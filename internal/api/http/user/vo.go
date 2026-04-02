package user

import "time"

// --- Request ---

type CreateUserRequest struct {
	Name     string `json:"name"     binding:"required"`
	Gender   uint8  `json:"gender"   binding:"oneof=0 1 2"`
	Role     uint8  `json:"role"     binding:"oneof=0 1"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"     binding:"required"`
	Gender   uint8  `json:"gender"   binding:"oneof=0 1 2"`
	Role     uint8  `json:"role"     binding:"oneof=0 1"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// --- Response ---

type UserResponse struct {
	UID       string    `json:"uid"`
	Name      string    `json:"name"`
	Gender    uint8     `json:"gender"`
	Role      uint8     `json:"role"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserResponse struct {
	UID string `json:"uid"`
}
