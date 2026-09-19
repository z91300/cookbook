// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"cookbook/internal/model"
)

type (
	IAttachment interface {
		// Upload 读取上传文件，以 BLOB 写入 attachments.content 并登记记录
		Upload(ctx context.Context) (res *model.AttachmentUploadOutput, err error)
		// GetContent 按 id 读取附件二进制内容（未删除才可读）。
		// width>0 时按宽度等比缩小（不放大），用于列表卡片缩略图；结果内存缓存。
		GetContent(ctx context.Context, in *model.AttachmentIdInput, width int) (res *model.AttachmentContentOutput, err error)
	}
)

var (
	localAttachment IAttachment
)

func Attachment() IAttachment {
	if localAttachment == nil {
		panic("implement not found for interface IAttachment, forgot register?")
	}
	return localAttachment
}

func RegisterAttachment(i IAttachment) {
	localAttachment = i
}
