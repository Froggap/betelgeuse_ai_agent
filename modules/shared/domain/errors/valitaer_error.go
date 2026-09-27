package commondomainerrors

import commondomainenums "scrapper-ai/modules/shared/domain/enums"

type DomainError struct {
	Code    commondomainenums.ErrorCode
	Field   string
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

func NewValidationError(field, message string) error {
	return &DomainError{
		Code:    commondomainenums.CodeValidation,
		Field:   field,
		Message: message,
	}
}
