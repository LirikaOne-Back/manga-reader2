package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/entity"
	"manga-reader2/internal/domain/repository"
	"time"
)

// ChapterUseCase интерфейс, определяющий бизнес-логику для работы с главами
type ChapterUseCase interface {
	Create(ctx context.Context, chapter *entity.Chapter) (*entity.Chapter, error)
	GetByID(ctx context.Context, id int64) (*entity.ChapterWithStats, error)
	ListByManga(ctx context.Context, mangaID int64) ([]*entity.Chapter, error)
	Update(ctx context.Context, chapter *entity.Chapter) (*entity.Chapter, error)
	Delete(ctx context.Context, id int64) error
	GetPages(ctx context.Context, chapterID int64) ([]*entity.Page, error)
}

// chapterUseCase реализация интерфейса ChapterUseCase
type chapterUseCase struct {
	chapterRepo   repository.ChapterRepository
	mangaRepo     repository.MangaRepository
	cacheRepo     repository.CacheRepository
	analyticsRepo repository.AnalyticsRepository
	log           logger.Logger
}

// NewChapterUseCase создает новый экземпляр ChapterUseCase
func NewChapterUseCase(
	chapterRepo repository.ChapterRepository,
	mangaRepo repository.MangaRepository,
	cacheRepo repository.CacheRepository,
	analyticsRepo repository.AnalyticsRepository,
	log logger.Logger,
) ChapterUseCase {
	return &chapterUseCase{
		chapterRepo:   chapterRepo,
		mangaRepo:     mangaRepo,
		cacheRepo:     cacheRepo,
		analyticsRepo: analyticsRepo,
		log:           log,
	}
}

// Create создает новую главу
func (uc *chapterUseCase) Create(ctx context.Context, chapter *entity.Chapter) (*entity.Chapter, error) {
	if chapter.Title == "" {
		return nil, errors.NewValidationError("Название главы не может быть пустым", nil)
	}

	if chapter.MangaID <= 0 {
		return nil, errors.NewValidationError("Не указан ID манги", nil)
	}

	_, err := uc.mangaRepo.GetByID(ctx, chapter.MangaID)
	if err != nil {
		return nil, err
	}

	id, err := uc.chapterRepo.Create(ctx, chapter)
	if err != nil {
		return nil, err
	}

	createdChapter, err := uc.chapterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := uc.invalidateChapterListCache(ctx, chapter.MangaID); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка глав", "error", err.Error(), "manga_id", chapter.MangaID)
	}

	return createdChapter, nil
}

// GetByID получает главу по ID с статистикой просмотров
func (uc *chapterUseCase) GetByID(ctx context.Context, id int64) (*entity.ChapterWithStats, error) {
	cacheKey := fmt.Sprintf("chapter:%d", id)
	cachedData, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var chapter entity.Chapter
		if err := json.Unmarshal([]byte(cachedData), &chapter); err == nil {
			views, err := uc.analyticsRepo.GetChapterViews(ctx, id)
			if err != nil {
				uc.log.Error("Ошибка получения просмотров главы", "error", err.Error(), "chapter_id", id)
				views = 0
			}

			if err := uc.analyticsRepo.RecordChapterView(ctx, id, chapter.MangaID); err != nil {
				uc.log.Error("Ошибка записи просмотра главы", "error", err.Error(), "chapter_id", id)
			}

			return &entity.ChapterWithStats{
				Chapter: chapter,
				Views:   views,
			}, nil
		}
		uc.log.Error("Ошибка декодирования главы из кеша", "error", err.Error())
	}

	chapter, err := uc.chapterRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	views, err := uc.analyticsRepo.GetChapterViews(ctx, id)
	if err != nil {
		uc.log.Error("Ошибка получения просмотров главы", "error", err.Error(), "chapter_id", id)
		views = 0
	}

	if err := uc.analyticsRepo.RecordChapterView(ctx, id, chapter.MangaID); err != nil {
		uc.log.Error("Ошибка записи просмотра главы", "error", err.Error(), "chapter_id", id)
	}

	if jsonData, err := json.Marshal(chapter); err == nil {
		if err := uc.cacheRepo.Set(ctx, cacheKey, string(jsonData), 30*time.Minute); err != nil {
			uc.log.Error("Ошибка кеширования главы", "error", err.Error())
		}
	}

	return &entity.ChapterWithStats{
		Chapter: *chapter,
		Views:   views,
	}, nil
}

// ListByManga возвращает список глав для манги
func (uc *chapterUseCase) ListByManga(ctx context.Context, mangaID int64) ([]*entity.Chapter, error) {
	_, err := uc.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("manga:%d:chapters", mangaID)
	cachedData, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var chapters []*entity.Chapter
		if err := json.Unmarshal([]byte(cachedData), &chapters); err == nil {
			return chapters, nil
		}
		uc.log.Error("Ошибка декодирования списка глав из кеша", "error", err.Error())
	}

	chapters, err := uc.chapterRepo.ListByManga(ctx, mangaID)
	if err != nil {
		return nil, err
	}

	if jsonData, err := json.Marshal(chapters); err == nil {
		if err := uc.cacheRepo.Set(ctx, cacheKey, string(jsonData), 15*time.Minute); err != nil {
			uc.log.Error("Ошибка кеширования списка глав", "error", err.Error())
		}
	}

	return chapters, nil
}

// Update обновляет главу
func (uc *chapterUseCase) Update(ctx context.Context, chapter *entity.Chapter) (*entity.Chapter, error) {
	existingChapter, err := uc.chapterRepo.GetByID(ctx, chapter.ID)
	if err != nil {
		return nil, err
	}

	if chapter.Title == "" {
		return nil, errors.NewValidationError("Название главы не может быть пустым", nil)
	}

	if err := uc.chapterRepo.Update(ctx, chapter); err != nil {
		return nil, err
	}

	updatedChapter, err := uc.chapterRepo.GetByID(ctx, chapter.ID)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("chapter:%d", chapter.ID)
	if err := uc.cacheRepo.Delete(ctx, cacheKey); err != nil {
		uc.log.Error("Ошибка инвалидации кеша главы", "error", err.Error())
	}

	if err := uc.invalidateChapterListCache(ctx, existingChapter.MangaID); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка глав", "error", err.Error(), "manga_id", existingChapter.MangaID)
	}

	return updatedChapter, nil
}

// Delete удаляет главу
func (uc *chapterUseCase) Delete(ctx context.Context, id int64) error {
	chapter, err := uc.chapterRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.chapterRepo.Delete(ctx, id); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("chapter:%d", id)
	if err := uc.cacheRepo.Delete(ctx, cacheKey); err != nil {
		uc.log.Error("Ошибка инвалидации кеша главы", "error", err.Error())
	}

	if err := uc.invalidateChapterListCache(ctx, chapter.MangaID); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка глав", "error", err.Error(), "manga_id", chapter.MangaID)
	}

	return nil
}

// GetPages возвращает список страниц для главы
func (uc *chapterUseCase) GetPages(ctx context.Context, chapterID int64) ([]*entity.Page, error) {
	// Этот метод будет реализован позже, когда мы добавим PageRepository
	// Сейчас это заглушка
	return []*entity.Page{}, nil
}

// invalidateChapterListCache инвалидирует кеш списка глав для манги
func (uc *chapterUseCase) invalidateChapterListCache(ctx context.Context, mangaID int64) error {
	cacheKey := fmt.Sprintf("manga:%d:chapters", mangaID)
	return uc.cacheRepo.Delete(ctx, cacheKey)
}
