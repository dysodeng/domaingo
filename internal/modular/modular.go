package modular

import (
	infraRdb "github.com/CXeon/domaingo/internal/infrastructure/rdb"
	tilesLogger "github.com/CXeon/tiles/logger"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	Rdb    *infraRdb.Rdb
	Logger tilesLogger.Logger
}

type Module interface {
	Build(deps Deps, g *gin.RouterGroup) error
}
