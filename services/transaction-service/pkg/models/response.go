package models

import (
	"encoding/json"
	"net/http"
	"time"
)

type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Envelope struct {
	Success bool           `json:"success"`
	Data    any            `json:"data,omitempty"`
	Error   *Error         `json:"error,omitempty"`
	Meta    map[string]any `json:"meta,omitempty"`
}

type SaleCreateResponse struct {
	ID           int64     `json:"id"`
	Status       string    `json:"status"`
	ResponseCode string    `json:"responseCode"`
	AuthCode     *string   `json:"authCode,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

type VoidCreateResponse struct {
	ID                    int64     `json:"id"`
	OriginalTransactionID int64     `json:"originalTransactionId"`
	Status                string    `json:"status"`
	ResponseCode          string    `json:"responseCode"`
	CreatedAt             time.Time `json:"createdAt"`
}

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Ok(w http.ResponseWriter, status int, data any, meta map[string]any) {
	JSON(w, status, Envelope{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

func Fail(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	JSON(w, status, Envelope{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
