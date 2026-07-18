<template>
  <BaseDialog
    :show="show"
    :title="t('admin.announcements.emailStatus')"
    width="wide"
    @close="handleClose"
  >
    <div class="space-y-4">
      <div class="flex justify-end">
        <button class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="load">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div v-if="loading && items.length === 0" class="py-10 text-center text-sm text-gray-500">
        {{ t('common.loading') }}
      </div>
      <div v-else-if="items.length === 0" class="py-10 text-center text-sm text-gray-500">
        {{ t('admin.announcements.noEmailBatches') }}
      </div>
      <div v-else class="space-y-3">
        <div
          v-for="batch in items"
          :key="batch.id"
          class="rounded-lg border border-gray-200 p-4 dark:border-gray-700"
        >
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="truncate font-medium text-gray-900 dark:text-white">{{ batch.title }}</div>
              <div class="mt-1 text-xs text-gray-500">{{ formatDateTime(batch.created_at) }}</div>
            </div>
            <span :class="['badge', statusClass(batch.status)]">
              {{ t(`admin.announcements.emailStatusLabels.${batch.status}`) }}
            </span>
          </div>

          <dl class="mt-4 grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
            <div>
              <dt class="text-gray-500">{{ t('admin.announcements.emailTotal') }}</dt>
              <dd class="mt-1 font-medium text-gray-900 dark:text-white">{{ batch.total_count }}</dd>
            </div>
            <div>
              <dt class="text-gray-500">{{ t('admin.announcements.emailProcessed') }}</dt>
              <dd class="mt-1 font-medium text-green-600">{{ batch.processed_count }}</dd>
            </div>
            <div>
              <dt class="text-gray-500">{{ t('admin.announcements.emailFailed') }}</dt>
              <dd class="mt-1 font-medium text-red-600">{{ batch.failed_count }}</dd>
            </div>
            <div>
              <dt class="text-gray-500">{{ t('admin.announcements.emailAttempts') }}</dt>
              <dd class="mt-1 font-medium text-gray-900 dark:text-white">
                {{ batch.attempt_count }} / {{ batch.max_attempts }}
              </dd>
            </div>
          </dl>

          <div v-if="batch.status === 'retrying'" class="mt-3 text-xs text-amber-700 dark:text-amber-300">
            {{ t('admin.announcements.emailNextRetry') }}: {{ formatDateTime(batch.next_attempt_at) }}
          </div>
          <div v-if="batch.last_error" class="mt-3 break-words rounded bg-red-50 p-2 text-xs text-red-700 dark:bg-red-900/20 dark:text-red-300">
            {{ batch.last_error }}
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button type="button" class="btn btn-secondary" @click="handleClose">{{ t('common.close') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import type { AnnouncementEmailBatch, AnnouncementEmailBatchStatus } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'

const props = defineProps<{ show: boolean; announcementId: number | null }>()
const emit = defineEmits<{ (e: 'close'): void }>()
const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const items = ref<AnnouncementEmailBatch[]>([])
let controller: AbortController | null = null

function statusClass(status: AnnouncementEmailBatchStatus) {
  if (status === 'completed') return 'badge-success'
  if (status === 'failed') return 'badge-danger'
  if (status === 'retrying') return 'badge-warning'
  return 'badge-gray'
}

async function load() {
  if (!props.show || !props.announcementId) return
  controller?.abort()
  const requestController = new AbortController()
  controller = requestController
  try {
    loading.value = true
    const result = await adminAPI.announcements.getEmailBatches(props.announcementId, { signal: requestController.signal })
    if (!requestController.signal.aborted && controller === requestController) items.value = result
  } catch (error: any) {
    if (requestController.signal.aborted || error?.code === 'ERR_CANCELED') return
    appStore.showError(error.response?.data?.detail || t('admin.announcements.failedToLoadEmailStatus'))
  } finally {
    if (controller === requestController) {
      loading.value = false
      controller = null
    }
  }
}

function handleClose() {
  controller?.abort()
  controller = null
  items.value = []
  emit('close')
}

watch(() => [props.show, props.announcementId], () => {
  if (props.show) load()
  else items.value = []
})

onUnmounted(() => controller?.abort())
</script>
