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
    error.value = err instanceof Error ? err.message : 'Registration failed'
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section class="auth-shell">
    <form class="auth-card" @submit.prevent="submit">
      <p class="eyebrow">{{ site?.name || 'Draw' }}</p>
      <h1>Register</h1>
      <p class="lede">Create an account for this downstream site.</p>

      <p v-if="error" class="form-error">{{ error }}</p>

      <div class="field">
        <label for="email">Email</label>
        <input id="email" v-model="form.email" type="email" autocomplete="email" required />
      </div>
      <div class="field">
        <label for="password">Password</label>
        <input id="password" v-model="form.password" type="password" autocomplete="new-password" required />
      </div>

      <button class="primary-button full-button" type="submit" :disabled="submitting">
        {{ submitting ? 'Creating account...' : 'Create account' }}
      </button>
      <p class="auth-meta">
        Already registered?
        <RouterLink to="/login">Sign in</RouterLink>
      </p>
    </form>
  </section>
</template>
