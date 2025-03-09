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

// MangaUseCase интерфейс, определяющий бизнес-логику для работы с мангой
type MangaUseCase interface {
	Create(ctx context.Context, manga *entity.Manga) (*entity.Manga, error)
	GetByID(ctx context.Context, id int64) (*entity.Manga, error)
	List(ctx context.Context, filter entity.MangaFilter) ([]*entity.Manga, error)
	Update(ctx context.Context, manga *entity.Manga) (*entity.Manga, error)
	Delete(ctx context.Context, id int64) error
	GetChapters(ctx context.Context, mangaID int64) ([]*entity.Chapter, error)
	GetPopular(ctx context.Context, period entity.StatsPeriod, limit int) ([]*entity.MangaStat, error)
}

// mangaUseCase реализация интерфейса MangaUseCase
type mangaUseCase struct {
	mangaRepo     repository.MangaRepository
	cacheRepo     repository.CacheRepository
	analyticsRepo repository.AnalyticsRepository
	log           logger.Logger
}

// NewMangaUseCase создает новый экземпляр MangaUseCase
func NewMangaUseCase(
	mangaRepo repository.MangaRepository,
	cacheRepo repository.CacheRepository,
	analyticsRepo repository.AnalyticsRepository,
	log logger.Logger,
) MangaUseCase {
	return &mangaUseCase{
		mangaRepo:     mangaRepo,
		cacheRepo:     cacheRepo,
		analyticsRepo: analyticsRepo,
		log:           log,
	}
}

// Create создает новую мангу
func (uc *mangaUseCase) Create(ctx context.Context, manga *entity.Manga) (*entity.Manga, error) {
	if manga.Title == "" {
		return nil, errors.NewValidationError("Название манги не может быть пустым", nil)
	}

	id, err := uc.mangaRepo.Create(ctx, manga)
	if err != nil {
		return nil, err
	}

	createdManga, err := uc.mangaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = uc.invalidateMangaListCache(ctx); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка манги", "error", err.Error())
	}

	return createdManga, nil
}

// GetByID получает мангу по ID
func (uc *mangaUseCase) GetByID(ctx context.Context, id int64) (*entity.Manga, error) {
	cacheKey := fmt.Sprintf("manga:%d", id)
	cachedData, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var manga entity.Manga
		if err = json.Unmarshal([]byte(cachedData), &manga); err == nil {
			if err = uc.analyticsRepo.RecordMangaView(ctx, id); err != nil {
				uc.log.Error("Ошибка записи просмотра манги", "error", err.Error(), "manga_id", id)
			}
			return &manga, nil
		}
		uc.log.Error("Ошибка декодирования манги из кеша", "error", err.Error())
	}

	manga, err := uc.mangaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := uc.analyticsRepo.RecordMangaView(ctx, id); err != nil {
		uc.log.Error("Ошибка записи просмотра манги", "error", err.Error(), "manga_id", id)
	}

	if jsonData, err := json.Marshal(manga); err == nil {
		if err := uc.cacheRepo.Set(ctx, cacheKey, string(jsonData), 30*time.Minute); err != nil {
			uc.log.Error("Ошибка кеширования манги", "error", err.Error())
		}
	}

	return manga, nil
}

// List возвращает список манги с фильтрацией
func (uc *mangaUseCase) List(ctx context.Context, filter entity.MangaFilter) ([]*entity.Manga, error) {
	if filter.Title == "" && filter.Status == "" && len(filter.Genres) == 0 {
		cacheKey := fmt.Sprintf("manga:list:%d:%d", filter.Limit, filter.Offset)
		cachedData, err := uc.cacheRepo.Get(ctx, cacheKey)
		if err == nil && cachedData != "" {
			var mangas []*entity.Manga
			if err := json.Unmarshal([]byte(cachedData), &mangas); err == nil {
				return mangas, nil
			}
			uc.log.Error("Ошибка декодирования списка манги из кеша", "error", err.Error())
		}
	}

	mangas, err := uc.mangaRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	if filter.Title == "" && filter.Status == "" && len(filter.Genres) == 0 {
		cacheKey := fmt.Sprintf("manga:list:%d:%d", filter.Limit, filter.Offset)
		if jsonData, err := json.Marshal(mangas); err == nil {
			if err := uc.cacheRepo.Set(ctx, cacheKey, string(jsonData), 10*time.Minute); err != nil {
				uc.log.Error("Ошибка кеширования списка манги", "error", err.Error())
			}
		}
	}

	return mangas, nil
}

// Update обновляет мангу
func (uc *mangaUseCase) Update(ctx context.Context, manga *entity.Manga) (*entity.Manga, error) {
	_, err := uc.mangaRepo.GetByID(ctx, manga.ID)
	if err != nil {
		return nil, err
	}

	if manga.Title == "" {
		return nil, errors.NewValidationError("Название манги не может быть пустым", nil)
	}

	if err := uc.mangaRepo.Update(ctx, manga); err != nil {
		return nil, err
	}

	updatedManga, err := uc.mangaRepo.GetByID(ctx, manga.ID)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("manga:%d", manga.ID)
	if err := uc.cacheRepo.Delete(ctx, cacheKey); err != nil {
		uc.log.Error("Ошибка инвалидации кеша манги", "error", err.Error())
	}

	if err := uc.invalidateMangaListCache(ctx); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка манги", "error", err.Error())
	}

	return updatedManga, nil
}

// Delete удаляет мангу
func (uc *mangaUseCase) Delete(ctx context.Context, id int64) error {
	_, err := uc.mangaRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.mangaRepo.Delete(ctx, id); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("manga:%d", id)
	if err := uc.cacheRepo.Delete(ctx, cacheKey); err != nil {
		uc.log.Error("Ошибка инвалидации кеша манги", "error", err.Error())
	}

	if err := uc.invalidateMangaListCache(ctx); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка манги", "error", err.Error())
	}

	return nil
}

// GetChapters возвращает список глав манги
func (uc *mangaUseCase) GetChapters(ctx context.Context, mangaID int64) ([]*entity.Chapter, error) {
	_, err := uc.mangaRepo.GetByID(ctx, mangaID)
	if err != nil {
		return nil, err
	}

	// Заглушка

	return []*entity.Chapter{}, nil
}

// GetPopular возвращает список популярной манги
func (uc *mangaUseCase) GetPopular(ctx context.Context, period entity.StatsPeriod, limit int) ([]*entity.MangaStat, error) {
	cacheKey := fmt.Sprintf("manga:popular:%s:%d", period, limit)
	cachedData, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var popular []*entity.MangaStat
		if err = json.Unmarshal([]byte(cachedData), &popular); err == nil {
			return popular, nil
		}
		uc.log.Error("Ошибка декодирования популярной манги из кеша", "error", err.Error())
	}

	popular, err := uc.analyticsRepo.GetTopManga(ctx, period, limit)
	if err != nil {
		return nil, err
	}

	var cacheTTL time.Duration
	switch period {
	case entity.StatsPeriodDaily:
		cacheTTL = 1 * time.Hour
	case entity.StatsPeriodWeekly:
		cacheTTL = 4 * time.Hour
	case entity.StatsPeriodMonthly:
		cacheTTL = 12 * time.Hour
	default:
		cacheTTL = 24 * time.Hour
	}

	if jsonData, err := json.Marshal(popular); err == nil {
		if err := uc.cacheRepo.Set(ctx, cacheKey, string(jsonData), cacheTTL); err != nil {
			uc.log.Error("Ошибка кеширования популярной манги", "error", err.Error())
		}
	}

	return popular, nil
}

// invalidateMangaListCache инвалидирует кеш списка манги
func (uc *mangaUseCase) invalidateMangaListCache(ctx context.Context) error {
	return uc.cacheRepo.Delete(ctx, "manga:list:*")
}
