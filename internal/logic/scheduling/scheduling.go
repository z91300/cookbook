// =================================================================================
// 食谱编排业务实现：按日期区间查询 / 新增 / 编辑 / 逻辑删除
// =================================================================================

package scheduling

import (
	"context"

	"cookbook/internal/dao"
	"cookbook/internal/model"
	"cookbook/internal/model/do"
	"cookbook/internal/service"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type sScheduling struct{}

func init() {
	service.RegisterScheduling(New())
}

func New() *sScheduling {
	return &sScheduling{}
}

// recipeExists 食谱存在性（未删除）
func (s *sScheduling) recipeExists(ctx context.Context, recipeId int64) (bool, error) {
	count, err := dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().Id, recipeId).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// List 按日期区间查编排，回填食谱卡片字段（标题/封面 URL）；按 日期, 时段, sort, id 升序
func (s *sScheduling) List(ctx context.Context, in model.SchedulingListInput) (out []model.SchedulingItem, err error) {
	out = make([]model.SchedulingItem, 0)
	m := dao.Schedulings.Ctx(ctx).Where(dao.Schedulings.Columns().IsDeleted, 0)
	if in.StartDate > 0 {
		m = m.WhereGTE(dao.Schedulings.Columns().PlanDate, in.StartDate)
	}
	if in.EndDate > 0 {
		m = m.WhereLTE(dao.Schedulings.Columns().PlanDate, in.EndDate)
	}
	err = m.
		Order(dao.Schedulings.Columns().PlanDate).
		Order(dao.Schedulings.Columns().Meal).
		Order(dao.Schedulings.Columns().Sort).
		Order(dao.Schedulings.Columns().Id).
		Scan(&out)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return out, nil
	}

	// 批量回填食谱卡片字段（标题 + 封面 URL），经 service.Recipe().ListByIds 一次取齐
	recipeIds := make([]int64, 0, len(out))
	for _, e := range out {
		recipeIds = append(recipeIds, e.RecipeId)
	}
	byId, err := service.Recipe().ListByIds(ctx, recipeIds)
	if err != nil {
		return nil, err
	}
	for i := range out {
		if recipe, ok := byId[out[i].RecipeId]; ok {
			out[i].RecipeTitle = recipe.Title
			out[i].CoverUrl = recipe.CoverUrl
		}
	}
	return out, nil
}

// Save 新增（id==0）或编辑（id>0）编排条目；recipeId 必须指向未删除食谱
func (s *sScheduling) Save(ctx context.Context, id int64, in model.SchedulingSaveInput) (int64, error) {
	exists, err := s.recipeExists(ctx, in.RecipeId)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, gerror.Newf("食谱不存在或已删除: %d", in.RecipeId)
	}
	now := gtime.Timestamp()
	if id > 0 {
		count, err := dao.Schedulings.Ctx(ctx).
			Where(dao.Schedulings.Columns().Id, id).
			Where(dao.Schedulings.Columns().IsDeleted, 0).
			Count()
		if err != nil {
			return 0, err
		}
		if count == 0 {
			return 0, gerror.Newf("编排条目不存在: %d", id)
		}
		_, err = dao.Schedulings.Ctx(ctx).
			Where(dao.Schedulings.Columns().Id, id).
			Data(do.Schedulings{
				PlanDate:  in.PlanDate,
				Meal:      in.Meal,
				RecipeId:  in.RecipeId,
				Servings:  in.Servings,
				Note:      in.Note,
				Sort:      in.Sort,
				UpdatedAt: now,
			}).
			Update()
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	result, err := dao.Schedulings.Ctx(ctx).
		Data(do.Schedulings{
			PlanDate:  in.PlanDate,
			Meal:      in.Meal,
			RecipeId:  in.RecipeId,
			Servings:  in.Servings,
			Note:      in.Note,
			Sort:      in.Sort,
			CreatedAt: now,
			UpdatedAt: now,
		}).
		InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// Delete 逻辑删除编排条目
func (s *sScheduling) Delete(ctx context.Context, id int64) (err error) {
	count, err := dao.Schedulings.Ctx(ctx).
		Where(dao.Schedulings.Columns().Id, id).
		Where(dao.Schedulings.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.Newf("编排条目不存在: %d", id)
	}
	_, err = dao.Schedulings.Ctx(ctx).
		Where(dao.Schedulings.Columns().Id, id).
		Data(g.Map{
			"is_deleted": 1,
			"updated_at": gtime.Timestamp(),
		}).
		Update()
	return err
}
