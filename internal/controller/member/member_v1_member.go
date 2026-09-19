package member

import (
	"context"

	"cookbook/api/member/v1"
	"cookbook/internal/model"
	"cookbook/internal/service"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	list, err := service.Member().List(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*model.MemberItem, 0)
	}
	items := make([]model.MemberItem, 0, len(list))
	for _, item := range list {
		items = append(items, *item)
	}
	return &v1.GetListRes{List: items}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := service.Member().Save(ctx, 0, req.MemberSaveInput)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	_, err = service.Member().Save(ctx, req.MemberIdInput.Id, req.MemberSaveInput)
	return nil, err
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = service.Member().Delete(ctx, req.MemberIdInput.Id)
	return nil, err
}