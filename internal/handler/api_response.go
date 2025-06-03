package handler

// API 공통 응답 구조체 (REST Best Practice)
type ApiResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	// Error   *ApiError   `json:"error,omitempty"`
}

func NewApiSuccess(data interface{}, code int, message ...string) ApiResponse {
	msg := "요청이 성공적으로 처리되었습니다."
	if len(message) > 0 {
		msg = message[0]
	}
	return ApiResponse{
		Success: true,
		Code:    code,
		Message: msg,
		Data:    data,
	}
}

// 공통 에러 응답 생성
func NewApiError(code int, errMsg string, details ...interface{}) ApiResponse {
	var det interface{}
	if len(details) > 0 && details[0] != nil {
		det = details[0]
	}
	return ApiResponse{
		Success: false,
		Code:    code,
		Message: errMsg,
		Data:    det,
	}
}
