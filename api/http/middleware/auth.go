package middleware

import (
	"errors"
	"strings"

	"github.com/CXeon/domaingo/api/http/response"
	authdomain "github.com/CXeon/domaingo/internal/domain/auth"
	tileerrors "github.com/CXeon/tiles/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type accessClaims struct {
	UID  string `json:"uid"`
	Role uint8  `json:"role"`
	jwt.RegisteredClaims
}

// Auth verifies the Bearer access token and injects uid/role into the gin context.
func Auth(secret string, store authdomain.TokenStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c)
		if tokenStr == "" {
			response.Fail(c, tileerrors.ErrUnauthorized)
			c.Abort()
			return
		}

		claims := &accessClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Fail(c, tileerrors.ErrUnauthorized)
			c.Abort()
			return
		}

		if store.IsAccessTokenBlocked(c.Request.Context(), claims.ID) {
			response.Fail(c, tileerrors.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set("uid", claims.UID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// Admin requires the requesting user to have role == 1 (admin).
// Must be used after Auth middleware.
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if r, ok := role.(uint8); !ok || r != 1 {
			response.Fail(c, tileerrors.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(header, "Bearer ")
}
