package models

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"memorylink/config"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 连接数据库并自动迁移表结构
func InitDB() {
	db, err := gorm.Open(mysql.Open(config.MySQLDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("连接 MySQL 失败: %v", err)
	}
	DB = db

	if err := db.AutoMigrate(&User{}, &Story{}, &Response{}, &CoinLedger{}); err != nil {
		log.Fatalf("自动迁移失败: %v", err)
	}
	if err := seed(db); err != nil {
		log.Printf("种子数据写入失败: %v", err)
	}
}

func hashPassword(pwd string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	return string(h)
}

// CheckPassword 校验明文密码
func (u *User) CheckPassword(pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pwd)) == nil
}

// PhotoList 把逗号分隔的照片字段拆成切片
func (s *Story) PhotoList() []string { return splitPhotos(s.OldPhotos) }
func (r *Response) PhotoList() []string { return splitPhotos(r.NewPhotos) }

func splitPhotos(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func seed(db *gorm.DB) error {
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		return nil
	}
	log.Println("写入种子数据 ...")

	uploadDir := filepath.Join(config.UploadDir, "seed")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return err
	}

	// 用内嵌的纯色 PNG 生成示意图片，避免依赖外部资源
	photo := func(name, color string) string {
		path := filepath.Join(uploadDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			_ = os.WriteFile(path, placeholderPNG(color), 0o644)
		}
		return "/uploads/seed/" + name
	}

	mkUser := func(username, nick, bio string, balance int) *User {
		u := &User{
			Username:     username,
			PasswordHash: hashPassword("123456"),
			Nickname:     nick,
			Bio:          bio,
			Balance:      balance,
			Avatar:       "",
		}
		if err := db.Create(u).Error; err != nil {
			log.Printf("创建用户 %s 失败: %v", username, err)
		}
		return u
	}

	ledger := func(u *User, change int, typ, remark string, storyID *uint, respID *uint) {
		dir := CoinDirectionIn
		if change < 0 {
			dir = CoinDirectionOut
		}
		db.Create(&CoinLedger{
			UserID: u.ID, Change: change, BalanceAfter: u.Balance,
			Direction: dir, Type: typ, Remark: remark,
			StoryID: storyID, ResponseID: respID,
		})
	}

	alice := mkUser("alice", "离家的阿栀", "成都人，在上海做设计。想念老巷子里的锅盔。", config.RegisterGiftCoins)
	ledger(alice, config.RegisterGiftCoins, CoinTypeRegister, "注册赠送初始硬币", nil, nil)
	bob := mkUser("bob", "还在城里的波波", "土生土长成都娃，周末喜欢扫街拍照。", config.RegisterGiftCoins)
	ledger(bob, config.RegisterGiftCoins, CoinTypeRegister, "注册赠送初始硬币", nil, nil)
	carol := mkUser("carol", "北上的小鹿", "长沙人，在北京读研，最惦记学校后门的米粉。", config.RegisterGiftCoins)
	ledger(carol, config.RegisterGiftCoins, CoinTypeRegister, "注册赠送初始硬币", nil, nil)

	type storySeed struct {
		author   *User
		title    string
		city     string
		loc      string
		mem      string
		oldImgs  []string
		bounty   int
		responder *User
		now      string
		newImgs  []string
		msg      string
		status   string
	}

	seeds := []storySeed{
		{
			author: alice, title: "帮我看看巷口的老锅盔摊还在吗", city: "成都",
			loc:  "锦江区镗钯街尽头的老巷口",
			mem:  "小时候爷爷每天早上牵着我去买锅盔，两毛钱一个，炭火炉子烤得两面金黄。摊位旁边有棵歪脖子槐树，我总在树下等。2018 年离开成都去上海后再没回去过，想知道那口炉子还生不生火。",
			oldImgs: []string{
				photo("old-guokui.png", "#b98a5a"),
				photo("old-lane.png", "#8c9a6b"),
			},
			bounty: 20,
			responder: bob,
			now:  "摊子还在！槐树被修剪过但还立着。老板换成了老师傅的儿子，炉子还是那口炭炉。锅盔涨到六块钱了，排队的人比以前还多。",
			newImgs: []string{
				photo("new-guokui.png", "#d9a441"),
				photo("new-lane.png", "#7d9464"),
			},
			msg:    "替你吃了一个，还是当年的味道。附上新出炉的照片，热乎着呢。",
			status: StoryStatusFulfilled,
		},
		{
			author: carol, title: "想再看一眼师大后山的旧铁轨", city: "长沙",
			loc:  "岳麓山下湖南师大后山废弃铁路",
			mem:  "考研失败的那个冬天，我每晚沿着旧铁轨走回宿舍。铁轨两侧长满了荒草，远处能看到橘子洲的灯。想请人帮我看看铁轨有没有被拆掉，道口那块红色警示牌还在不在。",
			oldImgs: []string{photo("old-rail.png", "#7d6b5a")},
			bounty:  15,
			status:  StoryStatusOpen,
		},
		{
			author: alice, title: "东郊记忆的涂鸦墙换成了什么样子", city: "成都",
			loc:  "成华区东郊记忆园区西南墙",
			mem:  "2016 年和大学室友在那面墙上画过一只巨大的鲸鱼，毕业时被物业刷白了。听说现在又允许涂鸦了，想知道墙上现在画着什么。",
			oldImgs: []string{photo("old-wall.png", "#5a7d9a")},
			bounty:  10,
			status:  StoryStatusOpen,
		},
	}

	for _, sd := range seeds {
		// 冻结硬币
		sd.author.Balance -= sd.bounty
		db.Model(&User{}).Where("id = ?", sd.author.ID).Update("balance", sd.author.Balance)

		story := &Story{
			UserID: sd.author.ID, Title: sd.title, City: sd.city, Location: sd.loc,
			MemoryText: sd.mem, OldPhotos: strings.Join(sd.oldImgs, ","),
			Bounty: sd.bounty, Status: sd.status,
		}
		if err := db.Create(story).Error; err != nil {
			return err
		}
		ledger(sd.author, -sd.bounty, CoinTypeFreeze, fmt.Sprintf("发布求看《%s》冻结悬赏", sd.title), &story.ID, nil)

		if sd.status == StoryStatusFulfilled && sd.responder != nil {
			resp := &Response{
				StoryID: story.ID, UserID: sd.responder.ID, NowText: sd.now,
				NewPhotos: strings.Join(sd.newImgs, ","), Message: sd.msg,
				Status: RespStatusAccepted,
			}
			if err := db.Create(resp).Error; err != nil {
				return err
			}
			story.AcceptedRespID = &resp.ID
			db.Model(story).Update("accepted_resp_id", resp.ID)

			// 结算：冻结的硬币付给代看人
			sd.responder.Balance += sd.bounty
			db.Model(&User{}).Where("id = ?", sd.responder.ID).Update("balance", sd.responder.Balance)
			ledger(sd.responder, sd.bounty, CoinTypeReward,
				fmt.Sprintf("代看《%s》被采纳获得悬赏", sd.title), &story.ID, &resp.ID)
		}
	}

	log.Println("种子数据完成：3 个用户（alice/bob/carol，密码均 123456），3 条求看")
	return nil
}

// placeholderPNG 生成一张 600x400 的纯色 PNG 作为示意照片
func placeholderPNG(hexColor string) []byte {
	c := parseHexColor(hexColor)
	img := image.NewRGBA(image.Rect(0, 0, 600, 400))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func parseHexColor(s string) color.RGBA {
	c := color.RGBA{A: 255}
	s = strings.TrimPrefix(s, "#")
	if len(s) == 6 {
		fmt.Sscanf(s, "%02x%02x%02x", &c.R, &c.G, &c.B)
	}
	return c
}
