package attachment

import (
	"context"
	"fmt"
	"hash/fnv"

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
	if r.Header.Get("If-None-Match") == etag {
		r.Response.WriteHeader(304)
		return nil, nil
	}
	r.Response.Header().Set("Content-Type", out.ContentType)
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", out.FileName))
	r.Response.Write(out.Content)
	return nil, nil
}
