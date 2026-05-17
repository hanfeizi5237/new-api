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

export const codexCliGuide: UsageGuide = {
  id: 'codex-cli',
  title: 'Codex CLI',
  shortTitle: 'Codex',
  description: 'OpenAI 的终端编码代理，适合补丁式修改、命令执行和计划追踪。',
  summary:
    'Codex CLI 更像一个对仓库直接动手的编码代理。第一版接入建议是先按官方要求装好 Node.js 和 CLI，再把配置切到 CCToken。Windows、macOS、Linux 都可以按各自终端环境完成接入。',
  officialUrl: 'https://github.com/openai/codex',
  tags: ['终端编码', '补丁编辑', '计划追踪'],
  recommendedFor: [
    '需要在本地仓库里直接编辑代码、运行命令和维护计划',
    '习惯命令行，不依赖图形界面',
    '希望通过 CCToken 承接 Codex 流量与密钥管理',
  ],
  prerequisites: [
    '已安装 Node.js 22 或当前工具要求的版本',
    '已能在终端执行 `npm` 或同等包管理命令',
    '准备好 CCToken 域名、令牌和想使用的模型',
  ],
  steps: [
    {
      title: '安装 Codex CLI',
      description:
        'Windows 环境推荐在 WSL2 中使用；macOS 与 Linux 可以直接通过 npm 全局安装。装完先确认 `codex --version` 可用。',
      code: 'npm install -g @openai/codex',
    },
    {
      title: '把 Codex 配置切到 CCToken',
      description:
        '可以手工修改配置，也可以沉淀团队辅助脚本。目标是让 Codex 的所有模型请求都走 CCToken 接入点，而不是默认官方账户。',
    },
    {
      title: '在项目目录中启动并校验权限策略',
      description:
        '进入一个真实项目后运行 `codex`，检查默认模型、沙箱策略和审批方式是否符合你的团队习惯。',
      code: 'cd /path/to/your/project\ncodex',
    },
    {
      title: '用一个小任务验证补丁工作流',
      description:
        '让 Codex 做一次轻量修改，例如补一个小注释或改一条文案。这样最容易确认模型连通性、文件写入权限和命令执行都已打通。',
    },
  ],
  verification: [
    '`codex --version` 正常输出',
    '在项目目录里可以正常启动交互会话',
    '执行一个小任务后，补丁和命令结果都能正常返回',
  ],
  troubleshooting: [
    {
      title: 'Windows 下体验不稳定',
      content:
        'Windows 环境建议优先在 WSL2 中运行，这样对文件系统权限、Node 环境和终端兼容性会更稳。',
    },
    {
      title: '改了接入点但看起来没生效',
      content:
        '先回头检查 Codex 当前使用的是哪份配置文件，再确认模型请求是否已经指向团队的 CCToken 域名。',
    },
  ],
}
