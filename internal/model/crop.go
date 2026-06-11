package model

// Crop 表示一个"作物/场所"选项，统一存储农作物（大豆、玉米）和使用场所（果园、路边）。
// 查询页第三步"选作物"的按钮数据即来自此表。
type Crop struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`      // 如"大豆"、"果园"、"草坪"
	SortOrder int    `json:"sort_order"` // 显示顺序
}
