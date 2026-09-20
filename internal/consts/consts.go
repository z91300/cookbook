package consts

// 业务错误码：GoFrame gcode 标准码之外的自定义码，前端按数值识别。
//
// 标准码（gcode）中与本项目相关的：0=成功、50=一般错误、51=参数校验失败、61=未认证/无权限。
const (
	// CodeTokenExpired 访问令牌（access）已过期/失效
	// 前端收到该码应带 refreshToken 调 /auth/refresh 换新令牌对后重试原请求；
	// 刷新也失败则视为登录态结束，清除本地令牌。
	CodeTokenExpired = 4401
)
