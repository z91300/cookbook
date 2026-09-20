// =================================================================================
// 用户与会话业务实现：注册（首个用户即管理员）/ 登录 / 登出 / 令牌刷新 / 当前用户 /
// 管理员侧用户管理（列表、启停、授权、重置密码）+ 鉴权辅助（令牌解析、登录与管理员校验）
//
// 登录态为双令牌（见 manifest/init.sql 的 user_sessions）：
//   · access  短效 2 小时，随每次请求走 Authorization: Bearer <token>
//   · refresh 长效 30 天（滑动续期），仅用于 access 过期后调 /auth/refresh，且每次刷新都轮换
// 令牌明文都不落库，库里只存 sha256。
//
// 权限模型极简（只有 users.is_admin 一个角色位）：
//   · 未登录：只读，任何写操作都被 MustLogin 拒掉
//   · 普通用户：除「标签管理 / 删除菜谱 / 用户管理」外与过去一致（含新建、编辑菜谱）
//   · 管理员：可管理标签、删除菜谱、管理用户
// =================================================================================

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"cookbook/internal/consts"
	"cookbook/internal/dao"
	"cookbook/internal/model"
	"cookbook/internal/model/do"
	"cookbook/internal/service"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
)

const (
	// accessTokenTTL 访问令牌有效期：短效，过期由前端用 refresh 换新
	accessTokenTTL = 2 * time.Hour
	// refreshTokenTTL 刷新令牌有效期：每次成功刷新都重置（滑动续期），长期未使用才失效
	refreshTokenTTL = 30 * 24 * time.Hour
)

type sAuth struct{}

func init() {
	service.RegisterAuth(New())
}

func New() *sAuth {
	return &sAuth{}
}

// ================================================================================
// 包级辅助（供 handler 中间件与其它 logic 包复用）
// ================================================================================

// Resolve 按访问令牌解析当前用户；令牌缺失/过期或用户被禁用时返回 nil
func Resolve(ctx context.Context, token string) *model.AuthUser {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	session, err := dao.UserSessions.Ctx(ctx).
		Where(dao.UserSessions.Columns().AccessTokenHash, hashToken(token)).
		Where("access_expires_at > ?", gtime.Timestamp()).
		One()
	if err != nil || session.IsEmpty() {
		return nil
	}
	userId := session[dao.UserSessions.Columns().UserId].Int64()
	record, err := dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, userId).
		Where(dao.Users.Columns().Status, model.UserStatusEnabled).
		One()
	if err != nil || record.IsEmpty() {
		return nil
	}
	var ent struct {
		Id       int64
		Username string
		Nickname string
		IsAdmin  int
	}
	if err = record.Struct(&ent); err != nil {
		return nil
	}
	return &model.AuthUser{
		Id:        ent.Id,
		Username:  ent.Username,
		Nickname:  ent.Nickname,
		IsAdmin:   ent.IsAdmin == 1,
		SessionId: session[dao.UserSessions.Columns().Id].Int64(),
	}
}

// Current 当前登录用户（未登录返回 nil）
func Current(ctx context.Context) *model.AuthUser {
	u, _ := model.UserFromCtx(ctx)
	return u
}

// tokenExpiredError 登录态过期错误（自定义码，前端据此走刷新流程）
func tokenExpiredError() error {
	return gerror.NewCode(
		gcode.New(consts.CodeTokenExpired, "token expired", nil),
		"登录态已过期",
	)
}

// MustLogin 要求已登录；未登录返回 61，令牌过期返回 4401（前端可自动刷新后重试）
func MustLogin(ctx context.Context) error {
	if Current(ctx) != nil {
		return nil
	}
	if model.TokenInvalidFromCtx(ctx) {
		return tokenExpiredError()
	}
	return gerror.NewCode(gcode.CodeNotAuthorized, "请先登录后再操作")
}

// MustAdmin 要求管理员；未登录/令牌过期同上，非管理员返回 61
func MustAdmin(ctx context.Context) error {
	u := Current(ctx)
	if u == nil {
		return MustLogin(ctx)
	}
	if !u.IsAdmin {
		return gerror.NewCode(gcode.CodeNotAuthorized, "仅管理员可执行该操作")
	}
	return nil
}

// ================================================================================
// 对外业务
// ================================================================================

// Register 注册：用户名唯一；库里还没有可用管理员时，首个注册用户自动成为管理员；注册即登录
func (s *sAuth) Register(ctx context.Context, in model.UserRegisterInput) (out *model.UserLoginOutput, err error) {
	username := strings.TrimSpace(in.Username)
	nickname := strings.TrimSpace(in.Nickname)
	if nickname == "" {
		nickname = username
	}
	var userId int64
	err = dao.Users.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		dup, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Username, username).Count()
		if err != nil {
			return err
		}
		if dup > 0 {
			return gerror.Newf("用户名「%s」已被注册", username)
		}
		// 首个注册用户即管理员：判据是「尚不存在可登录的管理员」，
		// 既满足默认规则，也兼容库中残留的、从未注册过的占位行（password_hash 为空）
		adminCount, err := dao.Users.Ctx(ctx).
			Where(dao.Users.Columns().IsAdmin, 1).
			Where(dao.Users.Columns().Status, model.UserStatusEnabled).
			Where("password_hash <> ''").
			Count()
		if err != nil {
			return err
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		now := gtime.Timestamp()
		userId, err = dao.Users.Ctx(ctx).Data(do.Users{
			Username:     username,
			Nickname:     nickname,
			PasswordHash: string(hash),
			IsAdmin:      boolToInt(adminCount == 0),
			Status:       model.UserStatusEnabled,
			LastLoginAt:  now,
			CreatedAt:    now,
			UpdatedAt:    now,
		}).InsertAndGetId()
		return err
	})
	if err != nil {
		return nil, err
	}
	// 初始化该用户的默认收藏夹「我的收藏」（幂等）：
	// 放在注册这种「初始化时机」，而不是收藏夹 LIST（GET）里做写操作
	if err = service.Favorite().EnsureDefaultFolder(ctx, userId); err != nil {
		return nil, err
	}
	pair, err := s.issueSession(ctx, userId)
	if err != nil {
		return nil, err
	}
	profile, err := profileOf(ctx, userId)
	if err != nil {
		return nil, err
	}
	return &model.UserLoginOutput{TokenPair: *pair, User: *profile}, nil
}

// Login 登录：校验密码与状态，签发双令牌并刷新最近登录时间
func (s *sAuth) Login(ctx context.Context, in model.UserLoginInput) (out *model.UserLoginOutput, err error) {
	username := strings.TrimSpace(in.Username)
	record, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Username, username).One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.New("用户名或密码错误")
	}
	var ent struct {
		Id           int64
		PasswordHash string
		Status       int
	}
	if err = record.Struct(&ent); err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(ent.PasswordHash), []byte(in.Password)) != nil {
		return nil, gerror.New("用户名或密码错误")
	}
	if ent.Status != model.UserStatusEnabled {
		return nil, gerror.New("账号已被禁用，请联系管理员")
	}
	now := gtime.Timestamp()
	if _, err = dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, ent.Id).
		Data(do.Users{LastLoginAt: now, UpdatedAt: now}).
		Update(); err != nil {
		return nil, err
	}
	pair, err := s.issueSession(ctx, ent.Id)
	if err != nil {
		return nil, err
	}
	profile, err := profileOf(ctx, ent.Id)
	if err != nil {
		return nil, err
	}
	return &model.UserLoginOutput{TokenPair: *pair, User: *profile}, nil
}

// Refresh 用 refreshToken 换新令牌对：
// 校验 refresh 未过期且用户正常 → 轮换 refresh 并把有效期重置为 30 天（滑动续期）。
// 旧 refresh 立即作废（同一会话只有一份有效），因此前端必须做「刷新单飞」。
func (s *sAuth) Refresh(ctx context.Context, in model.UserRefreshInput) (out *model.UserRefreshOutput, err error) {
	refreshToken := strings.TrimSpace(in.RefreshToken)
	if refreshToken == "" {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "缺少刷新令牌，请重新登录")
	}
	session, err := dao.UserSessions.Ctx(ctx).
		Where(dao.UserSessions.Columns().RefreshTokenHash, hashToken(refreshToken)).
		Where("expires_at > ?", gtime.Timestamp()).
		One()
	if err != nil {
		return nil, err
	}
	if session.IsEmpty() {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "登录态已失效，请重新登录")
	}
	userId := session[dao.UserSessions.Columns().UserId].Int64()
	status, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, userId).One()
	if err != nil {
		return nil, err
	}
	if status.IsEmpty() || status[dao.Users.Columns().Status].Int() != model.UserStatusEnabled {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "账号已被禁用，请联系管理员")
	}

	accessToken, refreshTokenNew, accessExpireAt, err := newTokenPair()
	if err != nil {
		return nil, err
	}
	now := gtime.Timestamp()
	if _, err = dao.UserSessions.Ctx(ctx).
		Where(dao.UserSessions.Columns().Id, session[dao.UserSessions.Columns().Id].Int64()).
		Data(do.UserSessions{
			AccessTokenHash:  hashToken(accessToken),
			AccessExpiresAt:  accessExpireAt,
			RefreshTokenHash: hashToken(refreshTokenNew),
			ExpiresAt:        now + int64(refreshTokenTTL.Seconds()),
			UpdatedAt:        now,
		}).Update(); err != nil {
		return nil, err
	}
	return &model.UserRefreshOutput{TokenPair: model.TokenPair{
		Token:           accessToken,
		RefreshToken:    refreshTokenNew,
		AccessExpiresAt: accessExpireAt,
	}}, nil
}

// Logout 登出：删除当前会话（access 与 refresh 同时失效）
func (s *sAuth) Logout(ctx context.Context) error {
	u := Current(ctx)
	if u == nil || u.SessionId == 0 {
		return nil
	}
	_, err := dao.UserSessions.Ctx(ctx).Where(dao.UserSessions.Columns().Id, u.SessionId).Delete()
	return err
}

// Profile 当前登录用户信息
func (s *sAuth) Profile(ctx context.Context) (*model.UserProfile, error) {
	u := Current(ctx)
	if u == nil {
		// 未登录 → 61；令牌过期 → 4401（前端据此刷新令牌后重试）
		return nil, MustLogin(ctx)
	}
	profile := u.Profile()
	return &profile, nil
}

// List 管理员：用户列表（关键词匹配登录名/昵称，含各用户创建的菜谱数）
func (s *sAuth) List(ctx context.Context, in model.UserListInput) (out *model.UserListOutput, err error) {
	if err = MustAdmin(ctx); err != nil {
		return nil, err
	}
	m := dao.Users.Ctx(ctx)
	if kw := strings.TrimSpace(in.Keywords); kw != "" {
		like := "%" + kw + "%"
		m = m.Where("(username LIKE ? OR nickname LIKE ?)", like, like)
	}
	var items []*model.UserItem
	var total int
	if err = m.Page(in.Page, in.PageSize).
		OrderDesc(dao.Users.Columns().Id).
		ScanAndCount(&items, &total, true); err != nil {
		return nil, err
	}
	if items == nil {
		items = make([]*model.UserItem, 0)
	}
	if err = s.fillRecipeCount(ctx, items); err != nil {
		return nil, err
	}
	list := make([]model.UserItem, 0, len(items))
	for _, item := range items {
		list = append(list, *item)
	}
	return model.NewPageRes(total, list), nil
}

// Update 管理员：修改用户昵称 / 启停 / 管理员授权
// 保护性约束：不能禁用或取消自己的管理员身份；系统需至少保留一名可用管理员
func (s *sAuth) Update(ctx context.Context, in model.UserUpdateInput) (err error) {
	if err = MustAdmin(ctx); err != nil {
		return err
	}
	me := Current(ctx)
	exists, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, in.Id).Count()
	if err != nil {
		return err
	}
	if exists == 0 {
		return gerror.Newf("用户不存在: %d", in.Id)
	}

	data := do.Users{UpdatedAt: gtime.Timestamp()}
	if in.Nickname != nil {
		data.Nickname = strings.TrimSpace(*in.Nickname)
	}
	if in.Status != nil {
		if *in.Status != model.UserStatusDisabled && *in.Status != model.UserStatusEnabled {
			return gerror.New("状态取值只能是 0（禁用）或 1（正常）")
		}
		if *in.Status == model.UserStatusDisabled {
			if in.Id == me.Id {
				return gerror.New("不能禁用当前登录账号")
			}
			if err = s.ensureAnotherAdmin(ctx, in.Id, "该用户是当前唯一的管理员，不能禁用"); err != nil {
				return err
			}
		}
		data.Status = *in.Status
	}
	if in.IsAdmin != nil {
		if !*in.IsAdmin {
			if in.Id == me.Id {
				return gerror.New("不能取消自己的管理员权限")
			}
			if err = s.ensureAnotherAdmin(ctx, in.Id, "系统需要至少保留一名管理员"); err != nil {
				return err
			}
		}
		data.IsAdmin = boolToInt(*in.IsAdmin)
	}
	if _, err = dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, in.Id).Data(data).Update(); err != nil {
		return err
	}
	// 被禁用 / 降权的用户，登录态立即失效（需要重新登录或以新身份使用）
	if (in.Status != nil && *in.Status == model.UserStatusDisabled) || (in.IsAdmin != nil && !*in.IsAdmin) {
		if _, err = dao.UserSessions.Ctx(ctx).Where(dao.UserSessions.Columns().UserId, in.Id).Delete(); err != nil {
			return err
		}
	}
	return nil
}

// ResetPassword 管理员重置指定用户密码，并清空其全部登录态
func (s *sAuth) ResetPassword(ctx context.Context, in model.UserPasswordResetInput) (err error) {
	if err = MustAdmin(ctx); err != nil {
		return err
	}
	exists, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, in.Id).Count()
	if err != nil {
		return err
	}
	if exists == 0 {
		return gerror.Newf("用户不存在: %d", in.Id)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err = dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().Id, in.Id).
		Data(do.Users{PasswordHash: string(hash), UpdatedAt: gtime.Timestamp()}).
		Update(); err != nil {
		return err
	}
	_, err = dao.UserSessions.Ctx(ctx).Where(dao.UserSessions.Columns().UserId, in.Id).Delete()
	return err
}

// ================================================================================
// 内部实现
// ================================================================================

// issueSession 为某用户新开一个会话（顺带清理已过期会话），返回双令牌对
func (s *sAuth) issueSession(ctx context.Context, userId int64) (out *model.TokenPair, err error) {
	_, _ = dao.UserSessions.Ctx(ctx).Where("expires_at < ?", gtime.Timestamp()).Delete()

	accessToken, refreshToken, accessExpireAt, err := newTokenPair()
	if err != nil {
		return nil, err
	}
	now := gtime.Timestamp()
	if _, err = dao.UserSessions.Ctx(ctx).Data(do.UserSessions{
		UserId:           userId,
		AccessTokenHash:  hashToken(accessToken),
		AccessExpiresAt:  accessExpireAt,
		RefreshTokenHash: hashToken(refreshToken),
		ExpiresAt:        now + int64(refreshTokenTTL.Seconds()),
		CreatedAt:        now,
		UpdatedAt:        now,
	}).Insert(); err != nil {
		return nil, err
	}
	return &model.TokenPair{
		Token:           accessToken,
		RefreshToken:    refreshToken,
		AccessExpiresAt: accessExpireAt,
	}, nil
}

// profileOf 按用户 id 读当前用户信息
func profileOf(ctx context.Context, userId int64) (*model.UserProfile, error) {
	record, err := dao.Users.Ctx(ctx).Where(dao.Users.Columns().Id, userId).One()
	if err != nil {
		return nil, err
	}
	var ent struct {
		Id       int64
		Username string
		Nickname string
		IsAdmin  int
	}
	if err = record.Struct(&ent); err != nil {
		return nil, err
	}
	return &model.UserProfile{
		Id:       ent.Id,
		Username: ent.Username,
		Nickname: ent.Nickname,
		IsAdmin:  ent.IsAdmin == 1,
	}, nil
}

// fillRecipeCount 批量统计各用户创建的菜谱数（不含已删除）
func (s *sAuth) fillRecipeCount(ctx context.Context, items []*model.UserItem) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.Id)
	}
	var rows []struct {
		UserId int64 `json:"user_id"`
		Cnt    int   `json:"cnt"`
	}
	if err := dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		WhereIn(dao.Recipes.Columns().UserId, ids).
		Fields("user_id, COUNT(*) AS cnt").
		Group("user_id").
		Scan(&rows); err != nil {
		return err
	}
	countById := make(map[int64]int, len(rows))
	for _, row := range rows {
		countById[row.UserId] = row.Cnt
	}
	for _, item := range items {
		item.RecipeCount = countById[item.Id]
	}
	return nil
}

// ensureAnotherAdmin 校验除 excludeId 外仍有可用（正常状态）的管理员
func (s *sAuth) ensureAnotherAdmin(ctx context.Context, excludeId int64, msg string) error {
	count, err := dao.Users.Ctx(ctx).
		Where(dao.Users.Columns().IsAdmin, 1).
		Where(dao.Users.Columns().Status, model.UserStatusEnabled).
		WhereNot(dao.Users.Columns().Id, excludeId).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.New(msg)
	}
	return nil
}

// newTokenPair 生成一对令牌：access（2 小时）+ refresh（30 天），返回明文与 access 过期时间
func newTokenPair() (accessToken, refreshToken string, accessExpireAt int64, err error) {
	if accessToken, err = newToken(); err != nil {
		return "", "", 0, err
	}
	if refreshToken, err = newToken(); err != nil {
		return "", "", 0, err
	}
	accessExpireAt = gtime.Timestamp() + int64(accessTokenTTL.Seconds())
	return accessToken, refreshToken, accessExpireAt, nil
}

// newToken 生成 32 字节随机令牌（十六进制 64 位）
func newToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// hashToken 令牌摘要（明文不落库，库中只存 sha256 十六进制）
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
