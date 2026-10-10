<script setup lang="ts">
import { computed } from 'vue'
import type { SiteConfig } from '../api'

const props = defineProps<{ site?: SiteConfig }>()

const brand = computed(() => props.site?.name || 'Draw')
const apiBase = computed(() => props.site?.api_base_url || 'https://draw.superai.sbs/v1')

// 每一项都对应站点已经实现的能力，避免展示不存在的功能。
const features = computed(() => [
  {
    tag: 'ACCOUNT',
    title: '统一账号',
    desc: '注册或登录后即可使用，账号、余额与平台保持一致。'
  },
  {
    tag: 'OVERVIEW',
    title: '账户概览',
    desc: '在一个页面查看余额、充值记录与调用消费，数据只读展示。'
  },
  {
    tag: 'USAGE',
    title: '调用统计',
    desc: '按当前账号汇总调用次数与消费金额，方便对账。'
  },
  {
    tag: 'ADMIN',
    title: '管理看板',
    desc: '管理员可查看成员、充值、调用与待结算汇总。'
  }
])
</script>

<template>
  <div class="landing">
    <section class="landing-hero">
      <div class="landing-hero__copy">
        <p class="landing-kicker">{{ brand }} · OpenAI 兼容接口</p>
        <h1>
          <span>一个账号</span>
          <span class="landing-gradient">接入 AI 工作流</span>
        </h1>
        <p class="landing-lede">
          使用标准 OpenAI 格式的接口地址，统一管理账号余额、充值与调用数据。
        </p>
        <div class="landing-actions">
          <RouterLink class="landing-primary" to="/register">开始使用</RouterLink>
          <a class="landing-secondary" href="#features">查看能力</a>
        </div>
        <RouterLink class="landing-text-link" to="/login">已有账号，进入控制台</RouterLink>
      </div>

      <div class="landing-visual" aria-hidden="true">
        <div class="orbital-sphere">
          <span class="orbital-grid"></span>
        </div>
        <div class="endpoint-card">
          <span class="endpoint-card__label">BASE URL</span>
          <code>{{ apiBase }}</code>
        </div>
        <span class="model-node model-node--compat">OpenAI 格式</span>
        <span class="model-node model-node--stream">流式响应</span>
        <span class="model-node model-node--key">API Key 鉴权</span>
      </div>
    </section>

    <section id="features" class="landing-section">
      <div class="landing-section__head">
        <div>
          <p class="section-kicker">账户能力</p>
          <h2>账户数据，清晰可控</h2>
          <p>登录后即可查看属于当前账号的余额、充值与调用数据。</p>
        </div>
        <RouterLink class="landing-secondary" to="/login">进入控制台</RouterLink>
      </div>

      <div class="feature-grid">
        <article v-for="feature in features" :key="feature.title" class="feature-card">
          <span class="feature-card__tag">{{ feature.tag }}</span>
          <h3>{{ feature.title }}</h3>
          <p>{{ feature.desc }}</p>
        </article>
      </div>
    </section>

    <section class="landing-section landing-integrate">
      <div class="landing-integrate__copy">
        <p class="section-kicker">快速接入</p>
        <h2>保持 OpenAI 格式，直接替换地址</h2>
        <p>
          沿用现有的 OpenAI 兼容客户端，只需要修改接口地址即可。模型与计费口径与平台保持一致。
        </p>
        <RouterLink class="landing-primary" to="/register">创建账号</RouterLink>
      </div>
      <div class="landing-integrate__code">
        <div class="code-window__head">
          <span class="code-dot"></span>
          <span class="code-dot"></span>
          <span class="code-dot"></span>
          <span class="code-window__title">环境变量</span>
        </div>
        <pre class="code-block">export OPENAI_BASE_URL={{ apiBase }}
export OPENAI_API_KEY=&lt;你的 API Key&gt;</pre>
      </div>
    </section>

    <section class="landing-cta">
      <p class="section-kicker">现在开始</p>
      <h2>用统一入口连接你的 AI 工作流</h2>
      <p>一个账号，一套数据，一处查看余额与调用。</p>
      <div class="landing-actions">
        <RouterLink class="landing-primary" to="/register">创建账号</RouterLink>
        <RouterLink class="landing-secondary" to="/login">进入控制台</RouterLink>
      </div>
    </section>
  </div>
</template>
