<template>
  <div class="auth-page">
    <div class="auth-card card fade-up">
      <div class="auth-card__header">
        <img v-if="siteConfig.siteLogo" :src="siteConfig.siteLogo" alt="Logo" class="auth-card__logo-img" />
        <PhSparkle v-else :size="28" weight="fill" class="auth-card__logo" />
        <h1 class="auth-card__title font-display">登录 {{ siteConfig.siteName }}</h1>
        <p class="auth-card__desc text-3">欢迎回到社区</p>
      </div>

      <!-- 密码登录 -->
      <form v-if="mode === 'password'" @submit.prevent="handleLogin" class="auth-form">
        <div class="auth-form__group">
          <label class="auth-form__label">用户名</label>
          <input
            v-model="pwForm.username"
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
            v-model="pwForm.password"
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

      <!-- 验证码登录 -->
      <form v-else @submit.prevent="handleCodeLogin" class="auth-form">
        <div class="auth-form__group">
          <label class="auth-form__label">邮箱</label>
          <input
            v-model="codeForm.email"
            type="email"
            class="input"
            :class="{ 'input--error': error }"
            placeholder="请输入邮箱"
            autocomplete="email"
          />
        </div>

        <div class="auth-form__group">
          <label class="auth-form__label">验证码</label>
          <div class="auth-form__code-row">
            <input
              v-model="codeForm.code"
              type="text"
              class="input"
              :class="{ 'input--error': error }"
              placeholder="6 位验证码"
              maxlength="6"
            />
            <button
              type="button"
              class="btn btn--outline btn--sm"
              :disabled="codeCooldown > 0 || sendingCode"
              @click="handleSendCode"
            >
              {{ sendingCode ? '发送中' : codeCooldown > 0 ? `${codeCooldown}s` : '发送验证码' }}
            </button>
          </div>
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
          <a href="#" @click.prevent="toggleMode" style="margin-left: 4px">
            {{ mode === 'password' ? '验证码登录' : '密码登录' }}
          </a>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { siteConfig } from '../config/site'
import { sendCode } from '../api'
import { PhSparkle, PhXCircle, PhCircleNotch, PhEnvelopeSimple } from '@phosphor-icons/vue'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const mode = ref('password')
const loading = ref(false)
const error = ref('')

const pwForm = reactive({ username: '', password: '' })
const codeForm = reactive({ email: '', code: '' })
const codeCooldown = ref(0)
const sendingCode = ref(false)

function toggleMode() {
  mode.value = mode.value === 'password' ? 'code' : 'password'
  error.value = ''
}

async function handleLogin() {
  error.value = ''
  if (!pwForm.username || !pwForm.password) {
    error.value = '请填写用户名和密码'
    return
  }
  loading.value = true
  try {
    await auth.login(pwForm)
    const redirect = route.query.redirect || '/'
    router.push(redirect)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function handleSendCode() {
  error.value = ''
  if (!codeForm.email) { error.value = '请输入邮箱'; return }
  sendingCode.value = true
  try {
    await sendCode({ email: codeForm.email })
    codeCooldown.value = 60
    const timer = setInterval(() => {
      codeCooldown.value--
      if (codeCooldown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (e) { error.value = e.message }
  finally { sendingCode.value = false }
}

async function handleCodeLogin() {
  error.value = ''
  if (!codeForm.email || !codeForm.code) { error.value = '请填写邮箱和验证码'; return }
  loading.value = true
  try {
    await auth.loginByCode({ email: codeForm.email, code: codeForm.code })
    const redirect = route.query.redirect || '/'
    router.push(redirect)
  } catch (e) { error.value = e.message }
  finally { loading.value = false }
}
</script>

<style scoped>
.auth-card__alt {
  text-align: center;
}

.auth-card__alt a {
  color: var(--accent);
  text-decoration: none;
}

.auth-card__alt a:hover {
  text-decoration: underline;
}

.auth-card__logo-img {
  width: 28px;
  height: 28px;
  object-fit: contain;
  border-radius: var(--radius-s);
  margin-bottom: 8px;
}
</style>
