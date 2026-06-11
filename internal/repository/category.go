package repository

import (
	"nongshaitong/internal/db"
	"nongshaitong/internal/model"
)

// GetCategoryTree 查询所有分类，并在内存中组装成树形结构返回。
// 返回的是顶级分类列表，每个节点下挂 Children 子节点列表。
// 前端一次性加载整棵树，避免多次请求。
func GetCategoryTree() ([]*model.CategoryTree, error) {
	rows, err := db.DB.Query(
		`SELECT id, parent_id, name, sort_order FROM categories ORDER BY sort_order ASC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 先把所有节点存成 map，key=id
	nodeMap := make(map[int64]*model.CategoryTree)
	var all []*model.CategoryTree
	for rows.Next() {
		node := &model.CategoryTree{}
		if err = rows.Scan(&node.ID, &node.ParentID, &node.Name, &node.SortOrder); err != nil {
			return nil, err
		}
		node.Children = []*model.CategoryTree{}
		nodeMap[node.ID] = node
		all = append(all, node)
	}

	// 再遍历一次，把子节点挂到父节点上
	var roots []*model.CategoryTree
	for _, node := range all {
		if node.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[node.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		}
	}
	return roots, nil
}

// GetCategories 返回平铺的所有分类列表（不组装树），供管理台列表展示使用。
func GetCategories() ([]model.Category, error) {
	rows, err := db.DB.Query(
		`SELECT id, parent_id, name, sort_order FROM categories ORDER BY sort_order ASC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Category
	for rows.Next() {
		var c model.Category
		if err = rows.Scan(&c.ID, &c.ParentID, &c.Name, &c.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

// CreateCategory 新增一个分类节点。sort_order 自动取当前最大值+1，保证按插入顺序排列。
func CreateCategory(parentID int64, name string) (int64, error) {
	// 取当前最大 sort_order，新节点排在最后
	var maxOrder int
	db.DB.QueryRow(`SELECT COALESCE(MAX(sort_order), -1) FROM categories`).Scan(&maxOrder)

	res, err := db.DB.Exec(
		`INSERT INTO categories (parent_id, name, sort_order) VALUES (?, ?, ?)`,
		parentID, name, maxOrder+1,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateCategory 修改分类名称（不支持移动父级，避免数据混乱）。
func UpdateCategory(id int64, name string) error {
	_, err := db.DB.Exec(`UPDATE categories SET name=? WHERE id=?`, name, id)
	return err
}

// DeleteCategory 删除分类。若该分类下已有产品，则拒绝删除并返回错误。
// 删除前检查直接子分类，若有子分类也拒绝（需先删子分类）。
func DeleteCategory(id int64) error {
	// 检查是否有子分类
	var childCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM categories WHERE parent_id=?`, id).Scan(&childCount)
	if childCount > 0 {
		return errorf("该分类下还有 %d 个子分类，请先删除子分类", childCount)
	}

	// 检查是否有关联产品
	var productCount int
	db.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE category_id=?`, id).Scan(&productCount)
	if productCount > 0 {
		return errorf("该分类下还有 %d 个产品，请先处理产品后再删除", productCount)
	}

	_, err := db.DB.Exec(`DELETE FROM categories WHERE id=?`, id)
	return err
}
