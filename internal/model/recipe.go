// =================================================================================
// 食谱领域模型 —— 字段唯一定义点
// =================================================================================

package model

// RecipeIngredientInput 食材项（JSON 存储形状）
type RecipeIngredientInput struct {
	Name     string `json:"name"     v:"required#食材名不能为空" dc:"食材名，如 五花肉"` // 食材名
	Amount   string `json:"amount"   dc:"用量，如 500g"`                      // 用量
	Optional int    `json:"optional" dc:"0=必选 1=可选"`                      // 0=必选 1=可选
}

// RecipeStepInput 步骤项（JSON 存储形状）
type RecipeStepInput struct {
	Title   string  `json:"title"   dc:"步骤标题（可选），如「焯水去腥」"`
	Content string  `json:"content" v:"required#步骤内容不能为空" dc:"步骤内容 markdown"` // 步骤内容
	Media   []int64 `json:"media"   dc:"步骤媒体附件 id 列表"`                        // 媒体附件 id 列表
}

// RecipeIdInput 食谱 id 入参
type RecipeIdInput struct {
	Id int64 `json:"id" v:"required|min:1#食谱id不能为空|食谱id不合法" dc:"食谱 id"`
}

// 用餐时间位掩码常量（recipes.meal_mask）
const (
	MealBreakfast = 1 // 早餐
	MealLunch     = 2 // 午餐
	MealDinner    = 4 // 晚餐
	MealSnack     = 8 // 加餐
)

// RecipeListInput 食谱列表查询入参
type RecipeListInput struct {
	PaginationInput
	Keywords string `json:"keywords" dc:"关键词，匹配标题/简介" v:"length:0,50#关键词最长50"`
	TagIds   []int  `json:"tagIds"   dc:"标签 id 列表，多选为 AND 叠加（菜谱需同时含所有选中标签）"` // 多选标签过滤
	Meal     int    `json:"meal"     dc:"用餐时间筛选：1=早餐 2=午餐 4=晚餐 8=加餐，0/缺省=不筛" v:"min:0|max:15#用餐时间参数不合法"`
}

// RecipeStepItem 步骤项（列表/编辑返回形状）
type RecipeStepItem struct {
	Title   string  `json:"title"   dc:"步骤标题（可选），如「焯水去腥」"`
	Content string  `json:"content" dc:"步骤内容 markdown"`
	Media   []int64 `json:"media"   dc:"步骤媒体附件 id 列表"`
}

// RecipeDetail 食谱详情（编辑弹窗数据源）
type RecipeDetail struct {
	Id                int64                   `json:"id"                dc:"食谱 id"`
	Title             string                  `json:"title"             dc:"菜谱标题"`
	Summary           string                  `json:"summary"           dc:"一句话简介"`
	CoverAttachmentId int64                   `json:"coverAttachmentId" dc:"封面附件 id，0=无封面"`
	CoverUrl          string                  `json:"coverUrl"          dc:"封面图访问 URL，空=无封面"`
	Tips              string                  `json:"tips"              dc:"注意事项"`
	Ingredients       []RecipeIngredientInput `json:"ingredients" dc:"食材列表"`
	Tools             []string                `json:"tools"             dc:"工具列表"`
	Steps             []RecipeStepItem        `json:"steps"             dc:"步骤列表（保序）"`
	Calories          int                     `json:"calories"          dc:"本菜谱总热量 kcal，0=未填"`
	Difficulty        int                     `json:"difficulty"        dc:"难度：0=未填 1=简单 2=中等 3=较难"`
	CookMinutes       int                     `json:"cookMinutes"       dc:"耗时(分钟)，0=未填"`
	MealMask          int                     `json:"mealMask"          dc:"适合用餐时间位掩码：1=早餐 2=午餐 4=晚餐 8=加餐"`
	Tags              []TagItem               `json:"tags"              dc:"菜谱标签列表"`
}

// RecipeUpdateInput 食谱编辑入参。
// 字段一律「指针语义」：null/不传 = 本次不修改；传 0 或空串 = 显式清空该字段
// （否则 summary/tips/封面/难度/热量等一旦设置就再也清不掉）。
// Steps/Ingredients/Tools 为整体覆盖：null=不改，[]=清空。
type RecipeUpdateInput struct {
	RecipeIdInput
	Title             *string                 `json:"title"               dc:"菜谱标题，不传=不修改" v:"length:1,100#标题不能为空|标题最长100"`
	Summary           *string                 `json:"summary"             dc:"一句话简介；不传=不修改，传空串=清空" v:"length:0,200#简介最长200"`
	CoverAttachmentId *int64                  `json:"coverAttachmentId"   dc:"封面附件 id；不传=不修改，传 0=清除封面" v:"min:0#封面 id 不合法"`
	Tips              *string                 `json:"tips"                dc:"注意事项；不传=不修改，传空串=清空" v:"length:0,500#注意事项最长500"`
	Ingredients       []RecipeIngredientInput `json:"ingredients"         dc:"食材列表，整体覆盖；null=不修改 []=清空"`
	Tools             []string                `json:"tools"               dc:"工具列表，整体覆盖；null=不修改 []=清空"`
	TagIds            []int                   `json:"tagIds"              dc:"标签 id 列表（分类），整体覆盖；null=不修改 []=清空"`
	Steps             []RecipeStepItem        `json:"steps"               dc:"步骤列表，整体覆盖；null=不修改 []=清空"`
	MealMask          *int                    `json:"mealMask"            dc:"用餐时间位掩码：1=早餐 2=午餐 4=晚餐 8=加餐；不传=不修改，传 0=未填" v:"min:0|max:15#用餐时间不合法"`
	Difficulty        *int                    `json:"difficulty"          dc:"难度：0=未填 1=简单 2=中等 3=较难；不传=不修改" v:"min:0|max:3#难度取值 0-3"`
	CookMinutes       *int                    `json:"cookMinutes"         dc:"耗时(分钟)；不传=不修改，传 0=未填" v:"min:0#耗时不能为负"`
	Calories          *int                    `json:"calories"            dc:"每份热量 kcal；不传=不修改，传 0=未填" v:"min:0#热量不能为负"`
}

// RecipeCreateInput 食谱新建入参（标题必填；列表字段整体写入）
type RecipeCreateInput struct {
	Title             string                  `json:"title"             v:"required|length:1,100#标题不能为空|标题最长100" dc:"菜谱标题"`
	Summary           string                  `json:"summary"           dc:"一句话简介" v:"length:0,200#简介最长200"`
	CoverAttachmentId int64                   `json:"coverAttachmentId" dc:"封面附件 id，0=无封面"`
	Tips              string                  `json:"tips"              dc:"注意事项" v:"length:0,500#注意事项最长500"`
	Ingredients       []RecipeIngredientInput `json:"ingredients"       dc:"食材列表"`
	Tools             []string                `json:"tools"             dc:"工具列表"`
	TagIds            []int                   `json:"tagIds"            dc:"标签 id 列表（分类），整体写入"`
	Steps             []RecipeStepItem        `json:"steps"             dc:"步骤列表"`
	MealMask          int                     `json:"mealMask"          dc:"用餐时间位掩码：1=早餐 2=午餐 4=晚餐 8=加餐；15=全部时段" v:"min:0|max:15#用餐时间不合法"`
	Difficulty        int                     `json:"difficulty"        dc:"难度：0=未填 1=简单 2=中等 3=较难" v:"min:0|max:3#难度取值 0-3"`
	CookMinutes       int                     `json:"cookMinutes"       dc:"耗时(分钟)，0=未填" v:"min:0#耗时不能为负"`
	Calories          int                     `json:"calories"          dc:"每份热量 kcal，0=未填" v:"min:0#热量不能为负"`
}

// RecipeListItem 食谱列表项（卡片展示字段）
type RecipeListItem struct {
	Id                int64     `json:"id"                    dc:"食谱 id"`
	Title             string    `json:"title"                 dc:"菜谱标题"`
	Summary           string    `json:"summary"               dc:"一句话简介"`
	CoverAttachmentId int64     `json:"coverAttachmentId"     dc:"封面附件 id，0=无封面"`
	CoverUrl          string    `json:"coverUrl"              dc:"封面图访问 URL，空=无封面"`
	Calories          int       `json:"calories"              dc:"本菜谱总热量 kcal，0=未填"`
	Difficulty        int       `json:"difficulty"            dc:"难度：0=未填 1=简单 2=中等 3=较难"`
	CookMinutes       int       `json:"cookMinutes"           dc:"耗时(分钟)，0=未填"`
	FavoriteCount     int       `json:"favoriteCount"         dc:"被收藏次数（收藏夹数）"`
	Tags              []TagItem `json:"tags"                  dc:"菜谱标签列表"`
}

// RecipeListOutput 食谱列表响应
type RecipeListOutput = PageRes[RecipeListItem]

// RecipeManageItem 菜谱管理表格行（设置页「菜谱管理」；比卡片多出排序位与时间）
type RecipeManageItem struct {
	Id                int64     `json:"id"                dc:"菜谱 id"`
	Title             string    `json:"title"             dc:"菜谱标题"`
	CoverAttachmentId int64     `json:"coverAttachmentId" dc:"封面附件 id，0=无封面"`
	CoverUrl          string    `json:"coverUrl"          dc:"封面图访问 URL，空=无封面"`
	Tags              []TagItem `json:"tags"              dc:"菜谱标签列表"`
	Difficulty        int       `json:"difficulty"        dc:"难度：0=未填 1=简单 2=中等 3=较难"`
	CookMinutes       int       `json:"cookMinutes"       dc:"耗时(分钟)，0=未填"`
	Calories          int       `json:"calories"          dc:"本菜谱总热量 kcal，0=未填"`
	MealMask          int       `json:"mealMask"          dc:"适合用餐时间位掩码：1=早餐 2=午餐 4=晚餐 8=加餐"`
	Sort              int       `json:"sort"              dc:"手动排序位：0=未参与排序（按 id 倒序=最新在前）"`
	CreatedAt         int64     `json:"createdAt"         dc:"创建时间 Unix 秒"`
	UpdatedAt         int64     `json:"updatedAt"         dc:"更新时间 Unix 秒"`
}

// RecipeManageListOutput 菜谱管理列表响应。
// 全量返回、不分页：手动排序提交的是「全部菜谱的完整顺序」，分页会让顺序残缺。
type RecipeManageListOutput struct {
	List []RecipeManageItem `json:"list" dc:"菜谱列表（按手动顺序）"`
}

// RecipeReorderInput 菜谱手动排序入参（仅管理员）
type RecipeReorderInput struct {
	Ids []int64 `json:"ids" v:"required#排序 id 列表不能为空" dc:"全部未删除菜谱的完整 id 顺序（不重不漏）"`
}
