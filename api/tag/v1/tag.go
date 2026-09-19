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
