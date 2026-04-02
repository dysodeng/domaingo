package cmd

import (
	usermodel "github.com/CXeon/domaingo/internal/infrastructure/persistence/user/model"
)

func migrateModels() []any {
	return []any{
		&usermodel.User{},
	}
}
