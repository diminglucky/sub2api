<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { register, type SiteConfig } from '../api'

defineProps<{ site?: SiteConfig }>()

const router = useRouter()
const form = reactive({ email: '', password: '' })
const error = ref('')
const submitting = ref(false)

async function submit() {
  error.value = ''
  submitting.value = true
  try {
    await register(form.email, form.password)
    window.dispatchEvent(new CustomEvent('draw-session-change', { detail: true }))
    await router.push('/dashboard')
  } catch (err) {
    error.value = err instanceof Error ? err.message : '注册失败'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="auth-shell">
    <form class="auth-card" @submit.prevent="submit">
      <p class="eyebrow">{{ site?.name || 'Draw' }}</p>
      <h1>创建账号</h1>
      <p class="lede">注册后即可进入控制台，查看账号余额、充值与调用数据。</p>

      <p v-if="error" class="form-error">{{ error }}</p>

      <div class="field">
        <label for="email">邮箱</label>
        <input id="email" v-model="form.email" type="email" autocomplete="email" required />
      </div>
      <div class="field">
        <label for="password">密码</label>
        <input id="password" v-model="form.password" type="password" autocomplete="new-password" required />
      </div>

      <button class="primary-button full-button" type="submit" :disabled="submitting">
        {{ submitting ? '正在注册...' : '注册账号' }}
      </button>
      <p class="auth-meta">
        已有账号？
        <RouterLink to="/login">返回登录</RouterLink>
      </p>
    </form>
  </section>
</template>
