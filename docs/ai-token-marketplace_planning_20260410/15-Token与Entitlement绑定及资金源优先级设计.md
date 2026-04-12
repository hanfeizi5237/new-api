# 15-Token与Entitlement绑定及资金源优先级设计

更新时间：2026-04-12

## 1. 文档目标

本文件冻结 `new-api-main` 现有 `Token / wallet / subscription` 体系与 Marketplace entitlement 之间的关系，重点解决：

1. 平台 `Token` 到底是认证载体还是计费载体
2. entitlement 归属于 `user` 还是 `token`
3. 用户同时拥有 entitlement、wallet、subscription 时，请求优先消耗谁
4. 如何避免一个请求被双重扣费
5. relay 层需要新增哪些 billing source

## 2. 现状约束

当前底座已有计费来源：

- `wallet`
- `subscription`

当前计费链路的核心特点：

- `Token` 既承担认证，也参与额度扣减
- `BillingSession` 会同时处理 token 额度与用户资金来源
- `service/funding_source.go` 只抽象了 `wallet/subscription`

如果直接把 entitlement 接进现有链路，又不重写优先级，会出现两个严重问题：

1. 买方买了 entitlement 后，请求仍然继续扣用户钱包或订阅
2. 同一次 entitlement 请求还会继续扣 `Token.RemainQuota`

这会造成双重扣费，必须在 `M1` 前冻结。

## 3. 冻结决策

### 决策 A：Marketplace entitlement 归属 `buyer_user_id`

- entitlement 的所有权归用户，不归 token
- 同一用户下的合法平台 token 都可消费该用户 entitlement
- `M1` 不做 token 级专属绑定

### 决策 B：平台 Token 在 entitlement 流程中只承担认证与访问控制

- token 仍负责鉴权
- token 仍负责模型白名单、分组限制、状态校验
- token 在 entitlement 请求中不再承担资金扣减

### 决策 C：引入第三种资金来源

新增：

- `BillingSourceMarketplaceEntitlement = "marketplace_entitlement"`

### 决策 D：单次请求只允许一种最终资金来源

优先级不是“分摊式”，而是“命中式”：

1. 命中 entitlement，则本次请求全部走 entitlement
2. 未命中 entitlement，才进入原有 `wallet/subscription` 决策

### 决策 E：`M1` 禁止 entitlement 不足时静默回落

如果用户存在 matching entitlement，但余额不足：

- 直接返回 `insufficient_marketplace_entitlement`
- 不自动改扣 wallet
- 不自动改扣 subscription

原因：

- 避免用户以为消费的是已购权益，实际却额外花钱包

## 4. 绑定关系冻结

## 4.1 `User` 与 `Token`

- `Token.UserId` 继续定义 token 所属用户
- 只有该用户的 token 才能访问该用户 entitlement

## 4.2 `User` 与 `BuyerEntitlement`

- `buyer_entitlement.buyer_user_id` 是 entitlement 唯一所有权锚点
- entitlement 聚合维度保持：
  - `buyer_user_id`
  - `vendor_id`
  - `model_name`

## 4.3 `Token` 与 `BuyerEntitlement`

`M1` 绑定规则：

- 不新增 `token_entitlement_binding` 表
- 不支持“某个 token 只能消费某个 entitlement”
- 只做“同用户 token 共享 entitlement”

未来可演进：

- 为企业用户增加 token 级 entitlement 隔离
- 为 API key 增加 marketplace-only / generic-only 策略

## 5. 资金源决策顺序

## 5.1 决策入口

在 relay 请求进入计费前，先做 `ResolveBillingSource`：

输入：

- `user_id`
- `token_id`
- `model_name`
- `pre_consumed_quota`
- 用户原有 billing preference

输出：

- `billing_source`
- 如命中 entitlement，返回 `entitlement_lot_id`

## 5.2 决策顺序

固定顺序：

1. 验证 token 是否属于当前用户
2. 验证 token 是否允许当前模型
3. 查询 matching entitlement
4. 如果存在可覆盖本次请求的 entitlement：
   - 选 `marketplace_entitlement`
5. 如果存在 matching entitlement 但余额不足：
   - 返回 entitlement 不足
6. 如果不存在 matching entitlement：
   - 再按原有 `wallet_first / subscription_first / only` 决策

## 5.3 决策表

| 场景 | 结果 |
| --- | --- |
| 有 matching entitlement，且余额充足 | 走 `marketplace_entitlement` |
| 有 matching entitlement，但余额不足 | 返回 entitlement 不足 |
| 无 matching entitlement，用户偏好 `subscription_first` | 走现有 subscription 优先逻辑 |
| 无 matching entitlement，用户偏好 `wallet_first` | 走现有 wallet 优先逻辑 |
| 无 matching entitlement，用户偏好 `subscription_only` 但无订阅 | 返回订阅不足 |

## 6. 为什么 `M1` 不做 entitlement 与 generic 资金混合

如果允许一个请求：

- 先吃一半 entitlement
- 剩下走 wallet 或 subscription

会产生这些复杂度：

- `usage_ledger` 不能绑定唯一资金来源
- settlement 无法直接归属
- 前端很难解释“本次请求到底花了什么”
- 退款与争议会明显复杂化

因此 `M1` 冻结：

- 单次请求只允许一个 funding source
- 不允许 entitlement + wallet/subscription 混合

## 7. 对 Token 额度的处理规则

## 7.1 generic 请求

未命中 entitlement 时：

- 继续沿用现有 token 额度预扣和结算逻辑

## 7.2 entitlement 请求

命中 `marketplace_entitlement` 时：

- token 不再扣减 `remain_quota`
- user 不再扣减 `quota`
- subscription 不再做预扣
- entitlement 成为唯一资金来源

## 7.3 token 在 entitlement 请求中仍保留的职责

- 鉴权
- 令牌状态检查
- 模型权限检查
- 分组权限检查
- 审计与日志定位

## 8. relay 层落地建议

## 8.1 新增 billing source 常量

建议在 `service/billing.go` 增加：

```go
const BillingSourceMarketplaceEntitlement = "marketplace_entitlement"
```

## 8.2 新增 funding source 实现

建议在 `service/funding_source.go` 或独立文件中新增：

- `MarketplaceEntitlementFunding`

职责：

- `PreConsume(amount)`
  - 冻结 entitlement lot
- `Settle(delta)`
  - 调整 entitlement 实扣
- `Refund()`
  - 回滚 entitlement 预冻结

## 8.3 `BillingSession` 的新分支

`BillingSession` 必须支持：

- tokenConsumed = 0
- fundingSource = marketplace_entitlement
- 不调用 `DecreaseTokenQuota`
- 不调用 `DecreaseUserQuota`

## 8.4 `RelayInfo` 建议新增或复用字段

至少要能携带：

- `BillingSource = marketplace_entitlement`
- `EntitlementLotId`
- `SellerId`
- `SupplyAccountId`
- `OrderId`
- `OrderItemId`

## 9. 日志与账务要求

## 9.1 `Log.Other`

至少新增：

- `billing_source`
- `entitlement_lot_id`
- `seller_id`
- `listing_id`
- `order_id`
- `order_item_id`

## 9.2 `usage_ledger`

必须固化：

- `event_key`
- `request_id`
- `billing_source = marketplace_entitlement`
- `entitlement_lot_id`
- `token_id`
- `user_id`

## 9.3 禁止的账务行为

- entitlement 请求扣用户钱包
- entitlement 请求扣用户订阅
- entitlement 请求扣 token remain quota
- generic 请求反向占用 marketplace lot

## 10. 异常场景冻结

## 10.1 用户存在 entitlement，但 token 不属于该用户

- 直接拒绝
- 不允许跨用户消费 entitlement

## 10.2 用户存在 entitlement，但 token 不允许该模型

- 仍按 token 模型权限拒绝
- entitlement 不改变 token 的 ACL 语义

## 10.3 用户同时有 entitlement 和 subscription

如果当前请求模型命中 entitlement：

- 强制 entitlement 优先
- 不进入 subscription_first 逻辑

## 10.4 用户同时有 entitlement 和 wallet

如果当前请求模型命中 entitlement：

- 强制 entitlement 优先
- 不进入 wallet_first 逻辑

## 11. 伪代码

```text
ResolveBillingSource(request):
  validate token ownership and token ACL

  lot = FindEligibleEntitlementLot(user_id, model_name, pre_consumed_quota)
  if lot exists:
    return marketplace_entitlement(lot)

  lot_insufficient = FindMatchingButInsufficientEntitlement(user_id, model_name)
  if lot_insufficient:
    return error(insufficient_marketplace_entitlement)

  return ResolveLegacyFundingSource(user_preference)
```

## 12. 与其他专题文档的关系

本文件需要与以下文档严格一致：

- `12-支付订单与权益发放时序设计.md`
- `13-entitlement路由分配与扣减算法设计.md`
- `14-库存口径、超卖控制与自动停售规则设计.md`
- `11-SellerKeyVault与密钥生命周期设计.md`

## 13. 不建议的实现

- entitlement 挂在 token 上而不是 user 上
- 命中 entitlement 后仍继续扣 token quota
- entitlement 不足时静默改走 wallet
- entitlement 不足时静默改走 subscription
- 一个请求拆两个 funding source

## 14. 验收标准

满足以下条件时，本专题视为完成：

1. entitlement 所有权、token 消费权限、资金来源优先级全部冻结
2. `marketplace_entitlement` 成为独立 billing source
3. entitlement 请求不会再触发 token/wallet/subscription 双重扣费
4. matching entitlement 存在但不足时，系统返回明确错误而非静默回落
5. 文档足以直接指导 `Task 4` 与 `BillingSession` 改造
