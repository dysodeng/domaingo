package router

import (
	"fmt"

	"github.com/CXeon/domaingo/api/http/middleware"
	"github.com/CXeon/domaingo/internal/config"
	"github.com/gin-gonic/gin"
)

func Default(r *gin.Engine) *gin.RouterGroup {
	prefix := fmt.Sprintf("/%s/%s/%s/api/v1",
		config.Config.Company,
		config.Config.Project,
		config.Config.ServiceName,
	)
	r.Use(middleware.CORS(), gin.Recovery())
	g := r.Group(prefix)
	
	return g
}
