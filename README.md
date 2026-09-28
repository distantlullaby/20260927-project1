# 记忆连接 · Memory Connect

> **你帮我再看一眼，我把记忆还给你。**

帮助离城青年，通过他人的镜头"代看"回忆角落的应用。离城的人发布一张求看明信片——**过去回忆 + 老照片**，还在那座城市的人拍下**当下现场 + 新照片与寄语**回应，发起人确认后，悬赏的「记忆硬币」结算给代看人，让回忆在硬币的循环里流动。

## 核心玩法（记忆硬币循环）

```
注册 ──赠送 100──▶ 有硬币可用
                      │
发布求看 ──冻结扣除悬赏（支持追加）──▶ 余额减少 / 冻结增加
                      │
他人「替他去拍」上传现场新照 + 寄语
                      │
发起人确认 ──事务结算──▶ 冻结硬币 → 代看人余额（本人余额不变）
                      │
代看人赚到硬币 ──再去发布自己的求看──▶ 循环
```

- **注册赠送**：新用户即得 100 枚初始硬币
- **发布冻结**：发布时从可用余额冻结悬赏硬币（余额不足不可发），支持随时**追加**
- **代看回应**：他人上传现场新照与寄语（同一人对同一需求只能回应一次，不能替自己拍）
- **确认结算**：发起人点击确认，冻结硬币在**数据库事务 + 行锁**下安全转给对方
- **全程留痕**：每次硬币变动都写入硬币流水表

## 双面明信片

首页 Feed 流里每张卡片都可展开，在两面之间切换：
- **过去面**：回忆文字 + 老照片 +「想请你拍什么」
- **当下面**：所有代看人上传的现场新照 + 寄语，发起人可在此确认结算

## 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3（Composition API）+ Vite + Vue Router |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | MySQL 8（自动建库建表，首次启动写入演示数据） |
| 账目安全 | 数据库事务 + `SELECT ... FOR UPDATE` 行锁 |

### 数据表

- `users`：用户（`balance` 可用硬币 / `frozen` 冻结硬币）
- `stories`：求看故事（悬赏金额、状态 open/settled、老照片、回忆）
- `responses`：代看回应（现场新照片、寄语、是否被采纳；同用户同故事唯一）
- `coin_ledgers`：硬币流水（注册/冻结/追加/支出/收入，含变动后余额快照）

## 目录结构

```
.
├── backend/                Go + Gin + GORM
│   ├── main.go             路由与启动
│   ├── config/             配置（环境变量可覆盖）
│   ├── models/             四张表模型
│   ├── services/coin.go    ★ 硬币事务：赠送/冻结/结算（行锁）
│   ├── database/           建库、迁移、种子数据
│   ├── handlers/           auth / story / user / upload
│   ├── middleware/         token 鉴权、CORS
│   └── seed/               演示明信片插画（SVG）
└── frontend/               Vue 3 + Vite
    └── src/
        ├── views/          Feed（首页）/ Profile（个人中心）/ Login
        ├── components/     双面明信片卡片、三个弹窗、图片上传、灯箱
        ├── api/            RESTful 请求封装
        └── store/          登录态、轻提示
```

## 快速开始

### 0. 前置环境

- Go 1.22+、Node.js 18+、本地 MySQL 8
- 默认连接 `root` / `123456` / `127.0.0.1:3306`，可用环境变量覆盖：
  `DB_USER`、`DB_PASSWORD`、`DB_HOST`、`DB_PORT`、`DB_NAME`、`SERVER_PORT`

### 1. 启动后端（:8091）

```powershell
cd backend
go run .
```

首次启动会自动创建数据库 `memory_connect`、建表并写入演示数据。

### 2. 启动前端（:5183）

```powershell
cd frontend
npm install
npm run dev
```

打开 http://localhost:5183 ，Vite 会把 `/api` 与 `/uploads` 代理到后端 8091。

### 演示账号（密码均为 `123456`）

| 账号 | 昵称 | 看点 |
|---|---|---|
| `alin` | 阿林 | 发布了「老槐树」求看，已被小满代拍、待确认 |
| `xiaoman` | 小满 | 发布了「三中小卖部」求看，暂无回应 |
| `oldchen` | 阿诚 | 发布的「江边风筝」已完成结算闭环 |

> 可用两个不同浏览器/隐身窗口分别登录两个账号，体验"代拍 → 对方确认 → 硬币到账"的完整链路。

## RESTful API 摘要

| 方法 | 路径 | 鉴权 | 说明 |
|---|---|---|---|
| POST | `/api/auth/register` | – | 注册并赠送初始硬币 |
| POST | `/api/auth/login` | – | 登录 |
| GET  | `/api/stories` | 可选 | Feed 列表（`scope=mine/responded`、`status`、`city`） |
| POST | `/api/stories` | ✅ | 发布求看（事务冻结） |
| POST | `/api/stories/:id/reward` | ✅ | 追加悬赏（事务冻结） |
| POST | `/api/stories/:id/responses` | ✅ | 替他去拍 |
| POST | `/api/stories/:id/responses/:rid/accept` | ✅ | 确认结算（事务转账） |
| GET  | `/api/me` | ✅ | 个人信息与统计 |
| GET  | `/api/me/ledgers` | ✅ | 硬币流水 |
| GET  | `/api/me/stories` | ✅ | 我发布的 |
| GET  | `/api/me/responses` | ✅ | 我代看的 |
| POST | `/api/uploads` | ✅ | 图片上传 |
