package handler

import (
	"nongshaitong/internal/model"
	"nongshaitong/internal/repository"

	"github.com/gin-gonic/gin"
)

// ListProducts 鍒嗛〉鏌ヨ浜у搧鍒楄〃锛屾敮鎸佸鏉′欢绛涢€夈€?// GET /api/products?category_id=1&crop_ids=2,3&target_ids=4&keyword=鑽夌敇鑶?is_active=1&page=1&page_size=30
func ListProducts(c *gin.Context) {
	filter := repository.ProductFilter{
		CategoryID: parseQueryInt64(c, "category_id", 0),
		CropIDs:    parseQueryInt64List(c, "crop_ids"),
		TargetIDs:  parseQueryInt64List(c, "target_ids"),
		Keyword:    c.Query("keyword"),
		Page:       int(parseQueryInt64(c, "page", 1)),
		PageSize:   int(parseQueryInt64(c, "page_size", 30)),
	}

	// is_active: 1=浠呬笂鏋?/ 0=浠呬笅鏋?/ 绌?鍏ㄩ儴(-1)
	switch c.Query("is_active") {
	case "1":
		filter.IsActive = 1
	case "0":
		filter.IsActive = 0
	default:
		filter.IsActive = -1
	}

	result, err := repository.ListProducts(filter)
	if err != nil {
		apiFail(c, "鏌ヨ浜у搧澶辫触: "+err.Error())
		return
	}
	apiOK(c, result)
}

// GetProductDetail 鏌ヨ鍗曚釜浜у搧璇︽儏锛堝惈瀹屾暣浣滅墿鍜岄槻娌诲璞″垪琛級銆?// GET /api/products/:id
func GetProductDetail(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	p, err := repository.GetProductDetail(id)
	if err != nil {
		apiFail(c, err.Error())
		return
	}
	apiOK(c, p)
}

// productBody 产品信息
type productBody struct {
	CategoryID int64   `json:"category_id"`
	Name       string  `json:"name"`
	Dosage     string  `json:"dosage"`
	Usage      string  `json:"usage"`
	Notes      string  `json:"notes"`
	IsActive   *int    `json:"is_active"` // 是否上架
	CropIDs    []int64 `json:"crop_ids"`
	TargetIDs  []int64 `json:"target_ids"`
}

// CreateProduct 创建产品信息// POST /api/products
func CreateProduct(c *gin.Context) {
	var body productBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apiFail(c, "参数错误:"+err.Error())
		return
	}
	if body.Name == "" {
		apiFail(c, "产品名称不允许为空")
		return
	}
	if body.CategoryID <= 0 {
		apiFail(c, "分类ID不允许为空")
		return
	}

	isActive := 1
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	p := model.Product{
		CategoryID: body.CategoryID,
		Name:       body.Name,
		Dosage:     body.Dosage,
		Usage:      body.Usage,
		Notes:      body.Notes,
		IsActive:   isActive,
	}
	if body.CropIDs == nil {
		body.CropIDs = []int64{}
	}
	if body.TargetIDs == nil {
		body.TargetIDs = []int64{}
	}

	id, err := repository.CreateProduct(p, body.CropIDs, body.TargetIDs)
	if err != nil {
		apiFail(c, "创建产品失败: "+err.Error())
		return
	}
	apiOK(c, gin.H{"id": id})
}

// UpdateProduct 修改产品信息// PUT /api/products/:id
func UpdateProduct(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var body productBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apiFail(c, "传递参数错误: "+err.Error())
		return
	}
	if body.Name == "" {
		apiFail(c, "名称不允许为空")
		return
	}

	isActive := 1
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	p := model.Product{
		ID:         id,
		CategoryID: body.CategoryID,
		Name:       body.Name,
		Dosage:     body.Dosage,
		Usage:      body.Usage,
		Notes:      body.Notes,
		IsActive:   isActive,
	}
	if body.CropIDs == nil {
		body.CropIDs = []int64{}
	}
	if body.TargetIDs == nil {
		body.TargetIDs = []int64{}
	}

	if err := repository.UpdateProduct(p, body.CropIDs, body.TargetIDs); err != nil {
		apiFail(c, "修改产品失败: "+err.Error())
		return
	}
	apiOK(c, nil)
}

// DeleteProduct 删除产品// DELETE /api/products/:id
func DeleteProduct(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := repository.DeleteProduct(id); err != nil {
		apiFail(c, err.Error())
		return
	}
	apiOK(c, nil)
}

// ToggleActive 修改产品状态// PUT /api/products/:id/active
// Body: {"is_active":0} 或 {"is_active":1}
func ToggleActive(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var body struct {
		IsActive int `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		apiFail(c, "非法参数")
		return
	}
	// is_active 设置产品上架状态
	isActive := 0
	if body.IsActive == 1 {
		isActive = 1
	}

	if err := repository.ToggleActive(id, isActive); err != nil {
		apiFail(c, "修改产品状态失败: "+err.Error())
		return
	}
	apiOK(c, gin.H{"is_active": isActive})
}
