// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package setting

import (
	"context"

	"cookbook/api/setting/v1"
)

type ISettingV1 interface {
	GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error)
	Upsert(ctx context.Context, req *v1.UpsertReq) (res *v1.UpsertRes, err error)
	Render(ctx context.Context, req *v1.RenderReq) (res *v1.RenderRes, err error)
}
