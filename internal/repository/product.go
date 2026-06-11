package repository

import (
	"database/sql"
	"nongshaitong/internal/db"
	"nongshaitong/internal/model"
	"strings"
)

// ProductFilter 是查询产品列表时支持的筛选条件。
// 各字段均为可选：零值/空值表示不过滤该条件。
type ProductFilter struct {
	CategoryID int64   // 按分类过滤（含子分类）
	CropIDs    []int64 // 按作物过滤（OR 逻辑，产品包含任一作物即匹配）
	TargetIDs  []int64 // 按防治对象过滤（OR 逻辑）
	Keyword    string  // 按产品名称关键词模糊搜索
	IsActive   int     // 1=仅上架 / 0=仅下架 / -1=全部（查询页传1，管理台传-1）
	Page       int     // 当前页码，从 1 开始
	PageSize   int     // 每页条数，默认 30
}

// ProductListResult 是分页查询结果，包含列表数据和总条数。
type ProductListResult struct {
	Total int                    `json:"total"`
	List  []model.ProductListItem `json:"list"`
}

// ListProducts 根据筛选条件分页查询产品列表，并聚合每个产品的作物和防治对象。
// 查询页传 IsActive=1 只看上架产品；管理台传 IsActive=-1 看全部。
func ListProducts(f ProductFilter) (*ProductListResult, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 30
	}

	// ── 构造 WHERE 子句 ──────────────────────────────────────────────────────
	var conds []string
	var args []interface{}

	// 分类过滤：匹配该分类本身或其直接子分类
	if f.CategoryID > 0 {
		conds = append(conds,
			`(p.category_id = ? OR p.category_id IN (SELECT id FROM categories WHERE parent_id = ?))`)
		args = append(args, f.CategoryID, f.CategoryID)
	}

	// 是否上架过滤
	if f.IsActive == 1 {
		conds = append(conds, `p.is_active = 1`)
	} else if f.IsActive == 0 {
		conds = append(conds, `p.is_active = 0`)
	}

	// 关键词模糊搜索产品名
	if f.Keyword != "" {
		conds = append(conds, `p.name LIKE ?`)
		args = append(args, "%"+f.Keyword+"%")
	}

	// 作物过滤（OR）：产品必须关联 cropIDs 中的至少一个
	if len(f.CropIDs) > 0 {
		ph := makePlaceholders(len(f.CropIDs))
		conds = append(conds,
			`p.id IN (SELECT product_id FROM product_crops WHERE crop_id IN (`+ph+`))`)
		for _, id := range f.CropIDs {
			args = append(args, id)
		}
	}

	// 防治对象过滤（OR）：产品必须关联 targetIDs 中的至少一个
	if len(f.TargetIDs) > 0 {
		ph := makePlaceholders(len(f.TargetIDs))
		conds = append(conds,
			`p.id IN (SELECT product_id FROM product_targets WHERE target_id IN (`+ph+`))`)
		for _, id := range f.TargetIDs {
			args = append(args, id)
		}
	}

	whereSQL := ""
	if len(conds) > 0 {
		whereSQL = "WHERE " + strings.Join(conds, " AND ")
	}

	// ── 查询总条数 ───────────────────────────────────────────────────────────
	countSQL := `SELECT COUNT(*) FROM products p ` + whereSQL
	var total int
	if err := db.DB.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, err
	}

	// ── 查询分页产品列表（含分类名） ─────────────────────────────────────────
	offset := (f.Page - 1) * f.PageSize
	listSQL := `
		SELECT p.id, p.category_id, p.name, p.dosage, p.is_active,
		       COALESCE(c.name, '') AS cat_name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		` + whereSQL + `
		ORDER BY p.id DESC
		LIMIT ? OFFSET ?`
	listArgs := append(args, f.PageSize, offset)

	rows, err := db.DB.Query(listSQL, listArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ProductListItem
	var ids []int64
	for rows.Next() {
		var item model.ProductListItem
		if err = rows.Scan(&item.ID, &item.CategoryID, &item.Name,
			&item.Dosage, &item.IsActive, &item.CategoryName); err != nil {
			return nil, err
		}
		item.Crops = []model.Crop{}
		item.Targets = []model.Target{}
		items = append(items, item)
		ids = append(ids, item.ID)
	}

	if len(ids) == 0 {
		return &ProductListResult{Total: total, List: []model.ProductListItem{}}, nil
	}

	// ── 批量查询这批产品的作物和防治对象，避免 N+1 查询 ─────────────────────
	if err = attachCropsToItems(items, ids); err != nil {
		return nil, err
	}
	if err = attachTargetsToItems(items, ids); err != nil {
		return nil, err
	}

	return &ProductListResult{Total: total, List: items}, nil
}

// GetProductDetail 查询单个产品的完整信息（含作物、防治对象、分类名）。
// 管理台"查看/编辑"和查询页"展开详情"均使用此接口。
func GetProductDetail(id int64) (*model.Product, error) {
	var p model.Product
	err := db.DB.QueryRow(`
		SELECT p.id, p.category_id, p.name, p.dosage, p.usage, p.notes, p.is_active,
		       p.created_at, p.updated_at,
		       COALESCE(c.name, '') AS cat_name
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.id = ?`, id,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Dosage, &p.Usage, &p.Notes,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName)
	if err == sql.ErrNoRows {
		return nil, errorf("产品不存在")
	}
	if err != nil {
		return nil, err
	}

	// 查询关联作物
	cropRows, err := db.DB.Query(`
		SELECT cr.id, cr.name, cr.sort_order
		FROM crops cr
		JOIN product_crops pc ON pc.crop_id = cr.id
		WHERE pc.product_id = ?
		ORDER BY cr.sort_order ASC`, id)
	if err != nil {
		return nil, err
	}
	defer cropRows.Close()
	for cropRows.Next() {
		var c model.Crop
		cropRows.Scan(&c.ID, &c.Name, &c.SortOrder)
		p.Crops = append(p.Crops, c)
	}
	if p.Crops == nil {
		p.Crops = []model.Crop{}
	}

	// 查询关联防治对象
	tgtRows, err := db.DB.Query(`
		SELECT t.id, t.name, t.type, t.sort_order
		FROM targets t
		JOIN product_targets pt ON pt.target_id = t.id
		WHERE pt.product_id = ?
		ORDER BY t.type ASC, t.sort_order ASC`, id)
	if err != nil {
		return nil, err
	}
	defer tgtRows.Close()
	for tgtRows.Next() {
		var t model.Target
		tgtRows.Scan(&t.ID, &t.Name, &t.Type, &t.SortOrder)
		p.Targets = append(p.Targets, t)
	}
	if p.Targets == nil {
		p.Targets = []model.Target{}
	}

	return &p, nil
}

// CreateProduct 新增产品，同时写入作物和防治对象的关联关系。
// cropIDs 和 targetIDs 可以为空切片（表示不关联任何作物/防治对象）。
func CreateProduct(p model.Product, cropIDs, targetIDs []int64) (int64, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	res, err := tx.Exec(`
		INSERT INTO products (category_id, name, dosage, usage, notes, is_active)
		VALUES (?, ?, ?, ?, ?, ?)`,
		p.CategoryID, p.Name, p.Dosage, p.Usage, p.Notes, p.IsActive,
	)
	if err != nil {
		return 0, err
	}
	productID, _ := res.LastInsertId()

	if err = insertCropRels(tx, productID, cropIDs); err != nil {
		return 0, err
	}
	if err = insertTargetRels(tx, productID, targetIDs); err != nil {
		return 0, err
	}

	return productID, tx.Commit()
}

// UpdateProduct 更新产品基本信息，并重新绑定作物和防治对象（先删后插）。
func UpdateProduct(p model.Product, cropIDs, targetIDs []int64) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		UPDATE products
		SET category_id=?, name=?, dosage=?, usage=?, notes=?, is_active=?,
		    updated_at=CURRENT_TIMESTAMP
		WHERE id=?`,
		p.CategoryID, p.Name, p.Dosage, p.Usage, p.Notes, p.IsActive, p.ID,
	)
	if err != nil {
		return err
	}

	// 清除旧关联，重新插入新关联
	if _, err = tx.Exec(`DELETE FROM product_crops WHERE product_id=?`, p.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM product_targets WHERE product_id=?`, p.ID); err != nil {
		return err
	}
	if err = insertCropRels(tx, p.ID, cropIDs); err != nil {
		return err
	}
	if err = insertTargetRels(tx, p.ID, targetIDs); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteProduct 删除产品及其所有关联关系（级联删除 product_crops / product_targets）。
func DeleteProduct(id int64) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM product_crops WHERE product_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM product_targets WHERE product_id=?`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM products WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

// ToggleActive 切换产品上架/下架状态（is_active 0↔1）。
func ToggleActive(id int64, isActive int) error {
	_, err := db.DB.Exec(
		`UPDATE products SET is_active=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		isActive, id,
	)
	return err
}

// ── 内部辅助函数 ──────────────────────────────────────────────────────────────

// insertCropRels 在事务中批量插入产品-作物关联。
func insertCropRels(tx *sql.Tx, productID int64, cropIDs []int64) error {
	for _, cropID := range cropIDs {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO product_crops (product_id, crop_id) VALUES (?, ?)`,
			productID, cropID,
		); err != nil {
			return err
		}
	}
	return nil
}

// insertTargetRels 在事务中批量插入产品-防治对象关联。
func insertTargetRels(tx *sql.Tx, productID int64, targetIDs []int64) error {
	for _, targetID := range targetIDs {
		if _, err := tx.Exec(
			`INSERT OR IGNORE INTO product_targets (product_id, target_id) VALUES (?, ?)`,
			productID, targetID,
		); err != nil {
			return err
		}
	}
	return nil
}

// attachCropsToItems 批量查询 ids 列表对应产品的作物，填充到 items 切片中。
// 使用 IN 查询避免循环查库（N+1 问题）。
func attachCropsToItems(items []model.ProductListItem, ids []int64) error {
	ph := makePlaceholders(len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := db.DB.Query(`
		SELECT pc.product_id, cr.id, cr.name, cr.sort_order
		FROM product_crops pc
		JOIN crops cr ON cr.id = pc.crop_id
		WHERE pc.product_id IN (`+ph+`)
		ORDER BY cr.sort_order ASC`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// 建立 productID -> index 映射，方便填充
	indexMap := make(map[int64]int, len(items))
	for i, item := range items {
		indexMap[item.ID] = i
	}

	for rows.Next() {
		var pid int64
		var c model.Crop
		if err = rows.Scan(&pid, &c.ID, &c.Name, &c.SortOrder); err != nil {
			return err
		}
		if idx, ok := indexMap[pid]; ok {
			items[idx].Crops = append(items[idx].Crops, c)
		}
	}
	return nil
}

// attachTargetsToItems 批量查询 ids 列表对应产品的防治对象，填充到 items 切片中。
func attachTargetsToItems(items []model.ProductListItem, ids []int64) error {
	ph := makePlaceholders(len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	rows, err := db.DB.Query(`
		SELECT pt.product_id, t.id, t.name, t.type, t.sort_order
		FROM product_targets pt
		JOIN targets t ON t.id = pt.target_id
		WHERE pt.product_id IN (`+ph+`)
		ORDER BY t.type ASC, t.sort_order ASC`, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	indexMap := make(map[int64]int, len(items))
	for i, item := range items {
		indexMap[item.ID] = i
	}

	for rows.Next() {
		var pid int64
		var t model.Target
		if err = rows.Scan(&pid, &t.ID, &t.Name, &t.Type, &t.SortOrder); err != nil {
			return err
		}
		if idx, ok := indexMap[pid]; ok {
			items[idx].Targets = append(items[idx].Targets, t)
		}
	}
	return nil
}
