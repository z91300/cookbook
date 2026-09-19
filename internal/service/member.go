// ================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// You can delete these comments if you wish manually maintain this interface file.
// ================================================================================

package service

import (
	"context"
	"cookbook/internal/model"
)

type (
	IMember interface {
		// List 查询全部未删除成员，按 sort,id 升序
		List(ctx context.Context) (out []*model.MemberItem, err error)
		// Save 新增（id==0）或编辑（id>0）成员
		Save(ctx context.Context, id int64, in model.MemberSaveInput) (int64, error)
		// Delete 逻辑删除成员
		Delete(ctx context.Context, id int64) (err error)
	}
)

var (
	localMember IMember
)

func Member() IMember {
	if localMember == nil {
		panic("implement not found for interface IMember, forgot register?")
	}
	return localMember
}

func RegisterMember(i IMember) {
	localMember = i
}
