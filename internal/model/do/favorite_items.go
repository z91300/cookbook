// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// FavoriteItems is the golang structure of table favorite_items for DAO operations like Where/Data.
type FavoriteItems struct {
	g.Meta     `orm:"table:favorite_items, do:true"`
	FavoriteId any //
	RecipeId   any //
	Note       any //
	CreatedAt  any //
}
