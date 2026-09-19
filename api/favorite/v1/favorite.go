// =================================================================================
// 收藏接口定义
// =================================================================================

package v1

import (
	"cookbook/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// GetListReq 收藏夹列表（无夹时后端自动建默认夹）
type GetListReq struct {
	g.Meta `path:"/favorites" tags:"Favorite管理" method:"get" summary:"查询收藏夹列表" operationId:"favorite_getList"`
}
type GetListRes struct {
	List []model.FavoriteFolder `json:"list" dc:"收藏夹列表"`
}

// CreateReq 新建收藏夹
type CreateReq struct {
	g.Meta `path:"/favorites" tags:"Favorite管理" method:"post" summary:"新建收藏夹" operationId:"favorite_create"`
	model.FavoriteSaveInput
}
type CreateRes struct {
	Id int64 `json:"id" dc:"新收藏夹 id"`
}

// UpdateReq 编辑收藏夹
type UpdateReq struct {
	g.Meta `path:"/favorites/{id}" tags:"Favorite管理" method:"put" summary:"编辑收藏夹" operationId:"favorite_update"`
	model.FavoriteIdInput
	model.FavoriteSaveInput
}
type UpdateRes struct{}

// DeleteReq 删除收藏夹（夹内收藏条目一并清除）
type DeleteReq struct {
	g.Meta `path:"/favorites/{id}" tags:"Favorite管理" method:"delete" summary:"删除收藏夹" operationId:"favorite_delete"`
	model.FavoriteIdInput
}
type DeleteRes struct{}

// GetRecipesReq 收藏夹内食谱列表（按收藏时间倒序，不分页）
type GetRecipesReq struct {
	g.Meta `path:"/favorites/{id}/recipes" tags:"Favorite管理" method:"get" summary:"收藏夹内食谱列表" operationId:"favorite_getRecipes"`
	model.FavoriteIdInput
}
type GetRecipesRes struct {
	List []model.RecipeListItem `json:"list" dc:"收藏的食谱列表"`
}

// GetByRecipeReq 查询食谱被收藏在哪些收藏夹
type GetByRecipeReq struct {
	g.Meta `path:"/recipes/{id}/favorites" tags:"Favorite管理" method:"get" summary:"查询食谱被收藏的收藏夹" operationId:"favorite_getByRecipe"`
	model.RecipeIdInput
}
type GetByRecipeRes struct {
	FavoriteIds []int64 `json:"favoriteIds" dc:"收藏夹 id 列表"`
}

// AddItemReq 收藏食谱到收藏夹
type AddItemReq struct {
	g.Meta `path:"/favorites/{id}/items" tags:"Favorite管理" method:"post" summary:"收藏食谱" operationId:"favorite_addItem"`
	model.FavoriteIdInput
	model.FavoriteAddItemInput
}
type AddItemRes struct{}

// RemoveItemReq 从收藏夹取消收藏
type RemoveItemReq struct {
	g.Meta `path:"/favorites/{id}/items/{recipeId}" tags:"Favorite管理" method:"delete" summary:"取消收藏" operationId:"favorite_removeItem"`
	model.FavoriteIdInput
	model.FavoriteRemoveItemInput
}
type RemoveItemRes struct{}