# 20-API契约与错误码规范

更新时间：2026-04-16

## 1. 文档目标

本文件用于冻结 Marketplace 在 `new-api-main` 中的接口风格、DTO 命名、分页规则、错误码规范和幂等约定，重点解决：

1. Marketplace 接口如何延续现有仓库风格
2. 买方、卖家、平台后台、支付 webhook 的接口边界如何划分
3. 分页、时间戳、金额、状态等字段如何统一命名
4. 业务错误码如何稳定输出给前端和任务系统
5. 幂等键、请求追踪、敏感动作错误响应如何表达

## 2. M1 适用边界

- `M1` 继续沿用现有 `common.ApiSuccess / common.ApiError` 风格
- 首期不重构全站 API 风格，不引入 GraphQL，不改现有前端请求层习惯
- Marketplace 的管理端、买方端、自助卖家端都继续挂在现有 `/api` 下
- 支付 webhook 属于第三方回调接口，响应需优先满足支付平台契约，不强制使用统一前端返回包
- 可选标量字段如需保留显式零值，应遵循指针字段策略，避免 `omitempty` 吞零值

## 3. 冻结决策

### 决策 A：优先兼容现有返回包装

管理端、买方端、自助卖家端接口默认使用：

成功：

```json
{
  "success": true,
  "message": "",
  "data": {}
}
```

失败：

```json
{
  "success": false,
  "message": "listing is not active",
  "error_code": "MARKET_LISTING_NOT_ACTIVE",
  "request_id": "req_xxx"
}
```

说明：

- 当前仓库已有 `success/message/data` 约定，Marketplace 不另起一套壳
- 为了让前端、任务系统和审计系统能稳定识别错误，Marketplace 失败响应应补充 `error_code`
- 若现有 `common.ApiError` 暂不支持 `error_code`，Marketplace controller 可新增兼容 helper

### 决策 B：分页参数沿用现有 `p + page_size`

请求分页参数统一：

- `p`
- `page_size`

兼容读取：

- `ps`
- `size`

响应统一返回 `PageInfo`：

```json
{
  "page": 1,
  "page_size": 20,
  "total": 153,
  "items": []
}
```

### 决策 C：字段命名统一使用 `snake_case`

适用于：

- query 参数
- JSON body
- JSON response

### 决策 D：金额、额度、时间字段必须带语义后缀

统一规则：

- 金额：`*_amount_minor`、`*_price_minor`、`*_fee_minor`
- token / entitlement 额度：`*_amount`、`*_quota`
- 时间戳：`*_at`
- 状态：`*_status`
- 原因：`*_reason`

### 决策 E：边界按调用方拆分，不按页面拆分

接口域划分：

- 平台后台：admin
- 卖家自助：self
- 买方：market
- 第三方 webhook：payment/*

## 4. 路由空间冻结

## 4.1 管理端

- `/api/seller/admin`
- `/api/marketplace/admin/seller-secrets`
- `/api/listing/admin`
- `/api/order/admin`
- `/api/settlement/admin`
- `/api/dispute/admin`

## 4.2 卖家端

- `/api/seller/self`
- `/api/marketplace/self/seller-secrets`
- `/api/seller/supplies`
- `/api/listing/self`

## 4.3 买方端

- `/api/market/listings`
- `/api/market/listings/:id`
- `/api/market/orders`
- `/api/market/orders/:id`
- `/api/market/orders/:id/pay`
- `/api/market/entitlements`

## 4.4 第三方回调端

- `/api/market/payment/epay/notify`
- `/api/market/payment/stripe/webhook`
- `/api/market/payment/creem/webhook`
- `/api/market/payment/waffo/webhook`

说明：

- 这些接口的返回值以支付平台验签、重试和 ACK 要求优先
- 但内部仍必须统一映射为 Marketplace 错误码与审计字段

## 5. DTO 设计规则

## 5.1 请求 DTO

规则：

1. 入参字段使用 `snake_case` JSON tag
2. ID 使用 `int`
3. 金额与额度使用 `int64`
4. 状态值使用稳定字符串枚举
5. 可选标量字段在需要保留显式零值时使用指针

## 5.2 响应 DTO

规则：

1. 管理端返回 View DTO，不直接暴露底层模型敏感字段
2. secret 默认只返回掩码，不返回明文
3. 状态字段必须稳定，不能临时拼文案
4. 后台列表优先返回结构化字段，不把完整 JSON blob 直接塞给前端

## 5.3 时间与状态表达

统一要求：

- 时间统一使用 Unix 时间戳秒级 `int64`
- 状态值统一小写下划线或小写单词
- 前端展示文案由 UI 层映射，不在接口里混合中英文描述

## 6. 分页、筛选与排序规范

## 6.1 分页参数

统一：

- `p`
- `page_size`

默认值：

- `p=1`
- `page_size=20`

上限：

- `page_size <= 100`

## 6.2 列表筛选

列表接口统一优先支持：

- `keyword`
- `status`
- 域专属筛选字段，例如：
  - `audit_status`
  - `seller_id`
  - `buyer_user_id`
  - `payment_status`
  - `entitlement_status`

## 6.3 排序

`M1` 先统一使用后端固定排序：

- 列表默认 `id desc`

## 7. 幂等与请求追踪

## 7.1 买方下单

冻结：

- `POST /api/market/orders` 使用 `client_nonce`
- 幂等键语义：`buyer_user_id + client_nonce`

## 7.2 支付回调

冻结：

- 幂等键来源于 `payment_provider + payment_trade_no + callback_stage`
- 重复回调必须返回可接受 ACK，但不能重复落账

## 7.3 人工敏感动作

建议：

- 支持 `X-Request-ID`
- 支持 `X-Idempotency-Key` 或 body 中稳定幂等字段

## 8. 错误响应规范

## 8.1 第一方接口

Marketplace 第一方接口建议返回：

```json
{
  "success": false,
  "message": "inventory is insufficient",
  "error_code": "MARKET_INVENTORY_INSUFFICIENT",
  "request_id": "req_20260416_xxx"
}
```

说明：

- HTTP 状态码继续兼容现有站内风格，可保持 `200`
- 机器可判断的错误语义由 `error_code` 提供
- `message` 面向人类阅读或 i18n 文案映射

## 8.2 第三方 webhook

第三方 webhook 返回：

- 优先遵循支付平台 ACK 规范
- 内部审计日志必须仍然落 Marketplace 错误码

## 8.3 任务系统与内部服务

内部服务错误结构至少应包含：

- `error_code`
- `message`
- `retryable`
- `request_id`

## 9. 错误码分层

统一建议格式：

```text
MARKET_<DOMAIN>_<DETAIL>
```

## 9.1 通用错误码

| 错误码 | 含义 | 是否建议重试 |
| --- | --- | --- |
| `MARKET_VALIDATION_INVALID_ARGUMENT` | 参数错误 | 否 |
| `MARKET_AUTH_UNAUTHORIZED` | 未登录或登录失效 | 否 |
| `MARKET_PERMISSION_FORBIDDEN` | 权限不足 | 否 |
| `MARKET_RESOURCE_NOT_FOUND` | 资源不存在 | 否 |
| `MARKET_STATE_CONFLICT` | 状态冲突 | 否 |
| `MARKET_INTERNAL_RETRYABLE` | 可重试内部错误 | 是 |
| `MARKET_INTERNAL_FATAL` | 不可自动恢复内部错误 | 否 |

## 9.2 卖家 / 商品域

| 错误码 | 含义 | 是否建议重试 |
| --- | --- | --- |
| `MARKET_SELLER_NOT_FOUND` | 卖家不存在 | 否 |
| `MARKET_SELLER_NOT_APPROVED` | 卖家未审核通过 | 否 |
| `MARKET_SUPPLY_NOT_FOUND` | 供给不存在 | 否 |
| `MARKET_SECRET_DISABLED` | 卖家 secret 已停用 | 否 |
| `MARKET_SECRET_VERIFY_FAILED` | secret 验证失败 | 否 |
| `MARKET_SECRET_REVEAL_FORBIDDEN` | 无权读取明文 | 否 |
| `MARKET_SECURE_VERIFICATION_REQUIRED` | 缺少二次验证 | 否 |
| `MARKET_LISTING_NOT_FOUND` | 商品不存在 | 否 |
| `MARKET_LISTING_NOT_ACTIVE` | 商品不可售 | 否 |
| `MARKET_SKU_NOT_FOUND` | SKU 不存在 | 否 |
| `MARKET_INVENTORY_INSUFFICIENT` | 可售库存不足 | 否 |

## 9.3 订单 / 支付域

| 错误码 | 含义 | 是否建议重试 |
| --- | --- | --- |
| `MARKET_ORDER_NOT_FOUND` | 订单不存在 | 否 |
| `MARKET_ORDER_EXPIRED` | 订单已过期 | 否 |
| `MARKET_ORDER_ALREADY_PAID` | 订单已支付 | 否 |
| `MARKET_ORDER_ALREADY_CLOSED` | 订单已关闭 | 否 |
| `MARKET_ORDER_NOT_PAYABLE` | 当前状态不可支付 | 否 |
| `MARKET_PAYMENT_PROVIDER_UNSUPPORTED` | 支付渠道不支持 | 否 |
| `MARKET_PAYMENT_CALLBACK_INVALID` | 回调验签失败或参数非法 | 否 |
| `MARKET_PAYMENT_CALLBACK_DUPLICATED` | 回调重复 | 否 |
| `MARKET_PAYMENT_PROVIDER_TIMEOUT` | 支付平台超时 | 是 |
| `MARKET_PAYMENT_RECONCILE_REQUIRED` | 需进入对账补扫 | 否 |

## 9.4 entitlement / relay / usage 域

| 错误码 | 含义 | 是否建议重试 |
| --- | --- | --- |
| `MARKET_ENTITLEMENT_NOT_FOUND` | 无对应 entitlement | 否 |
| `MARKET_ENTITLEMENT_INSUFFICIENT` | entitlement 不足 | 否 |
| `MARKET_FUNDING_SOURCE_CONFLICT` | 命中了非法资金源组合 | 否 |
| `MARKET_USAGE_FINALIZE_RETRYABLE` | usage finalize 暂失败 | 是 |
| `MARKET_UPSTREAM_TEMPORARY_UNAVAILABLE` | 上游暂不可用 | 是 |
| `MARKET_CHANNEL_NOT_AVAILABLE` | 候选 channel 不可用 | 否 |

## 9.5 任务 / 补偿域

| 错误码 | 含义 | 是否建议重试 |
| --- | --- | --- |
| `MARKET_TASK_LOCK_CONFLICT` | 任务抢锁失败 | 是 |
| `MARKET_TASK_NOT_READY` | 任务未到执行时间 | 否 |
| `MARKET_TASK_MAX_RETRY_EXCEEDED` | 超过最大重试次数 | 否 |
| `MARKET_COMPENSATION_REQUIRES_MANUAL_REVIEW` | 需人工处理 | 否 |

## 10. 与现有文档的关系

- 支付、发权事务边界对齐 `12-支付订单与权益发放时序设计.md`
- entitlement 选路与资金源优先级对齐 `15-Token与Entitlement绑定及资金源优先级设计.md`
- 结算与冲销错误语义对齐 `16-结算、佣金、退款与冲销公式设计.md`
- 敏感动作权限与 `SecureVerificationRequired` 对齐 `17-RBAC、敏感操作与审计权限矩阵设计.md`
- 异步重试、死信、补偿语义对齐 `19-异步事件、任务重试与补偿机制设计.md`

## 11. 验收标准

满足以下条件时，本专题视为完成：

1. Marketplace 接口已明确沿用现有 `success/message/data` 包装
2. 分页、字段命名、金额/时间后缀规则已冻结
3. 买方、卖家、管理端、支付 webhook 的路由边界已写清
4. 业务错误码已形成稳定清单，并标明是否可重试
5. 文档足以直接指导 controller DTO、前端 hooks 和异步任务错误处理改造
