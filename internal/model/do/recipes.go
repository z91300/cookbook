// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Recipes is the golang structure of table recipes for DAO operations like Where/Data.
type Recipes struct {
	g.Meta            `orm:"table:recipes, do:true"`
	Id                any //
	UserId            any //
	Title             any //
	Summary           any //
	CoverAttachmentId any //
	Tips              any //
	Ingredients       any //
	Tools             any //
	Steps             any //
	MealMask          any //
	Calories          any //
	Difficulty        any //
	Servings          any //
	CookMinutes       any //
	Source            any //
	AiModel           any //
	ReviewStatus      any //
	IsDeleted         any //
	CreatedAt         any //
	UpdatedAt         any //
}
