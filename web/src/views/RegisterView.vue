<template>
  <div class="auth-page">
    <div class="auth-card card fade-up">
      <div class="auth-card__header">
        <PhSparkle :size="28" weight="fill" class="auth-card__logo" />
        <h1 class="auth-card__title font-display">注册账号</h1>
        <p class="auth-card__desc text-3">加入 DevMoe 社区</p>
      </div>

      <form @submit.prevent="handleRegister" class="auth-form">
        <div class="auth-form__group">
          <label class="auth-form__label">用户名</label>
          <input
            v-model="form.username"
            type="text"
            class="input"
            :class="{ 'input--error': error }"
            placeholder="3-50 个字符"
            autocomplete="username"
          />
        </div>

        <div class="auth-form__group">
          <label class="auth-form__label">密码</label>
          <input
            v-model="form.password"
            type="password"
            class="input"
            :class="{ 'input--error': error }"
            placeholder="至少 6 位"
            autocomplete="new-password"
          />
        </div>

        <div class="auth-form__group">
          <label class="auth-form__label">确认密码</label>
          <input
            v-model="form.confirmPassword"
            type="password"
            class="input"
            :class="{ 'input--error': error }"
            placeholder="再次输入密码"
            autocomplete="new-password"
          />
        </div>

        <div v-if="error" class="auth-form__error">
          <PhXCircle :size="14" />
          {{ error }}
        </div>

        <div v-if="success" class="auth-form__success">
          <PhCheckCircle :size="14" />
          {{ success }}
        </div>

        <button type="submit" class="btn btn--primary auth-form__submit" :disabled="loading">
          <PhCircleNotch v-if="loading" :size="18" class="spin" />
          {{ loading ? '注册中...' : '注册' }}
        </button>
      </form>

      <div class="auth-card__footer">
        <span class="text-3">已有账号？</span>
        <router-link to="/login">登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { PhSparkle, PhXCircle, PhCheckCircle, PhCircleNotch } from '@phosphor-icons/vue'

const auth = useAuthStore()
const router = useRouter()

const form = reactive({ username: '', password: '', confirmPassword: '' })
const loading = ref(false)
const error = ref('')
const success = ref('')

async function handleRegister() {
  error.value = ''
  success.value = ''

  if (!form.username || !form.password) {
    error.value = '请填写用户名和密码'
    return
  }
  if (form.username.length < 3) {
    error.value = '用户名至少 3 个字符'
    return
  }
  if (form.password.length < 6) {
    error.value = '密码至少 6 位'
    return
  }
  if (form.password !== form.confirmPassword) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true
  try {
    await auth.register({ username: form.username, password: form.password })
    success.value = '注册成功，正在跳转登录...'
    setTimeout(() => router.push('/login'), 1500)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: var(--bg-page);
  padding: 24px;
}

.auth-card {
  width: 100%;
  max-width: 400px;
  padding: 32px;
  background: var(--bg-card);
}

.auth-card__header {
  text-align: center;
  margin-bottom: 28px;
}

.auth-card__logo {
  color: var(--accent);
  margin-bottom: 8px;
}

.auth-card__title {
  font-size: 22px;
  margin-bottom: 4px;
}

.auth-card__desc {
  font-size: 14px;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.auth-form__group {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  background: var(--bg-muted);
  border-radius: var(--radius-m);
  transition: background var(--duration-medium-2) var(--ease-standard);
}

.auth-form__group:focus-within {
  background: var(--bg-card);
  box-shadow: 0 0 0 2px var(--accent-glow);
}

.auth-form__label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-3);
  letter-spacing: 0.3px;
  text-transform: uppercase;
}

.auth-form__group .input:focus {
  box-shadow: none;
}

.auth-form__error {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #c47878;
}

.auth-form__success {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--mint);
}

.auth-form__submit {
  width: 100%;
  height: 42px;
  margin-top: 4px;
}

.auth-card__footer {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spin {
  animation: spin 0.8s linear infinite;
}
</style>
