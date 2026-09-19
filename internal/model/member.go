// =================================================================================
// 家庭成员领域模型 —— 字段唯一定义点
// =================================================================================

package model

// MemberItem 成员项（列表返回形状）
type MemberItem struct {
	Id   int64  `json:"id"   dc:"成员 id"`
	Name string `json:"name" dc:"成员名字"`
	Role string `json:"role" dc:"角色（爸爸 妈妈 儿子 妻子 妹妹，可自定义）"`
	Note string `json:"note" dc:"备注（口味、忌口、过敏等）"`
	Sort int    `json:"sort" dc:"展示排序"`
}

// MemberIdInput 成员 id 入参
type MemberIdInput struct {
	Id int64 `json:"id" v:"required|min:1#成员id不能为空|成员id不合法" dc:"成员 id"`
}

// MemberSaveInput 成员新增/编辑入参
type MemberSaveInput struct {
	Name string `json:"name" v:"required|length:1,20#成员名字不能为空|成员名字最长20" dc:"成员名字"`
	Role string `json:"role" v:"length:0,20#角色最长20" dc:"角色（爸爸 妈妈 儿子 妻子 妹妹，可自定义）"`
	Note string `json:"note" v:"length:0,500#备注最长500" dc:"备注（口味、忌口、过敏等）"`
	Sort int   `json:"sort" dc:"展示排序"`
}
