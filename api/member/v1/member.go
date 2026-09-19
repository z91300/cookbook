// =================================================================================
// 家庭成员接口定义
// =================================================================================

package v1

import (
	"cookbook/internal/model"

	"github.com/gogf/gf/v2/frame/g"
)

// GetListReq 成员列表
type GetListReq struct {
	g.Meta `path:"/members" tags:"Member管理" method:"get" summary:"查询家庭成员" operationId:"member_getList"`
}
type GetListRes struct {
	List []model.MemberItem `json:"list" dc:"成员列表"`
}

// CreateReq 新增成员
type CreateReq struct {
	g.Meta `path:"/members" tags:"Member管理" method:"post" summary:"新增成员" operationId:"member_create"`
	model.MemberSaveInput
}
type CreateRes struct {
	Id int64 `json:"id" dc:"新成员 id"`
}

// UpdateReq 编辑成员
type UpdateReq struct {
	g.Meta `path:"/members/{id}" tags:"Member管理" method:"put" summary:"编辑成员" operationId:"member_update"`
	model.MemberIdInput
	model.MemberSaveInput
}
type UpdateRes struct{}

// DeleteReq 删除成员（逻辑删除）
type DeleteReq struct {
	g.Meta `path:"/members/{id}" tags:"Member管理" method:"delete" summary:"删除成员" operationId:"member_delete"`
	model.MemberIdInput
}
type DeleteRes struct{}