package middleware

import (
	"context"
	"manga-reader2/internal/api/response"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"manga-reader2/internal/infrastructure/auth"
	"net/http"
	"strings"
)

// ContextKey используется для хранения данных в контексте
type ContextKey string

const (
	// UserIDKey ключ для ID пользователя в контексте
	UserIDKey ContextKey = "user_id"
	// UserRoleKey ключ для роли пользователя в контексте
	UserRoleKey ContextKey = "user_role"
	// UsernameKey ключ для имени пользователя в контексте
	UsernameKey ContextKey = "username"
)

// Authentication middleware для проверки JWT токена
func Authentication(jwtService *auth.JWTService, log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, log, errors.NewUnauthorizedError("Отсутствует заголовок Authorization", nil))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, log, errors.NewUnauthorizedError("Неверный формат заголовка Authorization", nil))
				return
			}

			claims, err := jwtService.ValidateAccessToken(parts[1])
			if err != nil {
				response.Error(w, log, errors.NewJWTInvalidError(err))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware для проверки роли пользователя
func RequireRole(roles ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				response.Unauthorized(w, nil, "Требуется авторизация")
				return
			}

			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				response.Forbidden(w, nil, "Отказано в доступе")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
