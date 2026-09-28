# 记忆连接 MemoryLink

> **你帮我再看一眼，我把记忆还给你**

一款帮助离城青年通过他人镜头代看回忆角落的应用。离开那座城后，老街还在吗、旧摊还生火吗——发布一条"求看"，冻结记忆硬币作为悬赏；仍在城里的人点击"替他去拍"，上传现场新照与寄语，凑成一张**双面明信片**（过去回忆 + 当下现场）；发起人确认采纳后，悬赏硬币结算给对方，记忆完成归还。

## 技术栈

- **前端**：Vue 3 + Vite + Vue Router + Axios
- **后端**：Go 1.22 + Gin + GORM
- **数据库**：MySQL 8.0
- **鉴权**：JWT

## 核心机制：记忆硬币循环

```
注册赠送 ◉100
   │
   ▼
发布求看 ── 冻结扣除悬赏（余额减少、故事记录 bounty）
   │            │
   │            └── 可继续「追加悬赏」（再次冻结）
   ▼
他人「替他去拍」上传现场新照 + 寄语（pending）
   │
   ▼
发起人「确认采纳」
   │
   ▼
冻结硬币结算给代看人（+reward 流水），故事 fulfilled，其它回应 rejected
```

**账目安全（数据库事务 + 行锁）**：
- 发布冻结：`SELECT ... FOR UPDATE` 锁定用户行 → 校验余额 → 扣减 → 建故事 → 写 `freeze` 流水，单事务完成，杜绝超额冻结。
- 追加悬赏：同上，锁定用户行与故事行，追加冻结并写 `add` 流水。
- 采纳结算：事务内锁定回应行 + 故事行 + 代看人行，校验状态后发放硬币、翻转状态、写 `reward` 流水；故事状态与回应状态双重校验保证**只能结算一次**。
- 每条流水都带 `balance_after`（发生后可用余额），形成可审计台账。

## 数据表

| 表 | 说明 |
|---|---|
| `users` | 用户，含可用硬币余额 `balance` |
| `stories` | 求看（明信片正面：回忆、老照片、城市地点、冻结悬赏、状态） |
| `responses` | 代看回应（明信片背面：现场文字、新照片、寄语、状态） |
| `coin_ledgers` | 硬币流水台账（register/freeze/add/unfreeze/reward） |

## 页面

1. **首页 Feed 流 `/`**：求看列表，城市/状态筛选；卡片可展开**双面明信片**（老照片回忆 ↔ 新照片现场）；发布求看、替他去拍、追加悬赏、确认采纳。
2. **个人中心 `/profile`**：可用/冻结/累计收入三栏硬币钱包；"我发布的求看 / 我代看的记录 / 硬币流水"三个标签页。

## RESTful API

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/api/auth/register` | | 注册（赠送 100 硬币） |
| POST | `/api/auth/login` | | 登录 |
| GET | `/api/me` | ✓ | 当前用户 |
| GET | `/api/profile` | ✓ | 个人中心聚合数据 |
| GET | `/api/stories` | | Feed 列表（可带 city/status） |
| GET | `/api/stories/:id` | | 故事详情 |
| POST | `/api/stories` | ✓ | 发布求看（事务冻结） |
| POST | `/api/stories/:id/bounty` | ✓ | 追加悬赏（事务冻结） |
| POST | `/api/stories/:id/responses` | ✓ | 替他去拍 |
| POST | `/api/responses/:rid/accept` | ✓ | 确认采纳（事务结算） |
| POST | `/api/upload` | ✓ | 上传图片 |

## 本地运行

### 1. 数据库

MySQL 本地启动，连接信息在 `backend/config/config.go`（默认 `root:123456@tcp(127.0.0.1:3306)`）。数据库 `memorylink` 由服务端自动创建表结构；也可手动建库：

```sql
CREATE DATABASE memorylink DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. 后端（:8080）

```bash
cd backend
go run .          # 或 go build -o memorylink.exe . && ./memorylink.exe
```

首次启动会自动建表并写入种子数据：
- 演示账号 **alice / bob / carol**，密码均为 **123456**
- 3 条示例求看（其中 1 条已完成双面明信片结算）

### 3. 前端（:5173）

```bash
cd frontend
npm install
npm run dev
```

打开 http://localhost:5173 （Vite 已代理 `/api` 与 `/uploads` 到 8080）。
