// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Settings is the golang structure for table settings.
type Settings struct {
	Id        int    `json:"id"        orm:"id"         description:""` //
	Key       string `json:"key"       orm:"key"        description:""` //
	Value     string `json:"value"     orm:"value"      description:""` //
	IsDeleted int    `json:"isDeleted" orm:"is_deleted" description:""` //
	CreatedAt int    `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt int    `json:"updatedAt" orm:"updated_at" description:""` //
}
