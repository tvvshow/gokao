# recommendation-service

高考志愿推荐服务，当前支持规则推荐引擎，并已预留兼容 OpenAI Chat Completions 风格的大模型分析接口。

## 运行端口

- 服务内部配置端口统一使用纯数字：`8084`
- HTTP Server 启动时由代码拼接为 `:8084`

## LLM 环境变量

- `LLM_ENABLED`：是否启用外部大模型分析
- `LLM_PROVIDER`：当前默认 `openai-compatible`
- `LLM_BASE_URL`：兼容 OpenAI 的基础地址，例如 `https://api.openai.com/v1`
- `LLM_API_KEY`：模型供应方 API Key
- `LLM_MODEL`：模型名，例如 `gpt-4o-mini`
- `LLM_TIMEOUT`：请求超时，例如 `20s`
- `LLM_MAX_TOKENS`：最大输出 token
- `LLM_TEMPERATURE`：采样温度
- `LLM_FALLBACK_ENABLED`：外部模型失败时是否允许降级到本地规则分析

## /api/v1/system/status 中的 analysis 字段

返回结构为 snake_case，核心字段如下：

- `enabled`
- `provider`
- `status`
- `model`
- `base_url`
- `max_tokens`
- `temperature`
- `fallback_mode`

### status 枚举

- `healthy`：外部 LLM 已启用且无降级依赖
- `degraded`：当前依赖降级链路，或仅使用本地回退分析
- `unhealthy`：依赖已知异常，无法正常提供分析能力
- `not_configured`：未配置外部 LLM

### fallback_mode 枚举

- `none`：无降级策略
- `local_rules`：本地规则文案回退
- `static_fallback`：静态降级逻辑（预留）

## Docker Compose

`docker-compose.yml` 中 recommendation-service 已对齐：

- `SERVER_PORT=8084`
- `SERVER_MODE=release`
- LLM 相关环境变量统一以 `RECOMMENDATION_LLM_*` 注入

