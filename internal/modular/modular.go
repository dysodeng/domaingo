package modular

import (
	authdomain "github.com/CXeon/domaingo/internal/domain/auth"
	infraRdb "github.com/CXeon/domaingo/internal/infrastructure/rdb"
	tilesLogger "github.com/CXeon/tiles/logger"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	Rdb        *infraRdb.Rdb
	Logger     tilesLogger.Logger
	TokenStore authdomain.TokenStore
}

type Module interface {
	Build(deps Deps, g *gin.RouterGroup) error
}
