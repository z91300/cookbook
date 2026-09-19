// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package attachment

import (
	"context"

	"cookbook/api/attachment/v1"
)

type IAttachmentV1 interface {
	Upload(ctx context.Context, req *v1.UploadReq) (res *v1.UploadRes, err error)
	GetContent(ctx context.Context, req *v1.GetContentReq) (res *v1.GetContentRes, err error)
}
