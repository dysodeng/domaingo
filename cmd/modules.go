package cmd

import (
	"github.com/CXeon/domaingo/internal/modular"
	usermodule "github.com/CXeon/domaingo/internal/modular/user"
)

func modules() []modular.Module {
	return []modular.Module{
		usermodule.New(),
		// order.New(),
	}
}
