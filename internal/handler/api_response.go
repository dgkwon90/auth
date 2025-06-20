// Package handler provides HTTP handlers and response types for the authentication service.
package handler

// APIResponse represents a standard API response structure for REST APIs.
type APIResponse struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// NewAPISuccess returns a successful APIResponse with optional custom message.
func NewAPISuccess(data any, code int, message ...string) APIResponse {
	msg := "요청이 성공적으로 처리되었습니다."
	if len(message) > 0 {
		msg = message[0]
	}
	return APIResponse{
		Success: true,
		Code:    code,
		Message: msg,
		Data:    data,
	}
}

// NewAPIError returns an error APIResponse with optional details.
func NewAPIError(code int, errMsg string, details ...any) APIResponse {
	var det any
	if len(details) > 0 && details[0] != nil {
		det = details[0]
	}
	return APIResponse{
		Success: false,
		Code:    code,
		Message: errMsg,
		Data:    det,
	}
}
