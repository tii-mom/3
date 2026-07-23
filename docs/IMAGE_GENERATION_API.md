# 生图 API 使用指南

本文说明如何通过现有的 3API Key，在同一个应用里同时调用文本和图片生成。

## 地址与密钥

| 用途 | 地址 |
|------|------|
| 控制台 / 用户网站 | `https://3api.shop` |
| API 网关 | `https://api.3api.shop` |
| OpenAI SDK `base_url` | `https://api.3api.shop/v1` |

无需为生图单独创建 API Key 或分组。只要当前用户分组允许生图、路由中存在可用的图片模型，同一个 Key 就可以继续使用文本模型并调用生图。

## 推荐：Responses 中使用图片工具

在日常开发中，推荐使用 `/v1/responses`，让文本和图片请求保持在同一套调用方式中。向 `tools` 显式加入 `image_generation` 即可：

```bash
curl https://api.3api.shop/v1/responses \
  -H "Authorization: Bearer sk-你的密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.6",
    "input": "生成一张科技风产品宣传图",
    "tools": [
      {"type": "image_generation", "model": "gpt-image-2"}
    ]
  }'
```

### Python

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-你的密钥",
    base_url="https://api.3api.shop/v1",
)

response = client.responses.create(
    model="gpt-5.6",
    input="生成一张科技风产品宣传图",
    tools=[{"type": "image_generation", "model": "gpt-image-2"}],
)
print(response)
```

### Node.js

```javascript
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: "sk-你的密钥",
  baseURL: "https://api.3api.shop/v1",
});

const response = await client.responses.create({
  model: "gpt-5.6",
  input: "生成一张科技风产品宣传图",
  tools: [{ type: "image_generation", model: "gpt-image-2" }],
});
console.log(response);
```

应用可以根据 `response.output` 中的图片生成结果继续保存、展示或下载图片。文本请求仍然使用同一个 `client` 和同一个 API Key。

## 仅需要图片时

也可以直接调用 OpenAI 兼容的图片接口；它仍然使用同一个 API Key：

```bash
curl https://api.3api.shop/v1/images/generations \
  -H "Authorization: Bearer sk-你的密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "一张科技风产品宣传图",
    "size": "1024x1024"
  }'
```

## 兼容边界

- 混合文本与生图请使用 `/v1/responses`，并显式传入 `image_generation` 工具。
- `/v1/chat/completions` 目前不会把 `image_generation` 工具透明转发到图片上游；仅切换模型不能让 Chat Completions 自动生图。
- Responses Lite 或会自行过滤工具的客户端不会自动触发生图，需要改用完整 Responses 请求。
- 管理后台仍需保持分组的“允许当前分组生图”权限，并确保账号/渠道有可用的图片模型；否则会返回模型或权限错误。

## CCS / Codex 模型选择

对于开启“允许当前分组生图”的 OpenAI 分组，网关会在 Codex `/models` 清单和标准 `/v1/models` 清单中动态加入 `gpt-image-2`。客户端选择该模型后，网关会把请求转换为 Responses 文本模型加 `image_generation` 工具调用；无需在 CCS 中新增 API Key 或单独配置生图接口。显式自定义模型列表仍以管理员配置为准。

## 快速自检

```bash
curl -sS https://api.3api.shop/health
```

返回 `status: "ok"` 只表示网关和基础依赖健康；真实生图还需要有效 API Key、账户额度，以及可用的图片模型路由。
