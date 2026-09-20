package cmd

import (
	"context"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"cookbook/internal/consts"
	"cookbook/internal/controller/attachment"
	"cookbook/internal/controller/favorite"
	"cookbook/internal/controller/hello"
	"cookbook/internal/controller/member"
	"cookbook/internal/controller/recipe"
	"cookbook/internal/controller/scheduling"
	"cookbook/internal/controller/setting"
	"cookbook/internal/controller/tag"
	"cookbook/internal/controller/user"
	"cookbook/internal/handler"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			// 容器部署：DB_* 环境变量存在则覆写数据库配置（优先于 config.yaml）
			applyDBEnv()
			s := g.Server()
			// HTTP 请求体上限：GoFrame 默认仅 8MB，会让「单文件 20MB」的上传限制形同虚设
			// （8~20MB 的请求在 multipart 解析阶段就 500）。与 consts.MaxUploadBytes 配套。
			s.SetClientMaxBodySize(consts.MaxRequestBodyBytes)
			s.Group("/", func(group *ghttp.RouterGroup) {
				// 顺序：① 多值查询参数归一（须在参数绑定前）② 鉴权（解析令牌写入当前用户）
				// ③ 统一响应信封（须在业务后、鉴权后，保证错误也走信封）
				group.Middleware(
					handler.MiddlewareQueryMultiValue,
					handler.MiddlewareAuth,
					handler.MiddlewareResponse,
				)
				group.Bind(
					hello.NewV1(),
					recipe.NewV1(),
					setting.NewV1(),
					tag.NewV1(),
					attachment.NewV1(),
					favorite.NewV1(),
					member.NewV1(),
					scheduling.NewV1(),
					user.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
