package repository

import (
	"context"
	"manga-reader2/internal/domain/entity"
)

// AnalyticsRepository определяет интерфейс для работы с аналитикой
type AnalyticsRepository interface {
	RecordMangaView(ctx context.Context, mangaID int64) error
	RecordChapterView(ctx context.Context, chapterID, mangaID int64) error
	RecordPageView(ctx context.Context, pageID, chapterID, mangaID int64) error

	GetMangaViews(ctx context.Context, mangaID int64) (int64, error)
	GetChapterViews(ctx context.Context, chapterID int64) (int64, error)

	GetTopManga(ctx context.Context, period entity.StatsPeriod, limit int) ([]*entity.MangaStat, error)
	GetTopChapters(ctx context.Context, period entity.StatsPeriod, limit int) ([]*entity.ChapterStat, error)

	ResetStats(ctx context.Context, period entity.StatsPeriod) error
}
