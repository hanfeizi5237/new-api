# 系统架构与模块关系图

本文档从系统设计角度说明当前项目的整体结构、模块边界和核心请求链路，适合作为以下场景的技术材料：

- 新成员快速理解项目结构
- 做模块重构前先确认边界
- 评审新功能应该放在哪一层
- 排查问题时快速定位入口

配套文档：

- `docs/project-feature-overview.zh-CN.md`：代码与功能能力总览
- `docs/product-feature-list.zh-CN.md`：面向产品和业务方的功能清单
- `docs/developer-reading-guide.zh-CN.md`：面向研发的阅读顺序与源码导读
- `docs/core-data-models.zh-CN.md`：核心数据模型与实体关系梳理
- `docs/core-business-sequences.zh-CN.md`：核心业务链路时序图

## 1. 架构结论

当前项目整体上是一个 **单体部署、模块化分层** 的系统，而不是微服务架构。

它的核心特征如下：

- 单个 Go 服务承载 API、Relay、后台接口和 Web 静态资源
- 后端遵循明确的分层结构：`Router -> Controller -> Service -> Model`
- Relay 子系统是后端中的一条专门链路，用于完成多协议解析、渠道选择、计费和上游转发
- 前端是一个 React 单页应用，由 Go 服务直接托管静态文件，或按配置重定向到外部前端地址
- 配置同时来自环境变量和数据库 `options`
- 数据层兼容 SQLite、MySQL、PostgreSQL

可以把它理解为：

`Gin Monolith + Layered Business Modules + Embedded React SPA + Specialized Relay Engine`

## 2. 高层架构图

```text
                           +----------------------+
                           |      Browser /       |
                           |    API Client /      |
                           |  SDK / CLI / Tools   |
                           +----------+-----------+
                                      |
                                      v
                         +------------+-------------+
                         |         Gin Server       |
                         |  RequestId / I18n / Log  |
                         |   Session / Recovery     |
                         +------------+-------------+
                                      |
               +----------------------+----------------------+
               |                                             |
               v                                             v
     +---------+----------+                       +-----------+----------+
     |   Dashboard/API    |                       |       Relay API      |
     |   /api, /dashboard |                       |   /v1, /v1beta,      |
     |   user/admin ops   |                       |   /mj, /suno, /pg    |
     +---------+----------+                       +-----------+----------+
               |                                              |
               v                                              v
     +---------+----------+                       +-----------+----------+
     |     Controller     |                       |       Controller      |
     | user/channel/...   |                       |        relay.go       |
     +---------+----------+                       +-----------+----------+
               |                                              |
               v                                              v
     +---------+----------+                       +-----------+----------+
     |      Service       |                       |  relay/* + service/* |
     | billing/auth/...   |                       | parse/select/billing |
     +---------+----------+                       +-----------+----------+
               |                                              |
               +----------------------+-----------------------+
                                      |
                                      v
                         +------------+-------------+
                         |       Model / GORM       |
                         | channel/user/token/...   |
                         +------------+-------------+
                                      |
                     +----------------+----------------+
                     |                                 |
                     v                                 v
               +-----+------+                   +------+------+
               | Main DB    |                   | Redis/Cache |
               | SQLite /   |                   | memory/disk |
               | MySQL / PG |                   +-------------+
               +------------+
```

## 3. 启动与初始化链路

系统启动入口在 `main.go`。

### 3.1 启动阶段做了什么

启动过程大致如下：

1. 加载 `.env` 与环境变量
2. 初始化公共环境与日志
3. 初始化模型相关配置
4. 初始化 HTTP 客户端
5. 初始化数据库与日志数据库
6. 初始化 OAuth provider、i18n、缓存与各种后台任务
7. 创建 Gin Server
8. 注册基础中间件、Session
9. 挂载 API/Relay/Web 路由
10. 启动 HTTP 服务

### 3.2 启动相关模块

- `main.go`
- `router/main.go`
- `model/main.go`
- `setting/config/config.go`

### 3.3 启动侧的设计特点

- 启动时会初始化数据库结构和必要数据
- 启动后会开启多个后台任务，而不是完全被动接收请求
- Go 服务既承载后端 API，也承载前端静态资源
- 在 `FRONTEND_BASE_URL` 配置存在时，也支持把 Web 请求重定向到外部前端

## 4. 后端分层结构

项目后端采用清晰的分层架构。

```text
HTTP Request
  -> Router
  -> Middleware
  -> Controller
  -> Service
  -> Model
  -> Database / Cache / Upstream API
```

### 4.1 Router 层

职责：

- 定义 URL 路由
- 按接口类别挂载中间件
- 按请求类型把流量导向 Controller

主要文件：

- `router/main.go`
- `router/api-router.go`
- `router/relay-router.go`
- `router/dashboard.go`
- `router/web-router.go`
- `router/video-router.go`

路由分工：

- `api-router.go`：后台、用户、支付、订阅、模型、渠道等业务接口
- `relay-router.go`：统一 Relay 接口
- `web-router.go`：前端静态资源与 SPA 路由兜底
- `dashboard.go`：兼容旧版 dashboard 路由

### 4.2 Middleware 层

职责：

- 认证与权限校验
- 速率限制
- CORS、Gzip、日志、恢复
- Token / 用户上下文注入
- 渠道分发与模型选择
- 安全校验

关键中间件：

- `middleware/auth.go`
- `middleware/distributor.go`
- `middleware/model-rate-limit.go`
- `middleware/rate-limit.go`
- `middleware/secure_verification.go`
- `middleware/turnstile-check.go`
- `middleware/stats.go`

这个层的特点是“业务很重”。尤其是 `Distribute()`，它不仅仅做上下文注入，还会：

- 解析模型请求
- 校验 Token 模型权限
- 处理 auto group
- 根据亲和性缓存或规则选择渠道
- 在请求上下文中写入所选渠道信息

因此这里是 Relay 体系的关键分流点。

### 4.3 Controller 层

职责：

- 接收 HTTP 请求
- 做参数绑定、返回格式控制
- 调用 Service / Relay 子系统
- 组织业务流程入口

主要 Controller 分成几类：

- 用户与认证：`user.go`、`oauth.go`、`passkey.go`、`twofa.go`
- 资金与订阅：`topup*.go`、`subscription*.go`、`billing.go`
- 渠道与模型：`channel.go`、`model*.go`、`ratio_sync.go`
- 日志与后台：`log.go`、`performance.go`、`option.go`
- Relay：`relay.go`、`playground.go`、`midjourney.go`
- 部署：`deployment.go`

其中 `controller/relay.go` 是 Relay 主控制器入口。

### 4.4 Service 层

职责：

- 承载业务规则
- 组织模型选择、计费、转换、通知、任务轮询等逻辑
- 降低 Controller 复杂度
- 为 Model 层和 Relay 层提供业务封装

典型 Service 类型：

- 渠道选择：`channel_select.go`
- 计费：`billing.go`、`task_billing.go`
- 协议转换：`convert.go`
- 敏感词：`sensitive.go`
- Midjourney / 任务：`midjourney.go`、`task.go`
- Codex 相关：`codex_oauth.go`、`codex_credential_refresh.go`
- 通知与额度预警：`user_notify.go`

### 4.5 Model 层

职责：

- 数据模型定义
- GORM 查询与更新
- 数据库迁移
- 缓存同步
- 用户、渠道、Token、日志、订阅等实体的持久化

主要模型：

- `User`
- `Token`
- `Channel`
- `Log`
- `Subscription`
- `TopUp`
- `ModelMeta`
- `PasskeyCredential`
- `Redemption`
- `Task`

`model/main.go` 还承担：

- 选择数据库类型
- 初始化主库和日志库
- 执行迁移
- 创建 Root 账号
- 检查系统初始化状态

## 5. Relay 子系统结构

Relay 是这个项目最有特色的部分。它不是简单的 HTTP 反代，而是一个具备协议解析、定价、预扣费、重试、渠道切换和多上游适配能力的专门子系统。

### 5.1 Relay 架构图

```text
Client Request
  -> router/relay-router.go
  -> TokenAuth / ModelRateLimit / Distribute
  -> controller/relay.go
  -> relay/helper.GetAndValidateRequest
  -> relay/common.GenRelayInfo
  -> service.EstimateRequestToken
  -> helper.ModelPriceHelper
  -> service.PreConsumeBilling
  -> getChannel / retry loop
  -> relay/* handler
  -> relay/channel/* adaptor
  -> Upstream Provider
  -> response / refund / violation fee / logs
```

### 5.2 Relay 关键模块

#### `controller/relay.go`

负责：

- 统一接收 Relay 请求
- 选择处理分支（文本、图片、音频、Rerank、Responses、Gemini、Realtime）
- 执行请求校验
- 生成 `RelayInfo`
- 敏感词检查
- Token 估算与价格计算
- 预扣费
- 渠道重试
- 错误格式转换
- 失败退款与违规费处理

#### `relay/helper/*`

负责：

- 不同协议请求的解析与校验
- 模型价格计算辅助
- 请求体格式兼容处理

#### `relay/common/*`

负责：

- 统一 RelayInfo 抽象
- 请求转换
- 覆盖参数
- 流式状态处理
- 通用账单辅助

#### `relay/*_handler.go`

负责：

- 按任务类型执行具体转发逻辑
- 文本、图像、音频、Embedding、Responses、Gemini、Claude 各自处理

#### `relay/channel/*`

负责：

- 上游厂商适配器
- 特定平台字段映射、请求构造和结果转换

这是上游供应商差异收敛的最底层。

## 6. 两条关键请求链路

### 6.1 普通后台业务请求链路

以“用户在控制台查看个人信息或管理员管理渠道”为例：

```text
Browser
  -> /api/...
  -> Router
  -> Auth Middleware
  -> Controller
  -> Service
  -> Model
  -> DB
  -> JSON Response
```

特点：

- 更偏传统后台业务接口
- 主要走 Session / AccessToken 鉴权
- 强依赖数据库读写

### 6.2 AI Relay 请求链路

以“客户端调用 `/v1/chat/completions`”为例：

```text
API Client
  -> /v1/chat/completions
  -> TokenAuth
  -> ModelRequestRateLimit
  -> Distribute
  -> controller.Relay
  -> validate request
  -> estimate tokens / compute billing
  -> pre-consume quota
  -> select channel
  -> relay handler
  -> upstream provider
  -> stream or final response
  -> settle billing / refund if needed
```

特点：

- 业务链路更长
- 中间涉及鉴权、模型权限、计费、重试、渠道切换
- 不只是“转发”，而是“统一编排”

## 7. 认证与权限模型

系统存在多种认证方式。

### 7.1 用户态认证

主要用于控制台与后台接口：

- Session
- Access Token

角色分层：

- 普通用户
- 管理员
- Root 管理员

### 7.2 API Token 认证

主要用于 Relay 接口：

- `Authorization` 中的 API Token
- 支持只读 Token 查询路径
- 支持 WebSocket/Reatime 特殊提取方式

### 7.3 增强安全能力

系统还提供：

- 2FA
- Passkey
- Turnstile
- 安全确认流程
- Email 域名限制
- SSRF 防护

说明这个系统的安全设计不是附带功能，而是正式产品能力的一部分。

## 8. 配置架构

项目配置来源分成两类。

### 8.1 环境变量

主要用于：

- 启动参数
- 数据库
- Redis
- Session
- 第三方服务地址
- 是否开启调试、pprof 等基础环境能力

### 8.2 数据库 Option 配置

主要用于：

- 平台开关
- 运营设置
- 计费策略
- 模型/分组配置
- 安全与展示配置

`setting/config/config.go` 提供了统一配置管理器：

- 配置模块注册
- 从 DB 加载配置
- 保存回 DB
- 结构体与字符串配置项互转

因此本项目的配置模型是：

`Env for infrastructure + DB options for runtime business configuration`

## 9. 数据存储与缓存

### 9.1 主数据库

当前兼容：

- SQLite
- MySQL
- PostgreSQL

数据库兼容要求很高，因此 Model 层大量逻辑都围绕 GORM 抽象和多数据库兼容来设计。

### 9.2 日志数据库

系统支持日志数据库与主数据库分离。

### 9.3 缓存层

缓存形态包括：

- Redis
- 内存缓存
- 磁盘缓存
- 渠道缓存
- Token 缓存
- 亲和性缓存

这些缓存主要用于：

- 提升渠道选择效率
- 减少高频查询压力
- 支持监控和统计

## 10. 后台任务与异步能力

项目启动后会拉起一组后台任务，而不是只有同步请求处理。

当前可见的后台任务包括：

- 渠道缓存同步
- 配置热同步
- 数据看板数据更新
- 自动更新渠道
- 自动测试渠道
- Codex 凭证自动刷新
- 订阅额度重置
- 渠道上游模型更新检查
- Midjourney / Task 批量更新

这说明系统的一部分业务能力依赖“调度 + 同步”机制，而不是全部发生在用户请求时。

## 11. 前端架构

前端是一个 React 单页应用，主要由以下部分组成：

- `pages/`：页面级入口
- `components/`：业务组件与 UI 组件
- `hooks/`：数据获取与行为复用
- `context/`：全局状态
- `helpers/`：前端工具函数
- `services/`：前端服务封装
- `constants/`：页面与业务常量

前端通过 `App.jsx` 定义页面路由，并通过侧边栏配置拼出后台信息架构。

它本质上是平台控制台，而不是简单的配置页面集合。

## 12. 模块关系视图

下面是更贴近目录结构的关系图。

```text
main.go
  -> common/
  -> logger/
  -> model/
  -> service/
  -> router/
  -> middleware/
  -> relay/
  -> oauth/
  -> setting/

router/
  -> controller/

controller/
  -> service/
  -> model/
  -> relay/    (Relay 场景)

service/
  -> model/
  -> setting/
  -> common/

relay/
  -> relay/helper
  -> relay/common
  -> relay/channel/*
  -> service/
  -> dto/
  -> types/

model/
  -> common/
  -> gorm/db

web/
  -> pages/
  -> components/
  -> hooks/
  -> helpers/
  -> context/
```

## 13. 如何判断代码该放哪一层

这是后续开发最常见的问题。

### 放在 Router 层的内容

- 路由路径定义
- 中间件挂载
- 接口分组

### 放在 Middleware 层的内容

- 与请求上下文强相关
- 与认证、限流、分发、安全校验相关
- 需要在 Controller 之前完成

### 放在 Controller 层的内容

- 参数绑定
- API 入参与出参格式
- 触发业务流程

### 放在 Service 层的内容

- 业务规则
- 跨多个 Model/外部系统的编排
- 计费、选择、转换、通知等逻辑

### 放在 Model 层的内容

- 数据表结构
- 数据库查询、更新、迁移
- 与持久化直接相关的逻辑

### 放在 Relay 层的内容

- 上游协议解析
- 上游请求构造
- 上游响应转换
- 流式结果处理

## 14. 当前架构的优点与代价

### 优点

- 部署简单，单体服务易落地
- 模块边界清晰，适合中等复杂度平台
- Relay 与后台能力放在同一系统中，联动方便
- 统一认证、计费、日志与配置，减少跨服务协调成本

### 代价

- 单体系统增长后，模块耦合风险会上升
- Relay 链路与后台链路共存，理解成本较高
- Middleware 与 Service 中存在较重业务逻辑，新人上手需要时间
- 后台任务、缓存、数据库兼容性要求提升了维护复杂度

## 15. 结论

这个项目当前最准确的架构定义是：

**一个以 Gin 为宿主、以 Layered Architecture 为骨架、以 Relay 子系统为核心特色、同时内嵌 React 平台控制台的模块化单体系统。**

如果后续继续演进，最值得重点关注的架构主题通常会是：

- Relay 子系统与普通后台业务的边界进一步清晰化
- 配置系统与业务设置的统一抽象
- 渠道、模型、计费三者的领域边界收敛
- 后台任务的可观测性与调度治理
