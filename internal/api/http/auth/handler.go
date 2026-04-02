package auth

import (
	"strings"

	"github.com/CXeon/domaingo/api/http/response"
	appauth "github.com/CXeon/domaingo/internal/application/auth"
	tilecontext "github.com/CXeon/tiles/context"
	tileerrors "github.com/CXeon/tiles/errors"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc appauth.Service
}

func NewHandler(svc appauth.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(g *gin.RouterGroup) {
	auth := g.Group("/auth")
	auth.POST("/register", h.Register_)
	auth.POST("/login", h.Login)
	auth.POST("/logout", h.Logout)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/introspect", h.Introspect)
}

// RegisterProtected registers routes that require a valid Access Token.
// The caller is responsible for applying the Auth middleware to g.
func (h *Handler) RegisterProtected(g *gin.RouterGroup) {
	auth := g.Group("/auth")
	auth.GET("/me", h.Me)
	auth.PUT("/password", h.ChangePassword)
}

func (h *Handler) Register_(c *gin.Context) {
	appCtx := tilecontext.From(c.Request.Context())
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, tileerrors.ErrBadRequest.Wrap(err))
		return
	}
	if err := h.svc.Register(appCtx, &appauth.RegisterDTO{
		Name:     req.Name,
		Gender:   req.Gender,
		Email:    req.Email,
		Password: req.Password,
	}); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *Handler) Login(c *gin.Context) {
	appCtx := tilecontext.From(c.Request.Context())
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, tileerrors.ErrBadRequest.Wrap(err))
		return
	}
	pair, err := h.svc.Login(appCtx, &appauth.LoginDTO{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, &TokenPairResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	appCtx := tilecontext.From(c.Request.Context())
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, tileerrors.ErrBadRequest.Wrap(err))
		return
	}
	accessToken := extractBearer(c)
	if accessToken == "" {
		response.Fail(c, tileerrors.ErrUnauthorized)
		return
	}
	if err := h.svc.Logout(appCtx, accessToken, req.RefreshToken); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func (h *Handler) Refresh(c *gin.Context) {
	appCtx := tilecontext.From(c.Request.Context())
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, tileerrors.ErrBadRequest.Wrap(err))
		return
	}
	pair, err := h.svc.Refresh(appCtx, req.RefreshToken)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, &TokenPairResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

func (h *Handler) Introspect(c *gin.Context) {
	appCtx := tilecontext.From(c.Request.Context())
	accessToken := extractBearer(c)
	if accessToken == "" {
		response.Fail(c, tileerrors.ErrUnauthorized)
		return
	}
	info, err := h.svc.Introspect(appCtx, accessToken)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, &IntrospectResponse{
		UID:  info.UID,
		Role: info.Role,
	})
}

func (h *Handler) Me(c *gin.Context) {
	appCtx := tilecontext.From(c.Request.Context())
	uid, _ := c.Get("uid")
	me, err := h.svc.Me(appCtx, uid.(string))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, &MeResponse{
		UID:       me.UID,
		Name:      me.Name,
		Gender:    me.Gender,
		Role:      me.Role,
		Email:     me.Email,
		CreatedAt: me.CreatedAt,
	})
}

func (h *Handler) ChangePassword(c *gin.Context) {
	appCtx := tilecontext.From(c.Request.Context())
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, tileerrors.ErrBadRequest.Wrap(err))
		return
	}
	uid, _ := c.Get("uid")
	if err := h.svc.ChangePassword(appCtx, &appauth.ChangePasswordDTO{
		UID:         uid.(string),
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}

func extractBearer(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}
