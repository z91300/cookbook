// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// UserSessions is the golang structure of table user_sessions for DAO operations like Where/Data.
type UserSessions struct {
	g.Meta           `orm:"table:user_sessions, do:true"`
	Id               any //
	UserId           any //
	ExpiresAt        any //
	CreatedAt        any //
	UpdatedAt        any //
	AccessTokenHash  any //
	AccessExpiresAt  any //
	RefreshTokenHash any //
}
