// =================================================================================
// 查询参数中间件 —— 同名重复键归一为数组（多选条件）
// GoFrame 的查询解析（gstr.Parse）对同名重复键只保留最后一个值（?k=a&k=b → b），
// 仅识别 k[]=a&k[]=b 的数组写法；而 ofetch/axios 等客户端对数组默认序列化为重复键，
// 会造成「多选」条件静默失效（只命中最后一个取值）。
// 本中间件在业务处理前把重复键转成切片写入查询参数，使 ?k=a&k=b 与 ?k[]=a&k[]=b 等价可用。
// =================================================================================

package handler

import (
	"strings"

	"github.com/gogf/gf/v2/net/ghttp"
)

// MiddlewareQueryMultiValue 把 ?k=a&k=b 归一为数组参数
// 仅在同名多值时注入切片，单值键与带 [] 后缀的写法原样交给框架解析
func MiddlewareQueryMultiValue(r *ghttp.Request) {
	for key, values := range r.URL.Query() {
		if len(values) > 1 && !strings.HasSuffix(key, "[]") {
			r.SetQuery(key, values)
		}
	}
	r.Middleware.Next()
}
