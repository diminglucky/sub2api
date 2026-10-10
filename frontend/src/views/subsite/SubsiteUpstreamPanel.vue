<template>
  <div class="space-y-4">
    <p v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
      {{ error }}
    </p>
    <div class="flex justify-end">
      <button class="btn btn-secondary" :disabled="loading" @click="reload">
        <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        <span class="ml-1">刷新</span>
      </button>
    </div>

    <div v-if="!groups.length && !loading" class="rounded-lg border border-dashed border-gray-200 py-10 text-center text-sm text-gray-400 dark:border-dark-700">
      主站还没有给本站分配分组
    </div>

    <section
      v-for="group in groups"
      :key="group.id"
      class="rounded-lg border border-gray-200 p-4 dark:border-dark-700"
    >
      <div class="flex flex-wrap items-center gap-2">
        <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ group.name }}</span>
        <span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-800">{{ group.platform }}</span>
        <span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-800">主站倍率 ×{{ group.mainMultiplier }}</span>
      </div>

      <div v-if="!group.rows.length" class="mt-3 rounded border border-dashed border-gray-200 py-5 text-center text-xs text-gray-400 dark:border-dark-700">
        该分组暂无可用模型（主站尚未配置渠道/模型）
      </div>

      <div v-else class="mt-3 overflow-x-auto rounded border border-gray-200 dark:border-dark-700">
        <table class="min-w-full text-sm">
          <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
            <tr>
              <th class="px-3 py-2">模型</th>
              <th class="px-3 py-2">平台</th>
              <th class="px-3 py-2">计费</th>
              <th class="px-3 py-2">主站倍率</th>
              <th class="px-3 py-2">子站倍率</th>
              <th class="px-3 py-2 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in group.rows" :key="row.model" class="border-t border-gray-100 dark:border-dark-700">
              <td class="px-3 py-2">{{ row.model }}</td>
              <td class="px-3 py-2 text-gray-500">{{ row.platform }}</td>
              <td class="px-3 py-2 text-gray-500">{{ billingLabel(row.billing_mode) }}</td>
              <td class="px-3 py-2">
                <span class="rounded bg-gray-100 px-2 py-1 font-mono text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                  ×{{ row.main_multiplier }}
                </span>
              </td>
              <td class="px-3 py-2">
                <input
                  v-model.number="draft[row.model]"
                  class="input w-24"
                  type="number"
                  step="0.01"
                  :min="row.main_multiplier"
                  :class="{ 'input-error': (draft[row.model] ?? row.main_multiplier) < row.main_multiplier }"
                />
              </td>
              <td class="px-3 py-2 text-right">
                <button
                  class="text-primary-600 hover:underline disabled:opacity-40"
                  :disabled="saving || (draft[row.model] ?? row.main_multiplier) < row.main_multiplier"
                  @click="saveModel(row)"
                >
                  保存
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import {
  createSubsitePrice,
  listSubsiteChannels,
  listSubsiteModels,
  listSubsitePrices,
  updateSubsitePrice,
  type SubsiteModelPrice
} from '@/api/subsiteAdmin'

interface GroupRows {
  id: number
  name: string
  platform: string
  mainMultiplier: number
  rows: SubsiteModelPrice[]
}

const groups = ref<GroupRows[]>([])
const overrideIds = ref<Record<string, number>>({})
const draft = reactive<Record<string, number>>({})
const loading = ref(false)
const saving = ref(false)
const error = ref('')

onMounted(reload)

function billingLabel(mode: string): string {
  if (mode === 'per_request') return '按次'
  if (mode === 'image') return '生图'
  if (mode === 'token') return '按量'
  return '—'
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const [channels, models, prices] = await Promise.all([
      listSubsiteChannels(),
      listSubsiteModels(),
      listSubsitePrices()
    ])
    overrideIds.value = {}
    for (const price of prices) {
      if (price.scope === 'model' && price.model) {
        overrideIds.value[price.model] = price.id
      }
    }
    groups.value = channels.map((channel) => ({
      id: channel.group_id,
      name: channel.name,
      platform: channel.platform,
      mainMultiplier: channel.main_rate_multiplier,
      rows: models.filter((model) => model.group_id === channel.group_id)
    }))
    for (const model of models) {
      draft[model.model] = model.subsite_multiplier ?? model.main_multiplier
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载失败'
  } finally {
    loading.value = false
  }
}

async function saveModel(row: SubsiteModelPrice) {
  const value = Number(draft[row.model]) || row.main_multiplier
  if (value < row.main_multiplier) {
    error.value = '子站倍率不能低于主站倍率'
    return
  }
  saving.value = true
  error.value = ''
  try {
    const payload = { scope: 'model' as const, model: row.model, rate_multiplier: value }
    const existing = overrideIds.value[row.model]
    if (existing) {
      await updateSubsitePrice(existing, payload)
    } else {
      await createSubsitePrice(payload)
    }
    await reload()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    saving.value = false
  }
}
</script>
