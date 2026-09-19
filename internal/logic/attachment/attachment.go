// =================================================================================
// 附件业务实现：multipart 上传以 BLOB 入库（attachments.content），
// 内容通过 GET /attachments/{id}/content 二进制直出
// =================================================================================

package attachment

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"strings"

	"cookbook/internal/dao"
	"cookbook/internal/model"
	"cookbook/internal/model/do"
	"cookbook/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type sAttachment struct{}

func init() {
	service.RegisterAttachment(New())
}

func New() *sAttachment {
	return &sAttachment{}
}

// randomName 生成随机文件名（16 字节 hex），保留原始扩展名
func randomName(original string) string {
	buf := make([]byte, 10)
	_, _ = rand.Read(buf)
	ext := strings.ToLower(filepath.Ext(original))
	return hex.EncodeToString(buf) + ext
}

// Upload 读取上传文件，以 BLOB 写入 attachments.content 并登记记录
func (s *sAttachment) Upload(ctx context.Context) (res *model.AttachmentUploadOutput, err error) {
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少上传文件字段 file")
	}
	// 大小限制 20MB
	if file.Size > 20*1024*1024 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "文件不能超过 20MB")
	}

	f, err := file.Open()
	if err != nil {
		return nil, gerror.Wrap(err, "读取上传文件失败")
	}
	defer f.Close()
	content := make([]byte, file.Size)
	if _, err = f.Read(content); err != nil {
		return nil, gerror.Wrap(err, "读取上传文件内容失败")
	}

	// 判断类型
	mimeType := file.Header.Get("Content-Type")
	kind := "file"
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		kind = "image"
	case strings.HasPrefix(mimeType, "video/"):
		kind = "video"
	}

	// 栅格图统一落库为 webp（svg/动图/已是 webp 或转换无收益则原样），降低存储与下游传输
	if kind == "image" {
		if webpContent, webpMime, converted := ToWebpBytes(content, mimeType); converted {
			content = webpContent
			mimeType = webpMime
		}
	}

	// storage_path 仅为唯一标识（blob 化后不再指向磁盘文件），按日期分目录
	storagePath := gtime.Now().Format("Ym") + "/" + randomName(file.Filename)
	now := gtime.Now().Unix()
	insertId, err := dao.Attachments.Ctx(ctx).Data(do.Attachments{
		OwnerUserId: 0, // TODO: 接入登录后取当前用户
		Kind:        kind,
		FileName:    file.Filename,
		StoragePath: storagePath,
		MimeType:    mimeType,
		SizeBytes:   file.Size,
		Content:     content,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}

	return &model.AttachmentUploadOutput{
		AttachmentId: insertId,
		Url:          ContentUrl(insertId),
	}, nil
}

// GetContent 按 id 读取附件二进制内容（未删除才可读）。
// width>0 时按宽度等比缩小（不放大），用于列表卡片缩略图；结果内存缓存。
func (s *sAttachment) GetContent(ctx context.Context, in *model.AttachmentIdInput, width int) (res *model.AttachmentContentOutput, err error) {
	record, err := dao.Attachments.Ctx(ctx).
		Where(dao.Attachments.Columns().Id, in.Id).
		Where(dao.Attachments.Columns().IsDeleted, 0).
		Fields(
			dao.Attachments.Columns().Content,
			dao.Attachments.Columns().MimeType,
			dao.Attachments.Columns().FileName,
		).
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("附件不存在: %d", in.Id)
	}
	content := record[dao.Attachments.Columns().Content].Bytes()
	if len(content) == 0 {
		return nil, gerror.Newf("附件内容缺失: %d", in.Id)
	}
	contentType := record[dao.Attachments.Columns().MimeType].String()
	if width > 0 {
		content, contentType, _ = ResizeBytes(content, contentType, width)
	}
	return &model.AttachmentContentOutput{
		Content:     content,
		ContentType: contentType,
		FileName:    record[dao.Attachments.Columns().FileName].String(),
	}, nil
}

// ContentUrl 附件内容访问 URL（全站统一由此拼装）
func ContentUrl(id int64) string {
	return "/attachments/" + g.NewVar(id).String() + "/content"
}