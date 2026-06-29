# DBManager 定义 — 资源控制层

> 本文档定义 DBManager 在整个服务系统中的定位、边界和核心设计。
> 发布于 2026-06-29

---

## 1. 一句话定义

DBManager 是数据库/Redis/存储资源与服务系统之间的**资源控制层**。

它不负责替代 PostgreSQL、Redis、pgAdmin、Bytebase，也不负责成为业务数据库查询的中转层。

它负责回答：

- 哪个服务使用哪个数据库？
- 哪个环境绑定哪个资源？
- 连接信息怎么隐藏？
- Runner 启动服务时怎么注入 env？
- Aegis 如何发布管理入口或临时访问入口？
- 删除、切换、复制数据库前会影响什么？

## 2. DBManager 不是什么

DBManager **不是**：

- SQL 编辑器
- 数据库表查看器
- pgAdmin 替代品
- Bytebase 替代品
- 数据库高可用系统
- 数据库运行时代理
- 所有数据库连接的唯一入口
- Postgres / Redis 的替代实现

这些事情交给现成工具：

| 工具 | 职责 |
|------|------|
| pgAdmin / DBeaver | 查看表、改表、跑 SQL |
| Bytebase | 复杂数据库变更流程、SQL 审核 |
| Postgres / Redis | 实际数据运行 |
| Aegis | 入口发布、反向代理、临时 TCP 暴露 |
| Runner | 启动服务、注入 env |

**DBManager 的核心价值不是"看数据库里面有什么"，而是"数据库资源在我的服务体系里扮演什么角色"。**

## 3. 核心定位

DBManager 是一个**资源事实来源**。它记录并管理：

- 数据库资源
- Redis 资源
- SQLite 本地资源
- 对象存储资源（后期）
- 队列资源（后期）
- 服务绑定关系
- 环境绑定关系
- secret_ref
- Aegis 访问入口
- Runner env 注入需求
- 备份记录
- 复制库/测试库记录
- 删除前影响分析
- 操作日志
- 审计日志

它的视角是：

- **service-first**
- **resource-first**
- **environment-first**

不是：

- table-first
- SQL-first
- schema-first

## 4. 系统边界

### DBManager 负责

| 能力 | 说明 |
|------|------|
| 资源登记 | 注册/查看/编辑资源，隐藏真实连接信息 |
| 服务绑定 | service→resource 绑定，environment→resource 绑定 |
| 环境隔离 | 环境维度隔离资源与绑定 |
| 连接信息隐藏 | 密码默认不可见，只暴露 secret_ref |
| secret_ref 管理 | 资源凭据引用规范 |
| Runner env 注入接口 | API 供 Runner 查询服务依赖 |
| Aegis 入口发布接口 | 发布/撤销 DBManager UI、pgAdmin、临时入口 |
| 临时数据库访问入口 | 创建/过期/撤销/记录临时 TCP 入口 |
| 删除前影响分析 | 绑定/路由/secret 依赖检查 |
| 绑定切换/回滚记录 | 切换历史，支持回滚 |
| 测试库/复制库/临时库记录 | 申请/记录/绑定/清理 |
| 备份记录 | 策略/时间/位置/状态追踪 |
| migration 关联记录 | 服务版本与 migration 的映射 |
| 操作日志 | 谁在什么时候做了什么 |
| 审计日志 | 完整变更追溯 |

### DBManager 不负责

- 业务服务运行时查询数据库
- 替业务服务转发 SQL 请求
- 长期承载数据库公网连接
- 自动高可用
- 自动主从切换
- 复杂 SQL 审核
- 数据库内部对象管理
- 完整备份恢复平台

## 5. 与其他模块的关系

### 和 Runner 的关系

Runner 负责启动服务。DBManager 负责告诉 Runner：

- 这个服务需要哪些资源
- 这些资源对应哪些 env key
- 这些 env key 对应哪些 secret_ref
- 缺少哪些资源
- 当前绑定是否有效

```
Runner → DBManager : query_binding
Runner → Service    : inject_env
```

DBManager **不直接启动服务**。

### 和 Aegis 的关系

Aegis 是反向代理/入口发布器。DBManager 使用 Aegis 做：

- 发布 DBManager 管理页面
- 发布 pgAdmin/RedisInsight/Bytebase 等工具入口
- 发布临时数据库 TCP 入口
- 关闭临时入口
- 记录入口对应哪个资源

```
DBManager → Aegis : publish_access
Aegis → DBManager : routes
```

Aegis **不应该成为数据库资源事实来源**。Aegis 管"入口怎么访问"；DBManager 管"入口背后是什么数据库资源"。

### 和 Postgres / Redis 的关系

DBManager 可以连接 Postgres / Redis 做控制操作：

- 测试连接
- 创建 database（后期）
- 创建 user（后期）
- 授权（后期）
- 生成连接串（后期）
- 备份记录
- 恢复记录

```
DBManager → Postgres : control
DBManager → Redis    : control
```

业务服务运行时**不经过 DBManager**：

- 正确：`Service → Postgres`
- 错误：`Service → DBManager → Postgres`

### 和 Auth 的关系

Auth 可以保护 DBManager 的管理页面。

需要避免的循环依赖：

- Auth 依赖 DBManager 才能启动
- DBManager 依赖 Auth 才能创建 Auth 的数据库

Auth 的基础身份库应**先静态配置**，不由 DBManager 首次创建。

## 6. 核心对象

### Resource

表示一个数据库、Redis、SQLite、对象存储、队列等资源。

```yaml
resource:
  id: db_poofnote_prod
  kind: postgres
  category: relational
  environment: prod
  owner_service: poofnote
  project_id: project_poofnote
  tenant_id: default

  logical_name: poofnote/prod/postgres
  endpoint_alias: pg-poofnote-prod.db.internal
  secret_ref: secret://poofnote/prod/postgres/write

  status:
    desired: active
    actual: unknown

  physical:
    hidden: true
    node: node_02
    host: 10.0.0.23
    port: 5432
    database: poofnote_prod_8f3a
```

### Resource Binding

表示某个服务在某个环境使用某个资源。

```yaml
binding:
  service: poofnote
  environment: prod
  resource: db_poofnote_prod
  env_key: DATABASE_URL
  access_role: read_write
  secret_ref: secret://poofnote/prod/postgres/write
  required: true
```

### Secret Ref

页面上默认不显示真实密码，只显示 `secret://` 格式。

Runner 启动服务时再解析成真实 env。

### Access Endpoint

表示通过 Aegis 发布的入口。

```yaml
access_endpoint:
  id: dbtmp_8f3a
  owner: dbmanager
  resource_id: db_poofnote_staging
  type: temp_tcp
  host: dbtmp-8f3a.example.com
  port: 15432
  target_host: 10.0.0.23
  target_port: 5432
  expires_at: 2026-06-29T21:00:00+08:00
  status: active
```

### Operation

表示创建、绑定、发布入口、测试连接、备份、恢复、复制等动作。

```yaml
operation:
  id: op_123
  type: publish_temp_tcp_access
  resource: db_poofnote_staging
  status: success
  actor: admin
  started_at: 2026-06-29T20:30:00+08:00
  finished_at: 2026-06-29T20:30:03+08:00
```

### Audit Log

记录谁做了什么。

```yaml
audit:
  actor: admin
  action: switch_binding
  service: poofnote
  environment: staging
  from: db_poofnote_staging
  to: db_poofnote_preview_42
  time: 2026-06-29T20:30:00+08:00
```

## 7. 链接与访问策略

### managed_internal（推荐）

系统管理的内部运行连接。

```
Service → internal endpoint → Database
```

- 用于服务运行
- 由 Runner 注入
- 默认不走公网

### temporary_external

临时外部入口。

```
Admin → Aegis → Database
```

- 临时，有过期时间
- 有 IP 白名单
- 有审计
- 用于调试和管理

### external_unmanaged

用户手动填写的外部数据库。

- DBManager 只登记
- 可以绑定服务
- 可以测试连接
- 但不承诺创建、备份、复制、权限管理

## 8. 隔离模型

即使第一版只有一个人用，字段应保留：

- `tenant_id`
- `project_id`
- `service_id`
- `environment`
- `owner_id`
- `created_by`

功能可以简化：

- 默认 tenant
- 默认 owner
- 项目可以创建
- 服务归属项目
- 资源归属项目和服务

不做：

- 团队邀请
- 组织权限
- 复杂 RBAC
- 计费
- 审批流
- 资源配额

## 9. 支持的资源类型

| 版本 | 资源类型 | 支持级别 |
|------|---------|----------|
| v0.1 | Postgres, Redis, SQLite, external_unmanaged | **L0**（登记、绑定、secret_ref、env 注入）+ **L1 部分**（Postgres/Redis 健康检查） |
| v0.2+ | MySQL/MariaDB, S3/R2/MinIO | L0 |
| 后期 | MongoDB, Elasticsearch/OpenSearch, 向量库, ClickHouse, Kafka/RabbitMQ/NATS | L0 |

### 支持级别定义

| 级别 | 能力 |
|------|------|
| L0 | 登记、绑定、secret_ref、env 注入 |
| L1 | 健康检查 |
| L2 | 创建资源、创建用户、授权 |
| L3 | 备份、恢复、复制、凭据轮换 |

## 10. 对 Aegis 的抽象

不要让 DBManager 直接绑定 Aegis 内部实现。抽象为 **AccessPublisher**：

```go
type AccessPublisher interface {
    PublishHTTPRoute(route HTTPRoute) (AccessEndpoint, error)
    PublishTempTCPAccess(tcp TempTCPAccess) (AccessEndpoint, error)
    RevokeAccess(endpointID string) error
    GetAccessStatus(endpointID string) (AccessEndpoint, error)
    SyncAccessSnapshot() ([]AccessEndpoint, error)
}
```

Aegis 是第一个 AccessPublisher 实现。

DBManager 使用 Aegis API Key，但必须是**受限权限**：

```
service_account: dbmanager
allowed:
  - publish_http_route
  - publish_temp_tcp_access
  - revoke_own_access
scope:
  hosts:
    - dbmanage.example.com
    - dbtmp-*.example.com
  tcp_ports:
    - 15000-15999
denied:
  - auth.example.com
  - runner.example.com
  - aegis-admin.example.com
  - global_routes
```

## 11. 最终定义

> DBManager 是一个面向个人/小团队自托管系统的**数据库与存储资源控制层**。
>
> 它不管理数据库内部数据，而是管理数据库资源在服务体系中的生命周期、绑定关系、连接抽象、secret 引用、运行时注入、访问入口、影响分析、备份记录、复制记录和操作审计。
>
> 它通过 Runner 完成服务启动时的环境变量注入，通过 Aegis 发布管理页面和临时外部访问入口，通过 Postgres/Redis 适配器完成基础连接检查和后续资源创建能力。
>
> DBManager 的目标是让服务使用数据库像申请资源一样清晰，而不是让使用者手动记住真实 IP、端口、数据库名、密码、服务绑定和路由关系。

更短一点：

> **pgAdmin 管数据库里面有什么；**
> **DBManager 管数据库在我的服务系统里属于谁、给谁用、怎么注入、怎么暴露、怎么切换、怎么删除才安全。**
