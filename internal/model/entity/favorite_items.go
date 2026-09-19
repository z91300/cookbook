// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// FavoriteItems is the golang structure for table favorite_items.
type FavoriteItems struct {
	FavoriteId int    `json:"favoriteId" orm:"favorite_id" description:""` //
	RecipeId   int    `json:"recipeId"   orm:"recipe_id"   description:""` //
	Note       string `json:"note"       orm:"note"        description:""` //
	CreatedAt  int    `json:"createdAt"  orm:"created_at"  description:""` //
}
