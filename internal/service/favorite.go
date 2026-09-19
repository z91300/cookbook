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
	IFavorite interface {
		// List 收藏夹列表；无任何夹时自动创建默认夹「我的收藏」；
		// recipeCount 用一条 GROUP BY 聚合填充
		List(ctx context.Context) (out []*model.FavoriteFolder, err error)
		// Create 新建收藏夹
		Create(ctx context.Context, in model.FavoriteSaveInput) (int64, error)
		// Update 编辑收藏夹
		Update(ctx context.Context, id int64, in model.FavoriteSaveInput) (err error)
		// Delete 删除收藏夹（逻辑删）并硬删夹内条目；食谱本体不动
		Delete(ctx context.Context, id int64) (err error)
		// GetRecipes 收藏夹内食谱，按收藏时间倒序；经 recipe.ListByIds 组装卡片字段
		GetRecipes(ctx context.Context, favoriteId int64) (out []model.RecipeListItem, err error)
		// GetByRecipe 查询食谱被收藏在哪些（未删除的）收藏夹
		GetByRecipe(ctx context.Context, recipeId int64) (out []int64, err error)
		// AddItem 收藏食谱到夹：已存在行则更新 note，否则插入
		AddItem(ctx context.Context, favoriteId int64, in model.FavoriteAddItemInput) (err error)
		// RemoveItem 从收藏夹取消收藏（硬删条目行）
		RemoveItem(ctx context.Context, favoriteId int64, recipeId int64) (err error)
	}
)

var (
	localFavorite IFavorite
)

func Favorite() IFavorite {
	if localFavorite == nil {
		panic("implement not found for interface IFavorite, forgot register?")
	}
	return localFavorite
}

func RegisterFavorite(i IFavorite) {
	localFavorite = i
}
