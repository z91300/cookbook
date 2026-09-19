package tag

import (
	"context"

	"cookbook/api/tag/v1"
	"cookbook/internal/model"
	"cookbook/internal/service"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	list, err := service.Tag().List(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*model.TagItem, 0)
	}
	items := make([]model.TagItem, 0, len(list))
	for _, item := range list {
		items = append(items, *item)
	}
	return &v1.GetListRes{List: items}, nil
}
