package repository

import (
	"nongshaitong/internal/db"
	"nongshaitong/internal/model"
)

// GetAllTargets 返回所有防治对象，可按 type 过滤（传空字符串则不过滤）。
// type 取值：weed=杂草 / pest=害虫 / disease=病害
func GetAllTargets(typ string) ([]model.Target, error) {
	query := `SELECT id, name, type, sort_order FROM targets WHERE 1=1`
	var args []interface{}
	if typ != "" {
		query += ` AND type=?`
		args = append(args, typ)
	}
	query += ` ORDER BY type ASC, sort_order ASC, id ASC`

	rows, err := db.DB.Query(query, args...)
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

// CreateTarget 新增一个防治对象。typ 必须是 weed/pest/disease 之一。
func CreateTarget(name, typ string) (int64, error) {
	var maxOrder int
	db.DB.QueryRow(`SELECT COALESCE(MAX(sort_order), -1) FROM targets`).Scan(&maxOrder)

	res, err := db.DB.Exec(
		`INSERT INTO targets (name, type, sort_order) VALUES (?, ?, ?)`,
		name, typ, maxOrder+1,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateTarget 修改防治对象名称和类型。
func UpdateTarget(id int64, name, typ string) error {
	_, err := db.DB.Exec(`UPDATE targets SET name=?, type=? WHERE id=?`, name, typ, id)
	return err
}

// DeleteTarget 删除防治对象。若已有产品关联则拒绝删除。
func DeleteTarget(id int64) error {
	var count int
	db.DB.QueryRow(`SELECT COUNT(*) FROM product_targets WHERE target_id=?`, id).Scan(&count)
	if count > 0 {
		return errorf("该防治对象已被 %d 个产品使用，无法删除", count)
	}
	_, err := db.DB.Exec(`DELETE FROM targets WHERE id=?`, id)
	return err
}
