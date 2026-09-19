package cmd

import (
	"context"

	_ "github.com/gogf/gf/contrib/drivers/pgsql/v2"
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"cookbook/internal/controller/attachment"
	"cookbook/internal/controller/favorite"
	"cookbook/internal/controller/hello"
	"cookbook/internal/controller/member"
	"cookbook/internal/controller/recipe"
	"cookbook/internal/controller/scheduling"
	"cookbook/internal/controller/setting"
	"cookbook/internal/controller/tag"
	"cookbook/internal/handler"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				// 多值查询参数归一（?k=a&k=b → ?k[]=a&k[]=b）须在参数绑定前执行
				group.Middleware(handler.MiddlewareQueryMultiValue, handler.MiddlewareResponse)
				group.Bind(
					hello.NewV1(),
					recipe.NewV1(),
					setting.NewV1(),
					tag.NewV1(),
					attachment.NewV1(),
					favorite.NewV1(),
					member.NewV1(),
					scheduling.NewV1(),
				)
			})
			s.Run()
			return nil
		},
	}
)
