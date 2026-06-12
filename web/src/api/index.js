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
      const url = err.response?.config?.url || ''
      // 登录接口的 401 表示凭证错误，不走"token 过期"逻辑
      if (url !== '/login' && url !== '/code-login') {
        localStorage.removeItem('token')
        if (router.currentRoute.value.name !== 'Login') {
          router.push({ name: 'Login', query: { redirect: router.currentRoute.value.fullPath } })
        }
        return Promise.reject(new Error('登录已过期，请重新登录'))
      }
    }
    // 未登录时 403 等同于需要登录
    if (err.response?.status === 403 && !localStorage.getItem('token')) {
      if (router.currentRoute.value.name !== 'Login') {
        router.push({ name: 'Login', query: { redirect: router.currentRoute.value.fullPath } })
      }
      return Promise.reject(new Error('请先登录'))
    }
    const data = err.response?.data
    const message = data?.message || '请求失败，请稍后重试'
    if (!err.response) {
      console.warn('[API] 未收到响应:', err.message, err.config?.url)
    }
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
  return api.get('/admin/users', { params })
}

export function getUser(id) {
  return api.get(`/admin/users/${id}`)
}

export function updateUser(id, data) {
  return api.put(`/admin/users/${id}`, data)
}

export function deleteUser(id) {
  return api.delete(`/admin/users/${id}`)
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

export function getPost(code) {
  return api.get(`/posts/${code}`)
}

export function createPost(data) {
  return api.post('/posts', data)
}

export function deletePost(code) {
  return api.delete(`/posts/${code}`)
}

export function toggleLikePost(code) {
  return api.put(`/posts/${code}/like`)
}

export function listFeed(params) {
  return api.get('/feed', { params })
}

// ---- User profile (public, by username) ----
export function getUserProfile(username) {
  return api.get(`/users/${username}`)
}

export function listUserPosts(username, params) {
  return api.get(`/users/${username}/posts`, { params })
}

export function listUserLikes(username, params) {
  return api.get(`/users/${username}/likes`, { params })
}

// ---- Comments ----
export function listComments(postCode, params) {
  return api.get(`/posts/${postCode}/comments`, { params })
}

export function createComment(postCode, data) {
  return api.post(`/posts/${postCode}/comments`, data)
}

export function listReplies(commentId, params) {
  return api.get(`/comments/${commentId}/replies`, { params })
}

export function toggleCommentLike(commentId) {
  return api.put(`/comments/${commentId}/like`)
}

// ---- Follow ----
export function toggleFollow(username) {
  return api.put(`/users/${username}/follow`)
}

export function getFollowers(username, params) {
  return api.get(`/users/${username}/followers`, { params })
}

export function getFollowings(username, params) {
  return api.get(`/users/${username}/followings`, { params })
}

export default api
