package model

import (
	"github.com/CXeon/domaingo/internal/domain/user/model"
	"github.com/google/uuid"
)

func FromUserEntity(u *model.User) *User {
	return &User{
		UID:      u.UID.String(),
		Name:     u.Name,
		Gender:   uint8(u.Gender),
		Role:     uint8(u.Role),
		Email:    u.Email,
		Password: u.Password,
	}
}

func (u *User) ToEntity() (*model.User, error) {
	uid, err := uuid.Parse(u.UID)
	if err != nil {
		return nil, err
	}
	return &model.User{
		UID:       uid,
		Name:      u.Name,
		Gender:    model.Gender(u.Gender),
		Role:      model.Role(u.Role),
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
	}, nil
}
