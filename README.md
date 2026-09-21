# Go Game Server

基于 Go 的后端服务，使用 Gin 提供 HTTP API，GORM 访问 MySQL，JWT 完成身份认证，Argon2id 完成密码哈希。当前包含基础认证、AI 模型配置管理和 AI 围棋对弈模块。

## 技术栈

| 组件 | 说明 |
|------|------|
| Go 1.26+ | 编程语言 |
| Gin | HTTP 框架 |
| GORM | ORM（单数表名策略） |
| MySQL 8+ | 数据库 |
| golang-migrate | 版本化 SQL 迁移 |
| JWT (HS256) | 登录后身份认证 |
| Argon2id | 密码哈希 |
| AES-256-GCM | AI API Key 加密存储 |
| Docker Compose | 本地 MySQL 环境 |
| OpenAPI 3.0 | 接口契约文档 |

## 快速开始

### 前置条件

- Go 1.26+
- Docker / Docker Compose
- MySQL 8+（可使用 Docker Compose 提供）

### 1. 克隆仓库

```bash
git clone <repo-url>
cd go-game-server
```

### 2. 准备环境变量

```bash
cp .env.example .env
```

编辑 `.env`，替换以下关键配置为本地实际值：

- `MYSQL_DATABASE`、`MYSQL_USER`、`MYSQL_PASSWORD`、`MYSQL_ROOT_PASSWORD` — Docker Compose 初始化数据库所需配置
- `MYSQL_IMAGE` — MySQL 镜像及版本，必须与已有数据目录兼容
- `MYSQL_PORT`、`HTTP_PORT` — 宿主机端口映射
- `MYSQL_DSN` — 宿主机直接运行服务时使用的 GORM 数据源名称
- `MYSQL_MIGRATE_URL` — 宿主机直接运行迁移命令时使用的 golang-migrate 连接 URL
- `JWT_SECRET` — 至少 32 字节的高熵随机密钥
- `MASTER_KEY` — AES-256-GCM 主密钥（base64 编码的 32 字节随机值）
- `DEFAULT_AI_API_KEY` — 默认 AI 产商的 API Key

使用 Docker Compose 时，`app` 和 `migrate` 服务会根据 `MYSQL_DATABASE`、`MYSQL_USER` 和 `MYSQL_PASSWORD` 自动生成容器内连接配置，覆盖 `.env` 中的 `MYSQL_DSN` 和 `MYSQL_MIGRATE_URL`。直接在宿主机运行 `go run ./cmd/server` 或 `go run ./cmd/migrate` 时，才使用这两个连接配置。

> ⚠️ 禁止将真实密钥、密码或连接串提交到代码仓库。`.env` 已在 `.gitignore` 中忽略。

### 3. 启动 Docker Compose 服务

```bash
docker compose up --build -d
```

Compose 会先启动 MySQL，等待健康检查通过，再执行一次性数据库迁移，迁移成功后启动项目服务。Compose 使用 `.env` 中的变量进行配置替换；`app` 服务加载 `.env`，`migrate` 服务使用 Compose 根据数据库配置生成的连接 URL。

仅启动数据库：

```bash
docker compose up -d mysql
```

### 4. 执行数据库迁移

`docker compose up --build -d` 会自动执行 `migrate up`。仅使用数据库容器时，可在宿主机执行：

```bash
go run ./cmd/migrate up
```

回滚最近一次迁移：

```bash
docker compose run --rm migrate down
```

### 5. 初始化首个管理员

```bash
go run ./cmd/admin-init
```

命令会交互式读取用户名和密码，生成 Argon2id 哈希后写入数据库。明文密码仅在进程内短暂存在。

> 也可以使用 `scripts/init_admin.sql.example` SQL 模板手动初始化，使用前必须替换占位符，禁止提交替换后的文件。

### 6. 本地直接启动 HTTP 服务（可选）

完整 Compose 已启动 `app` 时无需重复执行。仅使用数据库容器时：

```bash
docker compose up -d mysql
go run ./cmd/server
```

服务默认监听 `:8080`，可通过 `HTTP_ADDR` 环境变量修改。

收到 `SIGINT` / `SIGTERM` 信号后执行优雅停机，超时 10 秒。

## 项目结构

```text
cmd/
  server/       HTTP 服务入口
  migrate/      数据库迁移入口（up / down）
  admin-init/   首个管理员初始化入口
app/
  bootstrap/    应用组装与生命周期管理
  config/       配置加载、校验和 .env 读取
  database/     MySQL 连接初始化
  ginext/       Gin 框架扩展工具（统一响应包装）
  middleware/   请求 ID、JWT 鉴权、角色权限
  model/        持久化模型和领域常量（不直接作为 HTTP 响应）
  module/       按业务域划分的自包含模块
  response/     统一响应体、业务码、错误处理
  router/       路由编排与中间件挂载
  security/     Argon2id 密码哈希、JWT 签发与校验、AES 加解密
migrations/     版本化 SQL 迁移文件
docs/           OpenAPI 3.0 接口契约文档
scripts/        辅助脚本（管理员初始化 SQL 模板、Linux 打包脚本）
deploy/         部署配置（systemd 服务文件）
tests/          测试代码（不在 app/cmd 中放置 _test.go）
Dockerfile
docker-compose.yml
.dockerignore
.env.example
AGENTS.md
```

## 命令入口

| 命令 | 用途 |
|------|------|
| `go run ./cmd/server` | 启动 HTTP 服务 |
| `go run ./cmd/migrate up` | 执行数据库迁移 |
| `go run ./cmd/migrate down` | 回滚最近一次迁移 |
| `go run ./cmd/admin-init` | 交互式创建首个管理员 |

## 配置说明

所有配置通过环境变量提供，`.env.example` 包含全部可配置项及注释。

详细内容见配置示例文件。

## API 文档查询

项目使用 OpenAPI 3.0 作为接口契约文档，具体业务接口定义请通过以下方式获取：

### 1. 在线访问

服务启动后，可直接通过 HTTP 获取完整 OpenAPI 文档：

```bash
curl http://localhost:8080/docs/openapi.yaml
```

浏览器访问 `http://localhost:8080/docs/openapi.yaml` 即可查看。

### 2. 本地文件

文档源文件位于仓库内：

```
docs/openapi.yaml
```

可直接用编辑器或 Swagger Editor 打开查看。

### 3. Swagger UI（可选）

如需可视化交互界面，可使用 Docker 启动 Swagger UI：

```bash
docker run -p 8081:8080 -e SWAGGER_JSON=/docs/openapi.yaml -v $(pwd)/docs:/docs swaggerapi/swagger-ui
```

浏览器访问 `http://localhost:8081` 即可查看交互式文档。

> 接口变更时必须同步更新 `docs/openapi.yaml`，文档与代码实现保持一致，禁止提交与实现不一致的文档。

## 分层架构

每个业务模块自包含 Handler、Service、Repository、DTO 和路由注册，按业务域放在 `app/module/<域名>` 下：

```text
HTTP 请求
  │
  ▼
Middleware (RequestID → AuthRequired → AdminOnly)
  │
  ▼
Handler ── 参数解析、DTO 绑定、调用 Service、构造响应
  │
  ▼
Service ── 业务规则、输入校验、密码哈希、JWT 签发
  │
  ▼
Repository ── 数据库访问（参数化查询，不承载业务判断）
  │
  ▼
Model ── 持久化实体（不直接序列化为 HTTP 响应）
```

**关键约束：**

- Handler 不含业务逻辑，只做协议转换
- Service 不依赖 Gin，返回领域实体而非 HTTP 结构
- Repository 只做数据访问，不承载业务判断
- `model.User` 禁止直接序列化为接口响应，必须通过 DTO 映射脱敏（`PasswordHash` 永不出现在响应中）
- 统一响应和业务码复用 `app/response`
- 跨模块共享的数据访问错误复用 `app/model`

## 响应格式

所有接口统一返回 HTTP 200，前端通过 `code` 字段判断业务结果：

```json
{
  "code": 0,
  "message": "成功",
  "data": {}
}
```

请求追踪 ID 通过 `X-Request-ID` 响应头返回，不在响应体中包含。客户端可通过 `X-Request-ID` 请求头传入自定义追踪 ID，服务端原样回写。

业务码区间：

| 区间 | 用途 |
|------|------|
| `0` | 成功 |
| `1001-1999` | 参数校验异常 |
| `2001-2999` | 认证异常 |
| `3001-3999` | 权限异常 |
| `4001-4999` | 资源冲突 |
| `9000-9999` | 服务端异常 |

> 服务端异常不向客户端返回内部错误堆栈或敏感信息。

## 数据库

- 表名统一使用**单数形式**（`user`，非 `users`）
- GORM 连接启用 `SingularTable: true`
- 迁移 SQL、索引和约束命名与表名保持一致
- **禁止使用数据库外键约束**（FOREIGN KEY），表间关联关系由应用层维护
- 应用启动不会自动修改数据库结构，必须显式执行迁移

## 测试

### 单元测试

```bash
go test ./...
```

### 集成测试

集成测试使用独立测试数据库，不依赖开发数据库，开发 Compose 不提供测试库。通过 `.env` 之外的进程环境注入测试连接串。

Unix shell：

```bash
export MYSQL_TEST_DSN="<test-user>:<test-password>@tcp(<test-host>:3306)/<test-database>?charset=utf8mb4&parseTime=True&loc=Local"
go test -tags=integration ./tests
```

Windows PowerShell：

```powershell
$env:MYSQL_TEST_DSN = "<test-user>:<test-password>@tcp(<test-host>:3306)/<test-database>?charset=utf8mb4&parseTime=True&loc=Local"
go test -tags=integration ./tests
```

### 代码格式化与静态检查

```bash
gofmt -w app cmd tests
go vet ./...
```

### 覆盖率

```bash
go test -cover ./...
```

目标覆盖率不低于 80%。

测试目录根据开发代码文件结构进行划分。

## 安全要点

- 密码只以 Argon2id 哈希形式存储，使用独立随机盐
- JWT 强制校验签名算法（HS256）、签发方、过期时间
- 注册接口不允许客户端创建管理员
- 禁用用户不能登录
- 登录失败不暴露用户是否存在（统一返回 `用户名或密码错误`）
- 所有数据库查询使用参数化条件
- 密码、哈希、JWT、连接串不写入日志
- AI API Key 通过 AES-256-GCM 加密存储，传输时使用 RSA 公钥加密
- 错误响应不暴露内部堆栈
- 所有外部输入在 Service 边界校验
- 生产环境必须使用 HTTPS，并在网关层补充限流和安全响应头

## Docker Compose 服务部署

| 服务 | 宿主机端口 | 用途 |
|------|------------|------|
| `mysql` | `${MYSQL_PORT}` | 开发数据库 |
| `migrate` | 无 | 一次性执行数据库迁移 |
| `app` | `${HTTP_PORT}` | HTTP 服务 |

默认使用命名卷 `mysql_data`；设置 `MYSQL_DATA_PATH=./data/mysql` 后改用宿主机目录挂载。`data/` 已从 Git 和镜像构建上下文排除。默认数据库镜像为 `mysql:8.4` LTS；已有数据目录必须通过 `MYSQL_IMAGE` 指定兼容版本。镜像按目标 CPU 架构构建，支持 AMD64 和 ARM64。敏感配置统一从 `.env` 注入，不写入 `Dockerfile` 或 `docker-compose.yml`。

### 创建启动

```bash
docker compose up --build -d
```

### 暂时停止

```bash
docker compose stop
```

### 重新启动

```bash
docker compose start
```

### 停止并删除容器

```bash
docker compose down
```

该命令不会删除默认的 `mysql_data` 数据卷。如需同时删除数据库数据，执行 `docker compose down -v`。

## systemd 服务部署（Linux）

systemd 部署适用于将 HTTP 服务作为 Linux 系统服务运行。该方式只管理 `server` 进程，MySQL 必须由独立的数据库服务提供，不能与 Compose 的 `app` 服务同时占用同一个 `HTTP_PORT`。

### 部署目录要求

```bash
cd /opt/go-game-server
```

`deploy/systemd/go-game-server.service` 默认使用以下目录和文件：

```text
/opt/go-game-server/
├── .env
├── bin/server
├── bin/migrate
├── bin/admin-init
├── docs/openapi.yaml
└── migrations/
```

`.env` 必须包含宿主机运行服务所需的 `MYSQL_DSN`、`JWT_SECRET`、`MASTER_KEY` 等配置。服务以 `go-game` 用户运行，部署前需要确保该用户可以读取部署目录。

### 首次部署

#### 1. 准备环境配置

如果项目中还没有 `.env`，从示例文件创建并填写实际配置：

```bash
sudo cp .env.example .env
sudo vi .env
```

> `.env` 包含数据库连接串、JWT 密钥等敏感配置，禁止提交到代码仓库或公开传播。

#### 2. 构建 Linux 二进制文件

在服务器上执行以下命令，产物会直接生成到项目的 `bin` 目录：

```bash
go build -trimpath -ldflags "-s -w" -o ./bin/server ./cmd/server
go build -trimpath -ldflags "-s -w" -o ./bin/migrate ./cmd/migrate
go build -trimpath -ldflags "-s -w" -o ./bin/admin-init ./cmd/admin-init
```

Windows 构建 Linux 二进制：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\build-linux.ps1 -Architecture amd64
```

可选 `arm64`。构建产物需要放入服务器项目目录的 `bin` 目录。

如果使用已经包含 Linux 二进制文件的发布包，可以跳过构建步骤，但必须确认 `bin/server`、`bin/migrate` 和 `bin/admin-init` 已存在且适用于服务器 CPU 架构。

#### 3. 创建 systemd 运行用户

使用独立的系统用户运行服务，避免服务直接使用 `root` 权限。`/usr/sbin/nologin` 禁止该用户用于交互式登录。

```bash
sudo useradd --system --home-dir /opt/go-game-server --shell /usr/sbin/nologin go-game
```

如果 `go-game` 用户已经存在，跳过此命令。

#### 4. 设置目录所有者和文件权限

让 `go-game` 用户拥有项目目录，确保 systemd 启动的服务能够读取配置、二进制文件、接口文档和数据库迁移文件。

```bash
sudo chown -R go-game:go-game /opt/go-game-server
sudo chmod 600 /opt/go-game-server/.env
sudo chmod 755 /opt/go-game-server/bin/server /opt/go-game-server/bin/migrate /opt/go-game-server/bin/admin-init
```

#### 5. 安装 systemd 服务配置

将项目中的服务配置安装到 systemd 的系统级服务目录。该配置定义服务用户、工作目录、启动命令、自动重启和日志输出方式。

```bash
sudo cp /opt/go-game-server/deploy/systemd/go-game-server.service /etc/systemd/system/
```

#### 6. 重新加载 systemd 配置

通知 systemd 重新读取刚刚安装的服务配置。新增或修改 `.service` 文件后都需要执行。

```bash
sudo systemctl daemon-reload
```

#### 7. 执行数据库迁移

使用 `go-game` 用户执行数据库迁移，避免迁移产生由 `root` 所有的文件。

```bash
sudo -u go-game ./bin/migrate up
```

#### 8. 启动并设置开机自启

```bash
sudo systemctl enable --now go-game-server
sudo systemctl status go-game-server
```

### 服务管理

```bash
# 查看实时日志
sudo journalctl -u go-game-server -f

# 重启服务
sudo systemctl restart go-game-server

# 停止服务
sudo systemctl stop go-game-server

# 禁止开机自动启动
sudo systemctl disable go-game-server
```

### 发布新版本

在项目根目录重新构建并替换二进制文件后，重启服务：

```bash
sudo install -o go-game -g go-game -m 755 ./bin/server /opt/go-game-server/bin/server
sudo cp docs/openapi.yaml /opt/go-game-server/docs/openapi.yaml
sudo chown go-game:go-game /opt/go-game-server/docs/openapi.yaml
sudo systemctl restart go-game-server
```

如迁移文件发生变化，先同步 `migrations` 目录和 `migrate` 二进制，再执行：

```bash
sudo -u go-game ./bin/migrate up
```

> systemd 服务和 Compose 的 `app` 服务是两种互斥的 HTTP 服务运行方式。使用 systemd 时可通过 `docker compose up -d mysql` 仅启动数据库，但不要同时启动 Compose 的 `app` 服务。

## 开发流程

1. 明确接口契约、业务码、数据模型和安全边界
2. 先更新 `docs/openapi.yaml`（文档先行原则）
3. 编写失败测试
4. 实现最小可行代码使测试通过
5. 重构并保持测试通过
6. 增加迁移文件和 OpenAPI 文档
7. 执行格式化、静态检查、单元测试和集成测试
8. 检查 `git diff`，确认没有密钥和敏感信息
9. 更新 `AGENTS.md` 中受影响的约定

## 相关文档

- [AGENTS.md](AGENTS.md) — 开发规范与约定
- [docs/openapi.yaml](docs/openapi.yaml) — OpenAPI 3.0 接口契约
- [.env.example](.env.example) — 环境变量示例与注释
