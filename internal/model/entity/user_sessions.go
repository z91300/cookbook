// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// UserSessions is the golang structure for table user_sessions.
type UserSessions struct {
	Id               int    `json:"id"               orm:"id"                 description:""` //
	UserId           int    `json:"userId"           orm:"user_id"            description:""` //
	ExpiresAt        int    `json:"expiresAt"        orm:"expires_at"         description:""` //
	CreatedAt        int    `json:"createdAt"        orm:"created_at"         description:""` //
	UpdatedAt        int    `json:"updatedAt"        orm:"updated_at"         description:""` //
	AccessTokenHash  string `json:"accessTokenHash"  orm:"access_token_hash"  description:""` //
	AccessExpiresAt  int    `json:"accessExpiresAt"  orm:"access_expires_at"  description:""` //
	RefreshTokenHash string `json:"refreshTokenHash" orm:"refresh_token_hash" description:""` //
}
