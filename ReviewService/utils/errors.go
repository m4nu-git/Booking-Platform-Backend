package utils

// ServiceError carries an HTTP status code alongside a message so that
// the controller can respond with the correct status rather than always 500.
type ServiceError struct {
	Code    int
	Message string
}

func (e *ServiceError) Error() string { return e.Message }

func NewServiceError(code int, message string) *ServiceError {
	return &ServiceError{Code: code, Message: message}
}
