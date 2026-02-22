package models

import "net/http"

type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewAPIError(code, message string, status int, details any) *APIError {
	return &APIError{Code: code, Message: message, HTTPStatus: status, Details: details}
}

var (
	ErrValidation   = NewAPIError("VALIDATION_ERROR", "validation failed", http.StatusBadRequest, nil)
	ErrUnauthorized = NewAPIError("UNAUTHORIZED", "unauthorized", http.StatusUnauthorized, nil)
	ErrForbidden    = NewAPIError("FORBIDDEN", "forbidden", http.StatusForbidden, nil)
	ErrNotFound     = NewAPIError("NOT_FOUND", "resource not found", http.StatusNotFound, nil)
	ErrConflict     = NewAPIError("CONFLICT", "resource conflict", http.StatusConflict, nil)
	ErrInternal     = NewAPIError("INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, nil)
)
