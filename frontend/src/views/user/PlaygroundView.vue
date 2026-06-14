<template>
  <AppLayout>
    <div class="grid min-h-[calc(100vh-9rem)] gap-5 xl:grid-cols-[360px_minmax(0,1fr)]">
      <aside class="space-y-4">
        <section class="overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
          <div class="border-b border-gray-100 bg-gray-50/80 px-4 py-3 dark:border-dark-800 dark:bg-dark-950/40">
            <div class="flex items-center justify-between gap-3">
              <div class="flex min-w-0 items-center gap-3">
                <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary-50 text-primary-600 dark:bg-primary-500/10 dark:text-primary-300">
                  <Icon name="beaker" size="sm" />
                </div>
                <div class="min-w-0">
                  <h2 class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('playground.setup') }}</h2>
                  <p class="mt-0.5 text-xs leading-5 text-gray-500 dark:text-gray-400">{{ t('playground.setupHint') }}</p>
                </div>
              </div>
              <button
                type="button"
                class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-gray-200 bg-white text-gray-500 shadow-sm transition-colors hover:border-primary-200 hover:text-primary-600 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-400 dark:hover:border-primary-500/40 dark:hover:text-primary-300"
                :title="t('common.refresh')"
                @click="loadInitialData"
              >
                <Icon name="refresh" size="sm" :class="loadingKeys || loadingModels ? 'animate-spin' : ''" />
              </button>
            </div>
          </div>

          <div class="space-y-3 p-4">
            <label class="block rounded-md border border-gray-200 bg-gray-50/60 p-3 transition-colors focus-within:border-primary-400 focus-within:bg-white focus-within:ring-2 focus-within:ring-primary-500/10 dark:border-dark-700 dark:bg-dark-950/30 dark:focus-within:border-primary-500/60 dark:focus-within:bg-dark-900">
              <span class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon name="key" size="xs" />
                {{ t('playground.apiKey') }}
              </span>
              <div class="relative mt-2">
                <select
                  v-model="selectedKeyId"
                  class="h-10 w-full appearance-none rounded-md border border-gray-200 bg-white px-3 pr-9 text-sm font-medium text-gray-900 outline-none transition-colors focus:border-primary-400 dark:border-dark-700 dark:bg-dark-800 dark:text-white"
                >
                  <option :value="''">{{ t('playground.selectKey') }}</option>
                  <option
                    v-for="key in activeKeys"
                    :key="key.id"
                    :value="String(key.id)"
                  >
                    {{ key.name }} · {{ maskApiKey(key.key) }}
                  </option>
                </select>
                <Icon name="chevronDown" size="xs" class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400" />
              </div>
            </label>

            <label class="block rounded-md border border-gray-200 bg-gray-50/60 p-3 transition-colors focus-within:border-primary-400 focus-within:bg-white focus-within:ring-2 focus-within:ring-primary-500/10 dark:border-dark-700 dark:bg-dark-950/30 dark:focus-within:border-primary-500/60 dark:focus-within:bg-dark-900">
              <span class="flex items-center justify-between gap-2">
                <span class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                  <Icon name="brain" size="xs" />
                  {{ t('playground.model') }}
                </span>
                <span
                  v-if="loadingModels"
                  class="rounded-full bg-primary-50 px-2 py-0.5 text-[11px] font-medium text-primary-600 dark:bg-primary-500/10 dark:text-primary-300"
                >
                  Loading
                </span>
              </span>
              <div class="relative mt-2">
                <select
                  v-if="modelOptions.length > 0"
                  v-model="model"
                  class="h-10 w-full appearance-none rounded-md border border-gray-200 bg-white px-3 pr-9 font-mono text-sm text-gray-900 outline-none transition-colors focus:border-primary-400 dark:border-dark-700 dark:bg-dark-800 dark:text-white"
                >
                  <option v-for="name in modelOptions" :key="name" :value="name">{{ name }}</option>
                </select>
                <input
                  v-else
                  v-model="model"
                  class="h-10 w-full rounded-md border border-gray-200 bg-white px-3 font-mono text-sm text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-primary-400 dark:border-dark-700 dark:bg-dark-800 dark:text-white"
                  :placeholder="t('playground.modelPlaceholder')"
                />
                <Icon v-if="modelOptions.length > 0" name="chevronDown" size="xs" class="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-gray-400" />
              </div>
            </label>

            <label class="block rounded-md border border-gray-200 bg-gray-50/60 p-3 transition-colors focus-within:border-primary-400 focus-within:bg-white focus-within:ring-2 focus-within:ring-primary-500/10 dark:border-dark-700 dark:bg-dark-950/30 dark:focus-within:border-primary-500/60 dark:focus-within:bg-dark-900">
              <span class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                <Icon name="terminal" size="xs" />
                {{ t('playground.systemPrompt') }}
              </span>
              <textarea
                v-model="systemPrompt"
                rows="5"
                class="mt-2 min-h-[132px] w-full resize-none rounded-md border border-gray-200 bg-white px-3 py-2 text-sm leading-6 text-gray-900 outline-none transition-colors placeholder:text-gray-400 focus:border-primary-400 dark:border-dark-700 dark:bg-dark-800 dark:text-white"
                :placeholder="t('playground.systemPromptPlaceholder')"
              />
            </label>

            <div class="grid grid-cols-2 gap-3">
              <label class="block rounded-md border border-gray-200 bg-gray-50/60 p-3 transition-colors focus-within:border-primary-400 focus-within:bg-white focus-within:ring-2 focus-within:ring-primary-500/10 dark:border-dark-700 dark:bg-dark-950/30 dark:focus-within:border-primary-500/60 dark:focus-within:bg-dark-900">
                <span class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                  <Icon name="bolt" size="xs" />
                  {{ t('playground.temperature') }}
                </span>
                <input
                  v-model.number="temperature"
                  type="number"
                  min="0"
                  max="2"
                  step="0.1"
                  class="mt-2 h-10 w-full rounded-md border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-900 outline-none transition-colors focus:border-primary-400 dark:border-dark-700 dark:bg-dark-800 dark:text-white"
                />
              </label>
              <label class="block rounded-md border border-gray-200 bg-gray-50/60 p-3 transition-colors focus-within:border-primary-400 focus-within:bg-white focus-within:ring-2 focus-within:ring-primary-500/10 dark:border-dark-700 dark:bg-dark-950/30 dark:focus-within:border-primary-500/60 dark:focus-within:bg-dark-900">
                <span class="flex items-center gap-2 text-xs font-medium text-gray-500 dark:text-gray-400">
                  <Icon name="cpu" size="xs" />
                  {{ t('playground.maxTokens') }}
                </span>
                <input
                  v-model.number="maxTokens"
                  type="number"
                  min="1"
                  max="8192"
                  step="1"
                  class="mt-2 h-10 w-full rounded-md border border-gray-200 bg-white px-3 text-sm font-semibold text-gray-900 outline-none transition-colors focus:border-primary-400 dark:border-dark-700 dark:bg-dark-800 dark:text-white"
                />
              </label>
            </div>
          </div>
        </section>

      </aside>

      <main class="flex min-h-[620px] flex-col overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900">
        <header class="flex flex-wrap items-center justify-between gap-3 border-b border-gray-200 bg-white px-5 py-4 dark:border-dark-700 dark:bg-dark-900">
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-md bg-gray-900 text-white dark:bg-white dark:text-gray-950">
              <Icon name="chat" size="sm" />
            </div>
            <div>
              <h1 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('playground.title') }}</h1>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('playground.description') }}</p>
            </div>
          </div>
          <button type="button" class="inline-flex h-9 items-center gap-2 rounded-md border border-gray-200 bg-white px-3 text-sm font-medium text-gray-600 shadow-sm transition-colors hover:border-red-200 hover:text-red-600 disabled:cursor-not-allowed disabled:opacity-50 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-300 dark:hover:border-red-500/40 dark:hover:text-red-300" :disabled="sending || messages.length === 0" @click="clearChat">
            <Icon name="trash" size="xs" />
            {{ t('playground.clear') }}
          </button>
        </header>

        <div ref="messagesEl" class="flex-1 space-y-4 overflow-y-auto bg-gray-50 px-5 py-5 dark:bg-dark-950/50">
          <div v-if="messages.length === 0" class="flex h-full min-h-[360px] items-center justify-center text-center">
            <div>
              <div class="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-md border border-gray-200 bg-white text-gray-400 shadow-sm dark:border-dark-700 dark:bg-dark-900">
                <Icon name="chat" size="lg" />
              </div>
              <p class="text-base font-semibold text-gray-800 dark:text-gray-100">{{ t('playground.emptyTitle') }}</p>
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
                ? 'bg-gray-900 text-white dark:bg-primary-600'
                : 'border border-gray-200 bg-white text-gray-800 dark:border-dark-700 dark:bg-dark-800 dark:text-gray-100'"
            >
              <div class="mb-1 text-xs font-semibold opacity-75">
                {{ message.role === 'user' ? t('playground.you') : t('playground.assistant') }}
              </div>
              <div
                v-if="message.role === 'assistant'"
                class="playground-markdown leading-6"
                v-html="renderAssistantMessage(message.content)"
              />
              <div v-else class="whitespace-pre-wrap leading-6">{{ message.content }}</div>
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
import { authAPI, keysAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import type { ApiKey } from '@/types'
import { resolvePublicApiEndpoint } from '@/utils/apiEndpoint'
import { maskApiKey } from '@/utils/maskApiKey'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

marked.setOptions({
  breaks: true,
  gfm: true,
})

interface ChatMessage {
  id: number
  role: 'user' | 'assistant'
  content: string
}

const { t } = useI18n()

const keys = ref<ApiKey[]>([])
const selectedKeyId = ref('')
const model = ref('')
const gatewayModels = ref<string[]>([])
const systemPrompt = ref('你是 SuperAI Playground 的简洁中文测试助手。直接回答用户问题，不要介绍模型厂商、SDK 或内部实现。')
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

const activeKeys = computed(() => keys.value.filter((key) => key.status === 'active'))
const selectedKey = computed(() => activeKeys.value.find((key) => String(key.id) === selectedKeyId.value) || null)
const modelOptions = computed(() => gatewayModels.value)
const canSend = computed(() => Boolean(apiBaseUrl.value && selectedKey.value && model.value.trim() && draft.value.trim() && !sending.value))

function resolveDefaultEndpoint(configured: string): string {
  return resolvePublicApiEndpoint(configured)
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

async function loadModelsForSelectedKey() {
  gatewayModels.value = []
  if (!apiBaseUrl.value || !selectedKey.value) {
    model.value = ''
    return
  }

  loadingModels.value = true
  try {
    const response = await fetch(`${apiBaseUrl.value}/models`, {
      headers: {
        Authorization: `Bearer ${selectedKey.value.key}`,
      },
    })
    const payload = await response.json().catch(() => null)
    if (!response.ok) {
      const { code, message } = extractPlaygroundError(payload, response)
      throw new Error(normalizePlaygroundError(message, code))
    }

    const names = Array.isArray(payload?.data)
      ? payload.data
        .map((item: { id?: string }) => item?.id)
        .filter((name: unknown): name is string => typeof name === 'string' && isChatModelOption(name))
      : []
    gatewayModels.value = sortModelNames(Array.from(new Set(names)))
    applyDefaultModelForSelectedKey()
  } catch (error) {
    console.error('Failed to load playground models:', error)
    gatewayModels.value = []
    model.value = ''
    errorMessage.value = error instanceof Error ? error.message : t('playground.loadModelsFailed')
  } finally {
    loadingModels.value = false
  }
}

function sortModelNames(names: string[]): string[] {
  const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' })
  return names.sort((a, b) => collator.compare(b, a))
}

function isChatModelOption(name: string): boolean {
  const normalized = name.trim().toLowerCase()
  return Boolean(normalized) &&
    !normalized.includes('image') &&
    !normalized.includes('audio') &&
    !normalized.includes('realtime')
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
    await Promise.all([loadKeys(), loadPublicSettings()])
    await loadModelsForSelectedKey()
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

function extractPlaygroundError(payload: any, response: Response): { code: string; message: string } {
  const code = String(payload?.error?.code || payload?.code || '')
  const message = payload?.error?.message || payload?.message || `${response.status} ${response.statusText}`
  return { code, message }
}

function normalizePlaygroundError(message: string, code = ''): string {
  const normalizedCode = code.toUpperCase()
  const normalized = message.toLowerCase()
  if (normalizedCode === 'INSUFFICIENT_BALANCE' || normalized.includes('insufficient account balance')) {
    return t('playground.insufficientBalance')
  }
  if (normalized.includes('service temporarily unavailable') || normalized.includes('upstream service temporarily unavailable')) {
    return t('playground.serviceUnavailable')
  }
  return message || t('playground.sendFailed')
}

function renderAssistantMessage(content: string): string {
  if (!content) return ''
  const html = marked.parse(content) as string
  return DOMPurify.sanitize(html)
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
      const { code, message } = extractPlaygroundError(payload, response)
      throw new Error(normalizePlaygroundError(message, code))
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
    errorMessage.value = error instanceof Error ? error.message : t('playground.sendFailed')
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
  errorMessage.value = ''
  loadModelsForSelectedKey()
})
</script>

<style scoped>
.playground-markdown {
  overflow-wrap: anywhere;
}

.playground-markdown :deep(p) {
  margin: 0 0 0.75rem;
}

.playground-markdown :deep(p:last-child) {
  margin-bottom: 0;
}

.playground-markdown :deep(ul),
.playground-markdown :deep(ol) {
  margin: 0.5rem 0 0.75rem;
  padding-left: 1.25rem;
}

.playground-markdown :deep(ul) {
  list-style: disc;
}

.playground-markdown :deep(ol) {
  list-style: decimal;
}

.playground-markdown :deep(li) {
  margin: 0.25rem 0;
}

.playground-markdown :deep(strong) {
  font-weight: 700;
  color: rgb(17 24 39);
}

:global(.dark) .playground-markdown :deep(strong) {
  color: rgb(243 244 246);
}

.playground-markdown :deep(code) {
  border-radius: 0.25rem;
  background: rgb(243 244 246);
  padding: 0.1rem 0.3rem;
  font-size: 0.88em;
}

:global(.dark) .playground-markdown :deep(code) {
  background: rgb(31 41 55);
}
</style>
