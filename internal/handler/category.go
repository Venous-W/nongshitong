package handler

import (
	"nongshaitong/internal/repository"

	"github.com/gin-gonic/gin"
)

// GetCategoryTree 返回完整分类树（嵌套结构），前端一次加载。
// GET /api/categories
func GetCategoryTree(c *gin.Context) {
	tree, err := repository.GetCategoryTree()
	if err != nil {
		apiFail(c, "查询分类失败: "+err.Error())
		return
	}
	apiOK(c, tree)
}

// GetCategories 返回平铺分类列表，供管理台下拉/列表使用。
// GET /api/categories/flat
func GetCategories(c *gin.Context) {
	list, err := repository.GetCategories()
	if err != nil {
		apiFail(c, "查询分类失败: "+err.Error())
		return
	}
	apiOK(c, list)
}

// CreateCategory 新增分类。
// POST /api/categories
// Body: {"parent_id":0,"name":"除草"}
func CreateCategory(c *gin.Context) {
	var body struct {
		ParentID int64  `json:"parent_id"`
		Name     string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		apiFail(c, "参数错误：name 不能为空")
		return
	}

	id, err := repository.CreateCategory(body.ParentID, body.Name)
	if err != nil {
		apiFail(c, "创建分类失败: "+err.Error())
		return
	}
	apiOK(c, gin.H{"id": id})
}

// UpdateCategory 修改分类名称。
// PUT /api/categories/:id
// Body: {"name":"新名称"}
func UpdateCategory(c *gin.Context) {
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

	if err := repository.UpdateCategory(id, body.Name); err != nil {
		apiFail(c, "更新分类失败: "+err.Error())
		return
	}
	apiOK(c, nil)
}

// DeleteCategory 删除分类（有子分类或产品时拒绝）。
// DELETE /api/categories/:id
func DeleteCategory(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := repository.DeleteCategory(id); err != nil {
		apiFail(c, err.Error())
		return
	}
	apiOK(c, nil)
}
