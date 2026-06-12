import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getProfile, login as apiLogin, loginByCode as apiLoginByCode, register as apiRegister } from '../api'
import { ADMIN_ROLES, MANAGE_ROLES } from '../utils/role'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => ADMIN_ROLES.includes(user.value?.role))
  const canManagePosts = computed(() => MANAGE_ROLES.includes(user.value?.role))

  async function setTokenAndFetch(res) {
    token.value = res.data.token
    localStorage.setItem('token', res.data.token)
    await fetchProfile()
  }

  async function login(form) {
    const res = await apiLogin(form)
    await setTokenAndFetch(res)
  }

  async function loginByCode(form) {
    const res = await apiLoginByCode(form)
    await setTokenAndFetch(res)
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
  let initPromise = Promise.resolve()
  if (token.value) {
    initPromise = fetchProfile()
  }

  return { token, user, isLoggedIn, isAdmin, canManagePosts, login, loginByCode, register, fetchProfile, logout, initPromise }
})
