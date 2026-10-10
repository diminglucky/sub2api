<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">子站管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            选择哪些分组开放给子站，并为每个分组设置子站价格。子站用户页面复用主站，差异只在价格与入口页。
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

      <div class="grid gap-6 lg:grid-cols-[320px_minmax(0,1fr)]">
        <section class="rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <h2 class="mb-3 text-sm font-semibold text-gray-700 dark:text-gray-200">子站列表</h2>
          <div v-if="!subsites.length && !loading" class="py-6 text-center text-sm text-gray-400">
            还没有子站
          </div>
          <ul class="space-y-2">
            <li v-for="item in subsites" :key="item.id">
              <button
                type="button"
                class="w-full rounded-lg border px-3 py-2 text-left transition-colors"
                :class="
                  selected?.id === item.id
                    ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                    : 'border-gray-200 hover:border-primary-300 dark:border-dark-700'
                "
                @click="selectSubsite(item)"
              >
                <span class="block text-sm font-semibold text-gray-900 dark:text-white">
                  {{ item.name || item.slug }}
                </span>
                <span class="block text-xs text-gray-500 dark:text-gray-400">{{ item.domain }}</span>
                <span
                  class="mt-1 inline-block rounded-full px-2 py-0.5 text-[11px] font-semibold"
                  :class="
                    item.status === 'active'
                      ? 'bg-emerald-100 text-emerald-700'
                      : 'bg-gray-100 text-gray-500'
                  "
                >
                  {{ item.status === 'active' ? '已启用' : '已停用' }}
                </span>
              </button>
            </li>
          </ul>

          <div class="mt-5 border-t border-gray-100 pt-4 dark:border-dark-700">
            <button type="button" class="text-sm font-semibold text-primary-600" @click="showCreate = !showCreate">
              {{ showCreate ? '收起' : '新建子站' }}
            </button>
            <div v-if="showCreate" class="mt-3 space-y-2">
              <input v-model="createForm.slug" class="input" placeholder="slug，例如 draw" />
              <input v-model="createForm.domain" class="input" placeholder="域名，例如 draw.superai.sbs" />
              <input v-model="createForm.name" class="input" placeholder="名称，例如 Draw" />
              <input v-model="createForm.theme_color" class="input" placeholder="主题色，例如 #fb6415" />
              <button class="btn btn-primary w-full" :disabled="saving" @click="createSubsite">
                创建子站
              </button>
            </div>
          </div>
        </section>

        <section class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
          <div v-if="!selected" class="py-16 text-center text-sm text-gray-400">
            从左侧选择一个子站来管理分组与价格
          </div>
          <template v-else>
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ selected.name }}</h2>
                <p class="text-sm text-gray-500 dark:text-gray-400">{{ selected.domain }}</p>
              </div>
              <button
                class="btn btn-secondary"
                :disabled="selected.status !== 'active' || saving"
                @click="disableSelected"
              >
                停用子站
              </button>
            </div>

            <div class="mt-6">
              <h3 class="mb-2 text-sm font-semibold text-gray-700 dark:text-gray-200">子站分组</h3>
              <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
                勾选要开放给该子站的分组，并填写子站倍率。倍数 1 表示与主站价格相同。
              </p>

              <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
                <table class="min-w-full text-sm">
                  <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
                    <tr>
                      <th class="px-3 py-2">加入</th>
                      <th class="px-3 py-2">分组</th>
                      <th class="px-3 py-2">平台</th>
                      <th class="px-3 py-2">主站倍率</th>
                      <th class="px-3 py-2">子站倍率</th>
                      <th class="px-3 py-2 text-right">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-if="!groupRows.length" class="border-t border-gray-100 dark:border-dark-700">
                      <td colspan="6" class="px-3 py-6 text-center text-gray-400">没有可用分组</td>
                    </tr>
                    <tr
                      v-for="row in groupRows"
                      :key="row.id"
                      class="border-t border-gray-100 dark:border-dark-700"
                    >
                      <td class="px-3 py-2">
                        <input v-model="row.enabled" type="checkbox" />
                      </td>
                      <td class="px-3 py-2">{{ row.name }}</td>
                      <td class="px-3 py-2 text-gray-500">{{ row.platform }}</td>
                      <td class="px-3 py-2 text-gray-500">{{ row.mainMultiplier }}</td>
                      <td class="px-3 py-2">
                        <input
                          v-model.number="row.multiplier"
                          class="input w-24"
                          type="number"
                          step="0.01"
                          :disabled="!row.enabled"
                        />
                      </td>
                      <td class="px-3 py-2 text-right">
                        <button class="text-primary-600 hover:underline" :disabled="saving" @click="saveGroup(row)">
                          保存
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="mt-6">
              <h3 class="mb-2 text-sm font-semibold text-gray-700 dark:text-gray-200">模型价格覆盖（可选）</h3>
              <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
                针对单个模型设置更细的价格，优先于分组倍率。
              </p>

              <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
                <table class="min-w-full text-sm">
                  <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
                    <tr>
                      <th class="px-3 py-2">模型</th>
                      <th class="px-3 py-2">倍率</th>
                      <th class="px-3 py-2">状态</th>
                      <th class="px-3 py-2 text-right">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-if="!modelOverrides.length">
                      <td colspan="4" class="px-3 py-6 text-center text-gray-400">暂无模型价格覆盖</td>
                    </tr>
                    <tr
                      v-for="price in modelOverrides"
                      :key="price.id"
                      class="border-t border-gray-100 dark:border-dark-700"
                    >
                      <td class="px-3 py-2">{{ price.model }}</td>
                      <td class="px-3 py-2">{{ price.rate_multiplier }}</td>
                      <td class="px-3 py-2">{{ price.status }}</td>
                      <td class="px-3 py-2 text-right">
                        <button class="text-red-500 hover:underline" @click="removeModelOverride(price)">删除</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div class="mt-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_120px_auto]">
                <input v-model="modelForm.model" class="input" placeholder="模型名称，例如 gpt-image-1" />
                <input
                  v-model.number="modelForm.rate_multiplier"
                  class="input"
                  type="number"
                  step="0.01"
                  placeholder="倍率"
                />
                <button class="btn btn-primary" :disabled="saving" @click="addModelOverride">添加</button>
              </div>
            </div>
          </template>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { adminAPI } from '@/api/admin'
import {
  downstreamAdminAPI,
  type DownstreamPriceOverride,
  type DownstreamSubsite
} from '@/api/admin/downstream'

interface GroupRow {
  id: number
  name: string
  platform: string
  mainMultiplier: number
  enabled: boolean
  multiplier: number
}

const subsites = ref<DownstreamSubsite[]>([])
const selected = ref<DownstreamSubsite | null>(null)
const overrides = ref<DownstreamPriceOverride[]>([])
const groupRows = ref<GroupRow[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const showCreate = ref(false)

const createForm = reactive({ slug: '', domain: '', name: '', theme_color: '#fb6415' })
const modelForm = reactive({ model: '', rate_multiplier: 1 })

const modelOverrides = computed(() => overrides.value.filter((item) => item.scope === 'model'))

onMounted(reload)

async function reload() {
  loading.value = true
  error.value = ''
  try {
    subsites.value = await downstreamAdminAPI.listSubsites()
    if (selected.value) {
      const fresh = subsites.value.find((item) => item.id === selected.value?.id)
      if (fresh) {
        selected.value = fresh
        await loadDetail()
      } else {
        selected.value = null
        overrides.value = []
        groupRows.value = []
      }
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载子站失败'
  } finally {
    loading.value = false
  }
}

async function selectSubsite(item: DownstreamSubsite) {
  selected.value = item
  await loadDetail()
}

async function loadDetail() {
  if (!selected.value) return
  error.value = ''
  try {
    const [groups, prices] = await Promise.all([
      adminAPI.groups.getAll(),
      downstreamAdminAPI.listPrices(selected.value.id)
    ])
    overrides.value = prices
    const groupPrices = new Map(
      prices
        .filter((price) => price.scope === 'group' && price.group_id)
        .map((price) => [price.group_id as number, price])
    )
    groupRows.value = groups.map((group) => {
      const override = groupPrices.get(group.id)
      return {
        id: group.id,
        name: group.name,
        platform: group.platform,
        mainMultiplier: group.rate_multiplier,
        enabled: Boolean(override),
        multiplier: override?.rate_multiplier ?? group.rate_multiplier ?? 1
      }
    })
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载分组或价格失败'
  }
}

function overrideForGroup(groupId: number): DownstreamPriceOverride | undefined {
  return overrides.value.find((price) => price.scope === 'group' && price.group_id === groupId)
}

async function saveGroup(row: GroupRow) {
  if (!selected.value) return
  saving.value = true
  error.value = ''
  const existing = overrideForGroup(row.id)
  try {
    if (row.enabled) {
      const payload = {
        scope: 'group' as const,
        group_id: row.id,
        rate_multiplier: Number(row.multiplier) || 1
      }
      if (existing) {
        await downstreamAdminAPI.updatePrice(selected.value.id, existing.id, payload)
      } else {
        await downstreamAdminAPI.createPrice(selected.value.id, payload)
      }
    } else if (existing) {
      await downstreamAdminAPI.deletePrice(selected.value.id, existing.id)
    }
    await loadDetail()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存分组价格失败'
  } finally {
    saving.value = false
  }
}

async function addModelOverride() {
  if (!selected.value || !modelForm.model.trim()) return
  saving.value = true
  error.value = ''
  try {
    await downstreamAdminAPI.createPrice(selected.value.id, {
      scope: 'model',
      model: modelForm.model.trim(),
      rate_multiplier: Number(modelForm.rate_multiplier) || 1
    })
    modelForm.model = ''
    modelForm.rate_multiplier = 1
    await loadDetail()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存模型价格失败'
  } finally {
    saving.value = false
  }
}

async function removeModelOverride(price: DownstreamPriceOverride) {
  if (!selected.value) return
  saving.value = true
  error.value = ''
  try {
    await downstreamAdminAPI.deletePrice(selected.value.id, price.id)
    await loadDetail()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除模型价格失败'
  } finally {
    saving.value = false
  }
}

async function createSubsite() {
  saving.value = true
  error.value = ''
  try {
    await downstreamAdminAPI.createSubsite({ ...createForm })
    createForm.slug = ''
    createForm.domain = ''
    createForm.name = ''
    showCreate.value = false
    await reload()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '创建子站失败'
  } finally {
    saving.value = false
  }
}

async function disableSelected() {
  if (!selected.value) return
  saving.value = true
  error.value = ''
  try {
    await downstreamAdminAPI.disableSubsite(selected.value.id)
    await reload()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '停用子站失败'
  } finally {
    saving.value = false
  }
}
</script>
