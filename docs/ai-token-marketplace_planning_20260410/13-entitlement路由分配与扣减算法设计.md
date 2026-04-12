# 13-entitlement 路由分配与扣减算法设计

更新时间：2026-04-12

## 1. 文档目标

本文件将 entitlement 从“买方已购权益”映射到“可运行的 seller/channel”，并最终落到“预冻结、实扣、回滚、责任归属”的算法冻结到实现级。

本文件要回答：

1. 请求如何找到可用 entitlement lot
2. lot 与 seller/supply/channel 如何绑定
3. retry 到底允许切什么，不允许切什么
4. 流式请求如何冻结、实扣、回滚
5. usage、日志、结算如何保持一一可追踪

## 2. M1 适用边界

- 仅支持 `text/chat` 类模型
- entitlement 单位统一为 `token`
- 单次请求最终只允许归属 1 个 `entitlement_lot`
- `M1` 不支持一个请求拆分多个 lot
- `M1` 不支持跨 seller 自动切换
- `M1` 不支持 entitlement 与 wallet/subscription 在一个请求里混合扣费

## 3. 冻结决策

### 决策 A：先选 lot，再选 channel

顺序固定为：

1. 先选出唯一候选 `entitlement_lot`
2. 再在该 lot 绑定的 `supply_account` 范围内选 `channel`

### 决策 B：`M1` 只允许同一 supply 内的 channel fallback

- 可在同一 `seller_id + supply_account_id` 下切换多个 `channel`
- 不允许静默跨 seller
- 不允许同一请求一半算给 A seller，一半算给 B seller

### 决策 C：单次请求最终归属唯一 lot

- `usage_ledger` 必须固定写入 `entitlement_lot_id`
- `Log.Other` 必须固定写入 `seller_id/listing_id/order_id/order_item_id`
- 结算不得事后反推 lot

### 决策 D：预冻结先于上游调用

- 发请求前先冻结当前 lot 的 `FrozenAmount`
- 请求成功后转 `UsedAmount`
- 请求失败后回滚 `FrozenAmount`

## 4. 输入与输出

## 4.1 算法输入

运行时输入上下文：

- `user_id`
- `token_id`
- `model_name`
- `pre_consumed_quota`
- `request_id`
- `is_stream`
- `channel_group`

## 4.2 算法输出

算法最终必须返回：

- `entitlement_lot_id`
- `seller_id`
- `supply_account_id`
- `channel_id`
- `billing_source=marketplace_entitlement`
- `pre_frozen_amount`

## 5. 候选 lot 查询规则

## 5.1 查询条件

必须同时满足：

- `buyer_user_id = current_user_id`
- lot `status = active`
- `expire_at = 0` 或 `expire_at > now`
- `granted_amount - used_amount - refunded_amount - frozen_amount > 0`
- 关联 `buyer_entitlement.status = active`
- 关联 `listing.status = active`
- 关联 `supply_account.status = active`
- 关联 `seller_secret.status = active`

模型匹配规则：

- 先按 `buyer_entitlement.model_name = request.model_name`
- `M1` 不做模型别名映射
- 若需要模型 alias，放到 `M2`

## 5.2 候选排序

排序固定为：

1. `expire_at asc`，最早过期优先
2. `priority_seq asc`
3. `id asc`

## 5.3 容量过滤

候选 lot 必须满足：

- `remaining_amount >= pre_consumed_quota`

其中：

- `remaining_amount = granted_amount - used_amount - refunded_amount - frozen_amount`

说明：

- `M1` 不做一个请求拆多个 lot
- 如果没有任何一个 lot 能覆盖本次预冻结额度，则直接返回 `insufficient_marketplace_entitlement`

## 6. seller / supply / channel 映射规则

## 6.1 lot 与责任主体的绑定

`entitlement_lot` 自带以下责任字段：

- `seller_id`
- `supply_account_id`
- `listing_id`
- `order_id`
- `order_item_id`

因此：

- 一旦 lot 选中，本次请求的责任主体已经确定
- 后续 channel 只能在该 `supply_account` 绑定池内选择

## 6.2 channel 解析规则

对已选 lot，运行时按以下顺序找 channel：

1. 查 `supply_channel_binding.status = active`
2. 按 `binding_role=primary` 优先
3. 再按 `priority asc`
4. 过滤掉运行时已禁用 channel
5. 通过 `secret_runtime_resolver` 取到该 channel 的可用凭证

## 6.3 `M1` 不做的事

- 不跨 seller 自动找新 lot
- 不跨 supply 自动找新 lot
- 不因为 generic route 可用就跳出 entitlement 责任域

## 7. Retry 规则

## 7.1 允许的 retry

允许：

- 在同一 `entitlement_lot`
- 同一 `seller_id`
- 同一 `supply_account_id`
- 不同 `channel_id`

适用场景：

- channel 超时
- 上游 5xx
- channel 已自动 ban

## 7.2 不允许的 retry

不允许：

- 切到另一个 seller
- 切到另一个 lot
- entitlement 不足时自动回落到 wallet/subscription

原因：

- 否则责任归属和结算口径会漂移
- `M1` 先保证账务与审计正确，再追求跨 seller 高可用

## 7.3 retry 成功后的责任

- 即便切换了 channel，`seller_id / supply_account_id / entitlement_lot_id` 不变
- 只有 `channel_id` 可能变化

## 8. 预冻结、实扣与回滚

## 8.1 预冻结时机

在真正调用上游前完成：

1. 选择 lot
2. 对该 lot 做原子冻结
3. 将 `frozen_amount += pre_consumed_quota`
4. 将 lot 信息写入本次请求上下文

## 8.2 原子冻结规则

建议使用事务 + 条件更新：

- `WHERE id = ? AND status = 'active' AND remaining_amount >= pre_consumed_quota`

效果：

- 并发请求不会超扣同一 lot
- 失败时可立即返回 entitlement 不足

## 8.3 实扣规则

请求成功后：

- `used_amount += actual_quota`
- `frozen_amount -= pre_consumed_quota`
- 如 `pre_consumed_quota > actual_quota`
  - 差额自动释放

说明：

- `M1` 沿用现有 `preConsumedQuota` 作为预冻结上界
- 对 `text/chat` 请求，要求预冻结策略足够保守，原则上 `actual_quota <= pre_consumed_quota`

## 8.4 回滚规则

请求失败后：

- `frozen_amount -= pre_consumed_quota`
- `used_amount` 不变
- 写一条失败或退款型 `usage_ledger`

幂等键建议：

- `request_id + refund_final`

## 9. 流式场景

## 9.1 流式成功结束

处理规则：

1. 使用最终 usage 结算
2. lot 执行实扣
3. 写 `usage_ledger.status=success`

## 9.2 中途断流，但拿到最终 usage

- 仍按最终 usage 结算
- 释放多余冻结额度

## 9.3 中途断流，未拿到最终 usage

`M1` 冻结规则：

- 不允许直接按 0 扣减
- 使用平台可确认的最小可信 usage 结算
- 若无法得到可信 usage，则按 `pre_consumed_quota` 结算并记录异常

原因：

- entitlement 是已购权益，不能因为流式中断直接清零
- 相比“漏扣”，`M1` 优先保证账务不被钻空子

## 10. 责任归属与日志落点

每条最终成功的 `usage_ledger` 必须固定以下字段：

- `seller_id`
- `supply_account_id`
- `channel_id`
- `listing_id`
- `order_id`
- `order_item_id`
- `entitlement_lot_id`
- `billing_source = marketplace_entitlement`

`Log.Other` 至少挂入：

- `seller_id`
- `listing_id`
- `order_id`
- `order_item_id`
- `entitlement_lot_id`
- `billing_source`

## 11. 伪代码

```text
ResolveMarketplaceEntitlement(request):
  candidates = QueryEligibleLots(user_id, model_name, now)
  sort by expire_at asc, priority_seq asc, id asc

  for lot in candidates:
    if lot.remaining_amount < pre_consumed_quota:
      continue

    channels = ResolveChannelsBySupply(lot.supply_account_id)
    if channels is empty:
      continue

    if TryFreezeLot(lot.id, pre_consumed_quota) fails:
      continue

    return {
      lot,
      seller_id,
      supply_account_id,
      channels,
      billing_source = marketplace_entitlement
    }

  return insufficient_marketplace_entitlement
```

## 12. 与现有计费会话的集成建议

落点建议：

- `service/channel_select.go`
- `service/billing.go`
- `service/billing_session.go`
- `service/funding_source.go`
- `service/secret_runtime_resolver.go`

需要新增：

- `BillingSourceMarketplaceEntitlement`
- `MarketplaceEntitlementFunding`

核心要求：

- entitlement 请求不扣减 `user.quota`
- entitlement 请求不扣减 `token.remain_quota`
- entitlement 请求不走 subscription 预扣

## 13. 不建议的实现

- 请求成功后再反推 entitlement lot
- 同一次请求拆到多个 seller
- entitlement 不足时静默改走 wallet/subscription
- 流式失败直接按 0 扣减
- 明文 key 从 controller 直接传给 relay

## 14. 验收标准

满足以下条件时，本专题视为完成：

1. 单次请求可稳定选出唯一 `entitlement_lot`
2. retry 仅发生在同一 supply 的 channel 范围内
3. 预冻结、实扣、回滚都有明确数据动作
4. `usage_ledger` 与 `Log.Other` 能追溯唯一责任主体
5. entitlement 请求不会回落到 wallet/subscription
6. 文档足以直接指导 `Task 4` 编码
