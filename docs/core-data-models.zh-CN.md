# 核心数据模型关系梳理

本文档面向研发，梳理当前项目中最关键的数据实体、实体之间的关系，以及这些关系在业务链路中的作用。

这份文档重点回答：

- 系统里最核心的表有哪些
- 用户、Token、渠道、模型、订阅、日志之间如何关联
- 哪些关系是直接外键关系，哪些是间接映射关系
- 为什么有些字段看起来像“冗余”

配套文档：

- `docs/project-feature-overview.zh-CN.md`
- `docs/system-architecture.zh-CN.md`
- `docs/developer-reading-guide.zh-CN.md`

## 1. 先给结论

这个项目的数据模型可以分成 6 个领域：

1. 用户与身份
2. 访问令牌与权限
3. 渠道与模型能力
4. 支付、充值与订阅
5. 日志与统计
6. 平台配置与扩展元数据

其中最核心的主干关系是：

```text
User
  -> Token
  -> TopUp
  -> SubscriptionOrder
  -> UserSubscription
  -> Log
  -> PasskeyCredential
  -> UserOAuthBinding

Channel
  -> Ability
  <- Model Pricing / Enabled Models (indirectly)

ModelMeta
  -> Pricing View (derived)
  <- Ability (indirect match / aggregation)
```

这里有一个很重要的设计点：

**模型与渠道不是简单的直接多对多字符串关系，而是通过 `Ability` 表把 `group + model + channel` 三元关系显式建模。**

这决定了项目在“按分组开放模型”“多渠道多优先级选择”“定价视图聚合”上都比较灵活。

## 2. 核心实体总图

```text
                        +------------------+
                        |       User       |
                        +------------------+
                         |   |    |    |   \
                         |   |    |    |    \
                         v   v    v    v     v
                    +------+ +--------+ +------------------+
                    |Token | |TopUp   | | SubscriptionOrder|
                    +------+ +--------+ +------------------+
                       |                    |
                       |                    v
                       |             +------------------+
                       |             | UserSubscription |
                       |             +------------------+
                       |
                       v
                    +------+
                    | Log  |
                    +------+

User
  -> PasskeyCredential
  -> UserOAuthBinding

Channel
  -> Ability <- group + model + channel capability mapping

ModelMeta
  -> joined/aggregated with Ability + Vendor + Ratio configs
  -> Pricing (derived, not direct table entity)
```

## 3. 用户与身份领域

### 3.1 `User`

`User` 是最核心的业务主体。

它承载了 4 类信息：

- 基础身份信息：用户名、显示名、邮箱
- 权限信息：角色、状态、分组
- 资金与额度信息：Quota、UsedQuota、AffQuota
- 认证绑定信息：GitHubId、DiscordId、OidcId、WeChatId、TelegramId、LinuxDOId、AccessToken

额外特点：

- `Group` 是平台侧分组概念，直接影响可用渠道和可用模型
- `Setting` 是 JSON 文本，承载用户侧个性化设置和通知配置
- `InviterId`、`AffCode`、`AffQuota` 说明用户模型里同时承载了邀请体系
- `StripeCustomer` 把用户与外部支付客户体系关联起来

可以理解为：

`User = 账户主体 + 平台配额账户 + 身份绑定载体`

### 3.2 `PasskeyCredential`

这是用户与 WebAuthn/Passkey 的一对一绑定。

关键关系：

- `PasskeyCredential.UserID -> User.Id`

设计特征：

- 每个用户当前只保留一条 Passkey 凭据记录
- 存储的是 WebAuthn 所需的 CredentialID、公钥、SignCount、备份状态等

业务意义：

- 用于 Passkey 登录和安全验证

### 3.3 `UserOAuthBinding`

这是“自定义 OAuth Provider”与用户之间的绑定表。

关键关系：

- `UserOAuthBinding.UserId -> User.Id`
- `UserOAuthBinding.ProviderId -> CustomOAuthProvider.Id`

它和 `User` 表中的 GitHubId、DiscordId、OidcId 不同：

- `User` 表中的字段用于平台内置 OAuth
- `UserOAuthBinding` 用于动态配置的自定义 OAuth Provider

这说明系统对 OAuth 有两套建模方式：

- 固定字段型内置 Provider
- 扩展表型自定义 Provider

## 4. Token 与访问控制领域

### 4.1 `Token`

`Token` 是用户对外调用 API 的主要凭证。

关键关系：

- `Token.UserId -> User.Id`

它承载的不只是“Key”，还包括完整的访问控制策略：

- 过期时间
- 剩余额度
- 是否无限额度
- 模型限制
- IP 白名单
- 分组
- 跨分组重试

这里最重要的设计点有两个：

#### 1. Token 自己带 `Group`

说明系统不是只看用户所在分组，也允许把 Token 绑定到某个分组策略上。

这会直接影响：

- 能访问哪些模型
- 渠道选择落在哪个 group
- auto group 的重试逻辑

#### 2. Token 有 `ModelLimits`

说明模型访问控制不只发生在 User 层，还可以发生在 Token 层。

这让平台可以实现：

- 同一用户下发多个不同权限的 Token
- 把某些 Token 限定为只能访问部分模型

### 4.2 `Log` 与 `Token`

`Log` 中有 `TokenId` 和 `TokenName`。

这意味着日志在查询时同时保留：

- 可关联的 Token 主键
- 便于展示的 Token 名称快照

因此即便后续 Token 被重命名、删除，日志展示仍然有上下文。

## 5. 渠道、模型与能力映射领域

这是项目最容易误读的一块。

### 5.1 `Channel`

`Channel` 代表一个上游供应商接入配置。

它不是单纯的 API Key 存储，而是一个完整的接入单元，包含：

- 供应商类型
- Key
- BaseURL
- 状态
- 权重
- 分组
- 模型列表
- 优先级
- 余额
- 标签
- 额外参数与头覆盖
- 多 Key 模式信息

其中有几个字段特别关键：

- `Type`：渠道类型，决定走哪个适配器
- `Group`：该渠道服务于哪些业务分组
- `Models`：该渠道声明支持哪些模型
- `Priority` + `Weight`：决定选择顺序与负载分摊
- `ChannelInfo`：多 Key 模式的运行时状态

### 5.2 `Ability`

`Ability` 是渠道体系中最关键的中间表。

字段组合：

- `Group`
- `Model`
- `ChannelId`

这实际上表达的是：

**某个渠道在某个业务分组下，是否为某个模型提供能力。**

也就是一个三元关系：

```text
(group, model, channel) -> enabled / priority / weight / tag
```

这是理解整个路由与定价体系的关键。

#### 为什么不直接靠 `Channel.Models` 和 `Channel.Group`

因为系统需要支持：

- 同一个渠道服务多个 group
- 同一个渠道支持多个 model
- 每个 `(group, model)` 组合都可能有不同优先级和权重
- 渠道启停后需要同步更新能力映射

所以 `Channel.Models` 和 `Channel.Group` 更像输入配置，
而 `Ability` 才是运行时真正参与选择和聚合的关系表。

### 5.3 `Channel` 与 `Ability` 的关系

关系可以理解为：

- 一个 `Channel` 会展开成多条 `Ability`
- 一个 `Ability` 只指向一个 `Channel`

即：

`Channel 1 -> N Ability`

生成方式：

- `Channel.Group` 拆成多个 group
- `Channel.Models` 拆成多个 model
- 做笛卡尔积，生成多条 `(group, model, channel)` 记录

### 5.4 `Model`（模型元数据）

`model/model_meta.go` 中的实体名叫 `Model`，但语义上更接近“平台模型元数据”，这里称它为 `ModelMeta` 更容易理解。

它承载的是：

- 模型名
- 描述
- 图标
- 标签
- 供应商归属
- 自定义 endpoint 信息
- 展示状态
- 是否参与官方同步
- 名称匹配规则

它不是“渠道是否支持模型”的直接来源，而是：

- 平台对某个模型的官方描述
- 模型广场与定价视图的元信息来源

### 5.5 `ModelMeta` 与 `Ability` 的关系

这两个实体不是直接外键关系，而是通过模型名关联和聚合。

关系逻辑是：

- `Ability` 决定某模型在某 group 下是否真的可用
- `ModelMeta` 决定这个模型在平台上如何被描述、展示和归类

也就是说：

- `Ability` 偏运行时能力
- `ModelMeta` 偏展示和运营元数据

### 5.6 `Pricing`

`Pricing` 不是一个数据库表，而是一个聚合后的派生视图结构。

它来自多种来源组合：

- `Ability`
- `ModelMeta`
- Vendor 信息
- ratio_setting 配置
- endpoint 能力映射

因此它更像：

`Pricing View = 可用模型集合 + 元数据 + 分组可见性 + 定价配置`

这一点很重要，因为很多人看到 `Pricing` 容易误以为它是独立表。

## 6. 支付、充值与订阅领域

### 6.1 `TopUp`

`TopUp` 是一次充值订单。

关键关系：

- `TopUp.UserId -> User.Id`

它承载：

- 充值额度
- 支付金额
- 第三方交易号
- 支付方式
- 创建时间、完成时间
- 状态

业务上它的结果是：

- 支付成功后给 `User.Quota` 加值

### 6.2 `SubscriptionPlan`

订阅计划定义了一个可售卖的订阅产品。

它承载：

- 标题、价格、货币
- 时长与单位
- 是否启用
- 支付平台产品 ID
- 最大购买次数
- 升级用户分组
- 总额度
- 重置周期

它更像“商品定义”。

### 6.3 `SubscriptionOrder`

这是订阅支付订单。

关键关系：

- `SubscriptionOrder.UserId -> User.Id`
- `SubscriptionOrder.PlanId -> SubscriptionPlan.Id`

它表示：

- 用户发起了一次购买某个订阅计划的支付动作

### 6.4 `UserSubscription`

这是实际生效在用户身上的订阅实例。

关键关系：

- `UserSubscription.UserId -> User.Id`
- `UserSubscription.PlanId -> SubscriptionPlan.Id`

和 `SubscriptionOrder` 的区别：

- `SubscriptionOrder` 是交易订单
- `UserSubscription` 是生效订阅

它承载：

- 总额度与已用额度
- 开始/结束时间
- 生效状态
- 来源
- 重置周期时间点
- 升级后的用户分组
- 升级前分组

这说明订阅体系不是简单“付款后加余额”，而是单独建模了：

- 计划定义
- 交易订单
- 用户实例

### 6.5 订阅相关主链路

可以简化理解为：

```text
User
  -> buy SubscriptionPlan
  -> create SubscriptionOrder
  -> payment callback success
  -> create UserSubscription
  -> optionally upgrade User.Group
```

## 7. 日志与统计领域

### 7.1 `Log`

`Log` 是平台里的统一日志实体，但它不是纯系统日志，而是业务日志。

它能记录：

- 消费日志
- 充值日志
- 管理日志
- 系统日志
- 错误日志
- 退款日志

关键关系：

- `Log.UserId -> User.Id`
- `Log.TokenId -> Token.Id`
- `Log.ChannelId -> Channel.Id`

它同时冗余保存：

- `Username`
- `TokenName`
- `ModelName`
- `Group`
- `Ip`
- `RequestId`

这样做的原因是：

- 日志查询更方便
- 减少多表联查
- 保留当时调用上下文的快照

所以 `Log` 不是一个“纯规范化”模型，而是偏审计和查询友好的宽表。

### 7.2 `UseData` / 配额统计

虽然这里没有展开所有统计模型，但从日志和看板逻辑可以看出：

- 消费日志是原始事实
- 看板数据是聚合结果

因此系统的数据层是：

- 原始事件：`Log`
- 聚合视图：统计数据表/缓存

## 8. 配置与平台元数据领域

### 8.1 `Option`

`Option` 是平台全局配置的持久化入口。

语义上它是：

- 配置项键值表

但通过 `setting/config/config.go`，它又被映射成多个强类型配置模块。

因此：

- DB 中是扁平 `key -> value`
- 运行时是模块化配置结构

### 8.2 `CustomOAuthProvider`

这是平台的扩展身份源定义。

配合 `UserOAuthBinding` 使用，用来支持自定义 OAuth 登录与绑定。

### 8.3 `Vendor`

`Vendor` 用于承接模型供应商元数据。

`ModelMeta.VendorID -> Vendor.Id`

它主要影响：

- 模型展示
- 定价聚合
- 供应商筛选

## 9. 关键关系总表

下面用列表把最重要的关系压缩总结一遍。

### 9.1 直接主外键关系

- `Token.UserId -> User.Id`
- `TopUp.UserId -> User.Id`
- `SubscriptionOrder.UserId -> User.Id`
- `SubscriptionOrder.PlanId -> SubscriptionPlan.Id`
- `UserSubscription.UserId -> User.Id`
- `UserSubscription.PlanId -> SubscriptionPlan.Id`
- `PasskeyCredential.UserID -> User.Id`
- `UserOAuthBinding.UserId -> User.Id`
- `UserOAuthBinding.ProviderId -> CustomOAuthProvider.Id`
- `Log.UserId -> User.Id`
- `Log.TokenId -> Token.Id`
- `Log.ChannelId -> Channel.Id`
- `Ability.ChannelId -> Channel.Id`
- `ModelMeta.VendorID -> Vendor.Id`

### 9.2 间接或逻辑关系

- `Ability.Model <-> ModelMeta.ModelName`
- `Ability.Group <-> User.Group / Token.Group`
- `Pricing <- Ability + ModelMeta + Vendor + RatioConfig`
- `User.InviterId -> User.Id`

## 10. 业务主链路中的数据流

### 10.1 用户调用模型

```text
User
  -> Token
  -> Token.Group / User.Group
  -> Ability(group, model, channel)
  -> Channel
  -> Log
```

理解重点：

- 用户不是直接命中渠道
- 中间经过 Token / Group / Ability 选择

### 10.2 用户充值

```text
User
  -> TopUp
  -> payment callback
  -> User.Quota increase
  -> Log(topup)
```

### 10.3 用户购买订阅

```text
User
  -> SubscriptionPlan
  -> SubscriptionOrder
  -> UserSubscription
  -> optional User.Group change
```

### 10.4 后台展示模型广场

```text
ModelMeta
  + Vendor
  + Ability
  + Ratio settings
  -> Pricing view
```

## 11. 设计上的几个关键取舍

### 11.1 为什么很多表里有冗余字段

例如：

- `Log` 冗余保存 `Username`、`TokenName`、`ModelName`
- `UserSubscription` 冗余保存 `UpgradeGroup`、`PrevUserGroup`

原因是这个系统不是只追求范式，而是追求：

- 查询性能
- 审计可读性
- 历史快照保留

### 11.2 为什么 `Ability` 这么重要

因为这个系统真正的“模型可用性”不是挂在 User 或 Channel 上，而是挂在：

`group + model + channel`

没有这层表，很难优雅支持：

- 多分组
- 多模型
- 多渠道
- 权重与优先级
- 精准禁用/启用

### 11.3 为什么 `Pricing` 不是实体表

因为它本质上是一个聚合视图，而不是事实来源。

如果把它误当成事实表，后续改价格或模型元数据时容易设计错层。

## 12. 结论

这个项目的数据模型最值得记住的三点是：

1. `User` 是账户、额度和身份绑定的中心实体
2. `Token` 是用户对外访问控制的主要载体
3. `Ability` 是渠道、模型、分组三者关系的核心中间层

如果你后续要改模型路由、用户权限、计费或展示逻辑，先判断自己是在改哪一层关系：

- 用户关系
- Token 关系
- Ability 关系
- 订阅/充值关系
- 日志/审计关系

一旦关系层级判断对了，后面的改动位置通常也就清楚了。

