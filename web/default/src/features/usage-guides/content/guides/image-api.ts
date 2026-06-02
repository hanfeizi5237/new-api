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

export const imageApiGuide: UsageGuide = {
  id: 'image-api',
  title: '生图API',
  shortTitle: '生图API',
  description:
    '使用兼容 OpenAI Images API 的方式，通过 CCToken 发起图片生成请求。',
  summary:
    '这篇手册适合前端页面、自动化脚本、AI 工作流和第三方兼容客户端接入。核心链路很简单：向 `https://www.cctoken.fun/v1/images/generations` 发送 JSON，请求里带上模型、提示词和尺寸即可。接口兼容 OpenAI Images API 的主流调用方式，返回结果可能是图片 URL，也可能是 Base64 数据。',
  tags: ['OpenAI 兼容', 'Images API', '图片生成'],
  recommendedFor: [
    '需要在网页或服务端直接生成图片的开发者',
    '需要把生图能力接入自动化脚本、Agent 或工作流平台的团队',
    '已经习惯 OpenAI 接口风格，希望低成本迁移到 CCToken 的调用方',
  ],
  prerequisites: [
    '已在 CCToken 中创建可用的 API Key，并确认账号具备生图模型权限',
    '调用端支持发送 JSON 请求，并能正确处理较长的图片生成等待时间',
    '上线前已确认要使用的尺寸档位，避免因 fallback 导致结果或消耗不符合预期',
  ],
  steps: [
    {
      title: '快速开始：先打通最小请求',
      description:
        '默认入口是 `POST https://www.cctoken.fun/v1/images/generations`。最小请求只需要 `model`、`prompt` 和 `size` 三个字段。为了让图片质量和指令遵循更稳定，建议优先使用英文 prompt；如果 `size` 留空、填 `auto` 或传入非法值，系统会按 2K 档位处理。',
      code: `{
  "model": "gpt-image-2",
  "prompt": "A cute orange cat wearing an astronaut helmet, sticker style, clean background.",
  "size": "2048x2048"
}`,
      note: '如果你的客户端已经兼容 OpenAI Images API，通常只需要替换 Base URL 和 API Key 即可接入。',
    },
    {
      title: '认证方式：使用 Bearer Token',
      description:
        '所有请求都需要在请求头中携带 Bearer Token，同时将请求体声明为 JSON。最常见的调用错误不是参数本身，而是 `Authorization` 头缺失、格式不对，或者直接把令牌写成了裸字符串。',
      code: `Authorization: Bearer sk-你的APIKey
Content-Type: application/json`,
      note: '如果你在服务端做代理转发，记得同时保留 `Content-Type: application/json`。',
    },
    {
      title: '请求参数：先把最常用的四个字段用对',
      description:
        '日常接入优先关注下面几个字段：\n\n- `model`：必填，图片生成模型，例如 `gpt-image-2`\n- `prompt`：必填，图片描述，建议写清主体、风格、背景、构图和光线\n- `size`：必填，分辨率字符串，推荐只传文档支持的合法尺寸\n- `n`：可选，生成图片数量；如果你的后台限制单次只生成一张，请保持 `1`\n\n如果需要做稳定生产，建议把 prompt 模板化，把尺寸做成枚举，不要让前端自由拼接任意分辨率。',
      note: '接口是 OpenAI 兼容风格，但并不意味着所有第三方扩展字段都会被接受；遇到 400 时优先回到最小请求体排查。',
    },
    {
      title: '尺寸与计费：明确 1K、2K、4K 三档',
      description:
        '当前可参考的尺寸和消耗口径如下：\n\n- `1024x1024`：1K，1:1，消耗 `0.05`\n- `2048x2048`：2K，1:1，消耗 `0.10`\n- `1536x1024`：2K，3:2，消耗 `0.10`\n- `1024x1536`：2K，2:3，消耗 `0.10`\n- `3840x2160`：4K，16:9，消耗 `0.20`\n- `2160x3840`：4K，9:16，消耗 `0.20`\n\n如果 `size` 为空、`auto` 或传入不支持的分辨率，系统会 fallback 到 2K。最终扣费仍以 CCToken 控制台或账单里的实际规则为准。',
      note: '为了避免无意中升档或降档，生产环境里建议把尺寸固定成下拉选项，而不是开放任意字符串输入。',
    },
    {
      title: '调用示例：先用 cURL 跑通，再迁移到业务代码',
      description:
        '调试阶段推荐先用 cURL 确认接口可用，再迁移到 Python、JavaScript 或工作流节点。下面这段请求适合做联调自检。',
      code: `curl -X POST "https://www.cctoken.fun/v1/images/generations" \\
  -H "Authorization: Bearer sk-你的APIKey" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-image-2",
    "prompt": "A cinematic product photo of a futuristic gaming console, dark background, neon rim light, ultra detailed.",
    "size": "2048x2048"
  }'`,
      note: '如果你更习惯 Python requests 或 JavaScript fetch，可以沿用同样的 URL、Header 和 JSON 结构，不需要额外改协议。',
    },
    {
      title: '返回格式：兼容 URL 和 Base64 两种结果',
      description:
        '接口通常返回 OpenAI 兼容结构。不同上游可能给你图片 URL，也可能直接给 `b64_json`。如果拿到的是 Base64，前端可以拼接成 `data:image/png;base64,...` 后直接预览或下载。',
      code: `{
  "created": 1770000000,
  "data": [
    {
      "url": "https://..."
    }
  ]
}

{
  "created": 1770000000,
  "data": [
    {
      "b64_json": "iVBORw0KGgoAAAANSUhEUg..."
    }
  ]
}`,
      note: '处理返回值时，不要假定 `data[0]` 一定是 URL；先判断 `url` 和 `b64_json` 哪个存在，再决定展示或存储方式。',
    },
  ],
  verification: [
    '访问`https://www.cctoken.fun/` 确保能正常访问首页',
    '返回体中的 `data[0]` 至少包含 `url` 或 `b64_json` 其中一个字段',
    '切换 1K、2K、4K 合法尺寸时，结果分辨率和预期一致，没有误触 fallback',
  ],
  troubleshooting: [
    {
      title: '401 Unauthorized',
      content:
        '通常是 API Key 缺失、格式错误或无效。优先检查请求头是否是标准的 `Authorization: Bearer sk-...` 格式，并确认令牌没有过期或被禁用。',
    },
    {
      title: '400 Bad Request',
      content:
        '常见原因包括缺少 `model`、`prompt`、`size`，或者携带了当前接口不支持的字段。先退回最小请求体，再逐个加字段定位问题。',
    },
    {
      title: '429 Too Many Requests',
      content:
        '一般和请求过快、额度不足、模型限流或账号并发限制有关。可以先降低并发、检查余额，再确认模型当前是否可用。',
    },
    {
      title: '500 / 502 / 504',
      content:
        '这类错误通常来自上游生成超时、网关超时或服务暂时不可用。优先缩短 prompt、降低尺寸，或稍后重试；如果是批量任务，建议加入指数退避和失败重试机制。',
    },
  ],
}
