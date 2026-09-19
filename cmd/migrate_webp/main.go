// 一次性迁移：把 attachments 里所有栅格图片 content 重编码为 webp（有收益才替换），
// 同步更新 mime_type 与 file_name 扩展名。svg / 已是 webp / 动图跳过。
// 用法：go run ./cmd/migrate_webp            （默认 dry-run，只打印将发生的变更）
//       go run ./cmd/migrate_webp --apply    （实际写库）
package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/glebarez/go-sqlite"

	"cookbook/internal/logic/attachment"
)

func main() {
	apply := false
	for _, a := range os.Args[1:] {
		if a == "--apply" {
			apply = true
		}
	}

	db, err := sql.Open("sqlite", "manifest/cookbook.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, mime_type, file_name, content FROM attachments WHERE is_deleted=0 AND content IS NOT NULL`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	type update struct {
		id          int64
		content     []byte
		mime        string
		fileName    string
		origSize    int
	}
	var updates []update
	var origTotal, newTotal int

	for rows.Next() {
		var id int64
		var mime, fileName string
		var content []byte
		if err := rows.Scan(&id, &mime, &fileName, &content); err != nil {
			panic(err)
		}
		if !strings.HasPrefix(mime, "image/") {
			continue
		}
		out, outMime, converted := attachment.ToWebpBytes(content, mime)
		if !converted {
			continue
		}
		updates = append(updates, update{id: id, content: out, mime: outMime, fileName: newName(fileName), origSize: len(content)})
		origTotal += len(content)
		newTotal += len(out)
		fmt.Printf("id=%d %s %d B -> webp %d B (%.0f%%)\n", id, fileName, len(content), len(out), 100-float64(len(out))/float64(len(content))*100)
	}

	if len(updates) == 0 {
		fmt.Println("nothing to convert")
		return
	}
	fmt.Printf("\ntotal: %d files, %d B -> %d B, saved %.1f%%\n", len(updates), origTotal, newTotal, 100-float64(newTotal)/float64(origTotal)*100)

	if !apply {
		fmt.Println("dry-run only; re-run with --apply to write")
		return
	}

	for _, u := range updates {
		if _, err := db.Exec(`UPDATE attachments SET content=?, mime_type=?, file_name=?, size_bytes=?, updated_at=strftime('%s','now') WHERE id=?`,
			u.content, u.mime, u.fileName, len(u.content), u.id); err != nil {
			panic(fmt.Errorf("update id=%d: %w", u.id, err))
		}
	}
	fmt.Println("applied", len(updates), "updates")
}

// newName 替换扩展名为 .webp（保留基础名）
func newName(name string) string {
	return strings.TrimSuffix(name, filepath.Ext(name)) + ".webp"
}