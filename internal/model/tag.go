// =================================================================================
// 标签领域模型
// =================================================================================

package model

// TagItem 标签项
type TagItem struct {
	Id   int64  `json:"id"   dc:"标签 id"`
	Name string `json:"name" dc:"标签名"`
	Sort int    `json:"sort" dc:"展示排序"`
}

// TagIdInput 标签 id 入参
type TagIdInput struct {
	Id int64 `json:"id" v:"required|min:1#标签id不能为空|标签id不合法" dc:"标签 id"`
}

// TagSaveInput 标签新增/重命名入参
type TagSaveInput struct {
	Name string `json:"name" v:"required|length:1,20#标签名不能为空|标签名最长20" dc:"标签名"`
}

// TagManageItem 标签管理列表项：标签 + 使用中的菜谱数量（设置页标签管理用）
type TagManageItem struct {
	Id          int64  `json:"id"          dc:"标签 id"`
	Name        string `json:"name"        dc:"标签名"`
	Sort        int    `json:"sort"        dc:"展示排序"`
	RecipeCount int    `json:"recipeCount" dc:"使用该标签的菜谱数量（不含已删除菜谱）"`
}
