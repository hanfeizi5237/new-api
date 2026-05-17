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
  description: '按最新价格图查看模型输入、输出与缓存读价格，并据此估算成本。',
  summary:
    '本页已按你提供的最新价格截图重新整理，价格字段包含输入、输出、缓存读三列，以下表格即当前基准。',
  tags: ['价格', '用量', '缓存读'],
  recommendedFor: [
    '需要按模型做月度成本预算的团队',
    '需要比较不同模型输入/输出/缓存价格差异的产品与运营',
    '需要做模型替换或分层路由成本评估的工程团队',
  ],
  prerequisites: [
    '明确你的主模型与备选模型',
    '能统计业务的输入 Token、输出 Token 与缓存读 Token 用量',
    '上线前确认本页价格与最新价格截图一致',
  ],
  steps: [
    {
      title: '按价格表核对模型单价',
      description:
        '以下数据来自你提供的最新价格截图：\n\n| 模型 | 输入 | 输出 | 缓存读 |\n| --- | ---: | ---: | ---: |\n| GPT-5.5 | ¥17.50 | ¥105.00 | ¥1.75 |\n| GPT-5.4 | ¥8.75 | ¥52.50 | ¥0.88 |\n| GPT-5.4-mini | ¥2.63 | ¥15.75 | ¥0.26 |\n| GPT-5.3-Codex | ¥6.13 | ¥49.00 | ¥0.61 |\n| Claude Opus 4.7 | ¥17.50 | ¥87.50 | ¥1.75 |\n| Claude Opus 4.6 | ¥17.50 | ¥87.50 | ¥1.75 |\n| Claude Sonnet 4.6 | ¥10.50 | ¥52.50 | ¥1.05 |\n| DeepSeek V4 Flash | ¥0.02 | ¥1.00 | ¥2.00 |\n| DeepSeek V4 Pro | ¥0.025 | ¥3.00 | ¥6.00 |',
    },
    {
      title: '按输入/输出/缓存读拆分估算成本',
      description:
        '成本核算按输入、输出、缓存读三部分分别计算，然后相加得到总成本。',
      code: '输入成本 = 输入 Token 数 × 输入单价\n输出成本 = 输出 Token 数 × 输出单价\n缓存读成本 = 缓存读 Token 数 × 缓存读单价\n总成本 = 输入成本 + 输出成本 + 缓存读成本',
    },
    {
      title: '建立模型分层策略',
      description:
        '建议把高单价模型用于复杂任务，把轻量模型用于常规任务，降低整体平均成本。每次调整路由后，都应复盘输入/输出结构是否发生变化。',
      note: '如价格表更新，请以最新 Excel 为准并同步更新本页。',
    },
  ],
  verification: [
    '页面模型列表与最新价格截图一致（9个）',
    '每个模型都已包含输入、输出、缓存读三项价格',
    '成本估算包含输入/输出/缓存读三部分',
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
        '请先更新最新价格截图基准，然后同步更新本页价格表，避免旧数据继续被引用。',
    },
  ],
}
