package handlers

import (
	"errors"

	"memorylink/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	errInsufficientCoins = errors.New("硬币余额不足")
	errNotOwner          = errors.New("不是资源所有者")
	errStoryClosed       = errors.New("故事已结束")
	errNotPending        = errors.New("回应不在待确认状态")
)

// lockedForUpdate 返回 SELECT ... FOR UPDATE 行锁子句
func lockedForUpdate() clause.Expression {
	return clause.Locking{Strength: "UPDATE"}
}

// writeLedger 在事务内写一条硬币流水，balance_after 取该用户当前可用余额
func writeLedger(tx *gorm.DB, userID uint, change int, typ, remark string,
	storyID, responseID *uint) error {
	var user models.User
	if err := tx.Select("balance").First(&user, userID).Error; err != nil {
		return err
	}
	dir := models.CoinDirectionIn
	if change < 0 {
		dir = models.CoinDirectionOut
	}
	return tx.Create(&models.CoinLedger{
		UserID:       userID,
		Change:       change,
		Direction:    dir,
		BalanceAfter: user.Balance,
		Type:         typ,
		Remark:       remark,
		StoryID:      storyID,
		ResponseID:   responseID,
	}).Error
}
