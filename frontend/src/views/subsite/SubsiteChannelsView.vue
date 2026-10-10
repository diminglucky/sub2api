<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">上游管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            本站开放的分组与主站价格。价格完全跟随主站，由主站统一维护。
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

      <div v-if="!channels.length && !loading" class="rounded-xl border border-gray-200 bg-white py-12 text-center text-sm text-gray-400 dark:border-dark-700 dark:bg-dark-900">
        主站还没有给本站分配分组
      </div>

      <section
        v-for="channel in channels"
        :key="channel.group_id"
        class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="flex flex-wrap items-center gap-2">
          <span class="text-base font-semibold text-gray-900 dark:text-white">{{ channel.name }}</span>
          <span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-800">{{ channel.platform }}</span>
        </div>

        <div class="mt-4 grid gap-4 lg:grid-cols-[220px_minmax(0,1fr)]">
          <div>
            <p class="text-xs font-medium text-gray-500">倍率（token 模型）</p>
            <p class="mt-2 font-mono text-sm text-gray-900 dark:text-white">×{{ channel.main_rate_multiplier }}</p>
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500">生图一口价（USD / 张）</p>
            <div class="mt-2 grid gap-3 sm:grid-cols-3">
              <div v-for="tier in tiers" :key="tier.key">
                <span class="text-xs text-gray-500">{{ tier.label }}</span>
                <p class="mt-1 font-mono text-sm text-gray-900 dark:text-white">
                  {{ formatPrice(channel[tier.field]) }}
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { listSubsiteChannels, type SubsiteChannel } from '@/api/subsiteAdmin'

const tiers = [
  { key: '1k', label: '1K', field: 'main_image_price_1k' },
  { key: '2k', label: '2K', field: 'main_image_price_2k' },
  { key: '4k', label: '4K', field: 'main_image_price_4k' }
] as const

const channels = ref<SubsiteChannel[]>([])
const loading = ref(false)
const error = ref('')

onMounted(reload)

async function reload() {
  loading.value = true
  error.value = ''
  try {
    channels.value = await listSubsiteChannels()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载渠道失败'
  } finally {
    loading.value = false
  }
}

function formatPrice(value: number | null): string {
  if (value === null || value === undefined) return '—'
  return `$${Number(value).toFixed(4)}`
}
</script>
