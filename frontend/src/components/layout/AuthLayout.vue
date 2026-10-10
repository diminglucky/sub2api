<template>
  <div class="auth-split">
    <!-- 左侧品牌区：暖色渐变 + 网格，与右侧表单形成对比 -->
    <aside v-if="showBrand" class="auth-aside">
      <div class="auth-aside__glow" aria-hidden="true"></div>
      <div class="auth-aside__grid" aria-hidden="true"></div>

      <div class="auth-aside__inner">
        <div class="auth-brand-pill">
          <span class="auth-brand-pill__logo">
            <img :src="siteLogo || '/logo.svg'" alt="Logo" />
          </span>
          <span class="auth-brand-pill__text">
            <strong>{{ siteName }}</strong>
            <small>{{ siteSubtitle }}</small>
          </span>
        </div>

        <div class="auth-aside__hero">
          <p class="auth-aside__eyebrow">统一 AI 访问</p>
          <h1 class="auth-aside__title">
            <span>THE AI</span>
            <span>GATEWAY</span>
          </h1>
          <p class="auth-aside__lede">
            统一管理多家 AI 服务商、模型路由、额度与控制台访问。
          </p>
          <div class="auth-aside__cards">
            <div><strong>Multi</strong><span>AI 服务商</span></div>
            <div><strong>API</strong><span>统一网关</span></div>
            <div><strong>Quota</strong><span>额度控制</span></div>
          </div>
        </div>
      </div>
    </aside>

    <!-- 右侧表单区 -->
    <main class="auth-main">
      <div class="auth-main__inner" :class="contentClass">
        <div :class="cardClass">
          <slot />
        </div>

        <div :class="footerClass">
          <slot name="footer" />
        </div>

        <div v-if="showCopyright" :class="copyrightClass">
          &copy; {{ currentYear }} {{ siteName }}
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAppStore } from '@/stores'
import { sanitizeUrl } from '@/utils/url'

const props = withDefaults(defineProps<{
  showBrand?: boolean
  showCopyright?: boolean
  compact?: boolean
}>(), {
  showBrand: true,
  showCopyright: true,
  compact: false
})

const appStore = useAppStore()

const siteName = computed(() => appStore.siteName || 'SuperAI')
const siteLogo = computed(() => sanitizeUrl(appStore.siteLogo || '', { allowRelative: true, allowDataUrl: true }))
// 品牌胶囊里的副标题要短；过长的站点副标题回退到通用文案，避免撑破布局。
const siteSubtitle = computed(() => {
  const raw = (appStore.cachedPublicSettings?.site_subtitle || '').trim()
  return raw && raw.length <= 14 ? raw : 'AI 网关'
})

const currentYear = computed(() => new Date().getFullYear())
const contentClass = computed(() => [
  props.compact ? 'max-w-[390px]' : 'max-w-md'
])
const cardClass = computed(() => [
  'card-glass rounded-2xl shadow-glass',
  props.compact ? 'p-4 sm:p-5' : 'p-6 sm:p-8'
])
const footerClass = computed(() => [
  'text-center text-sm',
  props.compact ? 'mt-3' : 'mt-6'
])
const copyrightClass = computed(() => [
  'text-center text-xs text-gray-400 dark:text-dark-500',
  props.compact ? 'mt-4' : 'mt-8'
])

onMounted(() => {
  appStore.fetchPublicSettings()
})
</script>

<style scoped>
.auth-split {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  min-height: 100vh;
}

.auth-aside {
  position: relative;
  overflow: hidden;
  isolation: isolate;
  background: #fbf7f1;
}

.auth-aside__glow {
  position: absolute;
  inset: 0;
  z-index: 0;
  background:
    linear-gradient(118deg, rgba(255, 255, 255, 0) 34%, rgba(249, 168, 96, 0.42) 62%, rgba(252, 216, 176, 0.65) 100%),
    linear-gradient(160deg, #fdfaf6 0%, #fbf3e8 60%, #f7ecd9 100%);
}

.auth-aside__grid {
  position: absolute;
  inset: 0;
  z-index: 0;
  background-image:
    linear-gradient(rgba(15, 23, 42, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.05) 1px, transparent 1px);
  background-size: 84px 84px;
  mask-image: linear-gradient(180deg, rgba(0, 0, 0, 0.9), transparent 90%);
}

.auth-aside__inner {
  position: relative;
  z-index: 1;
  display: flex;
  min-height: 100vh;
  flex-direction: column;
  justify-content: space-between;
  gap: 48px;
  padding: 40px 56px 56px;
}

.auth-brand-pill {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  align-self: flex-start;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.72);
  padding: 8px 16px 8px 10px;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.06);
  backdrop-filter: blur(10px);
}

.auth-brand-pill__logo {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  overflow: hidden;
  border-radius: 10px;
}

.auth-brand-pill__logo img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.auth-brand-pill__text {
  display: grid;
  line-height: 1.15;
}

.auth-brand-pill__text strong {
  color: #111827;
  font-size: 17px;
  font-weight: 800;
}

.auth-brand-pill__text small {
  color: #6b7280;
  font-size: 11px;
}

.auth-aside__hero {
  max-width: 520px;
}

.auth-aside__eyebrow {
  position: relative;
  margin: 0 0 22px;
  padding-left: 16px;
  color: #ea580c;
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.auth-aside__eyebrow::before {
  position: absolute;
  top: 50%;
  left: 0;
  width: 4px;
  height: 18px;
  border-radius: 999px;
  background: #ea580c;
  transform: translateY(-50%);
  content: '';
}

.auth-aside__title {
  margin: 0;
  color: #0f172a;
  font-size: clamp(48px, 5.2vw, 76px);
  font-weight: 900;
  line-height: 0.98;
  letter-spacing: -0.02em;
  text-transform: uppercase;
}

.auth-aside__title span {
  display: block;
}

.auth-aside__title span:last-child {
  color: #334155;
}

.auth-aside__lede {
  max-width: 380px;
  margin: 26px 0 0;
  color: #52525b;
  font-size: 15px;
  line-height: 1.85;
}

.auth-aside__cards {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
  max-width: 380px;
  margin-top: 34px;
}

.auth-aside__cards > div {
  display: grid;
  gap: 4px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.7);
  padding: 14px 16px;
  box-shadow: 0 10px 26px rgba(15, 23, 42, 0.05);
}

.auth-aside__cards strong {
  color: #111827;
  font-size: 17px;
  font-weight: 800;
}

.auth-aside__cards span {
  color: #6b7280;
  font-size: 12px;
}

.auth-main {
  display: grid;
  place-items: center;
  padding: 28px 24px;
  background: #f7f7f8;
}

:global(.dark) .auth-main {
  background: #0b1120;
}

.auth-main__inner {
  width: 100%;
}

@media (max-width: 960px) {
  .auth-split {
    grid-template-columns: 1fr;
  }

  .auth-aside {
    display: none;
  }

  .auth-main {
    min-height: 100vh;
  }
}
</style>
