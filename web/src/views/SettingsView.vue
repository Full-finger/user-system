<template>
  <div class="settings fade-up">
    <h1 class="font-display" style="font-size: 26px; margin-bottom: 20px">设置</h1>

    <template v-if="auth.user">
      <!-- 修改密码 -->
      <div class="card settings__section">
        <h2 class="font-display settings__section-title">
          <PhKey :size="16" />
          修改密码
        </h2>
        <form @submit.prevent="handlePasswordChange" class="settings__form">
          <div class="settings__field">
            <label class="settings__label">新密码</label>
            <input v-model="pwForm.password" type="password" class="input" placeholder="至少 6 位" />
          </div>
          <div class="settings__field">
            <label class="settings__label">确认密码</label>
            <input v-model="pwForm.confirm" type="password" class="input" placeholder="再次输入" />
          </div>
          <div v-if="pwError" class="settings__error">
            <PhXCircle :size="14" />
            {{ pwError }}
          </div>
          <div v-if="pwSuccess" class="settings__success">
            <PhCheckCircle :size="14" />
            {{ pwSuccess }}
          </div>
          <button type="submit" class="btn btn--primary" :disabled="pwLoading">
            <PhCircleNotch v-if="pwLoading" :size="16" class="spin" />
            {{ pwLoading ? '保存中...' : '保存密码' }}
          </button>
        </form>
      </div>

      <!-- 绑定邮箱 -->
      <div class="card settings__section">
        <h2 class="font-display settings__section-title">
          <PhEnvelopeSimple :size="16" />
          绑定邮箱
        </h2>
        <form @submit.prevent="handleBindEmail" class="settings__form">
          <div class="settings__field">
            <label class="settings__label">邮箱地址</label>
            <input v-model="emailForm.email" type="email" class="input" placeholder="your@email.com" />
          </div>
          <div class="settings__field">
            <label class="settings__label">验证码</label>
            <div class="settings__code-row">
              <input v-model="emailForm.code" type="text" class="input" placeholder="6 位验证码" />
              <button
                type="button"
                class="btn btn--outline btn--sm"
                :disabled="codeCooldown > 0"
                @click="handleSendCode"
              >
                {{ codeCooldown > 0 ? `${codeCooldown}s` : '发送验证码' }}
              </button>
            </div>
          </div>
          <div v-if="emailError" class="settings__error">
            <PhXCircle :size="14" />
            {{ emailError }}
          </div>
          <div v-if="emailSuccess" class="settings__success">
            <PhCheckCircle :size="14" />
            {{ emailSuccess }}
          </div>
          <button type="submit" class="btn btn--primary" :disabled="emailLoading">
            <PhCircleNotch v-if="emailLoading" :size="16" class="spin" />
            {{ emailLoading ? '绑定中...' : '绑定邮箱' }}
          </button>
        </form>
        <p class="text-4" style="font-size: 12px; margin-top: 12px">
          当前邮箱：{{ auth.user.email || '未绑定' }}
        </p>
      </div>

      <!-- 外观设置 -->
      <div class="card settings__section">
        <h2 class="font-display settings__section-title">
          <PhPalette :size="16" />
          外观
        </h2>
        <div class="settings__theme-row">
          <span style="font-size: 14px">暗色模式</span>
          <button
            class="settings__toggle"
            :class="{ 'settings__toggle--on': isDark }"
            @click="toggleTheme"
          >
            <span class="settings__toggle-knob"></span>
          </button>
        </div>
      </div>

      <!-- 危险区域 -->
      <div class="card settings__section settings__section--danger">
        <h2 class="font-display settings__section-title" style="color: #c47878">
          <PhWarning :size="16" />
          危险操作
        </h2>
        <p class="text-3" style="font-size: 13px; margin-bottom: 12px">以下操作不可撤销，请谨慎操作。</p>
        <button class="btn btn--danger btn--sm" disabled>
          <PhTrash :size="14" />
          注销账号（开发中）
        </button>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useAuthStore } from '../stores/auth'
import { updateProfile, bindEmail, sendCode } from '../api'
import {
  PhKey, PhEnvelopeSimple, PhPalette, PhWarning, PhTrash,
  PhXCircle, PhCheckCircle, PhCircleNotch
} from '@phosphor-icons/vue'

const auth = useAuthStore()

const isDark = ref(document.documentElement.getAttribute('data-theme') === 'dark')

// ---- Password ----
const pwForm = reactive({ password: '', confirm: '' })
const pwLoading = ref(false)
const pwError = ref('')
const pwSuccess = ref('')

async function handlePasswordChange() {
  pwError.value = ''
  pwSuccess.value = ''
  if (pwForm.password.length < 6) { pwError.value = '密码至少 6 位'; return }
  if (pwForm.password !== pwForm.confirm) { pwError.value = '两次密码不一致'; return }
  pwLoading.value = true
  try {
    await updateProfile({ password: pwForm.password })
    pwSuccess.value = '密码修改成功'
    pwForm.password = ''
    pwForm.confirm = ''
  } catch (e) { pwError.value = e.message }
  finally { pwLoading.value = false }
}

// ---- Email ----
const emailForm = reactive({ email: '', code: '' })
const emailLoading = ref(false)
const emailError = ref('')
const emailSuccess = ref('')
const codeCooldown = ref(0)

async function handleSendCode() {
  emailError.value = ''
  if (!emailForm.email) { emailError.value = '请输入邮箱'; return }
  try {
    await sendCode({ email: emailForm.email })
    codeCooldown.value = 60
    const timer = setInterval(() => {
      codeCooldown.value--
      if (codeCooldown.value <= 0) clearInterval(timer)
    }, 1000)
  } catch (e) { emailError.value = e.message }
}

async function handleBindEmail() {
  emailError.value = ''
  emailSuccess.value = ''
  if (!emailForm.email || !emailForm.code) { emailError.value = '请填写邮箱和验证码'; return }
  emailLoading.value = true
  try {
    await bindEmail({ email: emailForm.email, code: emailForm.code })
    emailSuccess.value = '邮箱绑定成功'
    emailForm.email = ''
    emailForm.code = ''
    auth.fetchProfile()
  } catch (e) { emailError.value = e.message }
  finally { emailLoading.value = false }
}

// ---- Theme ----
function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.setAttribute('data-theme', isDark.value ? 'dark' : 'light')
}
</script>

<style scoped>
.settings {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 560px;
}

.settings__section {
  padding: 24px;
}

.settings__section--danger {
  border: 1px solid rgba(196, 120, 120, 0.2);
}

.settings__section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 16px;
}

.settings__form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.settings__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.settings__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-2);
}

.settings__code-row {
  display: flex;
  gap: 8px;
}

.settings__code-row .input {
  flex: 1;
}

.settings__error {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #c47878;
}

.settings__success {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--mint);
}

.settings__theme-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.settings__toggle {
  width: 44px;
  height: 24px;
  border-radius: var(--radius-full);
  border: none;
  background: var(--text-4);
  cursor: pointer;
  position: relative;
  transition: background var(--duration-medium-1) var(--ease-standard);
}

.settings__toggle--on {
  background: var(--accent);
}

.settings__toggle-knob {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #fff;
  transition: transform var(--duration-medium-3) var(--ease-emphasized);
}

.settings__toggle--on .settings__toggle-knob {
  transform: translateX(20px);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spin {
  animation: spin 0.8s linear infinite;
}
</style>
