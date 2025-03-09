package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/entity"
	"manga-reader2/internal/domain/repository"
	"os"
	"path/filepath"
	"time"
)

// PageUseCase интерфейс, определяющий бизнес-логику для работы со страницами
type PageUseCase interface {
	Create(ctx context.Context, page *entity.Page) (*entity.Page, error)
	GetByID(ctx context.Context, id int64) (*entity.Page, error)
	ListByChapter(ctx context.Context, chapterID int64) ([]*entity.Page, error)
	Update(ctx context.Context, page *entity.Page) (*entity.Page, error)
	Delete(ctx context.Context, id int64) error
	UploadImage(ctx context.Context, chapterID int64, number int, filename string, imageData []byte) (*entity.Page, error)
}

// pageUseCase реализация интерфейса PageUseCase
type pageUseCase struct {
	pageRepo      repository.PageRepository
	chapterRepo   repository.ChapterRepository
	cacheRepo     repository.CacheRepository
	analyticsRepo repository.AnalyticsRepository
	log           logger.Logger
}

// NewPageUseCase создает новый экземпляр PageUseCase
func NewPageUseCase(
	pageRepo repository.PageRepository,
	chapterRepo repository.ChapterRepository,
	cacheRepo repository.CacheRepository,
	analyticsRepo repository.AnalyticsRepository,
	log logger.Logger,
) PageUseCase {
	return &pageUseCase{
		pageRepo:      pageRepo,
		chapterRepo:   chapterRepo,
		cacheRepo:     cacheRepo,
		analyticsRepo: analyticsRepo,
		log:           log,
	}
}

// Create создает новую страницу
func (uc *pageUseCase) Create(ctx context.Context, page *entity.Page) (*entity.Page, error) {
	if page.ChapterID <= 0 {
		return nil, errors.NewValidationError("Не указан ID главы", nil)
	}

	if page.ImagePath == "" {
		return nil, errors.NewValidationError("Не указан путь к изображению", nil)
	}

	_, err := uc.chapterRepo.GetByID(ctx, page.ChapterID)
	if err != nil {
		return nil, err
	}

	id, err := uc.pageRepo.Create(ctx, page)
	if err != nil {
		return nil, err
	}

	createdPage, err := uc.pageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := uc.invalidatePageListCache(ctx, page.ChapterID); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка страниц", "error", err.Error(), "chapter_id", page.ChapterID)
	}

	return createdPage, nil
}

// GetByID получает страницу по ID
func (uc *pageUseCase) GetByID(ctx context.Context, id int64) (*entity.Page, error) {
	cacheKey := fmt.Sprintf("page:%d", id)
	cachedData, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var page entity.Page
		if err := json.Unmarshal([]byte(cachedData), &page); err == nil {
			chapter, err := uc.chapterRepo.GetByID(ctx, page.ChapterID)
			if err == nil {
				if err := uc.analyticsRepo.RecordPageView(ctx, id, page.ChapterID, chapter.MangaID); err != nil {
					uc.log.Error("Ошибка записи просмотра страницы", "error", err.Error(), "page_id", id)
				}
			} else {
				uc.log.Error("Ошибка получения главы для аналитики", "error", err.Error(), "chapter_id", page.ChapterID)
			}
			return &page, nil
		}
		uc.log.Error("Ошибка декодирования страницы из кеша", "error", err.Error())
	}

	page, err := uc.pageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	chapter, err := uc.chapterRepo.GetByID(ctx, page.ChapterID)
	if err == nil {
		if err := uc.analyticsRepo.RecordPageView(ctx, id, page.ChapterID, chapter.MangaID); err != nil {
			uc.log.Error("Ошибка записи просмотра страницы", "error", err.Error(), "page_id", id)
		}
	} else {
		uc.log.Error("Ошибка получения главы для аналитики", "error", err.Error(), "chapter_id", page.ChapterID)
	}

	if jsonData, err := json.Marshal(page); err == nil {
		if err := uc.cacheRepo.Set(ctx, cacheKey, string(jsonData), 30*time.Minute); err != nil {
			uc.log.Error("Ошибка кеширования страницы", "error", err.Error())
		}
	}

	return page, nil
}

// ListByChapter возвращает список страниц для главы
func (uc *pageUseCase) ListByChapter(ctx context.Context, chapterID int64) ([]*entity.Page, error) {
	_, err := uc.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("chapter:%d:pages", chapterID)
	cachedData, err := uc.cacheRepo.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var pages []*entity.Page
		if err := json.Unmarshal([]byte(cachedData), &pages); err == nil {
			return pages, nil
		}
		uc.log.Error("Ошибка декодирования списка страниц из кеша", "error", err.Error())
	}

	pages, err := uc.pageRepo.ListByChapter(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	if jsonData, err := json.Marshal(pages); err == nil {
		if err := uc.cacheRepo.Set(ctx, cacheKey, string(jsonData), 15*time.Minute); err != nil {
			uc.log.Error("Ошибка кеширования списка страниц", "error", err.Error())
		}
	}

	return pages, nil
}

// Update обновляет страницу
func (uc *pageUseCase) Update(ctx context.Context, page *entity.Page) (*entity.Page, error) {
	existingPage, err := uc.pageRepo.GetByID(ctx, page.ID)
	if err != nil {
		return nil, err
	}

	if page.ChapterID <= 0 {
		return nil, errors.NewValidationError("Не указан ID главы", nil)
	}

	if page.ImagePath == "" {
		return nil, errors.NewValidationError("Не указан путь к изображению", nil)
	}

	_, err = uc.chapterRepo.GetByID(ctx, page.ChapterID)
	if err != nil {
		return nil, err
	}

	if err := uc.pageRepo.Update(ctx, page); err != nil {
		return nil, err
	}

	updatedPage, err := uc.pageRepo.GetByID(ctx, page.ID)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("page:%d", page.ID)
	if err := uc.cacheRepo.Delete(ctx, cacheKey); err != nil {
		uc.log.Error("Ошибка инвалидации кеша страницы", "error", err.Error())
	}

	if existingPage.ChapterID != page.ChapterID {
		if err := uc.invalidatePageListCache(ctx, existingPage.ChapterID); err != nil {
			uc.log.Error("Ошибка инвалидации кеша списка страниц", "error", err.Error(), "chapter_id", existingPage.ChapterID)
		}
	}
	if err := uc.invalidatePageListCache(ctx, page.ChapterID); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка страниц", "error", err.Error(), "chapter_id", page.ChapterID)
	}

	return updatedPage, nil
}

// Delete удаляет страницу
func (uc *pageUseCase) Delete(ctx context.Context, id int64) error {
	page, err := uc.pageRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := os.Remove(page.ImagePath); err != nil && !os.IsNotExist(err) {
		uc.log.Error("Ошибка удаления файла изображения", "error", err.Error(), "path", page.ImagePath)
	}

	if err := uc.pageRepo.Delete(ctx, id); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("page:%d", id)
	if err := uc.cacheRepo.Delete(ctx, cacheKey); err != nil {
		uc.log.Error("Ошибка инвалидации кеша страницы", "error", err.Error())
	}

	if err := uc.invalidatePageListCache(ctx, page.ChapterID); err != nil {
		uc.log.Error("Ошибка инвалидации кеша списка страниц", "error", err.Error(), "chapter_id", page.ChapterID)
	}

	return nil
}

// UploadImage загружает изображение и создает новую страницу
func (uc *pageUseCase) UploadImage(ctx context.Context, chapterID int64, number int, filename string, imageData []byte) (*entity.Page, error) {
	_, err := uc.chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return nil, err
	}

	uploadDir := fmt.Sprintf("uploads/chapters/%d", chapterID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		uc.log.Error("Ошибка создания директории для загрузки", "error", err.Error(), "dir", uploadDir)
		return nil, errors.NewInternalError("Ошибка создания директории для загрузки", err)
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	newFilename := fmt.Sprintf("%d_%d%s", chapterID, number, ext)
	imagePath := filepath.Join(uploadDir, newFilename)

	if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
		uc.log.Error("Ошибка записи файла", "error", err.Error(), "path", imagePath)
		return nil, errors.NewInternalError("Ошибка записи файла", err)
	}

	page := &entity.Page{
		ChapterID: chapterID,
		Number:    number,
		ImagePath: imagePath,
	}

	return uc.Create(ctx, page)
}

// invalidatePageListCache инвалидирует кеш списка страниц для главы
func (uc *pageUseCase) invalidatePageListCache(ctx context.Context, chapterID int64) error {
	cacheKey := fmt.Sprintf("chapter:%d:pages", chapterID)
	return uc.cacheRepo.Delete(ctx, cacheKey)
}
