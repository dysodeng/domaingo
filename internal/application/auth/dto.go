package auth

import "time"

type RegisterDTO struct {
	Name     string
	Gender   uint8
	Email    string
	Password string
}

type LoginDTO struct {
	Email    string
	Password string
}

type TokenPairDTO struct {
	AccessToken  string
	RefreshToken string
}

type IntrospectDTO struct {
	UID  string
	Role uint8
}

type MeDTO struct {
	UID       string
	Name      string
	Gender    uint8
	Role      uint8
	Email     string
	CreatedAt time.Time
}

type ChangePasswordDTO struct {
	UID         string
	OldPassword string
	NewPassword string
}
