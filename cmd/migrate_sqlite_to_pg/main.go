// 一次性迁移：SQLite(manifest/cookbook.db) → PostgreSQL（链接见 config.yaml 的
// database.default.link；可从环境变量 PG_DSN 覆盖，如 "host=… user=… password=… dbname=…"）。
// 全表搬迁（含 attachments.content 二进制），显式 id 插入后逐表 setval 校准序列。
// 用法：go run ./cmd/migrate_sqlite_to_pg           （默认 dry-run，只统计不写库）
//       go run ./cmd/migrate_sqlite_to_pg --apply  （实际写库）
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/glebarez/go-sqlite"
	_ "github.com/lib/pq"

	"cookbook/internal/logic/attachment"
)

// 表清单与各表列（与 manifest/init.sql 一致）。按依赖顺序：先基础表后关联表。
// tags 有预置行（init.sql 已插入 16 条），迁移时 ON CONFLICT (id) DO NOTHING。
var tables = []struct {
	name  string
	cols  string
	upsert bool // PG 端冲突时跳过（tags 预置行）
}{
	{"users", "id,username,nickname,password_hash,avatar_attachment_id,status,created_at,updated_at", false},
	{"attachments", "id,owner_user_id,kind,file_name,storage_path,mime_type,size_bytes,width,height,duration,sha256,is_deleted,created_at,updated_at,content", false},
	{"recipes", "id,user_id,title,summary,cover_attachment_id,tips,ingredients,tools,steps,meal_mask,calories,difficulty,servings,cook_minutes,source,ai_model,review_status,is_deleted,created_at,updated_at", false},
	{"recipe_tags", "recipe_id,tag_id", false},
	{"members", "id,name,role,note,sort,is_deleted,created_at,updated_at", false},
	{"favorites", "id,user_id,name,description,cover_attachment_id,is_public,sort,is_deleted,created_at,updated_at", false},
	{"favorite_items", "favorite_id,recipe_id,note,created_at", false},
	{"ai_generations", "id,user_id,recipe_id,model,prompt,raw_output,status,error,created_at", false},
	{"schedulings", "id,plan_date,meal,recipe_id,servings,note,sort,is_deleted,created_at,updated_at", false},
	{"settings", "id,key,value,is_deleted,created_at,updated_at", false},
	{"tags", "id,name,sort", true},
}

func main() {
	apply := flag.Bool("apply", false, "实际写库（默认 dry-run 只统计）")
	flag.Parse()

	sqlite, err := sql.Open("sqlite", "manifest/cookbook.db")
	must(err, "open sqlite")
	defer sqlite.Close()

	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "migrate: PG_DSN 环境变量未设置（lib/pq 格式，如 host=… port=… user=… password=… dbname=…）")
		os.Exit(1)
	}
	pg, err := sql.Open("postgres", dsn)
	must(err, "open pg")
	defer pg.Close()

	must(pg.Ping(), "ping pg")
	must(sqlite.Ping(), "ping sqlite")

	total := 0
	start := time.Now()
	for _, t := range tables {
		n := migrateTable(sqlite, pg, t.name, t.cols, t.upsert, *apply)
		fmt.Printf("%-16s %6d rows%s\n", t.name, n, ternary(*apply, "", " (dry-run)"))
		total += n
	}

	// 序列校准：显式 id 插入后 PG identity 序列不前移，需 setval 到 max(id)
	if *apply {
		for _, t := range tables {
			if !hasAutoincrement(t.cols) {
				continue
			}
			_, err := pg.Exec(fmt.Sprintf(
				`SELECT setval(pg_get_serial_sequence('%s','id'), GREATEST((SELECT COALESCE(MAX(id),0) FROM %s), 1))`, t.name, t.name))
			must(err, "setval "+t.name)
		}
		fmt.Println("sequences calibrated")
	}

	fmt.Printf("\ntotal: %d rows in %s%s\n", total, time.Since(start).Round(time.Millisecond), ternary(*apply, "", " (dry-run only; re-run with --apply)"))
}

// migrateTable 迁移单表：SQLite 全量读 → PG 批量写。dry-run 只统计行数。
func migrateTable(sqlite, pg *sql.DB, table, cols string, upsert, apply bool) int {
	rows, err := sqlite.Query(fmt.Sprintf("SELECT %s FROM %s", cols, table))
	must(err, "select "+table)
	defer rows.Close()

	colNames := splitCols(cols)
	tgt := len(colNames)
	placeholders := make([]string, tgt)
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	conflict := "DO NOTHING"
	if !upsert {
		conflict = "DO NOTHING" // 幂等重跑安全：冲突即跳过
	}
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (id) %s", table, cols, join(placeholders, ","), conflict)
	if table == "recipe_tags" || table == "favorite_items" {
		// 复合主键表无 id 列
		stmt = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT DO NOTHING", table, cols, join(placeholders, ","))
	}

	vals := make([]any, tgt)
	ptrs := make([]any, tgt)
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	count, batch := 0, 0
	var tx *sql.Tx
	if apply {
		tx, err = pg.Begin()
		must(err, "begin "+table)
	}
	for rows.Next() {
		must(rows.Scan(ptrs...), "scan "+table)
		batch++
		count++
		if !apply {
			continue
		}
		// BLOB 列可能是 nil（content 可空）→ 转 []byte{} 避免 driver 报错
		for i, v := range vals {
			if v == nil {
				vals[i] = []byte{}
			}
		}
		_, err = tx.Exec(stmt, vals...)
		must(err, "insert "+table)
		// 每 200 行提交一次，避免单事务过大（content BLOB 表尤其）
		if batch >= 200 {
			must(tx.Commit(), "commit "+table)
			batch = 0
			tx, err = pg.Begin()
			must(err, "begin "+table)
		}
	}
	must(rows.Err(), "rows "+table)
	if apply && batch > 0 {
		must(tx.Commit(), "commit "+table)
	}
	return count
}

// hasAutoincrement 判断列里是否有 id（需 setval 校准的表）
func hasAutoincrement(cols string) bool {
	for _, c := range splitCols(cols) {
		if c == "id" {
			return true
		}
	}
	return false
}

func splitCols(cols string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(cols); i++ {
		if i == len(cols) || cols[i] == ',' {
			out = append(out, cols[start:i])
			start = i + 1
		}
	}
	return out
}

func join(items []string, sep string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += sep
		}
		out += s
	}
	return out
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func must(err error, what string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %s: %v\n", what, err)
		os.Exit(1)
	}
}

// 引用 attachment 包的 ToWebpBytes（与 migrate_webp 同款依赖），保持导入链一致
var _ = attachment.ContentUrl