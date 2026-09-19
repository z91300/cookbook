// =================================================================================
// 标签业务实现：全量标签列表（按 sort 排序）
// =================================================================================

package tag

import (
	"context"

	"cookbook/internal/dao"
	"cookbook/internal/model"
	"cookbook/internal/service"
)

type sTag struct{}

func init() {
	service.RegisterTag(New())
}

func New() *sTag {
	return &sTag{}
}

// List 查询全部标签（预置 + 后续新增），按 sort 升序
func (s *sTag) List(ctx context.Context) (out []*model.TagItem, err error) {
	err = dao.Tags.Ctx(ctx).Order(dao.Tags.Columns().Sort).Scan(&out)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = make([]*model.TagItem, 0)
	}
	return out, nil
}
