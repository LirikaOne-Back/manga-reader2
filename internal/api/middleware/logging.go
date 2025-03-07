package middleware

import (
	"manga-reader2/internal/common/logger"
	"net/http"
	"time"
)

// RequestLogging middleware для логирования запросов
func RequestLogging(log logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := NewWrapResponseWriter(w, r.ProtoMajor)

			requestLog := log.WithFields(map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			})

			requestID := r.Header.Get("X-Request-ID")
			if requestID != "" {
				requestLog = requestLog.With("request_id", requestID)
			}
			requestLog.Info("Начало обработки запроса")

			next.ServeHTTP(ww, r)

			duration := time.Since(start)

			requestLog = requestLog.WithFields(map[string]interface{}{
				"status":   ww.Status(),
				"duration": duration.String(),
				"bytes":    ww.BytesWritten(),
			})

			if ww.Status() >= 500 {
				requestLog.Error("Завершение обработки запроса")
			} else if ww.Status() >= 400 {
				requestLog.Warn("Завершение обработки запроса")
			} else {
				requestLog.Info("Завершение обработки запроса")
			}
		})
	}
}

// ResponseWriterWrapper интерфейс для обертки над http.ResponseWriter с доп. методами
type ResponseWriterWrapper interface {
	http.ResponseWriter
	Status() int
	BytesWritten() int
}

// WrapResponseWriter структура, реализующая ResponseWriterWrapper
type WrapResponseWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
}

// NewWrapResponseWriter создает новый экземпляр WrapResponseWriter
func NewWrapResponseWriter(w http.ResponseWriter, protoMajor int) ResponseWriterWrapper {
	return &WrapResponseWriter{ResponseWriter: w, status: http.StatusOK}
}

// Status возвращает код статуса ответа
func (w *WrapResponseWriter) Status() int {
	return w.status
}

// BytesWritten возвращает количество записанных байт
func (w *WrapResponseWriter) BytesWritten() int {
	return w.bytesWritten
}

// WriteHeader записывает код статуса ответа
func (w *WrapResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// Write записывает данные в ответ
func (w *WrapResponseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}
