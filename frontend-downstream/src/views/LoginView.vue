<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { login, type SiteConfig } from '../api'

defineProps<{ site?: SiteConfig }>()

const route = useRoute()
const router = useRouter()
const form = reactive({ email: '', password: '' })
const error = ref('')
const submitting = ref(false)

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    await login(form.email, form.password)
    window.dispatchEvent(new CustomEvent('draw-session-change', { detail: true }))
    await router.push(String(route.query.redirect || '/dashboard'))
  } catch (err) {
    error.value = err instanceof Error ? err.message : '登录失败'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="auth-shell">
    <form class="auth-card" @submit.prevent="submit">
      <p class="eyebrow">{{ site?.name || 'Draw' }}</p>
      <h1>欢迎回来</h1>
      <p class="lede">登录你的账号，继续管理余额、充值和 API 调用。</p>

      <p v-if="error" class="form-error">{{ error }}</p>

      <div class="field">
        <label for="email">邮箱</label>
        <input id="email" v-model="form.email" type="email" autocomplete="email" required />
      </div>
      <div class="field">
        <label for="password">密码</label>
        <input id="password" v-model="form.password" type="password" autocomplete="current-password" required />
      </div>

      <button class="primary-button full-button" type="submit" :disabled="submitting">
        {{ submitting ? '正在登录...' : '登录' }}
      </button>
      <p class="auth-meta">
        还没有账号？
        <RouterLink to="/register">立即注册</RouterLink>
      </p>
    </form>
  </section>
</template>
