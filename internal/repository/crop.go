package repository

import (
	"nongshaitong/internal/db"
	"nongshaitong/internal/model"
)

// GetAllCrops 返回所有作物/场所列表，按 sort_order 升序排列。
// 查询页"选作物"步骤和管理台的作物列表均使用此函数。
func GetAllCrops() ([]model.Crop, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, sort_order FROM crops ORDER BY sort_order ASC, id ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Crop
	for rows.Next() {
		var c model.Crop
		if err = rows.Scan(&c.ID, &c.Name, &c.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

// GetCropsByCategory 返回在指定分类下有产品关联的作物/场所列表，用于查询页动态加载按钮。
// 逻辑：先找该分类（及其所有子分类）下的上架产品，再取这些产品关联的作物去重。
// categoryID 传 0 时返回所有有产品的作物。
func GetCropsByCategory(categoryID int64) ([]model.Crop, error) {
	var query string
	var args []interface{}

	if categoryID == 0 {
		// 不限分类，取所有上架产品关联的作物
		query = `
			SELECT DISTINCT c.id, c.name, c.sort_order
			FROM crops c
			JOIN product_crops pc ON pc.crop_id = c.id
			JOIN products p ON p.id = pc.product_id AND p.is_active = 1
			ORDER BY c.sort_order ASC, c.id ASC`
	} else {
		// 先收集该 categoryID 以及其所有子分类 id，再查关联作物
		query = `
			SELECT DISTINCT c.id, c.name, c.sort_order
			FROM crops c
			JOIN product_crops pc ON pc.crop_id = c.id
			JOIN products p ON p.id = pc.product_id AND p.is_active = 1
			WHERE p.category_id = ?
			   OR p.category_id IN (SELECT id FROM categories WHERE parent_id = ?)
			ORDER BY c.sort_order ASC, c.id ASC`
		args = append(args, categoryID, categoryID)
	}

	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Crop
	for rows.Next() {
		var c model.Crop
		if err = rows.Scan(&c.ID, &c.Name, &c.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

// GetTargetsByCategoryAndCrops 返回在指定分类+作物组合下有产品关联的防治对象列表。
// 用于查询页"选防治对象"步骤的动态加载。
// cropIDs 为空时只按分类过滤；categoryID 为 0 时不过滤分类。
// 多个 cropID 之间是 OR 关系（产品关联任一作物即可）。
func GetTargetsByCategoryAndCrops(categoryID int64, cropIDs []int64) ([]model.Target, error) {
	// 构造 IN 子句占位符
	cropFilter := ""
	var args []interface{}

	if categoryID > 0 {
		args = append(args, categoryID, categoryID)
	}

	if len(cropIDs) > 0 {
		placeholders := makePlaceholders(len(cropIDs))
		cropFilter = `AND pc.crop_id IN (` + placeholders + `)`
		for _, id := range cropIDs {
			args = append(args, id)
		}
	}

	catFilter := ""
	if categoryID > 0 {
		catFilter = `AND (p.category_id = ? OR p.category_id IN (SELECT id FROM categories WHERE parent_id = ?))`
	}

	query := `
		SELECT DISTINCT t.id, t.name, t.type, t.sort_order
		FROM targets t
		JOIN product_targets pt ON pt.target_id = t.id
		JOIN products p ON p.id = pt.product_id AND p.is_active = 1
		` + func() string {
		if len(cropIDs) > 0 {
			return `JOIN product_crops pc ON pc.product_id = p.id ` + cropFilter
		}
		return ""
	}() + `
		WHERE 1=1 ` + catFilter + `
		ORDER BY t.type ASC, t.sort_order ASC, t.id ASC`

	// 重新整理参数顺序（cropIDs 在前，catFilter 在后）
	var finalArgs []interface{}
	if len(cropIDs) > 0 {
		for _, id := range cropIDs {
			finalArgs = append(finalArgs, id)
		}
	}
	if categoryID > 0 {
		finalArgs = append(finalArgs, categoryID, categoryID)
	}

	rows, err := db.DB.Query(query, finalArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.Target
	for rows.Next() {
		var t model.Target
		if err = rows.Scan(&t.ID, &t.Name, &t.Type, &t.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

// CreateCrop 新增一个作物/场所。sort_order 自动按插入顺序。
func CreateCrop(name string) (int64, error) {
	var maxOrder int
	db.DB.QueryRow(`SELECT COALESCE(MAX(sort_order), -1) FROM crops`).Scan(&maxOrder)

	res, err := db.DB.Exec(
		`INSERT INTO crops (name, sort_order) VALUES (?, ?)`, name, maxOrder+1,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateCrop 修改作物/场所名称。
func UpdateCrop(id int64, name string) error {
	_, err := db.DB.Exec(`UPDATE crops SET name=? WHERE id=?`, name, id)
	return err
}

// DeleteCrop 删除作物/场所。若已有产品关联则拒绝删除。
func DeleteCrop(id int64) error {
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM product_crops WHERE crop_id=?`, id).Scan(&count)
	if count > 0 {
		return errorf("该作物/场所已被 %d 个产品使用，无法删除", count)
	}
	_, err := db.DB.Exec(`DELETE FROM crops WHERE id=?`, id)
	return err
}
