package user

import tileerrors "github.com/CXeon/tiles/errors"

// 用户领域错误，服务ID = 001，错误码段：20010001 - 20019999
var (
	ErrUserNotFound   = tileerrors.New(20010001, "用户不存在")
	ErrUserSaveFail   = tileerrors.New(20010002, "保存用户失败")
	ErrUserDeleteFail = tileerrors.New(20010003, "删除用户失败")
	ErrUserConflict   = tileerrors.New(20010004, "用户已存在")
)
