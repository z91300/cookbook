// =================================================================================
// 用户领域模型 —— 字段唯一定义点
// 权限极简：只有 is_admin 一个角色位（首个注册用户自动为管理员），不做角色/权限表
// =================================================================================

package model

import "context"

// 用户状态（users.status）
const (
	UserStatusDisabled = 0 // 禁用
	UserStatusEnabled  = 1 // 正常
)

// UserProfile 当前登录用户信息（前端顶栏展示 + 权限判断）
type UserProfile struct {
	Id       int64  `json:"id"       dc:"用户 id"`
	Username string `json:"username" dc:"登录名"`
	Nickname string `json:"nickname" dc:"昵称"`
	IsAdmin  bool   `json:"isAdmin"  dc:"是否管理员"`
}

// UserRegisterInput 注册入参
type UserRegisterInput struct {
	Username string `json:"username" v:"required|length:2,20#用户名不能为空|用户名长度 2-20" dc:"登录名"`
	Password string `json:"password" v:"required|length:6,32#密码不能为空|密码长度 6-32" dc:"密码（明文，仅本次传输使用，落库为 bcrypt 哈希）"`
	Nickname string `json:"nickname" v:"length:0,20#昵称最长20" dc:"昵称，留空取用户名"`
}

// UserLoginInput 登录入参
type UserLoginInput struct {
	Username string `json:"username" v:"required#用户名不能为空" dc:"登录名"`
	Password string `json:"password" v:"required#密码不能为空" dc:"密码"`
}

// TokenPair 双令牌对
// access（短效）用于每次请求；refresh（长效）只用于 access 过期后换新令牌对，两者明文都不落库
type TokenPair struct {
	Token           string `json:"token"           dc:"访问令牌(access)，2 小时有效；请求头 Authorization: Bearer <token>"`
	RefreshToken    string `json:"refreshToken"    dc:"刷新令牌(refresh)，30 天有效（滑动续期）；access 过期后调 /auth/refresh 换新"`
	AccessExpiresAt int64  `json:"accessExpiresAt" dc:"访问令牌过期时间 Unix 秒"`
}

// UserLoginOutput 登录 / 注册成功返回
type UserLoginOutput struct {
	TokenPair
	User UserProfile `json:"user" dc:"当前用户信息"`
}

// UserRefreshInput 刷新令牌入参
type UserRefreshInput struct {
	RefreshToken string `json:"refreshToken" v:"required#刷新令牌不能为空" dc:"登录或上次刷新返回的 refreshToken"`
}

// UserRefreshOutput 刷新令牌返回（refresh 一并轮换，旧的立即作废）
type UserRefreshOutput struct {
	TokenPair
}

// UserIdInput 用户 id 入参
type UserIdInput struct {
	Id int64 `json:"id" v:"required|min:1#用户id不能为空|用户id不合法" dc:"用户 id"`
}

// UserUpdateInput 管理员修改用户入参（指针字段：nil=不修改，便于单独切换状态/角色）
type UserUpdateInput struct {
	UserIdInput
	Nickname *string `json:"nickname" dc:"昵称"`
	Status   *int    `json:"status"   dc:"状态：0=禁用 1=正常（禁用会清空该用户登录态）"`
	IsAdmin  *bool   `json:"isAdmin"  dc:"是否管理员"`
}

// UserPasswordResetInput 管理员重置密码入参
type UserPasswordResetInput struct {
	UserIdInput
	Password string `json:"password" v:"required|length:6,32#密码不能为空|密码长度 6-32" dc:"新密码"`
}

// UserItem 用户管理列表项
type UserItem struct {
	Id          int64  `json:"id"          dc:"用户 id"`
	Username    string `json:"username"    dc:"登录名"`
	Nickname    string `json:"nickname"    dc:"昵称"`
	IsAdmin     bool   `json:"isAdmin"     dc:"是否管理员"`
	Status      int    `json:"status"      dc:"状态：0=禁用 1=正常"`
	RecipeCount int    `json:"recipeCount" dc:"创建的菜谱数量（不含已删除）"`
	LastLoginAt int64  `json:"lastLoginAt" dc:"最近登录时间 Unix 秒，0=从未登录"`
	CreatedAt   int64  `json:"createdAt"   dc:"注册时间 Unix 秒"`
}

// UserListInput 用户管理列表查询入参
type UserListInput struct {
	PaginationInput
	Keywords string `json:"keywords" dc:"关键词，匹配登录名/昵称" v:"length:0,50#关键词最长50"`
}

// UserListOutput 用户管理列表响应
type UserListOutput = PageRes[UserItem]

// --------------------------------------------------------------------------------
// 当前登录用户（请求上下文载体，非接口出入参）
// 由 internal/handler.MiddlewareAuth 解析令牌后写入，logic 层用 model.UserFromCtx 读取
// --------------------------------------------------------------------------------

// AuthUser 鉴权通过后的当前登录用户
type AuthUser struct {
	Id        int64  // 用户 id
	Username  string // 登录名
	Nickname  string // 昵称
	IsAdmin   bool   // 是否管理员
	SessionId int64  // 当前会话 user_sessions.id（登出用）
}

// Profile 转成对外用户信息
func (u *AuthUser) Profile() UserProfile {
	return UserProfile{Id: u.Id, Username: u.Username, Nickname: u.Nickname, IsAdmin: u.IsAdmin}
}

// authUserCtxKey 上下文键（私有类型，避免与其它包的键冲突）
type authUserCtxKey struct{}

// ContextWithUser 把当前登录用户写入上下文
func ContextWithUser(ctx context.Context, u *AuthUser) context.Context {
	if u == nil {
		return ctx
	}
	return context.WithValue(ctx, authUserCtxKey{}, u)
}

// UserFromCtx 读取上下文中的当前登录用户（未登录返回 false）
func UserFromCtx(ctx context.Context) (*AuthUser, bool) {
	u, ok := ctx.Value(authUserCtxKey{}).(*AuthUser)
	return u, ok && u != nil
}

// authInvalidCtxKey 「带了令牌但已失效」上下文的键
type authInvalidCtxKey struct{}

// ContextMarkTokenInvalid 标记「请求带了登录令牌但已过期/失效」
// 用于把「未登录」与「登录态过期」区分开：后者前端应走 refresh 换新令牌后重试
func ContextMarkTokenInvalid(ctx context.Context) context.Context {
	return context.WithValue(ctx, authInvalidCtxKey{}, true)
}

// TokenInvalidFromCtx 请求是否带了失效的登录令牌
func TokenInvalidFromCtx(ctx context.Context) bool {
	invalid, _ := ctx.Value(authInvalidCtxKey{}).(bool)
	return invalid
}
