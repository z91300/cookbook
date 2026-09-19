// =================================================================================
// 食谱接口定义
// =================================================================================

package v1

import (
	"cookbook/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// GetListReq 食谱分页列表（关键词 + 标签筛选）
type GetListReq struct {
	g.Meta `path:"/recipes" tags:"Recipe管理" method:"get" summary:"查询食谱列表" operationId:"recipe_getList"`
	model.RecipeListInput
}
type GetListRes struct {
	model.RecipeListOutput
}

// GetOneReq 食谱详情（编辑弹窗数据源）
type GetOneReq struct {
	g.Meta `path:"/recipes/{id}" tags:"Recipe管理" method:"get" summary:"查询食谱详情" operationId:"recipe_getOne"`
	model.RecipeIdInput
}
type GetOneRes struct {
	model.RecipeDetail
}

// UpdateReq 食谱编辑保存
type UpdateReq struct {
	g.Meta `path:"/recipes/{id}" tags:"Recipe管理" method:"put" summary:"编辑食谱" operationId:"recipe_update"`
	model.RecipeUpdateInput
}
type UpdateRes struct{}

// DeleteReq 删除食谱（逻辑删除）
type DeleteReq struct {
	g.Meta `path:"/recipes/{id}" tags:"Recipe管理" method:"delete" summary:"删除食谱" operationId:"recipe_delete"`
	model.RecipeIdInput
}
type DeleteRes struct{}

// CreateReq 新建食谱
type CreateReq struct {
	g.Meta `path:"/recipes" tags:"Recipe管理" method:"post" summary:"新建食谱" operationId:"recipe_create"`
	model.RecipeCreateInput
}
type CreateRes struct {
	Id int64 `json:"id" dc:"新食谱 id"`
}
