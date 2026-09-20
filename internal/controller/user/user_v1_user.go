package user

import (
	"context"

	"cookbook/api/user/v1"
	"cookbook/internal/model"
	"cookbook/internal/service"
)

func (c *ControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	out, err := service.Auth().Register(ctx, model.UserRegisterInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		return nil, err
	}
	return &v1.RegisterRes{UserLoginOutput: *out}, nil
}

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	out, err := service.Auth().Login(ctx, model.UserLoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{UserLoginOutput: *out}, nil
}

func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	if err = service.Auth().Logout(ctx); err != nil {
		return nil, err
	}
	return &v1.LogoutRes{}, nil
}

func (c *ControllerV1) GetProfile(ctx context.Context, req *v1.GetProfileReq) (res *v1.GetProfileRes, err error) {
	profile, err := service.Auth().Profile(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetProfileRes{UserProfile: *profile}, nil
}

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetListReq) (res *v1.GetListRes, err error) {
	out, err := service.Auth().List(ctx, model.UserListInput{
		PaginationInput: model.PaginationInput{Page: req.Page, PageSize: req.PageSize},
		Keywords:        req.Keywords,
	})
	if err != nil {
		return nil, err
	}
	return &v1.GetListRes{UserListOutput: *out}, nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateReq) (res *v1.UpdateRes, err error) {
	err = service.Auth().Update(ctx, model.UserUpdateInput{
		UserIdInput: model.UserIdInput{Id: req.Id},
		Nickname:    req.Nickname,
		Status:      req.Status,
		IsAdmin:     req.IsAdmin,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRes{}, nil
}

func (c *ControllerV1) ResetPassword(ctx context.Context, req *v1.ResetPasswordReq) (res *v1.ResetPasswordRes, err error) {
	err = service.Auth().ResetPassword(ctx, model.UserPasswordResetInput{
		UserIdInput: model.UserIdInput{Id: req.Id},
		Password:    req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ResetPasswordRes{}, nil
}
func (c *ControllerV1) Refresh(ctx context.Context, req *v1.RefreshReq) (res *v1.RefreshRes, err error) {
	out, err := service.Auth().Refresh(ctx, model.UserRefreshInput{RefreshToken: req.RefreshToken})
	if err != nil {
		return nil, err
	}
	return &v1.RefreshRes{UserRefreshOutput: *out}, nil
}
