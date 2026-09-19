// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Members is the golang structure for table members.
type Members struct {
	Id        int    `json:"id"        orm:"id"         description:""` //
	Name      string `json:"name"      orm:"name"       description:""` //
	Role      string `json:"role"      orm:"role"       description:""` //
	Note      string `json:"note"      orm:"note"       description:""` //
	Sort      int    `json:"sort"      orm:"sort"       description:""` //
	IsDeleted int    `json:"isDeleted" orm:"is_deleted" description:""` //
	CreatedAt int    `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt int    `json:"updatedAt" orm:"updated_at" description:""` //
}
