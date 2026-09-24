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
          <label class="toolbar-field toolbar-field--select">
            <span>API Key</span>
            <ImageStudioSelect
              v-model="selectedKeyId"
              class="toolbar-image-select"
              data-testid="api-key-select"
              :options="apiKeyOptions"
              placeholder="选择密钥"
              direction="down"
            />
          </label>
          <label class="toolbar-field toolbar-field--select">
            <span>模型</span>
            <ImageStudioSelect
              v-model="imageModel"
              class="toolbar-image-select toolbar-image-select--model"
              data-testid="image-model-select"
              :options="imageModelOptions"
              :disabled="loadingModels || imageModels.length === 0"
              :placeholder="modelSelectPlaceholder"
              direction="down"
            />
          </label>
          <span v-if="modelLoadError" class="toolbar-error">{{ modelLoadError }}</span>
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
          <button
            type="button"
            class="composer-size-trigger"
            data-testid="image-size-trigger"
            @click="openSizeDialog"
          >
            <span>尺寸</span>
            <strong>{{ size }}</strong>
            <Icon name="chevronDown" size="xs" />
          </button>
          <label class="composer-field">
            <span>质量</span>
            <ImageStudioSelect
              v-model="quality"
              data-testid="quality-select"
              :options="qualityOptions"
            />
          </label>
          <label class="composer-field">
            <span>格式</span>
            <ImageStudioSelect
              v-model="format"
              data-testid="format-select"
              :options="formatOptions"
            />
          </label>
          <label class="composer-field">
            <span>背景</span>
            <ImageStudioSelect
              v-model="background"
              data-testid="background-select"
              :options="backgroundOptions"
            />
          </label>
          <div class="composer-field composer-field--count">
            <span>数量</span>
            <div class="count-stepper" data-testid="image-count-control">
              <button
                type="button"
                class="count-stepper__button"
                data-testid="count-decrement"
                aria-label="减少数量"
                :disabled="count <= 1"
                @click="adjustCount(-1)"
              >
                <Icon name="minus" size="xs" />
              </button>
              <input
                v-model.number="count"
                type="number"
                min="1"
                max="4"
                aria-label="生成数量"
                @blur="normalizeCount"
              />
              <button
                type="button"
                class="count-stepper__button"
                data-testid="count-increment"
                aria-label="增加数量"
                :disabled="count >= 4"
                @click="adjustCount(1)"
              >
                <Icon name="plus" size="xs" />
              </button>
            </div>
          </div>

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

    <BaseDialog
      :show="sizeDialogOpen"
      title="设置图像尺寸"
      width="normal"
      @close="closeSizeDialog"
    >
      <div class="size-dialog">
        <p class="size-dialog__current">当前：<strong>{{ size }}</strong></p>

        <div class="size-tabs" role="tablist" aria-label="尺寸设置方式">
          <button
            v-for="mode in sizeModeOptions"
            :key="mode.value"
            type="button"
            role="tab"
            class="size-tab"
            :class="{ active: draftSize.mode === mode.value }"
            :aria-selected="draftSize.mode === mode.value"
            :data-testid="`size-mode-${mode.value}`"
            @click="draftSize.mode = mode.value"
          >
            {{ mode.label }}
          </button>
        </div>

        <div class="size-dialog__content">
          <div v-if="draftSize.mode === 'auto'" class="size-auto-panel">
            <div class="size-auto-icon">
              <Icon name="bolt" size="lg" />
            </div>
            <p class="size-auto-title">自动尺寸</p>
            <p class="size-auto-description">
              不向模型传递具体的分辨率参数<br />
              由模型自己决定生成尺寸
            </p>
          </div>

          <template v-else-if="draftSize.mode === 'ratio'">
            <section class="size-section">
              <p class="size-section__label">基准分辨率</p>
              <div class="size-base-options">
                <button
                  v-for="base in aspectBaseOptions"
                  :key="base.value"
                  type="button"
                  class="size-base-option"
                  :class="{ active: draftSize.aspectBase === base.value }"
                  @click="draftSize.aspectBase = base.value"
                >
                  {{ base.label }}
                </button>
              </div>
            </section>

            <section class="size-section">
              <p class="size-section__label">图像比例</p>
              <div class="size-ratio-grid">
                <button
                  v-for="ratio in ratioOptions"
                  :key="ratio.value"
                  type="button"
                  class="size-ratio-option"
                  :class="{ active: !draftSize.useCustomRatio && draftSize.aspectRatio === ratio.value }"
                  @click="selectAspectRatio(ratio.value)"
                >
                  <span
                    class="size-ratio-shape"
                    :style="{ aspectRatio: ratio.css }"
                  />
                  <span>{{ ratio.value }}</span>
                </button>
              </div>
              <button
                type="button"
                class="size-custom-ratio-toggle"
                :class="{ active: draftSize.useCustomRatio }"
                @click="draftSize.useCustomRatio = !draftSize.useCustomRatio"
              >
                自定义比例
              </button>
              <div v-if="draftSize.useCustomRatio" class="size-custom-ratio">
                <input v-model.number="draftSize.customRatioWidth" type="number" min="1" max="64" />
                <span>:</span>
                <input v-model.number="draftSize.customRatioHeight" type="number" min="1" max="64" />
              </div>
            </section>
          </template>

          <template v-else>
            <section class="size-section">
              <p class="size-section__label">输入具体像素值</p>
              <div class="size-dimension-row">
                <label>
                  <span>宽度（Width）</span>
                  <input
                    v-model.number="draftSize.width"
                    data-testid="custom-width"
                    type="number"
                    min="256"
                    max="3840"
                    step="16"
                  />
                </label>
                <span class="size-dimension-separator">×</span>
                <label>
                  <span>高度（Height）</span>
                  <input
                    v-model.number="draftSize.height"
                    data-testid="custom-height"
                    type="number"
                    min="256"
                    max="3840"
                    step="16"
                  />
                </label>
              </div>
              <p class="size-dimension-hint">
                宽高会按 16 的倍数自动规整，最长边不超过 3840px。
              </p>
            </section>
          </template>
        </div>

        <div class="size-preview">
          <span>将使用</span>
          <strong>{{ draftSizeValue }}</strong>
        </div>
      </div>

      <template #footer>
        <div class="size-dialog__footer">
          <button type="button" class="size-dialog__cancel" @click="closeSizeDialog">取消</button>
          <button
            type="button"
            class="size-dialog__confirm"
            data-testid="confirm-image-size"
            :disabled="!canConfirmSize"
            @click="confirmSize"
          >
            确定
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ImageStudioSelect from './ImageStudioSelect.vue'
import Icon from '@/components/icons/Icon.vue'
import { authAPI, imageStudioGalleryAPI, keysAPI, type ImageStudioGalleryEntry } from '@/api'
import type { ApiKey } from '@/types'
import { resolvePlaygroundApiEndpoint } from '@/utils/apiEndpoint'

interface GeneratedImage {
  id: number | string
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

type SizeMode = 'auto' | 'ratio' | 'custom'

interface SizeSettings {
  mode: SizeMode
  aspectBase: number
  aspectRatio: string
  useCustomRatio: boolean
  customRatioWidth: number
  customRatioHeight: number
  width: number
  height: number
}

const keys = ref<ApiKey[]>([])
const selectedKeyId = ref('')
const apiBaseUrl = ref('')
const imageModel = ref('')
const imageModels = ref<string[]>([])
const loadingModels = ref(false)
const modelLoadError = ref('')
const prompt = ref('一只趴在桌上的橙色小猫，柔和窗边自然光，电影感，极简背景')
const appliedSize = reactive<SizeSettings>({
  mode: 'ratio',
  aspectBase: 1024,
  aspectRatio: '1:1',
  useCustomRatio: false,
  customRatioWidth: 16,
  customRatioHeight: 9,
  width: 1024,
  height: 1024,
})
const draftSize = reactive<SizeSettings>({ ...appliedSize })
const sizeDialogOpen = ref(false)
const quality = ref('auto')
const format = ref('png')
const background = ref('auto')
const count = ref(1)
const generating = ref(false)
const errorMessage = ref('')
const images = ref<GeneratedImage[]>([])
const references = ref<ReferenceItem[]>([])
const previewImage = ref('')
const fileInput = ref<HTMLInputElement | null>(null)

const size = computed(() => formatSize(appliedSize))
const draftSizeValue = computed(() => formatSize(draftSize))
const canConfirmSize = computed(() => {
  if (draftSize.mode === 'ratio' && draftSize.useCustomRatio) {
    return draftSize.customRatioWidth > 0 && draftSize.customRatioHeight > 0
  }
  if (draftSize.mode === 'custom') {
    return Number.isFinite(Number(draftSize.width)) &&
      Number.isFinite(Number(draftSize.height)) &&
      Number(draftSize.width) >= 256 &&
      Number(draftSize.height) >= 256
  }
  return true
})
const sizeModeOptions: Array<{ value: SizeMode; label: string }> = [
  { value: 'auto', label: '自动' },
  { value: 'ratio', label: '按比例' },
  { value: 'custom', label: '自定义宽高' },
]
const aspectBaseOptions = [
  { value: 1024, label: '1K' },
  { value: 2048, label: '2K' },
  { value: 3840, label: '4K' },
]
const ratioOptions = [
  { value: '1:1', css: '1 / 1' },
  { value: '3:2', css: '3 / 2' },
  { value: '2:3', css: '2 / 3' },
  { value: '16:9', css: '16 / 9' },
  { value: '9:16', css: '9 / 16' },
  { value: '4:3', css: '4 / 3' },
  { value: '3:4', css: '3 / 4' },
  { value: '21:9', css: '21 / 9' },
]
const qualityOptions = [
  { value: 'auto', label: 'auto' },
  { value: 'low', label: 'low' },
  { value: 'medium', label: 'medium' },
  { value: 'high', label: 'high' },
]
const formatOptions = [
  { value: 'png', label: 'PNG' },
  { value: 'jpeg', label: 'JPEG' },
  { value: 'webp', label: 'WEBP' },
]
const backgroundOptions = [
  { value: 'auto', label: '自动' },
  { value: 'transparent', label: '透明' },
  { value: 'opaque', label: '不透明' },
]
const imageModelOptions = computed(() => imageModels.value.map((model) => ({ value: model, label: model })))
const modelSelectPlaceholder = computed(() => {
  if (loadingModels.value) return '正在加载...'
  if (!imageModels.value.length) return '未找到图片模型'
  return '选择模型'
})
const activeKeys = computed(() => keys.value.filter((key) => key.status === 'active'))
const apiKeyOptions = computed(() => activeKeys.value.map((key) => ({ value: String(key.id), label: key.name })))
const selectedKey = computed(() => activeKeys.value.find((key) => String(key.id) === selectedKeyId.value) || null)
const canGenerate = computed(() => Boolean(apiBaseUrl.value && selectedKey.value?.key && imageModel.value.trim() && prompt.value.trim() && !generating.value))
const hasReferences = computed(() => references.value.length > 0)

let nextImageId = 1
let nextReferenceId = 1
let modelRequestId = 0

function resolveEndpoint(configured: string) {
  return resolvePlaygroundApiEndpoint(configured, typeof window === 'undefined' ? '' : window.location.hostname)
}

function openSizeDialog() {
  Object.assign(draftSize, appliedSize)
  const match = findSizeMatch(size.value)
  if (match) Object.assign(draftSize, match)
  sizeDialogOpen.value = true
}

function closeSizeDialog() {
  sizeDialogOpen.value = false
}

function confirmSize() {
  if (!canConfirmSize.value) return
  Object.assign(appliedSize, draftSize)
  sizeDialogOpen.value = false
}

function selectAspectRatio(value: string) {
  draftSize.aspectRatio = value
  draftSize.useCustomRatio = false
}

function adjustCount(delta: number) {
  count.value = Math.min(4, Math.max(1, Number(count.value || 1) + delta))
}

function normalizeCount() {
  count.value = Math.min(4, Math.max(1, Math.round(Number(count.value) || 1)))
}

function findSizeMatch(value: string): Partial<SizeSettings> | null {
  if (value === 'auto') return { mode: 'auto' }
  const dimensions = parseSize(value)
  if (!dimensions) return { mode: 'custom' }

  for (const base of aspectBaseOptions) {
    for (const ratio of ratioOptions) {
      if (ratioSize(ratio.value, base.value) === value) {
        return {
          mode: 'ratio',
          aspectBase: base.value,
          aspectRatio: ratio.value,
          useCustomRatio: false,
        }
      }
    }
  }

  return {
    mode: 'custom',
    width: dimensions.width,
    height: dimensions.height,
  }
}

function formatSize(settings: SizeSettings) {
  if (settings.mode === 'auto') return 'auto'
  if (settings.mode === 'custom') {
    return `${normalizeDimension(settings.width)}x${normalizeDimension(settings.height)}`
  }
  if (settings.useCustomRatio) {
    return ratioSize(`${settings.customRatioWidth}:${settings.customRatioHeight}`, settings.aspectBase)
  }
  return ratioSize(settings.aspectRatio, settings.aspectBase)
}

function ratioSize(ratio: string, base: number) {
  const [rawWidth, rawHeight] = ratio.split(':').map(Number)
  if (!rawWidth || !rawHeight) return `${base}x${base}`
  if (rawWidth >= rawHeight) {
    return `${normalizeDimension(base)}x${normalizeDimension(base * rawHeight / rawWidth)}`
  }
  return `${normalizeDimension(base * rawWidth / rawHeight)}x${normalizeDimension(base)}`
}

function normalizeDimension(value: number) {
  const numeric = Number.isFinite(Number(value)) ? Number(value) : 1024
  const clamped = Math.min(3840, Math.max(256, numeric))
  return Math.round(clamped / 16) * 16
}

function parseSize(value: string) {
  const match = /^(\d+)x(\d+)$/i.exec(value.trim())
  if (!match) return null
  return { width: Number(match[1]), height: Number(match[2]) }
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
    await loadGallery()
  } catch (error) {
    errorMessage.value = errorMessageFrom(error, '加载数据失败')
  }
}

async function loadGallery() {
  try {
    const entries = await imageStudioGalleryAPI.list()
    images.value = entries.map(galleryEntryToImage)
  } catch {
    // Gallery storage is optional; generation remains available without it.
  }
}

function galleryEntryToImage(entry: ImageStudioGalleryEntry, index: number): GeneratedImage {
  return {
    id: entry.id,
    title: `历史图片 ${index + 1}`,
    src: entry.url,
    meta: [entry.format?.toUpperCase(), entry.size, entry.model, '7 天内有效'].filter(Boolean).join(' · '),
  }
}

async function persistGeneratedImages(items: GeneratedImage[], prompt: string, model: string) {
  const results = await Promise.allSettled(items.map(async (item) => {
    const imageDataURL = await imageSourceToDataURL(item.src)
    return imageStudioGalleryAPI.save({
      image_data_url: imageDataURL,
      prompt,
      model,
      size: size.value,
      format: format.value,
    })
  }))
  const succeeded = results.filter((result) => result.status === 'fulfilled').length
  const failed = results.length - succeeded
  if (succeeded > 0) {
    await loadGallery()
  }
  if (failed > 0) {
    errorMessage.value = succeeded > 0
      ? '部分图片未能保存到图库，请检查对象存储配置。'
      : '图片已生成，但保存到图库失败。请检查对象存储是否已启用并保存配置。'
  }
}

async function imageSourceToDataURL(src: string) {
  if (src.startsWith('data:')) return src
  const response = await fetch(src)
  if (!response.ok) throw new Error(`Failed to read generated image: ${response.status}`)
  const blob = await response.blob()
  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onerror = () => reject(reader.error || new Error('Failed to read generated image'))
    reader.onload = () => resolve(String(reader.result || ''))
    reader.readAsDataURL(blob)
  })
}

async function loadModels() {
  const requestId = ++modelRequestId
  loadingModels.value = true
  imageModels.value = []
  imageModel.value = ''
  modelLoadError.value = ''
  const key = selectedKey.value
  const endpoint = apiBaseUrl.value
  if (!key || !endpoint) {
    loadingModels.value = false
    return
  }

  try {
    const [batchResult, genericResult] = await Promise.allSettled([
      fetchModelNames(`${endpoint}/images/batches/models`, key.key),
      fetchModelNames(`${endpoint}/models`, key.key, true),
    ])
    if (requestId !== modelRequestId) return

    const names = new Set<string>()
    if (batchResult.status === 'fulfilled') {
      batchResult.value.forEach((name) => names.add(name))
    }
    if (genericResult.status === 'fulfilled') {
      genericResult.value.forEach((name) => names.add(name))
    }

    imageModels.value = Array.from(names).sort((a, b) =>
      a.localeCompare(b, undefined, { numeric: true, sensitivity: 'base' }),
    )
    imageModel.value = imageModels.value[0] || ''

    if (batchResult.status === 'rejected' && genericResult.status === 'rejected') {
      modelLoadError.value = errorMessageFrom(batchResult.reason)
    }
  } catch (error) {
    if (requestId === modelRequestId) {
      modelLoadError.value = errorMessageFrom(error)
    }
  } finally {
    if (requestId === modelRequestId) {
      loadingModels.value = false
    }
  }
}

async function fetchModelNames(endpoint: string, apiKey: string, filterImages = false): Promise<string[]> {
  const response = await fetch(endpoint, {
    headers: { Authorization: `Bearer ${apiKey}` },
  })
  const payload = await response.json().catch(() => null)
  if (!response.ok) {
    throw new Error(payload?.error?.message || payload?.message || `${response.status} ${response.statusText}`)
  }

  const names: string[] = Array.isArray(payload?.data)
    ? payload.data
      .map((item: any) => {
        if (typeof item === 'string') return item
        return item?.id || item?.model || item?.name || ''
      })
      .map((name: unknown) => String(name).trim())
      .filter(Boolean)
    : []
  const uniqueNames = Array.from(new Set(names))
  return filterImages ? uniqueNames.filter(isImageModel) : uniqueNames
}

function errorMessageFrom(error: unknown, fallback = '加载图片模型失败') {
  const message = error instanceof Error ? error.message : ''
  const normalized = message.trim().toLowerCase()
  if (!normalized) return fallback
  if (normalized === 'failed to fetch' || normalized.includes('networkerror')) {
    return '无法连接到 API 服务，请检查网络或 API 地址。'
  }
  if (
    normalized.includes('insufficient account balance') ||
    normalized.includes('insufficient_balance') ||
    normalized.includes('insufficient balance') ||
    normalized.includes('余额不足')
  ) {
    return '账户余额不足，请先充值或换一个有余额的 API Key。'
  }
  if (
    normalized.includes('service temporarily unavailable') ||
    normalized.includes('upstream service temporarily unavailable')
  ) {
    return '当前模型暂时不可用，请稍后重试或换一个模型。'
  }
  return message || fallback
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
    await persistGeneratedImages(images.value, currentPrompt, imageModel.value)
  } catch (error) {
    errorMessage.value = errorMessageFrom(error, '生成失败')
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
  if (appliedSize.mode === 'ratio') {
    return appliedSize.useCustomRatio
      ? `${appliedSize.customRatioWidth}:${appliedSize.customRatioHeight}`
      : appliedSize.aspectRatio
  }
  if (value === 'auto') return '1:1'
  const dimensions = parseSize(value)
  if (!dimensions) return '1:1'
  const divisor = greatestCommonDivisor(dimensions.width, dimensions.height)
  return `${Math.round(dimensions.width / divisor)}:${Math.round(dimensions.height / divisor)}`
}

function greatestCommonDivisor(a: number, b: number): number {
  return b === 0 ? Math.max(1, a) : greatestCommonDivisor(b, a % b)
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

watch([selectedKeyId, apiBaseUrl], () => {
  void loadModels()
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

.toolbar-error {
  font-size: 0.75rem;
  color: rgb(220 38 38);
}

.toolbar-field .toolbar-image-select {
  min-width: 8.5rem;
}

.toolbar-field--select {
  min-height: auto;
  border: 0;
  border-radius: 0;
  background: transparent;
  padding: 0;
}

.toolbar-field .toolbar-image-select--model {
  min-width: 10rem;
}

.composer-size-trigger {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) auto;
  align-items: center;
  gap: 0.4rem;
  min-width: 8.5rem;
  height: 2.25rem;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.65rem;
  background: linear-gradient(180deg, rgb(255 255 255), rgb(240 253 250));
  padding: 0 0.55rem;
  color: rgb(51 65 85);
  text-align: left;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.composer-size-trigger:hover {
  border-color: rgb(45 212 191);
  box-shadow: 0 6px 16px rgb(20 184 166 / 0.14);
}

.composer-size-trigger span {
  font-size: 0.72rem;
  color: rgb(100 116 139);
}

.composer-size-trigger strong {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 0.78rem;
  font-weight: 600;
  color: rgb(30 41 59);
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
  min-width: 7rem;
}

.count-stepper {
  display: grid;
  grid-template-columns: 1.7rem minmax(2.2rem, 1fr) 1.7rem;
  align-items: center;
  height: 2.35rem;
  overflow: hidden;
  border: 1px solid rgb(226 232 240);
  border-radius: 0.75rem;
  background: rgb(255 255 255 / 0.92);
  box-shadow: 0 1px 2px rgb(15 23 42 / 0.04);
}

.count-stepper__button {
  display: flex;
  height: 100%;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: rgb(100 116 139);
  transition:
    background 0.16s ease,
    color 0.16s ease;
}

.count-stepper__button:hover:not(:disabled) {
  background: rgb(240 253 250);
  color: rgb(13 148 136);
}

.count-stepper__button:disabled {
  cursor: not-allowed;
  opacity: 0.35;
}

.count-stepper input {
  width: 100%;
  height: 100%;
  border: 0;
  border-right: 1px solid rgb(241 245 249);
  border-left: 1px solid rgb(241 245 249);
  border-radius: 0;
  background: transparent;
  padding: 0;
  color: rgb(30 41 59);
  font-size: 0.82rem;
  font-weight: 600;
  text-align: center;
  outline: none;
  appearance: textfield;
}

.count-stepper input::-webkit-outer-spin-button,
.count-stepper input::-webkit-inner-spin-button {
  margin: 0;
  appearance: none;
}

.size-dialog {
  --soft-text: rgb(30 41 59);
  --soft-muted: rgb(100 116 139);
  --soft-line: rgb(226 232 240);
  --soft-panel: rgb(248 250 252);
  --soft-accent: rgb(13 148 136);
  color: var(--soft-text);
}

.size-dialog__current {
  margin: -0.25rem 0 1rem;
  font-size: 0.78rem;
  color: var(--soft-muted);
}

.size-dialog__current strong {
  font-weight: 600;
  color: rgb(51 65 85);
}

.size-tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.35rem;
  border-radius: 999px;
  background: rgb(240 253 250);
  padding: 0.3rem;
}

.size-tab {
  min-height: 2.25rem;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: rgb(71 85 105);
  font-size: 0.82rem;
  transition:
    background 0.18s ease,
    color 0.18s ease,
    box-shadow 0.18s ease;
}

.size-tab.active {
  background: rgb(255 255 255);
  color: rgb(15 118 110);
  box-shadow: 0 4px 12px rgb(13 148 136 / 0.14);
}

.size-dialog__content {
  display: flex;
  min-height: 27rem;
  flex-direction: column;
  padding-top: 1.2rem;
}

.size-auto-panel {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem 1rem 4rem;
  text-align: center;
}

.size-auto-icon {
  display: flex;
  width: 3.5rem;
  height: 3.5rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: rgb(204 251 241);
  color: var(--soft-accent);
  box-shadow: 0 10px 26px rgb(13 148 136 / 0.14);
}

.size-auto-title {
  margin-top: 1.1rem;
  font-size: 1rem;
  font-weight: 600;
}

.size-auto-description {
  margin-top: 0.6rem;
  font-size: 0.82rem;
  line-height: 1.8;
  color: rgb(100 116 139);
}

.size-section + .size-section {
  margin-top: 1.15rem;
}

.size-section__label {
  margin-bottom: 0.55rem;
  font-size: 0.76rem;
  color: var(--soft-muted);
}

.size-base-options {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.65rem;
}

.size-base-option,
.size-ratio-option,
.size-custom-ratio-toggle {
  border: 1px solid var(--soft-line);
  background: rgb(255 255 255 / 0.72);
  color: rgb(71 85 105);
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.size-base-option {
  min-height: 2.45rem;
  border-radius: 0.65rem;
  font-size: 0.82rem;
}

.size-base-option.active,
.size-ratio-option.active,
.size-custom-ratio-toggle.active {
  border-color: rgb(45 212 191);
  background: rgb(240 253 250);
  box-shadow: 0 6px 18px rgb(13 148 136 / 0.13);
}

.size-ratio-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.65rem;
}

.size-ratio-option {
  display: flex;
  min-height: 4.2rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  border-radius: 0.75rem;
  font-size: 0.76rem;
}

.size-ratio-shape {
  display: block;
  width: 1.75rem;
  max-height: 1.8rem;
  border: 1.5px solid currentColor;
  border-radius: 0.25rem;
  opacity: 0.7;
}

.size-custom-ratio-toggle {
  width: 100%;
  min-height: 2.35rem;
  margin-top: 0.75rem;
  border-radius: 0.65rem;
  font-size: 0.78rem;
}

.size-custom-ratio {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 0.6rem;
  margin-top: 0.65rem;
}

.size-custom-ratio input,
.size-dimension-row input {
  width: 100%;
  height: 2.45rem;
  border: 1px solid var(--soft-line);
  border-radius: 0.65rem;
  background: rgb(255 255 255 / 0.75);
  padding: 0 0.7rem;
  color: rgb(30 41 59);
  font-size: 0.84rem;
  outline: none;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.size-custom-ratio input:focus,
.size-dimension-row input:focus {
  border-color: rgb(45 212 191);
  box-shadow: 0 0 0 3px rgb(20 184 166 / 0.13);
}

.size-dimension-row {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: end;
  gap: 0.75rem;
}

.size-dimension-row label {
  display: grid;
  gap: 0.45rem;
  font-size: 0.76rem;
  color: var(--soft-muted);
}

.size-dimension-separator {
  padding-bottom: 0.55rem;
  color: rgb(148 163 184);
}

.size-dimension-hint {
  margin-top: 0.9rem;
  border-left: 3px solid rgb(94 234 212);
  border-radius: 0.35rem;
  background: rgb(240 253 250);
  padding: 0.7rem 0.85rem;
  font-size: 0.72rem;
  line-height: 1.65;
  color: rgb(71 85 105);
}

.size-preview {
  display: grid;
  gap: 0.25rem;
  margin-top: auto;
  border-radius: 0.85rem;
  background: rgb(240 253 250);
  padding: 0.9rem 1rem;
}

.size-preview span {
  font-size: 0.72rem;
  color: var(--soft-muted);
}

.size-preview strong {
  font-size: 0.96rem;
  font-weight: 600;
  color: rgb(15 23 42);
}

.size-dialog__footer {
  display: grid;
  width: 100%;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.size-dialog__cancel,
.size-dialog__confirm {
  min-height: 2.7rem;
  border: 0;
  border-radius: 0.7rem;
  font-size: 0.86rem;
  transition:
    background 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.size-dialog__cancel {
  background: rgb(241 245 249);
  color: rgb(51 65 85);
}

.size-dialog__confirm {
  background: rgb(13 148 136);
  color: white;
  box-shadow: 0 8px 18px rgb(13 148 136 / 0.2);
}

.size-dialog__cancel:hover,
.size-dialog__confirm:hover {
  transform: translateY(-1px);
}

.size-dialog__confirm:disabled {
  cursor: not-allowed;
  opacity: 0.5;
  transform: none;
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

:global(.dark .composer-size-trigger) {
  border-color: rgb(51 65 85);
  background: rgb(30 41 59);
  color: rgb(241 245 249);
}

:global(.dark .composer-size-trigger span) {
  color: rgb(148 163 184);
}

:global(.dark .composer-size-trigger strong) {
  color: rgb(241 245 249);
}

:global(.dark .count-stepper) {
  border-color: rgb(51 65 85);
  background: rgb(30 41 59);
}

:global(.dark .count-stepper input) {
  border-color: rgb(51 65 85);
  color: rgb(241 245 249);
}

:global(.dark .size-dialog) {
  --soft-text: rgb(241 245 249);
  --soft-muted: rgb(148 163 184);
  --soft-line: rgb(51 65 85);
  --soft-panel: rgb(30 41 59);
  --soft-accent: rgb(45 212 191);
}

:global(.dark .size-tabs) {
  background: rgb(19 78 74 / 0.55);
}

:global(.dark .size-tab.active),
:global(.dark .size-base-option),
:global(.dark .size-ratio-option),
:global(.dark .size-custom-ratio-toggle),
:global(.dark .size-custom-ratio input),
:global(.dark .size-dimension-row input) {
  background: rgb(30 41 59);
  color: rgb(226 232 240);
}

:global(.dark .size-preview),
:global(.dark .size-dimension-hint),
:global(.dark .size-dialog__cancel) {
  background: rgb(30 41 59);
}

:global(.dark .size-auto-icon) {
  background: rgb(19 78 74 / 0.65);
}
</style>
