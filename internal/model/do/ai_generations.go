// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// AiGenerations is the golang structure of table ai_generations for DAO operations like Where/Data.
type AiGenerations struct {
	g.Meta    `orm:"table:ai_generations, do:true"`
	Id        any //
	UserId    any //
	RecipeId  any //
	Model     any //
	Prompt    any //
	RawOutput any //
	Status    any //
	Error     any //
	CreatedAt any //
}
