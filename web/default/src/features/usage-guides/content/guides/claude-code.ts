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

export const claudeCodeGuide: UsageGuide = {
  id: 'claude-code',
  title: 'Claude Code',
  shortTitle: 'Claude',
  description: 'Anthropic 的终端编程助手，可通过 Anthropic 兼容入口接入 CCToken。',
  summary:
    'Claude Code 的核心接入方式是安装官方 CLI，然后把它的请求基地址与鉴权令牌改指向 CCToken 暴露出来的 Anthropic 兼容入口。不同系统可以按各自终端环境保存变量或接入团队脚本。',
  officialUrl: 'https://www.anthropic.com/claude-code',
  tags: ['终端编码', 'Anthropic 兼容', '多文件编辑'],
  recommendedFor: [
    '需要深度代码库理解与多文件编辑',
    '偏好终端工作流，并希望直接在项目目录里运行 AI 助手',
    '已经在用 Claude 官方 CLI，希望把流量改走 CCToken',
  ],
  prerequisites: [
    '已安装 Claude Code CLI',
    '本机 PATH 已正确包含 Claude 可执行文件',
    '已准备好 CCToken 的兼容接入地址和令牌',
  ],
  steps: [
    {
      title: '先安装 Claude Code CLI',
      description:
        'Windows 环境建议准备 Node.js 和 Git Bash；macOS 与 Linux 可以按官方 CLI 安装方式处理。安装完成后先确认 `claude --version` 能正常输出。',
    },
    {
      title: '设置指向 CCToken 的环境变量',
      description:
        '把 Claude Code 的基地址和鉴权令牌改成你的 CCToken 配置。实际地址请使用你的部署所提供的 Anthropic 兼容入口。',
      code: 'export ANTHROPIC_BASE_URL="https://api.cctoken.fun"\nexport ANTHROPIC_AUTH_TOKEN="sk-your-cctoken-key"',
    },
    {
      title: '启动后选择可用模型',
      description:
        '进入 `claude` 后可以执行 `/model` 检查模型选择器。通常默认模型即可，但前提是 CCToken 侧已经向该令牌开放了对应模型。',
    },
    {
      title: '必要时使用团队一键脚本',
      description:
        '如果团队更关心统一落地速度，而不是逐台手配变量，可以把 PowerShell 或 shell 辅助脚本固化成内部标准流程。',
    },
  ],
  verification: [
    '`claude --version` 正常输出',
    '`echo $ANTHROPIC_BASE_URL` 或系统等价命令能看到已生效配置',
    '进入 `claude` 后执行一个简单代码任务，可以拿到正常回答',
  ],
  troubleshooting: [
    {
      title: '明明能启动，但还是走官方额度',
      content:
        '最常见原因是 `ANTHROPIC_BASE_URL` 没有在当前终端会话生效。重新打开终端，或手动 `source` 对应配置文件后再试一次。',
    },
    {
      title: 'Windows 下安装不顺',
      content:
        'Windows 环境更适合用 PowerShell 管理运行流程，用 Git Bash 处理安装阶段。',
    },
  ],
}
