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
	IScheduling interface {
		// List 按日期区间查编排，回填食谱卡片字段（标题/封面 URL）；按 日期, 时段, sort, id 升序
		List(ctx context.Context, in model.SchedulingListInput) (out []model.SchedulingItem, err error)
		// Save 新增（id==0）或编辑（id>0）编排条目；recipeId 必须指向未删除食谱
		Save(ctx context.Context, id int64, in model.SchedulingSaveInput) (int64, error)
		// Delete 逻辑删除编排条目
		Delete(ctx context.Context, id int64) (err error)
	}
)

var (
	localScheduling IScheduling
)

func Scheduling() IScheduling {
	if localScheduling == nil {
		panic("implement not found for interface IScheduling, forgot register?")
	}
	return localScheduling
}

func RegisterScheduling(i IScheduling) {
	localScheduling = i
}
