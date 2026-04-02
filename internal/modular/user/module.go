package user

import (
	userhttp "github.com/CXeon/domaingo/internal/api/http/user"
	appuser "github.com/CXeon/domaingo/internal/application/user"
	"github.com/CXeon/domaingo/internal/config"
	domainuser "github.com/CXeon/domaingo/internal/domain/user"
	userpersistence "github.com/CXeon/domaingo/internal/infrastructure/persistence/user"
	"github.com/CXeon/domaingo/internal/modular"
	"github.com/CXeon/domaingo/api/http/middleware"
	"github.com/gin-gonic/gin"
)

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Build(deps modular.Deps, g *gin.RouterGroup) error {
	repo := userpersistence.NewRepository(deps.Rdb)
	domainSvc := domainuser.NewSvc(repo)
	appSvc := appuser.NewService(domainSvc)
	handler := userhttp.NewHandler(appSvc)

	// All user management routes require admin privileges.
	adminGroup := g.Group("")
	adminGroup.Use(
		middleware.Auth(config.Config.Base.JWT.Secret, deps.TokenStore),
		middleware.Admin(),
	)
	handler.Register(adminGroup)
	return nil
}
