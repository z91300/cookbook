// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Attachments is the golang structure for table attachments.
type Attachments struct {
	Id          int    `json:"id"          orm:"id"            description:""` //
	OwnerUserId int    `json:"ownerUserId" orm:"owner_user_id" description:""` //
	Kind        string `json:"kind"        orm:"kind"          description:""` //
	FileName    string `json:"fileName"    orm:"file_name"     description:""` //
	StoragePath string `json:"storagePath" orm:"storage_path"  description:""` //
	MimeType    string `json:"mimeType"    orm:"mime_type"     description:""` //
	SizeBytes   int64  `json:"sizeBytes"   orm:"size_bytes"    description:""` //
	Width       int    `json:"width"       orm:"width"         description:""` //
	Height      int    `json:"height"      orm:"height"        description:""` //
	Duration    int    `json:"duration"    orm:"duration"      description:""` //
	Sha256      string `json:"sha256"      orm:"sha256"        description:""` //
	IsDeleted   int    `json:"isDeleted"   orm:"is_deleted"    description:""` //
	CreatedAt   int    `json:"createdAt"   orm:"created_at"    description:""` //
	UpdatedAt   int    `json:"updatedAt"   orm:"updated_at"    description:""` //
	Content     []byte `json:"content"     orm:"content"       description:""` //
}
