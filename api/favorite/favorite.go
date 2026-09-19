// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package favorite

import (
	"context"

	"cookbook/api/favorite/v1"
)

type IFavoriteV1 interface {
	GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error)
	Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error)
	GetRecipes(ctx context.Context, req *v1.GetRecipesReq) (res *v1.GetRecipesRes, err error)
	GetByRecipe(ctx context.Context, req *v1.GetByRecipeReq) (res *v1.GetByRecipeRes, err error)
	AddItem(ctx context.Context, req *v1.AddItemReq) (res *v1.AddItemRes, err error)
	RemoveItem(ctx context.Context, req *v1.RemoveItemReq) (res *v1.RemoveItemRes, err error)
}
