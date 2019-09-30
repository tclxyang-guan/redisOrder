package utils

import (
	"fmt"
	"github.com/jinzhu/gorm"
)

func Defer(tx *gorm.DB, status *bool) {
	if *status {
		//提交事务
		fmt.Print("commit")
		tx.Commit()
	} else {
		//回滚
		tx.Rollback()
	}
}
