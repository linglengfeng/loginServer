# Go Game Login Server

基于 Go 语言和 Gin 框架构建的游戏登录服务器，提供高性能的 HTTP API 服务。

## 技术栈

- **语言**: Go 1.24+
- **Web 框架**: Gin 1.11.0
- **配置管理**: Viper 1.21.0
- **数据库**: 
  - MySQL (GORM 1.31.1)
  - Redis (go-redis v9.17.1)
- **日志系统**: slog (golang.org/x/exp)
- **其他**:
  - JWT: github.com/dgrijalva/jwt-go
  - 邮件服务: SendGrid / QQ SMTP
  - 定时任务: cron/v3

## 项目结构

```
loginServer/
├── config/              # 配置管理模块
│   ├── cfg/            # 子配置文件目录
│   │   └── mailer.json # 邮件服务配置
│   └── config.go       # 配置加载器（支持主配置和子配置）
├── deploy/             # 部署脚本
│   ├── build_server.bat    # Windows 构建脚本
│   ├── run.bat             # Windows 运行脚本
│   ├── run_server.sh       # Linux 运行脚本
│   └── stop_server.sh      # Linux 停止脚本
├── logs/               # 日志文件目录（自动生成）
├── pkg/                # 核心工具包
│   ├── crypto/         # 加密/解密工具
│   ├── jwt/            # JWT Token 处理
│   ├── logger/         # 日志系统（基于 slog，支持文件和控制台输出）
│   ├── mysql/          # MySQL 连接管理
│   ├── myutil/         # 通用工具函数
│   └── redis/          # Redis 连接管理
├── request/            # HTTP 请求处理层
│   ├── req_cache.go    # 内存缓存管理（线程安全）
│   ├── req_handle.go   # 请求处理器
│   ├── req_register.go # 路由注册
│   ├── req_util.go     # 请求工具函数（中间件、参数解析等）
│   └── request.go      # HTTP 服务器启动和优雅关闭
├── shell/              # Shell 调试接口
│   └── shell.go        # 交互式命令行工具
├── src/                # 业务模块
│   ├── db/             # 数据库抽象层
│   │   ├── db.go       # 数据库初始化入口
│   │   ├── db_mysql/   # MySQL 实现
│   │   └── db_redis/   # Redis 实现
│   ├── log/            # 日志包装器
│   └── mailer/         # 邮件服务
│       ├── mailer.go   # 邮件服务入口
│       ├── emailqq/    # QQ 邮箱实现
│       └── sendgrid/   # SendGrid 实现
├── sql/                # SQL 脚本
│   └── server.sql      # 数据库初始化脚本
├── config.json         # 主配置文件
├── go.mod              # Go 模块定义
├── go.sum              # Go 模块校验和
└── main.go             # 应用入口
```

## 核心特性

### 配置管理
- 支持主配置文件（`config.json`）和子配置文件（`config/cfg/*.json`）
- 基于 Viper，支持 JSON 格式
- 配置加载失败时自动终止程序
- 提供配置验证和重新加载功能

### 日志系统
- 基于 `slog` 实现
- 支持多级别日志（debug/info/warn/error）
- 自动按日期和级别分割日志文件
- 支持文件和控制台双重输出
- 可配置日志保留天数

### 缓存系统
- 线程安全的内存缓存
- 支持懒加载（Lazy Loading）
- Double-Check 模式防止竞态条件
- 可扩展的缓存键管理

### HTTP 服务器
- 基于 Gin 框架
- 支持优雅关闭（Graceful Shutdown）
- 可配置的请求超时和连接管理
- IP 白名单中间件支持（按 API 分组管理，支持 CIDR 格式）
- CORS 跨域支持

### IP 白名单管理
- **数据库存储**: 白名单数据持久化到 MySQL 数据库
- **动态管理**: 支持通过 API 动态添加、删除、查询白名单
- **按分组管理**: 支持按 API 分组（sgame、adminServer、out、test）分别配置白名单
- **CIDR 支持**: 支持单个 IP 和 CIDR 网段格式（如 `192.168.1.0/24`）
- **配置同步**: 启动时自动将配置文件中的白名单同步到数据库
- **缓存加速**: 使用内存缓存提升白名单查询性能

### 数据库支持
- MySQL: 基于 GORM，支持连接池管理
- Redis: 基于 go-redis，支持连接池和超时配置
- 自动初始化数据库连接
- **数据表**:
  - `game_list`: 游戏服务器列表
  - `user_player_history`: 玩家历史记录
  - `login_notice`: 登录公告配置
  - `ip_whitelist`: IP 白名单配置（支持按 API 分组管理）

### 邮件服务
- 支持多种邮件服务商（QQ 邮箱、SendGrid）
- 可配置的邮件服务切换
- 统一的邮件发送接口

## 环境要求

- Go 1.24 或更高版本
- MySQL 5.7+ 或 MySQL 8.0+
- Redis 6.0+（可选，用于缓存）

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd loginServer
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 初始化数据库

```bash
mysql -hlocalhost -P3306 -uroot -p123456 -e "source sql/server.sql"
```

### 4. 配置应用

创建或修改 `config.json`：

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
        "password": ""
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

创建邮件服务配置 `config/cfg/mailer.json`：

```json
{
    "mailer": "qq",
    "sendgrid": {
        "username": "example@gmail.com",
        "api_key": "your_sendgrid_api_key"
    },
    "qq": {
        "ip": "smtp.qq.com",
        "port": "587",
        "username": "xxx@qq.com",
        "authcode": "xxx"
    }
}
```

### 5. 运行应用

#### Windows

```bash
# 方式1: 使用部署脚本（包含数据库初始化）
deploy\run.bat

# 方式2: 直接运行
go run .

# 方式3: 带 Shell 调试接口
go run . shell
```

#### Linux

```bash
# 构建 Linux 二进制文件
cd deploy
./build_server.bat  # 在 Windows 上交叉编译

# 部署到 Linux 服务器
./run_server.sh

# 停止服务
./stop_server.sh
```

## 配置说明

### 主配置文件 (config.json)

| 配置项 | 说明 | 可选值 |
|--------|------|--------|
| `server_mod` | 服务器运行模式 | `dev` / `release` |
| `server_name` | 服务器名称 | 任意字符串 |
| `gin.ip` | HTTP 服务器 IP | IP 地址 |
| `gin.port` | HTTP 服务器端口 | 端口号 |
| `gin.mod` | Gin 运行模式 | `debug` / `test` / `release` |
| `log.level` | 日志级别 | `debug` / `info` / `warn` / `error` |
| `log.path` | 日志文件目录 | 相对或绝对路径 |
| `log.remain_day` | 日志保留天数 | 正整数 |
| `log.showfile` | 是否显示文件名 | `0` / `1` |
| `log.showfunc` | 是否显示函数名 | `0` / `1` |
| `mysql.*` | MySQL 连接配置 | - |
| `redis.*` | Redis 连接配置（可选） | - |
| `ip_whitelist` | IP 白名单初始配置 | 按 API 分组配置，启动时自动同步到数据库 |

**注意**: `ip_whitelist` 配置项用于初始配置，启动时会自动将配置中的白名单同步到数据库。后续的白名单管理应通过 API 接口进行，数据存储在数据库中。

### 邮件服务配置 (config/cfg/mailer.json)

| 配置项 | 说明 |
|--------|------|
| `mailer` | 邮件服务类型 | `qq` / `sendgrid` |
| `qq.*` | QQ 邮箱 SMTP 配置 | - |
| `sendgrid.*` | SendGrid API 配置 | - |

### IP 白名单配置说明

IP 白名单支持按 API 分组管理，支持单个 IP 和 CIDR 网段格式。配置中的白名单会在启动时自动同步到数据库，后续可通过 `/loginServer/whitelist/` 接口动态管理。

## 开发模式

### 运行开发服务器

```bash
go run .
```

### 使用 Shell 调试接口

```bash
go run . shell
```

启动后可以通过交互式命令进行调试：
- `user` - 用户相关功能测试
- `db` - 数据库连接测试
- `config` - 显示配置信息
- `exit` - 退出 Shell

## 部署

### 构建 Linux 二进制文件

在 Windows 环境下交叉编译：

```bash
cd deploy
build_server.bat
```

生成的二进制文件位于 `deploy/loginServer`

### Linux 服务器部署

1. 将以下文件复制到服务器：
   - `deploy/loginServer`（二进制文件）
   - `config.json`（主配置文件）
   - `config/cfg/`（子配置目录）
   - `sql/server.sql`（数据库脚本，如需要）

2. 运行部署脚本：

```bash
cd deploy
chmod +x run_server.sh
./run_server.sh
```

3. 查看日志：

```bash
tail -f deploy/loginServer.log
```

### 优雅关闭

服务器支持优雅关闭，接收到以下信号时会安全关闭：
- `SIGTERM`
- `SIGQUIT`
- `SIGINT`

关闭超时时间为 10 秒。

## 架构设计

### 初始化流程

```
main.go
  ├── log.Start()              # 初始化日志系统
  ├── db.Start()               # 初始化数据库连接（MySQL + Redis）
  ├── mailer.Start()           # 初始化邮件服务
  └── request.Start()          # 启动 HTTP 服务器
      ├── initCache()          # 初始化缓存（服务器列表）
      ├── InitWhitelistFromDB() # 初始化IP白名单
      │   ├── 从数据库加载白名单
      │   ├── 检查配置文件并同步缺失的IP
      │   └── 写入缓存
      ├── setupMiddleware()     # 设置中间件（包含IP白名单验证）
      └── registerRoutes()      # 注册路由
```

### 请求处理流程

```
HTTP Request
  ├── 中间件层
  │   ├── CORS 处理
  │   ├── IP 白名单验证
  │   │   ├── 获取路由的 API 分组
  │   │   ├── 从缓存查询该分组的白名单
  │   │   ├── 验证客户端 IP（支持 CIDR 匹配）
  │   │   └── 拒绝或允许访问
  │   └── 请求日志记录
  ├── 路由匹配
  ├── 处理器执行
  │   ├── 参数解析
  │   ├── 业务逻辑处理
  │   └── 响应返回
  └── 响应输出
```

### 缓存机制

- **缓存策略**: 内存缓存 + 懒加载
- **线程安全**: 使用 `go-cache` 实现线程安全的缓存
- **缓存更新**: 支持手动刷新缓存
- **白名单缓存**: 
  - 启动时从数据库加载到缓存
  - 修改操作时自动更新缓存
  - 确保缓存与数据库数据一致

## 日志管理

日志文件按级别和日期自动分割：
- `log_debug_YYYYMMDD.log` - 调试日志
- `log_info_YYYYMMDD.log` - 信息日志
- `log_warn_YYYYMMDD.log` - 警告日志
- `log_error_YYYYMMDD.log` - 错误日志

日志文件自动清理，保留天数由 `log.remain_day` 配置项控制。

## 性能优化

- **连接池**: MySQL 和 Redis 均使用连接池管理
- **缓存**: 内存缓存减少数据库查询
- **超时控制**: HTTP 请求和数据库操作均有超时保护
- **优雅关闭**: 确保正在处理的请求完成后再关闭

## 故障排查

### 配置加载失败

检查：
1. `config.json` 文件是否存在且格式正确
2. 子配置文件路径是否正确
3. 查看启动日志中的错误信息

### 数据库连接失败

检查：
1. MySQL 服务是否运行
2. 连接配置是否正确（IP、端口、用户名、密码）
3. 数据库是否存在
4. 网络连接是否正常

### 服务启动失败

检查：
1. 端口是否被占用
2. 配置文件中的 IP 和端口是否正确
3. 查看日志文件中的错误信息

## License

[添加许可证信息]
