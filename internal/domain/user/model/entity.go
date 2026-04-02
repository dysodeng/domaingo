package model

import (
	"errors"
	"time"

	"github.com/CXeon/tiles/util/regex"
	"github.com/google/uuid"
)

type User struct {
	UID       uuid.UUID
	Name      string
	Gender    Gender
	Role      Role
	Email     string
	Password  string
	CreatedAt time.Time
}

func (u *User) Check() error {
	if !u.Gender.IsValid() {
		return errors.New("invalid gender")
	}

	if !u.Role.IsValid() {
		return errors.New("invalid role")
	}

	if len(u.Name) == 0 {
		return errors.New("name cannot be empty")
	}

	if !regex.IsEmail(u.Email) {
		return errors.New("invalid email")
	}

	if len(u.Password) == 0 {
		return errors.New("password cannot be empty")
	}

	return nil

}
