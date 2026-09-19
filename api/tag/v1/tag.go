// =================================================================================
// 标签接口定义
// =================================================================================

package v1

import (
	"cookbook/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// GetListReq 标签列表（预置 16 个 + 后续新增）
type GetListReq struct {
	g.Meta `path:"/tags" tags:"Tag管理" method:"get" summary:"查询标签列表" operationId:"tag_getList"`
}
type GetListRes struct {
	List []model.TagItem `json:"list" dc:"标签列表"`
}

// GetManageListReq 标签管理列表（含每个标签的使用中菜谱数量）
type GetManageListReq struct {
	g.Meta `path:"/tags/manage" tags:"Tag管理" method:"get" summary:"查询标签管理列表" operationId:"tag_getManageList"`
}
type GetManageListRes struct {
	List []model.TagManageItem `json:"list" dc:"标签管理列表"`
}

// CreateReq 新增标签（追加到末尾）
type CreateReq struct {
	g.Meta `path:"/tags" tags:"Tag管理" method:"post" summary:"新增标签" operationId:"tag_create"`
	model.TagSaveInput
}
type CreateRes struct {
	Id int64 `json:"id" dc:"新标签 id"`
}

// UpdateReq 重命名标签
type UpdateReq struct {
	g.Meta `path:"/tags/{id}" tags:"Tag管理" method:"put" summary:"重命名标签" operationId:"tag_update"`
	model.TagIdInput
	model.TagSaveInput
}
type UpdateRes struct{}

// DeleteReq 删除标签（从使用中的菜谱移除该标签，不删除菜谱）
type DeleteReq struct {
	g.Meta `path:"/tags/{id}" tags:"Tag管理" method:"delete" summary:"删除标签" operationId:"tag_delete"`
	model.TagIdInput
}
type DeleteRes struct {
	AffectedRecipes int `json:"affectedRecipes" dc:"受影响菜谱数量（被移除该标签的菜谱数）"`
}
