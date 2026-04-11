# 项目功能分析与总览

本文档用于梳理当前仓库在代码层面已经具备的功能能力，结论基于当前项目代码结构、后端路由、前端页面入口、系统设置页和 README 描述综合整理而成。

配套文档：

- `docs/product-feature-list.zh-CN.md`：面向产品、运营和业务方的功能清单
- `docs/system-architecture.zh-CN.md`：面向研发的系统架构与模块关系说明
- `docs/developer-reading-guide.zh-CN.md`：面向研发的阅读顺序与源码导读
- `docs/core-data-models.zh-CN.md`：面向研发的核心数据模型关系梳理
- `docs/core-business-sequences.zh-CN.md`：面向研发的核心业务链路时序图

## 1. 项目定位

从代码实现看，这个项目并不只是一个简单的 LLM API 转发层，而是一个完整的 AI 网关平台，至少包含以下几层能力：

- 统一 AI Relay 网关
- 多上游渠道与模型聚合
- 用户、认证与安全体系
- 额度、钱包、充值、订阅计费体系
- 管理后台与用户控制台
- 运维、监控、日志、风控与运营配置
- 模型部署管理能力

可以将其理解为：`AI API Gateway + Billing/Admin Console + User Workspace + Operations Platform`。

## 2. 分析依据

本次分析主要依据以下入口文件：

- `README.md`
- `router/relay-router.go`
- `router/api-router.go`
- `web/src/App.jsx`
- `web/src/components/layout/SiderBar.jsx`
- `web/src/pages/Setting/index.jsx`
- `web/src/components/settings/OperationSetting.jsx`
- `web/src/components/settings/SystemSetting.jsx`
- `web/src/components/settings/PersonalSetting.jsx`
- `web/src/constants/channel.constants.js`

分析方法如下：

1. 先从 README 提取项目对外声明的能力边界。
2. 再从后端路由确认哪些能力已经形成真实 API。
3. 再从前端路由与侧边栏确认哪些功能已经暴露为可操作页面。
4. 最后用系统设置页补足运营、风控、支付和安全类能力。

## 3. 功能模块总览

### 3.1 统一 AI Relay 网关

项目提供统一的 Relay 入口，支持多种上游协议和任务类型：

- OpenAI Chat Completions
- OpenAI Responses
- OpenAI Realtime
- Claude Messages
- Gemini 接口
- Embeddings
- Audio Transcriptions / Translations / Speech
- Images Generations / Edits
- Rerank
- Moderations

除此之外，还单独扩展了异步或专项任务通道：

- Midjourney 任务流
- Suno 音乐任务流

对应入口：

- `router/relay-router.go`

### 3.2 多上游渠道与模型聚合

项目核心价值之一是把大量上游供应商抽象成统一渠道。前端常量中已注册大量渠道类型，说明平台具备多供应商接入和统一管理能力。

当前代码中能明确看到的渠道类型包括但不限于：

- OpenAI
- Azure OpenAI
- Anthropic Claude
- AWS Claude
- Google Gemini
- OpenRouter
- Ollama
- Cohere
- DeepSeek
- Mistral AI
- xAI
- Moonshot
- Perplexity
- Vertex AI
- Dify
- Jina
- SiliconCloud
- Replicate
- Codex (OpenAI OAuth)
- Midjourney Proxy / Plus
- Suno API
- 多个国内模型平台与视频/绘图渠道
- 自定义渠道

这意味着项目的“渠道管理”不是简单存个 key，而是支持多种供应商协议、渠道类型和能力差异的聚合层。

对应入口：

- `web/src/constants/channel.constants.js`
- `router/api-router.go` 中 `/api/channel/*`

### 3.3 协议转换与智能路由

根据 README 和 Relay 相关代码，项目具备如下中间层能力：

- OpenAI Compatible 与 Claude Messages 之间转换
- OpenAI Compatible 到 Gemini 转换
- Gemini 到 OpenAI Compatible 转换
- Responses 相关兼容转换
- thinking 内容转换为普通内容
- 按模型/用户维度做路由分发
- 失败自动重试
- 用户级模型请求限流
- 渠道加权、分发与状态控制

这部分能力说明项目并不是“直连透传”，而是对上游协议做了格式适配与请求调度。

### 3.4 用户体系与认证能力

项目具备完整的用户账户体系，包括：

- 注册
- 登录
- 登出
- 密码重置
- 邮箱验证码
- 用户资料查询与更新
- 用户自助删除账号
- 获取用户可用模型与分组

多因子与现代认证能力包括：

- 2FA 状态查询、启用、禁用、备用码重置
- Passkey 登录
- Passkey 注册、校验、删除

OAuth 与第三方认证包括：

- GitHub
- Discord
- OIDC
- LinuxDO
- 微信
- Telegram
- 自定义 OAuth Provider

对应入口：

- `router/api-router.go` 中 `/api/user/*`
- `router/api-router.go` 中 `/api/oauth/*`
- `router/api-router.go` 中 `/api/custom-oauth-provider/*`
- `web/src/components/settings/SystemSetting.jsx`
- `web/src/components/settings/PersonalSetting.jsx`

### 3.5 钱包、额度、充值与订阅

项目内置了完整的额度与资金体系，而不是只做 API 认证。

#### 充值与钱包

- 在线充值
- 兑换码充值
- 充值金额估算
- 充值记录查询
- 用户钱包页

支持的支付相关能力：

- EPay
- Stripe
- Creem
- Waffo

#### 订阅体系

- 订阅计划查询
- 用户当前订阅查询
- 用户订阅计费偏好设置
- 订阅支付请求
- 后台创建/更新/上下线订阅计划
- 管理员为用户绑定订阅
- 管理员管理用户订阅生命周期

#### 邀请与额度流转

- 邀请码/邀请链接
- 邀请返利
- 额度转赠

对应入口：

- `router/api-router.go` 中 `/api/user/topup/*`
- `router/api-router.go` 中 `/api/user/pay`
- `router/api-router.go` 中 `/api/subscription/*`
- `web/src/components/topup/index.jsx`

### 3.6 Token、额度与使用统计

项目提供完整的 token 化访问控制与配额统计能力：

- 用户创建、更新、删除 API Token
- 按 Token 查询详情
- 获取 Token key
- 批量获取 Token keys
- Token 使用量查询
- 用户模型列表与分组
- 额度消耗日志
- 用户侧与管理员侧日志检索
- 数据导出/统计数据查询

这部分是平台商业化和管理员审计的基础。

对应入口：

- `router/api-router.go` 中 `/api/token/*`
- `router/api-router.go` 中 `/api/usage/token`
- `router/api-router.go` 中 `/api/log/*`
- `router/api-router.go` 中 `/api/data/*`

### 3.7 渠道管理后台

管理员可对上游渠道做完整管理，能力包括：

- 渠道列表、搜索、查看详情
- 新增渠道、更新渠道、删除渠道
- 删除禁用渠道
- 渠道测试与全量测试
- 查询/刷新渠道余额
- 拉取上游模型列表
- 批量打标签
- 按标签启停渠道
- 修复渠道能力映射
- 复制渠道
- 多 key 管理
- 检测与应用渠道上游模型更新
- 获取渠道密钥

专项能力还包括：

- Codex OAuth 授权与刷新
- 查询 Codex 渠道使用情况
- Ollama 拉取/删除模型与版本查询

对应入口：

- `router/api-router.go` 中 `/api/channel/*`

### 3.8 模型管理与定价管理

项目不只管理“渠道”，还管理“平台内模型元数据”和“计费规则”。

已具备的能力包括：

- 同步上游模型
- 预览上游同步结果
- 缺失模型查询
- 模型元数据 CRUD
- 分组与模型定价设置
- 比例配置同步
- 模型广场/模型页

对应入口：

- `router/api-router.go` 中 `/api/models/*`
- `router/api-router.go` 中 `/api/ratio_sync/*`
- `web/src/pages/Setting/index.jsx` 中的“分组与模型定价设置”

### 3.9 模型部署管理

项目中已经有独立的模型部署模块，说明它在“调用别人的模型”之外，也支持“部署和托管自己的模型工作负载”。

当前可见能力包括：

- 模型部署配置读取
- 部署连接测试
- 部署列表与搜索
- 查询硬件类型、地域、可用副本
- 价格估算
- 检查集群名称是否可用
- 创建部署
- 查询部署详情
- 查询部署日志
- 查看部署容器列表与容器详情
- 更新部署
- 更新部署名称
- 延长部署
- 删除部署

对应入口：

- `router/api-router.go` 中 `/api/deployments/*`
- `web/src/pages/ModelDeployment/index.jsx`

### 3.10 日志、任务与专项记录

除普通 API 使用日志外，项目还提供专项记录页面：

- 使用日志
- 绘图日志
- 任务日志
- Midjourney 记录

这些能力在前端侧已经作为独立页面存在，说明它们不是临时调试信息，而是正式业务数据。

对应入口：

- `web/src/components/layout/SiderBar.jsx`
- `router/api-router.go` 中 `/api/log/*`
- `router/api-router.go` 中 `/api/mj/*`
- `router/api-router.go` 中 `/api/task/*`

### 3.11 数据看板与用户工作台

前端存在完整的用户工作台，主要包括：

- 首页
- 数据看板
- API 信息展示
- 公告与 FAQ
- Uptime 状态面板
- Playground
- 聊天页面
- 模型广场
- 个人设置
- 钱包页

这意味着项目定位并非纯后端服务，还提供终端用户可直接使用的 Web 控制台。

对应入口：

- `web/src/App.jsx`
- `web/src/components/layout/SiderBar.jsx`
- `web/src/components/dashboard/index.jsx`

### 3.12 运营、风控与平台配置

系统设置页暴露了大量平台级配置，说明项目具备较强的运营能力。

#### 运营设置

- 新用户额度
- 预扣费策略
- 邀请奖励
- 文档链接
- 额度单位与汇率
- 是否展示 token 统计
- 是否默认折叠侧边栏
- Demo 模式
- 自用模式
- 顶栏模块显示配置
- 管理员侧边栏模块显示配置
- 敏感词过滤
- 日志策略
- 渠道监控与自动禁用
- 签到奖励
- 用户 Token 数量限制

#### 系统设置

- 密码登录/注册开关
- 邮箱验证
- SMTP
- GitHub / Discord / OIDC / LinuxDO OAuth
- 微信与 Telegram 认证
- Turnstile
- Passkey 参数
- Worker 相关配置
- 页脚、公告、站点地址
- 邮箱域名限制
- SSRF 防护
- 自定义 OAuth Provider

#### 其他设置页

从前端设置首页可见，当前系统设置被拆成多个一级分类：

- 运营设置
- 仪表盘设置
- 聊天设置
- 绘图设置
- 支付设置
- 分组与模型定价设置
- 速率限制设置
- 模型相关设置
- 模型部署设置
- 性能设置
- 系统设置
- 其他设置

对应入口：

- `web/src/pages/Setting/index.jsx`
- `web/src/components/settings/OperationSetting.jsx`
- `web/src/components/settings/SystemSetting.jsx`

### 3.13 性能与运维

项目还有一组偏运维的 Root 级接口：

- 性能统计
- 清理磁盘缓存
- 重置性能统计
- 主动 GC
- 查询日志文件
- 清理日志文件

另有状态与监控能力：

- 平台状态查询
- Uptime 状态查询
- 频道/渠道自动检测

对应入口：

- `router/api-router.go` 中 `/api/performance/*`
- `router/api-router.go` 中 `/api/status`
- `router/api-router.go` 中 `/api/uptime/status`

## 4. 前端页面入口视图

从前端路由和侧边栏看，当前主要页面入口如下。

### 4.1 公共页面

- `/`
- `/setup`
- `/login`
- `/register`
- `/reset`
- `/pricing`
- `/about`
- `/user-agreement`
- `/privacy-policy`

### 4.2 用户控制台

- `/console`
- `/console/token`
- `/console/log`
- `/console/midjourney`
- `/console/task`
- `/console/topup`
- `/console/personal`
- `/console/playground`
- `/console/chat/:id?`

### 4.3 管理后台

- `/console/channel`
- `/console/subscription`
- `/console/models`
- `/console/deployment`
- `/console/redemption`
- `/console/user`
- `/console/setting`

## 5. 后端接口分组视图

按路由可将后端能力大致分为：

- `/v1/*` 和 `/v1beta/*`：统一 Relay 接口
- `/mj/*`、`/suno/*`：专项任务 Relay
- `/api/user/*`：用户、认证、2FA、Passkey、充值、自助设置
- `/api/subscription/*`：订阅与订阅管理
- `/api/channel/*`：渠道后台
- `/api/token/*`：Token 管理
- `/api/log/*`：日志与统计
- `/api/models/*`：模型元数据
- `/api/deployments/*`：模型部署
- `/api/option/*`：平台全局配置
- `/api/performance/*`：性能与运维

## 6. 当前边界与未实现项

项目能力已经很完整，但从 Relay 路由仍能看到一部分接口尚未实现：

- `POST /v1/images/variations`
- `GET/POST/DELETE /v1/files*`
- `POST/GET /v1/fine-tunes*`
- `DELETE /v1/models/:model`

因此当前系统更偏向：

- 统一调用代理
- 多渠道与多协议兼容
- 计费与后台管理

而不是一个已经覆盖 OpenAI 全量接口的 100% 完整替代实现。

## 7. 结论

当前项目可以概括为一个面向 AI 模型接入、分发、计费和运营的综合平台，核心特征如下：

- 既有 API 网关能力，也有业务后台能力
- 既服务管理员，也服务终端用户
- 既支持多模型代理，也开始覆盖模型部署
- 既关注调用兼容，也关注计费、风控、监控和运营

如果后续需要继续深入，建议沿以下顺序阅读代码：

1. `router/relay-router.go`
2. `router/api-router.go`
3. `controller/relay.go` 及 `relay/`
4. `controller/channel.go`、`model/channel.go`
5. `controller/user.go`、`model/user.go`
6. `controller/subscription*.go`、`controller/topup*.go`
7. `web/src/App.jsx`
8. `web/src/pages/Setting/index.jsx`
