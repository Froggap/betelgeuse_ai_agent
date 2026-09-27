package shareddtovalidators

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	globalerrors "scrapper-ai/modules/shared/infrastructure/errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

type DTOValidator struct {
	validate *validator.Validate
}

func NewDTOValidator() *DTOValidator {
	v := validator.New()

	// Opcional: Esto hace que use el nombre de la etiqueta JSON en los errores
	// en lugar del nombre de la propiedad de la Struct en Go (ej: "model_version" en vez de "ModelVersion")
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.Split(fld.Tag.Get("json"), ",")[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &DTOValidator{validate: v}
}

// ValidateStruct valida cualquier DTO y si falla, escupe directamente tu AppError listo para responder
func (dv *DTOValidator) ValidateStruct(s interface{}) error {
	err := dv.validate.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		// Juntamos los errores en una lista de strings legibles
		var errorMessages []string
		for _, fieldErr := range validationErrors {
			// Ejemplo: "prompt: min=5" o "model_version: required"
			errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", fieldErr.Field(), fieldErr.Tag()))
		}

		// Unimos todo separado por comas: "prompt: min=5, model_version: required"
		fullMessage := "Fields validation failed: " + strings.Join(errorMessages, ", ")

		return globalerrors.NewAppError(
			http.StatusUnprocessableEntity, // 422
			"Validation Failed",
			fullMessage,
			err, // Le pasamos el error original de la librería para debugging interno
		)
	}

	return globalerrors.NewAppError(
		http.StatusBadRequest,
		"Bad Request",
		err.Error(),
		err,
	)
}
