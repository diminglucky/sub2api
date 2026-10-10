<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">子站管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            管理下游子站的品牌配置和价格覆盖。子站用户页面复用主站，差异只在价格与入口页。
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
            从左侧选择一个子站来管理价格
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
              <h3 class="mb-2 text-sm font-semibold text-gray-700 dark:text-gray-200">价格覆盖</h3>
              <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
                默认价格等于主站价格；这里设置的价格只影响该子站的用户扣费和结算差额。
              </p>

              <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
                <table class="min-w-full text-sm">
                  <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
                    <tr>
                      <th class="px-3 py-2">范围</th>
                      <th class="px-3 py-2">目标</th>
                      <th class="px-3 py-2">倍率</th>
                      <th class="px-3 py-2">状态</th>
                      <th class="px-3 py-2 text-right">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-if="!prices.length">
                      <td colspan="5" class="px-3 py-6 text-center text-gray-400">暂无价格覆盖</td>
                    </tr>
                    <tr v-for="price in prices" :key="price.id" class="border-t border-gray-100 dark:border-dark-700">
                      <td class="px-3 py-2">{{ price.scope === 'group' ? '分组' : '模型' }}</td>
                      <td class="px-3 py-2">{{ price.model || `分组 #${price.group_id}` }}</td>
                      <td class="px-3 py-2">{{ price.rate_multiplier }}</td>
                      <td class="px-3 py-2">{{ price.status }}</td>
                      <td class="px-3 py-2 text-right">
                        <button class="text-primary-600 hover:underline" @click="startEdit(price)">编辑</button>
                        <button class="ml-3 text-red-500 hover:underline" @click="removePrice(price)">删除</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <div class="mt-4 grid gap-3 sm:grid-cols-[140px_minmax(0,1fr)_120px_auto]">
                <select v-model="priceForm.scope" class="input">
                  <option value="group">分组</option>
                  <option value="model">模型</option>
                </select>
                <input
                  v-if="priceForm.scope === 'model'"
                  v-model="priceForm.model"
                  class="input"
                  placeholder="模型名称，例如 gpt-image-1"
                />
                <input
                  v-else
                  v-model.number="priceForm.group_id"
                  class="input"
                  type="number"
                  placeholder="分组 ID"
                />
                <input
                  v-model.number="priceForm.rate_multiplier"
                  class="input"
                  type="number"
                  step="0.01"
                  placeholder="倍率"
                />
                <button class="btn btn-primary" :disabled="saving" @click="submitPrice">
                  {{ editingPriceId ? '保存' : '添加' }}
                </button>
              </div>
              <button
                v-if="editingPriceId"
                class="mt-2 text-xs text-gray-500 hover:underline"
                @click="cancelEdit"
              >
                取消编辑
              </button>
            </div>
          </template>
        </section>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  downstreamAdminAPI,
  type DownstreamPriceOverride,
  type DownstreamSubsite
} from '@/api/admin/downstream'

const subsites = ref<DownstreamSubsite[]>([])
const selected = ref<DownstreamSubsite | null>(null)
const prices = ref<DownstreamPriceOverride[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const showCreate = ref(false)
const editingPriceId = ref<number | null>(null)

const createForm = reactive({ slug: '', domain: '', name: '', theme_color: '#fb6415' })
const priceForm = reactive({ scope: 'group' as 'group' | 'model', group_id: null as number | null, model: '', rate_multiplier: 1 })

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
        await loadPrices()
      } else {
        selected.value = null
        prices.value = []
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
  editingPriceId.value = null
  await loadPrices()
}

async function loadPrices() {
  if (!selected.value) return
  error.value = ''
  try {
    prices.value = await downstreamAdminAPI.listPrices(selected.value.id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载价格失败'
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

function startEdit(price: DownstreamPriceOverride) {
  editingPriceId.value = price.id
  priceForm.scope = price.scope
  priceForm.group_id = price.group_id
  priceForm.model = price.model || ''
  priceForm.rate_multiplier = price.rate_multiplier
}

function cancelEdit() {
  editingPriceId.value = null
  priceForm.scope = 'group'
  priceForm.group_id = null
  priceForm.model = ''
  priceForm.rate_multiplier = 1
}

async function submitPrice() {
  if (!selected.value) return
  saving.value = true
  error.value = ''
  const payload = {
    scope: priceForm.scope,
    group_id: priceForm.scope === 'group' ? priceForm.group_id : null,
    model: priceForm.scope === 'model' ? priceForm.model : null,
    rate_multiplier: Number(priceForm.rate_multiplier) || 1
  }
  try {
    if (editingPriceId.value) {
      await downstreamAdminAPI.updatePrice(selected.value.id, editingPriceId.value, payload)
    } else {
      await downstreamAdminAPI.createPrice(selected.value.id, payload)
    }
    cancelEdit()
    await loadPrices()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存价格失败'
  } finally {
    saving.value = false
  }
}

async function removePrice(price: DownstreamPriceOverride) {
  if (!selected.value) return
  saving.value = true
  error.value = ''
  try {
    await downstreamAdminAPI.deletePrice(selected.value.id, price.id)
    await loadPrices()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '删除价格失败'
  } finally {
    saving.value = false
  }
}
</script>
