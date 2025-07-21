package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	er "github.com/Joshdike/subscriptions_aggregator/internal/pkg/errors"
)

// ParseMonthYear takes a string in the format "MM-YYYY" and returns a time.Time object representing the corresponding month and year.
// If the input string is not in the expected format, an error is returned.
func ParseMonthYear(dateStr string) (time.Time, error) {
	// Split into month and year
	parts := strings.Split(dateStr, "-")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid date format, expected MM-YYYY")
	}

	// Parse the month and year
	layout := "01-2006" // Go's reference time format (MM-YYYY)
	return time.Parse(layout, dateStr)
}

// ErrorResponse structure
// swagger:model ErrorResponse
type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

// WriteError writes a structures JSON error response based on the error type
// 500 Internal Server Error is the default response for unknown errors
func WriteError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Default error message
	message := "An error occurred"
	var status int
	var details string

	// Unwrap the error to check for known types
	switch {
	case errors.Is(err, er.ErrDecodingJSON):
		message = "Invalid Request Format"
		status = http.StatusBadRequest
	case errors.Is(err, er.ErrEncodingJSON):
		message = "Failed to process Response"
		status = http.StatusInternalServerError
	case errors.Is(err, er.ErrInvalidInput):
		message = "Validation failed"
		status = http.StatusBadRequest
		details = err.Error()
	case errors.Is(err, er.ErrSubscriptionNotFound):
		message = "No subscription found"
		status = http.StatusNotFound

	case errors.Is(err, er.ErrAlreadyExists):
		message = "Validation failed"
		status = http.StatusConflict
		details = err.Error()
	case errors.Is(err, er.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = err.Error()
	default:
		status = http.StatusInternalServerError
		details = "Internal server error"
	}
	w.WriteHeader(status)

	// Write the error response
	writeErr := json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Details: details,
	})
	// Handle encoding errors
	if writeErr != nil {
		http.Error(w, `{"error": "failed to encode error"}`, http.StatusInternalServerError)
		return
	}
}
