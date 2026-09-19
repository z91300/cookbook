// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Recipes is the golang structure for table recipes.
type Recipes struct {
	Id                int    `json:"id"                orm:"id"                  description:""` //
	UserId            int    `json:"userId"            orm:"user_id"             description:""` //
	Title             string `json:"title"             orm:"title"               description:""` //
	Summary           string `json:"summary"           orm:"summary"             description:""` //
	CoverAttachmentId int    `json:"coverAttachmentId" orm:"cover_attachment_id" description:""` //
	Tips              string `json:"tips"              orm:"tips"                description:""` //
	Ingredients       string `json:"ingredients"       orm:"ingredients"         description:""` //
	Tools             string `json:"tools"             orm:"tools"               description:""` //
	Steps             string `json:"steps"             orm:"steps"               description:""` //
	MealMask          int    `json:"mealMask"          orm:"meal_mask"           description:""` //
	Calories          int    `json:"calories"          orm:"calories"            description:""` //
	Difficulty        int    `json:"difficulty"        orm:"difficulty"          description:""` //
	Servings          int    `json:"servings"          orm:"servings"            description:""` //
	CookMinutes       int    `json:"cookMinutes"       orm:"cook_minutes"        description:""` //
	Source            string `json:"source"            orm:"source"              description:""` //
	AiModel           string `json:"aiModel"           orm:"ai_model"            description:""` //
	ReviewStatus      int    `json:"reviewStatus"      orm:"review_status"       description:""` //
	IsDeleted         int    `json:"isDeleted"         orm:"is_deleted"          description:""` //
	CreatedAt         int    `json:"createdAt"         orm:"created_at"          description:""` //
	UpdatedAt         int    `json:"updatedAt"         orm:"updated_at"          description:""` //
}
