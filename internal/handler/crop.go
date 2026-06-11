package handler

import (
	"nongshaitong/internal/repository"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetAllCrops 返回所有作物/场所列表。
// GET /api/crops
func GetAllCrops(c *gin.Context) {
	list, err := repository.GetAllCrops()
	if err != nil {
		apiFail(c, "查询作物失败: "+err.Error())
		return
	}
	apiOK(c, list)
}

// GetFilterCrops 根据分类动态返回"有产品关联"的作物列表，用于查询页步骤三的按钮渲染。
// GET /api/filter/crops?category_id=1
func GetFilterCrops(c *gin.Context) {
	categoryID := parseQueryInt64(c, "category_id", 0)
	list, err := repository.GetCropsByCategory(categoryID)
	if err != nil {
		apiFail(c, "查询作物失败: "+err.Error())
		return
	}
	apiOK(c, list)
}

// GetFilterTargets 根据分类+作物动态返回"有产品关联"的防治对象列表，用于查询页步骤四的按钮渲染。
// GET /api/filter/targets?category_id=1&crop_ids=2,3,5
func GetFilterTargets(c *gin.Context) {
	categoryID := parseQueryInt64(c, "category_id", 0)
	cropIDs := parseQueryInt64List(c, "crop_ids")

	list, err := repository.GetTargetsByCategoryAndCrops(categoryID, cropIDs)
	if err != nil {
		apiFail(c, "查询防治对象失败: "+err.Error())
		return
	}
	apiOK(c, list)
}

// CreateCrop 新增作物/场所。
// POST /api/crops
// Body: {"name":"大豆"}
func CreateCrop(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		apiFail(c, "参数错误：name 不能为空")
		return
	}
	id, err := repository.CreateCrop(body.Name)
	if err != nil {
		apiFail(c, "创建失败: "+err.Error())
		return
	}
	apiOK(c, gin.H{"id": id})
}

// UpdateCrop 修改作物/场所名称。
// PUT /api/crops/:id
func UpdateCrop(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		apiFail(c, "参数错误：name 不能为空")
		return
	}
	if err := repository.UpdateCrop(id, body.Name); err != nil {
		apiFail(c, "更新失败: "+err.Error())
		return
	}
	apiOK(c, nil)
}

// DeleteCrop 删除作物/场所（已被产品使用时拒绝）。
// DELETE /api/crops/:id
func DeleteCrop(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := repository.DeleteCrop(id); err != nil {
		apiFail(c, err.Error())
		return
	}
	apiOK(c, nil)
}

// ── 辅助解析函数 ──────────────────────────────────────────────────────────────

// parseQueryInt64 从 URL 查询参数中解析 int64，解析失败时返回默认值 def。
func parseQueryInt64(c *gin.Context, key string, def int64) int64 {
	s := c.Query(key)
	if s == "" {
		return def
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return v
}

// parseQueryInt64List 从 URL 查询参数中解析逗号分隔的 int64 列表。
// 例如 "crop_ids=1,2,3" 返回 []int64{1,2,3}。
func parseQueryInt64List(c *gin.Context, key string) []int64 {
	s := c.Query(key)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if v, err := strconv.ParseInt(p, 10, 64); err == nil && v > 0 {
			result = append(result, v)
		}
	}
	return result
}
