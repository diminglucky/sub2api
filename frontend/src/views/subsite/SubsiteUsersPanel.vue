<template>
  <div class="space-y-4">
    <p v-if="error" class="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600">
      {{ error }}
    </p>
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-2">
        <input v-model="keyword" class="input max-w-xs" placeholder="搜索邮箱" />
        <span class="text-xs text-gray-500">共 {{ users.length }} 位用户</span>
      </div>
      <button class="btn btn-secondary" :disabled="loading" @click="reload">
        <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
        <span class="ml-1">刷新</span>
      </button>
    </div>

    <div class="overflow-x-auto rounded-lg border border-gray-200 dark:border-dark-700">
      <table class="min-w-full text-sm">
        <thead class="bg-gray-50 text-left text-xs uppercase text-gray-500 dark:bg-dark-800">
          <tr>
            <th class="px-3 py-2">邮箱</th>
            <th class="px-3 py-2">身份</th>
            <th class="px-3 py-2">余额</th>
            <th class="px-3 py-2">本站调用</th>
            <th class="px-3 py-2">本站消费</th>
            <th class="px-3 py-2">加入时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!filtered.length">
            <td colspan="6" class="px-3 py-6 text-center text-gray-400">
              {{ loading ? '加载中...' : '暂无用户' }}
            </td>
          </tr>
          <tr v-for="user in filtered" :key="user.user_id" class="border-t border-gray-100 dark:border-dark-700">
            <td class="px-3 py-2">{{ user.email }}</td>
            <td class="px-3 py-2">
              <span class="rounded-full bg-gray-100 px-2 py-0.5 text-[11px] font-semibold text-gray-500">
                {{ roleLabel(user.role) }}
              </span>
            </td>
            <td class="px-3 py-2">{{ money(user.balance) }}</td>
            <td class="px-3 py-2">{{ user.usage_count }}</td>
            <td class="px-3 py-2">{{ money(user.usage_cost) }}</td>
            <td class="px-3 py-2 text-gray-500">{{ formatDate(user.joined_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Icon from '@/components/icons/Icon.vue'
import { listSubsiteUsers, type SubsiteMember } from '@/api/subsiteAdmin'

const users = ref<SubsiteMember[]>([])
const loading = ref(false)
const error = ref('')
const keyword = ref('')

const filtered = computed(() => {
  const term = keyword.value.trim().toLowerCase()
  if (!term) return users.value
  return users.value.filter((user) => user.email.toLowerCase().includes(term))
})

onMounted(reload)

async function reload() {
  loading.value = true
  error.value = ''
  try {
    const page = await listSubsiteUsers()
    users.value = page.items || []
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载用户失败'
  } finally {
    loading.value = false
  }
}

function money(value: number | undefined): string {
  return `$${Number(value || 0).toFixed(2)}`
}

function roleLabel(role: string): string {
  if (role === 'owner') return '站长'
  if (role === 'admin') return '管理员'
  return '普通用户'
}

function formatDate(value: string): string {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}
</script>
