// =================================================================================
// HTTP 中间件 —— 统一响应输出（信封格式见 internal/model/response.go）
// 行为对齐 ghttp.MiddlewareHandlerResponse，仅信封结构收敛到自有 model
// =================================================================================

package handler

import (
	"mime"
	"net/http"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"

	"cookbook/internal/model"
)

// streamTypes 流式响应类型，不包统一信封
var streamTypes = []string{
	"text/event-stream",
	"application/octet-stream",
	"multipart/x-mixed-replace",
}

// MiddlewareResponse 统一响应中间件
// 成功：{"code":0,"message":"OK","data":{...}}
// 失败：{"code":非0,"message":"错误信息","data":null}
// 业务错误码由 gerror/gcode 携带（如 51 校验失败）；无码错误输出 50
// 已有自定义输出（导出/文件下载）与流式响应不包信封，直接透传
func MiddlewareResponse(r *ghttp.Request) {
	r.Middleware.Next()

	// 已有自定义输出，直接透传
	if r.Response.BufferLength() > 0 || r.Response.BytesWritten() > 0 {
		return
	}
	// 流式响应，不包信封
	mediaType, _, _ := mime.ParseMediaType(r.Response.Header().Get("Content-Type"))
	for _, ct := range streamTypes {
		if mediaType == ct {
			return
		}
	}

	var (
		err  = r.GetError()
		data = r.GetHandlerResponse()
		code = gerror.Code(err)
	)
	if err != nil {
		if code == gcode.CodeNil {
			code = gcode.CodeInternalError
		}
	} else if r.Response.Status > 0 && r.Response.Status != http.StatusOK {
		switch r.Response.Status {
		case http.StatusNotFound:
			code = gcode.CodeNotFound
		case http.StatusForbidden:
			code = gcode.CodeNotAuthorized
		default:
			code = gcode.CodeUnknown
		}
		// 记录错误，供后续中间件/日志使用
		r.SetError(gerror.NewCode(code))
	} else {
		code = gcode.CodeOK
	}

	msg := code.Message()
	if err != nil {
		msg = err.Error()
	}
	r.Response.WriteJson(model.StandardRes{
		Code:    code.Code(),
		Message: msg,
		Data:    data,
	})
}
