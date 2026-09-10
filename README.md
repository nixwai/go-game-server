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

- `MYSQL_DSN` — GORM 数据源名称
- `MYSQL_MIGRATE_URL` — golang-migrate 连接 URL
- `JWT_SECRET` — 至少 32 字节的高熵随机密钥
- `MASTER_KEY` — AES-256-GCM 主密钥（base64 编码的 32 字节随机值）
- `DEFAULT_AI_API_KEY` — 默认 AI 产商的 API Key

> ⚠️ 禁止将真实密钥、密码或连接串提交到代码仓库。`.env` 已在 `.gitignore` 中忽略。

### 3. 启动 MySQL

```bash
docker compose up -d mysql
```

测试数据库（端口 3307）可按需启动：

```bash
docker compose up -d mysql-test
```

### 4. 执行数据库迁移

```bash
go run ./cmd/migrate up
```

回滚最近一次迁移：

```bash
go run ./cmd/migrate down
```

### 5. 初始化首个管理员

```bash
go run ./cmd/admin-init
```

命令会交互式读取用户名和密码，生成 Argon2id 哈希后写入数据库。明文密码仅在进程内短暂存在。

> 也可以使用 `scripts/init_admin.sql.example` SQL 模板手动初始化，使用前必须替换占位符，禁止提交替换后的文件。

### 6. 启动 HTTP 服务

```bash
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
scripts/        辅助脚本（管理员初始化 SQL 模板）
tests/          测试代码（不在 app/cmd 中放置 _test.go）
docker-compose.yml
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

| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `APP_ENV` | 运行环境 | `development` |
| `HTTP_ADDR` | HTTP 监听地址 | `:8080` |
| `MYSQL_DSN` | GORM MySQL DSN | — |
| `MYSQL_MIGRATE_URL` | golang-migrate MySQL URL | — |
| `JWT_SECRET` | HS256 签名密钥（≥32 字节） | — |
| `JWT_ISSUER` | JWT 签发方 | `go-game-server` |
| `JWT_EXPIRES_IN` | Access Token 有效期 | `2h` |
| `ARGON2_TIME` | Argon2id 时间成本 | `1` |
| `ARGON2_MEMORY` | Argon2id 内存成本（KiB） | `65536` |
| `ARGON2_THREADS` | Argon2id 并行线程数 | `4` |
| `ARGON2_KEY_LEN` | Argon2id 哈希长度（字节） | `32` |
| `ARGON2_SALT_LEN` | Argon2id 随机盐长度（字节） | `16` |
| `MASTER_KEY` | AES-256-GCM 主密钥（base64 编码 32 字节） | — |
| `DEFAULT_AI_PROVIDER` | 默认 AI 产商名称 | `OpenAI` |
| `DEFAULT_AI_BASE_URL` | 默认 AI 产商 API 地址 | `https://api.openai.com/v1` |
| `DEFAULT_AI_MODEL` | 默认 AI 模型名称 | `gpt-4o-mini` |
| `DEFAULT_AI_API_KEY` | 默认产商 API Key（仅存内存） | — |
| `GO_LLM_TIMEOUT` | LLM API 调用超时 | `30s` |

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

集成测试使用 `mysql-test` 容器（端口 3307），不依赖开发数据库：

```bash
export MYSQL_TEST_DSN="go_game_test:go_game_test@tcp(127.0.0.1:3307)/go_game_test?charset=utf8mb4&parseTime=True&loc=Local"
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

测试目录划分：

```text
tests/
  config/                 配置加载与 .env 测试
  response/               统一响应与业务码测试
  security/               Argon2id、JWT 与 AES 加解密测试
  bootstrap/              应用启动测试
  module/                 模块测试
  router/                 路由与中间件测试
  mysql_integration_test.go  数据库集成测试
```

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

## Docker Compose 服务

| 服务 | 端口 | 用途 |
|------|------|------|
| `mysql` | 3306 | 开发数据库 |
| `mysql-test` | 3307 | 集成测试数据库 |

数据卷：`mysql_data`、`mysql_test_data`

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
