// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Favorites is the golang structure of table favorites for DAO operations like Where/Data.
type Favorites struct {
	g.Meta            `orm:"table:favorites, do:true"`
	Id                any //
	UserId            any //
	Name              any //
	Description       any //
	CoverAttachmentId any //
	IsPublic          any //
	Sort              any //
	IsDeleted         any //
	CreatedAt         any //
	UpdatedAt         any //
}
