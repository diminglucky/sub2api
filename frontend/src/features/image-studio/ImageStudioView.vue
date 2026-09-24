<template>
  <AppLayout>
    <div class="mx-auto flex min-h-[calc(100vh-8rem)] max-w-6xl flex-col gap-4">
      <header class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 dark:text-white">AI 图片生成</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
            选择你自己的 API Key，直接生成或编辑图片。
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <label class="toolbar-field">
            <span>API Key</span>
            <select v-model="selectedKeyId">
              <option value="">选择密钥</option>
              <option v-for="key in activeKeys" :key="key.id" :value="String(key.id)">
                {{ key.name }} · {{ maskApiKey(key.key) }}
              </option>
            </select>
          </label>
          <label class="toolbar-field">
            <span>模型</span>
            <select v-if="imageModels.length" v-model="imageModel" class="font-mono">
              <option v-for="model in imageModels" :key="model" :value="model">{{ model }}</option>
            </select>
            <input v-else v-model="imageModel" class="font-mono" placeholder="输入图片模型" />
          </label>
        </div>
      </header>

      <section class="min-h-[24rem] flex-1 overflow-auto rounded-xl border border-gray-200 bg-white p-4 dark:border-dark-700 dark:bg-dark-900">
        <div v-if="!images.length" class="flex min-h-[22rem] flex-col items-center justify-center text-center text-gray-500 dark:text-gray-400">
          <div class="flex h-16 w-16 items-center justify-center rounded-2xl border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-800">
            <Icon name="sparkles" size="xl" />
          </div>
          <p class="mt-4 font-medium text-gray-700 dark:text-gray-200">输入提示词开始生成图片</p>
          <p class="mt-1 text-sm">支持文生图，也可以上传参考图做图生图。</p>
        </div>

        <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <article
            v-for="image in images"
            :key="image.id"
            class="group overflow-hidden rounded-xl border border-gray-200 bg-gray-50 dark:border-dark-700 dark:bg-dark-800"
          >
            <button type="button" class="block aspect-square w-full overflow-hidden" @click="previewImage = image.src">
              <img :src="image.src" :alt="image.title" class="h-full w-full object-cover transition-transform duration-200 group-hover:scale-[1.02]" />
            </button>
            <div class="flex items-center justify-between gap-2 px-3 py-2">
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-gray-800 dark:text-gray-100">{{ image.title }}</p>
                <p class="truncate text-xs text-gray-400">{{ image.meta }}</p>
              </div>
              <div class="flex shrink-0 gap-1">
                <button type="button" class="mini-action" title="设为参考图" @click="useImageAsReference(image)">
                  <Icon name="paperclip" size="sm" />
                </button>
                <button type="button" class="mini-action" title="下载图片" @click="downloadImage(image)">
                  <Icon name="download" size="sm" />
                </button>
              </div>
            </div>
          </article>
        </div>
      </section>

      <form class="sticky bottom-3 rounded-2xl border border-gray-200 bg-white p-4 shadow-xl dark:border-dark-700 dark:bg-dark-900" @submit.prevent="generate">
        <div v-if="errorMessage" class="mb-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-600 dark:border-red-900/50 dark:bg-red-950/30 dark:text-red-300">
          {{ errorMessage }}
        </div>

        <div v-if="references.length" class="mb-3 flex flex-wrap gap-2">
          <div v-for="item in references" :key="item.id" class="flex max-w-48 items-center gap-2 rounded-lg border border-gray-200 bg-gray-50 px-2 py-1 dark:border-dark-700 dark:bg-dark-800">
            <img :src="item.preview" :alt="item.name" class="h-8 w-8 rounded object-cover" />
            <span class="truncate text-xs text-gray-600 dark:text-gray-300">{{ item.name }}</span>
            <button type="button" class="text-gray-400 hover:text-red-500" @click="removeReference(item.id)">
              <Icon name="x" size="xs" />
            </button>
          </div>
        </div>

        <textarea
          v-model="prompt"
          rows="2"
          class="w-full resize-none border-0 bg-transparent text-sm text-gray-900 outline-none placeholder:text-gray-400 dark:text-white"
          placeholder="描述你想生成的图片，支持上传参考图..."
          @paste="onPaste"
        />

        <div class="mt-3 flex flex-wrap items-end gap-3 border-t border-gray-100 pt-3 dark:border-dark-800">
          <label class="composer-field">
            <span>尺寸</span>
            <select v-model="size">
              <option v-for="item in sizeOptions" :key="item" :value="item">{{ item }}</option>
            </select>
          </label>
          <label class="composer-field">
            <span>质量</span>
            <select v-model="quality">
              <option value="high">high</option>
              <option value="medium">medium</option>
              <option value="low">low</option>
            </select>
          </label>
          <label class="composer-field">
            <span>格式</span>
            <select v-model="format">
              <option value="png">PNG</option>
              <option value="jpeg">JPEG</option>
              <option value="webp">WEBP</option>
            </select>
          </label>
          <label class="composer-field">
            <span>背景</span>
            <select v-model="background">
              <option value="auto">自动</option>
              <option value="transparent">透明</option>
              <option value="opaque">不透明</option>
            </select>
          </label>
          <label class="composer-field composer-field--count">
            <span>数量</span>
            <input v-model.number="count" type="number" min="1" max="4" />
          </label>

          <div class="flex-1" />
          <input ref="fileInput" type="file" accept="image/png,image/jpeg,image/webp" multiple class="hidden" @change="onFilesSelected" />
          <button type="button" class="secondary-button" title="上传参考图" @click="fileInput?.click()">
            <Icon name="paperclip" size="sm" />
          </button>
          <button type="submit" class="primary-button" :disabled="!canGenerate">
            <Icon name="sparkles" size="sm" />
            {{ generating ? '生成中' : '生成' }}
          </button>
        </div>
      </form>
    </div>

    <Teleport to="body">
      <div v-if="previewImage" class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 p-4" @click="previewImage = ''">
        <img :src="previewImage" alt="图片预览" class="max-h-[88vh] max-w-[92vw] rounded-xl object-contain shadow-2xl" />
      </div>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { authAPI, keysAPI } from '@/api'
import type { ApiKey } from '@/types'
import { resolvePlaygroundApiEndpoint } from '@/utils/apiEndpoint'
import { maskApiKey } from '@/utils/maskApiKey'

interface GeneratedImage {
  id: number
  title: string
  src: string
  meta: string
}

interface ReferenceItem {
  id: number
  name: string
  file: File
  preview: string
}

const keys = ref<ApiKey[]>([])
const selectedKeyId = ref('')
const apiBaseUrl = ref('')
const imageModel = ref('')
const imageModels = ref<string[]>([])
const prompt = ref('一只趴在桌上的橙色小猫，柔和窗边自然光，电影感，极简背景')
const size = ref('1024x1024')
const quality = ref('high')
const format = ref('png')
const background = ref('auto')
const count = ref(1)
const generating = ref(false)
const errorMessage = ref('')
const images = ref<GeneratedImage[]>([])
const references = ref<ReferenceItem[]>([])
const previewImage = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

const sizeOptions = ['1024x1024', '1536x1024', '1024x1536', '1792x1024']
const activeKeys = computed(() => keys.value.filter((key) => key.status === 'active'))
const selectedKey = computed(() => activeKeys.value.find((key) => String(key.id) === selectedKeyId.value) || null)
const canGenerate = computed(() => Boolean(apiBaseUrl.value && selectedKey.value?.key && imageModel.value.trim() && prompt.value.trim() && !generating.value))
const hasReferences = computed(() => references.value.length > 0)

let nextImageId = 1
let nextReferenceId = 1
let modelRequestId = 0

function resolveEndpoint(configured: string) {
  return resolvePlaygroundApiEndpoint(configured, typeof window === 'undefined' ? '' : window.location.hostname)
}

async function loadData() {
  errorMessage.value = ''
  try {
    const [keyResponse, settings] = await Promise.all([
      keysAPI.list(1, 100, { status: 'active' }),
      authAPI.getPublicSettings(),
    ])
    keys.value = keyResponse.items || []
    if (!selectedKeyId.value && activeKeys.value.length) {
      selectedKeyId.value = String(activeKeys.value[0].id)
    }
    apiBaseUrl.value = resolveEndpoint(settings.api_base_url || '')
    await loadModels()
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '加载数据失败'
  }
}

async function loadModels() {
  const requestId = ++modelRequestId
  imageModels.value = []
  imageModel.value = ''
  if (!selectedKey.value || !apiBaseUrl.value) return
  const response = await fetch(`${apiBaseUrl.value}/models`, {
    headers: { Authorization: `Bearer ${selectedKey.value.key}` },
  })
  const payload = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(payload?.error?.message || payload?.message || `${response.status} ${response.statusText}`)
  }
  const names: string[] = Array.isArray(payload?.data)
    ? payload.data.map((item: any) => String(item?.id || item?.model || item || '').trim()).filter(Boolean)
    : []
  if (requestId !== modelRequestId) return
  imageModels.value = Array.from(new Set(names.filter(isImageModel))).sort()
  if (!imageModel.value && imageModels.value.length) {
    imageModel.value = imageModels.value[0]
  }
}

function isImageModel(name: string) {
  const value = name.trim().toLowerCase()
  return value.startsWith('gpt-image-') ||
    value.includes('dall-e') ||
    value.includes('image') ||
    value.includes('imagen') ||
    value.includes('flux') ||
    value.includes('sdxl')
}

function isGeminiModel(name: string) {
  return name.toLowerCase().includes('gemini') && isImageModel(name)
}

async function generate() {
  if (!canGenerate.value || !selectedKey.value) return
  generating.value = true
  errorMessage.value = ''
  const currentPrompt = prompt.value.trim()
  try {
    const headers: Record<string, string> = { Authorization: `Bearer ${selectedKey.value.key}` }
    let url = `${apiBaseUrl.value}${hasReferences.value ? '/images/edits' : '/images/generations'}`
    let body: string | FormData
    if (isGeminiModel(imageModel.value)) {
      headers['Content-Type'] = 'application/json'
      body = JSON.stringify(buildGeminiPayload(currentPrompt))
      url = buildGeminiURL(imageModel.value)
    } else if (hasReferences.value) {
      body = buildEditForm(currentPrompt)
    } else {
      headers['Content-Type'] = 'application/json'
      body = JSON.stringify({
        model: imageModel.value.trim(),
        prompt: currentPrompt,
        size: size.value,
        quality: quality.value,
        output_format: format.value,
        background: background.value,
        n: Math.min(Math.max(Number(count.value) || 1, 1), 4),
        response_format: 'b64_json',
      })
    }
    const response = await fetch(url, { method: 'POST', headers, body })
    const payload = await response.json().catch(() => null)
    if (!response.ok) {
      throw new Error(payload?.error?.message || payload?.message || `${response.status} ${response.statusText}`)
    }
    const items = extractImages(payload)
    if (!items.length) throw new Error('上游没有返回图片')
    images.value = items.map((item: any, index: number) => {
      const mime = String(item.mimeType || item.mime_type || 'image/png')
      const b64 = String(item.b64_json || item.data || '')
      return {
        id: nextImageId++,
        title: `生成结果 ${index + 1}`,
        src: b64 ? `data:${mime};base64,${b64}` : String(item.url || ''),
        meta: `${format.value.toUpperCase()} · ${size.value} · ${imageModel.value}`,
      }
    }).filter((item: GeneratedImage) => item.src)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : '生成失败'
  } finally {
    generating.value = false
  }
}

function buildGeminiURL(model: string) {
  const root = apiBaseUrl.value.replace(/\/+$/, '').replace(/\/v1$/, '')
  const prefix = selectedKey.value?.group?.platform === 'antigravity' ? '/antigravity/v1beta' : '/v1beta'
  return `${root}${prefix}/models/${encodeURIComponent(model)}:generateContent`
}

function buildGeminiPayload(text: string) {
  const parts = references.value.map((item) => {
    const match = /^data:([^;,]+);base64,(.+)$/i.exec(item.preview)
    return match ? { inlineData: { mimeType: match[1], data: match[2] } } : { text: item.name }
  })
  parts.push({ text })
  return {
    contents: [{ role: 'user', parts }],
    generationConfig: {
      responseModalities: ['TEXT', 'IMAGE'],
      imageConfig: { aspectRatio: geminiAspectRatio(size.value) },
    },
  }
}

function geminiAspectRatio(value: string) {
  return ({ '1024x1024': '1:1', '1536x1024': '3:2', '1024x1536': '2:3', '1792x1024': '16:9' } as Record<string, string>)[value] || '1:1'
}

function buildEditForm(text: string) {
  const form = new FormData()
  form.append('model', imageModel.value.trim())
  form.append('prompt', text)
  form.append('size', size.value)
  form.append('quality', quality.value)
  form.append('output_format', format.value)
  form.append('background', background.value)
  form.append('n', String(Math.min(Math.max(Number(count.value) || 1, 1), 4)))
  form.append('response_format', 'b64_json')
  references.value.forEach((item) => form.append('image[]', item.file, item.name))
  return form
}

function extractImages(payload: any): any[] {
  if (!payload) return []
  if (Array.isArray(payload.data)) return payload.data.flatMap(extractImages)
  if (Array.isArray(payload.candidates)) {
    return payload.candidates.flatMap((candidate: any) =>
      (candidate?.content?.parts || [])
        .filter((part: any) => part?.inlineData?.data)
        .map((part: any) => ({
          b64_json: part.inlineData.data,
          mimeType: part.inlineData.mimeType || 'image/png',
        })),
    )
  }
  if (Array.isArray(payload.response?.output)) return payload.response.output
  if (Array.isArray(payload.output)) return payload.output
  if (payload.b64_json || payload.url || payload.result) return [payload]
  return []
}

function onFilesSelected(event: Event) {
  const input = event.target as HTMLInputElement
  addFiles(Array.from(input.files || []))
  input.value = ''
}

function onPaste(event: ClipboardEvent) {
  const files = Array.from(event.clipboardData?.files || []).filter((file) => file.type.startsWith('image/'))
  if (files.length) addFiles(files)
}

function addFiles(files: File[]) {
  const available = Math.max(0, 4 - references.value.length)
  files
    .filter((file) => ['image/png', 'image/jpeg', 'image/webp'].includes(file.type))
    .slice(0, available)
    .forEach((file) => {
      const reader = new FileReader()
      reader.onload = () => {
        references.value.push({
          id: nextReferenceId++,
          name: file.name || 'reference.png',
          file,
          preview: String(reader.result || ''),
        })
      }
      reader.readAsDataURL(file)
    })
}

function removeReference(id: number) {
  references.value = references.value.filter((item) => item.id !== id)
}

async function useImageAsReference(image: GeneratedImage) {
  const response = await fetch(image.src)
  const blob = await response.blob()
  addFiles([new File([blob], `reference-${image.id}.png`, { type: blob.type || 'image/png' })])
}

function downloadImage(image: GeneratedImage) {
  const link = document.createElement('a')
  link.href = image.src
  link.download = `${image.title}.png`
  link.click()
}

watch(selectedKeyId, () => {
  loadModels().catch((error) => {
    errorMessage.value = error instanceof Error ? error.message : '加载模型失败'
  })
})

onMounted(loadData)
</script>

<style scoped>
.toolbar-field {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-height: 2.4rem;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.5rem;
  background: rgb(255 255 255 / 0.85);
  padding: 0 0.6rem;
}

.toolbar-field span,
.composer-field span {
  white-space: nowrap;
  font-size: 0.75rem;
  color: rgb(100 116 139);
}

.toolbar-field select,
.toolbar-field input {
  min-width: 8rem;
  border: 0;
  background: transparent;
  color: rgb(15 23 42);
  font-size: 0.82rem;
  outline: none;
}

.composer-field {
  display: grid;
  gap: 0.2rem;
  min-width: 5.5rem;
}

.composer-field select,
.composer-field input {
  height: 2rem;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.4rem;
  background: rgb(248 250 252);
  padding: 0 0.45rem;
  color: rgb(30 41 59);
  font-size: 0.78rem;
}

.composer-field--count {
  min-width: 4.5rem;
}

.mini-action,
.secondary-button,
.primary-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  border-radius: 0.5rem;
  border: 1px solid rgb(226 232 240);
  transition: all 0.18s ease;
}

.mini-action {
  width: 2rem;
  height: 2rem;
  background: rgb(255 255 255 / 0.9);
  color: rgb(71 85 105);
}

.secondary-button {
  width: 2.25rem;
  height: 2.25rem;
  background: rgb(255 255 255 / 0.9);
  color: rgb(71 85 105);
}

.primary-button {
  min-height: 2.25rem;
  border-color: transparent;
  background: rgb(15 23 42);
  padding: 0 1rem;
  color: white;
  font-weight: 600;
}

.primary-button:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

:global(.dark .toolbar-field),
:global(.dark .mini-action),
:global(.dark .secondary-button) {
  border-color: rgb(63 63 70);
  background: rgb(24 24 27 / 0.9);
}

:global(.dark .toolbar-field select),
:global(.dark .toolbar-field input),
:global(.dark .composer-field select),
:global(.dark .composer-field input) {
  color: rgb(244 244 245);
}

:global(.dark .composer-field select),
:global(.dark .composer-field input) {
  border-color: rgb(63 63 70);
  background: rgb(9 9 11);
}
</style>
