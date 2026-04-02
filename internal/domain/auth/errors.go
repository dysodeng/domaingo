package auth

import tileerrors "github.com/CXeon/tiles/errors"

// 认证领域错误，服务ID = 002，错误码段：20020001 - 20029999
var (
	ErrInvalidCredentials = tileerrors.New(20020001, "用户名或密码错误")
	ErrInvalidToken       = tileerrors.New(20020002, "无效的令牌")
	ErrEmailAlreadyExists = tileerrors.New(20020003, "该邮箱已被注册")
)
