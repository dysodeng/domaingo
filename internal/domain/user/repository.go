package user

import (
	"context"

	"github.com/CXeon/domaingo/internal/domain/user/model"
	"github.com/google/uuid"
)

type Repository interface {
	FindByID(ctx context.Context, uid uuid.UUID) (*model.User, error)
	Create(ctx context.Context, user *model.User) (uuid.UUID, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, uid uuid.UUID) error
}
