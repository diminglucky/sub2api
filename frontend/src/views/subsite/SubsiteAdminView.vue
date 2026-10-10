<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">子站管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            本站概览。用户与用量只统计本站数据，账号和余额与主站共享。
          </p>
        </div>
        <button class="btn btn-secondary" :disabled="loading" @click="reload">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          <span class="ml-1">刷新</span>
        </button>
      </header>

      <p v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
        {{ error }}
      </p>

      <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        <article class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <p class="text-xs text-gray-500">用户数</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ summary?.member_count || 0 }}</p>
        </article>
        <article class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <p class="text-xs text-gray-500">充值金额</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ money(summary?.recharge_amount) }}</p>
        </article>
        <article class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <p class="text-xs text-gray-500">调用次数</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ summary?.usage_count || 0 }}</p>
        </article>
        <article class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <p class="text-xs text-gray-500">待结算金额</p>
          <p class="mt-2 text-2xl font-bold text-gray-900 dark:text-white">{{ money(summary?.settlement_pending) }}</p>
        </article>
      </div>

      <section class="grid gap-3 sm:grid-cols-3">
        <RouterLink
          to="/subsite-users"
          class="rounded-xl border border-gray-200 bg-white p-4 transition-colors hover:border-primary-400 dark:border-dark-700 dark:bg-dark-900"
        >
          <p class="text-sm font-semibold text-gray-900 dark:text-white">用户管理</p>
          <p class="mt-1 text-xs text-gray-500">查看本站用户、余额与用量</p>
        </RouterLink>
        <RouterLink
          to="/subsite-channels"
          class="rounded-xl border border-gray-200 bg-white p-4 transition-colors hover:border-primary-400 dark:border-dark-700 dark:bg-dark-900"
        >
          <p class="text-sm font-semibold text-gray-900 dark:text-white">上游管理</p>
          <p class="mt-1 text-xs text-gray-500">管理主站同步的渠道与价格</p>
        </RouterLink>
        <RouterLink
          to="/subsite-settings"
          class="rounded-xl border border-gray-200 bg-white p-4 transition-colors hover:border-primary-400 dark:border-dark-700 dark:bg-dark-900"
        >
          <p class="text-sm font-semibold text-gray-900 dark:text-white">子站设置</p>
          <p class="mt-1 text-xs text-gray-500">配置本站名称、Logo 与主题色</p>
        </RouterLink>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { getSubsiteAdminSummary, type SubsiteAdminSummary } from '@/api/subsiteAdmin'

const summary = ref<SubsiteAdminSummary | null>(null)
const loading = ref(false)
const error = ref('')

onMounted(reload)

async function reload() {
  loading.value = true
  error.value = ''
  try {
    summary.value = await getSubsiteAdminSummary()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载数据失败'
  } finally {
    loading.value = false
  }
}

function money(value: number | undefined): string {
  return `$${Number(value || 0).toFixed(2)}`
}
</script>
