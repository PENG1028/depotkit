# Gap Analysis: Depotly v0.1 → DBManager

> 分析日期：2026-06-29
> 目标：评估当前 Depotly 与 DBManager 愿景之间的差距，确定 v0.1 实施范围

---

## 总览

| 维度 | Depotly v0.1 现状 | DBManager 目标 |
|------|-------------------|----------------|
| 代码行数 | ~4,700 Go | 预估 ~25,000–40,000 |
| 包数量 | 8 (config/docker/postgres/redis/object/mongo/endpoint/utils) | 预估 20+ |
| 架构层 | Docker 容器编排 + 简易 CLI | 资源控制层 + 绑定 + 入口管理 |
| 数据存储 | YAML 文件 | 需要内部元数据库 (SQLite) |
| API | CLI 命令行 | CLI + HTTP API（至少为 Runner 暴露接口） |
| 概念 | 服务配置 (ServiceConfig) | 资源 (Resource) + 绑定 (Binding) + secret_ref |

## 各模块差距矩阵

### 1. 核心对象（所有缺失 ❌）

| DBManager 对象 | Depotly 现状 | 说明 |
|----------------|-------------|------|
| **Resource** | ❌ 不存在 | 当前只有 `ServiceConfig` 描述 Docker 容器，没有资源抽象 |
| **Resource Binding** | ❌ 不存在 | 没有 service→resource 绑定的概念 |
| **Secret Ref** | ❌ 不存在 | 密码是 depotly.yaml 中的明文 `string` |
| **Access Endpoint** | ❌ 不存在 | 有 `ExposureManifest` 声明结构体，无运行时入口记录 |
| **Operation** | ❌ 不存在 | 无操作记录系统 |
| **Audit Log** | ❌ 不存在 | 无审计日志系统 |
| **Tenant/Project/Env** | ❌ 不存在 | 无任何隔离维度 |

### 2. 功能域对比

#### 资源登记（0%）
| 能力 | 现状 | 优先级 |
|------|------|--------|
| Postgres 登记 | 配置中存在 | v0.1 ✅ |
| Redis 登记 | 配置中存在 | v0.1 ✅ |
| SQLite 登记 | ❌ 不存在 | v0.1 可选 |
| external_unmanaged | ❌ 不存在 | v0.1 ✅ |
| 隐藏真实连接信息 | ❌ | v0.1 ✅（secret_ref 核心） |
| 标记环境/服务 | ❌ | v0.1 ✅ |
| secret_ref 生成 | ❌ | v0.1 ✅ |

#### 服务绑定（0%）
| 能力 | 现状 | 优先级 |
|------|------|--------|
| service→database 绑定 | ❌ | v0.1 ✅ |
| environment→resource | ❌ | v0.1 ✅ |
| env_key 管理 | ❌ | v0.1 ✅ |
| access_role (read_only/read_write) | ❌ | v0.1 ✅ |
| secret_ref 关联 | ❌ | v0.1 ✅ |

#### Runner 注入 API（0%）
| 能力 | 现状 | 优先级 |
|------|------|--------|
| HTTP 查询接口 | ❌ CLI only | v0.1 ✅ |
| env key 映射返回 | ❌ | v0.1 ✅ |
| 缺少资源阻止启动标志 | ❌ | v0.1 ✅ |

#### AccessPublisher + Aegis（0%）
| 能力 | 现状 | 优先级 |
|------|------|--------|
| AccessPublisher 接口 | ❌ 当前是 ExposureProvider | v0.1 ✅ |
| Aegis 入口发布 | ❌ manifest-only | v0.1 ✅ |
| 临时 TCP 入口 | ❌ | v0.1 ✅ |
| 入口生命周期管理 | ❌ | v0.1 ✅ |

#### 临时入口记录（0%）
| 能力 | 现状 | 优先级 |
|------|------|--------|
| 创建记录 | ❌ | v0.1 ✅ |
| 过期时间 | ❌ | v0.1 ✅ |
| 资源关联 | ❌ | v0.1 ✅ |
| 撤销/关闭 | ❌ | v0.1 ✅ |
| 审计追溯 | ❌ | v0.1 ✅ |

#### 操作日志 + 审计日志（0%）
| 能力 | 现状 | 优先级 |
|------|------|--------|
| 操作记录存储 | ❌ | v0.1 ✅ |
| 审计日志 | ❌ | v0.1 ✅ |
| 查询/筛选 | ❌ | v0.1 基本 |

#### 基础影响分析（0%）
| 能力 | 现状 | 优先级 |
|------|------|--------|
| 删除资源前检查绑定 | ❌ | v0.1 ✅ |
| 检查活跃入口 | ❌ | v0.1 ✅ |
| 检查生产标记 | ❌ | v0.1 ✅ |
| 输出影响报告 | ❌ | v0.1 ✅ |

### 3. 可复用的 Depotly 能力

以下代码可以直接保留，作为 DBManager 的**底层执行层**：

| 包 | 文件 | 行数 | 用途 |
|----|------|------|------|
| `pkg/postgres/` | client.go | ~50 | PostgreSQL 连接 — 可用于健康检查和资源控制操作 |
| `pkg/postgres/` | migration.go | 245 | 带脏状态检测的迁移 — migration 关联记录的底层 |
| `pkg/postgres/` | backup.go | ~80 | pg_dump/restore — 备份记录的执行层 |
| `pkg/redis/` | client.go | ~50 | Redis 连接 |
| `pkg/redis/` | ops.go | ~86 | Redis 操作（scan/flush/ttl） |
| `pkg/object/` | client.go + ops.go | ~160 | MinIO/S3 操作 |
| `pkg/mongo/` | client.go + ops.go | ~170 | MongoDB 操作 |
| `pkg/docker/` | compose.go + docker.go | ~250 | Docker Compose 编排 — Docker 模式启动方式 |
| `pkg/endpoint/` | manifest.go | 213 | 资源元数据结构 — `InstanceInfo` 可演化为 `Resource` |
| `pkg/endpoint/` | provider.go | 35 | `ExposureProvider` 接口 — 可扩展为 `AccessPublisher` |
| `cmd/` | 全部 CLI | ~2,000 | 大部分命令可在 DBManager CLI 中保留 |
| `pkg/utils/` | confirm.go | ~30 | 确认提示逻辑 |

## 架构变化示意

```
当前 Depotly 结构：

  depotly.yaml (配置)
      ↓
  cmd/ (CLI 入口)
      ↓
  pkg/postgres/ pkg/redis/ pkg/object/ pkg/mongo/  (执行层)
      ↓
  Docker 容器 / 数据库连接


目标 DBManager 结构：

  dbmanager.yaml (注册数据源 / 元数据库配置)
      ↓
  cmd/ (CLI 入口) + api/ (HTTP 服务)
      ↓
  pkg/resource/   (资源模型 + CRUD)
  pkg/binding/    (绑定管理)
  pkg/secretref/  (secret 引用管理)
  pkg/access/     (AccessPublisher 接口 + Aegis 实现)
  pkg/audit/      (操作日志 + 审计日志)
  pkg/impact/     (影响分析)
  pkg/store/      (元数据持久层 - SQLite)
      ↓
  pkg/postgres/ pkg/redis/ pkg/object/ pkg/mongo/  (执行层 - 从 Depotly 继承)
  pkg/docker/                                       (从 Depotly 继承)
      ↓
  Docker 容器 / 数据库连接
```

## 关键架构决策

| 决策 | 选项 | 建议 |
|------|------|------|
| 元数据存储 | SQLite 文件 / Depotly 元数据库 | **SQLite 文件** — 零依赖，适合个人/小团队 |
| 项目结构 | 单仓库 / 多仓库 | **单仓库** — 与 Depotly 同仓库，pkg/ 分层隔离 |
| 命名 | dbmanager / depotly 扩展 | **depotly 扩展** — 保持品牌延续，新增 pkg/ 包 |
| HTTP API | 内嵌 server / 独立进程 | **内嵌** — depotly serve 命令启动 API 服务 |
| 对象关系 | 同仓库不同 namespace / 独立 module | **同 module** — `github.com/depotly/depotly` |

## 量化评估

| 能力域 | 完成度 | 预估工作量 |
|--------|--------|-----------|
| Depotly 现有能力保留（Docker + 迁移 + CLI） | 100% | 0 |
| 内部存储层（SQLite 元数据库） | 0% | 5-7 天 |
| Resource 模型 + CLI | 0% | 5-7 天 |
| Secret Ref 管理 | 0% | 3-5 天 |
| Service Binding 系统 | 0% | 5-7 天 |
| Runner 注入 API (HTTP) | 0% | 5-7 天 |
| AccessPublisher + Aegis 实现 | 0% | 5-7 天 |
| 临时入口管理 | 0% | 3-5 天 |
| 操作日志 + 审计日志 | 0% | 3-5 天 |
| 基础影响分析 | 0% | 3-5 天 |
| **v0.1 总计** | **~2%** | **~37-55 天（2-3 个人月）** |
