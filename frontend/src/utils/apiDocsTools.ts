export interface DocsToolGuide {
  slug: string
  envKey: string
  install: string
  configPath: string
  config: (baseUrl: string, apiRoot: string) => string
  run: string
  verify: string
  notes: string
}

export const docsToolGuides: DocsToolGuide[] = [
  {
    slug: 'cc-switch',
    envKey: 'SUPERAI_API_KEY',
    install: '从 GitHub Releases 下载并安装 CC Switch，再安装 Codex。',
    configPath: 'CC Switch Provider',
    config: () => '',
    run: '在 CC Switch 中启用 SuperAI Provider，然后重新打开 Codex。',
    verify: '在 Codex 中发送一条测试消息，并在 SuperAI 使用记录中确认请求。',
    notes: '为 CC Switch 单独创建一个 SuperAI API Key。'
  },
  {
    slug: 'zcf',
    envKey: 'SUPERAI_API_KEY',
    install: '安装并打开 ZCF 客户端，确认版本支持自定义 OpenAI Compatible Provider。',
    configPath: 'ZCF → 设置 → 模型服务 → 新增 OpenAI Compatible Provider',
    config: (baseUrl) => `Base URL: ${baseUrl}\nAPI Key: sk-...\nModel: YOUR_MODEL`,
    run: '在 ZCF 中选择刚配置的 Provider 和模型，然后发送一条测试消息。',
    verify: '在 ZCF 中选择模型并发送一条短消息，确认返回正常内容。',
    notes: '建议为 ZCF 单独创建一个 SuperAI API Key，便于撤销和统计用量。'
  },
  {
    slug: 'claude-code',
    envKey: 'ANTHROPIC_AUTH_TOKEN',
    install: 'npm install -g @anthropic-ai/claude-code\nclaude --version',
    configPath: 'Shell 环境变量（~/.zshrc / ~/.bashrc / PowerShell profile）',
    config: (_baseUrl, apiRoot) => `ANTHROPIC_BASE_URL="${apiRoot}"\nANTHROPIC_AUTH_TOKEN="sk-..."`,
    run: 'claude',
    verify: '进入项目目录运行 claude，发送一条短消息确认能正常回复。',
    notes: 'Claude Code 使用 Messages 协议，Base URL 填 API 根地址，不要追加 /v1。'
  },
  {
    slug: 'codex',
    envKey: 'SUPERAI_API_KEY',
    install: 'curl -fsSL https://chatgpt.com/codex/install.sh | sh\n\n# 或使用 npm\nnpm install -g @openai/codex\ncodex --version',
    configPath: '~/.codex/config.toml',
    config: (baseUrl) => `model = "gpt-5.6-sol"\nmodel_provider = "superai"\n\n[model_providers.superai]\nname = "SuperAI"\nbase_url = "${baseUrl}"\nenv_key = "SUPERAI_API_KEY"\nwire_api = "responses"\nstream_idle_timeout_ms = 300000`,
    run: 'codex',
    verify: '进入项目目录运行 codex，然后在会话中执行 /status，确认 provider 和 model。',
    notes: 'Codex CLI 与 IDE 扩展共用用户级配置。不要把真实 Key 写入 config.toml，也不要手工创建 auth.json。'
  },
  {
    slug: 'cursor',
    envKey: 'SUPERAI_API_KEY',
    install: '安装并打开 Cursor，确认可在 Settings → Models 中配置 OpenAI API Key。',
    configPath: 'Cursor Settings → Models → OpenAI API Key',
    config: (baseUrl) => `OpenAI Base URL: ${baseUrl}\nAPI Key: sk-...\nModel: YOUR_MODEL`,
    run: '在 Cursor Chat / Composer 中选择已配置模型。',
    verify: '在 Cursor Chat 中发送一条测试消息，确认模型能正常返回。',
    notes: '如果 Cursor 没有 OpenAI Compatible 选项，请升级到支持自定义 Base URL 的版本。'
  },
  {
    slug: 'roocode',
    envKey: 'SUPERAI_API_KEY',
    install: '在 VS Code 扩展市场安装 RooCode，并重新加载窗口。',
    configPath: 'RooCode Settings → Provider',
    config: (baseUrl) => `Provider: OpenAI Compatible\nBase URL: ${baseUrl}\nAPI Key: sk-...\nModel: YOUR_MODEL`,
    run: '保存配置后在 RooCode 面板中选择模型并开始任务。',
    verify: '在 RooCode 中发起一个简单任务，确认模型返回正常。',
    notes: 'RooCode 会复用 VS Code 网络设置，公司代理环境需要同时配置代理。'
  },
  {
    slug: 'opencode',
    envKey: 'SUPERAI_API_KEY',
    install: 'npm install -g opencode-ai\nopencode --version',
    configPath: 'OpenCode 自定义 Provider 配置（以官方文档为准）',
    config: (baseUrl) => `{
  "provider": {
    "superai": {
      "npm": "@ai-sdk/openai-compatible",
      "options": {
        "baseURL": "${baseUrl}",
        "apiKey": "sk-..."
      },
      "models": {
        "YOUR_MODEL": {}
      }
    }
  }
}`,
    run: 'opencode',
    verify: '启动后选择 superai Provider 和模型，发送一条测试消息。',
    notes: 'OpenCode 使用 OpenAI 兼容协议时，baseURL 要包含 /v1。'
  },
  {
    slug: 'openclaw',
    envKey: 'OPENAI_API_KEY',
    install: '按 OpenClaw 官方文档安装并完成首次初始化。',
    configPath: 'Shell 环境变量或 OpenClaw 运行时配置',
    config: (baseUrl) => `OPENAI_BASE_URL="${baseUrl}"\nOPENAI_API_KEY="sk-..."`,
    run: '启动 OpenClaw，并选择 YOUR_MODEL 对应的模型。',
    verify: '发送一条测试消息，确认 OpenClaw 能通过 SuperAI 调用模型。',
    notes: 'OpenClaw 通常读取 OPENAI_BASE_URL / OPENAI_API_KEY；修改后需要重启进程。'
  },
  {
    slug: 'workbuddy',
    envKey: 'SUPERAI_API_KEY',
    install: '安装并打开 WorkBuddy。',
    configPath: 'WorkBuddy → 自定义 Chat Completions',
    config: (baseUrl) => `Chat Completions URL: ${baseUrl}\nAPI Key: sk-...\nModel: YOUR_MODEL`,
    run: '保存后在 WorkBuddy 中选择该模型开始对话。',
    verify: '在 WorkBuddy 中发送一条测试消息，确认请求成功。',
    notes: 'WorkBuddy 使用 Chat Completions 协议时，URL 必须包含 /v1。'
  },
  {
    slug: 'droid-cli',
    envKey: 'SUPERAI_API_KEY',
    install: '按 Droid CLI 官方文档安装 CLI，并完成首次登录或初始化。',
    configPath: 'Droid CLI 自定义 OpenAI Provider',
    config: (baseUrl) => `base_url = "${baseUrl}"\napi_key  = "sk-..."\nmodel    = "YOUR_MODEL"`,
    run: 'droid',
    verify: '在 Droid 中发送一条测试消息，确认 provider 生效。',
    notes: '如果客户端支持协议选择，请选择 OpenAI Compatible 而不是原生 OpenAI。'
  },
  {
    slug: 'deepseek-harness',
    envKey: 'SUPERAI_API_KEY',
    install: '安装 DeepSeek Harness，并确认支持自定义 OpenAI Compatible Provider。',
    configPath: 'DeepSeek Harness → 自定义 OpenAI Compatible Provider',
    config: (baseUrl) => `base_url = "${baseUrl}"\napi_key  = "sk-..."\nmodel    = "YOUR_MODEL"`,
    run: '启动 DeepSeek Harness 并选择已配置模型。',
    verify: '发送一条测试消息，确认请求经过 SuperAI 网关。',
    notes: '模型 ID 必须以 Models 接口返回为准，不要直接填供应商官方名称。'
  },
  {
    slug: 'qwen-code',
    envKey: 'SUPERAI_API_KEY',
    install: 'npm install -g @qwen-code/qwen-code@latest\nqwen --version',
    configPath: '~/.qwen/settings.json',
    config: (baseUrl) => `{
  "modelProviders": {
    "openai": {
      "protocol": "openai",
      "models": [
        {
          "id": "gpt-5.6-sol",
          "name": "GPT-5.6 Sol via SuperAI",
          "baseUrl": "${baseUrl}",
          "envKey": "SUPERAI_API_KEY"
        },
        {
          "id": "claude-sonnet-4-5",
          "name": "Claude Sonnet 4.5 via SuperAI",
          "baseUrl": "${baseUrl}",
          "envKey": "SUPERAI_API_KEY"
        }
      ]
    }
  }
}`,
    run: 'qwen',
    verify: '进入 Qwen Code 后执行 /model，选择已配置模型并发送测试消息。',
    notes: '旧版 security.auth.apiKey / security.auth.baseUrl 正在弃用，优先使用 modelProviders + envKey。'
  }
]
