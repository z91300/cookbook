package attachment

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/url"
	"strings"

	"cookbook/api/attachment/v1"
	"cookbook/internal/model"
	"cookbook/internal/service"

	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error) {
	out, err := service.Attachment().Upload(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.UploadRes{AttachmentId: out.AttachmentId, Url: out.Url}, nil
}

// GetContent 附件内容二进制直出。
// 可选 ?w=<宽度>：等比缩小为缩略图（仅支持 360 档位，不放大）。
// 附件不可变，长缓存 + ETag 协商（命中返回 304）。
//
// 安全约定：只有白名单类型（栅格图/视频）允许 inline 同源直出；
// 其余一律 application/octet-stream + attachment 下载，
// 并统一加 X-Content-Type-Options: nosniff，防止浏览器把上传内容当脚本/HTML 执行。
func (c *ControllerV1) GetContent(ctx context.Context, req *v1.GetContentReq) (res *v1.GetContentRes, err error) {
	out, err := service.Attachment().GetContent(ctx, &model.AttachmentIdInput{Id: req.Id}, req.Width)
	if err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	// ETag：内容指纹（FNV-1a）+ 宽度参与协商；内容或宽度变化都会失效缓存
	h := fnv.New64a()
	h.Write(out.Content) // hash.Write never errors
	etag := fmt.Sprintf(`W/"%d-%d-%016x"`, req.Id, req.Width, h.Sum64())
	r.Response.Header().Set("ETag", etag)
	r.Response.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	r.Response.Header().Set("X-Content-Type-Options", "nosniff")
	if r.Header.Get("If-None-Match") == etag {
		r.Response.WriteHeader(304)
		return nil, nil
	}
	disposition, contentType := "attachment", "application/octet-stream"
	if out.Inline {
		disposition, contentType = "inline", out.ContentType
	}
	r.Response.Header().Set("Content-Type", contentType)
	r.Response.Header().Set("Content-Disposition", contentDisposition(disposition, out.FileName))
	r.Response.Write(out.Content)
	return nil, nil
}

// contentDisposition 拼装 Content-Disposition。
// 文件名同时给出 ASCII 兜底与 RFC 5987 的 filename*（中文名不乱码），
// 并剔除引号/控制字符，避免拼接出畸形或可注入的响应头。
func contentDisposition(disposition, fileName string) string {
	ascii := make([]rune, 0, len(fileName))
	for _, r := range fileName {
		if r < 0x20 || r > 0x7E || r == '"' || r == '\\' {
			continue
		}
		ascii = append(ascii, r)
	}
	name := strings.TrimSpace(string(ascii))
	if name == "" {
		name = "file"
	}
	return fmt.Sprintf("%s; filename=\"%s\"; filename*=UTF-8''%s", disposition, name, url.PathEscape(fileName))
}
