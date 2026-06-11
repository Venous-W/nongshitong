package repository

import (
	"fmt"
	"strings"
)

// errorf 是本包内部用的简单错误构造函数，避免引入 errors 包。
func errorf(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// makePlaceholders 生成 n 个 ? 占位符，以逗号拼接，用于 SQL IN 子句。
// 例如 makePlaceholders(3) 返回 "?,?,?"
func makePlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?,", n)[:n*2-1]
}
