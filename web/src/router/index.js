import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/LoginView.vue'),
    meta: { guest: true }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/RegisterView.vue'),
    meta: { guest: true }
  },
  {
    path: '/',
    component: () => import('../views/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'Home',
        component: () => import('../views/HomeView.vue')
      },
      {
        path: 'explore',
        name: 'Explore',
        component: () => import('../views/ExploreView.vue')
      },
      {
        path: 'nodes/:id',
        name: 'NodePosts',
        component: () => import('../views/NodePostsView.vue')
      },
      {
        path: 'posts/:code',
        name: 'PostDetail',
        component: () => import('../views/PostView.vue')
      },
      {
        path: 'users/:username',
        name: 'UserProfile',
        component: () => import('../views/UserProfileView.vue')
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('../views/ProfileView.vue'),
        meta: { auth: true }
      },
      {
        // TODO: 预留功能 — 后端 API 尚未实现，前端仅占位展示
        path: 'notifications',
        name: 'Notifications',
        component: () => import('../views/NotificationsView.vue'),
        meta: { auth: true }
      },
      {
        // TODO: 预留功能 — 后端 API 尚未实现，前端仅占位展示
        path: 'messages',
        name: 'Messages',
        component: () => import('../views/MessagesView.vue'),
        meta: { auth: true }
      },
      {
        path: 'admin',
        name: 'Admin',
        component: () => import('../views/AdminView.vue'),
        meta: { auth: true, admin: true }
      },
      {
        path: 'create-post',
        name: 'CreatePost',
        component: () => import('../views/CreatePostView.vue'),
        meta: { auth: true }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../views/SettingsView.vue'),
        meta: { auth: true }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const auth = useAuthStore()

  // 等待 auth 初始化完成（首次加载时 fetchProfile 可能尚未完成）
  if (auth.isLoggedIn && !auth.user) {
    await auth.initPromise
  }

  if (to.meta.auth && !auth.isLoggedIn) {
    return next({ name: 'Login', query: { redirect: to.fullPath } })
  }
  if (to.meta.admin && !auth.canManagePosts) {
    console.warn('权限不足，无法访问管理页面')
    return next({ name: 'Home' })
  }
  if (to.meta.guest && auth.isLoggedIn) {
    return next({ name: 'Home' })
  }
  next()
})

export default router
