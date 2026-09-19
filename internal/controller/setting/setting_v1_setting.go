package setting

import (
	"context"

	"cookbook/api/setting/v1"
	"cookbook/internal/model"
	"cookbook/internal/service"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	list, err := service.Setting().List(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]model.SettingItem, 0, len(list))
	for _, item := range list {
		items = append(items, *item)
	}
	return &v1.GetListRes{SettingListOutput: model.SettingListOutput{Settings: items}}, nil
}

func (c *ControllerV1) Upsert(ctx context.Context, req *v1.UpsertReq) (res *v1.UpsertRes, err error) {
	if err = service.Setting().Upsert(ctx, req.Items); err != nil {
		return nil, err
	}
	return &v1.UpsertRes{}, nil
}

func (c *ControllerV1) Render(ctx context.Context, req *v1.RenderReq) (res *v1.RenderRes, err error) {
	rendered, err := service.Setting().Render(ctx, req.Key, req.RecipeId)
	if err != nil {
		return nil, err
	}
	return &v1.RenderRes{SettingRenderOutput: model.SettingRenderOutput{Rendered: rendered}}, nil
}
