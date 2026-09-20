// ==========================================================================
// 设置业务实现：通用 key-value 存取 + 提示词模板渲染
// ==========================================================================

package setting

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"cookbook/internal/dao"
	"cookbook/internal/logic/auth"
	"cookbook/internal/model"
	"cookbook/internal/service"
)

type sSetting struct{}

func init() {
	service.RegisterSetting(New())
}

func New() *sSetting {
	return &sSetting{}
}

// List 查询全部正常设置项（按 key 升序）
func (s *sSetting) List(ctx context.Context) (out []*model.SettingItem, err error) {
	err = dao.Settings.Ctx(ctx).
		Fields(dao.Settings.Columns().Key, dao.Settings.Columns().Value).
		Where(dao.Settings.Columns().IsDeleted, 0).
		Order(dao.Settings.Columns().Key).
		Scan(&out)
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = make([]*model.SettingItem, 0)
	}
	return out, nil
}

// Upsert 批量保存：存在则更新 value 与 updated_at，不存在则插入
func (s *sSetting) Upsert(ctx context.Context, items model.SettingUpsertInputList) (err error) {
	if err = auth.MustLogin(ctx); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	if _, err = dao.Settings.Ctx(ctx).
		Data(toRecords(items)).
		OnConflict(dao.Settings.Columns().Key).
		Save(); err != nil {
		return err
	}
	return nil
}

// Render 取 key 对应模板，替换 {recipe.*} 变量为指定食谱内容
func (s *sSetting) Render(ctx context.Context, key string, recipeId int64) (string, error) {
	value, err := dao.Settings.Ctx(ctx).
		Fields(dao.Settings.Columns().Value).
		Where(dao.Settings.Columns().Key, key).
		Where(dao.Settings.Columns().IsDeleted, 0).
		Value()
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", gerror.NewCode(gcode.CodeNotFound, "未找到该设置项，请先在设置页面填写提示词模板")
	}
	tpl := value.String()
	if strings.TrimSpace(tpl) == "" {
		return "", gerror.NewCode(gcode.CodeInvalidOperation, "提示词模板为空，请先在设置页面填写")
	}

	detail, err := service.Recipe().GetOne(ctx, recipeId)
	if err != nil {
		return "", err
	}
	if detail == nil {
		return "", gerror.NewCode(gcode.CodeNotFound, "食谱不存在或已删除")
	}
	return renderTemplate(tpl, detail), nil
}

// renderTemplate 按模板语法替换变量：
// 基础：{recipe.title} {recipe.summary} {recipe.tips} {recipe.ingredients} {recipe.tools} {recipe.steps}
// 数值：{recipe.calories} {recipe.difficulty} {recipe.servings} {recipe.cookMinutes} {recipe.mealMask}
// 计数/汇总：{recipe.stepCount} {recipe.ingredientCount} {recipe.infoLine}
func renderTemplate(tpl string, d *model.RecipeDetail) string {
	replacer := strings.NewReplacer(
		"{recipe.title}", d.Title,
		"{recipe.summary}", d.Summary,
		"{recipe.tips}", tipsText(d.Tips),
		"{recipe.ingredients}", formatIngredients(d.Ingredients),
		"{recipe.tools}", formatTools(d.Tools),
		"{recipe.steps}", formatSteps(d.Steps),
		"{recipe.stepCount}", gconv.String(len(d.Steps)),
		"{recipe.ingredientCount}", gconv.String(len(d.Ingredients)),
		"{recipe.calories}", numText(d.Calories, " kcal"),
		"{recipe.difficulty}", difficultyText(d.Difficulty),
		"{recipe.servings}", numText(d.Servings, ""),
		"{recipe.cookMinutes}", numText(d.CookMinutes, " 分钟"),
		"{recipe.mealMask}", mealMaskText(d.MealMask),
		"{recipe.infoLine}", infoLine(d),
	)
	return replacer.Replace(tpl)
}

// formatIngredients 食材逐行：「名称 用量（可选）」
func formatIngredients(items []model.RecipeIngredientInput) string {
	if len(items) == 0 {
		return "（无）"
	}
	lines := make([]string, 0, len(items))
	for _, it := range items {
		line := strings.TrimSpace(it.Name)
		if amount := strings.TrimSpace(it.Amount); amount != "" {
			line += " " + amount
		}
		if it.Optional == 1 {
			line += "（可选）"
		}
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

// formatSteps 步骤逐行带圈数字编号：① 内容（有标题则「标题：内容」）
func formatSteps(items []model.RecipeStepItem) string {
	if len(items) == 0 {
		return "（无）"
	}
	lines := make([]string, 0, len(items))
	for i, it := range items {
		text := strings.TrimSpace(it.Content)
		if title := strings.TrimSpace(it.Title); title != "" {
			if text != "" {
				text = title + "：" + text
			} else {
				text = title
			}
		}
		lines = append(lines, circledNum(i+1)+" "+text)
	}
	return strings.Join(lines, "\n")
}

// circledNum 1-20 映射为带圈数字（①…⑳），超出则回退「n.」
func circledNum(n int) string {
	if n >= 1 && n <= 20 {
		return string(rune('①' + n - 1))
	}
	return gconv.String(n) + "."
}

// tipsText 小贴士文案：空时输出「（本次为空，按要求省略该模块）」
func tipsText(tips string) string {
	if strings.TrimSpace(tips) == "" {
		return "（本次为空，按要求省略该模块）"
	}
	return tips
}

// infoLine 信息条：约 X 分钟 · 难度Z（按已填字段拼接，全未填返回空；不含「人份」）
func infoLine(d *model.RecipeDetail) string {
	parts := make([]string, 0, 2)
	if d.CookMinutes > 0 {
		parts = append(parts, "约 "+gconv.String(d.CookMinutes)+" 分钟")
	}
	if t := difficultyText(d.Difficulty); t != "" {
		parts = append(parts, "难度"+t)
	}
	return strings.Join(parts, " · ")
}

// formatTools 工具逐行
func formatTools(items []string) string {
	if len(items) == 0 {
		return "（无）"
	}
	return strings.Join(items, "\n")
}

// numText 数值文案：0（未填）返回空串
func numText(v int, unit string) string {
	if v <= 0 {
		return ""
	}
	return gconv.String(v) + unit
}

// difficultyText 难度文案
func difficultyText(v int) string {
	switch v {
	case 1:
		return "简单"
	case 2:
		return "中等"
	case 3:
		return "较难"
	default:
		return ""
	}
}

// mealMaskText 用餐时段文案（位掩码）
func mealMaskText(mask int) string {
	if mask <= 0 {
		return ""
	}
	parts := make([]string, 0, 4)
	if mask&model.MealBreakfast != 0 {
		parts = append(parts, "早餐")
	}
	if mask&model.MealLunch != 0 {
		parts = append(parts, "午餐")
	}
	if mask&model.MealDinner != 0 {
		parts = append(parts, "晚餐")
	}
	if mask&model.MealSnack != 0 {
		parts = append(parts, "加餐")
	}
	return strings.Join(parts, "、")
}

// toRecords 批量保存项转 dao Data 行列表（setting 包内 helper，不改 model）
func toRecords(items model.SettingUpsertInputList) gdb.List {
	rows := make(gdb.List, 0, len(items))
	for _, it := range items {
		rows = append(rows, g.Map{
			"key":   it.Key,
			"value": it.Value,
		})
	}
	return rows
}
