// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Favorites is the golang structure for table favorites.
type Favorites struct {
	Id                int    `json:"id"                orm:"id"                  description:""` //
	UserId            int    `json:"userId"            orm:"user_id"             description:""` //
	Name              string `json:"name"              orm:"name"                description:""` //
	Description       string `json:"description"       orm:"description"         description:""` //
	CoverAttachmentId int    `json:"coverAttachmentId" orm:"cover_attachment_id" description:""` //
	IsPublic          int    `json:"isPublic"          orm:"is_public"           description:""` //
	Sort              int    `json:"sort"              orm:"sort"                description:""` //
	IsDeleted         int    `json:"isDeleted"         orm:"is_deleted"          description:""` //
	CreatedAt         int    `json:"createdAt"         orm:"created_at"          description:""` //
	UpdatedAt         int    `json:"updatedAt"         orm:"updated_at"          description:""` //
}
