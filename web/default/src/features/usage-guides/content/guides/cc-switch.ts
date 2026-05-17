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

export const ccSwitchGuide: UsageGuide = {
  id: 'cc-switch',
  title: 'CC Switch',
  shortTitle: 'CC Switch',
  description: '统一管理 Claude、Codex 与 Gemini CLI 的 Provider 配置。',
  summary:
    '适合同时使用多个终端 AI 助手的团队或个人。第一版接入重点不是手工填一堆环境变量，而是利用 CCToken 控制台里的快捷入口，把 Provider 和模型组合一次性导入到 CC Switch。',
  officialUrl: 'https://github.com/farion1231/cc-switch',
  tags: ['CLI 管理', 'Deep Link', '多模型切换'],
  recommendedFor: [
    '需要在 Claude Code、Codex CLI、Gemini CLI 之间频繁切换',
    '希望统一管理 MCP、Prompt 和模型档位',
    '需要桌面版、Web 版或命令行版都可用的入口',
  ],
  prerequisites: [
    '已安装 CC Switch 桌面版、Web 版或 CLI 版',
    'CCToken 中已经创建好可用令牌',
    '准备好主模型和各档位模型的候选列表',
  ],
  steps: [
    {
      title: '在 CCToken 聊天设置里启用快捷入口',
      description:
        '为了在令牌管理页看到一键导入按钮，可以先把 CC Switch 的快捷选项加入系统设置中的聊天设置。',
      code: '{ "CC Switch": "ccswitch" }',
      note: '配置完成后，令牌管理页的下拉菜单会出现 CC Switch 选项。',
    },
    {
      title: '从令牌管理页发起一键导入',
      description:
        '打开目标令牌的操作菜单，选择 `CC Switch`。系统会通过 `ccswitch://` 协议唤起本地应用，并打开配置弹窗。',
    },
    {
      title: '补齐应用类型和模型档位',
      description:
        '在弹窗中选择目标应用类型，例如 Claude、Codex 或 Gemini；再填写配置名称、主模型以及可选的 Haiku / Sonnet / Opus 档位模型。',
    },
    {
      title: '确认导入并切换到目标 CLI',
      description:
        '点击打开 CC Switch 后，新的 Provider 配置会进入 CC Switch。此时就可以在同一面板里切换不同 AI 编程助手共用的模型来源。',
    },
  ],
  verification: [
    'CC Switch 中能看到新建的 Provider 名称和目标应用类型',
    '主模型与分档模型都能正常出现在下拉列表中',
    '切换到对应 CLI 后，请求会通过 CCToken 发出',
  ],
  troubleshooting: [
    {
      title: '点击后没有唤起应用',
      content:
        '先确认本机已经安装支持 `ccswitch://` 的版本。若在远程机器或无头环境中使用，请改走 CC Switch Web 版，再手工导入配置。',
    },
    {
      title: '模型列表不完整',
      content:
        '通常是当前令牌可见模型不足，或 CCToken 侧尚未开放对应模型。先回到令牌权限与模型可见性设置中检查。',
    },
  ],
}
