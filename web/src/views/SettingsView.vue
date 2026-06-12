<template>
  <div class="settings fade-up">
    <h1 class="font-display settings__title">设置</h1>

    <!-- Tab Bar -->
    <div class="settings__tabs">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="settings__tab"
        :class="{ 'settings__tab--active': activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        <component :is="tab.icon" :size="16" />
        {{ tab.label }}
      </button>
    </div>

    <template v-if="auth.user">

      <!-- ====== 资料 Tab ====== -->
      <div v-show="activeTab === 'profile'" class="card settings__panel">
        <div class="settings__rows">
          <div class="settings__row">
            <div class="settings__row-label">
              <span>昵称</span>
              <span class="settings__row-hint">1-50 字符，公开显示</span>
            </div>
            <div class="settings__row-control">
              <input v-model="profileForm.nickname" type="text" class="input settings__input" placeholder="输入昵称" />
            </div>
          </div>
          <div class="settings__row">
            <div class="settings__row-label">
              <span>用户名</span>
              <span class="settings__row-hint">不可更改</span>
            </div>
            <div class="settings__row-control">
              <span class="settings__static-value">@{{ auth.user.username }}</span>
            </div>
          </div>
        </div>

        <!-- Unsaved hint -->
        <Transition name="fade">
          <div v-if="profileDirty" class="settings__unsaved-bar">
            <div class="settings__unsaved-hint">
              <PhWarningCircle :size="16" />
              你有未保存的更改
            </div>
            <div class="settings__unsaved-actions">
              <button class="btn btn--outline btn--sm" @click="resetProfile">重置</button>
              <button class="btn btn--primary btn--sm" :disabled="profileSaving" @click="saveProfile">
                <PhCircleNotch v-if="profileSaving" :size="14" class="spin" />
                {{ profileSaving ? '保存中...' : '保存更改' }}
              </button>
            </div>
          </div>
        </Transition>

        <div v-if="profileError" class="settings__msg settings__msg--error">
          <PhXCircle :size="14" />
          {{ profileError }}
        </div>
        <div v-if="profileSuccess" class="settings__msg settings__msg--success">
          <PhCheckCircle :size="14" />
          {{ profileSuccess }}
        </div>
      </div>

      <!-- ====== 安全 Tab ====== -->
      <div v-show="activeTab === 'security'" class="card settings__panel">
        <div class="settings__rows">
          <div class="settings__row">
            <div class="settings__row-label">
              <span>密码</span>
              <span class="settings__row-hint">至少 8 位，须包含字母和数字</span>
            </div>
            <div class="settings__row-control settings__row-control--end">
              <span class="settings__static-value">••••••••</span>
              <button class="btn btn--outline btn--sm" @click="showPasswordModal = true">
                <PhKey :size="14" />
                修改
              </button>
            </div>
          </div>
          <div class="settings__row">
            <div class="settings__row-label">
              <span>邮箱</span>
              <span class="settings__row-hint">{{ auth.user.email ? '已绑定' : '未绑定' }}</span>
            </div>
            <div class="settings__row-control settings__row-control--end">
              <span class="settings__static-value">{{ auth.user.email || '未绑定' }}</span>
              <button class="btn btn--outline btn--sm" @click="openEmailModal">
                <PhEnvelopeSimple :size="14" />
                {{ auth.user.email ? '更换' : '绑定' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- ====== 外观 Tab ====== -->
      <div v-show="activeTab === 'appearance'" class="card settings__panel">
        <div class="settings__rows">
          <div class="settings__row">
            <div class="settings__row-label">
              <span>主题</span>
              <span class="settings__row-hint">选择界面外观</span>
            </div>
            <div class="settings__row-control settings__row-control--end">
              <div class="settings__theme-group">
                <button
                  v-for="opt in themeOptions"
                  :key="opt.value"
                  class="settings__theme-btn"
                  :class="{ 'settings__theme-btn--active': theme.mode === opt.value }"
                  @click="theme.setTheme(opt.value)"
                >
                  <component :is="opt.icon" :size="16" />
                  {{ opt.label }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- ====== 危险操作 ====== -->
      <div class="card settings__panel settings__panel--danger">
        <h2 class="font-display settings__section-title" style="color: #c47878">
          <PhWarning :size="16" />
          危险操作
        </h2>
        <div class="settings__rows">
          <div class="settings__row">
            <div class="settings__row-label">
              <span>注销账号</span>
              <span class="settings__row-hint">永久删除你的账号和数据，不可恢复</span>
            </div>
            <div class="settings__row-control settings__row-control--end">
              <button class="btn btn--danger btn--sm" disabled>
                <PhTrash :size="14" />
                注销账号
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- ====== Password Modal ====== -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showPasswordModal" class="modal-overlay" @click.self="closePasswordModal">
          <div class="modal-panel">
            <div class="modal-panel__header">
              <h3 class="font-display" style="font-size: 16px">修改密码</h3>
              <button class="modal-panel__close" @click="closePasswordModal">
                <PhX :size="18" />
              </button>
            </div>
            <div class="modal-panel__body">
              <div class="settings__field">
                <label class="settings__label">新密码</label>
                <input v-model="pwForm.password" type="password" class="input" placeholder="至少 8 位，须包含字母和数字" />
              </div>
              <div class="settings__field">
                <label class="settings__label">确认密码</label>
                <input v-model="pwForm.confirm" type="password" class="input" placeholder="再次输入密码" />
              </div>
              <div v-if="pwError" class="settings__msg settings__msg--error">
                <PhXCircle :size="14" />
                {{ pwError }}
              </div>
            </div>
            <div class="modal-panel__footer">
              <button class="btn btn--outline btn--sm" @click="closePasswordModal">取消</button>
              <button class="btn btn--primary btn--sm" :disabled="pwLoading" @click="handlePasswordChange">
                <PhCircleNotch v-if="pwLoading" :size="14" class="spin" />
                {{ pwLoading ? '保存中...' : '确认修改' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>

    <!-- ====== Email Modal ====== -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showEmailModal" class="modal-overlay" @click.self="closeEmailModal">
          <div class="modal-panel">
            <div class="modal-panel__header">
              <h3 class="font-display" style="font-size: 16px">{{ auth.user?.email ? '更换邮箱' : '绑定邮箱' }}</h3>
              <button class="modal-panel__close" @click="closeEmailModal">
                <PhX :size="18" />
              </button>
            </div>
            <div class="modal-panel__body">
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
              <div v-if="emailError" class="settings__msg settings__msg--error">
                <PhXCircle :size="14" />
                {{ emailError }}
              </div>
            </div>
            <div class="modal-panel__footer">
              <button class="btn btn--outline btn--sm" @click="closeEmailModal">取消</button>
              <button class="btn btn--primary btn--sm" :disabled="emailLoading" @click="handleBindEmail">
                <PhCircleNotch v-if="emailLoading" :size="14" class="spin" />
                {{ emailLoading ? '绑定中...' : '确认绑定' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useThemeStore } from '../stores/theme'
import { updateProfile, bindEmail, sendCode } from '../api'
import {
  PhUser, PhKey, PhEnvelopeSimple, PhPalette, PhWarning, PhTrash,
  PhXCircle, PhCheckCircle, PhCircleNotch, PhSun, PhMoon, PhDesktop,
  PhWarningCircle, PhX
} from '@phosphor-icons/vue'

const auth = useAuthStore()
const theme = useThemeStore()

// ---- Tabs ----
const activeTab = ref('profile')

const tabs = [
  { key: 'profile', label: '资料', icon: PhUser },
  { key: 'security', label: '安全', icon: PhKey },
  { key: 'appearance', label: '外观', icon: PhPalette },
]

// ---- Theme ----
const themeOptions = [
  { value: 'light', label: '亮色', icon: PhSun },
  { value: 'dark', label: '暗色', icon: PhMoon },
  { value: 'system', label: '跟随系统', icon: PhDesktop },
]

// ---- Profile (资料 Tab) ----
const profileForm = reactive({ nickname: '' })
const initialNickname = ref('')
const profileSaving = ref(false)
const profileError = ref('')
const profileSuccess = ref('')

// Initialize form when user data is available
watch(() => auth.user, (user) => {
  if (user) {
    profileForm.nickname = user.nickname || ''
    initialNickname.value = user.nickname || ''
  }
}, { immediate: true })

const profileDirty = computed(() => profileForm.nickname !== initialNickname.value)

function resetProfile() {
  profileForm.nickname = initialNickname.value
  profileError.value = ''
  profileSuccess.value = ''
}

async function saveProfile() {
  profileError.value = ''
  profileSuccess.value = ''
  if (!profileForm.nickname || profileForm.nickname.length > 50) {
    profileError.value = '昵称长度 1-50 字符'
    return
  }
  profileSaving.value = true
  try {
    await updateProfile({ nickname: profileForm.nickname })
    initialNickname.value = profileForm.nickname
    profileSuccess.value = '资料已保存'
    auth.fetchProfile()
    setTimeout(() => { profileSuccess.value = '' }, 3000)
  } catch (e) {
    profileError.value = e.message
  } finally {
    profileSaving.value = false
  }
}

// ---- Password Modal ----
const showPasswordModal = ref(false)
const pwForm = reactive({ password: '', confirm: '' })
const pwLoading = ref(false)
const pwError = ref('')

function closePasswordModal() {
  showPasswordModal.value = false
  pwForm.password = ''
  pwForm.confirm = ''
  pwError.value = ''
}

async function handlePasswordChange() {
  pwError.value = ''
  if (pwForm.password.length < 8 || !/[a-zA-Z]/.test(pwForm.password) || !/\d/.test(pwForm.password)) {
    pwError.value = '密码至少 8 位，须包含字母和数字'
    return
  }
  if (pwForm.password !== pwForm.confirm) {
    pwError.value = '两次密码不一致'
    return
  }
  pwLoading.value = true
  try {
    await updateProfile({ password: pwForm.password })
    closePasswordModal()
  } catch (e) {
    pwError.value = e.message
  } finally {
    pwLoading.value = false
  }
}

// ---- Email Modal ----
const showEmailModal = ref(false)
const emailForm = reactive({ email: '', code: '' })
const emailLoading = ref(false)
const emailError = ref('')
const codeCooldown = ref(0)

function openEmailModal() {
  emailForm.email = auth.user?.email || ''
  emailForm.code = ''
  emailError.value = ''
  showEmailModal.value = true
}

function closeEmailModal() {
  showEmailModal.value = false
  emailForm.email = ''
  emailForm.code = ''
  emailError.value = ''
}

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
  if (!emailForm.email || !emailForm.code) { emailError.value = '请填写邮箱和验证码'; return }
  emailLoading.value = true
  try {
    await bindEmail({ email: emailForm.email, code: emailForm.code })
    await auth.fetchProfile()
    closeEmailModal()
  } catch (e) { emailError.value = e.message }
  finally { emailLoading.value = false }
}
</script>

<style scoped>
.settings {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 640px;
}

.settings__title {
  font-size: 26px;
  margin-bottom: 4px;
}

/* ---- Tabs ---- */
.settings__tabs {
  display: flex;
  gap: 4px;
  border-bottom: 1px solid var(--border);
  padding-bottom: 0;
  margin-bottom: 4px;
}

.settings__tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border: none;
  background: none;
  color: var(--text-3);
  font-family: var(--font-body);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border-radius: var(--radius-m) var(--radius-m) 0 0;
  position: relative;
  overflow: hidden;
  transition: color var(--duration-medium-1) var(--ease-standard),
              background var(--duration-medium-1) var(--ease-standard);
}

.settings__tab::after {
  content: '';
  position: absolute;
  bottom: -1px;
  left: 0;
  right: 0;
  height: 2px;
  background: transparent;
  transition: background var(--duration-medium-1) var(--ease-standard);
  pointer-events: none;
}

.settings__tab:hover {
  color: var(--text-1);
  background: var(--state-hover);
}

.settings__tab--active {
  color: var(--accent);
  font-weight: 600;
}

.settings__tab--active::after {
  background: var(--accent);
}

/* ---- Panel ---- */
.settings__panel {
  padding: 0;
  overflow: hidden;
}

.settings__section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  font-weight: 600;
  padding: 20px 24px 0;
  margin-bottom: 0;
}

/* ---- Rows (table-style) ---- */
.settings__rows {
  display: flex;
  flex-direction: column;
}

.settings__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px 24px;
  border-bottom: 1px solid var(--border);
}

.settings__row:last-child {
  border-bottom: none;
}

.settings__row-label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  flex-shrink: 0;
}

.settings__row-label > span:first-child {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-1);
}

.settings__row-hint {
  font-size: 12px;
  color: var(--text-3);
}

.settings__row-control {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  justify-content: flex-start;
}

.settings__row-control--end {
  justify-content: flex-end;
}

.settings__input {
  max-width: 320px;
}

.settings__static-value {
  font-size: 14px;
  color: var(--text-2);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* ---- Unsaved bar ---- */
.settings__unsaved-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 24px;
  background: var(--accent-light);
  border-top: 1px solid var(--accent-glow);
}

.settings__unsaved-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--accent);
  font-weight: 500;
}

.settings__unsaved-actions {
  display: flex;
  gap: 8px;
}

/* ---- Messages ---- */
.settings__msg {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  padding: 10px 24px;
}

.settings__msg--error {
  color: #c47878;
}

.settings__msg--success {
  color: var(--mint);
}

/* ---- Modal field ---- */
.settings__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.settings__label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-3);
  letter-spacing: 0.3px;
  text-transform: uppercase;
}

.settings__code-row {
  display: flex;
  gap: 8px;
}

.settings__code-row .input {
  flex: 1;
}

/* ---- Theme buttons ---- */
.settings__theme-group {
  display: flex;
  gap: 6px;
}

.settings__theme-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 8px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-m);
  background: transparent;
  color: var(--text-2);
  font-family: var(--font-body);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--duration-medium-1) var(--ease-standard);
}

.settings__theme-btn:hover {
  border-color: var(--border-hover);
  color: var(--text-1);
}

.settings__theme-btn--active {
  border-color: var(--accent);
  color: var(--accent);
  background: var(--accent-light);
}

/* ---- Danger panel ---- */
.settings__panel--danger {
  border: 1px solid rgba(196, 120, 120, 0.2);
}

/* ---- Responsive ---- */
@media (max-width: 600px) {
  .settings__row {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .settings__row-control,
  .settings__row-control--end {
    width: 100%;
    justify-content: flex-start;
  }

  .settings__input {
    max-width: 100%;
  }

  .settings__unsaved-bar {
    flex-direction: column;
    gap: 10px;
    align-items: flex-start;
  }

  .settings__theme-group {
    flex-wrap: wrap;
  }
}
</style>
