<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getAdminSummary, type AdminSummary, type SiteConfig } from '../api'

const props = defineProps<{ site?: SiteConfig }>()

const summary = ref<AdminSummary | null>(null)
const error = ref('')

const siteName = computed(() => props.site?.name || 'Draw')

onMounted(async () => {
  try {
    summary.value = await getAdminSummary()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '管理数据加载失败'
  }
})

function money(value: number | undefined): string {
  return Number(value || 0).toFixed(2)
}
</script>

<template>
  <section>
    <header class="page-heading">
      <div>
        <p class="eyebrow">{{ siteName }}</p>
        <h1>管理看板</h1>
        <p class="lede">查看 {{ siteName }} 的用户、充值、调用和待结算数据。</p>
      </div>
    </header>

    <p v-if="error" class="form-error">{{ error }}</p>

    <div class="grid">
      <article class="stat-card">
        <p class="stat-label">用户数</p>
        <p class="stat-value">{{ summary?.member_count || 0 }}</p>
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
        <p class="stat-label">调用次数</p>
        <p class="stat-value">{{ summary?.usage_count || 0 }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">调用消费</p>
        <p class="stat-value">{{ money(summary?.usage_cost) }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">待结算金额</p>
        <p class="stat-value">{{ money(summary?.settlement_pending) }}</p>
      </article>
    </div>

    <section class="panel">
      <div class="panel-heading">
        <div>
          <h2>结算口径</h2>
          <p>主站价格为成本价，子站售价超过成本的部分按实际 API 用量形成结算差额。</p>
        </div>
        <span class="status-chip">只读台账</span>
      </div>
    </section>
  </section>
</template>
