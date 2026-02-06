package domain

import (
	"time"

	"github.com/marcelofabianov/wisp"
)

// --- Input ---

type NewUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Phone string `json:"phone"`
}

type UpdateUserInput struct {
	Name string `json:"name"`
}

type UpdateUserPasswordInput struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
	UpdatedBy   wisp.AuditUser
}

// --- Entity ---

type User struct {
	ID           wisp.UUID           `db:"id"`
	Name         wisp.NonEmptyString `db:"name"`
	Email        wisp.Email          `db:"email"`
	PasswordHash wisp.NonEmptyString `db:"password"`
	Phone        wisp.Phone          `db:"phone"`
	Role         UserRole            `db:"role"`
	IsActive     bool                `db:"is_active"`
	wisp.Audit
}

func NewUser(input *NewUserInput, passwordHash wisp.NonEmptyString, createdBy wisp.AuditUser) (*User, error) {
	id, err := wisp.NewUUID()
	if err != nil {
		return nil, NewErrUserFailedGenerateUuid(err)
	}

	name, err := wisp.NewNonEmptyString(input.Name)
	if err != nil {
		return nil, NewErrUserInvalidName(input.Name)
	}

	email, err := wisp.NewEmail(input.Email)
	if err != nil {
		return nil, NewErrUserInvalidEmail(input.Email)
	}

	phone, err := wisp.NewPhone(input.Phone)
	if err != nil {
		return nil, NewErrUserInvalidPhone(input.Phone)
	}

	role, err := NewUserRole(input.Role)
	if err != nil {
		return nil, NewErrUserInvalidRole(input.Role)
	}

	return &User{
		ID:           id,
		Name:         name,
		Email:        email,
		Phone:        phone,
		Role:         role,
		IsActive:     true,
		PasswordHash: passwordHash,
		Audit:        wisp.NewAudit(createdBy),
	}, nil
}

func (u *User) Update(input *UpdateUserInput, updatedBy wisp.AuditUser) error {
	name, err := wisp.NewNonEmptyString(input.Name)
	if err != nil {
		return NewErrUserInvalidName(input.Name)
	}

	u.Name = name
	u.Audit.Touch(updatedBy)

	return nil
}

func (u *User) ChangePassword(newPasswordHash wisp.NonEmptyString, updatedBy wisp.AuditUser) {
	u.PasswordHash = newPasswordHash
	u.Audit.Touch(updatedBy)
}

func (u *User) Activate(updatedBy wisp.AuditUser) {
	u.IsActive = true
	u.Audit.Touch(updatedBy)
}

func (u *User) Deactivate(updatedBy wisp.AuditUser) error {
	if !u.IsActive {
		return NewErrUserAlreadyInactive()
	}

	u.IsActive = false
	u.Audit.Touch(updatedBy)

	return nil
}

func (u *User) IsActiveAndNotArchived() bool {
	return u.IsActive && !u.Audit.ArchivedAt.Valid
}

func (u *User) ToggleArchived(updatedBy wisp.AuditUser) error {
	if u.Audit.ArchivedAt.Valid {
		u.Audit.ArchivedAt = wisp.EmptyNullableTime
	} else {
		u.Audit.ArchivedAt = wisp.NewNullableTime(time.Now())
	}
	u.Audit.Touch(updatedBy)
	return nil
}

func (u *User) Delete(deletedBy wisp.AuditUser) {
	u.Audit.DeletedAt = wisp.NewNullableTime(time.Now())
	u.Audit.Touch(deletedBy)
}
