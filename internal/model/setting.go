// ==========================================================================
// 设置领域模型 —— 字段唯一定义点
// ==========================================================================

package model

// SettingUpsertInput 设置批量保存项（key=value 对）
type SettingUpsertInput struct {
	Key   string `json:"key"   v:"required|length:1,100#设置键不能为空|设置键最长100" dc:"设置键，如 image_gen_prompt"`
	Value string `json:"value" dc:"设置值（文本，长文本如提示词模板）"`
}

// SettingUpsertInputList 设置批量保存项列表（仅用于 swagger 文档展示）
type SettingUpsertInputList []SettingUpsertInput

// SettingItem 设置项（列表返回形状）
type SettingItem struct {
	Key   string `json:"key"   dc:"设置键"`
	Value string `json:"value" dc:"设置值"`
}

// SettingListOutput 设置列表响应
type SettingListOutput struct {
	Settings []SettingItem `json:"settings" dc:"全部设置项"`
}

// SettingRenderInput 提示词模板渲染入参（指定食谱 id）
type SettingRenderInput struct {
	SettingRenderPathInput
	Key string `json:"key" v:"required#设置键不能为空" dc:"模板设置键，如 image_gen_prompt"`
}

// SettingRenderPathInput 渲染接口路径参数（食谱 id）
type SettingRenderPathInput struct {
	RecipeId int64 `json:"recipeId" v:"required|min:1#食谱id不能为空|食谱id不合法" dc:"食谱 id"`
}

// SettingRenderOutput 渲染结果
type SettingRenderOutput struct {
	Rendered string `json:"rendered" dc:"变量替换后的提示词全文"`
}
