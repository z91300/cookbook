// =================================================================================
// 附件领域模型
// =================================================================================

package model

// AttachmentUploadOutput 附件上传响应
type AttachmentUploadOutput struct {
	AttachmentId int64  `json:"attachmentId" dc:"附件 id"`    // 附件 id
	Url          string `json:"url"          dc:"附件访问 URL"` // 附件访问 URL
}

// AttachmentIdInput 附件 id 入参
type AttachmentIdInput struct {
	Id int64 `json:"id" v:"required|min:1#附件id不能为空|附件id不合法" dc:"附件 id"` // 附件 id
}

// AttachmentContentOutput 附件内容读取输出
type AttachmentContentOutput struct {
	Content     []byte // 文件二进制内容
	ContentType string // MIME 类型，用于响应头
	FileName    string // 原始文件名，用于 Content-Disposition
}
