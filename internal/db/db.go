package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // 注册 sqlite 驱动（纯 Go，无需 CGO）
)

// DB 是全局数据库连接，由 Init() 初始化后供各 repository 使用。
var DB *sql.DB

// Init 打开（或创建）SQLite 数据库文件，建表，并写入初始种子数据。
// 程序启动时调用一次即可；若 data.db 文件已存在，建表语句因使用
// IF NOT EXISTS 不会重复执行，种子数据因使用 INSERT OR IGNORE 也不会重复插入。
func Init(dsn string) {
	var err error
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}

	// SQLite 不支持真正的并发写入，限制为单一连接避免 SQLITE_BUSY 错误。
	DB.SetMaxOpenConns(1)

	if err = createTables(); err != nil {
		log.Fatalf("建表失败: %v", err)
	}

	if err = seedData(); err != nil {
		log.Fatalf("初始化种子数据失败: %v", err)
	}
}
