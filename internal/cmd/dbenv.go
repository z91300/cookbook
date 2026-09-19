// 数据库环境变量装配：容器部署时通过 DB_TYPE / DB_HOST / DB_PORT / DB_USER /
// DB_PASSWORD / DB_NAME 直接注入数据库配置，无需挂载配置文件。
//   - DB_TYPE=postgres（默认）/ postgres 别名 pgsql；DB_TYPE=sqlite 时仅用 DB_PATH
//     （SQLite 文件路径，默认 /data/cookbook.db），其余变量忽略。
//   - 变量齐全时优先级高于镜像内 config.yaml（gdb.SetConfigGroup 先注册，
//     gins 的文件配置装配检测到已配置即跳过）。
//   - 未设 DB_HOST（sqlite 为 DB_DATA_PATH）时不动任何配置，回落 config.yaml。
package cmd

import (
	"os"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
)

// applyDBEnv 在 HTTP 服务启动前调用：环境变量存在则覆写数据库配置。
func applyDBEnv() {
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