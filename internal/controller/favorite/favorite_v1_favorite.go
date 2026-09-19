package favorite

import (
	"context"

	"cookbook/api/favorite/v1"
	"cookbook/internal/model"
	"cookbook/internal/service"
)

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	list, err := service.Favorite().List(ctx)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]*model.FavoriteFolder, 0)
	}
	items := make([]model.FavoriteFolder, 0, len(list))
	for _, item := range list {
		items = append(items, *item)
	}
	return &v1.GetListRes{List: items}, nil
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error) {
	id, err := service.Favorite().Create(ctx, req.FavoriteSaveInput)
	if err != nil {
		return nil, err
	}
	return &v1.CreateRes{Id: id}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = service.Favorite().Update(ctx, req.FavoriteIdInput.Id, req.FavoriteSaveInput)
	return nil, err
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteReq) (res *v1.DeleteRes, err error) {
	err = service.Favorite().Delete(ctx, req.FavoriteIdInput.Id)
	return nil, err
}

func (c *ControllerV1) GetRecipes(ctx context.Context, req *v1.GetRecipesReq) (res *v1.GetRecipesRes, err error) {
	list, err := service.Favorite().GetRecipes(ctx, req.FavoriteIdInput.Id)
	if err != nil {
		return nil, err
	}
	if list == nil {
		list = make([]model.RecipeListItem, 0)
	}
	return &v1.GetRecipesRes{List: list}, nil
}

func (c *ControllerV1) GetByRecipe(ctx context.Context, req *v1.GetByRecipeReq) (res *v1.GetByRecipeRes, err error) {
	ids, err := service.Favorite().GetByRecipe(ctx, req.RecipeIdInput.Id)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = make([]int64, 0)
	}
	return &v1.GetByRecipeRes{FavoriteIds: ids}, nil
}

func (c *ControllerV1) AddItem(ctx context.Context, req *v1.AddItemReq) (res *v1.AddItemRes, err error) {
	err = service.Favorite().AddItem(ctx, req.FavoriteIdInput.Id, req.FavoriteAddItemInput)
	return nil, err
}

func (c *ControllerV1) RemoveItem(ctx context.Context, req *v1.RemoveItemReq) (res *v1.RemoveItemRes, err error) {
	err = service.Favorite().RemoveItem(ctx, req.FavoriteIdInput.Id, req.FavoriteRemoveItemInput.RecipeId)
	return nil, err
}