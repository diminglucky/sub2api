<script setup lang="ts">
import { computed } from 'vue'
import { hasSession, type SiteConfig } from '../api'

const props = defineProps<{ site?: SiteConfig }>()

const brand = computed(() => props.site?.name || 'Draw')
const apiBase = computed(() => props.site?.api_base_url || 'https://draw.superai.sbs/v1')
const signedIn = computed(() => hasSession())
const startTarget = computed(() => (signedIn.value ? '/studio' : '/register'))

// 每一项都对应站点已经实现的绘图能力，避免展示不存在的功能。
const features = computed(() => [
  {
    tag: 'TEXT',
    title: '文生图',
    desc: '输入自然语言提示词即可生成图片，支持自定义尺寸与画质。'
  },
  {
    tag: 'IMAGE',
    title: '图生图',
    desc: '上传参考图后按提示词生成新的图片，保持主体与构图。'
  },
  {
    tag: 'MODEL',
    title: '多模型可选',
    desc: '在工作台中直接选择账号可用的图片模型。'
  },
  {
    tag: 'API',
    title: 'OpenAI 兼容接口',
    desc: '用标准 OpenAI 格式的地址直接调用图片生成接口。'
  }
])
</script>

<template>
  <div class="landing">
    <section class="landing-hero">
      <div class="landing-hero__copy">
        <p class="landing-kicker">{{ brand }} · AI 绘图工作台</p>
        <h1>
          <span>一句话</span>
          <span class="landing-gradient">生成你想要的图片</span>
        </h1>
        <p class="landing-lede">
          支持文生图与图生图，在同一处选择模型、生成图片并下载结果，账户余额与调用数据实时可见。
        </p>
        <div class="landing-actions">
          <RouterLink class="landing-primary" :to="startTarget">开始绘图</RouterLink>
          <a class="landing-secondary" href="#features">查看能力</a>
        </div>
        <RouterLink class="landing-text-link" to="/login">已有账号，进入控制台</RouterLink>
      </div>

      <div class="landing-visual" aria-hidden="true">
        <div class="orbital-sphere">
          <span class="orbital-grid"></span>
        </div>
        <div class="prompt-float prompt-float--one">戴着宇航头盔的橘猫</div>
        <div class="prompt-float prompt-float--two">赛博朋克城市夜景</div>
        <div class="prompt-float prompt-float--three">水彩风山间小屋</div>
        <div class="endpoint-card">
          <span class="endpoint-card__label">OPENAI 兼容</span>
          <code>{{ apiBase }}</code>
        </div>
      </div>
    </section>

    <section id="features" class="landing-section">
      <div class="landing-section__head">
        <div>
          <p class="section-kicker">绘图能力</p>
          <h2>从提示词到成图，一站完成</h2>
          <p>登录后进入绘图工作台，选择模型、填写提示词即可生成图片。</p>
        </div>
        <RouterLink class="landing-secondary" :to="startTarget">打开工作台</RouterLink>
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
        <p class="section-kicker">接口调用</p>
        <h2>也可以用自己的客户端生成图片</h2>
        <p>
          沿用现有 OpenAI 兼容客户端，只需要修改接口地址即可调用图片生成接口，计费口径与平台保持一致。
        </p>
        <RouterLink class="landing-primary" :to="startTarget">开始使用</RouterLink>
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
      <h2>把想法变成图片</h2>
      <p>一个账号，一套余额，一处生成与查看。</p>
      <div class="landing-actions">
        <RouterLink class="landing-primary" :to="startTarget">开始绘图</RouterLink>
        <RouterLink v-if="!signedIn" class="landing-secondary" to="/login">进入控制台</RouterLink>
        <RouterLink v-else class="landing-secondary" to="/dashboard">查看账户</RouterLink>
      </div>
    </section>
  </div>
</template>
