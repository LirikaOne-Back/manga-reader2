package entity

import "time"

// Manga представляет сущность манги
type Manga struct {
	ID          int64     `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	CoverImage  string    `json:"cover_image,omitempty" db:"cover_image"`
	Status      string    `json:"status" db:"status"` // ongoing, completed, hiatus
	Author      string    `json:"author" db:"author"`
	Artist      string    `json:"artist,omitempty" db:"artist"`
	Genres      []string  `json:"genres,omitempty"` // Связь многие-ко-многим
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// MangaFilter представляет фильтры для поиска манги
type MangaFilter struct {
	Title  string   `json:"title,omitempty"`
	Genres []string `json:"genres,omitempty"`
	Status string   `json:"status,omitempty"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
}
