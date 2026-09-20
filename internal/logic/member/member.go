// =================================================================================
// 家庭成员业务实现：列表 / 新增 / 编辑 / 逻辑删除
// =================================================================================

package member

import (
	"context"

	"cookbook/internal/dao"
	"cookbook/internal/logic/auth"
	"cookbook/internal/model"
	"cookbook/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type sMember struct{}

func init() {
	service.RegisterMember(New())
}

func New() *sMember {
	return &sMember{}
}

// List 查询全部未删除成员，按 sort,id 升序
func (s *sMember) List(ctx context.Context) (out []*model.MemberItem, err error) {
	err = dao.Members.Ctx(ctx).
		Where(dao.Members.Columns().IsDeleted, 0).
		Order(dao.Members.Columns().Sort).Order(dao.Members.Columns().Id).
		Scan(&out)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = make([]*model.MemberItem, 0)
	}
	return out, nil
}

// Save 新增（id==0）或编辑（id>0）成员
func (s *sMember) Save(ctx context.Context, id int64, in model.MemberSaveInput) (int64, error) {
	if err := auth.MustLogin(ctx); err != nil {
		return 0, err
	}
	now := gtime.Timestamp()
	if id > 0 {
		count, err := dao.Members.Ctx(ctx).
			Where(dao.Members.Columns().Id, id).
			Where(dao.Members.Columns().IsDeleted, 0).
			Count()
		if err != nil {
			return 0, err
		}
		if count == 0 {
			return 0, gerror.Newf("成员不存在: %d", id)
		}
		_, err = dao.Members.Ctx(ctx).
			Where(dao.Members.Columns().Id, id).
			Data(g.Map{
				"name":       in.Name,
				"role":       in.Role,
				"note":       in.Note,
				"sort":       in.Sort,
				"updated_at": now,
			}).
			Update()
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	result, err := dao.Members.Ctx(ctx).
		Data(g.Map{
			"name":       in.Name,
			"role":       in.Role,
			"note":       in.Note,
			"sort":       in.Sort,
			"created_at": now,
			"updated_at": now,
		}).
		InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Delete 逻辑删除成员
func (s *sMember) Delete(ctx context.Context, id int64) (err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return err
	}
	_, err = dao.Members.Ctx(ctx).
		Where(dao.Members.Columns().Id, id).
		Data(g.Map{
			"is_deleted": 1,
			"updated_at": gtime.Timestamp(),
		}).
		Update()
	return err
}
