package handler

import (
	"nongshaitong/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetAllTargets 返回防治对象列表，可按 type 过滤。
// GET /api/targets?type=weed
func GetAllTargets(c *gin.Context) {
	typ := c.Query("type") // weed / pest / disease / 空=全部
	list, err := repository.GetAllTargets(typ)
	if err != nil {
		apiFail(c, "查询防治对象失败: "+err.Error())
		return
	}
	apiOK(c, list)
}

// CreateTarget 新增防治对象。
// POST /api/targets
// Body: {"name":"稗草","type":"weed"}
func CreateTarget(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" || body.Type == "" {
		apiFail(c, "参数错误：name 和 type 不能为空")
		return
	}
	validTypes := map[string]bool{"weed": true, "pest": true, "disease": true}
	if !validTypes[body.Type] {
		apiFail(c, "type 只能是 weed / pest / disease")
		return
	}

	id, err := repository.CreateTarget(body.Name, body.Type)
	if err != nil {
		apiFail(c, "创建失败: "+err.Error())
		return
	}
	apiOK(c, gin.H{"id": id})
}

// UpdateTarget 修改防治对象名称和类型。
// PUT /api/targets/:id
func UpdateTarget(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var body struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" || body.Type == "" {
		apiFail(c, "参数错误：name 和 type 不能为空")
		return
	}
	if err := repository.UpdateTarget(id, body.Name, body.Type); err != nil {
		apiFail(c, "更新失败: "+err.Error())
		return
	}
	apiOK(c, nil)
}

// DeleteTarget 删除防治对象（已被产品使用时拒绝）。
// DELETE /api/targets/:id
func DeleteTarget(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := repository.DeleteTarget(id); err != nil {
		apiFail(c, err.Error())
		return
	}
	apiOK(c, nil)
}
