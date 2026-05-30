import axios from 'axios'
import router from '../router'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 — 自动带 token
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器 — 统一处理
api.interceptors.response.use(
  (res) => res.data,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      if (router.currentRoute.value.name !== 'Login') {
        router.push({ name: 'Login', query: { redirect: router.currentRoute.value.fullPath } })
      }
      return Promise.reject(new Error('登录已过期，请重新登录'))
    }
    const data = err.response?.data
    const message = data?.message || '网络错误，请稍后重试'
    return Promise.reject(new Error(message))
  }
)

// ---- Auth ----
export function checkUsername(username) {
  return api.get('/check-username', { params: { username } })
}

export function register(data) {
  return api.post('/register', data)
}

export function login(data) {
  return api.post('/login', data)
}

export function sendCode(data) {
  return api.post('/send-code', data)
}

export function loginByCode(data) {
  return api.post('/code-login', data)
}

// ---- Profile ----
export function getProfile() {
  return api.get('/profile')
}

export function updateProfile(data) {
  return api.put('/profile', data)
}

export function bindEmail(data) {
  return api.put('/profile/email', data)
}

// ---- Admin ----
export function listUsers(params) {
  return api.get('/users', { params })
}

export function getUser(id) {
  return api.get(`/users/${id}`)
}

export function updateUser(id, data) {
  return api.put(`/users/${id}`, data)
}

export function deleteUser(id) {
  return api.delete(`/users/${id}`)
}

// ---- Nodes ----
export function listNodes() {
  return api.get('/nodes')
}

export function getNode(id) {
  return api.get(`/nodes/${id}`)
}

export function getNodePosts(id, params) {
  return api.get(`/nodes/${id}/posts`, { params })
}

// ---- Posts ----
export function listPosts(params) {
  return api.get('/posts', { params })
}

export function getPost(id) {
  return api.get(`/posts/${id}`)
}

export function createPost(data) {
  return api.post('/posts', data)
}

export function deletePost(id) {
  return api.delete(`/posts/${id}`)
}

export function toggleLikePost(id) {
  return api.put(`/posts/${id}/like`)
}

export function listFeed(params) {
  return api.get('/feed', { params })
}

// ---- User profile (public) ----
export function getUserProfile(userId) {
  return api.get(`/users/${userId}`)
}

export function listUserPosts(userId, params) {
  return api.get(`/users/${userId}/posts`, { params })
}

export function listUserLikes(userId, params) {
  return api.get(`/users/${userId}/likes`, { params })
}

// ---- Follow ----
export function toggleFollow(userId) {
  return api.put(`/users/${userId}/follow`)
}

export function getFollowers(userId, params) {
  return api.get(`/users/${userId}/followers`, { params })
}

export function getFollowings(userId, params) {
  return api.get(`/users/${userId}/followings`, { params })
}

export default api
