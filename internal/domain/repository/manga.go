package repository

import (
	"context"
	"manga-reader2/internal/domain/entity"
)

// MangaRepository определяет интерфейс для репозитория манги
type MangaRepository interface {
	Create(ctx context.Context, manga *entity.Manga) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.Manga, error)
	List(ctx context.Context, filter entity.MangaFilter) ([]*entity.Manga, error)
	Update(ctx context.Context, manga *entity.Manga) error
	Delete(ctx context.Context, id int64) error

	// Дополнительные методы
	GetPopular(ctx context.Context, limit int) ([]*entity.MangaStat, error)
	AddGenreToManga(ctx context.Context, mangaID int64, genre string) error
	RemoveGenreFromManga(ctx context.Context, mangaID int64, genre string) error
	GetGenresForManga(ctx context.Context, mangaID int64) ([]string, error)
}
