// =================================================================================
// 用户接口定义：注册 / 登录 / 登出 / 当前用户 + 管理员侧用户管理
// =================================================================================

package v1

import (
	"cookbook/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// RegisterReq 注册（库中无用户时，首个注册者自动成为管理员）
type RegisterReq struct {
	g.Meta `path:"/auth/register" tags:"User管理" method:"post" summary:"注册用户" operationId:"user_register"`
	model.UserRegisterInput
}
type RegisterRes struct {
	model.UserLoginOutput
}

// LoginReq 登录
type LoginReq struct {
	g.Meta `path:"/auth/login" tags:"User管理" method:"post" summary:"用户登录" operationId:"user_login"`
	model.UserLoginInput
}
type LoginRes struct {
	model.UserLoginOutput
}

// LogoutReq 登出（删除当前会话）
type LogoutReq struct {
	g.Meta `path:"/auth/logout" tags:"User管理" method:"post" summary:"退出登录" operationId:"user_logout"`
}
type LogoutRes struct{}

// RefreshReq 用 refreshToken 换新令牌对（access 过期后调用，refresh 一并轮换）
type RefreshReq struct {
	g.Meta `path:"/auth/refresh" tags:"User管理" method:"post" summary:"刷新登录令牌" operationId:"user_refresh"`
	model.UserRefreshInput
}
type RefreshRes struct {
	model.UserRefreshOutput
}

// GetProfileReq 当前登录用户信息
type GetProfileReq struct {
	g.Meta `path:"/auth/profile" tags:"User管理" method:"get" summary:"查询当前用户" operationId:"user_getProfile"`
}
type GetProfileRes struct {
	model.UserProfile
}

// GetListReq 用户列表（仅管理员）
type GetListReq struct {
	g.Meta `path:"/users" tags:"User管理" method:"get" summary:"查询用户列表" operationId:"user_getList"`
	model.UserListInput
}
type GetListRes struct {
	model.UserListOutput
}

// UpdateReq 修改用户（昵称/启停/管理员授权，仅管理员）
type UpdateReq struct {
	g.Meta `path:"/users/{id}" tags:"User管理" method:"put" summary:"修改用户" operationId:"user_update"`
	model.UserUpdateInput
}
type UpdateRes struct{}

// ResetPasswordReq 重置用户密码（仅管理员）
type ResetPasswordReq struct {
	g.Meta `path:"/users/{id}/password" tags:"User管理" method:"put" summary:"重置用户密码" operationId:"user_resetPassword"`
	model.UserPasswordResetInput
}
type ResetPasswordRes struct{}
