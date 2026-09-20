// =================================================================================
// 标签业务实现：全量标签列表 + 设置页标签管理（增/重命名/删）
// =================================================================================

package tag

import (
	"context"
	"strings"
	"time"

	"cookbook/internal/dao"
	"cookbook/internal/logic/auth"
	"cookbook/internal/model"
	"cookbook/internal/model/do"
	"cookbook/internal/service"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
)

// gdb 命名缓存是「单槽位」：Name 非空时缓存 key 只由 Name 决定（不含 SQL），
// 同名查询共享同一份结果，写操作以 Duration=-1 + 同名清理，因此一个名字只能对应一条固定查询。
const tagListCacheName = "tag_list"

var (
	// tagListCacheOption List 查询缓存，TTL 兜底（绕过应用直接改库的场景），应用内写操作会主动清理
	tagListCacheOption = gdb.CacheOption{Duration: time.Hour, Name: tagListCacheName}
	// tagCacheClearOption 写操作后按名清理 List 缓存
	tagCacheClearOption = gdb.CacheOption{Duration: -1, Name: tagListCacheName}
)

type sTag struct{}

func init() {
	service.RegisterTag(New())
}

func New() *sTag {
	return &sTag{}
}

// List 查询全部标签（预置 + 后续新增），按 sort 升序；走 gdb 内存查询缓存，写操作时按名清理
func (s *sTag) List(ctx context.Context) (out []*model.TagItem, err error) {
	err = dao.Tags.Ctx(ctx).
		Order(dao.Tags.Columns().Sort).
		Cache(tagListCacheOption).
		Scan(&out)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = make([]*model.TagItem, 0)
	}
	return out, nil
}

// ManageList 标签管理列表：全部标签 + 使用中菜谱数（LEFT JOIN 聚合，已删除菜谱不计入）
func (s *sTag) ManageList(ctx context.Context) (out []*model.TagManageItem, err error) {
	err = dao.Tags.Ctx(ctx).As("t").
		LeftJoin("recipe_tags rt", "rt.tag_id=t.id").
		LeftJoin("recipes r", "r.id=rt.recipe_id AND r.is_deleted=0").
		Fields("t.id, t.name, t.sort, COUNT(r.id) AS recipe_count").
		Group("t.id").
		Order("t.sort ASC, t.id ASC").
		Scan(&out)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = make([]*model.TagManageItem, 0)
	}
	return out, nil
}

// Create 新增标签：查重后追加到末尾（sort = 当前最大 + 1）；仅管理员可操作
func (s *sTag) Create(ctx context.Context, in model.TagSaveInput) (int64, error) {
	if err := auth.MustAdmin(ctx); err != nil {
		return 0, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return 0, gerror.New("标签名不能为空")
	}
	dup, err := dao.Tags.Ctx(ctx).Where(dao.Tags.Columns().Name, name).Count()
	if err != nil {
		return 0, err
	}
	if dup > 0 {
		return 0, gerror.Newf("标签「%s」已存在", name)
	}
	maxSort, err := dao.Tags.Ctx(ctx).Fields("COALESCE(MAX(sort), 0) AS max_sort").One()
	if err != nil {
		return 0, err
	}
	id, err := dao.Tags.Ctx(ctx).
		Cache(tagCacheClearOption).
		Data(do.Tags{Name: name, Sort: maxSort["max_sort"].Int() + 1}).
		InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update 重命名标签：目标须存在，新名不得与其他标签重复；仅管理员可操作
func (s *sTag) Update(ctx context.Context, id int64, in model.TagSaveInput) (err error) {
	if err = auth.MustAdmin(ctx); err != nil {
		return err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return gerror.New("标签名不能为空")
	}
	exists, err := dao.Tags.Ctx(ctx).Where(dao.Tags.Columns().Id, id).Count()
	if err != nil {
		return err
	}
	if exists == 0 {
		return gerror.Newf("标签不存在: %d", id)
	}
	dup, err := dao.Tags.Ctx(ctx).
		Where(dao.Tags.Columns().Name, name).
		WhereNot(dao.Tags.Columns().Id, id).
		Count()
	if err != nil {
		return err
	}
	if dup > 0 {
		return gerror.Newf("标签「%s」已存在", name)
	}
	_, err = dao.Tags.Ctx(ctx).
		Cache(tagCacheClearOption).
		Where(dao.Tags.Columns().Id, id).
		Data(do.Tags{Name: name}).
		Update()
	return err
}

// Reorder 拖拽排序：按传入顺序把 sort 重写为 1、2、3…（越小越靠前）；仅管理员可操作
// 传入的 id 必须是当前全部标签的完整顺序（不重不漏）：否则会出现两个标签同 sort、
// 排序结果不确定，因此宁可整体拒绝也不写半截顺序
func (s *sTag) Reorder(ctx context.Context, ids []int64) error {
	if err := auth.MustAdmin(ctx); err != nil {
		return err
	}
	if len(ids) == 0 {
		return gerror.New("标签顺序不能为空")
	}
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return gerror.Newf("标签 id 不合法: %d", id)
		}
		if _, ok := seen[id]; ok {
			return gerror.Newf("标签 id 重复: %d", id)
		}
		seen[id] = struct{}{}
	}
	total, err := dao.Tags.Ctx(ctx).Count()
	if err != nil {
		return err
	}
	if total != len(ids) {
		return gerror.Newf("标签已变化（当前 %d 个，提交 %d 个），请刷新后重试", total, len(ids))
	}
	count, err := dao.Tags.Ctx(ctx).WhereIn(dao.Tags.Columns().Id, ids).Count()
	if err != nil {
		return err
	}
	if count != len(ids) {
		return gerror.New("存在不存在的标签 id，排序未生效")
	}
	return dao.Tags.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for i, id := range ids {
			if _, err := dao.Tags.Ctx(ctx).
				Cache(tagCacheClearOption).
				Where(dao.Tags.Columns().Id, id).
				Data(do.Tags{Sort: i + 1}).
				Update(); err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete 删除标签：事务内先清 recipe_tags 关联（即从菜谱移除该标签，菜谱保留）再删标签；
// 返回受影响菜谱数（使用中、未删除的）供前端提示；仅管理员可操作
func (s *sTag) Delete(ctx context.Context, id int64) (affected int, err error) {
	if err = auth.MustAdmin(ctx); err != nil {
		return 0, err
	}
	exists, err := dao.Tags.Ctx(ctx).Where(dao.Tags.Columns().Id, id).Count()
	if err != nil {
		return 0, err
	}
	if exists == 0 {
		return 0, gerror.Newf("标签不存在: %d", id)
	}
	affected, err = dao.RecipeTags.Ctx(ctx).As("rt").
		InnerJoin("recipes r", "r.id=rt.recipe_id AND r.is_deleted=0").
		Where("rt.tag_id", id).
		Count()
	if err != nil {
		return 0, err
	}
	err = dao.Tags.Ctx(ctx).Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.RecipeTags.Ctx(ctx).
			Where(dao.RecipeTags.Columns().TagId, id).
			Delete(); err != nil {
			return err
		}
		_, err := dao.Tags.Ctx(ctx).
			Cache(tagCacheClearOption).
			Where(dao.Tags.Columns().Id, id).
			Delete()
		return err
	})
	if err != nil {
		return 0, err
	}
	return affected, nil
}
