package utils

import "fmt"

// ErrorResponse - Validation errors or any others errors response body
type ErrorResponse struct {
	Code   string              `json:"code"`
	Errors map[string][]string `json:"errors,omitempty"`
	Error  map[string]string   `json:"error,omitempty"`
}

// NewErrorResponse - Single based error response body
func NewErrorResponse(statusCode int, field string, message string) ErrorResponse {

	var errorFields = map[string]string{}
	errorFields[field] = message

	return ErrorResponse{
		Code:  fmt.Sprintf("%d", statusCode),
		Error: errorFields,
	}
}

// NewValidationErrorResponse - Validation fields based errors
func NewValidationErrorResponse(statusCode int, verrs map[string][]string) ErrorResponse {
	return ErrorResponse{
		Code:   fmt.Sprintf("%d", statusCode),
		Errors: verrs,
	}
}
