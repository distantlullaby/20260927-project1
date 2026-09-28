package services

import (
	"errors"

	"memoryconnect/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 硬币流水类型
const (
	OpRegister = "register" // 注册赠送
	OpFreeze   = "freeze"   // 发布求看冻结
	OpAppend   = "append"   // 追加悬赏冻结
	OpPay      = "pay"      // 采纳时从冻结中支出
	OpIncome   = "income"   // 代看被采纳收入
)

// 业务错误，handler 层据此返回友好提示
var (
	ErrInvalidAmount = errors.New("硬币数量必须大于 0")
	ErrInsufficient  = errors.New("可用硬币不足")
	ErrFrozenShort   = errors.New("冻结硬币不足，无法结算")
	ErrSelfSettle    = errors.New("不能结算给自己")
)

// lockUser 在事务内以 SELECT ... FOR UPDATE 锁定用户行
func lockUser(tx *gorm.DB, id uint) (*models.User, error) {
	var u models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func writeLedger(tx *gorm.DB, u *models.User, change int, op string, storyID, responseID, refUserID *uint, remark string) error {
	l := models.CoinLedger{
		UserID:       u.ID,
		Change:       change,
		BalanceAfter: u.Balance,
		FrozenAfter:  u.Frozen,
		Type:         op,
		StoryID:      storyID,
		ResponseID:   responseID,
		RefUserID:    refUserID,
		Remark:       remark,
	}
	return tx.Create(&l).Error
}

// GrantInitial 注册赠送初始硬币（事务内调用）
func GrantInitial(tx *gorm.DB, userID uint, amount int) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	u, err := lockUser(tx, userID)
	if err != nil {
		return err
	}
	u.Balance += amount
	if err := tx.Save(u).Error; err != nil {
		return err
	}
	return writeLedger(tx, u, amount, OpRegister, nil, nil, nil, "注册赠送记忆硬币")
}

// FreezeBounty 从可用余额冻结悬赏硬币：发布 / 追加时调用（事务内）
func FreezeBounty(tx *gorm.DB, authorID uint, amount int, storyID uint, op, remark string) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	u, err := lockUser(tx, authorID)
	if err != nil {
		return err
	}
	if u.Balance < amount {
		return ErrInsufficient
	}
	u.Balance -= amount
	u.Frozen += amount
	if err := tx.Save(u).Error; err != nil {
		return err
	}
	sid := storyID
	return writeLedger(tx, u, -amount, op, &sid, nil, nil, remark)
}

// SettleBounty 采纳结算：发起人冻结硬币出账，代看人可用余额入账（事务内）
func SettleBounty(tx *gorm.DB, authorID, helperID uint, amount int, storyID, responseID uint) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if authorID == helperID {
		return ErrSelfSettle
	}

	// 固定加锁顺序（按 ID 升序），避免并发事务交叉等待导致死锁
	firstID, secondID := authorID, helperID
	if helperID < authorID {
		firstID, secondID = helperID, authorID
	}
	first, err := lockUser(tx, firstID)
	if err != nil {
		return err
	}
	second, err := lockUser(tx, secondID)
	if err != nil {
		return err
	}
	author, helper := first, second
	if first.ID == helperID {
		author, helper = second, first
	}

	if author.Frozen < amount {
		return ErrFrozenShort
	}
	author.Frozen -= amount
	helper.Balance += amount
	if err := tx.Save(author).Error; err != nil {
		return err
	}
	if err := tx.Save(helper).Error; err != nil {
		return err
	}

	sid, rid := storyID, responseID
	authorRef, helperRef := helper.ID, author.ID
	if err := writeLedger(tx, author, 0, OpPay, &sid, &rid, &authorRef, "确认代看，悬赏硬币结算支出"); err != nil {
		return err
	}
	return writeLedger(tx, helper, amount, OpIncome, &sid, &rid, &helperRef, "代看被采纳，获得悬赏硬币")
}
