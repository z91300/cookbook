// =================================================================================
// 收藏领域模型 —— 字段唯一定义点
// =================================================================================

package model

// FavoriteFolder 收藏夹（列表返回形状）
type FavoriteFolder struct {
	Id          int64  `json:"id"          dc:"收藏夹 id"`
	Name        string `json:"name"        dc:"收藏夹名"`
	Description string `json:"description" dc:"收藏夹描述"`
	RecipeCount int    `json:"recipeCount" dc:"收藏菜谱数"`
	Sort        int    `json:"sort"        dc:"展示排序"`
}

// FavoriteIdInput 收藏夹 id 入参
type FavoriteIdInput struct {
	Id int64 `json:"id" v:"required|min:1#收藏夹id不能为空|收藏夹id不合法" dc:"收藏夹 id"`
}

// FavoriteSaveInput 收藏夹新增/编辑入参
type FavoriteSaveInput struct {
	Name        string `json:"name"        v:"required|length:1,50#收藏夹名不能为空|收藏夹名最长50" dc:"收藏夹名"`
	Description string `json:"description" v:"length:0,200#描述最长200" dc:"收藏夹描述"`
	Sort        int    `json:"sort"        dc:"展示排序"`
}

// FavoriteAddItemInput 收藏食谱入参
type FavoriteAddItemInput struct {
	RecipeId int64  `json:"recipeId" v:"required|min:1#食谱id不能为空|食谱id不合法" dc:"食谱 id"`
	Note     string `json:"note"     v:"length:0,200#收藏备注最长200" dc:"收藏备注，可空"`
}

// FavoriteRemoveItemInput 取消收藏入参
type FavoriteRemoveItemInput struct {
	RecipeId int64 `json:"recipeId" v:"required|min:1#食谱id不能为空|食谱id不合法" dc:"食谱 id"`
}