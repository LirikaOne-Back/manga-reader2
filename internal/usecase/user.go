package usecase

import (
	"context"
	stderrors "errors"
	"golang.org/x/crypto/bcrypt"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/domain/entity"
	"manga-reader2/internal/domain/repository"
	"manga-reader2/internal/infrastructure/auth"
	"regexp"
	"strings"
)

// UserUseCase интерфейс, определяющий бизнес-логику для работы с пользователями
type UserUseCase interface {
	Register(ctx context.Context, reg *entity.UserRegistration) (*entity.User, error)
	Login(ctx context.Context, cred *entity.UserCredentials) (*entity.TokenPair, error)
	RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenPair, error)
	GetProfile(ctx context.Context, userID int64) (*entity.User, error)
	UpdateProfile(ctx context.Context, user *entity.User) (*entity.User, error)
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
}

// userUseCase реализация интерфейса UserUseCase
type userUseCase struct {
	userRepo   repository.UserRepository
	jwtService *auth.JWTService
	log        logger.Logger
}

// NewUserUseCase создает новый экземпляр UserUseCase
func NewUserUseCase(
	userRepo repository.UserRepository,
	jwtService *auth.JWTService,
	log logger.Logger,
) UserUseCase {
	return &userUseCase{
		userRepo:   userRepo,
		jwtService: jwtService,
		log:        log,
	}
}

// Register регистрирует нового пользователя
func (uc *userUseCase) Register(ctx context.Context, reg *entity.UserRegistration) (*entity.User, error) {
	if reg.Username == "" {
		return nil, errors.NewValidationError("Имя пользователя не может быть пустым", nil)
	}
	if reg.Email == "" {
		return nil, errors.NewValidationError("Email не может быть пустым", nil)
	}
	if reg.Password == "" {
		return nil, errors.NewValidationError("Пароль не может быть пустым", nil)
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(reg.Username) {
		return nil, errors.NewValidationError("Имя пользователя может содержать только буквы, цифры и символ подчеркивания", nil)
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`).MatchString(reg.Email) {
		return nil, errors.NewValidationError("Некорректный email", nil)
	}

	if len(reg.Password) < 6 {
		return nil, errors.NewValidationError("Пароль должен содержать минимум 6 символов", nil)
	}

	_, err := uc.userRepo.GetByUsername(ctx, reg.Username)
	if err == nil {
		return nil, errors.NewUserExistsError(reg.Username)
	}
	var appError *errors.AppError
	if !stderrors.As(err, &appError) {
		return nil, err
	}

	_, err = uc.userRepo.GetByEmail(ctx, reg.Email)
	if err == nil {
		return nil, errors.NewConflictError("Пользователь с таким email уже существует", nil)
	}

	if !stderrors.As(err, &appError) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		uc.log.Error("Ошибка хеширования пароля", "error", err.Error())
		return nil, errors.NewInternalError("Ошибка хеширования пароля", err)
	}

	user := &entity.User{
		Username: reg.Username,
		Email:    reg.Email,
		Password: string(hashedPassword),
		Role:     "user",
	}

	id, err := uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	createdUser, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	createdUser.Password = ""

	return createdUser, nil
}

// Login аутентифицирует пользователя и возвращает токен
func (uc *userUseCase) Login(ctx context.Context, cred *entity.UserCredentials) (*entity.TokenPair, error) {
	if cred.Username == "" {
		return nil, errors.NewValidationError("Имя пользователя не может быть пустым", nil)
	}
	if cred.Password == "" {
		return nil, errors.NewValidationError("Пароль не может быть пустым", nil)
	}

	user, err := uc.userRepo.GetByUsername(ctx, cred.Username)
	if err != nil {
		if errors.IsErrorCode(err, errors.ErrorCodeUserNotFound) {
			user, err = uc.userRepo.GetByEmail(ctx, cred.Username)
			if err != nil {
				return nil, errors.NewInvalidCredentialsError()
			}
		} else {
			return nil, err
		}
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(cred.Password)); err != nil {
		return nil, errors.NewInvalidCredentialsError()
	}

	tokenPair, err := uc.jwtService.GenerateTokenPair(user)
	if err != nil {
		uc.log.Error("Ошибка генерации токена", "error", err.Error(), "user_id", user.ID)
		return nil, errors.NewInternalError("Ошибка генерации токена", err)
	}

	return tokenPair, nil
}

// RefreshToken обновляет токен
func (uc *userUseCase) RefreshToken(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {
	claims, err := uc.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "токен истек") {
			return nil, errors.NewJWTExpiredError()
		}
		return nil, errors.NewJWTInvalidError(err)
	}

	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	tokenPair, err := uc.jwtService.GenerateTokenPair(user)
	if err != nil {
		uc.log.Error("Ошибка генерации токена", "error", err.Error(), "user_id", user.ID)
		return nil, errors.NewInternalError("Ошибка генерации токена", err)
	}

	return tokenPair, nil
}

// GetProfile получает профиль пользователя
func (uc *userUseCase) GetProfile(ctx context.Context, userID int64) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Password = ""

	return user, nil
}

// UpdateProfile обновляет профиль пользователя
func (uc *userUseCase) UpdateProfile(ctx context.Context, user *entity.User) (*entity.User, error) {
	currentUser, err := uc.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	currentUser.Username = user.Username
	currentUser.Email = user.Email

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(currentUser.Username) {
		return nil, errors.NewValidationError("Имя пользователя может содержать только буквы, цифры и символ подчеркивания", nil)
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`).MatchString(currentUser.Email) {
		return nil, errors.NewValidationError("Некорректный email", nil)
	}

	if err := uc.userRepo.Update(ctx, currentUser); err != nil {
		return nil, err
	}

	updatedUser, err := uc.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	updatedUser.Password = ""

	return updatedUser, nil
}

// ChangePassword изменяет пароль пользователя
func (uc *userUseCase) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.NewInvalidCredentialsError()
	}

	if len(newPassword) < 6 {
		return errors.NewValidationError("Новый пароль должен содержать минимум 6 символов", nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		uc.log.Error("Ошибка хеширования пароля", "error", err.Error())
		return errors.NewInternalError("Ошибка хеширования пароля", err)
	}

	user.Password = string(hashedPassword)
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return err
	}

	return nil
}
