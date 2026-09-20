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
		// Upload 读取上传文件，以 BLOB 写入 attachments.content 并登记记录。
		//
		// 安全约定（存储型 XSS 防线，见 mime.go）：
		//   - 类型只认魔数，不认客户端 Content-Type / 扩展名；
		//   - 非白名单类型统一存 application/octet-stream，内容路由按附件下载处置；
		//   - 内容 sha256 落库，同内容直接复用已有附件（去重/秒传）。
		Upload(ctx context.Context) (res *model.AttachmentUploadOutput, err error)
		// GetContent 按 id 读取附件二进制内容（未删除才可读）。
		// width>0 时按宽度等比缩小（不放大），用于列表卡片缩略图；结果内存缓存。
		// 同时给出 Inline 判定：只有白名单类型允许同源 inline 直出，其余必须下载。
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
