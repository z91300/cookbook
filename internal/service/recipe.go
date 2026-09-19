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
	IRecipe interface {
		// List 分页查询食谱，返回卡片展示字段（含标签、封面 URL）
		List(ctx context.Context, in model.RecipeListInput) (out *model.RecipeListOutput, err error)
		// GetOne 食谱详情（编辑弹窗数据源）：解析 JSON 列为结构化列表
		GetOne(ctx context.Context, id int64) (out *model.RecipeDetail, err error)
		// Update 编辑保存：标量列按需更新；ingredients/tools/steps 序列化为 JSON 整体覆盖
		Update(ctx context.Context, in model.RecipeUpdateInput) (err error)
		// Create 新建食谱：插入行（user_id=0, source=manual）后复用 saveRecipeColumns 写列表字段
		Create(ctx context.Context, in model.RecipeCreateInput) (id int64, err error)
		// Delete 逻辑删除食谱（置 is_deleted=1）
		Delete(ctx context.Context, id int64) (err error)
		// ListByIds 按 id 批量取食谱卡片项（保持传入顺序，剔除不存在/已删除）
		ListByIds(ctx context.Context, ids []int64) (out map[int64]model.RecipeListItem, err error)
	}
)

var (
	localRecipe IRecipe
)

func Recipe() IRecipe {
	if localRecipe == nil {
		panic("implement not found for interface IRecipe, forgot register?")
	}
	return localRecipe
}

func RegisterRecipe(i IRecipe) {
	localRecipe = i
}
