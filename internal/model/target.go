package model

// Target 表示防治对象（杂草/害虫/病害）。
// Type 字段将防治对象分为三类，管理台和查询页按类型分组展示。
type Target struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"` // 如"稗草"、"蚜虫"、"白粉病"
	Type      string `json:"type"` // weed=杂草 / pest=害虫 / disease=病害
	SortOrder int    `json:"sort_order"`
}
