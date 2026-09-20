// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Schedulings is the golang structure for table schedulings.
type Schedulings struct {
	Id        int    `json:"id"        orm:"id"         description:""` //
	PlanDate  int    `json:"planDate"  orm:"plan_date"  description:""` //
	Meal      int    `json:"meal"      orm:"meal"       description:""` //
	RecipeId  int    `json:"recipeId"  orm:"recipe_id"  description:""` //
	Note      string `json:"note"      orm:"note"       description:""` //
	Sort      int    `json:"sort"      orm:"sort"       description:""` //
	IsDeleted int    `json:"isDeleted" orm:"is_deleted" description:""` //
	CreatedAt int    `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt int    `json:"updatedAt" orm:"updated_at" description:""` //
}
