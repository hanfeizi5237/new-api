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
  description: '查看模型价格、缓存计费规则与 TTS 语音计费口径。',
  summary:
    '本栏目用于集中查看常见模型和语音能力的计费方式。计费口径按输入、输出和缓存命中分别统计，建议在上线前先核对模型单价与缓存策略，再做成本预估。\n\n以下价格表按公开规则整理为可读版本，便于在项目内快速检索。',
  officialUrl: 'https://docs.b.ai/zh-Hans/llmservice/pricing-and-usage/',
  tags: ['价格', '用量', '缓存计费', 'TTS'],
  recommendedFor: [
    '需要估算项目月度 Token 成本的团队',
    '希望优化输入缓存命中率的应用负责人',
    '需要同时评估 LLM 与 TTS 成本的产品与运营',
  ],
  prerequisites: [
    '已经明确业务主模型与备选模型',
    '知道系统中输入、输出与缓存的调用比例',
    '有基础的 Token 或字符用量统计能力',
  ],
  steps: [
    {
      title: '先确认 LLM 计费维度',
      description:
        'LLM 价格通常拆分为输入、输出和缓存命中三类。先按模型对齐这三类单价，再结合你的调用结构估算总成本。\n\n| 模型 | 每百万输入 Token (¥) | 每百万输出 Token (¥) | 每百万缓存命中 Token (¥) |\n| --- | ---: | ---: | ---: |\n| B.AI-DV-1T | 0.03 | 0.15 | 0.0000015 |\n| B.AI-4.5v | 0.03 | 0.15 | 0.0000015 |\n| B.AI-4.5V-preview | 0.03 | 0.15 | 0.0000015 |\n| B.AI-4o | 0.005 | 0.02 | 0.0000005 |\n| B.AI-4o-mini | 0.0012 | 0.0048 | 0.00000012 |\n| B.AI-4.1 | 0.01 | 0.04 | 0.0000005 |\n| B.AI-4.1-mini | 0.002 | 0.008 | 0.0000001 |\n| B.AI-4.1-nano | 0.0004 | 0.0016 | 0.00000002 |',
    },
    {
      title: '单独核对 o 系列特价模型',
      description:
        'o 系列模型建议独立评估，因为不同任务下输出比例差异更大，直接影响实际成本。\n\n| 模型 | 每百万输入 Token (¥) | 每百万输出 Token (¥) | 每百万缓存命中 Token (¥) |\n| --- | ---: | ---: | ---: |\n| B.AI-o1 | 0.06 | 0.24 | 0.000003 |\n| B.AI-o3-mini | 0.0088 | 0.0352 | 0.00000055 |\n| B.AI-o3 | 0.02 | 0.08 | 0.00000125 |\n| B.AI-o4-mini | 0.0044 | 0.0176 | 0.000000275 |',
    },
    {
      title: '按缓存规则计算提示词成本',
      description:
        '缓存命中部分按缓存价格计算，未命中部分按输入价格计算，输出部分仍按输出价格计算。缓存通常以短时窗口自动复用，命中率越高，输入成本越低。',
      code: '命中成本 = 命中 Token 数 × 缓存命中单价\n未命中输入成本 = 未命中 Token 数 × 输入单价\n输出成本 = 输出 Token 数 × 输出单价\n总成本 = 命中成本 + 未命中输入成本 + 输出成本',
    },
    {
      title: '补充评估 TTS 语音价格',
      description:
        'TTS 常见按输入字符数计费。普通音色与特色音色的模型计费口径不同，建议单独列预算。\n\n普通音色：\n\n| 模型 | 每百万输入字符价格 (¥) |\n| --- | ---: |\n| B.AI-TTS-1 | 0.11 |\n| B.AI-TTS-1-HD | 0.22 |\n\n特色音色：\n\n| 音色模型名称 | 每千字输入价格 (¥) |\n| --- | ---: |\n| alloy | 0.005 |\n| ash | 0.005 |\n| ballad | 0.005 |\n| coral | 0.005 |\n| echo | 0.005 |\n| fable | 0.005 |\n| nova | 0.005 |\n| onyx | 0.005 |\n| sage | 0.005 |\n| shimmer | 0.005 |\n| verse | 0.005 |',
      note: '价格策略可能随时间调整，正式上线前请再做一次价格核对。',
    },
  ],
  verification: [
    '每个上线模型都能对齐输入/输出/缓存三种单价',
    '预算表中单独列出了 o 系列与非 o 系列模型成本',
    '语音场景已拆分普通音色与特色音色的字符计费',
  ],
  troubleshooting: [
    {
      title: '预算和实际费用差距较大',
      content:
        '优先检查缓存命中率和输出 Token 占比是否与估算一致，这两项通常是偏差的主要来源。',
    },
    {
      title: '多模型切换后成本不可控',
      content:
        '建议为不同模型组设置独立配额和成本看板，并给高单价模型增加调用阈值或审批策略。',
    },
  ],
}
