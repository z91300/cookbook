// =================================================================================
// 标签业务实现：全量标签列表 + 设置页标签管理（增/重命名/删）
// =================================================================================

package tag

import (
	"context"
	"strings"
	"time"

	"cookbook/internal/dao"
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

// Create 新增标签：查重后追加到末尾（sort = 当前最大 + 1）
func (s *sTag) Create(ctx context.Context, in model.TagSaveInput) (int64, error) {
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

// Update 重命名标签：目标须存在，新名不得与其他标签重复
func (s *sTag) Update(ctx context.Context, id int64, in model.TagSaveInput) (err error) {
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

// Delete 删除标签：事务内先清 recipe_tags 关联（即从菜谱移除该标签，菜谱保留）再删标签；
// 返回受影响菜谱数（使用中、未删除的）供前端提示
func (s *sTag) Delete(ctx context.Context, id int64) (affected int, err error) {
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
