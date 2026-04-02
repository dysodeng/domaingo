package auth

import (
	"context"
	"errors"
	"time"

	authdomain "github.com/CXeon/domaingo/internal/domain/auth"
	domainuser "github.com/CXeon/domaingo/internal/domain/user"
	"github.com/CXeon/domaingo/internal/domain/user/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type Service interface {
	Register(ctx context.Context, dto *RegisterDTO) error
	Login(ctx context.Context, dto *LoginDTO) (*TokenPairDTO, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
	Refresh(ctx context.Context, refreshToken string) (*TokenPairDTO, error)
	Introspect(ctx context.Context, accessToken string) (*IntrospectDTO, error)
	Me(ctx context.Context, uid string) (*MeDTO, error)
	ChangePassword(ctx context.Context, dto *ChangePasswordDTO) error
}

type service struct {
	userRepo   domainuser.Repository
	tokenStore authdomain.TokenStore
	cfg        Config
}

func NewService(userRepo domainuser.Repository, tokenStore authdomain.TokenStore, cfg Config) Service {
	return &service{
		userRepo:   userRepo,
		tokenStore: tokenStore,
		cfg:        cfg,
	}
}

// --- JWT Claims ---

type accessClaims struct {
	UID  string `json:"uid"`
	Role uint8  `json:"role"`
	jwt.RegisteredClaims
}

type refreshClaims struct {
	UID string `json:"uid"`
	jwt.RegisteredClaims
}

// --- Service methods ---

func (s *service) Register(ctx context.Context, dto *RegisterDTO) error {
	if _, err := s.userRepo.FindByEmail(ctx, dto.Email); err == nil {
		return authdomain.ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u := &model.User{
		UID:      uuid.New(),
		Name:     dto.Name,
		Gender:   model.Gender(dto.Gender),
		Role:     model.RoleUser,
		Email:    dto.Email,
		Password: string(hash),
	}
	if err := u.Check(); err != nil {
		return err
	}
	_, err = s.userRepo.Create(ctx, u)
	return err
}

func (s *service) Login(ctx context.Context, dto *LoginDTO) (*TokenPairDTO, error) {
	u, err := s.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		return nil, authdomain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(dto.Password)); err != nil {
		return nil, authdomain.ErrInvalidCredentials
	}
	return s.issueTokenPair(ctx, u)
}

func (s *service) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// Block the access token if still valid
	if ac, err := s.parseAccessToken(accessToken); err == nil {
		_ = s.tokenStore.BlockAccessToken(ctx, ac.ID, ac.ExpiresAt.Time)
	}
	// Delete the refresh token regardless of access token validity
	if rc, err := s.parseRefreshToken(refreshToken); err == nil {
		_ = s.tokenStore.DeleteRefreshToken(ctx, rc.ID)
	}
	return nil
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*TokenPairDTO, error) {
	rc, err := s.parseRefreshToken(refreshToken)
	if err != nil {
		return nil, authdomain.ErrInvalidToken
	}

	uid, ok := s.tokenStore.RefreshTokenUID(ctx, rc.ID)
	if !ok {
		return nil, authdomain.ErrInvalidToken
	}

	// Rotation: delete old refresh token before issuing new pair
	_ = s.tokenStore.DeleteRefreshToken(ctx, rc.ID)

	u, err := s.userRepo.FindByID(ctx, uid)
	if err != nil {
		return nil, authdomain.ErrInvalidToken
	}
	return s.issueTokenPair(ctx, u)
}

func (s *service) Introspect(ctx context.Context, accessToken string) (*IntrospectDTO, error) {
	ac, err := s.parseAccessToken(accessToken)
	if err != nil {
		return nil, authdomain.ErrInvalidToken
	}
	if s.tokenStore.IsAccessTokenBlocked(ctx, ac.ID) {
		return nil, authdomain.ErrInvalidToken
	}
	return &IntrospectDTO{
		UID:  ac.UID,
		Role: ac.Role,
	}, nil
}

func (s *service) Me(ctx context.Context, uid string) (*MeDTO, error) {
	id, err := uuid.Parse(uid)
	if err != nil {
		return nil, authdomain.ErrInvalidToken
	}
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &MeDTO{
		UID:       u.UID.String(),
		Name:      u.Name,
		Gender:    uint8(u.Gender),
		Role:      uint8(u.Role),
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}, nil
}

func (s *service) ChangePassword(ctx context.Context, dto *ChangePasswordDTO) error {
	id, err := uuid.Parse(dto.UID)
	if err != nil {
		return authdomain.ErrInvalidToken
	}
	u, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return authdomain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(dto.OldPassword)); err != nil {
		return authdomain.ErrInvalidCredentials
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return s.userRepo.Update(ctx, u)
}

// --- helpers ---

func (s *service) issueTokenPair(ctx context.Context, u *model.User) (*TokenPairDTO, error) {
	now := time.Now()
	accessJTI := uuid.NewString()
	refreshJTI := uuid.NewString()

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &accessClaims{
		UID:  u.UID.String(),
		Role: uint8(u.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessJTI,
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}).SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, err
	}

	refreshExpiry := now.Add(s.cfg.RefreshTTL)
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &refreshClaims{
		UID: u.UID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshJTI,
			ExpiresAt: jwt.NewNumericDate(refreshExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}).SignedString([]byte(s.cfg.Secret))
	if err != nil {
		return nil, err
	}

	if err := s.tokenStore.SaveRefreshToken(ctx, refreshJTI, u.UID, refreshExpiry); err != nil {
		return nil, err
	}

	return &TokenPairDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) parseAccessToken(tokenStr string) (*accessClaims, error) {
	claims := &accessClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, authdomain.ErrInvalidToken
	}
	return claims, nil
}

func (s *service) parseRefreshToken(tokenStr string) (*refreshClaims, error) {
	claims := &refreshClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, s.keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, authdomain.ErrInvalidToken
	}
	return claims, nil
}

func (s *service) keyFunc(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, errors.New("unexpected signing method")
	}
	return []byte(s.cfg.Secret), nil
}
