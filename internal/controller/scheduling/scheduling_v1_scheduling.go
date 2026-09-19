package scheduling

import (
	"context"

	"cookbook/api/scheduling/v1"
	"cookbook/internal/model"
	"cookbook/internal/service"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	list, err := service.Scheduling().List(ctx, req.SchedulingListInput)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.SchedulingItem, 0)
	}
	return &v1.GetListRes{List: list}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := service.Scheduling().Save(ctx, 0, req.SchedulingSaveInput)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	_, err = service.Scheduling().Save(ctx, req.SchedulingIdInput.Id, req.SchedulingSaveInput)
	return nil, err
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = service.Scheduling().Delete(ctx, req.SchedulingIdInput.Id)
	return nil, err
}
