// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Users is the golang structure for table users.
type Users struct {
	Id                 int    `json:"id"                 orm:"id"                   description:""` //
	Username           string `json:"username"           orm:"username"             description:""` //
	Nickname           string `json:"nickname"           orm:"nickname"             description:""` //
	PasswordHash       string `json:"passwordHash"       orm:"password_hash"        description:""` //
	AvatarAttachmentId int    `json:"avatarAttachmentId" orm:"avatar_attachment_id" description:""` //
	Status             int    `json:"status"             orm:"status"               description:""` //
	CreatedAt          int    `json:"createdAt"          orm:"created_at"           description:""` //
	UpdatedAt          int    `json:"updatedAt"          orm:"updated_at"           description:""` //
}
