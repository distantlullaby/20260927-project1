package models

import "time"

// User 用户表：balance 为可用硬币，frozen 为发布求看时冻结的硬币
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password  string    `gorm:"size:100;not null" json:"-"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	Token     string    `gorm:"size:64;index" json:"-"`
	Balance   int       `gorm:"not null;default:0" json:"balance"`
	Frozen    int       `gorm:"not null;default:0" json:"frozen"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Story 故事表（求看需求 / 双面明信片的“过去”面）
type Story struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	AuthorID         uint      `gorm:"index;not null" json:"authorId"`
	Title            string    `gorm:"size:100;not null" json:"title"`
	City             string    `gorm:"size:50" json:"city"`
	Location         string    `gorm:"size:100" json:"location"`
	Memory           string    `gorm:"type:text" json:"memory"`     // 过去回忆
	OldPhotos        string    `gorm:"type:text" json:"-"`          // 老照片，逗号分隔的 URL
	Wish             string    `gorm:"type:text" json:"wish"`       // 想让对方拍什么
	Reward           int       `gorm:"not null;default:0" json:"reward"`
	Status           string    `gorm:"size:20;not null;default:open;index" json:"status"` // open / settled
	AcceptedResponseID *uint   `gorm:"index" json:"acceptedResponseId"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// Response 回应表（双面明信片的“当下”面：现场新照 + 寄语）
// 同一用户对同一条故事只能回应一次
type Response struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	StoryID   uint      `gorm:"uniqueIndex:idx_story_user;not null" json:"storyId"`
	UserID    uint      `gorm:"uniqueIndex:idx_story_user;not null" json:"userId"`
	Message   string    `gorm:"type:text" json:"message"`
	NewPhotos string    `gorm:"type:text" json:"-"` // 新照片，逗号分隔的 URL
	Accepted  bool      `gorm:"not null;default:false" json:"accepted"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// CoinLedger 硬币流水表：每一次硬币变动都留痕
type CoinLedger struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index;not null" json:"userId"`
	Change       int       `gorm:"not null" json:"change"` // 对可用余额的带符号变动
	BalanceAfter int       `gorm:"not null" json:"balanceAfter"`
	FrozenAfter  int       `gorm:"not null" json:"frozenAfter"`
	Type         string    `gorm:"size:30;index;not null" json:"type"`
	StoryID      *uint     `gorm:"index" json:"storyId"`
	ResponseID   *uint     `gorm:"index" json:"responseId"`
	RefUserID    *uint     `json:"refUserId"`
	Remark       string    `gorm:"size:255" json:"remark"`
	CreatedAt    time.Time `json:"createdAt"`
}
