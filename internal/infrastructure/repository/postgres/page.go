package postgres

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/entity"
	"manga-reader2/internal/domain/repository"
	"time"
)

// PageRepository реализация интерфейса repository.PageRepository для PostgreSQL
type PageRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewPageRepository создает новый экземпляр PageRepository
func NewPageRepository(db *sqlx.DB, log logger.Logger) repository.PageRepository {
	return &PageRepository{
		db:  db,
		log: log,
	}
}

// Create создает новую страницу в базе данных
func (r *PageRepository) Create(ctx context.Context, page *entity.Page) (int64, error) {
	query := `
		INSERT INTO pages (chapter_id, number, image_path, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	var id int64
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowxContext(
		ctx,
		query,
		page.ChapterID,
		page.Number,
		page.ImagePath,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		r.log.Error("Ошибка создания страницы", "error", err.Error())
		return 0, errors.NewDatabaseError("Ошибка создания страницы", err)
	}

	page.ID = id
	page.CreatedAt = createdAt
	page.UpdatedAt = updatedAt

	return id, nil
}

// GetByID получает страницу по идентификатору
func (r *PageRepository) GetByID(ctx context.Context, id int64) (*entity.Page, error) {
	query := `
		SELECT id, chapter_id, number, image_path, created_at, updated_at
		FROM pages
		WHERE id = $1
	`

	var page entity.Page
	err := r.db.GetContext(ctx, &page, query, id)

	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewPageNotFoundError(id)
		}
		r.log.Error("Ошибка получения страницы", "error", err.Error(), "id", id)
		return nil, errors.NewDatabaseError("Ошибка получения страницы", err)
	}

	return &page, nil
}

// ListByChapter получает список страниц для главы
func (r *PageRepository) ListByChapter(ctx context.Context, chapterID int64) ([]*entity.Page, error) {
	query := `
		SELECT id, chapter_id, number, image_path, created_at, updated_at
		FROM pages
		WHERE chapter_id = $1
		ORDER BY number
	`

	var pages []*entity.Page
	err := r.db.SelectContext(ctx, &pages, query, chapterID)

	if err != nil {
		r.log.Error("Ошибка получения списка страниц", "error", err.Error(), "chapter_id", chapterID)
		return nil, errors.NewDatabaseError("Ошибка получения списка страниц", err)
	}

	return pages, nil
}

// Update обновляет информацию о странице
func (r *PageRepository) Update(ctx context.Context, page *entity.Page) error {
	query := `
		UPDATE pages 
		SET chapter_id = $1, number = $2, image_path = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`

	var updatedAt time.Time
	result, err := r.db.QueryContext(
		ctx,
		query,
		page.ChapterID,
		page.Number,
		page.ImagePath,
		page.ID,
	)

	if err != nil {
		r.log.Error("Ошибка обновления страницы", "error", err.Error(), "id", page.ID)
		return errors.NewDatabaseError("Ошибка обновления страницы", err)
	}
	defer result.Close()

	if !result.Next() {
		return errors.NewPageNotFoundError(page.ID)
	}

	if err = result.Scan(&updatedAt); err != nil {
		r.log.Error("Ошибка сканирования даты обновления", "error", err.Error())
		return errors.NewDatabaseError("Ошибка обновления страницы", err)
	}

	page.UpdatedAt = updatedAt

	return nil
}

// Delete удаляет страницу по идентификатору
func (r *PageRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM pages WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Ошибка удаления страницы", "error", err.Error(), "id", id)
		return errors.NewDatabaseError("Ошибка удаления страницы", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Ошибка получения количества удаленных строк", "error", err.Error())
		return errors.NewDatabaseError("Ошибка удаления страницы", err)
	}

	if rowsAffected == 0 {
		return errors.NewPageNotFoundError(id)
	}

	return nil
}

// DeleteByChapterID удаляет все страницы для главы
func (r *PageRepository) DeleteByChapterID(ctx context.Context, chapterID int64) error {
	query := "DELETE FROM pages WHERE chapter_id = $1"

	_, err := r.db.ExecContext(ctx, query, chapterID)
	if err != nil {
		r.log.Error("Ошибка удаления страниц", "error", err.Error(), "chapter_id", chapterID)
		return errors.NewDatabaseError(fmt.Sprintf("Ошибка удаления страниц для главы %d", chapterID), err)
	}

	return nil
}
