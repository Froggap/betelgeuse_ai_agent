package authvalueobjects

import (
	"strings"

	commondomainerrors "scrapper-ai/modules/shared/domain/errors"
)

const (
	minUsernameLength = 3
	maxUsernameLength = 50
)

type UsernameVO struct {
	value string
}

func NewUsernameVO(raw string) (*UsernameVO, error) {
	value := strings.TrimSpace(raw)

	if value == "" {
		return nil, commondomainerrors.NewValidationError(
			"userName",
			"username is required",
		)
	}

	if len(value) < minUsernameLength || len(value) > maxUsernameLength {
		return nil, commondomainerrors.NewValidationError(
			"userName",
			"username must be between 3 and 50 characters",
		)
	}

	return &UsernameVO{value: value}, nil
}

func (u *UsernameVO) Value() string {
	return u.value
}

func (u *UsernameVO) String() string {
	return u.value
}
