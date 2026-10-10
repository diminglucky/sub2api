<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getSite, type SiteConfig } from './api'

// 子站前端只负责“入口”这一屏。登录后的所有用户页面直接复用主站前端，
// 所以这里的链接一律用真实的 <a href>，触发整页跳转到主站应用。
const site = ref<SiteConfig>({
  id: 0,
  slug: 'draw',
  domain: 'draw.superai.sbs',
  name: 'Draw',
  logo_url: '',
  theme_color: '#fb6415',
  api_base_url: 'https://draw.superai.sbs/v1'
})
const siteError = ref('')

const themeStyle = computed(() => ({
  '--site-color': site.value.theme_color || '#fb6415'
}))

onMounted(async () => {
  try {
    site.value = await getSite()
  } catch (error) {
    siteError.value = error instanceof Error ? error.message : '站点信息加载失败'
  }
})
</script>

<template>
  <div class="app-shell" :style="themeStyle">
    <header class="site-header">
      <a class="brand" href="/">
        <img v-if="site.logo_url" :src="site.logo_url" alt="" width="38" height="38" />
        <span v-else class="brand-mark">D</span>
        <span class="brand-name">{{ site.name }}</span>
      </a>

      <nav class="site-nav" aria-label="主导航">
        <a class="nav-link" href="/image-studio">图片工作台</a>
        <a class="nav-link" href="/dashboard">用户中心</a>
        <a class="nav-link nav-link--primary" href="/login">登录</a>
      </nav>
    </header>

    <main class="page">
      <div v-if="siteError" class="notice">{{ siteError }}</div>
      <RouterView v-slot="{ Component }">
        <component :is="Component" :site="site" />
      </RouterView>
    </main>
  </div>
</template>
