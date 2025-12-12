package errors

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewAPIError(code, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

func (e *APIError) Error() string {
	return e.Message
}
