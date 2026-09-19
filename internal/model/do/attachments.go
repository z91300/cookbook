// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Attachments is the golang structure of table attachments for DAO operations like Where/Data.
type Attachments struct {
	g.Meta      `orm:"table:attachments, do:true"`
	Id          any    //
	OwnerUserId any    //
	Kind        any    //
	FileName    any    //
	StoragePath any    //
	MimeType    any    //
	SizeBytes   any    //
	Width       any    //
	Height      any    //
	Duration    any    //
	Sha256      any    //
	IsDeleted   any    //
	CreatedAt   any    //
	UpdatedAt   any    //
	Content     []byte //
}
