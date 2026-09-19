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
	ISetting interface {
		// List 查询全部正常设置项（按 key 升序）
		List(ctx context.Context) (out []*model.SettingItem, err error)
		// Upsert 批量保存：存在则更新 value 与 updated_at，不存在则插入
		Upsert(ctx context.Context, items model.SettingUpsertInputList) (err error)
		// Render 取 key 对应模板，替换 {recipe.*} 变量为指定食谱内容
		Render(ctx context.Context, key string, recipeId int64) (string, error)
	}
)

var (
	localSetting ISetting
)

func Setting() ISetting {
	if localSetting == nil {
		panic("implement not found for interface ISetting, forgot register?")
	}
	return localSetting
}

func RegisterSetting(i ISetting) {
	localSetting = i
}
