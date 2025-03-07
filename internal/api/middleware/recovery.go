package middleware

import (
	"fmt"
	"manga-reader2/internal/api/response"
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
	"net/http"
	"runtime/debug"
)

// Recovery middleware для восстановления после паники
func Recovery(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()

					errMsg := fmt.Sprintf("Паника: %v", err)

					log.Error("Восстановление после паники",
						"error", errMsg,
						"stack", string(stack),
						"url", r.URL.String(),
						"method", r.Method,
					)

					appErr := errors.NewInternalError("Внутренняя ошибка сервера", nil)
					response.Error(w, log, appErr)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
