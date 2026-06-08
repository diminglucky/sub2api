<template>
  <AppLayout>
    <div class="grid min-h-[calc(100vh-9rem)] gap-5 xl:grid-cols-[320px_minmax(0,1fr)]">
      <aside class="space-y-4">
        <section class="rounded-lg border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
          <div class="mb-4 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('playground.setup') }}</h2>
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('playground.setupHint') }}</p>
            </div>
            <button
              type="button"
              class="rounded-lg p-2 text-gray-500 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-gray-400 dark:hover:bg-dark-800 dark:hover:text-white"
              :title="t('common.refresh')"
              @click="loadInitialData"
            >
              <Icon name="refresh" size="sm" :class="loadingKeys || loadingModels ? 'animate-spin' : ''" />
            </button>
          </div>

          <div class="space-y-4">
            <label class="block">
              <span class="input-label">{{ t('playground.apiKey') }}</span>
              <select v-model="selectedKeyId" class="input mt-1">
                <option :value="''">{{ t('playground.selectKey') }}</option>
                <option
                  v-for="key in activeKeys"
                  :key="key.id"
                  :value="String(key.id)"
                >
                  {{ key.name }} · {{ maskApiKey(key.key) }}
                </option>
              </select>
            </label>

            <label class="block">
              <span class="input-label">{{ t('playground.model') }}</span>
              <select
                v-if="modelOptions.length > 0"
                v-model="model"
                class="input mt-1 font-mono text-sm"
              >
                <option v-for="name in modelOptions" :key="name" :value="name">{{ name }}</option>
              </select>
              <input
                v-else
                v-model="model"
                class="input mt-1 font-mono text-sm"
                :placeholder="t('playground.modelPlaceholder')"
              />
            </label>

            <label class="block">
              <span class="input-label">{{ t('playground.systemPrompt') }}</span>
              <textarea
                v-model="systemPrompt"
                rows="4"
                class="input mt-1 resize-none text-sm"
                :placeholder="t('playground.systemPromptPlaceholder')"
              />
            </label>

            <div class="grid grid-cols-2 gap-3">
              <label class="block">
                <span class="input-label">{{ t('playground.temperature') }}</span>
                <input v-model.number="temperature" type="number" min="0" max="2" step="0.1" class="input mt-1" />
              </label>
              <label class="block">
                <span class="input-label">{{ t('playground.maxTokens') }}</span>
                <input v-model.number="maxTokens" type="number" min="1" max="8192" step="1" class="input mt-1" />
              </label>
            </div>
          </div>
        </section>

      </aside>

      <main class="flex min-h-[620px] flex-col overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900">
        <header class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 px-5 py-4 dark:border-dark-700">
          <div>
            <h1 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('playground.title') }}</h1>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('playground.description') }}</p>
          </div>
          <button type="button" class="btn btn-secondary" :disabled="sending || messages.length === 0" @click="clearChat">
            <Icon name="trash" size="sm" class="mr-2" />
            {{ t('playground.clear') }}
          </button>
        </header>

        <div ref="messagesEl" class="flex-1 space-y-4 overflow-y-auto bg-gray-50/70 px-5 py-5 dark:bg-dark-950/40">
          <div v-if="messages.length === 0" class="flex h-full min-h-[360px] items-center justify-center text-center">
            <div>
              <Icon name="chat" size="xl" class="mx-auto mb-3 text-gray-400" />
              <p class="text-base font-semibold text-gray-700 dark:text-gray-200">{{ t('playground.emptyTitle') }}</p>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('playground.emptyDescription') }}</p>
            </div>
          </div>

          <article
            v-for="message in messages"
            :key="message.id"
            class="flex"
            :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
          >
            <div
              class="max-w-[min(48rem,88%)] rounded-lg px-4 py-3 text-sm shadow-sm"
              :class="message.role === 'user'
                ? 'bg-primary-600 text-white'
                : 'border border-gray-200 bg-white text-gray-800 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-100'"
            >
              <div class="mb-1 text-xs font-semibold opacity-75">
                {{ message.role === 'user' ? t('playground.you') : t('playground.assistant') }}
              </div>
              <div class="whitespace-pre-wrap leading-6">{{ message.content }}</div>
            </div>
          </article>

          <article v-if="sending" class="flex justify-start">
            <div class="rounded-lg border border-gray-200 bg-white px-4 py-3 text-sm text-gray-500 shadow-sm dark:border-dark-700 dark:bg-dark-800 dark:text-gray-300">
              {{ t('playground.thinking') }}
            </div>
          </article>
        </div>

        <form class="border-t border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900" @submit.prevent="sendMessage">
          <div v-if="errorMessage" class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300">
            {{ errorMessage }}
          </div>
          <div class="flex gap-3">
            <textarea
              v-model="draft"
              rows="3"
              class="input min-h-[84px] flex-1 resize-none"
              :placeholder="t('playground.inputPlaceholder')"
              @keydown.meta.enter.prevent="sendMessage"
              @keydown.ctrl.enter.prevent="sendMessage"
            />
            <button type="submit" class="btn btn-primary self-end" :disabled="!canSend">
              <Icon name="play" size="sm" class="mr-2" />
              {{ sending ? t('playground.sending') : t('playground.send') }}
            </button>
          </div>
        </form>
      </main>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { authAPI, keysAPI, userChannelsAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import type { UserAvailableChannel } from '@/api/channels'
import type { ApiKey } from '@/types'
import { maskApiKey } from '@/utils/maskApiKey'

interface ChatMessage {
  id: number
  role: 'user' | 'assistant'
  content: string
}

const { t } = useI18n()

const keys = ref<ApiKey[]>([])
const selectedKeyId = ref('')
const model = ref('')
const availableChannels = ref<UserAvailableChannel[]>([])
const systemPrompt = ref('你是一个友好、准确、简洁的中文助手。请优先使用中文回答用户的问题。')
const temperature = ref(0.7)
const maxTokens = ref(1024)
const draft = ref('')
const messages = ref<ChatMessage[]>([])
const errorMessage = ref('')
const sending = ref(false)
const loadingKeys = ref(false)
const loadingModels = ref(false)
const apiBaseUrl = ref('')
const messagesEl = ref<HTMLElement | null>(null)
let nextMessageId = 1

const fallbackModelsByPlatform: Record<string, string[]> = {
  openai: [
    'gpt-5.5',
    'gpt-5.4',
    'gpt-5.4-mini',
    'gpt-5.3-codex',
    'gpt-5.3-codex-spark',
    'gpt-5.2',
  ],
  anthropic: [
    'claude-opus-4-8',
    'claude-opus-4-7',
    'claude-opus-4-6',
    'claude-sonnet-4-6',
    'claude-opus-4-5-20251101',
    'claude-sonnet-4-5-20250929',
    'claude-haiku-4-5-20251001',
  ],
  gemini: [
    'gemini-3.1-pro-preview',
    'gemini-3.1-flash-image',
    'gemini-3-pro-preview',
    'gemini-3-flash-preview',
    'gemini-2.5-pro',
    'gemini-2.5-flash',
    'gemini-2.0-flash',
  ],
  antigravity: [
    'claude-opus-4-8',
    'claude-opus-4-7',
    'claude-opus-4-6',
    'claude-opus-4-6-thinking',
    'claude-sonnet-4-6',
    'claude-sonnet-4-5',
    'claude-sonnet-4-5-thinking',
    'gemini-3.1-pro-high',
    'gemini-3.1-pro-low',
    'gemini-3.1-flash-image',
    'gemini-3-pro-preview',
    'gemini-2.5-flash',
  ],
}

const activeKeys = computed(() => keys.value.filter((key) => key.status === 'active'))
const selectedKey = computed(() => activeKeys.value.find((key) => String(key.id) === selectedKeyId.value) || null)
const selectedGroupId = computed(() => selectedKey.value?.group_id ?? selectedKey.value?.group?.id ?? null)
const selectedPlatform = computed(() => selectedKey.value?.group?.platform || '')
const modelOptions = computed(() => collectModelOptions(selectedPlatform.value, selectedGroupId.value))
const canSend = computed(() => Boolean(apiBaseUrl.value && selectedKey.value && model.value.trim() && draft.value.trim() && !sending.value))

function resolveDefaultEndpoint(configured: string): string {
  const trimmed = configured.trim().replace(/\/+$/, '')
  if (trimmed) return trimmed
  return `${window.location.origin.replace(/\/+$/, '')}/v1`
}

async function loadKeys() {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, 100, { status: 'active' })
    keys.value = response.items || []
    if (!selectedKeyId.value && activeKeys.value.length > 0) {
      selectedKeyId.value = String(activeKeys.value[0].id)
    }
  } finally {
    loadingKeys.value = false
  }
}

async function loadModels() {
  loadingModels.value = true
  try {
    availableChannels.value = await userChannelsAPI.getAvailableModels()
    applyDefaultModelForSelectedKey()
  } finally {
    loadingModels.value = false
  }
}

function collectModelOptions(platform: string, groupId: number | null): string[] {
  const names = new Set<string>()
  for (const name of fallbackModelsByPlatform[platform] || []) {
    names.add(name)
  }
  for (const channel of availableChannels.value) {
    for (const section of channel.platforms || []) {
      if (platform) {
        if (section.platform !== platform) continue
      } else if (groupId && !section.groups.some((group) => group.id === groupId)) {
        continue
      }

      for (const supported of section.supported_models || []) {
        if (supported.name) {
          names.add(supported.name)
        }
      }
    }
  }
  return sortModelNames(Array.from(names))
}

function sortModelNames(names: string[]): string[] {
  const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' })
  return names.sort((a, b) => collator.compare(b, a))
}

function applyDefaultModelForSelectedKey() {
  const options = modelOptions.value
  if (options.length === 0) return
  if (!options.includes(model.value)) {
    model.value = options[0]
  }
}

async function loadPublicSettings() {
  const settings = await authAPI.getPublicSettings()
  apiBaseUrl.value = resolveDefaultEndpoint(settings.api_base_url || '')
}

async function loadInitialData() {
  errorMessage.value = ''
  try {
    await Promise.all([loadKeys(), loadModels(), loadPublicSettings()])
  } catch (error) {
    console.error('Failed to load playground data:', error)
    errorMessage.value = t('playground.loadFailed')
  }
}

function buildRequestMessages() {
  const requestMessages: Array<{ role: 'system' | 'user' | 'assistant'; content: string }> = []
  const system = systemPrompt.value.trim()
  if (system) requestMessages.push({ role: 'system', content: system })
  for (const message of messages.value) {
    requestMessages.push({ role: message.role, content: message.content })
  }
  return requestMessages
}

function normalizePlaygroundError(message: string): string {
  const normalized = message.toLowerCase()
  if (normalized.includes('insufficient account balance')) {
    return t('playground.insufficientBalance')
  }
  if (normalized.includes('service temporarily unavailable') || normalized.includes('upstream service temporarily unavailable')) {
    return t('playground.serviceUnavailable')
  }
  return message || t('playground.sendFailed')
}

async function sendMessage() {
  if (!canSend.value || !selectedKey.value) return

  const content = draft.value.trim()
  draft.value = ''
  errorMessage.value = ''
  messages.value.push({ id: nextMessageId++, role: 'user', content })
  await scrollToBottom()

  sending.value = true
  try {
    const response = await fetch(`${apiBaseUrl.value}/chat/completions`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${selectedKey.value.key}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        model: model.value.trim(),
        messages: buildRequestMessages(),
        temperature: Number.isFinite(temperature.value) ? temperature.value : 0.7,
        max_tokens: Number.isFinite(maxTokens.value) ? maxTokens.value : 1024,
        stream: false,
      }),
    })

    const payload = await response.json().catch(() => null)
    if (!response.ok) {
      const message = payload?.error?.message || payload?.message || `${response.status} ${response.statusText}`
      throw new Error(message)
    }

    const answer = payload?.choices?.[0]?.message?.content || ''
    messages.value.push({
      id: nextMessageId++,
      role: 'assistant',
      content: answer || t('playground.emptyResponse'),
    })
    await scrollToBottom()
  } catch (error) {
    console.error('Playground request failed:', error)
    errorMessage.value = error instanceof Error ? normalizePlaygroundError(error.message) : t('playground.sendFailed')
  } finally {
    sending.value = false
  }
}

async function scrollToBottom() {
  await nextTick()
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  }
}

function clearChat() {
  messages.value = []
  errorMessage.value = ''
}

onMounted(() => {
  loadInitialData()
})

watch(selectedKeyId, () => {
  applyDefaultModelForSelectedKey()
})
</script>
