package request

import "strings"

type ValidationErrors struct {
	Message []string
}

func NewBadRequestError(message []string) *ValidationErrors {
	return &ValidationErrors{
		Message: message,
	}
}

func (e *ValidationErrors) Error() string {
	return strings.Join(e.Message, " ")
}
