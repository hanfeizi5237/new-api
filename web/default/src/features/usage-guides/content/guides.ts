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
export const guideIds = [
  'cc-switch',
  'cherry-studio',
  'openclaw',
  'claude-code',
  'codex-cli',
] as const

export type GuideId = (typeof guideIds)[number]

export type GuideStep = {
  title: string
  description: string
  code?: string
  note?: string
}

export type GuideTroubleshooting = {
  title: string
  content: string
}

export type UsageGuide = {
  id: GuideId
  title: string
  shortTitle: string
  description: string
  summary: string
  officialUrl?: string
  tags: string[]
  recommendedFor: string[]
  prerequisites: string[]
  steps: GuideStep[]
  verification: string[]
  troubleshooting: GuideTroubleshooting[]
}

export const usageGuides: UsageGuide[] = [
  {
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
  },
  {
    id: 'cherry-studio',
    title: 'Cherry Studio',
    shortTitle: 'Cherry',
    description: '桌面 AI 客户端，适合多模型对话和图像生成的日常使用。',
    summary:
      'Cherry Studio 的接入路径很直白：创建提供商、填入 CCToken 地址和密钥、补充模型，然后回到聊天界面切换即可。它也支持通过聊天设置做一键填充。',
    officialUrl: 'https://cherry-ai.com',
    tags: ['桌面客户端', '对话', '画图'],
    recommendedFor: [
      '希望用桌面客户端统一管理多模型会话',
      '既做文本问答，也需要图像模型入口',
      '偏好图形界面，不想长期停留在命令行',
    ],
    prerequisites: [
      '已安装 Cherry Studio',
      'CCToken 中有一个可复制的 API Key',
      '知道自己的站点地址以及要开放给客户端的模型名',
    ],
    steps: [
      {
        title: '可选：在聊天设置里配置一键填充',
        description:
          '如果希望从 CCToken 令牌页直接唤起 Cherry Studio，可在聊天设置中添加一个快捷入口。',
        code: '{ "Cherry Studio": "cherrystudio://providers/api-keys?v=1&data={cherryConfig}" }',
      },
      {
        title: '在 Cherry Studio 中创建提供商',
        description:
          '新增一个自定义提供商，将 API 地址指向你的 CCToken 站点，并填入刚刚复制的密钥。',
        code: '站点地址: https://www.cctoken.fun/\nAPI 地址: https://www.cctoken.fun/v1\nAPI 密钥: sk-xxx',
      },
      {
        title: '添加要使用的模型',
        description:
          '在模型管理页把需要的聊天模型补齐。若你还要做图像生成，再额外添加支持绘图的模型。',
      },
      {
        title: '切回聊天页选择 CCToken 模型',
        description:
          '保存设置后返回对话页，在模型切换器中选择刚接入的模型。如果要画图，切换到支持图像生成的模型再开始。',
      },
    ],
    verification: [
      '聊天页可以正常看到并切换到 CCToken 模型',
      '发送一轮对话后能收到正常响应',
      '如果配置了图像模型，绘图入口可以正常调用',
    ],
    troubleshooting: [
      {
        title: '已经填了 Key 但无法响应',
        content:
          '优先检查 API 地址是否补上了兼容接口路径，以及模型名是否与 CCToken 实际暴露出来的名称一致。',
      },
      {
        title: '图像模型可见但调用失败',
        content:
          '通常是该模型本身不支持图像生成，或当前令牌没有对应能力。把聊天模型和绘图模型分开配置会更稳妥。',
      },
    ],
  },
  {
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
  },
  {
    id: 'claude-code',
    title: 'Claude Code',
    shortTitle: 'Claude',
    description:
      'Anthropic 的终端编程助手，可通过 Anthropic 兼容入口接入 CCToken。',
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
        code: 'export ANTHROPIC_BASE_URL="https://www.cctoken.fun/"\nexport ANTHROPIC_AUTH_TOKEN="sk-your-cctoken-key"',
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
  },
  {
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
  },
]

export function isGuideId(value: string | undefined): value is GuideId {
  if (!value) return false
  return guideIds.includes(value as GuideId)
}

export function getGuideById(guideId: string | undefined): UsageGuide {
  if (!isGuideId(guideId)) {
    return usageGuides[0]
  }
  return usageGuides.find((guide) => guide.id === guideId) ?? usageGuides[0]
}
