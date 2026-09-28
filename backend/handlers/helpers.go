package handlers

import "strings"

// isDuplicateKey 判断是否为唯一索引冲突（MySQL 1062）
func isDuplicateKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "1062")
}
