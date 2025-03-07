package repository

import (
	"context"
	"manga-reader2/internal/domain/entity"
)

// ChapterRepository определяет интерфейс для репозитория глав
type ChapterRepository interface {
	Create(ctx context.Context, chapter *entity.Chapter) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Chapter, error)
	ListByManga(ctx context.Context, mangaID int64) ([]*entity.Chapter, error)
	Update(ctx context.Context, chapter *entity.Chapter) error
	Delete(ctx context.Context, id int64) error
	DeleteByMangaID(ctx context.Context, mangaID int64) error
}
