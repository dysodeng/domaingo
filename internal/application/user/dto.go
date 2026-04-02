package user

import (
	"time"

	"github.com/CXeon/domaingo/internal/domain/user/model"
)

// UserDTO is the response object returned to the caller.
// Password is intentionally excluded.
type UserDTO struct {
	UID       string
	Name      string
	Gender    uint8
	Role      uint8
	Email     string
	CreatedAt time.Time
}

type CreateUserDTO struct {
	Name     string
	Gender   uint8
	Role     uint8
	Email    string
	Password string
}

type UpdateUserDTO struct {
	UID      string
	Name     string
	Gender   uint8
	Role     uint8
	Email    string
	Password string
}

func toDTO(u *model.User) *UserDTO {
	return &UserDTO{
		UID:       u.UID.String(),
		Name:      u.Name,
		Gender:    uint8(u.Gender),
		Role:      uint8(u.Role),
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}
