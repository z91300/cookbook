// =================================================================================
// 附件接口定义
// =================================================================================

package v1

import (
	"cookbook/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// UploadReq 附件上传（multipart/form-data，字段名 file）
type UploadReq struct {
	g.Meta `path:"/attachments" tags:"Attachment管理" method:"post" summary:"上传附件" operationId:"attachment_upload" mime:"multipart/form-data"`
}
type UploadRes struct {
	AttachmentId int64  `json:"attachmentId" dc:"附件 id"`    // 附件 id
	Url          string `json:"url"          dc:"附件访问 URL"` // 附件访问 URL
}

// GetContentReq 附件内容下载（二进制直出，非 JSON 信封）
type GetContentReq struct {
	g.Meta `path:"/attachments/{id}/content" tags:"Attachment管理" method:"get" summary:"获取附件内容" operationId:"attachment_getContent"`
	model.AttachmentIdInput
	Width int `json:"-" p:"w" dc:"缩略图宽度（可选，仅支持 360，不放大）"` // ?w=360 缩略图
}
type GetContentRes struct {
	g.Meta `mime:"application/octet-stream"`
}

