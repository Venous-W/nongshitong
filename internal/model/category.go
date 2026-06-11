package model

// Category 表示农药品类分类节点。
// 分类支持多层嵌套（parent_id=0 表示一级分类，否则指向父节点 id）。
// 例如：除草（一级）-> 苗前除草剂（二级）-> 产品挂在二级 id 上。
type Category struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parent_id"` // 0 表示顶级分类，>0 指向父分类 id
	Name     string `json:"name"`
	SortOrder int   `json:"sort_order"` // 显示顺序，按插入顺序（数值越小越靠前）
}

// CategoryTree 是带子节点的分类树结构，用于前端一次性加载整棵树。
type CategoryTree struct {
	Category
	Children []*CategoryTree `json:"children"`
}
