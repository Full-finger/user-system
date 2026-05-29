<template>
  <div class="auth-page">
    <div class="auth-card card fade-up">
      <div class="auth-card__header">
        <PhSparkle :size="28" weight="fill" class="auth-card__logo" />
        <h1 class="auth-card__title font-display">登录 DevMoe</h1>
        <p class="auth-card__desc text-3">欢迎回到社区</p>
      </div>

      <form @submit.prevent="handleLogin" class="auth-form">
        <div class="auth-form__group">
          <label class="auth-form__label">用户名</label>
          <input
            v-model="form.username"
            type="text"
            class="input"
            :class="{ 'input--error': error }"
            placeholder="请输入用户名"
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
            placeholder="请输入密码"
            autocomplete="current-password"
          />
        </div>

        <div v-if="error" class="auth-form__error">
          <PhXCircle :size="14" />
          {{ error }}
        </div>

        <button type="submit" class="btn btn--primary auth-form__submit" :disabled="loading">
          <PhCircleNotch v-if="loading" :size="18" class="spin" />
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>

      <div class="auth-card__footer">
        <span class="text-3">还没有账号？</span>
        <router-link to="/register">注册</router-link>
      </div>

      <div class="auth-card__divider">
        <span class="text-4">或者</span>
      </div>

      <div class="auth-card__alt">
        <p class="text-3" style="font-size: 13px">
          <PhEnvelopeSimple :size="14" style="vertical-align: -2px" />
          验证码登录（开发中）
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { PhSparkle, PhXCircle, PhCircleNotch, PhEnvelopeSimple } from '@phosphor-icons/vue'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const form = reactive({ username: '', password: '' })
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  error.value = ''
  if (!form.username || !form.password) {
    error.value = '请填写用户名和密码'
    return
  }
  loading.value = true
  try {
    await auth.login(form)
    const redirect = route.query.redirect || '/'
    router.push(redirect)
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

.auth-card__divider {
  text-align: center;
  margin: 20px 0 16px;
  position: relative;
}

.auth-card__divider::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background: var(--border);
}

.auth-card__divider span {
  background: var(--bg-card);
  padding: 0 12px;
  position: relative;
  font-size: 12px;
}

.auth-card__alt {
  text-align: center;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spin {
  animation: spin 0.8s linear infinite;
}
</style>
