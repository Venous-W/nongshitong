package model

import "time"

// Product 表示一个农药产品，包含基本信息及关联的作物/防治对象。
// CategoryID 指向该产品所属的最深层分类（若分类有子级则存子级 id）。
// Crops 和 Targets 是查询时从关联表聚合的，不直接存储在 products 表中。
type Product struct {
	ID         int64     `json:"id"`
	CategoryID int64     `json:"category_id"` // 所属最深层分类的 id
	Name       string    `json:"name"`         // 格式：含量+成分+剂型，如"33%草甘膦 水剂"
	Dosage     string    `json:"dosage"`       // 用量，如"200ml/桶水"
	Usage      string    `json:"usage"`        // 使用方式/施药时机说明
	Notes      string    `json:"notes"`        // 重点及注意事项
	IsActive   int       `json:"is_active"`    // 1=上架（查询页可见）/ 0=下架
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 以下字段在 JOIN 查询后填充，数据库表中不存在这些列
	CategoryName string   `json:"category_name,omitempty"` // 分类名（含父级路径，如"除草 > 苗前除草剂"）
	Crops        []Crop   `json:"crops,omitempty"`         // 适用作物/场所列表
	Targets      []Target `json:"targets,omitempty"`       // 防治对象列表
}

// ProductListItem 是列表页的精简视图，减少不必要的数据传输。
type ProductListItem struct {
	ID           int64    `json:"id"`
	CategoryID   int64    `json:"category_id"`
	CategoryName string   `json:"category_name"`
	Name         string   `json:"name"`
	Dosage       string   `json:"dosage"`
	IsActive     int      `json:"is_active"`
	Crops        []Crop   `json:"crops"`
	Targets      []Target `json:"targets"`
}
