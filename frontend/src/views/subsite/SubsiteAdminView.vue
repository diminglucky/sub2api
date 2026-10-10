<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">子站管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            管理本站的分组价格与用量。这里只能操作当前子站，看不到主站后台。
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

      <section class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
        <h2 class="text-sm font-semibold text-gray-700 dark:text-gray-200">价格设置</h2>
        <p class="mt-1 mb-3 text-xs text-gray-500 dark:text-gray-400">
          分组价格由主站分配到本站后在此显示；倍率 1 表示与主站价格相同。模型价格优先于分组价格。
        </p>

        <div v-if="!prices.length && !loading" class="py-10 text-center text-sm text-gray-400">
          主站还没有给本站分配任何分组
        </div>

        <div v-else class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
          <table class="min-w-full text-sm">
            <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
              <tr>
                <th class="px-3 py-2">范围</th>
                <th class="px-3 py-2">目标</th>
                <th class="px-3 py-2">倍率</th>
                <th class="px-3 py-2 text-right">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="price in prices" :key="price.id" class="border-t border-gray-100 dark:border-dark-700">
                <td class="px-3 py-2">{{ price.scope === 'group' ? '分组' : '模型' }}</td>
                <td class="px-3 py-2">{{ price.model || `分组 #${price.group_id}` }}</td>
                <td class="px-3 py-2">
                  <input
                    v-model.number="draft[price.id]"
                    class="input w-24"
                    type="number"
                    step="0.01"
                  />
                </td>
                <td class="px-3 py-2 text-right">
                  <button class="text-primary-600 hover:underline" :disabled="saving" @click="savePrice(price)">
                    保存
                  </button>
                  <button
                    v-if="price.scope === 'model'"
                    class="ml-3 text-red-500 hover:underline"
                    :disabled="saving"
                    @click="removePrice(price)"
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
          <input v-model.number="newMultiplier" class="input" type="number" step="0.01" placeholder="倍率" />
          <button class="btn btn-primary" :disabled="saving || !newModel.trim()" @click="addModelPrice">
            添加模型价格
          </button>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  deleteSubsitePrice,
  getSubsiteAdminSummary,
  listSubsitePrices,
  updateSubsitePrice,
  createSubsitePrice,
  type SubsiteAdminSummary,
  type SubsitePriceOverride
} from '@/api/subsiteAdmin'

const summary = ref<SubsiteAdminSummary | null>(null)
const prices = ref<SubsitePriceOverride[]>([])
const draft = reactive<Record<number, number>>({})
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const newModel = ref('')
const newMultiplier = ref(1)

onMounted(reload)

function money(value: number | undefined): string {
  return `$${Number(value || 0).toFixed(2)}`
}

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const [siteSummary, sitePrices] = await Promise.all([
      getSubsiteAdminSummary(),
      listSubsitePrices()
    ])
    summary.value = siteSummary
    prices.value = sitePrices
    for (const price of sitePrices) {
      draft[price.id] = price.rate_multiplier
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载子站数据失败'
  } finally {
    loading.value = false
  }
}

async function savePrice(price: SubsitePriceOverride) {
  saving.value = true
  error.value = ''
  try {
    await updateSubsitePrice(price.id, {
      scope: price.scope,
      group_id: price.group_id,
      model: price.model,
      rate_multiplier: Number(draft[price.id]) || 1
    })
    await reload()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    saving.value = false
  }
}

async function removePrice(price: SubsitePriceOverride) {
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
  if (!newModel.value.trim()) return
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
