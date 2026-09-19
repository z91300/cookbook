// =================================================================================
// 统一响应模型 —— HTTP 出口信封，全部接口共用
// 配套输出中间件见 internal/handler.MiddlewareResponse
// =================================================================================

package model

// StandardRes 统一响应体
// code=0 成功；非 0 失败（业务错误码由 gerror/gcode 携带，无码错误输出 50）
type StandardRes struct {
	Code    int    `json:"code"    dc:"0 成功，非 0 失败"`     // 0 成功，非 0 失败
	Message string `json:"message" dc:"成功固定 OK，失败为错误信息"` // 成功固定 OK，失败为错误信息
	Data    any    `json:"data"    dc:"业务数据"`            // 业务数据
}
