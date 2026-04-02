package user

import (
	"context"

	domainuser "github.com/CXeon/domaingo/internal/domain/user"
	"github.com/CXeon/domaingo/internal/domain/user/model"
	"github.com/google/uuid"
)

type Service interface {
	GetUser(ctx context.Context, uid string) (*UserDTO, error)
	CreateUser(ctx context.Context, dto *CreateUserDTO) (string, error)
	UpdateUser(ctx context.Context, dto *UpdateUserDTO) error
	DeleteUser(ctx context.Context, uid string) error
}

type service struct {
	domainSvc domainuser.Service
}

func NewService(domainSvc domainuser.Service) Service {
	return &service{domainSvc: domainSvc}
}

func (s *service) GetUser(ctx context.Context, uid string) (*UserDTO, error) {
	id, err := uuid.Parse(uid)
	if err != nil {
		return nil, err
	}
	u, err := s.domainSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toDTO(u), nil
}

func (s *service) CreateUser(ctx context.Context, dto *CreateUserDTO) (string, error) {
	u := &model.User{
		UID:      uuid.New(),
		Name:     dto.Name,
		Gender:   model.Gender(dto.Gender),
		Role:     model.Role(dto.Role),
		Email:    dto.Email,
		Password: dto.Password,
	}
	if err := u.Check(); err != nil {
		return "", err
	}
	uid, err := s.domainSvc.Save(ctx, u)
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

func (s *service) UpdateUser(ctx context.Context, dto *UpdateUserDTO) error {
	uid, err := uuid.Parse(dto.UID)
	if err != nil {
		return err
	}
	u := &model.User{
		UID:      uid,
		Name:     dto.Name,
		Gender:   model.Gender(dto.Gender),
		Role:     model.Role(dto.Role),
		Email:    dto.Email,
		Password: dto.Password,
	}
	if err = u.Check(); err != nil {
		return err
	}
	_, err = s.domainSvc.Save(ctx, u)
	return err
}

func (s *service) DeleteUser(ctx context.Context, uid string) error {
	id, err := uuid.Parse(uid)
	if err != nil {
		return err
	}
	return s.domainSvc.Delete(ctx, id)
}
