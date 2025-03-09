package postgres

import (
	"context"
	"database/sql"
	stderrors "errors"
	"github.com/jmoiron/sqlx"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/entity"
	"manga-reader2/internal/domain/repository"
	"time"
)

// UserRepository реализация интерфейса repository.UserRepository для PostgreSQL
type UserRepository struct {
	db  *sqlx.DB
	log logger.Logger
}

// NewUserRepository создает новый экземпляр UserRepository
func NewUserRepository(db *sqlx.DB, log logger.Logger) repository.UserRepository {
	return &UserRepository{
		db:  db,
		log: log,
	}
}

// Create создает нового пользователя в базе данных
func (r *UserRepository) Create(ctx context.Context, user *entity.User) (int64, error) {
	query := `
		INSERT INTO users (username, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	var id int64
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowxContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_username_key\" (SQLSTATE 23505)" {
			return 0, errors.NewUserExistsError(user.Username)
		}
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return 0, errors.NewConflictError("Пользователь с таким email уже существует", nil)
		}

		r.log.Error("Ошибка создания пользователя", "error", err.Error())
		return 0, errors.NewDatabaseError("Ошибка создания пользователя", err)
	}

	user.ID = id
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt

	return id, nil
}

// GetByID получает пользователя по идентификатору
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user entity.User
	err := r.db.GetContext(ctx, &user, query, id)

	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewUserNotFoundError(id)
		}
		r.log.Error("Ошибка получения пользователя", "error", err.Error(), "id", id)
		return nil, errors.NewDatabaseError("Ошибка получения пользователя", err)
	}

	return &user, nil
}

// GetByUsername получает пользователя по имени пользователя
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user entity.User
	err := r.db.GetContext(ctx, &user, query, username)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewUserNotFoundError(username)
		}
		r.log.Error("Ошибка получения пользователя по имени", "error", err.Error(), "username", username)
		return nil, errors.NewDatabaseError("Ошибка получения пользователя", err)
	}

	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, username, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user entity.User
	err := r.db.GetContext(ctx, &user, query, email)

	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewUserNotFoundError(email)
		}
		r.log.Error("Ошибка получения пользователя по email", "error", err.Error(), "email", email)
		return nil, errors.NewDatabaseError("Ошибка получения пользователя", err)
	}

	return &user, nil
}

// Update обновляет информацию о пользователе
func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, password_hash = $3, role = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`

	var updatedAt time.Time
	result, err := r.db.QueryContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
		user.Role,
		user.ID,
	)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_username_key\" (SQLSTATE 23505)" {
			return errors.NewUserExistsError(user.Username)
		}
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return errors.NewConflictError("Пользователь с таким email уже существует", nil)
		}

		r.log.Error("Ошибка обновления пользователя", "error", err.Error(), "id", user.ID)
		return errors.NewDatabaseError("Ошибка обновления пользователя", err)
	}
	defer result.Close()

	if !result.Next() {
		return errors.NewUserNotFoundError(user.ID)
	}

	if err := result.Scan(&updatedAt); err != nil {
		r.log.Error("Ошибка сканирования даты обновления", "error", err.Error())
		return errors.NewDatabaseError("Ошибка обновления пользователя", err)
	}

	user.UpdatedAt = updatedAt

	return nil
}

// Delete удаляет пользователя по идентификатору
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.log.Error("Ошибка удаления пользователя", "error", err.Error(), "id", id)
		return errors.NewDatabaseError("Ошибка удаления пользователя", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Ошибка получения количества удаленных строк", "error", err.Error())
		return errors.NewDatabaseError("Ошибка удаления пользователя", err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundError(id)
	}

	return nil
}
