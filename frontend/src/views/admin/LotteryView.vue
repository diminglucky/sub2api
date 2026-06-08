<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">抽奖管理</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">创建活动、配置余额或卡密奖品，到时间自动开奖。</p>
        </div>
        <div class="flex gap-2">
          <button class="btn btn-secondary" :disabled="loading" @click="loadLotteries">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            刷新
          </button>
          <button class="btn btn-primary" @click="startCreate">
            <Icon name="plus" size="md" class="mr-1" />
            新建抽奖
          </button>
        </div>
      </div>

      <div class="grid gap-5 xl:grid-cols-[minmax(0,1fr)_440px]">
        <section class="card overflow-hidden">
          <div class="border-b border-gray-100 p-4 dark:border-dark-700">
            <div class="flex flex-wrap gap-3">
              <input v-model="search" class="input max-w-xs" placeholder="搜索活动" @input="debouncedLoad" />
              <select v-model="status" class="input w-36" @change="loadLotteries">
                <option value="">全部状态</option>
                <option value="draft">草稿</option>
                <option value="active">进行中</option>
                <option value="drawn">已开奖</option>
                <option value="archived">已归档</option>
              </select>
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="min-w-full divide-y divide-gray-100 dark:divide-dark-700">
              <thead class="bg-gray-50 text-left text-xs font-medium uppercase text-gray-500 dark:bg-dark-800 dark:text-dark-400">
                <tr>
                  <th class="px-4 py-3">活动</th>
                  <th class="px-4 py-3">状态</th>
                  <th class="px-4 py-3">开奖时间</th>
                  <th class="px-4 py-3">参与</th>
                  <th class="px-4 py-3 text-right">操作</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-if="loading">
                  <td colspan="5" class="px-4 py-10 text-center text-sm text-gray-500">加载中...</td>
                </tr>
                <tr v-else-if="events.length === 0">
                  <td colspan="5" class="px-4 py-10 text-center text-sm text-gray-500">暂无抽奖活动</td>
                </tr>
                <tr v-for="event in events" v-else :key="event.id" class="hover:bg-gray-50 dark:hover:bg-dark-800/60">
                  <td class="px-4 py-3">
                    <div class="font-medium text-gray-900 dark:text-white">{{ event.title }}</div>
                    <div class="mt-1 max-w-md truncate text-xs text-gray-500 dark:text-dark-400">#{{ event.id }} {{ event.description }}</div>
                  </td>
                  <td class="px-4 py-3"><span class="badge" :class="statusClass(event.status)">{{ statusLabel(event.status) }}</span></td>
                  <td class="px-4 py-3 text-sm text-gray-600 dark:text-dark-300">{{ formatTs(event.draw_at) }}</td>
                  <td class="px-4 py-3 text-sm text-gray-600 dark:text-dark-300">{{ event.entry_count }}</td>
                  <td class="px-4 py-3">
                    <div class="flex justify-end gap-1">
                      <button class="btn btn-ghost btn-sm" @click="editEvent(event)">编辑</button>
                      <button class="btn btn-ghost btn-sm" :disabled="event.status !== 'active'" @click="drawEvent(event)">开奖</button>
                      <button class="btn btn-ghost btn-sm text-red-600" @click="deleteEvent(event)">删除</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <aside class="space-y-5">
          <section class="card p-5">
            <div class="mb-4 flex items-center justify-between">
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ editingId ? '编辑抽奖' : '新建抽奖' }}</h2>
              <button v-if="editingId" class="btn btn-ghost btn-sm" @click="startCreate">清空</button>
            </div>

            <form class="space-y-4" @submit.prevent="saveEvent">
              <div>
                <label class="input-label">标题</label>
                <input v-model="form.title" class="input" required placeholder="例如：周末福利抽奖" />
              </div>
              <div>
                <label class="input-label">说明</label>
                <textarea v-model="form.description" class="input min-h-20" placeholder="活动说明，会展示给用户" />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div>
                  <label class="input-label">状态</label>
                  <select v-model="form.status" class="input">
                    <option value="draft">草稿</option>
                    <option value="active">进行中</option>
                    <option value="archived">归档</option>
                  </select>
                </div>
                <div>
                  <label class="input-label">开奖时间</label>
                  <input v-model="form.draw_at" class="input" type="datetime-local" required />
                </div>
              </div>
              <div>
                <label class="input-label">开始时间</label>
                <input v-model="form.starts_at" class="input" type="datetime-local" />
              </div>

              <div>
                <div class="mb-2 flex items-center justify-between">
                  <label class="input-label">奖品</label>
                  <button type="button" class="btn btn-secondary btn-sm" @click="addPrize">添加奖品</button>
                </div>
                <div class="space-y-3">
                  <div v-for="(prize, index) in form.prizes" :key="index" class="rounded-lg border border-gray-200 p-3 dark:border-dark-700">
                    <div class="mb-3 flex items-center justify-between">
                      <span class="text-sm font-medium text-gray-900 dark:text-white">奖品 {{ index + 1 }}</span>
                      <button type="button" class="text-sm text-red-600" @click="removePrize(index)">删除</button>
                    </div>
                    <div class="grid grid-cols-2 gap-2">
                      <select v-model="prize.type" class="input">
                        <option value="balance">余额</option>
                        <option value="card">卡密</option>
                      </select>
                      <input v-model.number="prize.quantity" class="input" type="number" min="1" placeholder="数量" />
                    </div>
                    <input v-model="prize.name" class="input mt-2" required placeholder="奖品名称" />
                    <input
                      v-if="prize.type === 'balance'"
                      v-model.number="prize.amount"
                      class="input mt-2"
                      type="number"
                      min="0.01"
                      step="0.01"
                      placeholder="到账金额，例如 5"
                    />
                    <textarea
                      v-else
                      v-model="prize.card_content"
                      class="input mt-2 min-h-20"
                      placeholder="卡密内容，中奖用户可见"
                    />
                  </div>
                </div>
              </div>

              <button class="btn btn-primary w-full" :disabled="saving || form.prizes.length === 0">
                <Icon v-if="saving" name="refresh" size="sm" class="mr-1 animate-spin" />
                {{ editingId ? '保存修改' : '创建活动' }}
              </button>
            </form>
          </section>

          <section v-if="selectedEvent" class="card p-5">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">中奖名单</h2>
            <div class="mt-3 space-y-2">
              <div v-if="!selectedEvent.winners?.length" class="rounded-lg bg-gray-50 p-4 text-sm text-gray-500 dark:bg-dark-800">
                暂无中奖记录
              </div>
              <div v-for="winner in selectedEvent.winners" :key="winner.id" class="rounded-lg border border-gray-100 p-3 dark:border-dark-700">
                <div class="flex items-center justify-between gap-2">
                  <div class="min-w-0">
                    <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ winner.user_email || `用户 #${winner.user_id}` }}</div>
                    <div class="text-xs text-gray-500 dark:text-dark-400">{{ winner.prize_name }}</div>
                  </div>
                  <span class="badge badge-success">{{ winner.prize_type === 'balance' ? `¥${formatMoney(winner.amount || 0)}` : '卡密' }}</span>
                </div>
                <textarea v-if="winner.card_content" class="input mt-2 min-h-16" readonly :value="winner.card_content" />
              </div>
            </div>
          </section>
        </aside>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import adminAPI from '@/api/admin'
import type { AdminLotteryPrizeRequest } from '@/api/admin/lottery'
import type { LotteryEvent } from '@/api/lottery'
import { formatDateTime, formatDateTimeLocalInput, parseDateTimeLocalInput } from '@/utils/format'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const events = ref<LotteryEvent[]>([])
const loading = ref(false)
const saving = ref(false)
const search = ref('')
const status = ref('')
const editingId = ref<number | null>(null)
const selectedEvent = computed(() => events.value.find((event) => event.id === editingId.value) || null)
let searchTimer: number | undefined

const form = reactive({
  title: '',
  description: '',
  status: 'active' as 'draft' | 'active' | 'archived',
  starts_at: '',
  draw_at: '',
  prizes: [] as AdminLotteryPrizeRequest[]
})

function formatTs(value?: number): string {
  return value ? formatDateTime(new Date(value * 1000)) : '-'
}

function formatMoney(value: number): string {
  return Number(value || 0).toFixed(2)
}

function statusLabel(value: LotteryEvent['status']): string {
  if (value === 'active') return '进行中'
  if (value === 'drawn') return '已开奖'
  if (value === 'archived') return '已归档'
  return '草稿'
}

function statusClass(value: LotteryEvent['status']): string {
  if (value === 'active') return 'badge-success'
  if (value === 'drawn') return 'badge-warning'
  return 'badge-gray'
}

function debouncedLoad(): void {
  window.clearTimeout(searchTimer)
  searchTimer = window.setTimeout(loadLotteries, 300)
}

async function loadLotteries(): Promise<void> {
  loading.value = true
  try {
    const res = await adminAPI.lottery.list(1, 50, {
      status: status.value || undefined,
      search: search.value || undefined,
      sort_by: 'created_at',
      sort_order: 'desc'
    })
    events.value = res.items
  } catch (error: any) {
    appStore.showError(error?.message || '加载抽奖活动失败')
  } finally {
    loading.value = false
  }
}

function startCreate(): void {
  editingId.value = null
  form.title = ''
  form.description = ''
  form.status = 'active'
  form.starts_at = ''
  form.draw_at = formatDateTimeLocalInput(Math.floor(Date.now() / 1000) + 3600)
  form.prizes = [{ type: 'balance', name: '¥5 余额', quantity: 1, amount: 5, sort_order: 1 }]
}

function editEvent(event: LotteryEvent): void {
  editingId.value = event.id
  form.title = event.title
  form.description = event.description || ''
  form.status = event.status === 'drawn' ? 'archived' : event.status
  form.starts_at = formatDateTimeLocalInput(event.starts_at || null)
  form.draw_at = formatDateTimeLocalInput(event.draw_at)
  form.prizes = event.prizes.map((prize, index) => ({
    type: prize.type,
    name: prize.name,
    quantity: prize.quantity,
    amount: prize.amount,
    sort_order: prize.sort_order || index + 1,
    card_content: (prize as any).card_content
  }))
}

function addPrize(): void {
  form.prizes.push({ type: 'balance', name: '¥5 余额', quantity: 1, amount: 5, sort_order: form.prizes.length + 1 })
}

function removePrize(index: number): void {
  form.prizes.splice(index, 1)
}

function buildPrizePayload(): AdminLotteryPrizeRequest[] {
  return form.prizes.map((prize, index) => ({
    type: prize.type,
    name: prize.name,
    quantity: Number(prize.quantity || 1),
    amount: prize.type === 'balance' ? Number(prize.amount || 0) : undefined,
    card_content: prize.type === 'card' ? prize.card_content : undefined,
    sort_order: index + 1
  }))
}

async function saveEvent(): Promise<void> {
  const drawAt = parseDateTimeLocalInput(form.draw_at)
  if (!drawAt) {
    appStore.showError('请选择开奖时间')
    return
  }
  saving.value = true
  try {
    const payload = {
      title: form.title,
      description: form.description,
      status: form.status,
      starts_at: parseDateTimeLocalInput(form.starts_at),
      draw_at: drawAt,
      prizes: buildPrizePayload()
    }
    if (editingId.value) {
      await adminAPI.lottery.update(editingId.value, payload)
      appStore.showSuccess('抽奖已更新')
    } else {
      const created = await adminAPI.lottery.create(payload)
      editingId.value = created.id
      appStore.showSuccess('抽奖已创建')
    }
    await loadLotteries()
  } catch (error: any) {
    appStore.showError(error?.message || '保存抽奖失败')
  } finally {
    saving.value = false
  }
}

async function drawEvent(event: LotteryEvent): Promise<void> {
  if (!window.confirm(`确定立即开奖「${event.title}」吗？`)) return
  try {
    await adminAPI.lottery.draw(event.id)
    appStore.showSuccess('开奖完成')
    await loadLotteries()
  } catch (error: any) {
    appStore.showError(error?.message || '开奖失败')
  }
}

async function deleteEvent(event: LotteryEvent): Promise<void> {
  if (!window.confirm(`确定删除「${event.title}」吗？`)) return
  try {
    await adminAPI.lottery.delete(event.id)
    if (editingId.value === event.id) startCreate()
    appStore.showSuccess('抽奖已删除')
    await loadLotteries()
  } catch (error: any) {
    appStore.showError(error?.message || '删除失败')
  }
}

onMounted(() => {
  startCreate()
  loadLotteries()
})
</script>
