// =================================================================================
// 食谱编排接口定义
// =================================================================================

package v1

import (
	"cookbook/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// GetListReq 编排列表（按日期区间，回填食谱卡片字段）
type GetListReq struct {
	g.Meta `path:"/schedulings" tags:"Scheduling管理" method:"get" summary:"查询食谱编排列表" operationId:"scheduling_getList"`
	model.SchedulingListInput
}
type GetListRes struct {
	List []model.SchedulingItem `json:"list" dc:"编排列表"`
}

// CreateReq 新增编排条目
type CreateReq struct {
	g.Meta `path:"/schedulings" tags:"Scheduling管理" method:"post" summary:"新增食谱编排" operationId:"scheduling_create"`
	model.SchedulingSaveInput
}
type CreateRes struct {
	Id int64 `json:"id" dc:"新编排条目 id"`
}

// UpdateReq 编辑编排条目（全量覆盖）
type UpdateReq struct {
	g.Meta `path:"/schedulings/{id}" tags:"Scheduling管理" method:"put" summary:"编辑食谱编排" operationId:"scheduling_update"`
	model.SchedulingIdInput
	model.SchedulingSaveInput
}
type UpdateRes struct{}

// DeleteReq 删除编排条目（逻辑删除）
type DeleteReq struct {
	g.Meta `path:"/schedulings/{id}" tags:"Scheduling管理" method:"delete" summary:"删除食谱编排" operationId:"scheduling_delete"`
	model.SchedulingIdInput
}
type DeleteRes struct{}
