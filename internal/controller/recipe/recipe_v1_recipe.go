package recipe

import (
	"context"

	"cookbook/api/recipe/v1"
	"cookbook/internal/service"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	out, err := service.Recipe().List(ctx, req.RecipeListInput)
	if err != nil {
		return nil, err
	}
	return &v1.GetListRes{RecipeListOutput: *out}, nil
}

func (c *ControllerV1) GetOne(ctx context.Context, req *v1.GetOneReq) (res *v1.GetOneRes, err error) {
	detail, err := service.Recipe().GetOne(ctx, req.RecipeIdInput.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetOneRes{RecipeDetail: *detail}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = service.Recipe().Update(ctx, req.RecipeUpdateInput)
	return nil, err
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = service.Recipe().Delete(ctx, req.RecipeIdInput.Id)
	return nil, err
}
func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := service.Recipe().Create(ctx, req.RecipeCreateInput)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}
