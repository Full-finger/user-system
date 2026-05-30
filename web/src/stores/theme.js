import { ref, computed, watch } from 'vue'
import { defineStore } from 'pinia'

const STORAGE_KEY = 'theme'

function getSystemDark() {
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? false
}

function applyTheme(dark) {
  document.documentElement.setAttribute('data-theme', dark ? 'dark' : 'light')
}

// 初始化时立即同步一次，防止闪烁
const saved = localStorage.getItem(STORAGE_KEY) || 'system'
if (saved === 'system') {
  applyTheme(getSystemDark())
} else {
  applyTheme(saved === 'dark')
}

export const useThemeStore = defineStore('theme', () => {
  const mode = ref(saved)
  const systemDark = ref(getSystemDark())

  // 监听系统主题变化
  const mql = window.matchMedia('(prefers-color-scheme: dark)')
  mql.addEventListener('change', (e) => {
    systemDark.value = e.matches
  })

  const isDark = computed(() => (mode.value === 'system' ? systemDark.value : mode.value === 'dark'))

  watch(isDark, (dark) => applyTheme(dark), { immediate: true })

  function setTheme(m) {
    mode.value = m
    localStorage.setItem(STORAGE_KEY, m)
  }

  function toggleTheme() {
    const order = ['light', 'dark', 'system']
    const next = order[(order.indexOf(mode.value) + 1) % 3]
    setTheme(next)
  }

  return { mode, isDark, setTheme, toggleTheme }
})