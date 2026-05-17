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
  description: '面向业务与运营的模型价格说明页，用于快速对比模型成本结构。',
  summary:
    '定价信息按输入、输出、缓存读三部分展示，适合做模型选型、路由分层和预算评估。价格高低并不直接等于性价比，建议结合任务复杂度与输出占比一起评估。',
  tags: ['价格', '用量', '预算'],
  recommendedFor: [
    '需要按模型做月度成本预算的团队',
    '需要比较不同模型输入/输出/缓存价格差异的产品与运营',
    '需要做模型替换或分层路由成本评估的工程团队',
  ],
  prerequisites: [
    '明确你的主模型与备选模型',
    '能统计业务的输入 Token、输出 Token 与缓存读 Token 用量',
    '上线前确认本页价格与当前执行价格基准一致',
  ],
  pricing: {
    unit: '单位：¥ / 百万 Token',
    rows: [
      {
        model: 'GPT-5.5',
        input: '¥17.50',
        output: '¥105.00',
        cacheRead: '¥1.75',
      },
      {
        model: 'GPT-5.4',
        input: '¥8.75',
        output: '¥52.50',
        cacheRead: '¥0.88',
      },
      {
        model: 'GPT-5.4-mini',
        input: '¥2.63',
        output: '¥15.75',
        cacheRead: '¥0.26',
      },
      {
        model: 'GPT-5.3-Codex',
        input: '¥6.13',
        output: '¥49.00',
        cacheRead: '¥0.61',
      },
      {
        model: 'Claude Opus 4.7',
        input: '¥17.50',
        output: '¥87.50',
        cacheRead: '¥1.75',
      },
      {
        model: 'Claude Opus 4.6',
        input: '¥17.50',
        output: '¥87.50',
        cacheRead: '¥1.75',
      },
      {
        model: 'Claude Sonnet 4.6',
        input: '¥10.50',
        output: '¥52.50',
        cacheRead: '¥1.05',
      },
      {
        model: 'DeepSeek V4 Flash',
        input: '¥0.02',
        output: '¥1.00',
        cacheRead: '¥2.00',
      },
      {
        model: 'DeepSeek V4 Pro',
        input: '¥0.025',
        output: '¥3.00',
        cacheRead: '¥6.00',
      },
    ],
    notes: [
      '同一模型的输入、输出和缓存读价格可能体现不同的成本结构。',
      '高复杂度任务通常对输出价格更敏感，长上下文任务通常对输入和缓存读更敏感。',
      '建议按业务类型分层路由模型，避免所有请求都落到高单价模型。',
    ],
  },
  steps: [
    {
      title: '成本核算公式',
      description: '按输入、输出、缓存读三段计费，再合并为总成本。',
      code: '输入成本 = 输入 Token 数 × 输入单价\n输出成本 = 输出 Token 数 × 输出单价\n缓存读成本 = 缓存读 Token 数 × 缓存读单价\n总成本 = 输入成本 + 输出成本 + 缓存读成本',
    },
    {
      title: '预算建议',
      description:
        '预算评估时建议分别看模型单价、输出占比和缓存命中率，避免只按输入价格判断。',
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
        '请按最新价格基准同步更新本页，避免预算依据滞后。',
    },
  ],
}
