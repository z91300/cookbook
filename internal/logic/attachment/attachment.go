// =================================================================================
// 附件业务实现：multipart 上传以 BLOB 入库（attachments.content），
// 内容通过 GET /attachments/{id}/content 二进制直出
// =================================================================================

package attachment

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"path/filepath"
	"strings"

	"cookbook/internal/consts"
	"cookbook/internal/dao"
	"cookbook/internal/logic/auth"
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

// Upload 读取上传文件，以 BLOB 写入 attachments.content 并登记记录。
//
// 安全约定（存储型 XSS 防线，见 mime.go）：
//   - 类型只认魔数，不认客户端 Content-Type / 扩展名；
//   - 非白名单类型统一存 application/octet-stream，内容路由按附件下载处置；
//   - 内容 sha256 落库，同内容直接复用已有附件（去重/秒传）。
func (s *sAttachment) Upload(ctx context.Context) (res *model.AttachmentUploadOutput, err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return nil, err
	}
	r := g.RequestFromCtx(ctx)
	file := r.GetUploadFile("file")
	if file == nil {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "缺少上传文件字段 file")
	}
	if file.Size <= 0 {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "文件内容为空")
	}
	if file.Size > consts.MaxUploadBytes {
		return nil, gerror.NewCode(gcode.CodeInvalidParameter, "文件不能超过 20MB")
	}

	f, err := file.Open()
	if err != nil {
		return nil, gerror.Wrap(err, "读取上传文件失败")
	}
	defer f.Close()
	content := make([]byte, file.Size)
	// 单次 Read 不保证读满（尤其大文件），必须 io.ReadFull
	if _, err = io.ReadFull(f, content); err != nil {
		return nil, gerror.Wrap(err, "读取上传文件内容失败")
	}

	// 类型识别：按魔数嗅探，客户端 Content-Type 与扩展名均可伪造、不采信
	mimeType := detectMime(content)
	// 栅格图统一落库为 webp（svg/动图/已是 webp 或转换无收益则原样），降低存储与下游传输
	if kindOfMime(mimeType) == "image" {
		if webpContent, webpMime, converted := ToWebpBytes(content, mimeType); converted {
			content, mimeType = webpContent, webpMime
		}
	}
	// 白名单判定：白名单内保留真实类型（可 inline 直出），其余一律降级为
	// application/octet-stream + kind=file，由内容路由改为下载处置
	kind := "file"
	if InlineSafe(mimeType) {
		kind = kindOfMime(mimeType)
	} else {
		mimeType = "application/octet-stream"
	}

	// 内容哈希：去重/秒传依据
	sum := sha256.Sum256(content)
	sha := hex.EncodeToString(sum[:])

	// 秒传：同内容（未删除）直接复用已有附件，不重复落库
	existing, err := dao.Attachments.Ctx(ctx).
		Where(dao.Attachments.Columns().Sha256, sha).
		Where(dao.Attachments.Columns().IsDeleted, 0).
		Fields(dao.Attachments.Columns().Id).
		OrderDesc(dao.Attachments.Columns().Id).
		One()
	if err != nil {
		return nil, err
	}
	if !existing.IsEmpty() {
		id := existing[dao.Attachments.Columns().Id].Int64()
		return &model.AttachmentUploadOutput{AttachmentId: id, Url: ContentUrl(id)}, nil
	}

	// 文件名与真实类型对齐（防伪造扩展名），并归一化以免疫响应头注入
	fileName := alignFileName(sanitizeFileName(file.Filename), mimeType)
	// storage_path 仅为唯一标识（blob 化后不再指向磁盘文件），按日期分目录
	storagePath := gtime.Now().Format("Ym") + "/" + randomName(fileName)
	ownerId := int64(0)
	if u := auth.Current(ctx); u != nil {
		ownerId = u.Id
	}
	now := gtime.Now().Unix()
	insertId, err := dao.Attachments.Ctx(ctx).Data(do.Attachments{
		OwnerUserId: ownerId,
		Kind:        kind,
		FileName:    fileName,
		StoragePath: storagePath,
		MimeType:    mimeType,
		SizeBytes:   int64(len(content)), // 落库体的真实字节数（webp 重编码后可能与上传体不同）
		Sha256:      sha,
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
// 同时给出 Inline 判定：只有白名单类型允许同源 inline 直出，其余必须下载。
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
	// 仅图片允许缩放（缩略图）；视频/其他类型原样直出
	if width > 0 && strings.HasPrefix(contentType, "image/") {
		content, contentType, _ = ResizeBytes(content, contentType, width)
	}
	return &model.AttachmentContentOutput{
		Content:     content,
		ContentType: contentType,
		FileName:    record[dao.Attachments.Columns().FileName].String(),
		// 历史数据兜底：库里可能存着客户端声明的 image/svg+xml、text/html 等，
		// 老行同样按白名单判定，非白名单只能下载
		Inline: InlineSafe(contentType),
	}, nil
}

// ContentUrl 附件内容访问 URL（全站统一由此拼装）
func ContentUrl(id int64) string {
	return "/attachments/" + g.NewVar(id).String() + "/content"
}