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

export const cherryStudioGuide: UsageGuide = {
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
}
