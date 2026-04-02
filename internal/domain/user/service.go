package user

import (
	"context"

	"github.com/CXeon/domaingo/internal/domain/user/model"
	"github.com/google/uuid"
)

type Service interface {
	GetByID(ctx context.Context, uid uuid.UUID) (*model.User, error)

	Save(ctx context.Context, user *model.User) (uuid.UUID, error)

	Delete(ctx context.Context, uid uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewSvc(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, uid uuid.UUID) (*model.User, error) {
	u, err := s.repo.FindByID(ctx, uid)
	if err != nil {
		return nil, ErrUserNotFound.Wrap(err)
	}
	return u, nil
}

func (s *service) Save(ctx context.Context, user *model.User) (uuid.UUID, error) {
	// 检查用户是否存在
	if _, err := s.repo.FindByID(ctx, user.UID); err == nil {
		// 用户存在，更新用户信息
		if err = s.repo.Update(ctx, user); err != nil {
			return uuid.Nil, ErrUserSaveFail.Wrap(err)
		}
		return user.UID, nil
	}
	uid, err := s.repo.Create(ctx, user)
	if err != nil {
		return uuid.Nil, ErrUserSaveFail.Wrap(err)
	}
	return uid, nil
}

func (s *service) Delete(ctx context.Context, uid uuid.UUID) error {
	if err := s.repo.Delete(ctx, uid); err != nil {
		return ErrUserDeleteFail.Wrap(err)
	}
	return nil
}
