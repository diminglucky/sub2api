<template>
  <AppLayout>
    <div class="space-y-6">
      <header>
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">子站设置</h1>
        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
          本站的名称、上游价格与用户都在这里管理。价格完全以本站倍率为准。
        </p>
      </header>

      <div class="flex flex-wrap gap-2 border-b border-gray-200 dark:border-dark-700">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="-mb-px border-b-2 px-4 py-2 text-sm font-semibold transition-colors"
          :class="
            activeTab === tab.key
              ? 'border-primary-500 text-primary-600'
              : 'border-transparent text-gray-500 hover:text-gray-700'
          "
          @click="activeTab = tab.key"
        >
          {{ tab.label }}
        </button>
      </div>

      <section v-show="activeTab === 'basic'">
        <p v-if="error" class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
          {{ error }}
        </p>
        <p v-if="saved" class="mb-3 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
          已保存。刷新页面后品牌样式会全部生效。
        </p>

        <div class="rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
          <div class="grid gap-5 sm:grid-cols-2">
            <label class="grid gap-1.5">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-200">站点名称</span>
              <input v-model="form.name" class="input" placeholder="例如 Draw" />
            </label>
            <label class="grid gap-1.5">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-200">主题色</span>
              <span class="flex items-center gap-2">
                <input v-model="form.theme_color" type="color" class="h-10 w-14 cursor-pointer rounded border border-gray-200 bg-white" />
                <input v-model="form.theme_color" class="input" placeholder="#fb6415" />
              </span>
            </label>
            <label class="grid gap-1.5 sm:col-span-2">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-200">Logo 地址</span>
              <span class="flex items-center gap-3">
                <input v-model="form.logo_url" class="input" placeholder="例如 /draw-logo.svg" />
                <span class="flex h-10 w-10 flex-shrink-0 items-center justify-center overflow-hidden rounded-lg border border-gray-200 bg-gray-50">
                  <img v-if="form.logo_url" :src="form.logo_url" alt="Logo 预览" class="h-full w-full object-contain" />
                </span>
              </span>
            </label>
          </div>
          <div class="mt-6 flex justify-end">
            <button class="btn btn-primary" :disabled="saving || !form.name.trim()" @click="save">
              {{ saving ? '保存中...' : '保存设置' }}
            </button>
          </div>
        </div>

        <div class="mt-5 rounded-xl border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
          <h2 class="mb-3 text-sm font-semibold text-gray-700 dark:text-gray-200">站点信息（只读）</h2>
          <dl class="grid gap-3 text-sm sm:grid-cols-2">
            <div>
              <dt class="text-gray-500">子站标识</dt>
              <dd class="mt-1 font-mono text-gray-900 dark:text-white">{{ settings?.slug || '-' }}</dd>
            </div>
            <div>
              <dt class="text-gray-500">域名</dt>
              <dd class="mt-1 font-mono text-gray-900 dark:text-white">{{ settings?.domain || '-' }}</dd>
            </div>
            <div>
              <dt class="text-gray-500">状态</dt>
              <dd class="mt-1 text-gray-900 dark:text-white">
                {{ settings?.status === 'active' ? '已启用' : '已停用' }}
              </dd>
            </div>
          </dl>
        </div>
      </section>

      <section v-show="activeTab === 'upstream'">
        <SubsiteUpstreamPanel />
      </section>

      <section v-show="activeTab === 'users'">
        <SubsiteUsersPanel />
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import SubsiteUpstreamPanel from './SubsiteUpstreamPanel.vue'
import SubsiteUsersPanel from './SubsiteUsersPanel.vue'
import { getSubsiteSettings, updateSubsiteSettings, type SubsiteSettings } from '@/api/subsiteAdmin'

const tabs = [
  { key: 'basic', label: '基本设置' },
  { key: 'upstream', label: '上游管理' },
  { key: 'users', label: '用户管理' }
] as const

const activeTab = ref<(typeof tabs)[number]['key']>('basic')
const settings = ref<SubsiteSettings | null>(null)
const form = reactive({ name: '', logo_url: '', theme_color: '#fb6415' })
const saving = ref(false)
const error = ref('')
const saved = ref(false)

onMounted(load)

async function load() {
  error.value = ''
  try {
    const data = await getSubsiteSettings()
    settings.value = data
    form.name = data.name
    form.logo_url = data.logo_url
    form.theme_color = data.theme_color || '#fb6415'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载子站设置失败'
  }
}

async function save() {
  if (!form.name.trim()) return
  saving.value = true
  error.value = ''
  saved.value = false
  try {
    const data = await updateSubsiteSettings({
      name: form.name.trim(),
      logo_url: form.logo_url.trim(),
      theme_color: form.theme_color.trim()
    })
    settings.value = data
    saved.value = true
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存失败'
  } finally {
    saving.value = false
  }
}
</script>
