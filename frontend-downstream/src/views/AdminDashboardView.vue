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
    error.value = err instanceof Error ? err.message : 'Failed to load admin summary'
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
        <h1>Scoped admin dashboard</h1>
        <p class="lede">Only data attributed to {{ siteName }} is returned by the internal API.</p>
      </div>
    </header>

    <p v-if="error" class="form-error">{{ error }}</p>

    <div class="grid">
      <article class="stat-card">
        <p class="stat-label">Members</p>
        <p class="stat-value">{{ summary?.member_count || 0 }}</p>
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
        <p class="stat-label">Usage count</p>
        <p class="stat-value">{{ summary?.usage_count || 0 }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">Usage cost</p>
        <p class="stat-value">{{ money(summary?.usage_cost) }}</p>
      </article>
      <article class="stat-card">
        <p class="stat-label">Pending settlement</p>
        <p class="stat-value">{{ money(summary?.settlement_pending) }}</p>
      </article>
    </div>

    <section class="panel">
      <h2>Scope guard</h2>
      <p>The backend filters every metric by the Host-resolved subsite ID. Main-site and other subsite rows are not included.</p>
    </section>
  </section>
</template>
