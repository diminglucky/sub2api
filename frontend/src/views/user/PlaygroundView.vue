<template>
  <AppLayout>
    <div class="playground-shell">
      <div class="playground-mode-switch">
        <div class="mode-pills" role="tablist" :aria-label="t('playground.modeLabel')">
          <button
            type="button"
            class="mode-pill"
            :class="{ active: mode === 'chat' }"
            :aria-pressed="mode === 'chat'"
            @click="mode = 'chat'"
          >
            {{ t('playground.chatMode') }}
          </button>
          <button
            type="button"
            class="mode-pill"
            :class="{ active: mode === 'image' }"
            :aria-pressed="mode === 'image'"
            @click="mode = 'image'"
          >
            {{ t('playground.imageMode') }}
          </button>
        </div>

        <div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
          <Icon name="grid" size="sm" />
          <span>{{ mode === 'chat' ? t('playground.chatWorkbenchHint') : t('playground.imageWorkbenchHint') }}</span>
        </div>
      </div>

      <div v-if="mode === 'chat'" class="playground-grid">
        <aside class="playground-sidebar playground-sidebar-image card-glass">
          <div class="sidebar-section">
            <div class="section-title">
              <Icon name="key" size="sm" />
              <span>{{ t('playground.chatApiKey') }}</span>
            </div>
            <select v-model="selectedChatKeyId" class="input">
              <option :value="''">{{ t('playground.selectKey') }}</option>
              <option v-for="key in activeKeys" :key="key.id" :value="String(key.id)">
                {{ key.name }} · {{ maskApiKey(key.key) }}
              </option>
            </select>
          </div>

          <div class="sidebar-section">
            <div class="section-title">
              <Icon name="brain" size="sm" />
              <span>{{ t('playground.model') }}</span>
              <span v-if="loadingChatModels" class="section-badge">{{ t('common.loading') }}</span>
            </div>
            <select v-if="chatModelOptions.length > 0" v-model="model" class="input font-mono">
              <option v-for="name in chatModelOptions" :key="name" :value="name">{{ name }}</option>
            </select>
            <input
              v-else
              v-model="model"
              class="input font-mono"
              :placeholder="t('playground.modelPlaceholder')"
            />
            <p v-if="chatModelLoadError" class="model-load-error">{{ chatModelLoadError }}</p>
          </div>

          <div class="sidebar-grid">
            <label class="control-card">
              <span class="control-label">
                <Icon name="bolt" size="xs" />
                {{ t('playground.temperature') }}
              </span>
              <input v-model.number="temperature" type="number" min="0" max="2" step="0.1" class="input input-tight font-semibold" />
            </label>
            <label class="control-card">
              <span class="control-label">
                <Icon name="cpu" size="xs" />
                {{ t('playground.maxTokens') }}
              </span>
              <input v-model.number="maxTokens" type="number" min="1" max="8192" step="1" class="input input-tight font-semibold" />
            </label>
          </div>

          <button type="button" class="btn btn-primary btn-lg w-full" :disabled="!canSubmitChat" @click="sendMessage">
            <Icon name="sparkles" size="sm" />
            {{ sending ? t('playground.sending') : t('playground.send') }}
          </button>
        </aside>

        <main class="playground-panel card-glass">
          <div ref="messagesEl" class="panel-body">
            <div v-if="messages.length === 0" class="empty-stage">
              <div class="empty-stage__icon">
                <Icon name="sparkles" size="lg" />
              </div>
              <p class="empty-stage__title">{{ t('playground.chatEmptyTitle') }}</p>
              <p class="empty-stage__desc">{{ t('playground.chatEmptyDescription') }}</p>
            </div>

            <article
              v-for="message in messages"
              :key="message.id"
              class="message-row"
              :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
            >
              <div class="message-bubble" :class="message.role === 'user' ? 'message-bubble-user' : 'message-bubble-assistant'">
                <div class="message-meta">
                  {{ message.role === 'user' ? t('playground.you') : t('playground.assistant') }}
                </div>
                <div v-if="message.role === 'assistant'" class="playground-markdown" v-html="renderAssistantMessage(message.content)" />
                <div v-else class="whitespace-pre-wrap leading-6">{{ message.content }}</div>
              </div>
            </article>

            <div v-if="sending" class="message-row">
              <div class="thinking-chip">
                <Icon name="sparkles" size="sm" />
                {{ t('playground.thinking') }}
              </div>
            </div>
          </div>

          <form class="panel-composer" @submit.prevent="sendMessage">
            <div v-if="errorMessage" class="error-banner">
              {{ errorMessage }}
            </div>
            <div class="composer-shell">
              <textarea
                v-model="draft"
                rows="3"
                class="composer-input"
                :placeholder="t('playground.inputPlaceholder')"
                data-testid="playground-chat-input"
                @keydown.meta.enter.prevent="sendMessage"
                @keydown.ctrl.enter.prevent="sendMessage"
              />
              <button type="submit" class="send-button" :disabled="!canSubmitChat">
                <Icon name="arrowRight" size="sm" />
                {{ t('playground.send') }}
              </button>
            </div>
          </form>
        </main>
      </div>

      <div v-else class="playground-grid playground-grid-image">
        <aside class="playground-sidebar card-glass">
          <div class="sidebar-section">
            <div class="section-title">
              <Icon name="key" size="sm" />
              <span>{{ t('playground.imageApiKey') }}</span>
            </div>
            <select v-model="selectedImageKeyId" class="input">
              <option :value="''">{{ t('playground.selectKey') }}</option>
              <option v-for="key in activeKeys" :key="key.id" :value="String(key.id)">
                {{ key.name }} · {{ maskApiKey(key.key) }}
              </option>
            </select>
          </div>

          <div class="sidebar-section">
            <div class="section-title">
              <Icon name="brain" size="sm" />
              <span>{{ t('playground.model') }}</span>
              <span v-if="loadingImageModels" class="section-badge">{{ t('common.loading') }}</span>
            </div>
            <select v-if="imageModelOptions.length > 0" v-model="imageModel" class="input font-mono">
              <option v-for="name in imageModelOptions" :key="name" :value="name">{{ name }}</option>
            </select>
            <input
              v-else
              v-model="imageModel"
              class="input font-mono"
              :placeholder="t('playground.modelPlaceholder')"
            />
            <p v-if="imageModelLoadError" class="model-load-error">{{ imageModelLoadError }}</p>
          </div>

          <div class="sidebar-grid">
            <label class="control-card">
              <span class="control-label">
                <Icon name="grid" size="xs" />
                {{ t('playground.imageSize') }}
              </span>
              <select v-model="imageSize" class="input input-tight">
                <option v-for="option in imageSizeOptions" :key="option" :value="option">{{ option }}</option>
              </select>
            </label>
            <label class="control-card">
              <span class="control-label">
                <Icon name="badge" size="xs" />
                {{ t('playground.imageQuality') }}
              </span>
              <select v-model="imageQuality" class="input input-tight">
                <option v-for="option in imageQualityOptions" :key="option" :value="option">{{ option }}</option>
              </select>
            </label>
          </div>

          <div class="sidebar-grid">
            <label class="control-card">
              <span class="control-label">
                <Icon name="copy" size="xs" />
                {{ t('playground.imageCount') }}
              </span>
              <input v-model.number="imageCount" type="number" min="1" max="4" step="1" class="input input-tight font-semibold" />
            </label>
            <label class="control-card">
              <span class="control-label">
                <Icon name="swap" size="xs" />
                {{ t('playground.asyncImage') }}
              </span>
              <button type="button" class="toggle-like" :class="{ active: asyncImage }" @click="asyncImage = !asyncImage">
                <span>{{ asyncImage ? 'On' : 'Off' }}</span>
                <span class="toggle-like__dot" />
              </button>
            </label>
          </div>

          <details class="sidebar-section compact-extra" :open="Boolean(referenceImage)">
            <summary class="section-title compact-summary">
              <Icon name="paperclip" size="sm" />
              <span>{{ t('playground.referenceImage') }}</span>
            </summary>
            <ImageUpload
              v-model="referenceImage"
              :upload-label="t('playground.uploadImage')"
              :remove-label="t('playground.removeImage')"
              :hint="t('playground.referenceImageHint')"
            />
          </details>
        </aside>

        <main class="playground-panel card-glass">
          <div class="image-stage">
            <div v-if="images.length === 0" class="empty-stage empty-stage-image">
              <div class="empty-stage__icon">
                <Icon name="sparkles" size="lg" />
              </div>
              <p class="empty-stage__title">{{ t('playground.imageEmptyTitle') }}</p>
              <p class="empty-stage__desc">{{ t('playground.imageEmptyDescription') }}</p>
            </div>

            <div v-else class="image-results">
              <div class="image-viewer">
                <template v-if="selectedImage">
                  <div class="image-viewer__canvas">
                    <img :src="selectedImage.src" :alt="selectedImage.title" class="image-viewer__img" />
                    <button
                      type="button"
                      class="image-viewer__download"
                      :aria-label="t('playground.downloadImage')"
                      @click="downloadImage(selectedImage)"
                    >
                      <Icon name="download" size="sm" />
                    </button>
                  </div>
                  <div v-if="images.length > 1" class="image-thumbs" :aria-label="t('playground.result')">
                    <button
                      v-for="(image, index) in images"
                      :key="image.id"
                      type="button"
                      class="image-thumb"
                      :class="{ active: image.id === selectedImage.id }"
                      @click="selectImage(image.id)"
                    >
                      <img :src="image.src" :alt="`${t('playground.result')} ${index + 1}`" />
                    </button>
                  </div>
                </template>
                <div class="image-viewer__empty" v-else>
                  <Icon name="sparkles" size="lg" />
                  <p>{{ t('playground.imageViewerEmpty') }}</p>
                </div>
              </div>
            </div>
          </div>

          <form class="panel-composer panel-composer-image" @submit.prevent="generateImage">
            <div v-if="errorMessage" class="error-banner">
              {{ errorMessage }}
            </div>
            <div class="composer-shell">
              <textarea
                v-model="imagePrompt"
                rows="3"
                class="composer-input"
                :placeholder="t('playground.imagePromptPlaceholder')"
                data-testid="playground-image-prompt"
                @keydown.meta.enter.prevent="generateImage"
                @keydown.ctrl.enter.prevent="generateImage"
              />
              <button type="submit" class="send-button" :disabled="!canGenerateImage" data-testid="playground-image-generate">
                <Icon name="sparkles" size="sm" />
                {{ generatingImage ? t('playground.generating') : t('playground.generate') }}
              </button>
            </div>
          </form>
        </main>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import ImageUpload from '@/components/common/ImageUpload.vue'
import { authAPI, keysAPI } from '@/api'
import type { ApiKey } from '@/types'
import { resolvePlaygroundApiEndpoint } from '@/utils/apiEndpoint'
import { maskApiKey } from '@/utils/maskApiKey'

marked.setOptions({ breaks: true, gfm: true })

interface ChatMessage {
  id: number
  role: 'user' | 'assistant'
  content: string
}

interface GeneratedImage {
  id: number
  title: string
  src: string
  meta: string
}

const { t } = useI18n()

const mode = ref<'chat' | 'image'>('image')
const keys = ref<ApiKey[]>([])
const selectedChatKeyId = ref('')
const selectedImageKeyId = ref('')
const model = ref('')
const imageModel = ref('')
const chatModelOptions = ref<string[]>([])
const imageModelOptions = ref<string[]>([])
const chatModelLoadError = ref('')
const imageModelLoadError = ref('')
const temperature = ref(0.7)
const maxTokens = ref(4096)
const draft = ref('')
const imagePrompt = ref('一只趴在桌上的橙色小猫，柔和的窗边自然光，电影感，极简背景')
const messages = ref<ChatMessage[]>([])
const images = ref<GeneratedImage[]>([])
const selectedImageId = ref<number | null>(null)
const errorMessage = ref('')
const sending = ref(false)
const generatingImage = ref(false)
const loadingKeys = ref(false)
const loadingChatModels = ref(false)
const loadingImageModels = ref(false)
const apiBaseUrl = ref('')
const messagesEl = ref<HTMLElement | null>(null)
const referenceImage = ref('')
const imageSize = ref('1024x1024')
const imageQuality = ref('high')
const imageCount = ref(1)
const asyncImage = ref(true)
const imageSizeOptions = ['1024x1024', '1536x1024', '1024x1536', '1792x1024']
const imageQualityOptions = ['high', 'medium', 'low']
let nextMessageId = 1
let nextImageId = 1

const activeKeys = computed(() => keys.value.filter((key) => key.status === 'active'))
const selectedChatKey = computed(() => activeKeys.value.find((key) => String(key.id) === selectedChatKeyId.value) || null)
const selectedImageKey = computed(() => activeKeys.value.find((key) => String(key.id) === selectedImageKeyId.value) || null)
const modelOptions = computed(() => chatModelOptions.value)
const canSubmitChat = computed(() => Boolean(apiBaseUrl.value && selectedChatKey.value && model.value.trim() && draft.value.trim() && !sending.value))
const canGenerateImage = computed(() => Boolean(apiBaseUrl.value && selectedImageKey.value && imageModel.value.trim() && imagePrompt.value.trim() && !generatingImage.value))
const selectedImage = computed(() => images.value.find((image) => image.id === selectedImageId.value) || images.value[0] || null)
const hasReferenceImage = computed(() => Boolean(referenceImage.value.trim()))
let chatModelsRequestId = 0
let imageModelsRequestId = 0

function resolveDefaultEndpoint(configured: string): string {
  return resolvePlaygroundApiEndpoint(configured, typeof window === 'undefined' ? '' : window.location.hostname)
}

async function loadKeys() {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, 100, { status: 'active' })
    keys.value = response.items || []
    if (!selectedChatKeyId.value && activeKeys.value.length > 0) {
      selectedChatKeyId.value = String(activeKeys.value[0].id)
    }
    if (!selectedImageKeyId.value && activeKeys.value.length > 0) {
      selectedImageKeyId.value = String(activeKeys.value[0].id)
    }
  } finally {
    loadingKeys.value = false
  }
}

async function loadChatModelsForSelectedKey() {
  const requestId = ++chatModelsRequestId
  chatModelOptions.value = []
  chatModelLoadError.value = ''
  if (!apiBaseUrl.value || !selectedChatKey.value) {
    loadingChatModels.value = false
    model.value = ''
    return
  }

  const endpoint = apiBaseUrl.value
  const key = selectedChatKey.value
  loadingChatModels.value = true
  try {
    const names = await fetchModelNames(endpoint, key.key)
    if (requestId !== chatModelsRequestId) return
    const options = sortModelNames(names.filter(isChatModelOption))
    chatModelOptions.value = options
    if (options.length === 0) {
      model.value = ''
      chatModelLoadError.value = names.length > 0
        ? t('playground.noChatModelsAvailable')
        : t('playground.noModelsAvailable')
      return
    }
    applyDefaultModelForSelectedKey()
  } catch (error) {
    if (requestId !== chatModelsRequestId) return
    console.error('Failed to load playground models:', error)
    chatModelOptions.value = []
    model.value = ''
    const message = error instanceof Error ? error.message : t('playground.loadModelsFailed')
    chatModelLoadError.value = message
    if (mode.value === 'chat') errorMessage.value = message
  } finally {
    if (requestId === chatModelsRequestId) {
      loadingChatModels.value = false
    }
  }
}

async function loadImageModelsForSelectedKey() {
  const requestId = ++imageModelsRequestId
  imageModelOptions.value = []
  imageModelLoadError.value = ''
  if (!apiBaseUrl.value || !selectedImageKey.value) {
    loadingImageModels.value = false
    imageModel.value = ''
    return
  }

  const endpoint = apiBaseUrl.value
  const key = selectedImageKey.value
  loadingImageModels.value = true
  try {
    const names = await fetchModelNames(endpoint, key.key)
    if (requestId !== imageModelsRequestId) return
    imageModelOptions.value = sortImageModelNames(names)
    if (imageModelOptions.value.length === 0) {
      imageModel.value = ''
      imageModelLoadError.value = names.length > 0
        ? t('playground.noImageModelsAvailable')
        : t('playground.noModelsAvailable')
      return
    }
    applyDefaultImageModelForSelectedKey()
  } catch (error) {
    if (requestId !== imageModelsRequestId) return
    console.error('Failed to load playground models:', error)
    imageModelOptions.value = []
    imageModel.value = ''
    const message = error instanceof Error ? error.message : t('playground.loadModelsFailed')
    imageModelLoadError.value = message
    if (mode.value === 'image') errorMessage.value = message
  } finally {
    if (requestId === imageModelsRequestId) {
      loadingImageModels.value = false
    }
  }
}

async function fetchModelNames(endpoint: string, apiKey: string): Promise<string[]> {
  const response = await fetch(`${endpoint}/models`, {
    headers: {
      Authorization: `Bearer ${apiKey}`,
    },
  })
  const payload = await response.json().catch(() => null)
  if (!response.ok) {
    const { code, message } = extractPlaygroundError(payload, response)
    throw new Error(normalizePlaygroundError(message, code))
  }
  return extractModelNames(payload)
}

function extractModelNames(payload: any): string[] {
  if (!Array.isArray(payload?.data)) return []
  const names = payload.data
    .map((item: unknown) => {
      if (typeof item === 'string') return item
      if (item && typeof item === 'object') {
        const candidate = item as { id?: unknown; model?: unknown; name?: unknown }
        return candidate.id ?? candidate.model ?? candidate.name
      }
      return ''
    })
    .filter((name: unknown): name is string => typeof name === 'string')
    .map((name: string) => name.trim())
    .filter(Boolean)
  return Array.from(new Set(names))
}

function sortModelNames(names: string[]): string[] {
  const collator = new Intl.Collator(undefined, { numeric: true, sensitivity: 'base' })
  return names.sort((a, b) => collator.compare(b, a))
}

function sortImageModelNames(names: string[]): string[] {
  const unique = Array.from(new Set(names.map((name) => name.trim()).filter(isImageModelOption)))
  return sortModelNames(unique)
}

function isChatModelOption(name: string): boolean {
  const normalized = name.trim().toLowerCase()
  return Boolean(normalized) &&
    !isImageModelOption(normalized) &&
    !normalized.includes('audio') &&
    !normalized.includes('realtime')
}

function isImageModelOption(name: string): boolean {
  const normalized = name.trim().toLowerCase()
  return Boolean(normalized) && (
    normalized.startsWith('gpt-image-') ||
    normalized.includes('dall-e') ||
    normalized.includes('imagen') ||
    normalized.includes('flux') ||
    normalized.includes('midjourney') ||
    normalized.includes('stable-diffusion') ||
    normalized.includes('sdxl') ||
    normalized.includes('image')
  )
}

function applyDefaultModelForSelectedKey() {
  const options = modelOptions.value
  if (options.length === 0) return
  if (!options.includes(model.value)) {
    model.value = options[0]
  }
}

function applyDefaultImageModelForSelectedKey() {
  const options = imageModelOptions.value
  if (options.length === 0) return
  if (!options.includes(imageModel.value)) {
    imageModel.value = options[0]
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
    await Promise.all([loadChatModelsForSelectedKey(), loadImageModelsForSelectedKey()])
  } catch (error) {
    console.error('Failed to load playground data:', error)
    errorMessage.value = t('playground.loadFailed')
  }
}

function buildRequestMessages() {
  const requestMessages: Array<{ role: 'user' | 'assistant'; content: string }> = []
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
  if (normalized === 'failed to fetch' || normalized.includes('networkerror')) {
    return t('playground.networkFailed')
  }
  if (normalizedCode === 'INSUFFICIENT_BALANCE' || normalized.includes('insufficient account balance')) {
    return t('playground.insufficientBalance')
  }
  if (normalized.includes('service temporarily unavailable') || normalized.includes('upstream service temporarily unavailable')) {
    return t('playground.serviceUnavailable')
  }
  return message || t('playground.sendFailed')
}

function normalizePlaygroundException(error: unknown): string {
  if (error instanceof Error) {
    return normalizePlaygroundError(error.message)
  }
  return t('playground.sendFailed')
}

function renderAssistantMessage(content: string): string {
  if (!content) return ''
  return DOMPurify.sanitize(marked.parse(content) as string)
}

async function sendMessage() {
  const content = draft.value.trim()
  if (!content || sending.value) return
  if (!apiBaseUrl.value) {
    errorMessage.value = t('playground.apiEndpointMissing')
    return
  }
  if (!selectedChatKey.value) {
    errorMessage.value = t('playground.chatKeyRequired')
    return
  }
  if (!model.value.trim()) {
    errorMessage.value = t('playground.modelRequired')
    return
  }

  draft.value = ''
  errorMessage.value = ''
  messages.value.push({ id: nextMessageId++, role: 'user', content })
  await scrollToBottom()

  sending.value = true
  try {
    const response = await fetch(`${apiBaseUrl.value}/chat/completions`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${selectedChatKey.value.key}`,
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
    errorMessage.value = normalizePlaygroundException(error)
  } finally {
    sending.value = false
  }
}

async function generateImage() {
  if (!canGenerateImage.value || !selectedImageKey.value) return

  const prompt = imagePrompt.value.trim()
  errorMessage.value = ''
  generatingImage.value = true
  try {
    const headers: Record<string, string> = {
      Authorization: `Bearer ${selectedImageKey.value.key}`,
    }
    headers['Content-Type'] = 'application/json'
    const response = await fetch(`${apiBaseUrl.value}${hasReferenceImage.value ? '/images/edits' : '/images/generations'}`, {
      method: 'POST',
      headers,
      body: JSON.stringify(buildImagePayload(prompt)),
    })

    const payload = await readImageResponse(response)
    if (!response.ok) {
      const { code, message } = extractPlaygroundError(payload, response)
      throw new Error(normalizePlaygroundError(message, code))
    }

    const data = extractImageItems(payload)
    const nextImages = data
      .map((item: any, index: number) => {
        const b64 = String(item?.b64_json || item?.result || '')
        const src = b64
          ? `data:image/png;base64,${b64}`
          : String(item?.url || item?.download_url || '')
        if (!src) return null
        return {
          id: nextImageId++,
          title: item?.revised_prompt ? String(item.revised_prompt) : `${t('playground.result')} ${index + 1}`,
          src,
          meta: [item?.output_format, imageSize.value, imageQuality.value, item?.model ? String(item.model) : imageModel.value].filter(Boolean).join(' · '),
        }
      })
      .filter((item: GeneratedImage | null): item is GeneratedImage => Boolean(item))

    if (nextImages.length === 0) {
      throw new Error(t('playground.noImageResponse'))
    }

    images.value = nextImages
    selectedImageId.value = nextImages[0]?.id ?? null
  } catch (error) {
    console.error('Image generation request failed:', error)
    errorMessage.value = normalizePlaygroundException(error)
  } finally {
    generatingImage.value = false
  }
}

function buildImagePayload(prompt: string) {
  const payload: Record<string, unknown> = {
    model: imageModel.value.trim(),
    prompt,
    size: imageSize.value,
    quality: imageQuality.value,
    n: Math.min(Math.max(Number(imageCount.value) || 1, 1), 4),
    stream: asyncImage.value,
    response_format: 'b64_json',
  }
  if (hasReferenceImage.value) {
    payload.images = [{ image_url: referenceImage.value.trim() }]
  }
  return payload
}

async function readImageResponse(response: Response): Promise<any> {
  const contentType = response.headers.get('content-type') || ''
  if (contentType.includes('text/event-stream')) {
    return parseImageStream(await response.text())
  }
  return response.json().catch(() => null)
}

function parseImageStream(text: string): any {
  const blocks = text.split(/\n\s*\n/)
  const events: any[] = []
  for (const block of blocks) {
    const dataLines = block
      .split('\n')
      .filter((line) => line.startsWith('data:'))
      .map((line) => line.slice(5).trimStart())
    if (dataLines.length === 0) continue
    const payloadText = dataLines.join('\n').trim()
    if (!payloadText || payloadText === '[DONE]') continue
    try {
      events.push(JSON.parse(payloadText))
    } catch {
      continue
    }
  }
  return { data: events }
}

function extractImageItems(payload: any): any[] {
  if (!payload) return []
  if (Array.isArray(payload.data)) {
    return payload.data.flatMap((item: any) => extractImageItems(item))
  }
  if (Array.isArray(payload.response?.output)) return payload.response.output
  if (Array.isArray(payload.output)) return payload.output
  if (payload.b64_json || payload.result || payload.url) return [payload]
  return []
}

async function scrollToBottom() {
  await nextTick()
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  }
}

function selectImage(id: number) {
  selectedImageId.value = id
}

function downloadImage(image: GeneratedImage | null) {
  if (!image) return
  const link = document.createElement('a')
  link.href = image.src
  link.download = `${image.title || 'playground-image'}.png`
  link.click()
}

onMounted(() => {
  loadInitialData()
})

watch(selectedChatKeyId, () => {
  errorMessage.value = ''
  void loadChatModelsForSelectedKey()
})

watch(selectedImageKeyId, () => {
  errorMessage.value = ''
  void loadImageModelsForSelectedKey()
})

watch(mode, () => {
  errorMessage.value = ''
})
</script>

<style scoped>
.playground-shell {
  display: grid;
  grid-template-rows: auto minmax(0, 1fr);
  gap: 0.38rem;
  margin: -0.75rem -0.75rem 0;
  height: calc(100vh - 8rem);
  min-height: 34rem;
  overflow: hidden;
}

.playground-mode-switch {
  border-radius: 1rem;
  padding: 0 0.18rem 0.15rem;
}

.playground-mode-switch {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.45rem;
  background: transparent;
  border: 0;
}

.mode-pills {
  display: inline-flex;
  gap: 0.15rem;
  padding: 0.14rem;
  border-radius: 9999px;
  border: 1px solid rgb(203 213 225 / 0.9);
  background: rgb(255 255 255 / 0.86);
  box-shadow: 0 10px 26px rgb(15 23 42 / 0.08);
}

:global(.dark .playground-shell .mode-pills) {
  border-color: rgb(55 65 81 / 0.36);
  background: rgb(17 24 39 / 0.48);
  box-shadow: none;
}

.mode-pill {
  min-width: 6rem;
  border-radius: 9999px;
  padding: 0.42rem 0.8rem;
  color: rgb(71 85 105);
  transition: all 0.2s ease;
}

:global(.dark .playground-shell .mode-pill) {
  color: rgb(156 163 175);
}

.mode-pill.active {
  background: linear-gradient(135deg, rgb(15 23 42), rgb(30 41 59));
  color: white;
  box-shadow: 0 8px 24px rgb(0 0 0 / 0.2);
}

.playground-grid {
  display: grid;
  gap: 0.45rem;
  align-items: stretch;
  grid-template-columns: minmax(0, 330px) minmax(0, 1fr);
  min-height: 0;
}

.playground-grid-image {
  grid-template-columns: minmax(300px, 340px) minmax(0, 1fr);
}

.playground-sidebar,
.playground-panel {
  border-radius: 1.05rem;
  padding: 0.62rem;
  border: 1px solid rgb(203 213 225 / 0.8);
  min-height: 0;
}

:global(.dark .playground-shell .playground-sidebar),
:global(.dark .playground-shell .playground-panel) {
  border-color: rgb(55 65 81 / 0.38);
}

.playground-sidebar {
  display: grid;
  gap: 0.4rem;
  align-content: start;
  overflow: auto;
}

.playground-sidebar-image {
  align-content: start;
  gap: 0.36rem;
}

.playground-sidebar-image .sidebar-section {
  gap: 0.28rem;
}

.playground-sidebar-image .sidebar-grid {
  gap: 0.35rem;
}

.playground-sidebar-image .control-card {
  gap: 0.25rem;
  padding: 0.42rem;
  border-radius: 0.8rem;
}

.compact-extra {
  border-radius: 0.8rem;
  border: 1px solid rgb(203 213 225 / 0.82);
  background: rgb(255 255 255 / 0.72);
  padding: 0.42rem;
}

:global(.dark .playground-shell .compact-extra) {
  border-color: rgb(55 65 81 / 0.28);
  background: rgb(17 24 39 / 0.22);
}

.compact-summary {
  cursor: pointer;
  list-style: none;
}

.compact-summary::-webkit-details-marker {
  display: none;
}

.compact-extra > :not(summary) {
  margin-top: 0.45rem;
}

.sidebar-section {
  display: grid;
  gap: 0.34rem;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.86rem;
  font-weight: 600;
  color: rgb(17 24 39);
}

:global(.dark .playground-shell .section-title) {
  color: rgb(255 255 255);
}

.section-badge {
  border-radius: 9999px;
  border: 1px solid rgb(20 184 166 / 0.24);
  background: rgb(240 253 250 / 0.92);
  padding: 0.3rem 0.6rem;
  font-size: 0.75rem;
  color: rgb(15 118 110);
}

:global(.dark .playground-shell .section-badge) {
  border-color: rgb(55 65 81 / 0.6);
  background: rgb(17 24 39 / 0.75);
  color: rgb(209 213 219);
}

.sidebar-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.38rem;
}

.control-card {
  display: grid;
  gap: 0.26rem;
  border-radius: 0.78rem;
  border: 1px solid rgb(203 213 225 / 0.72);
  background: rgb(255 255 255 / 0.86);
  padding: 0.44rem;
  box-shadow: 0 10px 26px rgb(15 23 42 / 0.05);
}

:global(.dark .playground-shell .control-card) {
  border-color: rgb(55 65 81 / 0.24);
  background: rgb(15 23 42 / 0.34);
  box-shadow: none;
}

.control-label {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.8rem;
  color: rgb(71 85 105);
}

:global(.dark .playground-shell .control-label) {
  color: rgb(156 163 175);
}

.input-tight {
  min-height: 2rem;
}

.toggle-like {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-radius: 9999px;
  border: 1px solid rgb(203 213 225 / 0.9);
  background: rgb(255 255 255 / 0.9);
  padding: 0.42rem 0.7rem;
  color: rgb(51 65 85);
}

:global(.dark .playground-shell .toggle-like) {
  border-color: rgb(55 65 81 / 0.55);
  background: rgb(17 24 39 / 0.55);
  color: rgb(209 213 219);
}

.toggle-like.active {
  border-color: rgb(249 115 22 / 0.8);
  background: rgb(255 237 213 / 0.9);
  color: rgb(154 52 18);
}

:global(.dark .playground-shell .toggle-like.active) {
  background: rgb(124 45 18 / 0.55);
  color: rgb(209 213 219);
}

.toggle-like__dot {
  width: 1.9rem;
  height: 1rem;
  border-radius: 9999px;
  background: rgb(148 163 184);
  position: relative;
}

:global(.dark .playground-shell .toggle-like__dot) {
  background: rgb(75 85 99);
}

.toggle-like.active .toggle-like__dot {
  background: rgb(249 115 22);
}

.toggle-like__dot::after {
  content: '';
  position: absolute;
  top: 0.08rem;
  left: 0.08rem;
  width: 0.85rem;
  height: 0.85rem;
  border-radius: 9999px;
  background: white;
  transition: transform 0.2s ease;
}

.toggle-like.active .toggle-like__dot::after {
  transform: translateX(0.9rem);
}

.playground-panel {
  display: grid;
  grid-template-rows: minmax(0, 1fr) auto;
  gap: 0.45rem;
  overflow: hidden;
}

.playground-grid-image .playground-panel {
  grid-template-rows: minmax(0, 1fr) auto;
}

.panel-body {
  min-height: 0;
  overflow-y: auto;
  border-radius: 0.95rem;
  border: 1px solid rgb(226 232 240 / 0.88);
  background: linear-gradient(180deg, rgb(255 255 255 / 0.94), rgb(248 250 252 / 0.9));
  padding: 0.65rem;
  display: grid;
  gap: 0.55rem;
  align-content: start;
  color: rgb(15 23 42);
}

:global(.dark .playground-shell .panel-body) {
  border-color: transparent;
  background: rgb(9 9 11 / 0.52);
  color: rgb(243 244 246);
}

.empty-stage {
  min-height: 100%;
  display: grid;
  place-items: center;
  text-align: center;
  padding: 0.8rem 0.6rem;
  color: rgb(100 116 139);
}

:global(.dark .playground-shell .empty-stage) {
  color: rgb(156 163 175);
}

.empty-stage__icon {
  margin-bottom: 0.45rem;
  display: grid;
  place-items: center;
  height: 2.9rem;
  width: 2.9rem;
  border-radius: 0.9rem;
  border: 1px solid rgb(203 213 225 / 0.86);
  background: rgb(255 255 255 / 0.94);
  color: rgb(15 118 110);
  box-shadow: 0 12px 30px rgb(15 23 42 / 0.08);
}

:global(.dark .playground-shell .empty-stage__icon) {
  border-color: transparent;
  background: rgb(31 41 55 / 0.9);
  color: white;
  box-shadow: none;
}

.empty-stage__title {
  font-size: 1.05rem;
  font-weight: 700;
  color: rgb(15 23 42);
}

:global(.dark .playground-shell .empty-stage__title) {
  color: rgb(255 255 255);
}

.empty-stage__desc {
  margin-top: 0.32rem;
  max-width: 26rem;
  line-height: 1.55;
}

.message-row {
  display: flex;
}

.message-bubble {
  max-width: min(48rem, 88%);
  border-radius: 1.25rem;
  padding: 0.78rem 0.9rem;
  font-size: 0.95rem;
  box-shadow: 0 8px 24px rgb(0 0 0 / 0.15);
}

.message-bubble-user {
  margin-left: auto;
  background: linear-gradient(135deg, rgb(249 115 22), rgb(234 88 12));
  color: white;
}

.message-bubble-assistant {
  border: 1px solid rgb(226 232 240);
  background: rgb(255 255 255 / 0.96);
  color: rgb(30 41 59);
}

:global(.dark .playground-shell .message-bubble-assistant) {
  border-color: rgb(55 65 81 / 0.55);
  background: rgb(17 24 39 / 0.85);
  color: rgb(243 244 246);
}

.message-meta {
  margin-bottom: 0.4rem;
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  opacity: 0.7;
}

.thinking-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 9999px;
  border: 1px solid rgb(226 232 240);
  background: rgb(255 255 255 / 0.9);
  padding: 0.75rem 1rem;
  color: rgb(51 65 85);
}

:global(.dark .playground-shell .thinking-chip) {
  border-color: rgb(55 65 81 / 0.55);
  background: rgb(17 24 39 / 0.75);
  color: rgb(229 231 235);
}

.panel-composer {
  display: grid;
  gap: 0.35rem;
}

.composer-shell {
  display: flex;
  align-items: stretch;
  gap: 0.45rem;
  border-radius: 1rem;
  border: 1px solid rgb(203 213 225 / 0.82);
  background: rgb(255 255 255 / 0.92);
  padding: 0.46rem;
  box-shadow: 0 10px 28px rgb(15 23 42 / 0.06);
}

:global(.dark .playground-shell .composer-shell) {
  border-color: rgb(55 65 81 / 0.55);
  background: rgb(9 9 11 / 0.74);
  box-shadow: none;
}

.composer-input {
  flex: 1;
  min-height: 3.4rem;
  max-height: 7rem;
  resize: none;
  border: 0;
  background: transparent;
  color: rgb(15 23 42);
  outline: none;
  padding: 0.38rem 0.42rem;
}

.composer-input::placeholder {
  color: rgb(100 116 139);
}

:global(.dark .playground-shell .composer-input) {
  color: white;
}

:global(.dark .playground-shell .composer-input::placeholder) {
  color: rgb(156 163 175);
}

.panel-composer-image .composer-input {
  min-height: 3rem;
}

.panel-composer-image .composer-shell {
  border-radius: 1.15rem;
  padding: 0.48rem;
}

.send-button {
  align-self: end;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  border-radius: 0.82rem;
  background: linear-gradient(135deg, rgb(249 115 22), rgb(234 88 12));
  padding: 0.58rem 0.84rem;
  font-weight: 700;
  color: white;
  transition: transform 0.2s ease;
}

.send-button:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.error-banner {
  border-radius: 0.75rem;
  border: 1px solid rgb(254 202 202);
  background: rgb(254 242 242 / 0.9);
  padding: 0.5rem 0.7rem;
  color: rgb(185 28 28);
}

:global(.dark .playground-shell .error-banner) {
  border-color: rgb(185 28 28 / 0.45);
  background: rgb(127 29 29 / 0.22);
  color: rgb(252 165 165);
}

.model-load-error {
  font-size: 0.78rem;
  line-height: 1.45;
  color: rgb(248 113 113);
}

.image-stage {
  display: grid;
  gap: 0.5rem;
  min-height: 0;
  overflow: hidden;
}

.empty-stage-image {
  min-height: min(42vh, 390px);
  grid-column: 1 / -1;
}

.image-results {
  display: grid;
  place-items: center;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}

.image-viewer {
  width: min(100%, 860px);
  height: 100%;
  max-height: 100%;
  overflow: hidden;
  border-radius: 1rem;
  border: 1px solid rgb(226 232 240 / 0.92);
  background: rgb(255 255 255 / 0.9);
  display: grid;
  grid-template-rows: minmax(0, 1fr) auto;
  min-height: 0;
}

:global(.dark .playground-shell .image-viewer) {
  border-color: rgb(55 65 81 / 0.28);
  background: rgb(9 9 11 / 0.42);
}

.image-viewer__empty {
  min-height: min(38vh, 340px);
  display: grid;
  place-items: center;
  gap: 0.5rem;
  text-align: center;
  color: rgb(100 116 139);
}

:global(.dark .playground-shell .image-viewer__empty) {
  color: rgb(156 163 175);
}

.image-viewer__canvas {
  position: relative;
  display: grid;
  place-items: center;
  min-height: 0;
  overflow: hidden;
  padding: 0.65rem;
  background: linear-gradient(180deg, rgb(248 250 252 / 0.95), rgb(241 245 249 / 0.9));
}

:global(.dark .playground-shell .image-viewer__canvas) {
  background: rgb(3 7 18 / 0.42);
}

.image-viewer__img {
  display: block;
  width: 100%;
  height: 100%;
  min-height: 0;
  object-fit: contain;
  border-radius: 0.85rem;
}

.image-viewer__download {
  position: absolute;
  top: 0.9rem;
  right: 0.9rem;
  display: inline-grid;
  place-items: center;
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  border: 1px solid rgb(203 213 225 / 0.88);
  background: rgb(255 255 255 / 0.82);
  color: rgb(30 41 59);
  box-shadow: 0 10px 24px rgb(15 23 42 / 0.14);
  backdrop-filter: blur(10px);
  transition: background 0.18s ease, transform 0.18s ease;
}

.image-viewer__download:hover {
  background: rgb(255 255 255 / 0.96);
  transform: translateY(-1px);
}

:global(.dark .playground-shell .image-viewer__download) {
  border-color: rgb(148 163 184 / 0.34);
  background: rgb(2 6 23 / 0.58);
  color: rgb(226 232 240);
  box-shadow: 0 10px 24px rgb(0 0 0 / 0.22);
}

:global(.dark .playground-shell .image-viewer__download:hover) {
  background: rgb(15 23 42 / 0.82);
}

.image-thumbs {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.35rem;
  border-top: 1px solid rgb(226 232 240 / 0.88);
  background: rgb(255 255 255 / 0.72);
  padding: 0.55rem;
}

:global(.dark .playground-shell .image-thumbs) {
  border-top-color: rgb(55 65 81 / 0.28);
  background: transparent;
}

.image-thumb {
  height: 3.5rem;
  width: 3.5rem;
  overflow: hidden;
  border-radius: 0.75rem;
  border: 2px solid rgb(226 232 240);
  background: rgb(255 255 255 / 0.92);
}

:global(.dark .playground-shell .image-thumb) {
  border-color: transparent;
  background: rgb(15 23 42 / 0.62);
}

.image-thumb.active {
  border-color: rgb(249 115 22);
}

.image-thumb img {
  display: block;
  height: 100%;
  width: 100%;
  object-fit: cover;
}

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

:global(.dark .playground-shell .playground-markdown strong) {
  color: rgb(243 244 246);
}

.playground-markdown :deep(code) {
  border-radius: 0.25rem;
  background: rgb(243 244 246);
  padding: 0.1rem 0.3rem;
  font-size: 0.88em;
}

:global(.dark .playground-shell .playground-markdown code) {
  background: rgb(31 41 55);
}

@media (max-width: 1280px) {
  .playground-grid,
  .playground-grid-image {
    grid-template-columns: minmax(0, 1fr);
  }

  .image-stage {
    grid-template-columns: minmax(0, 1fr);
  }
}

@media (max-width: 768px) {
  .playground-shell {
    margin: 0;
  }

  .sidebar-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .composer-shell,
  .playground-mode-switch {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
