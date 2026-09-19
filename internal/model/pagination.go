// =================================================================================
// 分页通用模型 —— 入参嵌入复用、响应用泛型别名复用
// 用法：
//   入参：领域查询输入直接嵌入 PaginationInput
//   出参：领域输出定义别名 type XxxListOutput = PageRes[XxxItem]
// =================================================================================

package model

// PaginationInput 分页入参
type PaginationInput struct {
	Page     int `json:"page"     d:"1"  dc:"页码，从 1 开始" v:"min:0#分页码不能为负"`
	PageSize int `json:"pageSize" d:"20" dc:"每页数量" v:"min:1|max:100#每页数量最小1|每页数量最大100"`
}

// PageRes 分页响应
type PageRes[T any] struct {
	List  []T `json:"list"  dc:"当前页数据（空页为 []）"` // 当前页数据（保证非 nil，空页序列化为 []）
	Total int `json:"total" dc:"过滤后总条数"`        // 过滤后总条数
}

// NewPageRes 构造分页响应（list 保证非 nil）
func NewPageRes[T any](total int, list []T) *PageRes[T] {
	if list == nil {
		list = make([]T, 0)
	}
	return &PageRes[T]{List: list, Total: total}
}
