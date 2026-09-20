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
		// Create 新建食谱：插入行（user_id=当前登录用户，未登录为 0=系统；source=manual）后复用 saveRecipeColumns 写列表字段
		Create(ctx context.Context, in model.RecipeCreateInput) (id int64, err error)
		// Delete 逻辑删除食谱（置 is_deleted=1）并同事务清理关联引用；仅管理员可操作。
		// 不做清理的话 recipe_tags / favorite_items 的孤儿行会随删除次数持续累积，
		// 排期行保留（历史编排可追溯）但标记删除，避免继续指向已删食谱。
		Delete(ctx context.Context, id int64) (err error)
		// ManageList 菜谱管理列表（设置页「菜谱管理」表格）：全量非删除菜谱，按手动顺序返回。
		// 不分页是刻意的：手动排序提交的是「全部菜谱的完整顺序」，分页会让顺序残缺。
		// 仅管理员可操作。
		ManageList(ctx context.Context) (out *model.RecipeManageListOutput, err error)
		// Reorder 菜谱手动排序：按传入顺序把 sort 重写为 1、2、3…（越小越靠前）；仅管理员可操作。
		// 传入的 id 必须是当前全部未删除菜谱的完整顺序（不重不漏）：只提交子集会让未参与的菜谱
		// sort 停滞，出现两个菜谱同 sort 的歧义顺序，因此宁可整体拒绝也不写半截顺序。
		Reorder(ctx context.Context, ids []int64) (err error)
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
