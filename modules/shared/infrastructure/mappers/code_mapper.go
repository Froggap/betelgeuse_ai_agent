package sharedinframappers

import (
	"net/http"
	commondomainenums "scrapper-ai/modules/shared/domain/enums"
	commondomainerrors "scrapper-ai/modules/shared/domain/errors"
	globalerrors "scrapper-ai/modules/shared/infrastructure/errors"
)

type ErrorConfig struct {
	Status int
	Title  string
}

var domainErrorMap = map[commondomainenums.ErrorCode]ErrorConfig{
	commondomainenums.CodeValidation: {
		Status: http.StatusBadRequest,
		Title:  "Validation Error",
	},
	commondomainenums.CodeConflict: {
		Status: http.StatusConflict,
		Title:  "Conflict",
	},
	commondomainenums.CodeNotFound: {
		Status: http.StatusNotFound,
		Title:  "Not Found",
	},
	commondomainenums.CodeUnauthorized: {
		Status: http.StatusUnauthorized,
		Title:  "Unauthorized",
	},
}

func MapDomainError(err *commondomainerrors.DomainError) *globalerrors.AppError {

	config, exists := domainErrorMap[err.Code]

	if !exists {
		return globalerrors.NewAppError(
			http.StatusInternalServerError,
			"Internal Server Error",
			"An unexpected error occurred",
			nil,
		)
	}

	return globalerrors.NewAppError(
		config.Status,
		config.Title,
		err.Message,
		nil,
	)
}
