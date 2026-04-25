# 17-RBAC、敏感操作与审计权限矩阵设计

更新时间：2026-04-16

## 1. 文档目标

本文件冻结 Marketplace 在 `new-api-main` 上的最小权限模型、敏感操作二次验证规则和审计落点，重点解决：

1. 谁能看、谁能改、谁能审核
2. 哪些操作必须 `SecureVerificationRequired`
3. 哪些行为必须写审计记录
4. 卖家密钥、订单补偿、库存恢复、发权重放等高风险动作如何管控

## 2. M1 适用边界

- 沿用现有 `UserAuth / AdminAuth / RootAuth`
- `M1` 不要求先引入复杂权限中心，但必须冻结动作级权限矩阵
- `M1` 卖家自助能力默认只读或关闭，写操作以平台后台代办为主
- 前端菜单可做可见性控制，但后端权限校验必须独立成立

## 3. 角色定义

## 3.1 基础角色

- `buyer`
  - 普通买方用户
- `seller_readonly`
  - 已入驻卖家的一期只读视角
- `ops_admin`
  - 运营管理员，负责 seller/listing 审核与日常维护
- `risk_admin`
  - 风控管理员，负责暂停、恢复、风控动作
- `finance_admin`
  - 财务管理员，负责退款复核、结算释放、对账
- `root`
  - 最高权限，仅处理明文读取、强制补偿、系统级配置
- `system_task`
  - 定时任务与系统内部动作

## 3.2 与现有认证方式映射

| 角色 | 建议认证门槛 |
| --- | --- |
| `buyer` | `UserAuth` |
| `seller_readonly` | `UserAuth` + seller ownership |
| `ops_admin` | `AdminAuth` |
| `risk_admin` | `AdminAuth` + risk scope |
| `finance_admin` | `AdminAuth` + finance scope |
| `root` | `RootAuth` |
| `system_task` | internal task context |

说明：

- 若当前仓库尚未有精细化 admin scope，可先在 controller/service 层用显式校验函数收口
- 不允许只靠前端隐藏按钮实现权限控制

## 4. 冻结决策

### 决策 A：密钥明文读取只能由 `root` 执行

必须同时满足：

- `RootAuth`
- `SecureVerificationRequired`
- 必填 `reason`
- 记录 reveal audit

### 决策 B：涉及资金、库存、发权补偿的管理动作必须二次验证

至少包括：

- 手动关闭已支付订单
- entitlement 重放
- 手动退款
- 库存人工恢复上架
- 结算冲销
- 强制禁用/恢复 seller secret

### 决策 C：所有状态机跃迁都必须可审计

至少覆盖：

- seller 审核
- listing 审核与状态切换
- seller_secret 生命周期动作
- order 支付与关闭
- entitlement 发放与重放
- settlement 冻结、释放、冲销
- dispute 裁决执行

### 决策 D：系统任务与人工操作必须区分 actor 类型

审计中必须明确：

- `actor_type = admin/root/system/task/user`
- 不允许把系统动作伪装成人工动作

## 5. 动作权限矩阵

| 动作 | 最低角色 | 二次验证 | 审计 |
| --- | --- | --- | --- |
| 浏览市场商品 | `buyer` | 否 | 否 |
| 查看自己的订单/entitlement | `buyer` | 否 | 否 |
| 浏览卖家只读信息 | `seller_readonly` | 否 | 否 |
| 创建 seller | `ops_admin` | 否 | 是 |
| 审核 seller | `ops_admin` | 否 | 是 |
| 创建/编辑 listing | `ops_admin` | 否 | 是 |
| 审核 listing | `ops_admin` | 否 | 是 |
| 手动暂停 listing | `ops_admin` | 否 | 是 |
| 人工恢复 listing | `risk_admin` | 是 | 是 |
| 导入 seller secret | `ops_admin` | 否 | 是 |
| 验证 seller secret | `ops_admin` | 否 | 是 |
| 禁用 seller secret | `risk_admin` | 是 | 是 |
| 恢复 seller secret | `risk_admin` | 是 | 是 |
| 查看 seller secret 明文 | `root` | 是 | 是 |
| 强制同步 secret 到 channel mirror | `root` | 是 | 是 |
| 创建市场订单 | `buyer` | 否 | 是 |
| 拉起支付 | `buyer` | 否 | 是 |
| 接收支付 webhook | `system_task` | 否 | 是 |
| entitlement 发放 | `system_task` | 否 | 是 |
| entitlement 手动重放 | `finance_admin` | 是 | 是 |
| 手动关闭已支付订单 | `finance_admin` | 是 | 是 |
| 手动退款 | `finance_admin` | 是 | 是 |
| 执行 settlement release | `finance_admin` | 是 | 是 |
| 执行 settlement reversal | `finance_admin` | 是 | 是 |
| 创建 dispute | `buyer` | 否 | 是 |
| 裁决 dispute | `risk_admin` | 是 | 是 |

## 6. 卖家端边界

`M1` 冻结：

- 卖家端默认不开放明文密钥读取
- 卖家端默认不开放 listing 自主上架写接口
- 卖家端默认不开放提现和结算出账操作
- 如果保留 seller self 接口，应只读返回：
  - 自己的供给状态
  - 商品审核状态
  - 订单汇总报表
  - 脱敏后的 secret 验证状态

## 7. 二次验证规则

## 7.1 必须 `SecureVerificationRequired` 的动作

- seller secret 明文读取
- seller secret disable / recover / rotate
- inventory 异常后的人工恢复上架
- entitlement 重放
- 已支付订单手动关闭
- 手动退款
- settlement release / reversal
- dispute 裁决执行

## 7.2 必须填写 `reason` 的动作

- 所有上表中的二次验证动作
- 任何越过正常状态机的人工 override

`reason` 要求：

- 不允许为空
- 不允许只写“test”或“fix”
- 应可被审计人员理解

## 8. 审计模型要求

建议统一形成 `marketplace operation audit` 口径；若 `M1` 不新建总审计表，也必须保证各域有可回查记录。

每条审计至少包含：

- `actor_user_id`
- `actor_type`
- `action`
- `target_type`
- `target_id`
- `request_id`
- `ip`
- `reason`
- `result`
- `before_state`
- `after_state`
- `meta`
- `created_at`

说明：

- `seller_secret_audit` 继续承担 secret 域事实审计
- 其他域至少应通过操作日志或 domain audit 字段补齐
- `Log.Other` 只能作为调用证据，不替代管理审计

## 9. 接口与权限落点建议

建议在 controller 层统一收口：

- `RequireMarketplaceAdminScope(ctx, scope)`
- `RequireSecureMarketplaceAction(ctx, action)`
- `RequireOwnedBuyerResource(ctx, userID)`
- `RequireOwnedSellerResource(ctx, sellerUserID)`

scope 建议至少分：

- `market_ops`
- `market_risk`
- `market_finance`
- `market_secret_root`

## 10. 不建议的实现

- 只靠前端菜单控制 admin 权限
- `AdminAuth` 获得所有 secret 明文读取能力
- entitlement 重放不留审计
- 系统任务与人工操作共用同一 actor 标识
- 高风险动作不要求 `reason`

## 11. 验收标准

满足以下条件时，本专题视为完成：

1. Marketplace 主要动作都有最小角色边界
2. 高风险动作都已冻结二次验证规则
3. secret、订单、发权、库存、结算、争议都有审计要求
4. `M1` 卖家端与平台后台边界不再模糊
5. 文档足以直接指导 controller/service 权限改造
