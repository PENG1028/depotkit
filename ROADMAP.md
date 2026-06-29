# DBManager v0.1 实施路线图

> 基于 Depotly v0.1 现有能力，进化到 DBManager 第一阶段。
> 发布于 2026-06-29

---

## 设计原则

1. **保留 Depotly 执行层** — Docker Compose、Postgres 连接、migration、backup、Redis/S3/Mongo 操作、endpoint provider 全部保留
2. **新增概念层** — Resource、Binding、SecretRef、AccessEndpoint、Operation、AuditLog 是新增的核心抽象
3. **存储层** — SQLite 作为元数据库，轻量零依赖
4. **不重复造轮子** — 不写 SQL 编辑器、不做 pgAdmin、不做 Bytebase
5. **增量交付** — 每个阶段产出可测试、可体验的功能

## v0.1 范围（8 个核心能力）

```
v0.1 做 ✅：
1. Resource 模型
2. Resource Binding
3. SecretRef
4. Runner 查询接口
5. AccessPublisher + Aegis 实现
6. 临时入口记录
7. 操作日志与审计日志
8. 基础影响分析

v0.1 不做 ❌：
- 自动创建数据库
- 自动备份恢复
- 复制库
- 绑定回滚
- 完整 Web UI
- 复杂多租户权限
- SQL 编辑器
- 数据库表浏览器
```

## 整体架构

```
┌──────────────────────────────────────────────┐
│  cmd/ (CLI 层)                                │
│  init, resource, binding, access, serve, ...  │
├──────────────────────────────────────────────┤
│  api/ (HTTP 服务 - depotly serve)             │
│  GET /bindings?service=&env=                  │
│  Runner 注入接口                              │
├──────────────────────────────────────────────┤
│  pkg/resource/  (资源模型 + CRUD)             │
│  pkg/binding/   (绑定管理)                     │
│  pkg/secretref/ (secret 引用)                  │
│  pkg/access/    (AccessPublisher 接口)         │
│  pkg/audit/     (操作日志 + 审计日志)           │
│  pkg/impact/    (影响分析)                     │
│  pkg/store/     (SQLite 持久层)                │
├──────────────────────────────────────────────┤
│  pkg/postgres/ pkg/redis/ pkg/object/         │
│  pkg/mongo/ pkg/docker/ pkg/endpoint/         │
│  ↑ 保持不动，作为底层执行能力来源               │
└──────────────────────────────────────────────┘
```

## 阶段计划

### Phase 1: 基础设施（约 1 周）

**目标：元数据存储 + Resource 基本模型**

```
新增文件：
  pkg/store/
    store.go          - SQLite 打开/关闭/迁移
    models.go         - Resource / Binding / AccessEndpoint schema
    resource_store.go - Resource CRUD

  cmd/resource.go       - resource 子命令（list, show, create, delete）

修改文件：
  cmd/init.go           - init 时初始化 SQLite 元数据库
  pkg/config/config.go  - 添加 metadata.db 路径
```

**验收标准：**
- `depotly init` 创建 `.depotly/dbmanager.db`
- `depotly resource create --kind postgres --name demo-prod` 创建资源
- `depotly resource list` 列出资源
- `depotly resource show <id>` 查看资源详情

### Phase 2: SecretRef + Binding（约 1 周）

**目标：secret 引用机制 + 服务绑定系统**

```
新增文件：
  pkg/secretref/
    secretref.go   - SecretRef 结构体 + 解析/验证

  pkg/binding/
    binding.go     - Binding CRUD
    binding_store.go - Binding SQLite 持久化

  cmd/binding.go    - binding 子命令（list, create, delete, show）
```

**验收标准：**
- `depotly binding create --service poofnote --resource demo-prod --env staging --env-key DATABASE_URL`
- `depotly binding list --service poofnote`
- `depotly binding list --resource demo-prod` — 显示资源被哪些服务使用
- secret_ref 格式：`secret://<project>/<env>/<service>/<key>`

### Phase 3: Runner 注入 API（约 1 周）

**目标：depotly serve 命令 + HTTP 查询接口**

```
新增文件：
  cmd/serve.go         - HTTP 服务启动命令
  api/
    router.go          - 路由注册
    binding_handler.go - GET /bindings?service=&env=

  pkg/env/
    resolver.go        - secret_ref 解析为真实值
```

**验收标准：**
- `depotly serve --addr :8080` 启动 HTTP 服务
- `GET /api/v1/bindings?service=poofnote&env=staging` 返回：

```json
{
  "service": "poofnote",
  "environment": "staging",
  "status": "ready",
  "env": {
    "DATABASE_URL": "secret://depotly/staging/poofnote/DATABASE_URL"
  }
}
```

### Phase 4: AccessPublisher + Aegis（约 1 周）

**目标：入口发布抽象 + Aegis API 集成**

```
新增文件：
  pkg/access/
    publisher.go      - AccessPublisher 接口定义
    aegis.go          - Aegis HTTP/TCP 入口发布实现
    endpoint.go       - AccessEndpoint 模型
    endpoint_store.go - 入口 SQLite 持久化
    noop.go           - NoopPublisher（测试用）

  cmd/access.go        - access 子命令（publish, revoke, list）
```

**验收标准：**
- `depotly access publish-tcp --resource demo-prod --ttl 30m` 发布临时入口
- `depotly access list` 列出所有活跃入口
- `depotly access revoke <id>` 撤销入口
- Aegis 实现可配置 API endpoint + API key

### Phase 5: 操作日志 + 审计日志 + 基础影响分析（约 1 周）

**目标：操作可追溯 + 删除安全保护**

```
新增文件：
  pkg/audit/
    operation.go     - Operation 模型 + 记录
    audit.go         - AuditLog 模型 + 记录
    operation_store.go - SQLite 持久化

  pkg/impact/
    analyzer.go      - 影响分析引擎
    check_binding.go - 检查是否被绑定
    check_access.go  - 检查是否有活跃入口
    check_prod.go    - 检查是否生产资源
    report.go        - 输出影响报告

  cmd/audit.go        - audit 子命令
  cmd/impact.go       - impact analyze 子命令
```

**验收标准：**
- 每次资源/绑定/入口变更自动记录 Operation
- Audit Log 可追溯"谁在什么时候做了什么"
- `depotly impact analyze --resource demo-prod` 输出：

```
⚠ 影响分析：demo-prod (postgres)

  绑定数：        2 个服务
  活跃入口：      1 个
  生产资源：      是
  风险等级：      高

  影响服务：
    - poofnote (staging, DATABASE_URL)
    - blog (production, DATABASE_URL)

  活跃入口：
    - dbtmp_8f3a (过期 2026-06-30 03:00+08)

  建议：解除所有绑定并撤销入口后再删除。
```

### Phase 6: 集成 + 测试 + 文档（约 1 周）

**目标：全链路可用 + 自动化测试覆盖**

```
修改文件：
  cmd/init.go     - 增强 init，初始化元数据库 + 绑定内置服务
  cmd/root.go     - 合并 DBManager 子命令

新增文件：
  test_dbmanager.sh - 集成测试脚本

文档：
  README.md - 更新
  DBMANAGER_SPEC.md (已完成)
  GAP_ANALYSIS.md   (已完成)
```

**验收标准：**
- 全链路：init → serve → binding create → access publish → impact analyze 完整流程
- 回归测试覆盖：Depotly 原有 35 项测试 + DBManager 新增测试
- 所有资源/绑定/入口变更都记录审计日志

## 文件新增/修改总览

### 新增文件

```
pkg/store/store.go             # SQLite 引擎
pkg/store/models.go            # 数据库 schema
pkg/store/resource_store.go    # Resource CRUD
pkg/store/binding_store.go     # Binding CRUD
pkg/store/access_store.go      # AccessEndpoint CRUD
pkg/store/operation_store.go   # Operation CRUD
pkg/resource/resource.go       # Resource 模型
pkg/binding/binding.go         # Binding 模型
pkg/secretref/secretref.go     # SecretRef 模型
pkg/access/publisher.go        # AccessPublisher 接口
pkg/access/aegis.go            # Aegis 实现
pkg/access/endpoint.go         # AccessEndpoint 模型
pkg/access/noop.go             # Noop 实现
pkg/env/resolver.go            # 环境变量注入解析
pkg/audit/operation.go         # 操作日志
pkg/audit/audit.go             # 审计日志
pkg/impact/analyzer.go         # 影响分析引擎
pkg/impact/check_*.go          # 各维度检查
pkg/impact/report.go           # 报告输出
api/router.go                  # HTTP 路由
api/binding_handler.go         # 绑定查询处理
cmd/resource.go                # resource CLI
cmd/binding.go                 # binding CLI
cmd/access.go                  # access CLI
cmd/audit.go                   # audit CLI
cmd/impact.go                  # impact analyze CLI
cmd/serve.go                   # HTTP 服务启动
test_dbmanager.sh              # 集成测试
```

**预估新增：约 60-80 个文件**

### 修改文件

```
cmd/init.go     # 增加元数据库初始化
cmd/root.go     # 注册新子命令
pkg/config/config.go   # 增加元数据库配置项
go.mod          # 增加 sqlite 依赖
README.md       # 更新
```

## 依赖新增

```
modernc.org/sqlite   # SQLite driver（CGo-free, 纯 Go 实现）
```

## 技术风险

| 风险 | 缓解 |
|------|------|
| SQLite 并发写冲突 | CLI 场景基本单用户，HTTP API 场景用 WAL 模式 + 单 writer |
| Aegis API 未就绪 | AccessPublisher 接口支持 NoopPublisher，Aegis 实现可延迟到集成测试 |
| SecretRef 解析安全 | 暂不实现 secret 存储，只做引用格式。真实凭据仍在 depotly.yaml |
| 影响分析准确性 | v0.1 基于元数据检查（绑定数/入口数），不做运行时探测 |
| 与现有 config 的集成 | 混合模式：depotly.yaml 仍作为 Docker 配置，DBManager 元数据库作为资源抽象 |

## 总计估算

| 阶段 | 内容 | 预估时间 |
|------|------|----------|
| Phase 1 | 基础设施 + Resource 模型 | 5-7 天 |
| Phase 2 | SecretRef + Binding | 5-7 天 |
| Phase 3 | Runner 注入 API | 5-7 天 |
| Phase 4 | AccessPublisher + Aegis | 5-7 天 |
| Phase 5 | 审计日志 + 影响分析 | 5-7 天 |
| Phase 6 | 集成 + 测试 + 文档 | 5-7 天 |
| **总计** | | **30-42 天（约 6-8 周）** |
