package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode представляет код ошибки
type ErrorCode string

// Константы кодов ошибок
const (
	// Общие ошибки
	ErrorCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrorCodeBadRequest   ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrorCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrorCodeConflict     ErrorCode = "CONFLICT"
	ErrorCodeValidation   ErrorCode = "VALIDATION_ERROR"

	// Ошибки базы данных
	ErrorCodeDatabase ErrorCode = "DATABASE_ERROR"

	// Ошибки манги
	ErrorCodeMangaNotFound   ErrorCode = "MANGA_NOT_FOUND"
	ErrorCodeChapterNotFound ErrorCode = "CHAPTER_NOT_FOUND"
	ErrorCodePageNotFound    ErrorCode = "PAGE_NOT_FOUND"

	// Ошибки пользователей
	ErrorCodeUserNotFound ErrorCode = "USER_NOT_FOUND"
	ErrorCodeUserExists   ErrorCode = "USER_ALREADY_EXISTS"
	ErrorCodeInvalidCreds ErrorCode = "INVALID_CREDENTIALS"

	// Ошибки JWT
	ErrorCodeJWTInvalid ErrorCode = "JWT_INVALID"
	ErrorCodeJWTExpired ErrorCode = "JWT_EXPIRED"
)

// AppError представляет ошибку приложения
type AppError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	StatusCode int         `json:"-"`
	Err        error       `json:"-"`
}

// Error реализует интерфейс error
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap возвращает вложенную ошибку
func (e *AppError) Unwrap() error {
	return e.Err
}

// Функции для создания различных типов ошибок

// NewInternalError создает ошибку внутреннего сервера
func NewInternalError(msg string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeInternal,
		Message:    msg,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewBadRequestError создает ошибку некорректного запроса
func NewBadRequestError(msg string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeBadRequest,
		Message:    msg,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// NewUnauthorizedError создает ошибку авторизации
func NewUnauthorizedError(msg string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeUnauthorized,
		Message:    msg,
		StatusCode: http.StatusUnauthorized,
		Err:        err,
	}
}

// NewForbiddenError создает ошибку доступа
func NewForbiddenError(msg string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeForbidden,
		Message:    msg,
		StatusCode: http.StatusForbidden,
		Err:        err,
	}
}

// NewNotFoundError создает ошибку "не найдено"
func NewNotFoundError(msg string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeNotFound,
		Message:    msg,
		StatusCode: http.StatusNotFound,
		Err:        err,
	}
}

// NewConflictError создает ошибку конфликта
func NewConflictError(msg string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeConflict,
		Message:    msg,
		StatusCode: http.StatusConflict,
		Err:        err,
	}
}

// NewValidationError создает ошибку валидации
func NewValidationError(msg string, details interface{}) *AppError {
	return &AppError{
		Code:       ErrorCodeValidation,
		Message:    msg,
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

// NewDatabaseError создает ошибку базы данных
func NewDatabaseError(msg string, err error) *AppError {
	return &AppError{
		Code:       ErrorCodeDatabase,
		Message:    msg,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// Специфические ошибки для доменных объектов

// NewMangaNotFoundError создает ошибку "манга не найдена"
func NewMangaNotFoundError(id interface{}) *AppError {
	return &AppError{
		Code:       ErrorCodeMangaNotFound,
		Message:    fmt.Sprintf("Манга с ID %v не найдена", id),
		StatusCode: http.StatusNotFound,
	}
}

// NewChapterNotFoundError создает ошибку "глава не найдена"
func NewChapterNotFoundError(id interface{}) *AppError {
	return &AppError{
		Code:       ErrorCodeChapterNotFound,
		Message:    fmt.Sprintf("Глава с ID %v не найдена", id),
		StatusCode: http.StatusNotFound,
	}
}

// NewPageNotFoundError создает ошибку "страница не найдена"
func NewPageNotFoundError(id interface{}) *AppError {
	return &AppError{
		Code:       ErrorCodePageNotFound,
		Message:    fmt.Sprintf("Страница с ID %v не найдена", id),
		StatusCode: http.StatusNotFound,
	}
}

// NewUserNotFoundError создает ошибку "пользователь не найден"
func NewUserNotFoundError(identifier interface{}) *AppError {
	return &AppError{
		Code:       ErrorCodeUserNotFound,
		Message:    fmt.Sprintf("Пользователь %v не найден", identifier),
		StatusCode: http.StatusNotFound,
	}
}

// NewUserExistsError создает ошибку "пользователь уже существует"
func NewUserExistsError(username string) *AppError {
	return &AppError{
		Code:       ErrorCodeUserExists,
		Message:    fmt.Sprintf("Пользователь с именем %s уже существует", username),
		StatusCode: http.StatusConflict,
	}
}

// NewInvalidCredentialsError создает ошибку "неверные учетные данные"
func NewInvalidCredentialsError() *AppError {
	return &AppError{
		Code:       ErrorCodeInvalidCreds,
		Message:    "Неверное имя пользователя или пароль",
		StatusCode: http.StatusUnauthorized,
	}
}

// NewJWTInvalidError создает ошибку "недействительный JWT токен"
func NewJWTInvalidError(err error) *AppError {
	return &AppError{
		Code:       ErrorCodeJWTInvalid,
		Message:    "Недействительный токен авторизации",
		StatusCode: http.StatusUnauthorized,
		Err:        err,
	}
}

// NewJWTExpiredError создает ошибку "истекший JWT токен"
func NewJWTExpiredError() *AppError {
	return &AppError{
		Code:       ErrorCodeJWTExpired,
		Message:    "Срок действия токена авторизации истек",
		StatusCode: http.StatusUnauthorized,
	}
}

// IsErrorCode проверяет, соответствует ли ошибка указанному коду ошибки
func IsErrorCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// IsNotFoundError проверяет, является ли ошибка ошибкой "не найдено"
func IsNotFoundError(err error) bool {
	return IsErrorCode(err, ErrorCodeNotFound) ||
		IsErrorCode(err, ErrorCodeMangaNotFound) ||
		IsErrorCode(err, ErrorCodeChapterNotFound) ||
		IsErrorCode(err, ErrorCodePageNotFound) ||
		IsErrorCode(err, ErrorCodeUserNotFound)
}

// IsValidationError проверяет, является ли ошибка ошибкой валидации
func IsValidationError(err error) bool {
	return IsErrorCode(err, ErrorCodeValidation)
}

// IsConflictError проверяет, является ли ошибка ошибкой конфликта
func IsConflictError(err error) bool {
	return IsErrorCode(err, ErrorCodeConflict) ||
		IsErrorCode(err, ErrorCodeUserExists)
}

// IsDatabaseError проверяет, является ли ошибка ошибкой базы данных
func IsDatabaseError(err error) bool {
	return IsErrorCode(err, ErrorCodeDatabase)
}

// IsUnauthorizedError проверяет, является ли ошибка ошибкой авторизации
func IsUnauthorizedError(err error) bool {
	return IsErrorCode(err, ErrorCodeUnauthorized) ||
		IsErrorCode(err, ErrorCodeJWTInvalid) ||
		IsErrorCode(err, ErrorCodeJWTExpired) ||
		IsErrorCode(err, ErrorCodeInvalidCreds)
}

// IsForbiddenError проверяет, является ли ошибка ошибкой доступа
func IsForbiddenError(err error) bool {
	return IsErrorCode(err, ErrorCodeForbidden)
}
