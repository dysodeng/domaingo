package user

import (
	"context"

	domainmodel "github.com/CXeon/domaingo/internal/domain/user/model"
	"github.com/CXeon/domaingo/internal/infrastructure/persistence/user/model"
	"github.com/CXeon/domaingo/internal/infrastructure/rdb"
	"github.com/google/uuid"
)

type repository struct {
	rdb *rdb.Rdb
}

func NewRepository(rdb *rdb.Rdb) *repository {
	return &repository{rdb: rdb}
}

func (r *repository) FindByID(ctx context.Context, uid uuid.UUID) (*domainmodel.User, error) {
	var u model.User
	if err := r.rdb.Reader().WithContext(ctx).Where("uid = ?", uid.String()).First(&u).Error; err != nil {
		return nil, err
	}
	return u.ToEntity()
}

func (r *repository) Create(ctx context.Context, user *domainmodel.User) (uuid.UUID, error) {
	u := model.FromUserEntity(user)
	if err := r.rdb.Writer().WithContext(ctx).Create(u).Error; err != nil {
		return uuid.Nil, err
	}
	return user.UID, nil
}

func (r *repository) Update(ctx context.Context, user *domainmodel.User) error {
	u := model.FromUserEntity(user)
	return r.rdb.Writer().WithContext(ctx).Where("uid = ?", u.UID).Updates(u).Error
}

func (r *repository) Delete(ctx context.Context, uid uuid.UUID) error {
	return r.rdb.Writer().WithContext(ctx).Where("uid = ?", uid.String()).Delete(&model.User{}).Error
}
