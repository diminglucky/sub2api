<template>
  <div v-if="visible" class="inline-flex items-center gap-1">
    <button
      type="button"
      class="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-[9px] font-medium text-teal-600 transition-colors hover:bg-teal-50 disabled:cursor-not-allowed disabled:opacity-60 dark:text-teal-400 dark:hover:bg-teal-900/30"
      :disabled="loading"
      :title="queryTitle"
      @click="handleQuery"
    >
      <svg
        class="h-2.5 w-2.5"
        :class="{ 'animate-spin': loading }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      {{ creditsLabel }}
    </button>
    <button
      v-if="canReset"
      type="button"
      class="inline-flex items-center gap-0.5 rounded px-1.5 py-0.5 text-[9px] font-medium text-amber-600 transition-colors hover:bg-amber-50 disabled:cursor-not-allowed disabled:opacity-60 dark:text-amber-400 dark:hover:bg-amber-900/30"
      :disabled="resetting || loading || !canReset"
      :title="t('admin.accounts.openaiQuotaReset.resetTooltipReady')"
      @click="handleReset"
    >
      <svg
        class="h-2.5 w-2.5"
        :class="{ 'animate-spin': resetting }"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v6h6M20 20v-6h-6M5 19a9 9 0 0014-7.5M19 5A9 9 0 005 12.5" />
      </svg>
      {{ t('admin.accounts.openaiQuotaReset.reset') }}
    </button>
    <span v-if="error" class="max-w-[120px] truncate text-[9px] text-red-500" :title="error">
      {{ error }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { OpenAIQuotaUsage } from '@/api/admin/accounts'
import type { Account } from '@/types'

const props = defineProps<{
  account: Account
}>()

const emit = defineEmits<{
  reset: []
}>()

const { t } = useI18n()

const quota = ref<OpenAIQuotaUsage | null>(null)
const loading = ref(false)
const resetting = ref(false)
const error = ref<string | null>(null)

const visible = computed(() => props.account.platform === 'openai' && props.account.type === 'oauth')
const availableCredits = computed(() => quota.value?.rate_limit_reset_credits?.available_count ?? null)
const canReset = computed(() => (availableCredits.value ?? 0) > 0)

const creditsLabel = computed(() => {
  if (availableCredits.value == null) return t('admin.accounts.openaiQuotaReset.count')
  return `${t('admin.accounts.openaiQuotaReset.count')} ${availableCredits.value}`
})

const queryTitle = computed(() => {
  const fetchedAt = quota.value?.fetched_at
  if (!fetchedAt) return t('admin.accounts.openaiQuotaReset.countTooltipLoad')
  return t('admin.accounts.openaiQuotaReset.countTooltipRefreshTime', {
    time: new Date(fetchedAt * 1000).toLocaleString()
  })
})

const extractErrorMessage = (e: unknown, fallback: string): string => {
  const err = e as {
    message?: string
    reason?: string
    response?: { data?: { message?: string; error?: string } }
  }
  return (
    err?.message ||
    err?.reason ||
    err?.response?.data?.message ||
    err?.response?.data?.error ||
    fallback
  )
}

const handleQuery = async () => {
  loading.value = true
  error.value = null
  try {
    quota.value = await adminAPI.accounts.queryOpenAIQuota(props.account.id)
  } catch (e) {
    error.value = extractErrorMessage(e, t('admin.accounts.openaiQuotaReset.queryFailed'))
  } finally {
    loading.value = false
  }
}

const handleReset = async () => {
  if (!canReset.value) return
  resetting.value = true
  error.value = null
  try {
    await adminAPI.accounts.resetOpenAIQuota(props.account.id)
    await handleQuery()
    emit('reset')
  } catch (e) {
    error.value = extractErrorMessage(e, t('admin.accounts.openaiQuotaReset.resetFailed'))
  } finally {
    resetting.value = false
  }
}

watch(
  () => props.account.id,
  () => {
    quota.value = null
    error.value = null
    loading.value = false
    resetting.value = false
  }
)
</script>
