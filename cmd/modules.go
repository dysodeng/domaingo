package cmd

import (
	"github.com/CXeon/domaingo/internal/modular"
	authmodule "github.com/CXeon/domaingo/internal/modular/auth"
	usermodule "github.com/CXeon/domaingo/internal/modular/user"
)

func modules() []modular.Module {
	return []modular.Module{
		authmodule.New(),
		usermodule.New(),
	}
}
