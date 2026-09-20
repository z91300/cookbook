// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package recipe

import (
	"context"

	"cookbook/api/recipe/v1"
)

type IRecipeV1 interface {
	GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error)
	GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	GetManageList(ctx context.Context, req *v1.GetManageListReq) (res *v1.GetManageListRes, err error)
	Reorder(ctx context.Context, req *v1.ReorderReq) (res *v1.ReorderRes, err error)
}
