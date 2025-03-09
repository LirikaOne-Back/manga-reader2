package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/entity"
	"manga-reader2/internal/domain/repository"
	"time"
)

// ChapterRepository реализация интерфейса repository.ChapterRepository для PostgreSQL
type ChapterRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewChapterRepository создает новый экземпляр ChapterRepository
func NewChapterRepository(db *sqlx.DB, log logger.Logger) repository.ChapterRepository {
	return &ChapterRepository{
		db:  db,
		log: log,
	}
}

// Create создает новую главу в базе данных
func (r *ChapterRepository) Create(ctx context.Context, chapter *entity.Chapter) (int64, error) {
	query := `
		INSERT INTO chapters (manga_id, number, title, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	var id int64
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowxContext(
		ctx,
		query,
		chapter.MangaID,
		chapter.Number,
		chapter.Title,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		r.log.Error("Ошибка создания главы", "error", err.Error())
		return 0, errors.NewDatabaseError("Ошибка создания главы", err)
	}

	chapter.ID = id
	chapter.CreatedAt = createdAt
	chapter.UpdatedAt = updatedAt

	return id, nil
}

// GetByID получает главу по идентификатору
func (r *ChapterRepository) GetByID(ctx context.Context, id int64) (*entity.Chapter, error) {
	query := `
		SELECT id, manga_id, number, title, created_at, updated_at
		FROM chapters
		WHERE id = $1
	`

	var chapter entity.Chapter
	err := r.db.GetContext(ctx, &chapter, query, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewChapterNotFoundError(id)
		}
		r.log.Error("Ошибка получения главы", "error", err.Error(), "id", id)
		return nil, errors.NewDatabaseError("Ошибка получения главы", err)
	}

	return &chapter, nil
}

// ListByManga получает список глав для манги
func (r *ChapterRepository) ListByManga(ctx context.Context, mangaID int64) ([]*entity.Chapter, error) {
	query := `
		SELECT id, manga_id, number, title, created_at, updated_at
		FROM chapters
		WHERE manga_id = $1
		ORDER BY number
	`

	var chapters []*entity.Chapter
	err := r.db.SelectContext(ctx, &chapters, query, mangaID)

	if err != nil {
		r.log.Error("Ошибка получения списка глав", "error", err.Error(), "manga_id", mangaID)
		return nil, errors.NewDatabaseError("Ошибка получения списка глав", err)
	}

	return chapters, nil
}

// Update обновляет информацию о главе
func (r *ChapterRepository) Update(ctx context.Context, chapter *entity.Chapter) error {
	query := `
		UPDATE chapters 
		SET number = $1, title = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`

	var updatedAt time.Time
	result, err := r.db.QueryContext(
		ctx,
		query,
		chapter.Number,
		chapter.Title,
		chapter.ID,
	)

	if err != nil {
		r.log.Error("Ошибка обновления главы", "error", err.Error(), "id", chapter.ID)
		return errors.NewDatabaseError("Ошибка обновления главы", err)
	}
	defer result.Close()

	if !result.Next() {
		return errors.NewChapterNotFoundError(chapter.ID)
	}

	if err := result.Scan(&updatedAt); err != nil {
		r.log.Error("Ошибка сканирования даты обновления", "error", err.Error())
		return errors.NewDatabaseError("Ошибка обновления главы", err)
	}

	chapter.UpdatedAt = updatedAt

	return nil
}

// Delete удаляет главу по идентификатору
func (r *ChapterRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM chapters WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Ошибка удаления главы", "error", err.Error(), "id", id)
		return errors.NewDatabaseError("Ошибка удаления главы", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Ошибка получения количества удаленных строк", "error", err.Error())
		return errors.NewDatabaseError("Ошибка удаления главы", err)
	}

	if rowsAffected == 0 {
		return errors.NewChapterNotFoundError(id)
	}

	return nil
}

// DeleteByMangaID удаляет все главы для манги
func (r *ChapterRepository) DeleteByMangaID(ctx context.Context, mangaID int64) error {
	query := "DELETE FROM chapters WHERE manga_id = $1"

	_, err := r.db.ExecContext(ctx, query, mangaID)
	if err != nil {
		r.log.Error("Ошибка удаления глав", "error", err.Error(), "manga_id", mangaID)
		return errors.NewDatabaseError(fmt.Sprintf("Ошибка удаления глав для манги %d", mangaID), err)
	}

	return nil
}
