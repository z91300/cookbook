// 容器部署环境变量装配（无配置文件启动）：
//
// 数据库（internal/cmd/dbenv.go 原职责）：
//   DB_TYPE      postgres（默认，别名 pg/pgsql）| sqlite
//   postgres 模式：DB_HOST / DB_PORT(默认5432) / DB_USER(默认root) / DB_PASSWORD / DB_NAME(默认cookbook)
//   sqlite   模式：DB_DATA_PATH（默认 /data/cookbook.db）
//
// HTTP 服务（镜像内无 config.yaml，server 段必须由环境变量兜底，否则 GoFrame
//   回退默认 :0 随机端口且关闭 OpenAPI）：
//   API_PORT     后端监听端口（默认 8000）
//
// 变量齐全时优先级高于任何 config.yaml（gdb/gcfg 先注册，文件配置装配检测到已
// 配置即跳过）。未设 DB_HOST 时数据库回落 config.yaml；API_PORT 恒生效（容器
// 环境本就该由环境变量决定端口）。
package cmd

import (
	"os"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// applyDBEnv 在 HTTP 服务启动前调用。
func applyDBEnv() {
	// server 段：容器内无配置文件，端口必须显式给（镜像 Env 已设 PORT=3000 给前端，
	// 后端用 API_PORT 区分，避免复用同一变量）。
	if port := strings.TrimSpace(os.Getenv("API_PORT")); port != "" {
		g.Server().SetAddr(":" + port)
		g.Server().SetOpenApiPath("/api.json")
		g.Server().SetSwaggerPath("/swagger")
	}

	dbType := strings.ToLower(envOr("DB_TYPE", "postgres"))
	if dbType == "pg" || dbType == "postgres" {
		dbType = "pgsql"
	}

	var node gdb.ConfigNode
	switch dbType {
	case "sqlite":
		path := envOr("DB_DATA_PATH", "/data/cookbook.db")
		node = gdb.ConfigNode{
			Type: "sqlite",
			Name: path, // sqlite 驱动把 Name 当文件路径用
		}
	case "pgsql":
		host := os.Getenv("DB_HOST")
		if host == "" {
			return // 未提供环境变量：回落镜像内 config.yaml
		}
		node = gdb.ConfigNode{
			Type: "pgsql",
			Host: host,
			Port: envOr("DB_PORT", "5432"),
			User: envOr("DB_USER", "root"),
			Pass: os.Getenv("DB_PASSWORD"),
			Name: envOr("DB_NAME", "cookbook"),
		}
	default:
		return // 未识别的 DB_TYPE：回落 config.yaml（启动日志自然暴露配置缺失）
	}

	if err := gdb.SetConfigGroup(gdb.DefaultGroupName, gdb.ConfigGroup{node}); err != nil {
		panic("apply DB env: " + err.Error())
	}
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}