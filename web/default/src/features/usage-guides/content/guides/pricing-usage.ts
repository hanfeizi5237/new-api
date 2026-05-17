/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import type { UsageGuide } from '../types'

export const pricingUsageGuide: UsageGuide = {
  id: 'pricing-usage',
  title: '定价与用量',
  shortTitle: '定价',
  description: '按最新价格表查看模型输入/输出价格，并据此估算实际用量成本。',
  summary:
    '本页已按你提供的 `/Users/hanfei/Downloads/价格表.xlsx` 重新整理。当前只保留价格表中的模型输入/输出单价，不再展示旧版缓存或 TTS 计费内容。',
  tags: ['价格', '用量', '模型单价'],
  recommendedFor: [
    '需要按模型做月度成本预算的团队',
    '需要比较不同模型输入/输出价格差异的产品与运营',
    '需要做模型替换或分层路由成本评估的工程团队',
  ],
  prerequisites: [
    '明确你的主模型与备选模型',
    '能统计业务的输入 Token 与输出 Token 用量',
    '上线前确认本页价格与最新内部价格表一致',
  ],
  steps: [
    {
      title: '按价格表核对模型单价',
      description:
        '以下数据来自你提供的价格表文件（`/Users/hanfei/Downloads/价格表.xlsx`）：\n\n| 模型 | 输入 | 输出 |\n| --- | ---: | ---: |\n| GPT-5.4 | 2.5 | 15 |\n| GPT-5.5 | 5 | 30 |\n| GPT-5.5 Instant | 5 | 30 |\n| GPT-5.4 Pro | 30 | 180 |\n| GPT-5.2 | 1.75 | 14 |\n| GPT-5.4 Mini | 0.75 | 4.5 |\n| GPT-5 Mini | 0.25 | 2 |\n| GPT-5.4 Nano | 0.2 | 1.25 |\n| GPT-5 Nano | 0.05 | 0.4 |\n| Claude Opus 4.7 | 5 | 25 |\n| Claude Opus 4.6 | 5 | 25 |\n| Claude Opus 4.5 | 5 | 25 |\n| Claude Sonnet 4.6 | 3 | 15 |\n| Claude Sonnet 4.5 | 3 | 15 |\n| Claude Haiku 4.5 | 1 | 5 |\n| Gemini 3.1 Pro | 2 | 12 |\n| Gemini 3 Flash | 0.5 | 3 |\n| DeepSeek V3.2 | 0.29 | 0.44 |\n| DeepSeek V4 Flash | 0.28 | 0.56 |\n| DeepSeek V4 Pro | 0.87 | 1.74 |\n| GLM-5.1 | 1.4 | 4.4 |\n| GLM-5 | 1 | 3.2 |\n| MiniMax M2.7 | 0.3 | 1.2 |\n| MiniMax M2.5 | 0.3 | 1.2 |\n| Kimi K2.6 | 0.95 | 4 |\n| Kimi K2.5 | 0.59 | 3 |',
    },
    {
      title: '按输入/输出拆分估算成本',
      description:
        '成本核算按输入和输出两部分分别计算，然后相加得到总成本。建议把线上真实 Token 占比带入计算，而不是只看单价。',
      code: '输入成本 = 输入 Token 数 × 输入单价\n输出成本 = 输出 Token 数 × 输出单价\n总成本 = 输入成本 + 输出成本',
    },
    {
      title: '建立模型分层策略',
      description:
        '建议把高单价模型用于复杂任务，把轻量模型用于常规任务，降低整体平均成本。每次调整路由后，都应复盘输入/输出结构是否发生变化。',
      note: '如价格表更新，请以最新 Excel 为准并同步更新本页。',
    },
  ],
  verification: [
    '页面模型列表与价格表.xlsx中的模型数量一致（26个）',
    '每个模型都已包含输入和输出价格',
    '成本估算仅按输入/输出两项计算，不再混入旧版规则',
  ],
  troubleshooting: [
    {
      title: '预算与实际仍有偏差',
      content:
        '优先检查实际输出 Token 占比是否高于预估值，输出占比变化通常会显著拉高总成本。',
    },
    {
      title: '模型变更后价格没更新',
      content:
        '请先更新 `/Users/hanfei/Downloads/价格表.xlsx`，然后同步更新本页价格表，避免旧数据继续被引用。',
    },
  ],
}
