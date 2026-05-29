<template>
  <div class="auth-page">
    <div class="auth-card card fade-up">
      <div class="auth-card__header">
        <PhSparkle :size="28" weight="fill" class="auth-card__logo" />
        <h1 class="auth-card__title font-display">注册账号</h1>
        <p class="auth-card__desc text-3">加入 DevMoe 社区</p>
      </div>

      <!-- 步骤指示器 -->
      <div class="step-indicator">
        <span class="step-indicator__dot" :class="{ 'step-indicator__dot--active': step === 1 }">1</span>
        <span class="step-indicator__line" :class="{ 'step-indicator__line--active': step === 2 }"></span>
        <span class="step-indicator__dot" :class="{ 'step-indicator__dot--active': step === 2 }">2</span>
      </div>

      <form @submit.prevent class="auth-form">
        <Transition name="step" mode="out-in">
          <!-- 步骤 1：用户名 + 密码 -->
          <div v-if="step === 1" key="step1" class="auth-form__step">
            <div class="auth-form__group">
              <label class="auth-form__label">用户名</label>
              <input
                v-model="form.username"
                type="text"
                class="input"
                :class="{ 'input--error': error }"
                placeholder="3-50 个字符"
                autocomplete="username"
                @blur="handleCheckUsername"
              />
              <span v-if="usernameStatus === 'available'" class="auth-form__hint auth-form__hint--ok">
                <PhCheck :size="14" /> 用户名可用
              </span>
              <span v-else-if="usernameStatus === 'taken'" class="auth-form__hint auth-form__hint--err">
                <PhX :size="14" /> 用户名已被占用
              </span>
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
                @input="updatePasswordStrength"
              />
              <div v-if="passwordStrength.level" class="password-strength">
                <div class="password-strength__bar">
                  <div
                    class="password-strength__fill"
                    :class="`password-strength__fill--${passwordStrength.level}`"
                    :style="{ width: passwordStrength.percent + '%' }"
                  ></div>
                </div>
                <span class="password-strength__label" :class="`password-strength__label--${passwordStrength.level}`">
                  {{ passwordStrength.text }}
                </span>
              </div>
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

            <button type="button" class="btn btn--primary auth-form__submit" @click="handleNext">
              下一步
            </button>
          </div>

          <!-- 步骤 2：邮箱 + 验证码 -->
          <div v-else key="step2" class="auth-form__step">
            <div class="auth-form__group">
              <label class="auth-form__label">邮箱</label>
              <input
                v-model="form.email"
                type="email"
                class="input"
                :class="{ 'input--error': error }"
                placeholder="your@email.com"
                autocomplete="email"
              />
            </div>

            <div class="auth-form__group">
              <label class="auth-form__label">验证码</label>
              <div class="auth-form__code-row">
                <input
                  v-model="form.code"
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

            <div v-if="success" class="auth-form__success">
              <PhCheckCircle :size="14" />
              {{ success }}
            </div>

            <div class="auth-form__actions">
              <button type="button" class="btn btn--ghost" @click="step = 1; error = ''">上一步</button>
              <button type="submit" class="btn btn--primary" :disabled="loading" @click="handleRegister">
                <PhCircleNotch v-if="loading" :size="18" class="spin" />
                {{ loading ? '注册中...' : '注册' }}
              </button>
            </div>
          </div>
        </Transition>
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
import { sendCode, checkUsername } from '../api'
import { PhSparkle, PhXCircle, PhCheckCircle, PhCircleNotch, PhCheck, PhX } from '@phosphor-icons/vue'

const auth = useAuthStore()
const router = useRouter()

const step = ref(1)
const form = reactive({ username: '', password: '', confirmPassword: '', email: '', code: '' })
const loading = ref(false)
const error = ref('')
const success = ref('')
const codeCooldown = ref(0)
const sendingCode = ref(false)
const usernameStatus = ref('') // '' | 'checking' | 'available' | 'taken'
const passwordStrength = reactive({ level: '', percent: 0, text: '' })

async function handleCheckUsername() {
  if (form.username.length < 3) {
    usernameStatus.value = ''
    return
  }
  usernameStatus.value = 'checking'
  try {
    await checkUsername(form.username)
    usernameStatus.value = 'available'
  } catch {
    usernameStatus.value = 'taken'
  }
}

function updatePasswordStrength() {
  const pwd = form.password
  if (!pwd) {
    passwordStrength.level = ''
    passwordStrength.percent = 0
    passwordStrength.text = ''
    return
  }
  let score = 0
  if (pwd.length >= 6) score++
  if (pwd.length >= 10) score++
  if (/[a-z]/.test(pwd) && /[A-Z]/.test(pwd)) score++
  if (/\d/.test(pwd)) score++
  if (/[^a-zA-Z0-9]/.test(pwd)) score++

  if (score <= 2) {
    passwordStrength.level = 'weak'
    passwordStrength.percent = 33
    passwordStrength.text = '弱'
  } else if (score <= 3) {
    passwordStrength.level = 'medium'
    passwordStrength.percent = 66
    passwordStrength.text = '中'
  } else {
    passwordStrength.level = 'strong'
    passwordStrength.percent = 100
    passwordStrength.text = '强'
  }
}

function handleNext() {
  error.value = ''
  if (!form.username || !form.password) {
    error.value = '请填写用户名和密码'
    return
  }
  if (form.username.length < 3) {
    error.value = '用户名至少 3 个字符'
    return
  }
  if (usernameStatus.value === 'taken') {
    error.value = '用户名已被占用，请更换'
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
  step.value = 2
}

async function handleSendCode() {
  error.value = ''
  if (!form.email) { error.value = '请输入邮箱'; return }
  sendingCode.value = true
  try {
    await sendCode({ email: form.email })
    codeCooldown.value = 60
    const timer = setInterval(() => {
      codeCooldown.value--
      if (codeCooldown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (e) { error.value = e.message }
  finally { sendingCode.value = false }
}

async function handleRegister() {
  error.value = ''
  success.value = ''

  if (!form.email || !form.code) {
    error.value = '请填写邮箱和验证码'
    return
  }

  loading.value = true
  try {
    await auth.register({ username: form.username, password: form.password, email: form.email, code: form.code })
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
  margin-bottom: 20px;
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

/* 步骤指示器 */
.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0;
  margin-bottom: 24px;
}

.step-indicator__dot {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-3);
  background: var(--bg-muted);
  transition: all 0.25s ease;
}

.step-indicator__dot--active {
  color: #fff;
  background: var(--accent);
}

.step-indicator__line {
  width: 48px;
  height: 2px;
  background: var(--bg-muted);
  margin: 0 8px;
  transition: background 0.25s ease;
}

.step-indicator__line--active {
  background: var(--accent);
}

/* 表单 */
.auth-form {
  display: flex;
  flex-direction: column;
}

.auth-form__step {
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

.auth-form__hint {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}

.auth-form__hint--ok {
  color: var(--mint);
}

.auth-form__hint--err {
  color: #c47878;
}

/* 密码强度 */
.password-strength {
  display: flex;
  align-items: center;
  gap: 8px;
}

.password-strength__bar {
  flex: 1;
  height: 3px;
  background: var(--bg-muted);
  border-radius: 2px;
  overflow: hidden;
}

.password-strength__fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.25s ease, background 0.25s ease;
}

.password-strength__fill--weak {
  background: #c47878;
}

.password-strength__fill--medium {
  background: #d4a843;
}

.password-strength__fill--strong {
  background: var(--mint);
}

.password-strength__label {
  font-size: 12px;
  font-weight: 500;
  min-width: 16px;
}

.password-strength__label--weak {
  color: #c47878;
}

.password-strength__label--medium {
  color: #d4a843;
}

.password-strength__label--strong {
  color: var(--mint);
}

.auth-form__code-row {
  display: flex;
  gap: 8px;
}

.auth-form__code-row .input {
  flex: 1;
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

.auth-form__actions {
  display: flex;
  gap: 12px;
  margin-top: 4px;
}

.auth-form__actions .btn {
  flex: 1;
  height: 42px;
}

.auth-card__footer {
  text-align: center;
  margin-top: 20px;
  font-size: 14px;
}

/* 步骤过渡动画 */
.step-enter-active,
.step-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.step-enter-from {
  opacity: 0;
  transform: translateX(16px);
}

.step-leave-to {
  opacity: 0;
  transform: translateX(-16px);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spin {
  animation: spin 0.8s linear infinite;
}
</style>