package auth

import (
	"time"

	authhttp "github.com/CXeon/domaingo/internal/api/http/auth"
	"github.com/CXeon/domaingo/api/http/middleware"
	appauth "github.com/CXeon/domaingo/internal/application/auth"
	"github.com/CXeon/domaingo/internal/config"
	domainuser "github.com/CXeon/domaingo/internal/domain/user"
	userpersistence "github.com/CXeon/domaingo/internal/infrastructure/persistence/user"
	"github.com/CXeon/domaingo/internal/modular"
	"github.com/gin-gonic/gin"
)

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Build(deps modular.Deps, g *gin.RouterGroup) error {
	jwtCfg := config.Config.Base.JWT
	repo := userpersistence.NewRepository(deps.Rdb)

	var userRepo domainuser.Repository = repo

	svc := appauth.NewService(userRepo, deps.TokenStore, appauth.Config{
		Secret:     jwtCfg.Secret,
		AccessTTL:  time.Duration(jwtCfg.AccessTTL) * time.Second,
		RefreshTTL: time.Duration(jwtCfg.RefreshTTL) * time.Second,
	})
	handler := authhttp.NewHandler(svc)

	// Public routes: register, login, logout, refresh, introspect
	handler.Register(g)

	// Protected routes: me, change_password
	protected := g.Group("")
	protected.Use(middleware.Auth(jwtCfg.Secret, deps.TokenStore))
	handler.RegisterProtected(protected)

	return nil
}
