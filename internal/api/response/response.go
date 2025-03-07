package response

import (
	"encoding/json"
	"net/http"

	stderrors "errors" // Стандартный пакет errors с псевдонимом
	"manga-reader2/internal/common/errors"
	"manga-reader2/internal/common/logger"
)

// Response содержит общую структуру ответа API
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// MetaPagination содержит информацию о пагинации
type MetaPagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
}

// ErrorResponse описывает структуру ошибки в ответе API
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// JSON отправляет JSON-ответ с указанным кодом состояния
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Success отправляет успешный ответ с данными
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	resp := Response{
		Success: true,
		Data:    data,
	}

	JSON(w, statusCode, resp)
}

// SuccessWithMeta отправляет успешный ответ с данными и метаданными
func SuccessWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta interface{}) {
	resp := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}

	JSON(w, statusCode, resp)
}

// Error отправляет ответ с ошибкой
func Error(w http.ResponseWriter, log logger.Logger, err error) {
	var statusCode int
	var errorResp interface{}

	var appErr *errors.AppError
	if stderrors.As(err, &appErr) {
		statusCode = appErr.StatusCode
		errorResp = ErrorResponse{
			Code:    string(appErr.Code),
			Message: appErr.Message,
			Details: appErr.Details,
		}

		if appErr.Err != nil {
			log.Error("Ошибка обработки запроса",
				"statusCode", statusCode,
				"errorCode", appErr.Code,
				"message", appErr.Message,
				"error", appErr.Err.Error(),
			)
		} else {
			log.Error("Ошибка обработки запроса",
				"statusCode", statusCode,
				"errorCode", appErr.Code,
				"message", appErr.Message,
			)
		}
	} else {
		statusCode = http.StatusInternalServerError
		errorResp = ErrorResponse{
			Code:    string(errors.ErrorCodeInternal),
			Message: "Внутренняя ошибка сервера",
		}
		log.Error("Неизвестная ошибка", "error", err.Error())
	}

	resp := Response{
		Success: false,
		Error:   errorResp,
	}

	JSON(w, statusCode, resp)
}

// Created отправляет ответ с кодом 201 (Created) и данными
func Created(w http.ResponseWriter, data interface{}) {
	Success(w, http.StatusCreated, data)
}

// NoContent отправляет ответ с кодом 204 (No Content)
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// BadRequest отправляет ответ с кодом 400 (Bad Request) и сообщением об ошибке
func BadRequest(w http.ResponseWriter, log logger.Logger, message string) {
	appErr := errors.NewBadRequestError(message, nil)
	Error(w, log, appErr)
}

// Unauthorized отправляет ответ с кодом 401 (Unauthorized) и сообщением об ошибке
func Unauthorized(w http.ResponseWriter, log logger.Logger, message string) {
	appErr := errors.NewUnauthorizedError(message, nil)
	Error(w, log, appErr)
}

// Forbidden отправляет ответ с кодом 403 (Forbidden) и сообщением об ошибке
func Forbidden(w http.ResponseWriter, log logger.Logger, message string) {
	appErr := errors.NewForbiddenError(message, nil)
	Error(w, log, appErr)
}

// NotFound отправляет ответ с кодом 404 (Not Found) и сообщением об ошибке
func NotFound(w http.ResponseWriter, log logger.Logger, message string) {
	appErr := errors.NewNotFoundError(message, nil)
	Error(w, log, appErr)
}

// InternalServerError отправляет ответ с кодом 500 (Internal Server Error) и сообщением об ошибке
func InternalServerError(w http.ResponseWriter, log logger.Logger, message string) {
	appErr := errors.NewInternalError(message, nil)
	Error(w, log, appErr)
}
