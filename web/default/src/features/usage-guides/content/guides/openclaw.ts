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

export const openclawGuide: UsageGuide = {
  id: 'openclaw',
  title: 'OpenClaw',
  shortTitle: 'OpenClaw',
  description: '自托管 AI 助手平台，可把多种消息渠道接到自己的代理上。',
  summary:
    'OpenClaw 的接入重点是把 CCToken 声明成一个 `models.providers` 里的自定义 provider，再让默认代理模型指向 `cctoken/<模型ID>`。推荐用环境变量保存密钥。',
  officialUrl: 'https://openclaw.ai',
  tags: ['自托管', '消息渠道', 'Agent'],
  recommendedFor: [
    '需要把 Telegram、Discord、WhatsApp 等渠道接到自己的 AI 代理',
    '希望长期运行、保留本地状态和会话记忆',
    '需要多代理协同或计划任务能力',
  ],
  prerequisites: [
    'Node.js 22 或更高版本',
    'OpenClaw Gateway 与 Control UI 已按官方流程跑通',
    '已准备好 CCToken 地址和 API Key',
  ],
  steps: [
    {
      title: '先把密钥放进环境变量',
      description:
        '把密钥保存在 shell、服务环境或 `.env` 中，而不是直接写进配置文件。',
      code: 'export CCTOKEN_API_KEY="sk-your-cctoken-key"',
    },
    {
      title: '在 `models.providers` 中声明 `cctoken` provider',
      description:
        '将 `baseUrl` 指向你的 CCToken 地址，并确保包含 `/v1`；接口类型使用 OpenAI 兼容的 completions 方式即可。',
      code: '"cctoken": {\n  "baseUrl": "https://www.cctoken.fun/v1",\n  "apiKey": "${CCTOKEN_API_KEY}",\n  "api": "openai-completions"\n}',
    },
    {
      title: '把需要的模型列进 provider',
      description:
        '在 provider 下声明模型 ID 和展示名称，保持模型 ID 与 CCToken 侧暴露出来的一致。',
    },
    {
      title: '切换默认代理模型',
      description:
        '在 `agents.defaults.model.primary` 中使用 `cctoken/<模型ID>`，必要时再给出一个备用模型列表。',
      code: '"primary": "cctoken/gemini-2.5-flash"\n"fallbacks": ["cctoken/kimi-k2.5"]',
    },
  ],
  verification: [
    'OpenClaw 启动后没有 provider 解析报错',
    '代理默认模型已经显示为 `cctoken/...`',
    '从任一接入渠道发消息时，能收到经 CCToken 转发的响应',
  ],
  troubleshooting: [
    {
      title: '请求发出但始终没有响应',
      content:
        '先确认 `baseUrl` 是否带有 `/v1`，再检查模型 ID 是否与控制台里的实际名称一致。这两处是最常见的失配点。',
    },
    {
      title: '想自定义配置目录',
      content:
        'OpenClaw 支持通过 `OPENCLAW_HOME`、`OPENCLAW_STATE_DIR`、`OPENCLAW_CONFIG_PATH` 调整配置和状态文件位置。',
    },
  ],
}
