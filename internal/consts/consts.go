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

// 附件上传大小限制。
const (
	// MaxUploadBytes 单个上传文件的大小上限（logic 层校验，超出返回参数错误）
	MaxUploadBytes = 20 * 1024 * 1024

	// MaxRequestBodyBytes HTTP 请求体上限（server 层，见 internal/cmd/cmd.go 的 ClientMaxBodySize）。
	// 必须略大于单文件上限：multipart 的边界、其他表单字段都要占字节，
	// 且 GoFrame 该项默认只有 8MB——不显式放开的话 8~20MB 的文件会在
	// multipart 解析阶段直接报 "request body too large"，让单文件上限形同虚设。
	MaxRequestBodyBytes = MaxUploadBytes + 2*1024*1024
)
