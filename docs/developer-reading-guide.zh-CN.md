# 开发者阅读顺序与源码导读

本文档面向第一次进入仓库的开发者，目标不是解释所有细节，而是回答三个最实际的问题：

- 第一次读这个项目，先看什么
- 不同需求应该从哪条链路切入
- 哪些文件是“核心文件”，哪些文件是“专项扩展”

配套文档：

- `docs/product-feature-list.zh-CN.md`
- `docs/project-feature-overview.zh-CN.md`
- `docs/system-architecture.zh-CN.md`

## 1. 推荐阅读策略

不要一上来就平铺整个目录。

这个项目体量已经不小，而且同时包含：

- 后台业务接口
- Relay 专项链路
- 前端控制台
- 配置系统
- 后台任务

如果从细节开始读，很容易迷失。更有效的方式是：

1. 先理解服务怎么启动
2. 再理解路由怎么分流
3. 再理解两条主链路
4. 最后再按你要改的业务方向深入

## 2. 第一次进入仓库，建议先看这 10 个文件

如果你时间有限，先看下面这些：

1. `main.go`
2. `router/main.go`
3. `router/api-router.go`
4. `router/relay-router.go`
5. `middleware/auth.go`
6. `middleware/distributor.go`
7. `controller/relay.go`
8. `model/main.go`
9. `web/src/App.jsx`
10. `web/src/components/layout/SiderBar.jsx`

这 10 个文件基本能回答：

- 服务怎么起来
- 后端入口有哪些
- Relay 请求怎么走
- 用户/管理员页面有哪些
- 项目的主要信息架构是什么

## 3. 第一轮阅读顺序

### 第一步：先看服务入口

看：

- `main.go`

要回答的问题：

- 服务启动时做了哪些初始化
- 有哪些后台任务会自动启动
- Gin 中间件是在哪一层全局挂载的
- 前端静态资源是如何被嵌入和托管的

读完后你应该知道：

- 这是一个单体 Go 服务
- 它不是纯 API 服务，还会托管前端
- 它启动后会跑缓存同步、渠道检测、订阅重置等后台任务

### 第二步：看总路由入口

看：

- `router/main.go`

要回答的问题：

- 系统有几套路由
- Web 与 API 的边界在哪里
- 前端是本地托管还是外部重定向

读完后你应该形成一个最粗粒度的地图：

- `/api/*`：后台业务接口
- `/v1/*`、`/v1beta/*`、`/mj/*`、`/suno/*`：Relay 接口
- `/`：前端 SPA

### 第三步：分开读两条主路由

先看：

- `router/api-router.go`

再看：

- `router/relay-router.go`

这样做的目的，是把“后台业务系统”和“AI Relay 系统”分开理解。

#### 读 `api-router.go` 时重点看

- 用户相关接口有哪些
- 支付、充值、订阅怎么分组
- 渠道、模型、部署、日志、配置分别在哪些路由组

#### 读 `relay-router.go` 时重点看

- 支持哪些协议和任务类型
- 哪些路径对应 OpenAI、Claude、Gemini、Midjourney、Suno
- 哪些中间件只在 Relay 路由上生效

## 4. 第二轮阅读：理解两条核心链路

### 4.1 后台业务链路

推荐阅读顺序：

1. `router/api-router.go`
2. `middleware/auth.go`
3. 某个具体 controller，例如 `controller/user.go`
4. 对应 service，例如 `service/*`
5. 对应 model，例如 `model/user.go`

用“登录”做例子时可以读：

- `router/api-router.go`
- `controller/user.go`
- `model/user.go`
- `model/twofa.go`
- `model/passkey.go`

这条线适合帮助你理解：

- 控制台接口是怎么组织的
- session / access token 如何工作
- 用户实体与认证逻辑在哪里

### 4.2 Relay 链路

推荐阅读顺序：

1. `router/relay-router.go`
2. `middleware/auth.go`
3. `middleware/model-rate-limit.go`
4. `middleware/distributor.go`
5. `controller/relay.go`
6. `relay/helper/valid_request.go`
7. `relay/common/*`
8. 某个具体 handler，例如 `relay/claude_handler.go`
9. 某个具体 adaptor，例如 `relay/channel/openai/*`

这条线适合帮助你理解：

- 请求如何被识别为哪种协议
- 模型名如何被解析
- 渠道如何被选择
- 预扣费、重试和退款如何工作
- 上游厂商适配点在哪里

## 5. 按需求类型选择阅读入口

这是最实用的一部分。

### 5.1 如果你要改用户登录、注册、2FA、Passkey

先看：

- `router/api-router.go`
- `controller/user.go`
- `controller/twofa.go`
- `controller/passkey.go`
- `middleware/auth.go`
- `model/user.go`
- `model/twofa.go`
- `model/passkey.go`
- `web/src/components/auth/*`
- `web/src/components/settings/PersonalSetting.jsx`

### 5.2 如果你要改渠道管理

先看：

- `router/api-router.go`
- `controller/channel.go`
- `controller/channel-test.go`
- `controller/channel-billing.go`
- `controller/channel_upstream_update.go`
- `model/channel.go`
- `model/ability.go`
- `service/channel.go`
- `service/channel_select.go`
- `web/src/pages/Channel/index.jsx`
- `web/src/components/table/channels/*`

### 5.3 如果你要改 Relay 行为

先看：

- `router/relay-router.go`
- `middleware/distributor.go`
- `controller/relay.go`
- `relay/helper/valid_request.go`
- `relay/common/relay_info.go`
- `relay/common/request_conversion.go`
- `relay/*_handler.go`
- `relay/channel/*`
- `service/billing.go`
- `service/quota.go`

### 5.4 如果你要接入一个新模型厂商

先看：

- `web/src/constants/channel.constants.js`
- `constant/channel.go`
- `relay/channel/*` 中已有相似厂商
- `relay/helper/model_mapped.go`
- `controller/channel.go`
- `service/channel_select.go`
- `router/relay-router.go`

你需要重点搞清楚 4 件事：

- 新厂商属于哪种协议族
- 模型列表如何获取
- 请求和响应需要做哪些字段映射
- 计费和能力标签怎么挂到现有体系里

### 5.5 如果你要改充值、支付、订阅

先看：

- `router/api-router.go`
- `controller/topup.go`
- `controller/topup_stripe.go`
- `controller/topup_creem.go`
- `controller/topup_waffo.go`
- `controller/subscription.go`
- `controller/subscription_payment_*.go`
- `service/billing.go`
- `service/funding_source.go`
- `model/topup.go`
- `model/subscription.go`
- `web/src/components/topup/*`

### 5.6 如果你要改模型定价、分组、可见模型

先看：

- `controller/pricing.go`
- `controller/ratio_config.go`
- `controller/ratio_sync.go`
- `model/pricing.go`
- `model/model_meta.go`
- `setting/ratio_setting/*`
- `setting/model_setting/*`
- `web/src/hooks/model-pricing/*`
- `web/src/components/settings/RatioSetting.jsx`

### 5.7 如果你要改系统设置或全局开关

先看：

- `controller/option.go`
- `model/option.go`
- `model/main.go`
- `setting/config/config.go`
- `setting/operation_setting/*`
- `setting/system_setting/*`
- `setting/performance_setting/*`
- `web/src/pages/Setting/index.jsx`
- `web/src/components/settings/*`

### 5.8 如果你要改模型部署

先看：

- `controller/deployment.go`
- `service/http_client.go`
- `docs/ionet-client.md`
- `web/src/pages/ModelDeployment/index.jsx`
- `web/src/components/table/model-deployments/*`
- `web/src/hooks/model-deployments/*`

## 6. 哪些文件最关键

下面这些文件是项目中的“枢纽文件”。

### 6.1 后端枢纽文件

- `main.go`
- `router/api-router.go`
- `router/relay-router.go`
- `middleware/auth.go`
- `middleware/distributor.go`
- `controller/relay.go`
- `model/main.go`
- `service/channel_select.go`

### 6.2 前端枢纽文件

- `web/src/App.jsx`
- `web/src/components/layout/SiderBar.jsx`
- `web/src/pages/Setting/index.jsx`
- `web/src/components/settings/PersonalSetting.jsx`

这些文件的特点是：

- 改动频率高
- 承上启下
- 很多需求最终都会经过它们

## 7. 哪些目录可以后读

第一次读仓库时，这些目录不建议一开始就深挖：

- `relay/channel/*` 的所有厂商适配细节
- `docs/openapi/*`
- `i18n/locales/*`
- 各种单独支付网关的细节实现
- 各种专项任务和视频通道的细节实现

原因不是它们不重要，而是它们更偏“扩展实现”，不先理解主骨架的话，读这些很难形成整体认知。

## 8. 推荐的阅读地图

### 8.1 90 分钟快速上手版

如果你只有 90 分钟：

1. `main.go`
2. `router/main.go`
3. `router/api-router.go`
4. `router/relay-router.go`
5. `middleware/auth.go`
6. `middleware/distributor.go`
7. `controller/relay.go`
8. `web/src/App.jsx`
9. `web/src/components/layout/SiderBar.jsx`

这会让你先形成“系统地图”。

### 8.2 1 天深入版

如果你有半天到一天时间：

1. 完成上面的 90 分钟版
2. 看 `model/main.go`
3. 看一个用户链路：`controller/user.go` + `model/user.go`
4. 看一个渠道链路：`controller/channel.go` + `model/channel.go`
5. 看一个 Relay 链路：`controller/relay.go` + `relay/helper/*` + `relay/common/*`
6. 看一个前端设置链路：`web/src/pages/Setting/index.jsx`

这时你基本就能开始改需求了。

### 8.3 按业务模块上手版

如果你已经知道自己要改什么，就直接按“第 5 节”的场景入口跳读，不需要按目录顺序平铺。

## 9. 阅读时常见误区

### 误区 1：一开始就读 `relay/channel/*`

问题：

- 你会看到大量厂商适配细节，但不知道它们在系统里的位置

更好的顺序：

- 先读 `controller/relay.go`
- 再读 `middleware/distributor.go`
- 最后再读某个具体 adaptor

### 误区 2：把 `router/api-router.go` 当作实现文件

问题：

- 路由文件只是入口地图，不是业务实现

正确用法：

- 用它确认“这个功能在哪个 controller”

### 误区 3：分不清后台业务和 Relay 业务

问题：

- 这两条链路共用一些层，但目标不同

建议：

- 看代码时始终问自己：这是“控制台/后台业务”还是“模型调用链路”

### 误区 4：先从前端页面看完整业务

问题：

- 前端展示的是结果，不一定能解释后端真实边界

更好的方式：

- 先从路由和 controller 定位
- 再回到前端看交互和参数

## 10. 新功能落点建议

如果你要加新功能，可以先用下面的判断法：

- 新 HTTP 路由：放 `router/*`
- 认证、限流、上下文注入：放 `middleware/*`
- 接口入参与返回编排：放 `controller/*`
- 业务逻辑与流程编排：放 `service/*`
- 数据读写与迁移：放 `model/*`
- 上游协议适配：放 `relay/*`
- 平台配置项：放 `setting/*` 和 `option`
- 前端页面：放 `web/src/pages/*`
- 前端复用组件：放 `web/src/components/*`

## 11. 结论

阅读这个项目最重要的不是“把所有文件读完”，而是尽快建立下面这个判断框架：

- 这是后台业务还是 Relay 业务
- 它属于哪一层
- 真实入口在哪个 router / controller
- 核心规则藏在 service 还是 middleware
- 持久化落在哪个 model

一旦这个框架建立起来，后面的阅读效率会高很多。

