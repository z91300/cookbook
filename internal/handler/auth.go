// =================================================================================
// 鉴权中间件 —— 解析访问令牌（access），把当前登录用户写入请求上下文
// 令牌来源：请求头 Authorization: Bearer <token>（前端统一携带），兼容裸 X-Token 头。
// 说明：本中间件只做「识别身份」，不做拦截——未登录/令牌失效一律按匿名继续，
// 需要登录或管理员权限的接口在 logic 层用 auth.MustLogin / auth.MustAdmin 自行校验。
// 这样浏览类接口保持零门槛（未登录可看），写接口由 logic 层统一把住。
// =================================================================================

package handler

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"

	"cookbook/internal/logic/auth"
	"cookbook/internal/model"
)

// authHeaderPrefix 令牌前缀（标准 Bearer 方案）
const authHeaderPrefix = "bearer "

// MiddlewareAuth 解析访问令牌并注入当前登录用户；只识别身份、不拦截：
// 未带令牌 → 匿名；带了令牌但解析不出来（过期/会话已删/账号被禁用）→ 标记为「登录态失效」，
// 需要登录的接口在 logic 层用 auth.MustLogin/MustAdmin 判断，登录态失效返回 4401 供前端刷新令牌后重试。
func MiddlewareAuth(r *ghttp.Request) {
	token := ExtractToken(r)
	if token != "" {
		if user := auth.Resolve(r.Context(), token); user != nil {
			r.SetCtx(model.ContextWithUser(r.Context(), user))
		} else {
			r.SetCtx(model.ContextMarkTokenInvalid(r.Context()))
		}
	}
	r.Middleware.Next()
}

// ExtractToken 从请求头取登录令牌：优先 Authorization: Bearer <token>，兼容裸 X-Token 头
func ExtractToken(r *ghttp.Request) string {
	if raw := strings.TrimSpace(r.Header.Get("Authorization")); raw != "" {
		if gstr.HasPrefix(gstr.ToLower(raw), authHeaderPrefix) {
			return strings.TrimSpace(raw[len(authHeaderPrefix):])
		}
		return raw
	}
	return strings.TrimSpace(r.Header.Get("X-Token"))
}
