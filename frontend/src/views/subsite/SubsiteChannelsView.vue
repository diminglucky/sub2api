<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">上游管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            管理主站同步过来的上游渠道。主站价格只读，子站价格不能低于主站价格。
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

      <section class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-200">渠道价格</h2>
        <p class="mt-1 mb-3 text-xs text-gray-500 dark:text-gray-400">
          勾选要开放给本站的渠道并填写子站价格。主站价格由主站维护，不能修改；子站价格不能低于主站价格。
        </p>

        <div v-if="!rows.length && !loading" class="py-10 text-center text-sm text-gray-400">
          主站还没有可用的渠道
        </div>

        <div v-else class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full text-sm">
            <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
              <tr>
                <th class="px-3 py-2">开放</th>
                <th class="px-3 py-2">渠道</th>
                <th class="px-3 py-2">平台</th>
                <th class="px-3 py-2">主站价格</th>
                <th class="px-3 py-2">子站价格</th>
                <th class="px-3 py-2 text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in rows" :key="row.groupId" class="border-t border-gray-100 dark:border-dark-700">
                <td class="px-3 py-2">
                  <input v-model="row.enabled" type="checkbox" />
                </td>
                <td class="px-3 py-2">{{ row.name }}</td>
                <td class="px-3 py-2 text-gray-500">{{ row.platform }}</td>
                <td class="px-3 py-2">
                  <span class="rounded bg-gray-100 px-2 py-1 font-mono text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                    {{ row.main }}
                  </span>
                </td>
                <td class="px-3 py-2">
                  <input
                    v-model.number="row.sub"
                    class="input w-28"
                    type="number"
                    step="0.01"
                    :min="row.main"
                    :disabled="!row.enabled"
                    :class="{ 'input-error': row.enabled && row.sub < row.main }"
                  />
                  <p v-if="row.enabled && row.sub < row.main" class="mt-1 text-xs text-red-500">
                    不能低于主站价格
                  </p>
                </td>
                <td class="px-3 py-2 text-right">
                  <button
                    class="text-primary-600 hover:underline disabled:opacity-40"
                    :disabled="saving || (row.enabled && row.sub < row.main)"
                    @click="saveChannel(row)"
                  >
                    保存
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-200">模型价格（可选）</h2>
        <p class="mt-1 mb-3 text-xs text-gray-500 dark:text-gray-400">
          针对单个模型设置更细的价格，优先于渠道价格，同样不能低于主站价格。
        </p>

        <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full text-sm">
            <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
              <tr>
                <th class="px-3 py-2">模型</th>
                <th class="px-3 py-2">子站价格</th>
                <th class="px-3 py-2 text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!modelOverrides.length">
                <td colspan="3" class="px-3 py-6 text-center text-gray-400">暂无模型价格</td>
              </tr>
              <tr v-for="price in modelOverrides" :key="price.id" class="border-t border-gray-100 dark:border-dark-700">
                <td class="px-3 py-2">{{ price.model }}</td>
                <td class="px-3 py-2">
                  <input v-model.number="draft[price.id]" class="input w-24" type="number" step="0.01" min="1" />
                </td>
                <td class="px-3 py-2 text-right">
                  <button class="text-primary-600 hover:underline" :disabled="saving" @click="saveModelPrice(price)">
                    保存
                  </button>
                  <button
                    class="ml-3 text-red-500 hover:underline"
                    :disabled="saving"
                    @click="removeModelPrice(price)"
                  >
                    删除
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="mt-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_120px_auto]">
          <input v-model="newModel" class="input" placeholder="新增模型价格，例如 gpt-image-1" />
          <input v-model.number="newMultiplier" class="input" type="number" step="0.01" min="1" placeholder="价格" />
          <button class="btn btn-primary" :disabled="saving || !newModel.trim() || newMultiplier < 1" @click="addModelPrice">
            添加模型价格
          </button>
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
  deleteSubsitePrice,
  listSubsiteChannels,
  listSubsitePrices,
  updateSubsitePrice,
  type SubsitePriceOverride
} from '@/api/subsiteAdmin'

interface ChannelRow {
  groupId: number
  name: string
  platform: string
  main: number
  sub: number
  enabled: boolean
  overrideId: number | null
}

const rows = ref<ChannelRow[]>([])
const prices = ref<SubsitePriceOverride[]>([])
const draft = reactive<Record<number, number>>({})
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const newModel = ref('')
const newMultiplier = ref(1)

const modelOverrides = computed(() => prices.value.filter((price) => price.scope === 'model'))

onMounted(reload)

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const [channels, sitePrices] = await Promise.all([listSubsiteChannels(), listSubsitePrices()])
    prices.value = sitePrices
    const groupOverrides = new Map(
      sitePrices
        .filter((price) => price.scope === 'group' && price.group_id)
        .map((price) => [price.group_id as number, price])
    )
    rows.value = channels.map((channel) => {
      const override = groupOverrides.get(channel.group_id)
      return {
        groupId: channel.group_id,
        name: channel.name,
        platform: channel.platform,
        main: channel.main_rate_multiplier,
        sub: override?.rate_multiplier ?? channel.main_rate_multiplier,
        enabled: Boolean(override),
        overrideId: override?.id ?? null
      }
    })
    for (const price of sitePrices) {
      draft[price.id] = price.rate_multiplier
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载渠道失败'
  } finally {
    loading.value = false
  }
}

async function saveChannel(row: ChannelRow) {
  if (row.enabled && row.sub < row.main) {
    error.value = '子站价格不能低于主站价格'
    return
  }
  saving.value = true
  error.value = ''
  try {
    if (row.enabled) {
      const payload = {
        scope: 'group' as const,
        group_id: row.groupId,
        rate_multiplier: Number(row.sub) || row.main
      }
      if (row.overrideId) {
        await updateSubsitePrice(row.overrideId, payload)
      } else {
        await createSubsitePrice(payload)
      }
    } else if (row.overrideId) {
      await deleteSubsitePrice(row.overrideId)
    }
    await reload()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function saveModelPrice(price: SubsitePriceOverride) {
  const value = Number(draft[price.id]) || 1
  if (value < 1) {
    error.value = '子站价格不能低于主站价格'
    return
  }
  saving.value = true
  error.value = ''
  try {
    await updateSubsitePrice(price.id, { scope: 'model', model: price.model, rate_multiplier: value })
    await reload()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function removeModelPrice(price: SubsitePriceOverride) {
  saving.value = true
  error.value = ''
  try {
    await deleteSubsitePrice(price.id)
    await reload()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除失败'
  } finally {
    saving.value = false
  }
}

async function addModelPrice() {
  if (!newModel.value.trim() || newMultiplier.value < 1) return
  saving.value = true
  error.value = ''
  try {
    await createSubsitePrice({
      scope: 'model',
      model: newModel.value.trim(),
      rate_multiplier: Number(newMultiplier.value) || 1
    })
    newModel.value = ''
    newMultiplier.value = 1
    await reload()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '添加失败'
  } finally {
    saving.value = false
  }
}
</script>
