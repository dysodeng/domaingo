package user

import (
	"net/http"

	"github.com/CXeon/domaingo/api/http/response"
	appuser "github.com/CXeon/domaingo/internal/application/user"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc appuser.Service
}

func NewHandler(svc appuser.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(g *gin.RouterGroup) {
	users := g.Group("/users")
	users.GET("/:uid", h.GetUser)
	users.POST("", h.CreateUser)
	users.PUT("/:uid", h.UpdateUser)
	users.DELETE("/:uid", h.DeleteUser)
}

func (h *Handler) GetUser(c *gin.Context) {
	uid := c.Param("uid")
	user, err := h.svc.GetUser(c.Request.Context(), uid)
	if err != nil {
		response.Fail(c, http.StatusNotFound, 404, err.Error())
		return
	}
	response.OK(c, &UserResponse{
		UID:       user.UID,
		Name:      user.Name,
		Gender:    user.Gender,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	uid, err := h.svc.CreateUser(c.Request.Context(), &appuser.CreateUserDTO{
		Name:     req.Name,
		Gender:   req.Gender,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, &CreateUserResponse{UID: uid})
}

func (h *Handler) UpdateUser(c *gin.Context) {
	uid := c.Param("uid")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, 400, err.Error())
		return
	}
	if err := h.svc.UpdateUser(c.Request.Context(), &appuser.UpdateUserDTO{
		UID:      uid,
		Name:     req.Name,
		Gender:   req.Gender,
		Email:    req.Email,
		Password: req.Password,
	}); err != nil {
		response.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	uid := c.Param("uid")
	if err := h.svc.DeleteUser(c.Request.Context(), uid); err != nil {
		response.Fail(c, http.StatusInternalServerError, 500, err.Error())
		return
	}
	response.OK(c, nil)
}
