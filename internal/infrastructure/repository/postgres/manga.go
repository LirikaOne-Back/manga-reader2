package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	stderrors "errors"

	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/entity"
	"manga-reader2/internal/domain/repository"
)

// MangaRepository реализует интерфейс repository.MangaRepository для PostgreSQL
type MangaRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewMangaRepository создает новый экземпляр MangaRepository
func NewMangaRepository(db *sqlx.DB, log logger.Logger) repository.MangaRepository {
	return &MangaRepository{
		db:  db,
		log: log,
	}
}

// Create создает новую мангу в базе данных
func (r *MangaRepository) Create(ctx context.Context, manga *entity.Manga) (int64, error) {
	query := `
		INSERT INTO manga (title, description, cover_image, status, author, artist, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowxContext(
		ctx,
		query,
		manga.Title,
		manga.Description,
		manga.CoverImage,
		manga.Status,
		manga.Author,
		manga.Artist,
	).Scan(&id)

	if err != nil {
		r.log.Error("Ошибка создания манги", "error", err.Error())
		return 0, errors.NewDatabaseError("Ошибка создания манги", err)
	}

	if len(manga.Genres) > 0 {
		for _, genre := range manga.Genres {
			err = r.AddGenreToManga(ctx, id, genre)
			if err != nil {
				r.log.Error("Ошибка добавления жанра к манге", "error", err.Error(), "manga_id", id, "genre", genre)
			}
		}
	}

	return id, nil
}

// GetByID получает мангу по идентификатору
func (r *MangaRepository) GetByID(ctx context.Context, id int64) (*entity.Manga, error) {
	query := `
		SELECT id, title, description, cover_image, status, author, artist, created_at, updated_at
		FROM manga
		WHERE id = $1
	`

	manga := &entity.Manga{}
	err := r.db.GetContext(ctx, manga, query, id)

	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewMangaNotFoundError(id)
		}
		r.log.Error("Ошибка получения манги", "error", err.Error(), "id", id)
		return nil, errors.NewDatabaseError("Ошибка получения манги", err)
	}

	genres, err := r.GetGenresForManga(ctx, id)
	if err != nil {
		r.log.Error("Ошибка получения жанров для манги", "error", err.Error(), "manga_id", id)
	} else {
		manga.Genres = genres
	}

	return manga, nil
}

// List получает список манг с пагинацией и фильтрацией
func (r *MangaRepository) List(ctx context.Context, filter entity.MangaFilter) ([]*entity.Manga, error) {
	queryParts := []string{
		"SELECT id, title, description, cover_image, status, author, artist, created_at, updated_at FROM manga",
	}

	var where []string
	var args []interface{}
	argIndex := 1

	if filter.Title != "" {
		where = append(where, fmt.Sprintf("title ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Title+"%")
		argIndex++
	}

	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if len(filter.Genres) > 0 {
		queryParts = append(queryParts, "JOIN manga_genres mg ON manga.id = mg.manga_id")
		queryParts = append(queryParts, "JOIN genres g ON mg.genre_id = g.id")

		genreConditions := make([]string, len(filter.Genres))
		for i, genre := range filter.Genres {
			genreConditions[i] = fmt.Sprintf("g.name = $%d", argIndex)
			args = append(args, genre)
			argIndex++
		}
		where = append(where, "("+strings.Join(genreConditions, " OR ")+")")

		queryParts = append(queryParts, "GROUP BY manga.id")
	}

	if len(where) > 0 {
		queryParts = append(queryParts, "WHERE "+strings.Join(where, " AND "))
	}

	queryParts = append(queryParts, "ORDER BY updated_at DESC")

	if filter.Limit > 0 {
		queryParts = append(queryParts, fmt.Sprintf("LIMIT $%d", argIndex))
		args = append(args, filter.Limit)
		argIndex++

		queryParts = append(queryParts, fmt.Sprintf("OFFSET $%d", argIndex))
		args = append(args, filter.Offset)
	}

	query := strings.Join(queryParts, " ")

	var mangas []*entity.Manga
	err := r.db.SelectContext(ctx, &mangas, query, args...)

	if err != nil {
		r.log.Error("Ошибка получения списка манги", "error", err.Error())
		return nil, errors.NewDatabaseError("Ошибка получения списка манги", err)
	}

	for _, manga := range mangas {
		genres, err := r.GetGenresForManga(ctx, manga.ID)
		if err != nil {
			r.log.Error("Ошибка получения жанров для манги", "error", err.Error(), "manga_id", manga.ID)
		} else {
			manga.Genres = genres
		}
	}

	return mangas, nil
}

// Update обновляет информацию о манге
func (r *MangaRepository) Update(ctx context.Context, manga *entity.Manga) error {
	query := `
		UPDATE manga 
		SET title = $1, description = $2, cover_image = $3, status = $4, author = $5, artist = $6, updated_at = NOW()
		WHERE id = $7
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		manga.Title,
		manga.Description,
		manga.CoverImage,
		manga.Status,
		manga.Author,
		manga.Artist,
		manga.ID,
	)

	if err != nil {
		r.log.Error("Ошибка обновления манги", "error", err.Error(), "id", manga.ID)
		return errors.NewDatabaseError("Ошибка обновления манги", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Ошибка получения количества обновленных строк", "error", err.Error())
		return errors.NewDatabaseError("Ошибка обновления манги", err)
	}

	if rowsAffected == 0 {
		return errors.NewMangaNotFoundError(manga.ID)
	}

	currentGenres, err := r.GetGenresForManga(ctx, manga.ID)
	if err != nil {
		r.log.Error("Ошибка получения текущих жанров", "error", err.Error(), "manga_id", manga.ID)
	}

	var genresToAdd, genresToRemove []string

	for _, newGenre := range manga.Genres {
		found := false
		for _, currentGenre := range currentGenres {
			if newGenre == currentGenre {
				found = true
				break
			}
		}
		if !found {
			genresToAdd = append(genresToAdd, newGenre)
		}
	}

	for _, currentGenre := range currentGenres {
		found := false
		for _, newGenre := range manga.Genres {
			if currentGenre == newGenre {
				found = true
				break
			}
		}
		if !found {
			genresToRemove = append(genresToRemove, currentGenre)
		}
	}

	for _, genre := range genresToAdd {
		err = r.AddGenreToManga(ctx, manga.ID, genre)
		if err != nil {
			r.log.Error("Ошибка добавления жанра к манге", "error", err.Error(), "manga_id", manga.ID, "genre", genre)
		}
	}

	for _, genre := range genresToRemove {
		err = r.RemoveGenreFromManga(ctx, manga.ID, genre)
		if err != nil {
			r.log.Error("Ошибка удаления жанра у манги", "error", err.Error(), "manga_id", manga.ID, "genre", genre)
		}
	}

	return nil
}

// Delete удаляет мангу по идентификатору
func (r *MangaRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM manga WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Ошибка удаления манги", "error", err.Error(), "id", id)
		return errors.NewDatabaseError("Ошибка удаления манги", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Ошибка получения количества удаленных строк", "error", err.Error())
		return errors.NewDatabaseError("Ошибка удаления манги", err)
	}

	if rowsAffected == 0 {
		return errors.NewMangaNotFoundError(id)
	}

	return nil
}

// GetPopular получает список популярных манг (by views)
func (r *MangaRepository) GetPopular(ctx context.Context, limit int) ([]*entity.MangaStat, error) {
	query := `
		SELECT m.id as manga_id, m.title, COUNT(mv.id) as views
		FROM manga m
		JOIN manga_views mv ON m.id = mv.manga_id
		GROUP BY m.id, m.title
		ORDER BY views DESC
		LIMIT $1
	`

	var stats []*entity.MangaStat
	err := r.db.SelectContext(ctx, &stats, query, limit)

	if err != nil {
		r.log.Error("Ошибка получения популярных манг", "error", err.Error())
		return nil, errors.NewDatabaseError("Ошибка получения популярных манг", err)
	}

	return stats, nil
}

// AddGenreToManga добавляет жанр к манге
func (r *MangaRepository) AddGenreToManga(ctx context.Context, mangaID int64, genre string) error {
	genreID, err := r.getOrCreateGenre(ctx, genre)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO manga_genres (manga_id, genre_id)
		VALUES ($1, $2)
		ON CONFLICT (manga_id, genre_id) DO NOTHING
	`

	_, err = r.db.ExecContext(ctx, query, mangaID, genreID)
	if err != nil {
		r.log.Error("Ошибка добавления жанра к манге", "error", err.Error(), "manga_id", mangaID, "genre_id", genreID)
		return errors.NewDatabaseError("Ошибка добавления жанра к манге", err)
	}

	return nil
}

// RemoveGenreFromManga удаляет жанр у манги
func (r *MangaRepository) RemoveGenreFromManga(ctx context.Context, mangaID int64, genre string) error {
	query := `SELECT id FROM genres WHERE name = $1`
	var genreID int64
	err := r.db.QueryRowxContext(ctx, query, genre).Scan(&genreID)

	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil
		}
		r.log.Error("Ошибка получения ID жанра", "error", err.Error(), "genre", genre)
		return errors.NewDatabaseError("Ошибка получения ID жанра", err)
	}

	query = `DELETE FROM manga_genres WHERE manga_id = $1 AND genre_id = $2`
	_, err = r.db.ExecContext(ctx, query, mangaID, genreID)

	if err != nil {
		r.log.Error("Ошибка удаления жанра у манги", "error", err.Error(), "manga_id", mangaID, "genre_id", genreID)
		return errors.NewDatabaseError("Ошибка удаления жанра у манги", err)
	}

	return nil
}

// GetGenresForManga получает список жанров для манги
func (r *MangaRepository) GetGenresForManga(ctx context.Context, mangaID int64) ([]string, error) {
	query := `
		SELECT g.name
		FROM genres g
		JOIN manga_genres mg ON g.id = mg.genre_id
		WHERE mg.manga_id = $1
		ORDER BY g.name
	`

	var genres []string
	err := r.db.SelectContext(ctx, &genres, query, mangaID)

	if err != nil {
		r.log.Error("Ошибка получения жанров для манги", "error", err.Error(), "manga_id", mangaID)
		return nil, errors.NewDatabaseError("Ошибка получения жанров для манги", err)
	}

	return genres, nil
}

// getOrCreateGenre получает существующий жанр или создает новый
func (r *MangaRepository) getOrCreateGenre(ctx context.Context, genre string) (int64, error) {
	query := `SELECT id FROM genres WHERE name = $1`
	var id int64
	err := r.db.QueryRowxContext(ctx, query, genre).Scan(&id)

	if err == nil {
		return id, nil
	}

	if stderrors.Is(err, sql.ErrNoRows) {
		r.log.Error("Ошибка проверки существования жанра", "error", err.Error(), "genre", genre)
		return 0, errors.NewDatabaseError("Ошибка проверки существования жанра", err)
	}

	query = `INSERT INTO genres (name) VALUES ($1) RETURNING id`
	err = r.db.QueryRowxContext(ctx, query, genre).Scan(&id)

	if err != nil {
		r.log.Error("Ошибка создания жанра", "error", err.Error(), "genre", genre)
		return 0, errors.NewDatabaseError("Ошибка создания жанра", err)
	}

	return id, nil
}
