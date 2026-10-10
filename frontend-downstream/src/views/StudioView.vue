<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { listApiKeys, type ApiKeyInfo, type SiteConfig } from '../api'
import {
  DRAW_FORMATS,
  DRAW_QUALITIES,
  DRAW_SIZES,
  editImage,
  errorMessageFrom,
  generateImage,
  listImageModels,
  loadGallery,
  saveGallery,
  type DrawFormat,
  type DrawResult
} from '../draw'

const props = defineProps<{ site?: SiteConfig }>()

const apiKeys = ref<ApiKeyInfo[]>([])
const selectedKeyId = ref('')
const models = ref<string[]>([])
const model = ref('')
const prompt = ref('')
const size = ref(DRAW_SIZES[0].value)
const quality = ref(DRAW_QUALITIES[0].value)
const format = ref<DrawFormat>('png')
const count = ref(1)
const reference = ref<File | null>(null)
const referencePreview = ref('')
const results = ref<DrawResult[]>([])
const generating = ref(false)
const loadingKeys = ref(false)
const loadingModels = ref(false)
const error = ref('')
const preview = ref<DrawResult | null>(null)

const activeKeys = computed(() => apiKeys.value.filter((key) => key.status === 'active'))
const selectedKey = computed(
  () => activeKeys.value.find((key) => String(key.id) === selectedKeyId.value) || null
)
const canGenerate = computed(
  () => Boolean(selectedKey.value?.key && model.value && prompt.value.trim()) && !generating.value
)

onMounted(async () => {
  results.value = loadGallery()
  await loadKeys()
})

watch(selectedKeyId, () => {
  void loadModels()
})

async function loadKeys() {
  loadingKeys.value = true
  error.value = ''
  try {
    const page = await listApiKeys()
    apiKeys.value = page.items || []
    if (!selectedKeyId.value && activeKeys.value.length) {
      selectedKeyId.value = String(activeKeys.value[0].id)
    }
  } catch (err) {
    error.value = errorMessageFrom(err, '无法加载 API Key')
  } finally {
    loadingKeys.value = false
  }
}

async function loadModels() {
  models.value = []
  model.value = ''
  if (!selectedKey.value?.key) return
  loadingModels.value = true
  error.value = ''
  try {
    const names = await listImageModels(selectedKey.value.key, props.site)
    models.value = names
    model.value = names[0] || ''
    if (!names.length) {
      error.value = '当前密钥没有可用的图片模型，请换一个密钥。'
    }
  } catch (err) {
    error.value = errorMessageFrom(err, '无法加载图片模型')
  } finally {
    loadingModels.value = false
  }
}

function onReferenceChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  if (!['image/png', 'image/jpeg', 'image/webp'].includes(file.type)) {
    error.value = '参考图仅支持 PNG、JPEG 或 WEBP。'
    return
  }
  clearReference()
  reference.value = file
  const reader = new FileReader()
  reader.onload = () => {
    referencePreview.value = String(reader.result || '')
  }
  reader.readAsDataURL(file)
}

function clearReference() {
  reference.value = null
  referencePreview.value = ''
}

async function generate() {
  if (!canGenerate.value || !selectedKey.value) return
  generating.value = true
  error.value = ''
  const request = {
    model: model.value.trim(),
    prompt: prompt.value.trim(),
    size: size.value,
    quality: quality.value,
    format: format.value,
    count: Math.min(Math.max(Number(count.value) || 1, 1), 4)
  }
  try {
    const generated = reference.value
      ? await editImage(selectedKey.value.key, request, reference.value, props.site)
      : await generateImage(selectedKey.value.key, request, props.site)
    if (!generated.length) {
      throw new Error('接口没有返回图片，请调整提示词后重试。')
    }
    results.value = [...generated, ...results.value]
    saveGallery(results.value)
    clearReference()
  } catch (err) {
    error.value = errorMessageFrom(err)
  } finally {
    generating.value = false
  }
}

function download(item: DrawResult) {
  const link = document.createElement('a')
  link.href = item.src
  link.download = `draw-${item.createdAt}.${item.format}`
  link.click()
}

function clearGallery() {
  results.value = []
  saveGallery([])
}

function formatTime(value: number): string {
  if (!value) return ''
  const date = new Date(value)
  return `${date.getMonth() + 1}/${date.getDate()} ${String(date.getHours()).padStart(2, '0')}:${String(
    date.getMinutes()
  ).padStart(2, '0')}`
}
</script>

<template>
  <section class="studio">
    <header class="page-heading">
      <div>
        <p class="eyebrow">{{ site?.name || 'Draw' }} · 绘图工作台</p>
        <h1>AI 图片生成</h1>
        <p class="lede">输入提示词即可生成图片，也可以上传参考图进行图生图。</p>
      </div>
      <div class="studio-toolbar">
        <label class="studio-field">
          <span>API Key</span>
          <select v-model="selectedKeyId" :disabled="loadingKeys || !activeKeys.length">
            <option v-if="!activeKeys.length" value="">
              {{ loadingKeys ? '正在加载...' : '暂无可用密钥' }}
            </option>
            <option v-for="key in activeKeys" :key="key.id" :value="String(key.id)">
              {{ key.name }}
            </option>
          </select>
        </label>
        <label class="studio-field">
          <span>模型</span>
          <select v-model="model" :disabled="loadingModels || !models.length">
            <option v-if="!models.length" value="">
              {{ loadingModels ? '正在加载...' : '未找到图片模型' }}
            </option>
            <option v-for="name in models" :key="name" :value="name">{{ name }}</option>
          </select>
        </label>
        <button
          class="icon-button"
          type="button"
          title="刷新密钥和模型"
          :disabled="loadingKeys || loadingModels"
          @click="loadKeys().then(loadModels)"
        >
          刷新
        </button>
      </div>
    </header>

    <p v-if="error" class="form-error">{{ error }}</p>

    <div v-if="!activeKeys.length && !loadingKeys" class="studio-empty">
      <h2>还没有可用的 API Key</h2>
      <p>绘图接口使用 API Key 鉴权。请先在账户中创建密钥，再回到这里生成图片。</p>
    </div>

    <div v-else-if="!results.length" class="studio-empty">
      <h2>输入提示词开始生成图片</h2>
      <p>支持文生图，也可以上传参考图做图生图。生成结果会自动保存在当前浏览器。</p>
    </div>

    <div v-else class="studio-gallery">
      <div class="studio-gallery__head">
        <span>生成结果 · {{ results.length }}</span>
        <button class="text-button" type="button" @click="clearGallery">清空记录</button>
      </div>
      <div class="studio-grid">
        <article v-for="item in results" :key="item.createdAt + item.src.slice(-12)" class="studio-card">
          <button class="studio-card__preview" type="button" @click="preview = item">
            <img :src="item.src" :alt="item.prompt" loading="lazy" />
          </button>
          <div class="studio-card__meta">
            <p class="studio-card__prompt">{{ item.prompt }}</p>
            <p class="studio-card__sub">{{ item.size }} · {{ item.model }} · {{ formatTime(item.createdAt) }}</p>
          </div>
          <button class="studio-card__download" type="button" title="下载图片" @click="download(item)">
            下载
          </button>
        </article>
      </div>
    </div>

    <form class="studio-composer" @submit.prevent="generate">
      <div v-if="referencePreview" class="studio-reference">
        <img :src="referencePreview" alt="参考图" />
        <span>已选择参考图</span>
        <button type="button" @click="clearReference">移除</button>
      </div>

      <textarea
        v-model="prompt"
        rows="3"
        placeholder="描述你想生成的图片，例如：一只戴着宇航头盔的橘猫，卡通风格，柔和光线"
      ></textarea>

      <div class="studio-controls">
        <label class="studio-field">
          <span>尺寸</span>
          <select v-model="size">
            <option v-for="option in DRAW_SIZES" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </label>
        <label class="studio-field">
          <span>质量</span>
          <select v-model="quality">
            <option v-for="option in DRAW_QUALITIES" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </label>
        <label class="studio-field">
          <span>格式</span>
          <select v-model="format">
            <option v-for="option in DRAW_FORMATS" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </label>
        <label class="studio-field studio-field--narrow">
          <span>数量</span>
          <select v-model.number="count">
            <option v-for="n in 4" :key="n" :value="n">{{ n }}</option>
          </select>
        </label>

        <label class="studio-upload">
          <input type="file" accept="image/png,image/jpeg,image/webp" @change="onReferenceChange" />
          <span>上传参考图</span>
        </label>

        <button class="primary-button studio-submit" type="submit" :disabled="!canGenerate">
          {{ generating ? '正在生成...' : '开始生成' }}
        </button>
      </div>
    </form>
  </section>

  <div v-if="preview" class="studio-lightbox" @click.self="preview = null">
    <div class="studio-lightbox__body">
      <img :src="preview.src" :alt="preview.prompt" />
      <div class="studio-lightbox__bar">
        <p>{{ preview.prompt }}</p>
        <button class="primary-button" type="button" @click="download(preview)">下载</button>
        <button class="secondary-button" type="button" @click="preview = null">关闭</button>
      </div>
    </div>
  </div>
</template>
