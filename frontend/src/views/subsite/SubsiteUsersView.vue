<template>
  <AppLayout>
    <div class="space-y-6">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">用户管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            本站的用户列表。余额与账号信息与主站共享，用量和充值只统计本站数据。
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
        <div class="mb-3 flex items-center gap-2">
          <input v-model="keyword" class="input max-w-xs" placeholder="搜索邮箱" />
          <span class="text-xs text-gray-500">共 {{ users.length }} 位用户</span>
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
                <th class="px-3 py-2">本站充值</th>
                <th class="px-3 py-2">加入时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!filtered.length">
                <td colspan="7" class="px-3 py-6 text-center text-gray-400">
                  {{ loading ? '加载中...' : '暂无用户' }}
                </td>
              </tr>
              <tr v-for="user in filtered" :key="user.user_id" class="border-t border-gray-100 dark:border-dark-700">
                <td class="px-3 py-2">{{ user.email }}</td>
                <td class="px-3 py-2">
                  <span
                    class="rounded-full px-2 py-0.5 text-[11px] font-semibold"
                    :class="
                      user.role === 'owner' || user.role === 'admin'
                        ? 'bg-primary-50 text-primary-600'
                        : 'bg-gray-100 text-gray-500'
                    "
                  >
                    {{ roleLabel(user.role) }}
                  </span>
                </td>
                <td class="px-3 py-2">{{ money(user.balance) }}</td>
                <td class="px-3 py-2">{{ user.usage_count }}</td>
                <td class="px-3 py-2">{{ money(user.usage_cost) }}</td>
                <td class="px-3 py-2">{{ money(user.recharge_amount) }}</td>
                <td class="px-3 py-2 text-gray-500">{{ formatDate(user.joined_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
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
