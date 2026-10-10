<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">上游管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            本站开放模型的倍率。主站倍率只读，子站倍率不能低于主站倍率；模型基础价由主站统一维护。
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

      <div v-if="!rows.length && !loading" class="rounded-xl border border-gray-200 bg-white py-12 text-center text-sm text-gray-400 dark:border-dark-700 dark:bg-dark-900">
        主站还没有给本站分配分组或模型
      </div>

      <section
        v-for="group in grouped"
        :key="group.id"
        class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="mb-3 flex items-center gap-2">
          <span class="text-base font-semibold text-gray-900 dark:text-white">{{ group.name }}</span>
          <span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-800">主站倍率 ×{{ group.mainMultiplier }}</span>
        </div>

        <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
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
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  createSubsitePrice,
  listSubsiteModels,
  listSubsitePrices,
  updateSubsitePrice,
  type SubsiteModelPrice
} from '@/api/subsiteAdmin'

interface GroupRows {
  id: number
  name: string
  mainMultiplier: number
  rows: SubsiteModelPrice[]
}

const rows = ref<SubsiteModelPrice[]>([])
const overrideIds = ref<Record<string, number>>({})
const draft = reactive<Record<string, number>>({})
const loading = ref(false)
const saving = ref(false)
const error = ref('')

const grouped = computed<GroupRows[]>(() => {
  const map = new Map<number, GroupRows>()
  for (const row of rows.value) {
    let group = map.get(row.group_id)
    if (!group) {
      group = { id: row.group_id, name: row.group_name, mainMultiplier: row.main_multiplier, rows: [] }
      map.set(row.group_id, group)
    }
    group.rows.push(row)
  }
  return Array.from(map.values())
})

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
    const [models, prices] = await Promise.all([listSubsiteModels(), listSubsitePrices()])
    rows.value = models
    overrideIds.value = {}
    for (const price of prices) {
      if (price.scope === 'model' && price.model) {
        overrideIds.value[price.model] = price.id
      }
    }
    for (const row of models) {
      draft[row.model] = row.subsite_multiplier ?? row.main_multiplier
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载模型失败'
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
