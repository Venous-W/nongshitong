package db

// createTables 创建系统所需的 6 张表（若已存在则跳过）。
// 表结构说明：
//   - categories  : 分类树，parent_id=0 为顶级
//   - crops       : 作物/使用场所（统一存储，不区分农作物和场所）
//   - targets     : 防治对象，type 字段区分杂草/害虫/病害/调节剂
//   - products    : 农药产品，category_id 指向最深层分类
//   - product_crops   : 产品↔作物多对多关联
//   - product_targets : 产品↔防治对象多对多关联
func createTables() error {
	ddl := `
-- 分类表：支持多层树形结构，parent_id=0 表示顶级分类
CREATE TABLE IF NOT EXISTS categories (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id  INTEGER NOT NULL DEFAULT 0,
    name       TEXT    NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- 作物/场所表：农作物（大豆、玉米）和使用场所（果园、路边）统一存放
CREATE TABLE IF NOT EXISTS crops (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL UNIQUE,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- 防治对象表：type 取值 weed(杂草) / pest(害虫) / disease(病害) / regulator(调节剂)
CREATE TABLE IF NOT EXISTS targets (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL UNIQUE,
    type       TEXT    NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- 产品表：核心业务表，is_active=1 上架/0 下架
CREATE TABLE IF NOT EXISTS products (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    name        TEXT    NOT NULL,
    dosage      TEXT    NOT NULL DEFAULT '',
    usage       TEXT    NOT NULL DEFAULT '',
    notes       TEXT    NOT NULL DEFAULT '',
    is_active   INTEGER NOT NULL DEFAULT 1,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 产品-作物 多对多关联表
CREATE TABLE IF NOT EXISTS product_crops (
    product_id INTEGER NOT NULL,
    crop_id    INTEGER NOT NULL,
    PRIMARY KEY (product_id, crop_id)
);

-- 产品-防治对象 多对多关联表
CREATE TABLE IF NOT EXISTS product_targets (
    product_id INTEGER NOT NULL,
    target_id  INTEGER NOT NULL,
    PRIMARY KEY (product_id, target_id)
);
`
	_, err := DB.Exec(ddl)
	return err
}
