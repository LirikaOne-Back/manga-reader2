package entity

// MangaStat представляет статистику по манге
type MangaStat struct {
	MangaID int64  `json:"manga_id" db:"manga_id"`
	Title   string `json:"title" db:"title"`
	Views   int64  `json:"views" db:"views"`
}

// ChapterStat представляет статистику по главе
type ChapterStat struct {
	ChapterID int64   `json:"chapter_id" db:"chapter_id"`
	MangaID   int64   `json:"manga_id" db:"manga_id"`
	Number    float64 `json:"number" db:"number"`
	Title     string  `json:"title" db:"title"`
	Views     int64   `json:"views" db:"views"`
}

// StatsPeriod представляет период статистики
type StatsPeriod string

const (
	StatsPeriodDaily   StatsPeriod = "daily"
	StatsPeriodWeekly  StatsPeriod = "weekly"
	StatsPeriodMonthly StatsPeriod = "monthly"
	StatsPeriodAllTime StatsPeriod = "all_time"
)
