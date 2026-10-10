<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">上游管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            管理主站同步过来的渠道。主站倍率与生图一口价只读，子站价格不能低于主站价格。
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
        主站还没有可用的渠道
      </div>

      <section
        v-for="row in rows"
        :key="row.groupId"
        class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900"
      >
        <div class="flex flex-wrap items-center justify-between gap-3">
          <label class="flex items-center gap-2">
            <input v-model="row.enabled" type="checkbox" />
            <span class="text-base font-semibold text-gray-900 dark:text-white">{{ row.name }}</span>
            <span class="rounded bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-800">{{ row.platform }}</span>
          </label>
          <button
            class="btn btn-primary"
            :disabled="saving || !rowValid(row)"
            @click="saveChannel(row)"
          >
            保存
          </button>
        </div>

        <div class="mt-4 grid gap-4 lg:grid-cols-[220px_minmax(0,1fr)]">
          <div>
            <p class="text-xs font-medium text-gray-500">加价倍率（token 模型，1 = 与主站同价）</p>
            <div class="mt-2 flex items-center gap-2 text-sm">
              <span class="text-gray-500">主站</span>
              <span class="rounded bg-gray-100 px-2 py-1 font-mono text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                ×{{ row.mainRate }}
              </span>
            </div>
            <label class="mt-2 flex items-center gap-2 text-sm">
              <span class="text-gray-500">子站</span>
              <input
                v-model.number="row.subRate"
                class="input w-24"
                type="number"
                step="0.01"
                :min="1"
                :disabled="!row.enabled"
                :class="{ 'input-error': row.enabled && row.subRate < 1 }"
              />
            </label>
            <p v-if="row.enabled && row.subRate < 1" class="mt-1 text-xs text-red-500">
              不能低于 1（低于 1 会低于主站价格）
            </p>
          </div>

          <div>
            <p class="text-xs font-medium text-gray-500">生图一口价（USD / 张）</p>
            <div class="mt-2 grid gap-3 sm:grid-cols-3">
              <div v-for="tier in tiers" :key="tier.key">
                <span class="text-xs text-gray-500">{{ tier.label }}</span>
                <div class="mt-1 flex items-center gap-2">
                  <span class="rounded bg-gray-100 px-2 py-1 font-mono text-xs text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                    {{ formatPrice(row.main[tier.key]) }}
                  </span>
                  <span class="text-gray-400">→</span>
                  <input
                    v-model.number="row.sub[tier.key]"
                    class="input w-24"
                    type="number"
                    step="0.01"
                    :min="row.main[tier.key] ?? 0"
                    :disabled="!row.enabled"
                    :class="{ 'input-error': row.enabled && belowFloor(row.sub[tier.key], row.main[tier.key]) }"
                  />
                </div>
              </div>
            </div>
            <p v-if="row.enabled && !imageRowValid(row)" class="mt-1 text-xs text-red-500">
              子站生图价格不能低于主站价格
            </p>
          </div>
        </div>
      </section>

      <section class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-200">模型价格（可选）</h2>
        <p class="mt-1 mb-3 text-xs text-gray-500 dark:text-gray-400">
          针对单个模型覆盖倍率，优先于渠道倍率，同样不能低于主站价格。
        </p>

        <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full text-sm">
            <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
              <tr>
                <th class="px-3 py-2">模型</th>
                <th class="px-3 py-2">子站倍率</th>
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
                  <button class="ml-3 text-red-500 hover:underline" :disabled="saving" @click="removeModelPrice(price)">
                    删除
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="mt-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_120px_auto]">
          <input v-model="newModel" class="input" placeholder="新增模型，例如 gpt-image-1" />
          <input v-model.number="newMultiplier" class="input" type="number" step="0.01" min="1" placeholder="倍率" />
          <button class="btn btn-primary" :disabled="saving || !newModel.trim() || newMultiplier < 1" @click="addModelPrice">
            添加模型
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

type TierKey = '1k' | '2k' | '4k'

interface ChannelRow {
  groupId: number
  name: string
  platform: string
  mainRate: number
  subRate: number
  main: Record<TierKey, number | null>
  sub: Record<TierKey, number | null>
  enabled: boolean
  overrideId: number | null
}

const tiers: Array<{ key: TierKey; label: string }> = [
  { key: '1k', label: '1K' },
  { key: '2k', label: '2K' },
  { key: '4k', label: '4K' }
]

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

function belowFloor(sub: number | null, main: number | null): boolean {
  if (sub === null || main === null) return false
  return sub + 1e-9 < main
}

function imageRowValid(row: ChannelRow): boolean {
  return tiers.every((tier) => !belowFloor(row.sub[tier.key], row.main[tier.key]))
}

function rowValid(row: ChannelRow): boolean {
  if (!row.enabled) return true
  return row.subRate + 1e-9 >= 1 && imageRowValid(row)
}

function formatPrice(value: number | null): string {
  if (value === null || value === undefined) return '—'
  return `$${Number(value).toFixed(4)}`
}

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
        mainRate: channel.main_rate_multiplier,
        subRate: override?.rate_multiplier ?? channel.main_rate_multiplier,
        main: {
          '1k': channel.main_image_price_1k,
          '2k': channel.main_image_price_2k,
          '4k': channel.main_image_price_4k
        },
        sub: {
          '1k': override?.image_price_1k ?? channel.main_image_price_1k,
          '2k': override?.image_price_2k ?? channel.main_image_price_2k,
          '4k': override?.image_price_4k ?? channel.main_image_price_4k
        },
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
  if (!rowValid(row)) {
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
        rate_multiplier: Number(row.subRate) || 1,
        image_price_1k: row.sub['1k'],
        image_price_2k: row.sub['2k'],
        image_price_4k: row.sub['4k']
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
