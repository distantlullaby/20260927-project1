package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	Nickname     string         `gorm:"size:50" json:"nickname"`
	Avatar       string         `gorm:"size:500" json:"avatar"`
	Bio          string         `gorm:"size:200" json:"bio"`
	Balance      int            `gorm:"not null;default:0" json:"balance"` // 可用记忆硬币（冻结部分不在此列）
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// 故事状态
const (
	StoryStatusOpen      = "open"      // 求看中（等待他人代看）
	StoryStatusFulfilled = "fulfilled" // 已采纳（记忆已归还）
)

// Story 故事表（求看请求 / 双面明信片的正面：过去回忆）
type Story struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"index;not null" json:"user_id"`
	Title         string         `gorm:"size:100;not null" json:"title"`
	City          string         `gorm:"size:50;not null;index" json:"city"`
	Location      string         `gorm:"size:200;not null" json:"location"`
	MemoryText    string         `gorm:"type:text;not null" json:"memory_text"`  // 过去回忆
	OldPhotos     string         `gorm:"size:1000" json:"old_photos"`            // 老照片 URL，逗号分隔（多张）
	Bounty        int            `gorm:"not null;default:0" json:"bounty"`       // 当前悬赏总额（冻结中）
	Status        string         `gorm:"size:20;not null;default:open;index" json:"status"`
	AcceptedRespID *uint         `gorm:"column:accepted_response_id;index" json:"accepted_response_id"` // 被采纳的回应
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联的展示字段
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Responses []Response `gorm:"foreignKey:StoryID" json:"responses,omitempty"`
}

// 回应状态
const (
	RespStatusPending  = "pending"  // 已提交，等待发起人确认
	RespStatusAccepted = "accepted" // 已采纳，硬币已结算
	RespStatusRejected = "rejected" // 未采纳
)

// Response 回应表（双面明信片的背面：当下现场）
type Response struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	StoryID    uint           `gorm:"index;not null" json:"story_id"`
	UserID     uint           `gorm:"index;not null" json:"user_id"` // 代看人
	NowText    string         `gorm:"type:text;not null" json:"now_text"`
	NewPhotos  string         `gorm:"size:1000" json:"new_photos"` // 现场新照，逗号分隔
	Message    string         `gorm:"type:text" json:"message"`     // 寄语
	Status     string         `gorm:"size:20;not null;default:pending;index" json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	User  *User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Story *Story `gorm:"foreignKey:StoryID" json:"story,omitempty"`
}

// 硬币流水类型
const (
	CoinTypeRegister = "register" // 注册赠送
	CoinTypeFreeze   = "freeze"   // 发布冻结
	CoinTypeAdd      = "add"      // 追加冻结
	CoinTypeUnfreeze = "unfreeze" // 未采纳退回
	CoinTypeReward   = "reward"   // 采纳结算（代看人收入）
)

// 硬币流水中的资金方向
const (
	CoinDirectionIn  = "in"  // 入账
	CoinDirectionOut = "out" // 出账（冻结也视为出账）
)

// CoinLedger 硬币流水表（财务台账）
type CoinLedger struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index;not null" json:"user_id"`
	Change     int       `gorm:"not null" json:"change"` // 正数入账 / 负数出账
	Direction  string    `gorm:"size:4;not null" json:"direction"` // in / out
	BalanceAfter int     `gorm:"not null" json:"balance_after"`
	Type       string    `gorm:"size:20;not null;index" json:"type"`
	StoryID    *uint     `gorm:"index" json:"story_id"`
	ResponseID *uint     `gorm:"index" json:"response_id"`
	Remark     string    `gorm:"size:200" json:"remark"`
	CreatedAt  time.Time `json:"created_at"`
}
