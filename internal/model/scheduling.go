// =================================================================================
// 食谱编排领域模型 —— 字段唯一定义点
// =================================================================================

package model

// SchedulingItem 编排项（列表返回形状，含食谱卡片展示字段）
type SchedulingItem struct {
	Id          int64  `json:"id"          dc:"编排条目 id"`
	PlanDate    int    `json:"planDate"    dc:"编排日期 YYYYMMDD，如 20260916"`
	Meal        int    `json:"meal"        dc:"用餐时段：1=早餐 2=午餐 4=晚餐 8=加餐"`
	RecipeId    int64  `json:"recipeId"    dc:"食谱 id"`
	RecipeTitle string `json:"recipeTitle" dc:"食谱标题（冗余展示字段，列表接口回填）"`
	CoverUrl    string `json:"coverUrl"    dc:"食谱封面 URL，空=无封面"`
	Note        string `json:"note"        dc:"编排备注"`
	Sort        int    `json:"sort"        dc:"同槽位内排序"`
}

// SchedulingIdInput 编排条目 id 入参
type SchedulingIdInput struct {
	Id int64 `json:"id" v:"required|min:1#编排id不能为空|编排id不合法" dc:"编排条目 id"`
}

// SchedulingListInput 编排列表查询入参（按日期区间，含端点）
type SchedulingListInput struct {
	StartDate int `json:"startDate" dc:"起始日期 YYYYMMDD（含），0=不限制"`
	EndDate   int `json:"endDate"   dc:"结束日期 YYYYMMDD（含），0=不限制"`
}

// SchedulingSaveInput 编排新增/编辑入参（全量覆盖语义；meal 取 1/2/4/8 之一）
type SchedulingSaveInput struct {
	PlanDate int    `json:"planDate" v:"required|min:20000101|max:99991231#编排日期不能为空|编排日期不合法" dc:"编排日期 YYYYMMDD"`
	Meal     int    `json:"meal"     v:"required|in:1,2,4,8#用餐时段不能为空|用餐时段取值 1/2/4/8" dc:"用餐时段：1=早餐 2=午餐 4=晚餐 8=加餐"`
	RecipeId int64  `json:"recipeId" v:"required|min:1#食谱id不能为空|食谱id不合法" dc:"食谱 id"`
	Note     string `json:"note"     v:"length:0,200#备注最长200" dc:"编排备注"`
	Sort     int    `json:"sort"     dc:"同槽位内排序"`
}
