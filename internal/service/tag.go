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
	ITag interface {
		// List 查询全部标签（预置 + 后续新增），按 sort 升序
		List(ctx context.Context) (out []*model.TagItem, err error)
		// ManageList 标签管理列表：全部标签 + 使用中菜谱数（LEFT JOIN 聚合，已删除菜谱不计入）
		ManageList(ctx context.Context) (out []*model.TagManageItem, err error)
		// Create 新增标签：查重后追加到末尾（sort = 当前最大 + 1）
		Create(ctx context.Context, in model.TagSaveInput) (int64, error)
		// Update 重命名标签：目标须存在，新名不得与其他标签重复
		Update(ctx context.Context, id int64, in model.TagSaveInput) (err error)
		// Delete 删除标签：事务内先清 recipe_tags 关联（即从菜谱移除该标签，菜谱保留）再删标签；
		// 返回受影响菜谱数（使用中、未删除的）供前端提示
		Delete(ctx context.Context, id int64) (affected int, err error)
	}
)

var (
	localTag ITag
)

func Tag() ITag {
	if localTag == nil {
		panic("implement not found for interface ITag, forgot register?")
	}
	return localTag
}

func RegisterTag(i ITag) {
	localTag = i
}
