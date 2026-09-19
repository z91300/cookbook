// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AiGenerations is the golang structure for table ai_generations.
type AiGenerations struct {
	Id        int    `json:"id"        orm:"id"         description:""` //
	UserId    int    `json:"userId"    orm:"user_id"    description:""` //
	RecipeId  int    `json:"recipeId"  orm:"recipe_id"  description:""` //
	Model     string `json:"model"     orm:"model"      description:""` //
	Prompt    string `json:"prompt"    orm:"prompt"     description:""` //
	RawOutput string `json:"rawOutput" orm:"raw_output" description:""` //
	Status    int    `json:"status"    orm:"status"     description:""` //
	Error     string `json:"error"     orm:"error"      description:""` //
	CreatedAt int    `json:"createdAt" orm:"created_at" description:""` //
}
