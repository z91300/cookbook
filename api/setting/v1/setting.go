// =================================================================================
// 设置接口定义（key-value 通用存储）
// =================================================================================

package v1

import (
	"cookbook/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// GetListReq 设置项列表（全部正常键值对）
type GetListReq struct {
	g.Meta `path:"/settings" tags:"Setting管理" method:"get" summary:"查询设置列表" operationId:"setting_getList"`
}
type GetListRes struct {
	model.SettingListOutput
}

// UpsertReq 批量保存设置（存在则更新 value，不存在则插入）
type UpsertReq struct {
	g.Meta `path:"/settings" tags:"Setting管理" method:"put" summary:"批量保存设置" operationId:"setting_upsert"`
	Items  model.SettingUpsertInputList `json:"items" v:"required|min-length:1#设置项不能为空|至少一条设置项" dc:"设置项列表，整体写入"`
}
type UpsertRes struct{}

// RenderReq 提示词模板渲染：取 key 对应模板，替换 {recipe.*} 变量为指定食谱内容
type RenderReq struct {
	g.Meta `path:"/settings/{key}/render/{recipeId}" tags:"Setting管理" method:"get" summary:"渲染提示词模板" operationId:"setting_render"`
	model.SettingRenderInput
}
type RenderRes struct {
	model.SettingRenderOutput
}
