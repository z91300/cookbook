// =================================================================================
// 食谱业务实现：分页列表 / 详情 / 编辑保存
// =================================================================================
package recipe

import (
	"context"
	"encoding/json"
	"strings"

	"cookbook/internal/dao"
	"cookbook/internal/logic/attachment"
	"cookbook/internal/logic/auth"
	"cookbook/internal/model"
	"cookbook/internal/model/do"
	"cookbook/internal/service"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type sRecipe struct{}

func init() {
	service.RegisterRecipe(New())
}

func New() *sRecipe {
	return &sRecipe{}
}

// List 分页查询食谱，返回卡片展示字段（含标签、封面 URL）
func (s *sRecipe) List(ctx context.Context, in model.RecipeListInput) (out *model.RecipeListOutput, err error) {
	m := dao.Recipes.Ctx(ctx).Where(dao.Recipes.Columns().IsDeleted, 0)

	// 关键词：标题/简介模糊匹配（OR 分组加括号，避免与 is_deleted 等叠加 WHERE 时优先级错乱）
	if in.Keywords != "" {
		kw := "%" + in.Keywords + "%"
		m = m.Where("(title LIKE ? OR summary LIKE ?)", kw, kw)
	}
	// 多选标签：AND 叠加，菜谱需同时含所有选中标签
	for _, tagId := range in.TagIds {
		m = m.Where("id IN (SELECT recipe_id FROM recipe_tags WHERE tag_id=?)", tagId)
	}
	// 用餐时间：位与筛选；meal_mask=0 视为"全部时段"，任意时段筛选都命中
	if in.Meal > 0 {
		m = m.Where("(meal_mask = 0 OR meal_mask & ? > 0)", in.Meal)
	}

	out = &model.RecipeListOutput{}
	var entities []*model.RecipeListItem
	// Fields 白名单：卡片墙用不到 ingredients/tools/steps 三个 JSON 大列，
	// 不指定就会 SELECT * 把它们整列拉出来（每行可达数十 KB 的无效 IO）
	// 排序：sort ASC（手动顺序，越小越靠前）+ id DESC；
	// sort=0 视为「未参与排序」，天然排在手动序列之前 —— 也就是新菜谱仍置顶
	err = m.Fields(
		dao.Recipes.Columns().Id,
		dao.Recipes.Columns().Title,
		dao.Recipes.Columns().Summary,
		dao.Recipes.Columns().CoverAttachmentId,
		dao.Recipes.Columns().Calories,
		dao.Recipes.Columns().Difficulty,
		dao.Recipes.Columns().CookMinutes,
	).Page(in.Page, in.PageSize).
		OrderAsc(dao.Recipes.Columns().Sort).
		OrderDesc(dao.Recipes.Columns().Id).
		ScanAndCount(&entities, &out.Total, true)
	if err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		out.List = make([]model.RecipeListItem, 0)
		return out, nil
	}

	// 组装封面 URL（批量一次查附件表）
	coverIds := make([]int64, 0, len(entities))
	for _, e := range entities {
		if e.CoverAttachmentId > 0 {
			coverIds = append(coverIds, e.CoverAttachmentId)
		}
	}
	urlById, err := s.attachmentUrlMap(ctx, coverIds)
	if err != nil {
		return nil, err
	}

	// 组装标签（批量两次查：关联表 + 标签表）
	recipeIds := make([]int64, 0, len(entities))
	for _, e := range entities {
		recipeIds = append(recipeIds, e.Id)
	}
	tagsMap, err := s.recipeTagMap(ctx, recipeIds)
	if err != nil {
		return nil, err
	}
	favCntMap, err := s.favoriteCountMap(ctx, recipeIds)
	if err != nil {
		return nil, err
	}

	out.List = make([]model.RecipeListItem, 0, len(entities))
	for _, e := range entities {
		e.CoverUrl = urlById[e.CoverAttachmentId]
		if tags, ok := tagsMap[e.Id]; ok {
			e.Tags = tags
		} else {
			e.Tags = make([]model.TagItem, 0)
		}
		e.FavoriteCount = favCntMap[e.Id]
		out.List = append(out.List, *e)
	}
	return out, nil
}

// favoriteCountMap 按食谱 id 批量统计被收藏次数（收藏夹数，未删除的收藏夹）
func (s *sRecipe) favoriteCountMap(ctx context.Context, recipeIds []int64) (map[int64]int, error) {
	out := make(map[int64]int, len(recipeIds))
	if len(recipeIds) == 0 {
		return out, nil
	}
	type countRow struct {
		RecipeId int64 `json:"recipe_id"`
		Cnt      int   `json:"cnt"`
	}
	var rows []countRow
	err := dao.FavoriteItems.Ctx(ctx).
		InnerJoin("favorites", "favorites.id=favorite_items.favorite_id").
		WhereIn("favorite_items.recipe_id", recipeIds).
		Where("favorites.is_deleted", 0).
		Fields("favorite_items.recipe_id, COUNT(*) AS cnt").
		Group("favorite_items.recipe_id").
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.RecipeId] = r.Cnt
	}
	return out, nil
}

// attachmentUrlMap 按 id 批量取附件访问 URL（blob 内容路由）
func (s *sRecipe) attachmentUrlMap(ctx context.Context, ids []int64) (map[int64]string, error) {
	result := make(map[int64]string, len(ids))
	if len(ids) == 0 {
		return result, nil
	}
	records, err := dao.Attachments.Ctx(ctx).
		WhereIn(dao.Attachments.Columns().Id, ids).
		Where(dao.Attachments.Columns().IsDeleted, 0).
		Where(dao.Attachments.Columns().Content + " IS NOT NULL").
		Fields(dao.Attachments.Columns().Id).
		All()
	if err != nil {
		return nil, err
	}
	for _, rec := range records {
		id := rec[dao.Attachments.Columns().Id].Int64()
		result[id] = attachment.ContentUrl(id)
	}
	return result, nil
}

// recipeTagMap 按食谱 id 批量取标签（按 sort 排序）
func (s *sRecipe) recipeTagMap(ctx context.Context, recipeIds []int64) (map[int64][]model.TagItem, error) {
	result := make(map[int64][]model.TagItem, len(recipeIds))
	if len(recipeIds) == 0 {
		return result, nil
	}
	links, err := dao.RecipeTags.Ctx(ctx).
		WhereIn(dao.RecipeTags.Columns().RecipeId, recipeIds).
		All()
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		return result, nil
	}
	tagIds := make([]int64, 0, len(links))
	for _, link := range links {
		tagIds = append(tagIds, link[dao.RecipeTags.Columns().TagId].Int64())
	}
	tagRecords, err := dao.Tags.Ctx(ctx).WhereIn(dao.Tags.Columns().Id, tagIds).Order(dao.Tags.Columns().Sort).All()
	if err != nil {
		return nil, err
	}
	tagById := make(map[int64]model.TagItem, len(tagRecords))
	for _, rec := range tagRecords {
		var item model.TagItem
		if err = rec.Struct(&item); err != nil {
			return nil, err
		}
		tagById[item.Id] = item
	}
	for _, link := range links {
		recipeId := link[dao.RecipeTags.Columns().RecipeId].Int64()
		tagId := link[dao.RecipeTags.Columns().TagId].Int64()
		if tag, ok := tagById[tagId]; ok {
			result[recipeId] = append(result[recipeId], tag)
		}
	}
	return result, nil
}

// GetOne 食谱详情（编辑弹窗数据源）：解析 JSON 列为结构化列表
func (s *sRecipe) GetOne(ctx context.Context, id int64) (out *model.RecipeDetail, err error) {
	record, err := dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().Id, id).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, gerror.Newf("食谱不存在或已删除: %d", id)
	}
	out = &model.RecipeDetail{}
	if err = record.Struct(out); err != nil {
		return nil, err
	}

	// 封面 URL
	if out.CoverAttachmentId > 0 {
		urlById, err := s.attachmentUrlMap(ctx, []int64{out.CoverAttachmentId})
		if err != nil {
			return nil, err
		}
		out.CoverUrl = urlById[out.CoverAttachmentId]
	}

	// JSON 列解析（空串视为空列表）
	if err = json.Unmarshal([]byte(record["ingredients"].String()), &out.Ingredients); err != nil {
		out.Ingredients = make([]model.RecipeIngredientInput, 0)
	}
	if err = json.Unmarshal([]byte(record["tools"].String()), &out.Tools); err != nil {
		out.Tools = make([]string, 0)
	}
	if err = json.Unmarshal([]byte(record["steps"].String()), &out.Steps); err != nil {
		out.Steps = make([]model.RecipeStepItem, 0)
	}
	if out.Tags == nil {
		out.Tags = make([]model.TagItem, 0)
	}

	// 标签
	tagsMap, err := s.recipeTagMap(ctx, []int64{id})
	if err != nil {
		return nil, err
	}
	if tags, ok := tagsMap[id]; ok {
		out.Tags = tags
	}
	return out, nil
}

// Update 编辑保存：标量列按需更新；ingredients/tools/steps 序列化为 JSON 整体覆盖
func (s *sRecipe) Update(ctx context.Context, in model.RecipeUpdateInput) (err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return err
	}
	// 存在性校验
	count, err := dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().Id, in.Id).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.Newf("食谱不存在或已删除: %d", in.Id)
	}

	// 步骤内容非空校验（动态表单可增减，但不允许空行落库）
	for i, step := range in.Steps {
		if step.Content == "" {
			return gerror.Newf("第 %d 步内容不能为空", i+1)
		}
	}
	for i, ing := range in.Ingredients {
		if ing.Name == "" {
			return gerror.Newf("第 %d 个食材名不能为空", i+1)
		}
	}
	// 标题不允许被清空（传了就必须非空）
	if in.Title != nil && strings.TrimSpace(*in.Title) == "" {
		return gerror.New("标题不能为空")
	}

	// 标量列按需更新（指针非 nil 才写，传空值=清空）；JSON 列与标签整体覆盖
	if err = s.saveRecipeColumns(ctx, in.Id, saveRecipeColumnsInput{
		Title:             in.Title,
		Summary:           in.Summary,
		Tips:              in.Tips,
		CoverAttachmentId: in.CoverAttachmentId,
		MealMask:          in.MealMask,
		Difficulty:        in.Difficulty,
		CookMinutes:       in.CookMinutes,
		Calories:          in.Calories,
		Ingredients:       in.Ingredients,
		Tools:             in.Tools,
		Steps:             in.Steps,
		TagIds:            in.TagIds,
	}); err != nil {
		return err
	}
	return nil
}

// saveRecipeColumnsInput 标量列 + JSON 列 + 标签的共享写入集（Create 与 Update 复用）。
// 标量列一律指针语义：nil=本次不修改，非 nil=显式写入（允许写成空串/0 以清空字段）；
// JSON 列 nil=不修改、非 nil（含空切片）=整体覆盖；TagIds nil=不修改、非 nil=整体重建。
type saveRecipeColumnsInput struct {
	Title             *string
	Summary           *string
	Tips              *string
	CoverAttachmentId *int64
	MealMask          *int
	Difficulty        *int
	CookMinutes       *int
	Calories          *int
	Ingredients       []model.RecipeIngredientInput
	Tools             []string
	Steps             []model.RecipeStepItem
	TagIds            []int
}

// saveRecipeColumns 保存 recipes 标量列与 JSON 列，并按 TagIds 重建标签关联。
// 整个过程在同一个事务里：UPDATE recipes + DELETE recipe_tags + 批量 INSERT 要么全成要么全不成，
// 避免中途失败留下「标签删了没插回来」的半更新状态。
func (s *sRecipe) saveRecipeColumns(ctx context.Context, recipeId int64, in saveRecipeColumnsInput) (err error) {
	// 标签存在性校验放在事务外：传了不存在的 tag id 会建出悬空关联（列表侧静默丢弃但脏数据留库）
	tagIds, err := s.normalizeTagIds(ctx, in.TagIds)
	if err != nil {
		return err
	}

	data := do.Recipes{}
	if in.Title != nil {
		data.Title = *in.Title
	}
	if in.Summary != nil {
		data.Summary = *in.Summary
	}
	if in.Tips != nil {
		data.Tips = *in.Tips
	}
	if in.CoverAttachmentId != nil {
		data.CoverAttachmentId = *in.CoverAttachmentId
	}
	if in.MealMask != nil {
		data.MealMask = *in.MealMask
	}
	if in.Difficulty != nil {
		data.Difficulty = *in.Difficulty
	}
	if in.CookMinutes != nil {
		data.CookMinutes = *in.CookMinutes
	}
	if in.Calories != nil {
		data.Calories = *in.Calories
	}
	if in.Ingredients != nil {
		b, err := json.Marshal(in.Ingredients)
		if err != nil {
			return err
		}
		data.Ingredients = string(b)
	}
	if in.Tools != nil {
		b, err := json.Marshal(in.Tools)
		if err != nil {
			return err
		}
		data.Tools = string(b)
	}
	if in.Steps != nil {
		b, err := json.Marshal(in.Steps)
		if err != nil {
			return err
		}
		data.Steps = string(b)
	}
	// 编辑即刷新更新时间（列表按 id 倒序时看不出差别，但「最近修改」排序依赖它）
	data.UpdatedAt = gtime.Timestamp()

	return dao.Recipes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.Recipes.Ctx(ctx).
			Where(dao.Recipes.Columns().Id, recipeId).
			Data(data).
			Update(); err != nil {
			return err
		}

		// 标签整体覆盖：nil=不修改，非 nil=重建关联（含空列表=清空）；批量一次插入
		if in.TagIds == nil {
			return nil
		}
		if _, err := dao.RecipeTags.Ctx(ctx).
			Where(dao.RecipeTags.Columns().RecipeId, recipeId).
			Delete(); err != nil {
			return err
		}
		if len(tagIds) == 0 {
			return nil
		}
		rows := make(g.List, 0, len(tagIds))
		for _, tagId := range tagIds {
			rows = append(rows, g.Map{
				dao.RecipeTags.Columns().RecipeId: recipeId,
				dao.RecipeTags.Columns().TagId:    tagId,
			})
		}
		_, err := dao.RecipeTags.Ctx(ctx).Data(rows).Insert()
		return err
	})
}

// normalizeTagIds 标签 id 去重并校验存在性；nil 原样返回（=不修改标签）
func (s *sRecipe) normalizeTagIds(ctx context.Context, ids []int) ([]int, error) {
	if ids == nil {
		return nil, nil
	}
	// 去重：tagIds 重复会让复合主键 (recipe_id, tag_id) 冲突
	seen := make(map[int]struct{}, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, gerror.Newf("标签 id 不合法: %d", id)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	if len(out) == 0 {
		return out, nil
	}
	count, err := dao.Tags.Ctx(ctx).
		WhereIn(dao.Tags.Columns().Id, out).
		Count()
	if err != nil {
		return nil, err
	}
	if count != len(out) {
		return nil, gerror.Newf("存在不存在的标签 id（提交 %d 个，命中 %d 个），请刷新后重试", len(out), count)
	}
	return out, nil
}

// Create 新建食谱：插入行（user_id=当前登录用户，未登录为 0=系统；source=manual）后复用 saveRecipeColumns 写列表字段
func (s *sRecipe) Create(ctx context.Context, in model.RecipeCreateInput) (id int64, err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return 0, err
	}
	// 步骤/食材内容非空校验（与 Update 同规则）
	for i, step := range in.Steps {
		if step.Content == "" {
			return 0, gerror.Newf("第 %d 步内容不能为空", i+1)
		}
	}
	for i, ing := range in.Ingredients {
		if ing.Name == "" {
			return 0, gerror.Newf("第 %d 个食材名不能为空", i+1)
		}
	}
	// 记录作者（未登录时保持 0=系统），便于后续按用户统计
	authorId := int64(0)
	if u := auth.Current(ctx); u != nil {
		authorId = u.Id
	}
	now := gtime.Timestamp()
	id, err = dao.Recipes.Ctx(ctx).
		Data(g.Map{
			"user_id":    authorId,
			"title":      in.Title,
			"summary":    in.Summary,
			"tips":       in.Tips,
			"meal_mask":  in.MealMask,
			"source":     "manual",
			"created_at": now,
			"updated_at": now,
		}).
		InsertAndGetId()
	if err != nil {
		return 0, err
	}
	// 新建时所有标量列都显式写入（含零值/空串），与 Update 的「不传不改」语义区分
	err = s.saveRecipeColumns(ctx, id, saveRecipeColumnsInput{
		Title:             &in.Title,
		Summary:           &in.Summary,
		Tips:              &in.Tips,
		CoverAttachmentId: &in.CoverAttachmentId,
		MealMask:          &in.MealMask,
		Difficulty:        &in.Difficulty,
		CookMinutes:       &in.CookMinutes,
		Calories:          &in.Calories,
		Ingredients:       in.Ingredients,
		Tools:             in.Tools,
		Steps:             in.Steps,
		TagIds:            in.TagIds,
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Delete 逻辑删除食谱（置 is_deleted=1）并同事务清理关联引用；仅管理员可操作。
// 不做清理的话 recipe_tags / favorite_items 的孤儿行会随删除次数持续累积，
// 排期行保留（历史编排可追溯）但标记删除，避免继续指向已删食谱。
func (s *sRecipe) Delete(ctx context.Context, id int64) (err error) {
	if err = auth.MustAdmin(ctx); err != nil {
		return err
	}
	count, err := dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().Id, id).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.Newf("食谱不存在或已删除: %d", id)
	}
	now := gtime.Timestamp()
	return dao.Recipes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := dao.Recipes.Ctx(ctx).
			Where(dao.Recipes.Columns().Id, id).
			Data(do.Recipes{IsDeleted: 1, UpdatedAt: now}).
			Update(); err != nil {
			return err
		}
		// 标签关联、收藏条目无独立语义，直接硬删（留着只会是垃圾行）
		if _, err := dao.RecipeTags.Ctx(ctx).
			Where(dao.RecipeTags.Columns().RecipeId, id).
			Delete(); err != nil {
			return err
		}
		if _, err := dao.FavoriteItems.Ctx(ctx).
			Where(dao.FavoriteItems.Columns().RecipeId, id).
			Delete(); err != nil {
			return err
		}
		_, err := dao.Schedulings.Ctx(ctx).
			Where(dao.Schedulings.Columns().RecipeId, id).
			Where(dao.Schedulings.Columns().IsDeleted, 0).
			Data(do.Schedulings{IsDeleted: 1, UpdatedAt: now}).
			Update()
		return err
	})
}

// ManageList 菜谱管理列表（设置页「菜谱管理」表格）：全量非删除菜谱，按手动顺序返回。
// 不分页是刻意的：手动排序提交的是「全部菜谱的完整顺序」，分页会让顺序残缺。
// 仅管理员可操作。
func (s *sRecipe) ManageList(ctx context.Context) (out *model.RecipeManageListOutput, err error) {
	if err = auth.MustAdmin(ctx); err != nil {
		return nil, err
	}
	out = &model.RecipeManageListOutput{List: make([]model.RecipeManageItem, 0)}
	var entities []*model.RecipeManageItem
	err = dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		// 同 List：卡片/表格都用不到 ingredients/tools/steps 三个 JSON 大列
		Fields(
			dao.Recipes.Columns().Id,
			dao.Recipes.Columns().Title,
			dao.Recipes.Columns().CoverAttachmentId,
			dao.Recipes.Columns().Difficulty,
			dao.Recipes.Columns().CookMinutes,
			dao.Recipes.Columns().Calories,
			dao.Recipes.Columns().MealMask,
			dao.Recipes.Columns().Sort,
			dao.Recipes.Columns().CreatedAt,
			dao.Recipes.Columns().UpdatedAt,
		).
		OrderAsc(dao.Recipes.Columns().Sort).
		OrderDesc(dao.Recipes.Columns().Id).
		Scan(&entities)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return out, nil
	}

	// 封面 URL 与标签：复用 List 的两个批量查询（各一次）
	coverIds := make([]int64, 0, len(entities))
	recipeIds := make([]int64, 0, len(entities))
	for _, e := range entities {
		recipeIds = append(recipeIds, e.Id)
		if e.CoverAttachmentId > 0 {
			coverIds = append(coverIds, e.CoverAttachmentId)
		}
	}
	urlById, err := s.attachmentUrlMap(ctx, coverIds)
	if err != nil {
		return nil, err
	}
	tagsMap, err := s.recipeTagMap(ctx, recipeIds)
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		e.CoverUrl = urlById[e.CoverAttachmentId]
		if tags, ok := tagsMap[e.Id]; ok {
			e.Tags = tags
		} else {
			e.Tags = make([]model.TagItem, 0)
		}
		out.List = append(out.List, *e)
	}
	return out, nil
}

// Reorder 菜谱手动排序：按传入顺序把 sort 重写为 1、2、3…（越小越靠前）；仅管理员可操作。
// 传入的 id 必须是当前全部未删除菜谱的完整顺序（不重不漏）：只提交子集会让未参与的菜谱
// sort 停滞，出现两个菜谱同 sort 的歧义顺序，因此宁可整体拒绝也不写半截顺序。
func (s *sRecipe) Reorder(ctx context.Context, ids []int64) (err error) {
	if err = auth.MustAdmin(ctx); err != nil {
		return err
	}
	seen := make(map[int64]struct{}, len(ids))
	unique := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return gerror.Newf("菜谱 id 不合法: %d", id)
		}
		if _, ok := seen[id]; ok {
			return gerror.New("提交的菜谱顺序有重复 id，排序未生效")
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	total, err := dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return err
	}
	if total != len(unique) {
		return gerror.Newf("菜谱已变化（当前 %d 个，提交 %d 个），请刷新后重试", total, len(unique))
	}
	count, err := dao.Recipes.Ctx(ctx).
		WhereIn(dao.Recipes.Columns().Id, unique).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return err
	}
	if count != len(unique) {
		return gerror.New("存在不存在的菜谱 id，排序未生效")
	}
	return dao.Recipes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for i, id := range unique {
			// 只改 sort：重排不是「编辑菜谱」，不动 updated_at
			if _, err := dao.Recipes.Ctx(ctx).
				Where(dao.Recipes.Columns().Id, id).
				Data(do.Recipes{Sort: i + 1}).
				Update(); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListByIds 按 id 批量取食谱卡片项（保持传入顺序，剔除不存在/已删除）
func (s *sRecipe) ListByIds(ctx context.Context, ids []int64) (out map[int64]model.RecipeListItem, err error) {
	out = make(map[int64]model.RecipeListItem, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	var entities []*model.RecipeListItem
	err = dao.Recipes.Ctx(ctx).
		WhereIn(dao.Recipes.Columns().Id, ids).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		// 同 List：卡片字段白名单，不拉 ingredients/tools/steps 大列
		Fields(
			dao.Recipes.Columns().Id,
			dao.Recipes.Columns().Title,
			dao.Recipes.Columns().Summary,
			dao.Recipes.Columns().CoverAttachmentId,
			dao.Recipes.Columns().Calories,
			dao.Recipes.Columns().Difficulty,
			dao.Recipes.Columns().CookMinutes,
		).
		Scan(&entities)
	if err != nil {
		return nil, err
	}
	if len(entities) == 0 {
		return out, nil
	}
	coverIds := make([]int64, 0, len(entities))
	for _, e := range entities {
		if e.CoverAttachmentId > 0 {
			coverIds = append(coverIds, e.CoverAttachmentId)
		}
	}
	urlById, err := s.attachmentUrlMap(ctx, coverIds)
	if err != nil {
		return nil, err
	}
	recipeIds := make([]int64, 0, len(entities))
	for _, e := range entities {
		recipeIds = append(recipeIds, e.Id)
	}
	tagsMap, err := s.recipeTagMap(ctx, recipeIds)
	if err != nil {
		return nil, err
	}
	favCntMap, err := s.favoriteCountMap(ctx, recipeIds)
	if err != nil {
		return nil, err
	}
	for _, e := range entities {
		e.CoverUrl = urlById[e.CoverAttachmentId]
		if tags, ok := tagsMap[e.Id]; ok {
			e.Tags = tags
		} else {
			e.Tags = make([]model.TagItem, 0)
		}
		e.FavoriteCount = favCntMap[e.Id]
		out[e.Id] = *e
	}
	return out, nil
}
