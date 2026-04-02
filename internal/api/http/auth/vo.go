package auth

import "time"

// --- Request ---

type RegisterRequest struct {
	Name     string `json:"name"     binding:"required"`
	Gender   uint8  `json:"gender"   binding:"oneof=0 1 2"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// --- Response ---

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type IntrospectResponse struct {
	UID  string `json:"uid"`
	Role uint8  `json:"role"`
}

type MeResponse struct {
	UID       string    `json:"uid"`
	Name      string    `json:"name"`
	Gender    uint8     `json:"gender"`
	Role      uint8     `json:"role"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
