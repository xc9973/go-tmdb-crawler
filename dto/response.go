package dto

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// Success creates a success response
func Success(data interface{}) Response {
	return Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage creates a success response with custom message
func SuccessWithMessage(message string, data interface{}) Response {
	return Response{
		Code:    0,
		Message: message,
		Data:    data,
	}
}

// Error creates an error response
func Error(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
	}
}

// BadRequest creates a 400 bad request response
func BadRequest(message string) Response {
	return Error(400, message)
}

// NotFound creates a 404 not found response
func NotFound(message string) Response {
	return Error(404, message)
}

// InternalError creates a 500 internal server error response
func InternalError(message string) Response {
	return Error(500, message)
}
