package domain

import "fmt"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func NewRole(roleStr string) (Role, error) {
	switch Role(roleStr) {
	case RoleUser:
		return RoleUser, nil
	case RoleAdmin:
		return RoleAdmin, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidRole, roleStr)
	}
}

func (r Role) String() string {
	return string(r)
}
