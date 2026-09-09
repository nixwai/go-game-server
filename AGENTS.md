# Go Game Server 开发规范

## 1. 项目介绍

本项目是基于 Go 的认证后端服务，使用 Gin 提供 HTTP API，使用 GORM 访问 MySQL，使用 JWT 完成登录后的身份认证，使用 Argon2id 完成密码哈希。

当前仅实现基础认证能力：健康检查、普通用户注册、用户名密码登录、当前用户信息查询、管理员权限校验和管理员初始化。暂不包含具体业务模块、刷新令牌、邮箱验证和密码找回。

## 2. 技术栈

- Go 1.26+
- Gin
- GORM
- MySQL 8+
- golang-migrate
- JWT HS256
- Argon2id
- Docker Compose
- OpenAPI 3.0

## 3. 目录结构

```text
cmd/
  server/       HTTP 服务入口
  migrate/      数据库迁移入口
  admin-init/   首个管理员初始化入口
app/
  config/       配置加载和校验
  database/     数据库连接
  handler/      HTTP Handler
  middleware/   请求 ID、JWT、角色权限中间件
  model/        持久化模型和领域常量
  dto/          按业务域划分的接口请求与响应 DTO
  repository/   数据访问抽象与 GORM 实现
  response/     统一响应、业务码、错误处理
  router/       路由注册
  security/     Argon2id 和 JWT
  service/      认证业务逻辑
migrations/     版本化 SQL 迁移
docs/           OpenAPI 文档
tests/          按 config、dto、response、security、service、router 划分的测试与集成测试
docker-compose.yml
.env.example
AGENTS.md
```

核心代码必须放在 `app` 下，不使用 `internal`。

## 4. API 规范

### 方法约定

- **GET**：仅用于读取数据，不产生任何状态变更；
- **POST**：用于所有数据变更操作（创建、更新、删除），禁止使用 PUT、PATCH、DELETE。

### 参数传递约定

- **GET**：参数通过 query string 传递；
- **POST**：参数统一通过 JSON Body 传递，包括资源 ID，路径不携带资源 ID；
- 禁止混合使用路径参数与 Body 传递业务参数。

### 路由命名约定

- 所有业务接口统一挂在 `/api/v1` 前缀下；
- 资源集合路径以名词命名（如 `/providers`、`/models`）；
- 操作语义通过路径后缀显式声明，固定使用以下后缀：
  - `/list`：列表查询；
  - `/create`：新增；
  - `/update`：更新；
  - `/delete`：删除；
- 无副作用的单一动作接口（如 `/login`、`/register`、`/me`）以动作词命名，不加后缀；
- 具体路由以 `app/router/router.go` 和 `docs/openapi.yaml` 为准。

### 认证约定

需要认证的接口必须在请求头携带：

`text
Authorization: Bearer <jwt>
`

## 5. HTTP 状态码与业务码

所有接口统一返回 HTTP 200，客户端通过响应体 `code` 字段判断业务结果。响应体结构：

```json
{
  "code": 0,
  "message": "成功",
  "data": {}
}
```

请求追踪 ID 通过 `X-Request-ID` 响应头返回，不在响应体中包含。客户端可通过 `X-Request-ID` 请求头传入自定义追踪 ID，服务端原样回写。

规则：

- HTTP 状态码始终为 200，不通过 HTTP 状态码区分业务结果；
- `code` 表示业务码，成功固定为 0，非 0 表示具体错误类型；
- 参数校验异常使用 1001-1999；
- 认证异常使用 2001-2999；
- 权限异常使用 3001-3999；
- 资源冲突使用 4001-4999；
- 服务端异常使用 9000-9999；
- 服务端异常不得向客户端返回内部错误堆栈或敏感信息。

## 6. 配置与密钥

配置统一通过环境变量提供。`.env.example` 仅允许包含示例值，不得包含真实密码、JWT 密钥或生产连接串。

关键配置：

- `MYSQL_DSN`：GORM 使用的 MySQL DSN；
- `MYSQL_MIGRATE_URL`：golang-migrate 使用的 MySQL URL；
- `JWT_SECRET`：至少 32 字节的高熵密钥；
- `JWT_EXPIRES_IN`：JWT 有效期；
- `ARGON2_*`：Argon2id 参数。

生产环境必须使用密钥管理系统或安全的环境变量注入机制。禁止将密钥写入代码、测试固定值之外的配置文件或日志。

## 7. 数据库迁移与管理员初始化

数据库表名统一使用单数形式，例如用户表使用 `user`，禁止新建复数表名。GORM 连接必须启用单数表名策略，迁移 SQL、初始化脚本、索引和约束命名应与表名保持一致。
禁止使用数据库层面的外键约束（FOREIGN KEY），表间关联关系由应用层维护，保证数据完整性与水平扩展能力。

应用启动不会自动修改数据库结构。执行迁移：

```bash
go run ./cmd/migrate up
go run ./cmd/migrate down
```

本地启动 MySQL：

```bash
docker compose up -d mysql
```

初始化首个管理员：

```bash
go run ./cmd/admin-init
```

命令会交互式读取用户名和密码，生成 Argon2id 哈希后写入数据库。仓库同时提供 `scripts/init_admin.sql.example` SQL 模板，使用前必须替换占位符，禁止提交替换后的真实密码或哈希。禁止在 SQL、命令行参数、代码仓库中保存固定管理员密码。

## 8. 安全规范

- 密码只能以 Argon2id 哈希形式存储；
- JWT 必须校验签名、签发方、过期时间和允许的算法；
- 注册接口不能由客户端创建管理员；
- 禁用用户不能登录；
- 登录失败不得暴露用户是否存在；
- 查询必须使用参数化条件；
- 密码、密码哈希、JWT、数据库连接串不得写入日志；
- 错误响应不得暴露内部堆栈；
- 所有外部输入必须在服务边界校验；
- 生产部署必须使用 HTTPS，并在网关层补充限流和安全响应头。

## 9. 分层与复用规范

- Handler 只负责 HTTP 参数解析、调用 Service 和构造响应；
- Service 负责业务规则，不直接依赖 Gin；
- Repository 负责数据库访问，不承载业务判断；
- Security 只提供密码和令牌能力；
- Model 只定义持久化结构和领域常量，不直接作为 HTTP 响应；
- DTO 负责接口请求和响应结构，按业务域放在 `app/dto` 下；
- Handler 负责将 Service 返回的 Model 映射为响应 DTO；
- Service 返回业务实体或服务结果，不包含 JSON 标签和 HTTP 协议结构；
- 禁止直接将 `model.User` 序列化为接口响应，避免泄露 `PasswordHash`；
- 公共响应和错误码统一复用 `app/response`；
- 新增数据库实体时必须提供 Repository 接口和实现；
- 不复制粘贴认证、权限和错误响应逻辑；
- 不为了复用而提前创建无业务价值的抽象层。

## 10. 代码规范

- 提交前执行 `gofmt`；
- 错误必须显式处理；
- 函数保持单一职责；
- 避免深层嵌套和超大文件；
- 不修改传入对象，优先返回新值；
- 导出的类型、函数和字段应有清晰命名；
- 不使用无意义的全局可变状态；
- 不提交调试代码、临时文件或本地密钥。

### 10.1 注释规范

- 导出的类型、函数、变量和配置字段必须补充 GoDoc 风格注释；
- 注释只解释代码做什么：描述函数、类型或逻辑块的实际行为与意图；
- 注释禁止引用编码规范、项目约定、安全要求、分层规则等代码规范本身；
- 注释禁止复述代码本身或描述语法行为；
- 代码迭代时注释同步更新为当前代码行为的解释，删除过时、重复或已被替代的残留注释；
- 不在注释中解释实现流程、数据流转路径或跨层调用关系；
- 非导出的辅助函数和内部逻辑仅在行为不明显时添加注释。

## 11. 测试规范

测试至少覆盖：

- Argon2id 哈希和校验；
- JWT 生成、解析、过期和非法算法；
- 注册、重复用户名和非法输入；
- 登录成功、错误密码、不存在用户和禁用用户；
- JWT 缺失、无效和过期；
- 普通用户和管理员权限差异；
- HTTP 状态码与业务码；
- 敏感字段不出现在响应中；
- 数据库约束和异常。

执行：

```bash
gofmt -w app cmd tests
go vet ./...
go test ./...
go test -cover ./...
```

目标测试覆盖率不低于 80%。所有测试文件统一放在 `tests` 目录，不在 `app`、`cmd` 业务代码目录中放置 `*_test.go`。单元测试按组件划分为 `tests/config`、`tests/dto`、`tests/response`、`tests/security`、`tests/service`、`tests/router`；涉及数据库行为的集成测试放在 `tests` 根目录并使用 `mysql-test` 容器，不得依赖开发数据库。设置 `MYSQL_TEST_DSN` 后执行 `go test -tags=integration ./tests`。

## 12. 新功能开发流程

1. 明确接口、业务码、数据模型和安全边界；
2. 先编写失败测试；
3. 实现最小可行代码使测试通过；
4. 重构并保持测试通过；
5. 增加迁移文件和 OpenAPI 文档；
6. 执行格式化、静态检查、单元测试和集成测试；
7. 检查 `git diff`，确认没有密钥和敏感信息；
8. 更新本文件中受影响的约定。

## 13. 接口文档同步规范

- `docs/openapi.yaml` 是前端对接的唯一接口契约，必须与代码实现保持一致；
- 任何接口变更（新增接口、修改路径、修改请求或响应结构、修改业务码、修改参数校验规则）必须同步更新 OpenAPI 文档；
- 文档更新与代码实现在同一个提交中完成，禁止拆分为独立提交或延后补充；
- 先更新文档再编写实现代码，文档先行原则；
- PR 审查时必须检查 `docs/openapi.yaml` 是否与 Handler、DTO、Response 的实际行为一致；
- 新增业务码时必须在 `info.description` 的业务码表中同步登记；
- 新增或修改 Schema 时必须补充 `description` 和 `example`；
- 禁止提交与实现不一致的接口文档。

## 14. 规范同步

本文件（AGENTS.md）是项目核心规范的唯一事实来源，必须与代码实现保持一致。当代码变更涉及以下内容时，必须在同一提交中同步更新本文件：

- 响应体结构或业务码变更；
- API 路由约定或参数传递方式变更；
- 目录结构变更；
- 分层职责或复用规则变更；
- 安全约定变更；
- 数据库表名或迁移约定变更；
- 配置项或密钥约定变更；
- 测试策略或覆盖目标变更。

禁止将规范更新拆分为独立提交或延后补充。PR 审查时必须检查本文件是否与实际代码行为一致。
