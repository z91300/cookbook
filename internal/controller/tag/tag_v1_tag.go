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

func (c *ControllerV1) GetManageList(ctx context.Context, req *v1.GetManageListReq) (res *v1.GetManageListRes, err error) {
	list, err := service.Tag().ManageList(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*model.TagManageItem, 0)
	}
	items := make([]model.TagManageItem, 0, len(list))
	for _, item := range list {
		items = append(items, *item)
	}
	return &v1.GetManageListRes{List: items}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := service.Tag().Create(ctx, model.TagSaveInput{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = service.Tag().Update(ctx, req.Id, model.TagSaveInput{Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}

func (c *ControllerV1) Reorder(ctx context.Context, req *v1.ReorderReq) (res *v1.ReorderRes, err error) {
	if err = service.Tag().Reorder(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.ReorderRes{}, nil
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	affected, err := service.Tag().Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteRes{AffectedRecipes: affected}, nil
}
