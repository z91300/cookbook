// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Schedulings is the golang structure of table schedulings for DAO operations like Where/Data.
type Schedulings struct {
	g.Meta    `orm:"table:schedulings, do:true"`
	Id        any //
	PlanDate  any //
	Meal      any //
	RecipeId  any //
	Servings  any //
	Note      any //
	Sort      any //
	IsDeleted any //
	CreatedAt any //
	UpdatedAt any //
}
