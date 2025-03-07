package repository

import (
	"context"
	"manga-reader2/internal/domain/entity"
)

// UserRepository определяет интерфейс для репозитория пользователей
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (int64, error)
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id int64) error
}
