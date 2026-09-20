// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"cookbook/internal/model"
)

type (
	IAuth interface {
		// Register 注册：用户名唯一；库里还没有可用管理员时，首个注册用户自动成为管理员；注册即登录
		Register(ctx context.Context, in model.UserRegisterInput) (out *model.UserLoginOutput, err error)
		// Login 登录：校验密码与状态，签发双令牌并刷新最近登录时间
		Login(ctx context.Context, in model.UserLoginInput) (out *model.UserLoginOutput, err error)
		// Refresh 用 refreshToken 换新令牌对：
		// 校验 refresh 未过期且用户正常 → 轮换 refresh 并把有效期重置为 30 天（滑动续期）。
		// 旧 refresh 立即作废（同一会话只有一份有效），因此前端必须做「刷新单飞」。
		Refresh(ctx context.Context, in model.UserRefreshInput) (out *model.UserRefreshOutput, err error)
		// Logout 登出：删除当前会话（access 与 refresh 同时失效）
		Logout(ctx context.Context) error
		// Profile 当前登录用户信息
		Profile(ctx context.Context) (*model.UserProfile, error)
		// List 管理员：用户列表（关键词匹配登录名/昵称，含各用户创建的菜谱数）
		List(ctx context.Context, in model.UserListInput) (out *model.UserListOutput, err error)
		// Update 管理员：修改用户昵称 / 启停 / 管理员授权
		// 保护性约束：不能禁用或取消自己的管理员身份；系统需至少保留一名可用管理员
		Update(ctx context.Context, in model.UserUpdateInput) (err error)
		// ResetPassword 管理员重置指定用户密码，并清空其全部登录态
		ResetPassword(ctx context.Context, in model.UserPasswordResetInput) (err error)
	}
)

var (
	localAuth IAuth
)

func Auth() IAuth {
	if localAuth == nil {
		panic("implement not found for interface IAuth, forgot register?")
	}
	return localAuth
}

func RegisterAuth(i IAuth) {
	localAuth = i
}
