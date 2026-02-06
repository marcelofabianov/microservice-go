package domain

import (
	"database/sql/driver"
	"fmt"

	"github.com/marcelofabianov/wisp"
)

type UserRole wisp.Role

const (
	RoleAdmin  UserRole = "admin"
	RoleCommon UserRole = "common"
	RoleGuest  UserRole = "guest"
)

func init() {
	wisp.RegisterRoles(
		wisp.Role(RoleAdmin),
		wisp.Role(RoleCommon),
		wisp.Role(RoleGuest),
	)
}

func NewUserRole(role string) (UserRole, error) {
	r, err := wisp.NewRole(role)
	if err != nil {
		return "", err
	}

	return UserRole(r), nil
}

func (r UserRole) IsAdmin() bool {
	return r == RoleAdmin
}

func (r UserRole) IsCommon() bool {
	return r == RoleCommon
}

func (r UserRole) IsGuest() bool {
	return r == RoleGuest
}

func (r UserRole) CanLogin() bool {
	return r.IsAdmin() || r.IsCommon()
}

func (r UserRole) RequiresAction() bool {
	return r.IsGuest()
}

func (r UserRole) String() string {
	return string(r)
}

// Value implements the driver.Valuer interface.
func (r UserRole) Value() (driver.Value, error) {
	return r.String(), nil
}

// Scan implements the sql.Scanner interface.
func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into UserRole", value)
	}
	*r = UserRole(str)
	return nil
}
