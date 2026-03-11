# Go Game Login Server

基于 Go 语言和 Gin 框架构建的游戏登录服务器，提供高性能的 HTTP API 服务，支持多实例部署。

## 技术栈

- **语言**: Go 1.24+
- **Web 框架**: Gin 1.11.0
- **配置管理**: Viper 1.21.0
- **数据库**:
  - MySQL (GORM 1.31.1)
  - Redis (go-redis v9.17.1) — 必选，用于缓存与多实例共享
- **日志系统**: slog (golang.org/x/exp)
- **其他**:
  - JWT: github.com/golang-jwt/jwt/v5
  - 定时任务: cron/v3
  - Shell 调试: yaegi 动态解释器

## 项目结构

```
loginServer/
├── config/              # 配置管理
│   └── config.go        # 主配置加载（config.json）
├── deploy/              # 部署脚本
│   ├── build_server.bat # 交叉编译 Linux 二进制
│   ├── run_server.bat   # Windows 本地运行（含数据库初始化）
│   ├── run_server.sh    # Linux 运行脚本
│   ├── stop_server.sh   # Linux 停止脚本
│   └── config.json      # 部署用配置
├── logs/                # 日志目录（自动生成）
├── pkg/                 # 核心工具包
│   ├── crypto/          # 加密/解密
│   ├── jwt/             # JWT Token 处理
│   ├── logger/          # 日志（基于 slog）
│   ├── mysql/           # MySQL 连接
│   ├── myutil/          # 通用工具
│   └── redis/           # Redis 连接
├── request/             # HTTP 请求层
│   ├── req_handle_adminServer.go  # 后台管理接口
│   ├── req_handle_out.go          # 对外接口
│   ├── req_handle_sgame.go        # 游戏服接口
│   ├── req_handle_test_group.go   # 测试接口
│   ├── req_params.go    # 参数解析
│   ├── req_register.go  # 路由注册
│   ├── req_util.go      # 中间件、响应工具
│   ├── req_whitelist.go  # 白名单处理
│   └── request.go       # 服务器启动与优雅关闭
├── shell/               # Shell 调试接口（yaegi REPL）
├── src/
│   ├── db/              # 数据层（统一入口，外部仅依赖本包）
│   │   ├── db.go        # 启动/关闭、业务与缓存入口
│   │   ├── db_init.go   # 初始化入口
│   │   ├── db_types.go  # 类型重新导出（供 request 等外部包使用）
│   │   ├── db_server_list.go     # 服务器列表
│   │   ├── db_login_notice.go    # 登录公告
│   │   ├── db_user_player_history.go  # 用户历史
│   │   ├── db_ip_whitelist.go    # IP 白名单
│   │   ├── db_mysql/    # MySQL 实现（仅 db 包内部使用）
│   │   └── db_redis/    # Redis 缓存实现（仅 db 包内部使用）
│   └── log/             # 日志包装
├── sql/
│   └── server.sql       # 数据库初始化
├── config.json          # 主配置文件
├── go.mod
├── go.sum
└── main.go              # 应用入口
```

## 核心特性

### 配置管理
- 单一主配置文件 `config.json`
- 基于 Viper，JSON 格式
- 加载失败时程序退出

### 日志系统
- 基于 `slog`
- 多级别（debug/info/warn/error）
- 按日期与级别分割文件
- 可配置保留天数

### 缓存系统
- **Redis 缓存**：服务器列表、公告、白名单
- **多实例共享**：支持水平扩展
- **懒加载**：GetOrLoad 模式
- **键命名**：`loginServer:server:list`、`loginServer:whitelist:<分组>` 等

### HTTP 服务器
- 基于 Gin
- 优雅关闭（SIGTERM/SIGQUIT/SIGINT）
- IP 白名单中间件（按 API 分组，支持 CIDR）
- CORS 支持

### IP 白名单
- **存储**：MySQL + Redis 缓存
- **分组**：sgame、adminServer、out、test
- **格式**：单 IP 或 CIDR（如 `192.168.1.0/24`）
- **动态管理**：通过 API 增删改查
- **启动同步**：配置中的白名单会同步到数据库

### 数据库
- **MySQL**：GORM，连接池
- **Redis**：必选，连接池，Ping 校验
- **数据表**：`game_list`、`user_player_history`、`login_notice`、`ip_whitelist`

## 环境要求

- Go 1.24+
- MySQL 5.7+ / 8.0+
- Redis 6.0+（必选）

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 初始化数据库

```bash
mysql -hlocalhost -P3306 -uroot -p123456 -e "source sql/server.sql"
```

### 3. 配置应用

编辑 `config.json`：

```json
{
    "server_mod": "dev",
    "server_name": "loginServer",
    "gin": {
        "ip": "127.0.0.1",
        "port": "9100",
        "mod": "debug"
    },
    "log": {
        "level": "debug",
        "path": "./logs",
        "remain_day": "60",
        "showfile": "1",
        "showfunc": "0"
    },
    "mysql": {
        "ip": "127.0.0.1",
        "port": "3306",
        "user": "root",
        "password": "123456",
        "db": "loginServer"
    },
    "redis": {
        "ip": "127.0.0.1",
        "port": "6379",
        "db": "1",
        "password": "123456"
    },
    "ip_whitelist": {
        "sgame": [
            "127.0.0.1",
            "192.168.1.0/24"
        ],
        "adminServer": [
            "10.0.1.141"
        ],
        "out": [],
        "test": []
    }
}
```

### 4. 运行

#### Windows

```bash
# 方式1：使用部署脚本（含数据库初始化）
deploy\run_server.bat

# 方式2：直接运行
go run .

# 方式3：带 Shell 调试
go run . shell
```

#### Linux

```bash
# 在 Windows 上交叉编译
cd deploy
build_server.bat

# 部署到 Linux
./run_server.sh

# 停止
./stop_server.sh
```

## 配置说明

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `server_mod` | 运行模式 | `dev` / `release` |
| `server_name` | 服务名称 | 任意字符串 |
| `gin.ip` | 监听 IP | 空或 127.0.0.1 时监听 0.0.0.0 |
| `gin.port` | 监听端口 | 端口号 |
| `gin.mod` | Gin 模式 | `debug` / `test` / `release` |
| `log.level` | 日志级别 | `debug` / `info` / `warn` / `error` |
| `log.path` | 日志目录 | 路径 |
| `log.remain_day` | 日志保留天数 | 正整数 |
| `mysql.*` | MySQL 连接 | - |
| `redis.*` | Redis 连接（必填） | - |
| `ip_whitelist` | 白名单初始配置 | 按分组配置，启动时同步到数据库 |

## 开发模式

```bash
# 普通运行
go run .

# 带 Shell 调试（REPL）
go run . shell
```

Shell 支持交互式命令：`user`、`db`、`config`、`exit` 等。

## 部署

### 交叉编译

```bash
cd deploy
build_server.bat
```

输出：`deploy/loginServer`（Linux amd64）

### Linux 部署

1. 复制到服务器：`loginServer`、`config.json`、`sql/server.sql`（如需）
2. 运行：`./run_server.sh`
3. 日志：`tail -f deploy/loginServer.log`

### 优雅关闭

收到 SIGTERM/SIGQUIT/SIGINT 时安全关闭，超时 10 秒。

## 架构设计

### 初始化流程

```
main.go
  ├── log.Start()            # 日志
  ├── db.Start()             # MySQL + Redis
  └── request.Start()        # HTTP 服务
      ├── initCache()        # 预加载服务器列表
      ├── InitWhitelistFromDB()  # 白名单
      ├── setupMiddleware()  # CORS、白名单、日志
      └── registerRoutes()   # 路由
```

### 缓存策略

- 服务器列表、公告、白名单使用 Redis
- `GetOrLoad` 懒加载，未命中时回源 MySQL
- 多实例共享，支持水平扩展

### 数据层（db 包）设计

数据层采用**分层隔离**设计，外部只通过 `package db` 访问数据，不得直接使用 `db_mysql` 或 `db_redis`。

| 层级 | 说明 |
|------|------|
| **db** | 统一入口。提供业务函数（如 `GetServerList`、`CreateLoginNotice`）和类型重新导出。 |
| **db_mysql** | MySQL 实现，仅被 db 包内部调用。 |
| **db_redis** | Redis 缓存实现，仅被 db 包内部调用。 |

**类型重新导出**：`db_types.go` 中将 `db_mysql` 中供业务使用的类型以类型别名形式重新导出（如 `GameList`、`PlayerHistoryItem`、`LoginNotice`），使 `request` 等外部包只需导入 `db` 即可完成类型与函数的调用，无需接触 `db_mysql`/`db_redis`。

**调用关系**：

```
request、main、config、shell 等
        │
        └── import db  ← 唯一数据访问入口
               │
               ├── db_mysql（内部实现）
               └── db_redis（内部实现）
```

## 日志管理

按级别和日期分割：
- `log_debug_YYYYMMDD.log`
- `log_info_YYYYMMDD.log`
- `log_warn_YYYYMMDD.log`
- `log_error_YYYYMMDD.log`

保留天数由 `log.remain_day` 控制。

## 故障排查

**配置失败**：检查 `config.json` 是否存在且格式正确。

**数据库失败**：检查 MySQL/Redis 服务、连接参数、网络。

**启动失败**：检查端口占用、配置、日志错误信息。

## License

[添加许可证信息]
