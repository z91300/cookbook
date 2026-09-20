// =================================================================================
// 食谱业务实现：分页列表 / 详情 / 编辑保存
// =================================================================================
package recipe

import (
	"context"
	"encoding/json"

	"cookbook/internal/dao"
	"cookbook/internal/logic/attachment"
	"cookbook/internal/logic/auth"
	"cookbook/internal/model"
	"cookbook/internal/model/do"
	"cookbook/internal/service"

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
	err = m.Page(in.Page, in.PageSize).OrderDesc(dao.Recipes.Columns().Id).ScanAndCount(&entities, &out.Total, true)
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

	// 标量列按需更新；JSON 列与标签整体覆盖
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

// saveRecipeColumnsInput 标量列 + JSON 列 + 标签的共享写入集
// （Create 与 Update 复用；空值不覆盖，TagIds nil=不修改）
type saveRecipeColumnsInput struct {
	Title             string
	Summary           string
	Tips              string
	CoverAttachmentId int64
	MealMask          int
	Difficulty        int
	CookMinutes       int
	Calories          int
	Ingredients       []model.RecipeIngredientInput
	Tools             []string
	Steps             []model.RecipeStepItem
	TagIds            []int
}

// saveRecipeColumns 保存 recipes 标量列（空值跳过）与 JSON 列，并按 TagIds 重建关联
func (s *sRecipe) saveRecipeColumns(ctx context.Context, recipeId int64, in saveRecipeColumnsInput) (err error) {
	data := do.Recipes{}
	if in.Title != "" {
		data.Title = in.Title
	}
	if in.Summary != "" {
		data.Summary = in.Summary
	}
	if in.Tips != "" {
		data.Tips = in.Tips
	}
	if in.CoverAttachmentId > 0 {
		data.CoverAttachmentId = in.CoverAttachmentId
	}
	if in.MealMask > 0 {
		data.MealMask = in.MealMask
	}
	if in.Difficulty > 0 {
		data.Difficulty = in.Difficulty
	}
	if in.CookMinutes > 0 {
		data.CookMinutes = in.CookMinutes
	}
	if in.Calories > 0 {
		data.Calories = in.Calories
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
	if _, err = dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().Id, recipeId).
		Data(data).
		Update(); err != nil {
		return err
	}

	// 标签整体覆盖：nil=不修改，非 nil=重建关联
	if in.TagIds != nil {
		if _, err = dao.RecipeTags.Ctx(ctx).
			Where(dao.RecipeTags.Columns().RecipeId, recipeId).
			Delete(); err != nil {
			return err
		}
		for _, tagId := range in.TagIds {
			if _, err = dao.RecipeTags.Ctx(ctx).Data(do.RecipeTags{
				RecipeId: recipeId,
				TagId:    tagId,
			}).Insert(); err != nil {
				return err
			}
		}
	}
	return nil
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
	err = s.saveRecipeColumns(ctx, id, saveRecipeColumnsInput{
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
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Delete 逻辑删除食谱（置 is_deleted=1）；仅管理员可操作
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
	_, err = dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().Id, id).
		Data(do.Recipes{IsDeleted: 1}).
		Update()
	return err
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
