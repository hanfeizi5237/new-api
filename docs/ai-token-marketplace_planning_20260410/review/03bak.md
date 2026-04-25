1 +# AI Token 集市平台 — M1 修复清单                                                                                                                                                             
2 +                                                                                                                                                                                              
3 +> **整理日期**: 2026-04-16                                                                                                                                                                    
4 +> **评审范围**: `dev` 分支当前头部 commit `17df9868`（`fix M1完成`）                                                                                                                          
5 +> **依据文档**:                                                                                                                                                                               
6 +> - `07-ClawTeam_基于new-api-main的完整实施蓝图.md`                                                                                                                                           
7 +> - `09-M1_交易闭环开发任务书.md`                                                                                                                                                             
8 +> - `12-支付订单与权益发放时序设计.md`                                                                                                                                                        
9 +> - `13-entitlement路由分配与扣减算法设计.md`                                                                                                                                                 
10 +> - `15-Token与Entitlement绑定及资金源优先级设计.md`                                                                                                                                          
11 +> - `review/01-代码审查报告_20260413.md`                                                                                                                                                      
12 +> - `review/02-修复验证审查报告_20260413.md`                                                                                                                                                  
13 +                                                                                                                                                                                              
14 +---                                                                                                                                                                                           
15 +                                                                                                                                                                                              
16 +## 一、当前结论                                                                                                                                                                               
17 +                                                                                                                                                                                              
18 +当前 `dev` 分支已经具备 M1 的主要代码骨架，但**还不能视为“可稳定交付的 M1 完成态”**。                                                                                                         
19 +                                                                                                                                                                                              
20 +原因不是“功能完全没做完”，而是仍存在几处会直接影响：                                                                                                                                          
21 +                                                                                                                                                                                              
22 +- 并发安全                                                                                                                                                                                    
23 +- 支付成功后的失败补偿                                                                                                                                                                        
24 +- entitlement 使用账口径                                                                                                                                                                      
25 +- usage 审计可追溯性                                                                                                                                                                          
26 +                                                                                                                                                                                              
27 +建议结论：                                                                                                                                                                                    
28 +                                                                                                                                                                                              
29 +- **允许继续开发和修复**                                                                                                                                                                      
30 +- **不建议宣布 M1 已完成**                                                                                                                                                                    
31 +- **不建议在当前状态直接进入 M2**                                                                                                                                                             
32 +                                                                                                                                                                                              
33 +---                                                                                                                                                                                           
34 +                                                                                                                                                                                              
35 +## 二、修复优先级总表                                                                                                                                                                         
36 +                                                                                                                                                                                              
37 +| 优先级 | 问题 | 结论 |                                                                                                                                                                      
38 +| --- | --- | --- |                                                                                                                                                                           
39 +| `P0` | `lockForUpdate` 实现错误，实际未加行锁 | 必须先修 |                                                                                                                                  
40 +| `P0` | 支付成功但发权失败时，`entitlement_status=failed` 无法真正落库 | 必须先修 |                                                                                                          
41 +| `P1` | entitlement 成功结算后未回写 `market_order_item.used_amount` | 必须补齐 |                                                                                                            
42 +| `P1` | `usage_ledger` 未写入 token usage 明细字段 | 必须补齐 |                                                                                                                              
43 +| `P2` | marketplace 列表接口存在明显 N+1 查询 | 建议在 M1 收口前修 |                                                                                                                         
44 +                                                                                                                                                                                              
45 +---                                                                                                                                                                                           
46 +                                                                                                                                                                                              
47 +## 三、P0 修复项                                                                                                                                                                              
48 +                                                                                                                                                                                              
49 +### 1. 修复 `lockForUpdate` 实现错误                                                                                                                                                          
50 +                                                                                                                                                                                              
51 +**问题文件**                                                                                                                                                                                  
52 +                                                                                                                                                                                              
53 +- `service/marketplace_inventory.go:11`                                                                                                                                                       
54 +                                                                                                                                                                                              
55 +**当前问题**                                                                                                                                                                                  
56 +                                                                                                                                                                                              
57 +当前实现：                                                                                                                                                                                    
58 +                                                                                                                                                                                              
59 +```go                                                                                                                                                                                         
     60 +if common.UsingSQLite || common.UsingMySQL || common.UsingPostgreSQL {                                                                                                                        
     61 +    return query                                                                                                                                                                              
     62 +}                                                                                                                                                                                             
     63 +return query.Set("gorm:query_option", "FOR UPDATE")                                                                                                                                           
     64 +```                                                                                                                                                                                           
65 +                                                                                                                                                                                              
66 +这会导致：                                                                                                                                                                                    
67 +                                                                                                                                                                                              
68 +- SQLite 不加锁                                                                                                                                                                               
69 +- MySQL 不加锁                                                                                                                                                                                
70 +- PostgreSQL 也不加锁                                                                                                                                                                         
71 +                                                                                                                                                                                              
72 +等价于所有受支持数据库都跳过 `FOR UPDATE`。                                                                                                                                                   
73 +                                                                                                                                                                                              
74 +**影响范围**                                                                                                                                                                                  
75 +                                                                                                                                                                                              
76 +- `service/market_order_service.go`                                                                                                                                                           
77 +- `service/market_payment_service.go`                                                                                                                                                         
78 +- `service/entitlement_service.go`                                                                                                                                                            
79 +- `service/marketplace_entitlement_funding.go`                                                                                                                                                
80 +- `service/inventory_sync_task.go`                                                                                                                                                            
81 +                                                                                                                                                                                              
82 +**影响**                                                                                                                                                                                      
83 +                                                                                                                                                                                              
84 +- 下单冻结库存时并发不安全                                                                                                                                                                    
85 +- webhook 并发下订单状态竞争                                                                                                                                                                  
86 +- entitlement 预冻结/结算/退款可能并发串扰                                                                                                                                                    
87 +- inventory sync 期间状态判断不可靠                                                                                                                                                           
88 +                                                                                                                                                                                              
89 +**修复要求**                                                                                                                                                                                  
90 +                                                                                                                                                                                              
91 +- SQLite: 明确跳过 `FOR UPDATE`                                                                                                                                                               
92 +- MySQL/PostgreSQL: 正常附加 `FOR UPDATE`                                                                                                                                                     
93 +- 不要再使用“支持数据库全部跳过”的写法                                                                                                                                                        
94 +                                                                                                                                                                                              
95 +**建议修复方式**                                                                                                                                                                              
96 +                                                                                                                                                                                              
97 +```go                                                                                                                                                                                         
     98 +func lockForUpdate(query *gorm.DB) *gorm.DB {                                                                                                                                                 
     99 +    if common.UsingSQLite {                                                                                                                                                                   
    100 +        return query                                                                                                                                                                          
    101 +    }                                                                                                                                                                                         
    102 +    return query.Set("gorm:query_option", "FOR UPDATE")                                                                                                                                       
    103 +}                                                                                                                                                                                             
    104 +```                                                                                                                                                                                           
105 +                                                                                                                                                                                              
106 +**验收标准**                                                                                                                                                                                  
107 +                                                                                                                                                                                              
108 +- SQLite 下不报 `FOR UPDATE` 语法错误                                                                                                                                                         
109 +- MySQL/PostgreSQL 下生成的 SQL 确实带 `FOR UPDATE`                                                                                                                                           
110 +- marketplace 相关事务逻辑重新具备最基本的串行保护                                                                                                                                            
111 +                                                                                                                                                                                              
112 +---                                                                                                                                                                                           
113 +                                                                                                                                                                                              
114 +### 2. 修复“支付成功但发权失败”状态无法落库的问题                                                                                                                                             
115 +                                                                                                                                                                                              
116 +**问题文件**                                                                                                                                                                                  
117 +                                                                                                                                                                                              
118 +- `service/market_payment_service.go:194-220`                                                                                                                                                 
119 +                                                                                                                                                                                              
120 +**当前问题**                                                                                                                                                                                  
121 +                                                                                                                                                                                              
122 +当前流程先把订单对象设为：                                                                                                                                                                    
123 +                                                                                                                                                                                              
124 +- `order_status=paid`                                                                                                                                                                         
125 +- `payment_status=paid`                                                                                                                                                                       
126 +- `entitlement_status=created`                                                                                                                                                                
127 +                                                                                                                                                                                              
128 +然后在 `grantEntitlementsForOrderTx` 失败时，尝试把 `entitlement_status` 改成 `failed` 后直接 `return err`。                                                                                  
129 +                                                                                                                                                                                              
130 +但该逻辑仍在同一个事务里，`return err` 会导致：                                                                                                                                               
131 +                                                                                                                                                                                              
132 +- 订单支付成功状态回滚                                                                                                                                                                        
133 +- `entitlement_status=failed` 也一起回滚                                                                                                                                                      
134 +                                                                                                                                                                                              
135 +最终数据库里无法稳定留下设计要求的失败态。                                                                                                                                                    
136 +                                                                                                                                                                                              
137 +**设计要求**                                                                                                                                                                                  
138 +                                                                                                                                                                                              
139 +见 `12-支付订单与权益发放时序设计.md`：                                                                                                                                                       
140 +                                                                                                                                                                                              
141 +- `payment_status=paid`                                                                                                                                                                       
142 +- `order_status=paid`                                                                                                                                                                         
143 +- `entitlement_status=failed`                                                                                                                                                                 
144 +- 后台必须可对单订单项重放                                                                                                                                                                    
145 +                                                                                                                                                                                              
146 +**风险**                                                                                                                                                                                      
147 +                                                                                                                                                                                              
148 +- PSP 已确认成功，但平台仍把订单留在未支付态                                                                                                                                                  
149 +- 后台无法基于真实失败态做重放                                                                                                                                                                
150 +- 补偿链路与客服排障会失真                                                                                                                                                                    
151 +                                                                                                                                                                                              
152 +**修复要求**                                                                                                                                                                                  
153 +                                                                                                                                                                                              
154 +需要把“支付成功状态落库”和“发权失败标记落库”做成可保留的终态，而不是跟着事务整体回滚。                                                                                                        
155 +                                                                                                                                                                                              
156 +可选方向：                                                                                                                                                                                    
157 +                                                                                                                                                                                              
158 +1. 同事务内先持久化 paid 状态，发权失败后改为 `entitlement_status=failed`，最终事务仍提交成功，再把错误作为返回值或告警向上抛出                                                               
159 +2. 将发权逻辑拆成独立步骤，确保 paid 状态可保留，失败时显式落 `failed`                                                                                                                        
160 +                                                                                                                                                                                              
161 +**额外要求**                                                                                                                                                                                  
162 +                                                                                                                                                                                              
163 +- 保留现有 `retryEntitlementGrantOnly` 逻辑                                                                                                                                                   
164 +- 不允许支付重试再次累加库存                                                                                                                                                                  
165 +                                                                                                                                                                                              
166 +**验收标准**                                                                                                                                                                                  
167 +                                                                                                                                                                                              
168 +- 能稳定构造并保留 `paid + entitlement_failed` 状态                                                                                                                                           
169 +- 后续再次调用支付完成逻辑时，仅补做 entitlement grant                                                                                                                                        
170 +- 不重复修改 `InventorySnapshot.SoldAmount`                                                                                                                                                   
171 +- 不重复修改 `SupplyAccount.ReservedCapacity`                                                                                                                                                 
172 +                                                                                                                                                                                              
173 +---                                                                                                                                                                                           
174 +                                                                                                                                                                                              
175 +## 四、P1 修复项                                                                                                                                                                              
176 +                                                                                                                                                                                              
177 +### 3. 成功结算后补写 `market_order_item.used_amount`                                                                                                                                         
178 +                                                                                                                                                                                              
179 +**问题文件**                                                                                                                                                                                  
180 +                                                                                                                                                                                              
181 +- `service/marketplace_entitlement_funding.go:154-202`                                                                                                                                        
182 +- `model/market_order.go:57`                                                                                                                                                                  
183 +                                                                                                                                                                                              
184 +**当前问题**                                                                                                                                                                                  
185 +                                                                                                                                                                                              
186 +当前 settle 逻辑会更新：                                                                                                                                                                      
187 +                                                                                                                                                                                              
188 +- `entitlement_lot.used_amount`                                                                                                                                                               
189 +- `buyer_entitlement.total_used`                                                                                                                                                              
190 +- `supply.used_capacity`                                                                                                                                                                      
191 +- `inventory_snapshot.consumed_amount`                                                                                                                                                        
192 +                                                                                                                                                                                              
193 +但没有更新：                                                                                                                                                                                  
194 +                                                                                                                                                                                              
195 +- `market_order_item.used_amount`                                                                                                                                                             
196 +                                                                                                                                                                                              
197 +**任务书要求**                                                                                                                                                                                
198 +                                                                                                                                                                                              
199 +见 `09-M1_交易闭环开发任务书.md`：                                                                                                                                                            
200 +                                                                                                                                                                                              
201 +- 成功结算后回写 `usage_ledger`                                                                                                                                                               
202 +- 回写 `entitlement_lot.used_amount`                                                                                                                                                          
203 +- 回写 `market_order_item.used_amount`                                                                                                                                                        
204 +                                                                                                                                                                                              
205 +**影响**                                                                                                                                                                                      
206 +                                                                                                                                                                                              
207 +- 订单项维度无法准确反映已消费量                                                                                                                                                              
208 +- 后续 settlement / dispute / admin 对账缺失直接依据                                                                                                                                          
209 +                                                                                                                                                                                              
210 +**修复要求**                                                                                                                                                                                  
211 +                                                                                                                                                                                              
212 +- marketplace entitlement 成功结算时，同步累加对应 `order_item_id` 的 `used_amount`                                                                                                           
213 +- 失败退款时不得错误增加 `used_amount`                                                                                                                                                        
214 +                                                                                                                                                                                              
215 +**验收标准**                                                                                                                                                                                  
216 +                                                                                                                                                                                              
217 +- 单次成功请求后，`market_order_item.used_amount` 正确增加                                                                                                                                    
218 +- retry 成功不会重复累加                                                                                                                                                                      
219 +- failed/refund 流程不会产生脏数据                                                                                                                                                            
220 +                                                                                                                                                                                              
221 +---                                                                                                                                                                                           
222 +                                                                                                                                                                                              
223 +### 4. 补全 `usage_ledger` 的真实 usage 字段                                                                                                                                                  
224 +                                                                                                                                                                                              
225 +**问题文件**                                                                                                                                                                                  
226 +                                                                                                                                                                                              
227 +- `model/usage_ledger.go:21-32`                                                                                                                                                               
228 +- `service/marketplace_entitlement_funding.go:263-346`                                                                                                                                        
229 +                                                                                                                                                                                              
230 +**当前问题**                                                                                                                                                                                  
231 +                                                                                                                                                                                              
232 +`UsageLedger` 模型已经定义了：                                                                                                                                                                
233 +                                                                                                                                                                                              
234 +- `prompt_tokens`                                                                                                                                                                             
235 +- `completion_tokens`                                                                                                                                                                         
236 +- `total_tokens`                                                                                                                                                                              
237 +- `other`                                                                                                                                                                                     
238 +                                                                                                                                                                                              
239 +但当前 marketplace 写账逻辑只写了：                                                                                                                                                           
240 +                                                                                                                                                                                              
241 +- request 归属                                                                                                                                                                                
242 +- billing source                                                                                                                                                                              
243 +- entitlement/seller/order 维度                                                                                                                                                               
244 +- pre-consume / actual quota                                                                                                                                                                  
245 +                                                                                                                                                                                              
246 +没有写 token usage 明细。                                                                                                                                                                     
247 +                                                                                                                                                                                              
248 +**影响**                                                                                                                                                                                      
249 +                                                                                                                                                                                              
250 +- usage ledger 只能证明“扣了多少配额”                                                                                                                                                         
251 +- 不能证明“这次请求真实用了多少 prompt/completion/total tokens”                                                                                                                               
252 +- 审计、争议、结算支撑不足                                                                                                                                                                    
253 +                                                                                                                                                                                              
254 +**修复要求**                                                                                                                                                                                  
255 +                                                                                                                                                                                              
256 +- 在最终结算成功时，把 relay 已知 usage 明细同步写入 `UsageLedger`                                                                                                                            
257 +- 在 failed/refund 场景下，至少明确写入：                                                                                                                                                     
258 +  - `ledger_status`                                                                                                                                                                           
259 +  - `error_code`                                                                                                                                                                              
260 +  - request 基础归属                                                                                                                                                                          
261 +- 如有 `Log.Other` 中的 marketplace 归属信息，也应考虑同步进 `UsageLedger.Other`                                                                                                              
262 +                                                                                                                                                                                              
263 +**验收标准**                                                                                                                                                                                  
264 +                                                                                                                                                                                              
265 +- success ledger 具备 token usage 明细                                                                                                                                                        
266 +- failed ledger 具备失败归属与错误码                                                                                                                                                          
267 +- `usage_ledger` 能独立支撑最小审计追溯                                                                                                                                                       
268 +                                                                                                                                                                                              
269 +---                                                                                                                                                                                           
270 +                                                                                                                                                                                              
271 +## 五、P2 修复项                                                                                                                                                                              
272 +                                                                                                                                                                                              
273 +### 5. 优化 marketplace 列表接口的 N+1 查询                                                                                                                                                   
274 +                                                                                                                                                                                              
275 +**问题文件**                                                                                                                                                                                  
276 +                                                                                                                                                                                              
277 +- `controller/listing.go:43-53`                                                                                                                                                               
278 +- `controller/market_order.go:40-48`                                                                                                                                                          
279 +- `controller/market_order.go:100-109`                                                                                                                                                        
280 +                                                                                                                                                                                              
281 +**当前问题**                                                                                                                                                                                  
282 +                                                                                                                                                                                              
283 +存在典型 N+1：                                                                                                                                                                                
284 +                                                                                                                                                                                              
285 +- 商品列表逐条查 detail                                                                                                                                                                       
286 +- 后台 listing 列表逐条查 SKU                                                                                                                                                                 
287 +- 买家订单列表逐条查 order items                                                                                                                                                              
288 +                                                                                                                                                                                              
289 +**影响**                                                                                                                                                                                      
290 +                                                                                                                                                                                              
291 +- 数据量小时问题不明显                                                                                                                                                                        
292 +- 数据量上来后，后台页和 marketplace 页会明显变慢                                                                                                                                             
293 +                                                                                                                                                                                              
294 +**修复要求**                                                                                                                                                                                  
295 +                                                                                                                                                                                              
296 +- 把 listing -> sku                                                                                                                                                                           
297 +- order -> items                                                                                                                                                                              
298 +- public listing -> detail                                                                                                                                                                    
299 +                                                                                                                                                                                              
300 +尽量改成批量查询后组装，而不是逐条调 service/model。                                                                                                                                          
301 +                                                                                                                                                                                              
302 +**验收标准**                                                                                                                                                                                  
303 +                                                                                                                                                                                              
304 +- 单页 20 条 listing / order 时查询次数明显下降                                                                                                                                               
305 +- controller 层不再按记录循环打 DB                                                                                                                                                            
306 +                                                                                                                                                                                              
307 +---                                                                                                                                                                                           
308 +                                                                                                                                                                                              
309 +## 六、建议修复顺序                                                                                                                                                                           
310 +                                                                                                                                                                                              
311 +建议严格按下面顺序推进：                                                                                                                                                                      
312 +                                                                                                                                                                                              
313 +1. 修复 `lockForUpdate`                                                                                                                                                                       
314 +2. 修复支付成功但发权失败状态保留                                                                                                                                                             
315 +3. 补写 `market_order_item.used_amount`                                                                                                                                                       
316 +4. 补全 `usage_ledger` usage 字段                                                                                                                                                             
317 +5. 优化 N+1 查询                                                                                                                                                                              
318 +                                                                                                                                                                                              
319 +原因：                                                                                                                                                                                        
320 +                                                                                                                                                                                              
321 +- `P0` 不先修，所有并发与补偿行为都不可信                                                                                                                                                     
322 +- `P1` 不补，M1 闭环虽“能跑”，但账务口径不完整                                                                                                                                                
323 +- `P2` 可放最后收口                                                                                                                                                                           
324 +                                                                                                                                                                                              
325 +---                                                                                                                                                                                           
326 +                                                                                                                                                                                              
327 +## 七、建议补充测试                                                                                                                                                                           
328 +                                                                                                                                                                                              
329 +当前已有测试能覆盖一部分 happy path，但仍需补以下用例：                                                                                                                                       
330 +                                                                                                                                                                                              
331 +### 必补测试 A：`lockForUpdate` 行为测试                                                                                                                                                      
332 +                                                                                                                                                                                              
333 +- SQLite 下不附加 `FOR UPDATE`                                                                                                                                                                
334 +- MySQL/PostgreSQL 下附加 `FOR UPDATE`                                                                                                                                                        
335 +                                                                                                                                                                                              
336 +### 必补测试 B：支付成功但发权失败状态保留                                                                                                                                                    
337 +                                                                                                                                                                                              
338 +构造：                                                                                                                                                                                        
339 +                                                                                                                                                                                              
340 +- PSP 回调成功                                                                                                                                                                                
341 +- `grantEntitlementsForOrderTx` 人为返回 error                                                                                                                                                
342 +                                                                                                                                                                                              
343 +断言：                                                                                                                                                                                        
344 +                                                                                                                                                                                              
345 +- `order_status=paid`                                                                                                                                                                         
346 +- `payment_status=paid`                                                                                                                                                                       
347 +- `entitlement_status=failed`                                                                                                                                                                 
348 +                                                                                                                                                                                              
349 +### 必补测试 C：marketplace settle 回写 `order_item.used_amount`                                                                                                                              
350 +                                                                                                                                                                                              
351 +断言：                                                                                                                                                                                        
352 +                                                                                                                                                                                              
353 +- success 后 `order_item.used_amount` 正确增加                                                                                                                                                
354 +- retry 不重复增加                                                                                                                                                                            
355 +                                                                                                                                                                                              
356 +### 必补测试 D：usage ledger 写入完整 usage                                                                                                                                                   
357 +                                                                                                                                                                                              
358 +断言 success ledger 至少包含：                                                                                                                                                                
359 +                                                                                                                                                                                              
360 +- `prompt_tokens`                                                                                                                                                                             
361 +- `completion_tokens`                                                                                                                                                                         
362 +- `total_tokens`                                                                                                                                                                              
363 +- `billing_source`                                                                                                                                                                            
364 +- `entitlement_lot_id`                                                                                                                                                                        
365 +                                                                                                                                                                                              
366 +---                                                                                                                                                                                           
367 +                                                                                                                                                                                              
368 +## 八、回归验证命令                                                                                                                                                                           
369 +                                                                                                                                                                                              
370 +建议每个修复批次后至少执行：                                                                                                                                                                  
371 +                                                                                                                                                                                              
372 +```bash                                                                                                                                                                                       
    373 +go test ./service -run 'TestNewBillingSession|TestMarketPayment|TestMarketplace' -count=1                                                                                                     
    374 +```                                                                                                                                                                                           
375 +                                                                                                                                                                                              
376 +完成全部修复后执行：                                                                                                                                                                          
377 +                                                                                                                                                                                              
378 +```bash                                                                                                                                                                                       
    379 +go test ./service ./controller ./model -count=1                                                                                                                                               
    380 +go test ./...                                                                                                                                                                                 
    381 +```                                                                                                                                                                                           
382 +                                                                                                                                                                                              
383 +如有条件，建议补一次：                                                                                                                                                                        
384 +                                                                                                                                                                                              
385 +- SQLite 跑全量 marketplace 测试                                                                                                                                                              
386 +- MySQL 或 PostgreSQL 跑一次关键事务测试                                                                                                                                                      
387 +                                                                                                                                                                                              
388 +---                                                                                                                                                                                           
389 +                                                                                                                                                                                              
390 +## 九、完成标志                                                                                                                                                                               
391 +                                                                                                                                                                                              
392 +当以下条件同时满足时，才建议将 M1 视为“可收口”：                                                                                                                                              
393 +                                                                                                                                                                                              
394 +1. `P0` 两项全部关闭                                                                                                                                                                          
395 +2. `P1` 两项全部关闭                                                                                                                                                                          
396 +3. marketplace entitlement 请求能稳定写出完整 usage 账                                                                                                                                        
397 +4. 支付成功但发权失败可稳定保留失败态并支持重放                                                                                                                                               
398 +5. 不再存在“代码说已完成，但账务状态落不住”的问题                                                                                                                                             
399 +                                                                                                                                                                                              
400 +---                                                                                                                                                                                           
401 +                                                                                                                                                                                              
402 +## 十、附注                                                                                                                                                                                   
403 +                                                                                                                                                                                              
404 +本清单是基于 2026-04-16 对当前 `dev` 分支代码的再次全面 review 收敛出的执行清单。                                                                                                             
405 +                                                                                                                                                                                              
406 +它不是新的需求文档，而是：                                                                                                                                                                    
407 +                                                                                                                                                                                              
408 +- 对 M1 当前剩余问题的收口列表                                                                                                                                                                
409 +- 对修复先后顺序的明确约束                                                                                                                                                                    
410 +- 对“何时可以继续宣布 M1 完成”的判断门槛
