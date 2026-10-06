// Package httpx defines the frozen JSON envelope and application error type
// shared by every handler.
package httpx

import (
	"fmt"
	"net/http"
)

// Frozen error codes (English). Human-readable messages are Bahasa Indonesia.
const (
	CodeInvalidRequest              = "INVALID_REQUEST"
	CodeRouteNotFound               = "ROUTE_NOT_FOUND"
	CodeMethodNotAllowed            = "METHOD_NOT_ALLOWED"
	CodeItemNotFound                = "ITEM_NOT_FOUND"
	CodeInsufficientStock           = "INSUFFICIENT_STOCK"
	CodeReservationNotFound         = "RESERVATION_NOT_FOUND"
	CodeReservationExpired          = "RESERVATION_EXPIRED"
	CodeReservationAlreadyConfirmed = "RESERVATION_ALREADY_CONFIRMED"
	CodeInternalError               = "INTERNAL_ERROR"
)

// ErrorDetail describes one invalid field.
type ErrorDetail struct {
	Field   string `json:"field,omitempty" example:"quantity"`
	Rule    string `json:"rule,omitempty" example:"gt"`
	Message string `json:"message" example:"quantity harus lebih dari 0"`
}

// ErrorInfo is the error body of the envelope.
type ErrorInfo struct {
	Code    string        `json:"code" example:"INSUFFICIENT_STOCK"`
	Message string        `json:"message" example:"stok tidak mencukupi"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorResponse is the frozen error envelope.
type ErrorResponse struct {
	Status string    `json:"status" example:"error"`
	Error  ErrorInfo `json:"error"`
}

// AppError carries everything a handler needs to render the envelope.
// It deliberately carries no exported JSON tags: it is never serialized
// directly.
type AppError struct {
	Status  int
	Code    string
	Message string
	Details []ErrorDetail
	Cause   error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the internal cause to errors.Is/errors.As.
func (e *AppError) Unwrap() error { return e.Cause }

// New builds a base application error.
func New(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// WithDetails returns a copy carrying field-level validation details.
func (e *AppError) WithDetails(details ...ErrorDetail) *AppError {
	clone := *e
	clone.Details = details
	return &clone
}

// Wrap returns a copy carrying an internal cause (never serialized).
func (e *AppError) Wrap(cause error) *AppError {
	clone := *e
	clone.Cause = cause
	return &clone
}

// Base platform errors. Domain modules build their own on top of New.
var (
	ErrInvalidRequest   = New(http.StatusBadRequest, CodeInvalidRequest, "request tidak valid")
	ErrRouteNotFound    = New(http.StatusNotFound, CodeRouteNotFound, "endpoint tidak ditemukan")
	ErrMethodNotAllowed = New(http.StatusMethodNotAllowed, CodeMethodNotAllowed, "method tidak diizinkan")
	ErrInternal         = New(http.StatusInternalServerError, CodeInternalError, "terjadi kesalahan pada server")
)

// Internal wraps any unexpected error into the frozen 500 response.
func Internal(cause error) *AppError {
	return ErrInternal.Wrap(cause)
}
