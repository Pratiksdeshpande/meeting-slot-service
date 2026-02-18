package utils

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool           `json:"success"`
	Data       interface{}    `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

// WriteJSON writes a JSON response with the given status code
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteSuccess writes a success response
func WriteSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	response := Response{
		Success: true,
		Data:    data,
	}
	WriteJSON(w, statusCode, response)
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, statusCode int, code, message string) {
	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	WriteJSON(w, statusCode, response)
}

// WritePaginatedResponse writes a paginated response
func WritePaginatedResponse(w http.ResponseWriter, data interface{}, page, limit, total int) {
	response := PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: PaginationInfo{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	}
	WriteJSON(w, http.StatusOK, response)
}

// Common error responses
func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, "NOT_FOUND", message)
}

func WriteInternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
