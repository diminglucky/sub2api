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
    error.value = err instanceof Error ? err.message : 'Failed to load account summary'
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
        <h1>Account dashboard</h1>
        <p class="lede">Balance, recharge and API usage stay synchronized with the main site.</p>
      </div>
    </header>

    <p v-if="error" class="form-error">{{ error }}</p>

    <div class="grid">
      <article class="stat-card">
        <p class="stat-label">Balance</p>
        <p class="stat-value">{{ money(summary?.balance) }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">Usage</p>
        <p class="stat-value">{{ summary?.usage_count || 0 }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">Usage cost</p>
        <p class="stat-value">{{ money(summary?.usage_cost) }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">Recharges</p>
        <p class="stat-value">{{ summary?.recharge_count || 0 }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">Recharge amount</p>
        <p class="stat-value">{{ money(summary?.recharge_amount) }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">Account</p>
        <p class="stat-value">{{ summary?.email || '...' }}</p>
      </article>
    </div>

    <section class="panel">
      <h2>OpenAI-compatible API</h2>
      <p>Use this base URL with the API key managed on the main site.</p>
      <pre class="code-block">export OPENAI_BASE_URL={{ apiBase }}
export OPENAI_API_KEY=&lt;main-site-api-key&gt;</pre>
    </section>
  </section>
</template>
