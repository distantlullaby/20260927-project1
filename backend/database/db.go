package database

import (
	"fmt"
	"log"

	"memoryconnect/config"
	"memoryconnect/models"
	"memoryconnect/services"
	"memoryconnect/seed"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Init 连接 MySQL、自动建库、迁移表结构，并在空库时写入演示数据
func Init(cfg *config.Config) *gorm.DB {
	// 先连服务器（不指定库），自动创建数据库
	root, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 MySQL 失败：%v（可用 DB_USER / DB_PASSWORD / DB_HOST 等环境变量调整）", err)
	}
	if err := root.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		cfg.DBName)).Error; err != nil {
		log.Fatalf("创建数据库失败：%v", err)
	}

	db, err := gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("打开数据库失败：%v", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Story{}, &models.Response{}, &models.CoinLedger{}); err != nil {
		log.Fatalf("数据库迁移失败：%v", err)
	}
	DB = db

	seedIfEmpty(db, cfg)
	return db
}

// seedIfEmpty 首次启动且没有用户时写入演示账号与故事
func seedIfEmpty(db *gorm.DB, cfg *config.Config) {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		return
	}

	type seedUser struct {
		u        *models.User
		password string
	}
	mkUser := func(username, nickname string) seedUser {
		pw, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
		return seedUser{
			u: &models.User{
				Username: username,
				Password: string(pw),
				Nickname: nickname,
				Avatar:   "",
				Token:    "",
			},
			password: "123456",
		}
	}

	aLin := mkUser("alin", "阿林")
	xiaoMan := mkUser("xiaoman", "小满")
	oldChen := mkUser("oldchen", "城南的阿诚")

	users := []seedUser{aLin, xiaoMan, oldChen}
	for _, su := range users {
		if err := db.Create(su.u).Error; err != nil {
			log.Printf("种子用户创建失败：%v", err)
			return
		}
	}

	// 注册赠送：每个账号初始 100 硬币
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, su := range users {
			if err := services.GrantInitial(tx, su.u.ID, cfg.InitCoins); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("种子硬币赠送失败：%v", err)
		return
	}

	// 生成演示用的老照片（SVG 手绘明信片风格）
	old1 := seed.OldPhotoURL(cfg.UploadDir, "seed-old-alley", seed.SceneAlley)
	old2 := seed.OldPhotoURL(cfg.UploadDir, "seed-old-school", seed.SceneSchool)
	old3 := seed.OldPhotoURL(cfg.UploadDir, "seed-old-riverside", seed.SceneRiverside)

	type seedStory struct {
		st       *models.Story
		reward   int
		respBy   *models.User
		message  string
		newScene seed.Scene
		settle   bool
	}
	stories := []seedStory{
		{
			st: &models.Story{
				AuthorID:  aLin.u.ID,
				Title:     "再看一眼巷口的老槐树",
				City:      "长沙",
				Location:  "天心区坡子街深处的槐树巷",
				Memory:    "小时候放了学总在这棵老槐树下等爷爷下班，他的自行车铃一响我就冲过去。后来巷子拆迁，全家搬来了深圳，我已经八年没回去过了。",
				OldPhotos: old1,
				Wish:      "想请路过的人帮忙拍一张老槐树现在的样子，树下要是还有下棋的老人就更好了。",
				Status:    "open",
			},
			reward:   20,
			respBy:   xiaoMan.u,
			message:  "槐树还在！树下今天有两位爷爷在下象棋，我帮你把铃声也一起记下来了。它比你记忆里粗了一圈，枝桠上多了好几个新鸟窝。",
			newScene: seed.SceneAlley,
		},
		{
			st: &models.Story{
				AuthorID:  xiaoMan.u.ID,
				Title:     "三中门口的小卖部还在吗",
				City:      "武汉",
				Location:  "武昌区第三中学正门",
				Memory:    "五毛钱的橘子汽水、放学排队买的辣条，老板记得每个人的名字。听说整条街都翻新了，不知道那个绿漆铁皮棚还在不在。",
				OldPhotos: old2,
				Wish:      "想看看三中门口现在是什么样，如果小卖部还在，替我买瓶橘子汽水吧（笑）。",
				Status:    "open",
			},
			reward: 10,
		},
		{
			st: &models.Story{
				AuthorID:  oldChen.u.ID,
				Title:     "江边台阶上的风筝线",
				City:      "九江",
				Location:  "长江大堤老渡口台阶",
				Memory:    "爸爸每年春天带我在这里放风筝，我的那只老鹰风筝最后挂在了江对面的树上。来上海打工第六年，江风的味道都快忘了。",
				OldPhotos: old3,
				Wish:      "想看看大堤和台阶现在的样子，春天江堤上还有没有人放风筝。",
				Status:    "open",
			},
			reward:   30,
			respBy:   aLin.u,
			message:  "今天风很好，台阶上有三户人家在放风筝，我拍了飞得最高的那只。渡口翻新了，但台阶还是老样子，一级不少。",
			newScene: seed.SceneRiverside,
			settle:   true, // 已结算的完整闭环示例
		},
	}

	for _, ss := range stories {
		ss.st.Reward = ss.reward // 悬赏写入故事行，与冻结的硬币保持一致
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(ss.st).Error; err != nil {
				return err
			}
			return services.FreezeBounty(tx, ss.st.AuthorID, ss.reward, ss.st.ID, services.OpFreeze, "发布求看，冻结悬赏硬币")
		})
		if err != nil {
			log.Printf("种子故事创建失败：%v", err)
			return
		}

		if ss.respBy == nil {
			continue
		}
		newPhoto := seed.NewPhotoURL(cfg.UploadDir, fmt.Sprintf("seed-new-%d", ss.st.ID), ss.newScene)
		resp := &models.Response{
			StoryID:   ss.st.ID,
			UserID:    ss.respBy.ID,
			Message:   ss.message,
			NewPhotos: newPhoto,
		}
		if err := db.Create(resp).Error; err != nil {
			log.Printf("种子回应创建失败：%v", err)
			return
		}

		if ss.settle {
			err := db.Transaction(func(tx *gorm.DB) error {
				if err := services.SettleBounty(tx, ss.st.AuthorID, ss.respBy.ID, ss.reward, ss.st.ID, resp.ID); err != nil {
					return err
				}
				return tx.Model(&models.Story{}).Where("id = ?", ss.st.ID).Updates(map[string]interface{}{
					"status":               "settled",
					"accepted_response_id": resp.ID,
				}).Error
			})
			if err != nil {
				log.Printf("种子结算失败：%v", err)
				return
			}
			if err := db.Model(&models.Response{}).Where("id = ?", resp.ID).Update("accepted", true).Error; err != nil {
				log.Printf("种子回应更新失败：%v", err)
			}
			ss.st.Status = "settled"
		}
	}

	log.Println("演示数据已就绪：账号 alin / xiaoman / oldchen，密码均为 123456")
}
