# shortlink —— 一个 Go 实现的短链接服务

基于 Gin + GORM + Redis 的短链接服务，支持短链生成、302 跳转与访问统计（今日 PV/UV、近 7 日 PV）。

这是一个学习向项目，重点实践了：雪花 ID 生成短码、缓存穿透防护、异步日志写入、优雅停机与容器化部署。

仓库地址：https://github.com/Fall-zy/shortlink

## 功能特性

- **短链生成**：雪花 ID + Base62 编码，短码紧凑且不可预测，不依赖数据库自增
- **302 跳转**：Redis 缓存加速，缓存未命中自动回源数据库
- **缓存穿透防护**：查询不存在的短码时写入短 TTL 占位符，防止恶意请求打穿数据库
- **访问统计**：今日 PV/UV（IP 去重）、近 7 日每日 PV
- **异步日志写入**：跳转路径不阻塞，带缓冲 channel + worker 池后台落库
- **优雅停机**：先停止接收请求，再排空日志缓冲，最后关闭数据库
- **双数据库支持**：SQLite（默认，纯 Go 驱动）与 MySQL，配置驱动切换
- **配置灵活**：YAML 配置文件 + 环境变量覆盖，容器内无需改动配置文件
- **容器化**：多阶段构建静态二进制，Docker Compose 一键编排 App + MySQL + Redis

## 技术栈

| 组件 | 选型 |
| --- | --- |
| 语言 | Go 1.25 |
| Web 框架 | Gin |
| ORM | GORM（sqlite / mysql 驱动） |
| 缓存 | Redis（go-redis v9） |
| ID 生成 | sonyflake（雪花算法变体） |
| 配置 | Viper（YAML + 环境变量） |
| 日志 | zap + gin-contrib/zap |
| 部署 | Docker 多阶段构建 + Docker Compose（MySQL 8 / Redis 7） |

## 项目结构

```
shortlink
├── cmd/main.go              # 入口：初始化配置/依赖/路由，优雅停机
├── config/                  # 配置加载（viper，YAML + 环境变量）
│   ├── config.go
│   └── config.yaml
├── internal
│   ├── handler/             # HTTP 层：参数校验、响应组装
│   ├── service/             # 业务层：短链生成、缓存策略、统计聚合、异步日志
│   ├── repository/          # 数据层：GORM 访问 SQLite/MySQL
│   ├── model/               # ORM 模型
│   └── utils/               # 雪花 ID、Base62、Redis、日志
├── web/index.html           # 前端页面（原生 JS）
├── Dockerfile               # 多阶段构建
└── docker-compose.yml       # App + MySQL + Redis 编排
```

分层依赖方向：`handler → service → repository → model`，`utils`/`config` 提供横切能力。

## 快速开始

### 方式一：Docker Compose（推荐）

需要 Docker 及 Docker Compose。

```bash
docker compose up -d
docker compose logs -f app
```

启动后访问 http://localhost:8080 。

> 注意：MySQL 容器首次启动需要约 10~20 秒完成初始化，期间 app 容器可能因数据库未就绪而自动重启（`restart: always` 兜底），等待 MySQL 就绪后即自动恢复正常，属预期现象。

### 方式二：本地运行

需要 Go 1.25+ 及本地 Redis（默认 `localhost:6379`，无密码）。

```bash
# 1. 启动 Redis（默认配置即可）
# 2. 下载依赖并运行
go mod download
go run ./cmd
```

默认使用 SQLite，数据库文件 `shortlink.db` 会在当前目录自动创建（首次启动自动建表）。

## 配置说明

配置文件为 `config/config.yaml`，所有配置项均可通过环境变量覆盖（前缀 `SHORTLINK_`，`.` 替换为 `_`）：

| 配置项 | 环境变量 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `server.port` | `SHORTLINK_SERVER_PORT` | `8080` | 监听端口 |
| `server.base_url` | `SHORTLINK_SERVER_BASE_URL` | `http://localhost:8080` | 生成短链时的域名前缀 |
| `database.driver` | `SHORTLINK_DATABASE_DRIVER` | `sqlite` | `sqlite` 或 `mysql` |
| `database.dsn` | `SHORTLINK_DATABASE_DSN` | `shortlink.db` | 数据库连接串 |
| `redis.addr` | `SHORTLINK_REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `redis.password` | `SHORTLINK_REDIS_PASSWORD` | 空 | Redis 密码 |
| `redis.db` | `SHORTLINK_REDIS_DB` | `0` | Redis 库号 |
| `redis.pool_size` | `SHORTLINK_REDIS_POOL_SIZE` | `10` | 连接池大小 |

## API 文档

### 1. 创建短链

```
POST /api/v1/shorten
Content-Type: application/json
```

请求体：

```json
{ "url": "https://example.com/a/very/long/path?query=1" }
```

成功响应（`201 Created`）：

```json
{ "short_url": "http://localhost:8080/r/1NqzG3mQ", "code": "1NqzG3mQ" }
```

示例：

```bash
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/a/very/long/path?query=1"}'
```

### 2. 短链跳转

```
GET /r/:code
```

命中时返回 `302 Found` 并携带 `Location` 头；短码不存在时返回 `404`。

```bash
curl -I http://localhost:8080/r/1NqzG3mQ
```

### 3. 访问统计

```
GET /api/v1/stats/:code
```

响应示例：

```json
{
  "today_pv": 3,
  "today_uv": 2,
  "daily_pvs": [
    { "date": "2026-08-18", "pv": 5 },
    { "date": "2026-08-19", "pv": 3 }
  ]
}
```

```bash
curl http://localhost:8080/api/v1/stats/1NqzG3mQ
```

## 设计要点

### 短码生成：雪花 ID + Base62

使用 sonyflake 生成 64 位雪花 ID，再编码为 Base62 短码。相比数据库自增 ID：

- 不依赖数据库、天然支持未来分布式扩展
- 短码不可预测，无法被遍历枚举
- 单机部署时机器 ID 随机生成，无需配置

### 读路径：Cache-Aside + 缓存穿透防护

```
请求 → Redis 命中？ → 直接返回
        ↓ 未命中
      查数据库
        ↓ 存在          → 回填缓存（TTL 1h）→ 返回
        ↓ 不存在        → 写入 __INVALID__ 占位符（TTL 1min）→ 返回 404
```

- 不存在的短码同样进缓存（占位符），避免恶意构造短码把请求打到数据库
- 占位符 TTL（1min）远短于正常缓存 TTL（1h），在防穿透与内存占用间取平衡
- Redis 故障时自动降级为直查数据库，缓存不可用不影响服务可用性

### 写路径：异步日志，不阻塞跳转

跳转是热路径，日志落库不能拖慢响应：

- 带缓冲 channel（容量 1000）+ 2 个 worker 协程后台落库
- 缓冲满时丢弃日志并告警（宁可丢日志，不可堵跳转）
- 优雅停机顺序：**先停止 HTTP 接收 → 再排空日志缓冲 → 最后关闭数据库**，保证停机时日志不丢

### 容器化：静态编译 + 多阶段构建

- `CGO_ENABLED=0` 静态编译，配合纯 Go 的 sqlite 驱动，二进制可直接跑在 alpine 上
- 多阶段构建：构建阶段体积大、依赖多；运行阶段只保留二进制 + 配置 + 前端 + CA 证书，镜像精简

## 部署注意事项

- **`SHORTLINK_SERVER_BASE_URL` 必须设置为公网域名**，否则生成的短链对用户不可访问
- 建议设置时区 `TZ=Asia/Shanghai`，统计的"今日"按服务器本地时间切分
- 部署在 Nginx/云负载均衡等反向代理后时，需配置可信代理（`SetTrustedProxies`），否则 `ClientIP()` 取到的是代理 IP，UV 统计会失真

## 已知限制

- 未实现限流与防滥用，`/shorten` 接口公开可调
- 统计接口无鉴权
- 相同 URL 会生成多条短链（未做去重）
- 仅提供统计 API，暂无统计展示页面
- 暂无单元测试
