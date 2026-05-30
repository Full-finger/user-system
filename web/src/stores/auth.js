import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getProfile, login as apiLogin, loginByCode as apiLoginByCode, register as apiRegister } from '../api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function login(form) {
    const res = await apiLogin(form)
    token.value = res.data.token
    localStorage.setItem('token', res.data.token)
    await fetchProfile()
  }

  async function loginByCode(form) {
    const res = await apiLoginByCode(form)
    token.value = res.data.token
    localStorage.setItem('token', res.data.token)
    await fetchProfile()
  }

  async function register(form) {
    await apiRegister(form)
  }

  async function fetchProfile() {
    try {
      const res = await getProfile()
      user.value = res.data
    } catch {
      logout()
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('token')
  }

  // 初始化时拉取 profile
  if (token.value) {
    fetchProfile()
  }

  return { token, user, isLoggedIn, isAdmin, login, loginByCode, register, fetchProfile, logout }
})
