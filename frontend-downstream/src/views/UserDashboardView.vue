<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getMeSummary, type SiteConfig, type UserSummary } from '../api'

const props = defineProps<{ site?: SiteConfig }>()

const summary = ref<UserSummary | null>(null)
const error = ref('')

const apiBase = computed(() => props.site?.api_base_url || 'https://draw.superai.sbs/v1')

onMounted(async () => {
  try {
    summary.value = await getMeSummary()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '账户数据加载失败'
  }
})

function money(value: number | undefined): string {
  return `$${Number(value || 0).toFixed(2)}`
}
</script>

<template>
  <section>
    <header class="page-heading">
      <div>
        <p class="eyebrow">{{ site?.name || 'Draw' }}</p>
        <h1>用户中心</h1>
        <p class="lede">查看账户余额、充值记录和调用情况，数据来自当前账号的统一账本。</p>
      </div>
    </header>

    <p v-if="error" class="form-error">{{ error }}</p>

    <div class="grid">
      <article class="stat-card">
        <p class="stat-label">账户余额</p>
        <p class="stat-value">{{ money(summary?.balance) }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">调用次数</p>
        <p class="stat-value">{{ summary?.usage_count || 0 }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">调用消费</p>
        <p class="stat-value">{{ money(summary?.usage_cost) }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">充值笔数</p>
        <p class="stat-value">{{ summary?.recharge_count || 0 }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">充值金额</p>
        <p class="stat-value">{{ money(summary?.recharge_amount) }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">登录账号</p>
        <p class="stat-value">{{ summary?.email || '...' }}</p>
      </article>
    </div>

    <section class="panel">
      <div class="panel-heading">
        <div>
          <h2>OpenAI 兼容接入</h2>
          <p>使用标准 OpenAI 格式的接口地址，通过 API Key 完成鉴权。</p>
        </div>
        <span class="status-chip">已启用</span>
      </div>
      <pre class="code-block">export OPENAI_BASE_URL={{ apiBase }}
export OPENAI_API_KEY=&lt;你的 API Key&gt;</pre>
    </section>

    <section class="panel muted-panel">
      <h2>数据说明</h2>
      <p>本页展示归属于 {{ site?.name || '当前账号' }} 的余额、充值和调用统计，均为只读数据。</p>
    </section>
  </section>
</template>
