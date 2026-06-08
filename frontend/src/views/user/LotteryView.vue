<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">抽奖</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">点击参与，到期开奖后自动发放余额或展示卡密。</p>
        </div>
        <button class="btn btn-secondary" :disabled="loading" @click="loadLotteries">
          <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          刷新
        </button>
      </div>

      <div v-if="loading" class="grid gap-4 md:grid-cols-2">
        <div v-for="i in 4" :key="i" class="card h-64 animate-pulse bg-gray-100 dark:bg-dark-800" />
      </div>

      <div v-else-if="lotteries.length === 0" class="card p-10 text-center">
        <Icon name="gift" size="xl" class="mx-auto text-gray-400" />
        <h2 class="mt-3 text-lg font-semibold text-gray-900 dark:text-white">暂无可参与抽奖</h2>
        <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">有新活动时会显示在这里。</p>
      </div>

      <div v-else class="grid gap-4 lg:grid-cols-2">
        <article
          v-for="item in lotteries"
          :key="item.id"
          class="card overflow-hidden border-gray-200 dark:border-dark-700"
        >
          <div class="border-b border-gray-100 bg-white p-5 dark:border-dark-700 dark:bg-dark-900">
            <div class="flex items-start justify-between gap-4">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <h2 class="truncate text-lg font-semibold text-gray-900 dark:text-white">{{ item.title }}</h2>
                  <span :class="statusClass(item.status)" class="badge">{{ statusLabel(item.status) }}</span>
                </div>
                <p v-if="item.description" class="mt-2 line-clamp-2 text-sm text-gray-600 dark:text-dark-300">
                  {{ item.description }}
                </p>
              </div>
              <div class="shrink-0 rounded-lg bg-amber-50 px-3 py-2 text-right dark:bg-amber-900/20">
                <div class="text-xs text-amber-700 dark:text-amber-300">参与人数</div>
                <div class="text-xl font-semibold text-amber-900 dark:text-amber-100">{{ item.entry_count }}</div>
              </div>
            </div>
          </div>

          <div class="space-y-4 p-5">
            <div class="grid gap-3 sm:grid-cols-2">
              <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
                <div class="text-xs text-gray-500 dark:text-dark-400">开奖时间</div>
                <div class="mt-1 text-sm font-medium text-gray-900 dark:text-white">{{ formatTs(item.draw_at) }}</div>
              </div>
              <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800">
                <div class="text-xs text-gray-500 dark:text-dark-400">我的状态</div>
                <div class="mt-1 text-sm font-medium text-gray-900 dark:text-white">
                  {{ item.my_winner ? '已中奖' : item.joined ? '已参与' : item.status === 'drawn' ? '未中奖' : '未参与' }}
                </div>
              </div>
            </div>

            <div>
              <div class="mb-2 text-sm font-medium text-gray-900 dark:text-white">奖品</div>
              <div class="grid gap-2">
                <div
                  v-for="prize in item.prizes"
                  :key="prize.id"
                  class="flex items-center justify-between rounded-lg border border-gray-100 px-3 py-2 dark:border-dark-700"
                >
                  <div class="min-w-0">
                    <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ prize.name }}</div>
                    <div class="text-xs text-gray-500 dark:text-dark-400">{{ prize.type === 'balance' ? `余额 ¥${formatMoney(prize.amount || 0)}` : '卡密' }}</div>
                  </div>
                  <span class="badge badge-gray">x{{ prize.quantity }}</span>
                </div>
              </div>
            </div>

            <div
              v-if="item.my_winner"
              class="rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-800/50 dark:bg-emerald-900/20"
            >
              <div class="text-sm font-semibold text-emerald-800 dark:text-emerald-200">
                恭喜中奖：{{ item.my_winner.prize_name }}
              </div>
              <p v-if="item.my_winner.prize_type === 'balance'" class="mt-1 text-sm text-emerald-700 dark:text-emerald-300">
                ¥{{ formatMoney(item.my_winner.amount || 0) }} 已自动加入你的余额。
              </p>
              <div v-else class="mt-3 flex gap-2">
                <input class="input flex-1" readonly :value="item.my_winner.card_content || ''" />
                <button class="btn btn-secondary" @click="copyCard(item.my_winner.card_content || '')">复制</button>
              </div>
            </div>

            <div v-if="item.winners?.length" class="rounded-xl border border-gray-100 p-4 dark:border-dark-700">
              <div class="mb-3 flex items-center justify-between gap-2">
                <div class="text-sm font-semibold text-gray-900 dark:text-white">中奖名单</div>
                <span class="badge badge-gray">{{ item.winners.length }} 人</span>
              </div>
              <div class="grid gap-2">
                <div
                  v-for="winner in item.winners"
                  :key="winner.id"
                  class="flex items-center justify-between gap-3 rounded-lg bg-gray-50 px-3 py-2 dark:bg-dark-800"
                >
                  <div class="min-w-0">
                    <div class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ winnerDisplayName(winner) }}</div>
                    <div class="truncate text-xs text-gray-500 dark:text-dark-400">{{ winner.prize_name }}</div>
                  </div>
                  <span class="badge" :class="winner.prize_type === 'balance' ? 'badge-success' : 'badge-warning'">
                    {{ winner.prize_type === 'balance' ? '余额' : '卡密' }}
                  </span>
                </div>
              </div>
            </div>

            <div class="flex items-center justify-between gap-3">
              <span class="text-xs text-gray-500 dark:text-dark-400">
                {{ item.status === 'active' ? countdownText(item.draw_at) : item.drawn_at ? `已于 ${formatTs(item.drawn_at)} 开奖` : '' }}
              </span>
              <button
                class="btn btn-primary"
                :disabled="joiningId === item.id || item.joined || item.status !== 'active' || Date.now() >= item.draw_at * 1000"
                @click="joinLottery(item)"
              >
                <Icon v-if="joiningId === item.id" name="refresh" size="sm" class="mr-1 animate-spin" />
                {{ item.joined ? '已参与' : item.status === 'drawn' ? '已开奖' : '立即参与' }}
              </button>
            </div>
          </div>
        </article>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { lotteryAPI } from '@/api'
import type { LotteryEvent, LotteryWinner } from '@/api/lottery'
import { formatDateTime } from '@/utils/format'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const lotteries = ref<LotteryEvent[]>([])
const loading = ref(false)
const joiningId = ref<number | null>(null)
let refreshTimer: number | undefined

const hasPendingDraw = computed(() =>
  lotteries.value.some((item) => item.status === 'active' && Date.now() >= item.draw_at * 1000)
)

function formatTs(value?: number): string {
  if (!value) return '-'
  return formatDateTime(new Date(value * 1000))
}

function formatMoney(value: number): string {
  return Number(value || 0).toFixed(2)
}

function winnerDisplayName(winner: LotteryWinner): string {
  const displayName = winner.user_display_name?.trim()
  if (displayName) return displayName
  const email = winner.user_email?.trim()
  if (email) return email.split('@')[0] || email
  return `用户${winner.user_id}`
}

function statusLabel(status: LotteryEvent['status']): string {
  if (status === 'active') return '进行中'
  if (status === 'drawn') return '已开奖'
  if (status === 'archived') return '已归档'
  return '草稿'
}

function statusClass(status: LotteryEvent['status']): string {
  if (status === 'active') return 'badge-success'
  if (status === 'drawn') return 'badge-warning'
  return 'badge-gray'
}

function countdownText(drawAt: number): string {
  const diff = drawAt * 1000 - Date.now()
  if (diff <= 0) return '等待系统开奖'
  const minutes = Math.ceil(diff / 60000)
  if (minutes < 60) return `${minutes} 分钟后开奖`
  const hours = Math.floor(minutes / 60)
  const rest = minutes % 60
  return `${hours} 小时 ${rest} 分钟后开奖`
}

async function loadLotteries(): Promise<void> {
  loading.value = true
  try {
    lotteries.value = await lotteryAPI.list()
  } catch (error: any) {
    appStore.showError(error?.message || '加载抽奖失败')
  } finally {
    loading.value = false
  }
}

async function joinLottery(item: LotteryEvent): Promise<void> {
  joiningId.value = item.id
  try {
    const updated = await lotteryAPI.join(item.id)
    const index = lotteries.value.findIndex((event) => event.id === item.id)
    if (index >= 0) lotteries.value[index] = updated
    appStore.showSuccess('参与成功')
  } catch (error: any) {
    appStore.showError(error?.message || '参与失败')
  } finally {
    joiningId.value = null
  }
}

async function copyCard(value: string): Promise<void> {
  if (!value) return
  await navigator.clipboard.writeText(value)
  appStore.showSuccess('卡密已复制')
}

onMounted(() => {
  loadLotteries()
  refreshTimer = window.setInterval(() => {
    if (hasPendingDraw.value) {
      loadLotteries()
    }
  }, 5000)
})

onUnmounted(() => {
  window.clearInterval(refreshTimer)
})
</script>
