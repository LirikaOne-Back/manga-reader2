package repository

import (
	"context"
	"manga-reader2/internal/domain/entity"
)

// PageRepository определяет интерфейс для репозитория страниц
type PageRepository interface {
	Create(ctx context.Context, page *entity.Page) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Page, error)
	ListByChapter(ctx context.Context, chapterID int64) ([]*entity.Page, error)
	Update(ctx context.Context, page *entity.Page) error
	Delete(ctx context.Context, id int64) error
	DeleteByChapterID(ctx context.Context, chapterID int64) error
}
