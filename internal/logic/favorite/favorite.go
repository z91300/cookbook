// =================================================================================
// 收藏业务实现：收藏夹管理 + 收藏条目
// =================================================================================

package favorite

import (
	"context"

	"cookbook/internal/dao"
	"cookbook/internal/logic/auth"
	"cookbook/internal/model"
	"cookbook/internal/service"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type sFavorite struct{}

func init() {
	service.RegisterFavorite(New())
}

func New() *sFavorite {
	return &sFavorite{}
}

// defaultFavoriteName 默认收藏夹名
const defaultFavoriteName = "我的收藏"

// EnsureDefaultFolder 确保该用户至少有一个收藏夹「我的收藏」（幂等）。
// 在注册等「初始化时机」调用：LIST 是 GET，绝不能带写副作用（前端在弹窗/页面里
// 反复调用 getList，旧实现每次无夹都插一行，且多用户下会把夹挂到 user_id=0 造成归属错乱）。
// 用一条 INSERT ... WHERE NOT EXISTS 原子完成，天然免疫并发重复插入。
func (s *sFavorite) EnsureDefaultFolder(ctx context.Context, userId int64) (err error) {
	now := gtime.Timestamp()
	_, err = dao.Favorites.DB().Exec(ctx, `
INSERT INTO favorites (user_id, name, description, is_public, sort, is_deleted, created_at, updated_at)
SELECT ?, ?, '', 0, 0, 0, ?, ?
WHERE NOT EXISTS (
    SELECT 1 FROM favorites WHERE user_id = ? AND is_deleted = 0
)`, userId, defaultFavoriteName, now, now, userId)
	return err
}

// List 收藏夹列表（纯读，无任何写副作用）；
// recipeCount 用一条 GROUP BY 聚合填充
func (s *sFavorite) List(ctx context.Context) (out []*model.FavoriteFolder, err error) {
	err = dao.Favorites.Ctx(ctx).
		Where(dao.Favorites.Columns().IsDeleted, 0).
		Order(dao.Favorites.Columns().Sort).Order(dao.Favorites.Columns().Id).
		Scan(&out)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = make([]*model.FavoriteFolder, 0)
	}

	// 每夹收藏数：一条聚合查询
	type countRow struct {
		FavoriteId int64 `json:"favorite_id"`
		Cnt        int   `json:"cnt"`
	}
	var rows []countRow
	if err = dao.FavoriteItems.Ctx(ctx).
		Fields("favorite_id, COUNT(*) AS cnt").
		Group("favorite_id").
		Scan(&rows); err != nil {
		return nil, err
	}
	cntByFid := make(map[int64]int, len(rows))
	for _, r := range rows {
		cntByFid[r.FavoriteId] = r.Cnt
	}
	for _, folder := range out {
		folder.RecipeCount = cntByFid[folder.Id]
	}
	return out, nil
}

// Create 新建收藏夹
func (s *sFavorite) Create(ctx context.Context, in model.FavoriteSaveInput) (int64, error) {
	if err := auth.MustLogin(ctx); err != nil {
		return 0, err
	}
	now := gtime.Timestamp()
	id, err := dao.Favorites.Ctx(ctx).
		Data(g.Map{
			"user_id":    0,
			"name":       in.Name,
			"description": in.Description,
			"is_public":  0,
			"sort":       in.Sort,
			"created_at": now,
			"updated_at": now,
		}).
		InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Update 编辑收藏夹
func (s *sFavorite) Update(ctx context.Context, id int64, in model.FavoriteSaveInput) (err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return err
	}
	count, err := dao.Favorites.Ctx(ctx).
		Where(dao.Favorites.Columns().Id, id).
		Where(dao.Favorites.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return gerror.Newf("收藏夹不存在: %d", id)
	}
	_, err = dao.Favorites.Ctx(ctx).
		Where(dao.Favorites.Columns().Id, id).
		Data(g.Map{
			"name":        in.Name,
			"description": in.Description,
			"sort":        in.Sort,
			"updated_at":  gtime.Timestamp(),
		}).
		Update()
	return err
}

// Delete 删除收藏夹（逻辑删）并硬删夹内条目；食谱本体不动
func (s *sFavorite) Delete(ctx context.Context, id int64) (err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return err
	}
	err = dao.Favorites.Transaction(ctx, func(ctx context.Context, tx gdb.TX) (err error) {
		if _, err = dao.Favorites.Ctx(ctx).
			Where(dao.Favorites.Columns().Id, id).
			Data(g.Map{"is_deleted": 1, "updated_at": gtime.Timestamp()}).
			Update(); err != nil {
			return err
		}
		_, err = dao.FavoriteItems.Ctx(ctx).
			Where(dao.FavoriteItems.Columns().FavoriteId, id).
			Delete()
		return err
	})
	return err
}

// GetRecipes 收藏夹内食谱，按收藏时间倒序；经 recipe.ListByIds 组装卡片字段
func (s *sFavorite) GetRecipes(ctx context.Context, favoriteId int64) (out []model.RecipeListItem, err error) {
	out = make([]model.RecipeListItem, 0)
	count, err := dao.Favorites.Ctx(ctx).
		Where(dao.Favorites.Columns().Id, favoriteId).
		Where(dao.Favorites.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gerror.Newf("收藏夹不存在: %d", favoriteId)
	}
	type itemRow struct {
		RecipeId int64 `json:"recipe_id"`
	}
	var rows []itemRow
	if err = dao.FavoriteItems.Ctx(ctx).
		Where(dao.FavoriteItems.Columns().FavoriteId, favoriteId).
		OrderDesc(dao.FavoriteItems.Columns().CreatedAt).
		OrderDesc(dao.FavoriteItems.Columns().RecipeId).
		Scan(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return out, nil
	}
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.RecipeId)
	}
	byId, err := service.Recipe().ListByIds(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		if item, ok := byId[id]; ok {
			out = append(out, item)
		}
	}
	return out, nil
}

// GetByRecipe 查询食谱被收藏在哪些（未删除的）收藏夹
func (s *sFavorite) GetByRecipe(ctx context.Context, recipeId int64) (out []int64, err error) {
	out = make([]int64, 0)
	records, err := dao.FavoriteItems.Ctx(ctx).
		InnerJoin("favorites", "favorites.id=favorite_items.favorite_id").
		Where("favorite_items.recipe_id", recipeId).
		Where("favorites.is_deleted", 0).
		Fields("favorite_items.favorite_id").
		All()
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		out = append(out, record["favorite_id"].Int64())
	}
	return out, nil
}

// AddItem 收藏食谱到夹：已存在行则更新 note，否则插入
func (s *sFavorite) AddItem(ctx context.Context, favoriteId int64, in model.FavoriteAddItemInput) (err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return err
	}
	// 收藏夹存在性
	fCount, err := dao.Favorites.Ctx(ctx).
		Where(dao.Favorites.Columns().Id, favoriteId).
		Where(dao.Favorites.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return err
	}
	if fCount == 0 {
		return gerror.Newf("收藏夹不存在: %d", favoriteId)
	}
	// 食谱存在性
	rCount, err := dao.Recipes.Ctx(ctx).
		Where(dao.Recipes.Columns().Id, in.RecipeId).
		Where(dao.Recipes.Columns().IsDeleted, 0).
		Count()
	if err != nil {
		return err
	}
	if rCount == 0 {
		return gerror.Newf("食谱不存在或已删除: %d", in.RecipeId)
	}
	now := gtime.Timestamp()
	existing, err := dao.FavoriteItems.Ctx(ctx).
		Where(dao.FavoriteItems.Columns().FavoriteId, favoriteId).
		Where(dao.FavoriteItems.Columns().RecipeId, in.RecipeId).
		Count()
	if err != nil {
		return err
	}
	if existing > 0 {
		_, err = dao.FavoriteItems.Ctx(ctx).
			Where(dao.FavoriteItems.Columns().FavoriteId, favoriteId).
			Where(dao.FavoriteItems.Columns().RecipeId, in.RecipeId).
			Data(g.Map{"note": in.Note}).
			Update()
		return err
	}
	_, err = dao.FavoriteItems.Ctx(ctx).
		Data(g.Map{
			"favorite_id": favoriteId,
			"recipe_id":   in.RecipeId,
			"note":        in.Note,
			"created_at":  now,
		}).
		Insert()
	return err
}

// RemoveItem 从收藏夹取消收藏（硬删条目行）
func (s *sFavorite) RemoveItem(ctx context.Context, favoriteId int64, recipeId int64) (err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return err
	}
	_, err = dao.FavoriteItems.Ctx(ctx).
		Where(dao.FavoriteItems.Columns().FavoriteId, favoriteId).
		Where(dao.FavoriteItems.Columns().RecipeId, recipeId).
		Delete()
	return err
}