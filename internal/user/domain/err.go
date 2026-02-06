package domain

import (
	"errors"

	"github.com/marcelofabianov/fault"
)

const USER_AGGREGATE = "user"

var (
	// --- Validation ---
	ErrUserInvalidName     = errors.New("invalid name")
	ErrUserInvalidEmail    = errors.New("invalid email")
	ErrUserInvalidPassword = errors.New("invalid password")
	ErrUserInvalidRole     = errors.New("invalid role")
	ErrUserInvalidPhone    = errors.New("invalid phone")
	ErrUserAlreadyInactive = errors.New("user is already inactive")

	// --- Infrastructure ---
	ErrUserFailedGenerateUuid = errors.New("failed to generate user ID")
)

// --- Validation ---

func NewErrUserInvalidName(name string) error {
	return fault.Wrap(
		ErrUserInvalidName,
		ErrUserInvalidName.Error(),
		fault.WithCode(fault.DomainViolation),
		fault.WithContext("name", name),
		fault.WithContext("aggregate", USER_AGGREGATE),
	)
}

func NewErrUserInvalidEmail(email string) error {
	return fault.Wrap(
		ErrUserInvalidEmail,
		ErrUserInvalidEmail.Error(),
		fault.WithCode(fault.DomainViolation),
		fault.WithContext("email", email),
		fault.WithContext("aggregate", USER_AGGREGATE),
	)
}

func NewErrUserInvalidPassword() error {
	return fault.Wrap(
		ErrUserInvalidPassword,
		ErrUserInvalidPassword.Error(),
		fault.WithCode(fault.DomainViolation),
		fault.WithContext("aggregate", USER_AGGREGATE),
	)
}

func NewErrUserInvalidRole(role string) error {
	return fault.Wrap(
		ErrUserInvalidRole,
		ErrUserInvalidRole.Error(),
		fault.WithCode(fault.DomainViolation),
		fault.WithContext("role", role),
		fault.WithContext("aggregate", USER_AGGREGATE),
	)
}

func NewErrUserInvalidPhone(phone string) error {
	return fault.Wrap(
		ErrUserInvalidPhone,
		ErrUserInvalidPhone.Error(),
		fault.WithCode(fault.DomainViolation),
		fault.WithContext("phone", phone),
		fault.WithContext("aggregate", USER_AGGREGATE),
	)
}

func NewErrUserAlreadyInactive() error {
	return fault.Wrap(
		ErrUserAlreadyInactive,
		ErrUserAlreadyInactive.Error(),
		fault.WithCode(fault.DomainViolation),
		fault.WithContext("aggregate", USER_AGGREGATE),
	)
}

// --- Infrastructure ---

func NewErrUserFailedGenerateUuid(err error) error {
	return fault.Wrap(
		ErrUserFailedGenerateUuid,
		ErrUserFailedGenerateUuid.Error(),
		fault.WithCode(fault.Internal),
		fault.WithContext("error", err.Error()),
		fault.WithContext("aggregate", USER_AGGREGATE),
	)
}
