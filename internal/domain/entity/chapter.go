package entity

import "time"

// Chapter представляет главу манги
type Chapter struct {
	ID        int64     `json:"id" db:"id"`
	MangaID   int64     `json:"manga_id" db:"manga_id"`
	Number    float64   `json:"number" db:"number"` // Используем float для поддержки глав типа 1.5
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ChapterWithStats представляет главу со статистикой
type ChapterWithStats struct {
	Chapter
	Views int64 `json:"views"`
}
