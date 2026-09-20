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

// TagReorderInput 标签拖拽排序入参
// 语义：按传入顺序整体重写 sort（从 1 起递增）。必须是当前全部标签的完整顺序，
// 数量对不上（如已被其他端增删标签）就整体拒绝，避免出现同 sort 的歧义排序
type TagReorderInput struct {
	Ids []int64 `json:"ids" v:"required|length:1,1000#标签顺序不能为空|标签顺序不合法" dc:"拖拽后的标签 id 顺序（当前全部标签的完整顺序，从前往后）"`
}

// TagManageItem 标签管理列表项：标签 + 使用中的菜谱数量（设置页标签管理用）
type TagManageItem struct {
	Id          int64  `json:"id"          dc:"标签 id"`
	Name        string `json:"name"        dc:"标签名"`
	Sort        int    `json:"sort"        dc:"展示排序"`
	RecipeCount int    `json:"recipeCount" dc:"使用该标签的菜谱数量（不含已删除菜谱）"`
}
