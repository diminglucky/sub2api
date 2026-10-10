<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">子站管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            选择哪些分组开放给子站。子站价格完全跟随主站，无需单独设置倍率或模型。
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
              <input v-model="createForm.logo_url" class="input" placeholder="Logo 地址，例如 /draw-logo.svg" />
              <input v-model="createForm.theme_color" class="input" placeholder="主题色，例如 #fb6415" />
              <button class="btn btn-primary w-full" :disabled="saving" @click="createSubsite">
                创建子站
              </button>
            </div>
          </div>
        </section>

        <section class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
          <div v-if="!selected" class="py-16 text-center text-sm text-gray-400">
            从左侧选择一个子站来分配分组
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
              <h3 class="mb-2 text-sm font-semibold text-gray-700 dark:text-gray-200">开放分组</h3>
              <p class="mb-3 text-xs text-gray-500 dark:text-gray-400">
                勾选后该分组对子站开放，价格与主站完全一致（含模型价格与生图一口价）。
              </p>

              <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
                <table class="min-w-full text-sm">
                  <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
                    <tr>
                      <th class="px-3 py-2">开放</th>
                      <th class="px-3 py-2">分组</th>
                      <th class="px-3 py-2">平台</th>
                      <th class="px-3 py-2">主站倍率</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-if="!groups.length" class="border-t border-gray-100 dark:border-dark-700">
                      <td colspan="4" class="px-3 py-6 text-center text-gray-400">没有可用分组</td>
                    </tr>
                    <tr v-for="row in groups" :key="row.group_id" class="border-t border-gray-100 dark:border-dark-700">
                      <td class="px-3 py-2">
                        <input
                          type="checkbox"
                          :checked="row.assigned"
                          :disabled="saving"
                          @change="toggleGroup(row, ($event.target as HTMLInputElement).checked)"
                        />
                      </td>
                      <td class="px-3 py-2">{{ row.name }}</td>
                      <td class="px-3 py-2 text-gray-500">{{ row.platform }}</td>
                      <td class="px-3 py-2">
                        <span class="rounded bg-gray-100 px-2 py-1 font-mono text-gray-600 dark:bg-dark-800 dark:text-gray-300">
                          ×{{ row.main_rate_multiplier }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
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
  type DownstreamGroupAssignment,
  type DownstreamSubsite
} from '@/api/admin/downstream'

const subsites = ref<DownstreamSubsite[]>([])
const selected = ref<DownstreamSubsite | null>(null)
const groups = ref<DownstreamGroupAssignment[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref('')
const showCreate = ref(false)

const createForm = reactive({ slug: '', domain: '', name: '', logo_url: '', theme_color: '#fb6415' })

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
        await loadGroups()
      } else {
        selected.value = null
        groups.value = []
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
  await loadGroups()
}

async function loadGroups() {
  if (!selected.value) return
  error.value = ''
  try {
    groups.value = await downstreamAdminAPI.listGroups(selected.value.id)
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载分组失败'
  }
}

async function toggleGroup(row: DownstreamGroupAssignment, assigned: boolean) {
  if (!selected.value) return
  saving.value = true
  error.value = ''
  try {
    if (assigned) {
      await downstreamAdminAPI.assignGroup(selected.value.id, row.group_id)
    } else {
      await downstreamAdminAPI.unassignGroup(selected.value.id, row.group_id)
    }
    await loadGroups()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败'
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
    createForm.logo_url = ''
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
