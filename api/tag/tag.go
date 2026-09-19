// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package tag

import (
	"context"

	"cookbook/api/tag/v1"
)

type ITagV1 interface {
	GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error)
}
