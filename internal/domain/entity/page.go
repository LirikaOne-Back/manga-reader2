package entity

import "time"

// Page представляет страницу главы
type Page struct {
	ID        int64     `json:"id" db:"id"`
	ChapterID int64     `json:"chapter_id" db:"chapter_id"`
	Number    int       `json:"number" db:"number"`
	ImagePath string    `json:"image_path" db:"image_path"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
