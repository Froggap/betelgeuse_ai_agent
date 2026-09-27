package sharedvalueobjects

import (
	commondomainerrors "scrapper-ai/modules/shared/domain/errors"
	"strings"

	"github.com/google/uuid"
)

type UUIDVO struct {
	value uuid.UUID
}

func NewUUIDVO(value string) (*UUIDVO, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil, commondomainerrors.NewValidationError(
			"uuid",
			"uuid is required",
		)
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return nil, commondomainerrors.NewValidationError(
			"uuid",
			"uuid must be a valid UUID",
		)
	}

	return &UUIDVO{
		value: id,
	}, nil
}

func (u UUIDVO) Value() uuid.UUID {
	return u.value
}

func (u UUIDVO) String() string {
	return u.value.String()
}
