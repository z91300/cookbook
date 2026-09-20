// =================================================================================
// 附件类型识别与安全白名单
//
// 上传时按「文件魔数」识别真实类型（客户端声明的 Content-Type / 扩展名一律不采信），
// 只有白名单内的栅格图与视频允许以 inline 方式同源直出；其余类型
// （text/html、image/svg+xml、脚本、未知二进制…）一律降级为
// application/octet-stream + Content-Disposition: attachment，
// 避免「上传一个带脚本的 HTML/SVG，再打开同源 URL」造成存储型 XSS。
// =================================================================================

package attachment

import (
	"net/http"
	"path/filepath"
	"strings"
)

// inlineSafeMimes 允许 inline 直出的 MIME 白名单：只放「浏览器不会当脚本执行」的类型。
// 特别注意 image/svg+xml 不在白名单：SVG 内可嵌 <script>，同源 inline 直出即 XSS；
// 同理 text/html、application/xhtml+xml、text/xml 等文本类型也一律不放。
var inlineSafeMimes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/gif":       true,
	"image/webp":      true,
	"image/bmp":       true,
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,
}

// canonicalExt 白名单类型对应的规范扩展名（纠正客户端伪造的扩展名）
var canonicalExt = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"image/bmp":       ".bmp",
	"video/mp4":       ".mp4",
	"video/webm":      ".webm",
	"video/quicktime": ".mov",
}

// InlineSafe 该 MIME 是否允许 inline 直出（白名单判定，服务端唯一判据）
func InlineSafe(mimeType string) bool {
	return inlineSafeMimes[strings.ToLower(mimeType)]
}

// kindOfMime 按 MIME 归类附件用途（attachments.kind）
func kindOfMime(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	}
	return "file"
}

// detectMime 按魔数嗅探真实 MIME，返回不带 charset 等参数的小写类型。
// 嗅探不出来的内容返回 application/octet-stream（会被白名单判定拦下）。
func detectMime(content []byte) string {
	// ISO BMFF（mp4/mov/m4v…）：ftyp box 位于第 4~8 字节，
	// net/http 的嗅探表不覆盖 mp4，这里按 major brand 补上。
	if len(content) >= 12 && string(content[4:8]) == "ftyp" {
		if strings.HasPrefix(strings.ToLower(string(content[8:12])), "qt") {
			return "video/quicktime"
		}
		return "video/mp4"
	}
	ct := http.DetectContentType(content)
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i] // 去掉 "; charset=utf-8" 之类的参数
	}
	return strings.ToLower(strings.TrimSpace(ct))
}

// sanitizeFileName 归一文件名：只留基础名，剔除控制字符/引号/路径分隔符
// （响应头里的 Content-Disposition 不能被文件名注入换行），并限制长度。
func sanitizeFileName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	var b strings.Builder
	for _, r := range name {
		if r < 0x20 || r == 0x7F || r == '"' || r == '\'' || r == '/' || r == '\\' {
			continue
		}
		b.WriteRune(r)
	}
	name = strings.TrimSpace(b.String())
	if name == "" || name == "." || name == ".." {
		return "file"
	}
	if runes := []rune(name); len(runes) > 120 {
		name = string(runes[:120])
	}
	return name
}

// alignFileName 让文件名扩展名与嗅探出的真实类型一致（客户端可伪造扩展名）。
// 非白名单类型不改名：它们只做下载处置，浏览器不会按扩展名执行。
func alignFileName(name, mimeType string) string {
	want, ok := canonicalExt[mimeType]
	if !ok {
		return name
	}
	ext := strings.ToLower(filepath.Ext(name))
	if ext == want || (mimeType == "image/jpeg" && ext == ".jpeg") {
		return name
	}
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if base == "" {
		base = "file"
	}
	return base + want
}
