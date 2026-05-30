<template>
  <div class="layout">
    <!-- Topbar -->
    <header class="topbar glass">
      <div class="topbar__inner">
        <div class="topbar__left">
          <button class="topbar__menu-btn" @click="sidebarOpen = !sidebarOpen">
            <PhList :size="20" />
          </button>
          <router-link to="/" class="topbar__logo">
            <img v-if="siteConfig.siteLogo" :src="siteConfig.siteLogo" alt="Logo" class="topbar__logo-img" />
            <PhSparkle v-else :size="22" weight="fill" />
            <span class="font-display">{{ siteConfig.siteName }}</span>
          </router-link>
        </div>

        <div class="topbar__search">
          <PhMagnifyingGlass :size="16" class="topbar__search-icon" />
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="siteConfig.searchPlaceholder"
            class="topbar__search-input"
          />
        </div>

        <div class="topbar__right">
          <router-link v-if="auth.isLoggedIn" to="/notifications" class="topbar__icon-btn">
            <PhBell :size="20" />
            <span class="topbar__badge"></span>
          </router-link>

          <router-link v-if="auth.isLoggedIn" to="/messages" class="topbar__icon-btn">
            <PhEnvelopeSimple :size="20" />
          </router-link>

          <router-link to="/explore" class="topbar__icon-btn">
            <PhCompass :size="20" />
          </router-link>

          <template v-if="auth.isLoggedIn">
            <router-link to="/profile" class="topbar__avatar" :title="auth.user?.username">
              <div class="avatar avatar--sm">
                {{ (auth.user?.username || '?')[0].toUpperCase() }}
              </div>
            </router-link>
          </template>
          <template v-else>
            <router-link to="/login" class="btn btn--outline btn--sm">登录</router-link>
          </template>

          <button class="topbar__icon-btn" @click="theme.toggleTheme" :title="theme.mode === 'light' ? '暗色模式' : theme.mode === 'dark' ? '亮色模式' : '跟随系统'">
            <PhSun v-if="theme.mode === 'light'" :size="20" />
            <PhMoon v-else-if="theme.mode === 'dark'" :size="20" />
            <PhDesktop v-else :size="20" />
          </button>
        </div>
      </div>
    </header>

    <!-- Sidebar overlay (mobile) -->
    <div v-if="sidebarOpen" class="sidebar-overlay" @click="sidebarOpen = false"></div>

    <!-- Sidebar -->
    <aside class="sidebar" :class="{ 'sidebar--open': sidebarOpen }">
      <nav class="sidebar__nav">
        <router-link to="/" class="sidebar__item" @click="sidebarOpen = false">
          <PhHouse :size="20" />
          <span>发现</span>
        </router-link>
        <router-link to="/explore" class="sidebar__item" @click="sidebarOpen = false">
          <PhCompass :size="20" />
          <span>探索</span>
        </router-link>
        <div class="sidebar__divider"></div>
        <router-link to="/profile" class="sidebar__item" @click="sidebarOpen = false" v-if="auth.isLoggedIn">
          <PhUserCircle :size="20" />
          <span>个人中心</span>
        </router-link>
        <router-link to="/notifications" class="sidebar__item" @click="sidebarOpen = false" v-if="auth.isLoggedIn">
          <PhBell :size="20" />
          <span>通知</span>
        </router-link>
        <router-link to="/messages" class="sidebar__item" @click="sidebarOpen = false" v-if="auth.isLoggedIn">
          <PhEnvelopeSimple :size="20" />
          <span>私信</span>
        </router-link>
        <router-link to="/settings" class="sidebar__item" @click="sidebarOpen = false" v-if="auth.isLoggedIn">
          <PhGearSix :size="20" />
          <span>设置</span>
        </router-link>
        <router-link to="/admin" class="sidebar__item" @click="sidebarOpen = false" v-if="auth.isAdmin">
          <PhSquaresFour :size="20" />
          <span>管理后台</span>
        </router-link>
      </nav>

      <div class="sidebar__footer" v-if="auth.isLoggedIn">
        <button class="sidebar__item sidebar__item--danger" @click="handleLogout">
          <PhSignOut :size="20" />
          <span>退出登录</span>
        </button>
      </div>
    </aside>

    <!-- Main -->
    <main class="main-content">
      <div class="main-content__inner">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>

      <!-- Right Bar -->
      <aside class="rightbar">
        <div class="card rightbar__section">
          <h3 class="rightbar__title">
            <PhUsers :size="16" />
            在线用户
          </h3>
          <p class="text-3" style="font-size: 13px">功能开发中...</p>
        </div>

        <div class="card rightbar__section">
          <h3 class="rightbar__title">
            <PhChartBar :size="16" />
            今日统计
          </h3>
          <p class="text-3" style="font-size: 13px">功能开发中...</p>
        </div>

        <div class="card rightbar__section">
          <h3 class="rightbar__title">
            <PhMegaphone :size="16" />
            公告
          </h3>
          <p class="text-3" style="font-size: 13px">{{ siteConfig.announcement }}</p>
        </div>
      </aside>
    </main>

    <!-- Toast -->
    <Transition name="fade">
      <div v-if="toast.show" class="toast glass" :class="`toast--${toast.type}`">
        <PhCheckCircle v-if="toast.type === 'success'" :size="18" />
        <PhXCircle v-else-if="toast.type === 'error'" :size="18" />
        <PhInfo v-else :size="18" />
        <span>{{ toast.message }}</span>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useThemeStore } from '../stores/theme'
import { useRouter } from 'vue-router'
import { siteConfig } from '../config/site'

import {
  PhList, PhSparkle, PhMagnifyingGlass, PhBell, PhEnvelopeSimple,
  PhCompass, PhMoon, PhSun, PhDesktop, PhHouse, PhUserCircle, PhGearSix,
  PhSquaresFour, PhSignOut, PhUsers, PhChartBar, PhMegaphone,
  PhCheckCircle, PhXCircle, PhInfo
} from '@phosphor-icons/vue'

const auth = useAuthStore()
const theme = useThemeStore()
const router = useRouter()
const sidebarOpen = ref(false)
const searchQuery = ref('')

const toast = reactive({ show: false, message: '', type: 'info' })

function handleLogout() {
  auth.logout()
  sidebarOpen.value = false
  router.push('/login')
}

// expose toast for child views
function showToast(message, type = 'success') {
  toast.message = message
  toast.type = type
  toast.show = true
  setTimeout(() => { toast.show = false }, 3000)
}

</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* ---- Topbar ---- */
.topbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: var(--topbar-h);
  background: rgba(248, 245, 251, 0.82);
  z-index: 100;
  box-shadow: var(--shadow-3);
}

[data-theme="dark"] .topbar {
  background: rgba(18, 16, 22, 0.78);
}

.topbar__inner {
  height: 100%;
  display: flex;
  align-items: center;
  padding: 0 20px 0 calc(var(--sidebar-w) + 20px);
  gap: 16px;
}

.topbar__left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.topbar__menu-btn {
  display: none;
  background: none;
  border: none;
  color: var(--text-2);
  cursor: pointer;
  padding: 6px;
  border-radius: var(--radius-s);
  position: relative;
  overflow: hidden;
}

.topbar__menu-btn::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: transparent;
  transition: background var(--duration-medium-1) var(--ease-standard);
  pointer-events: none;
}

.topbar__menu-btn:hover::after {
  background: var(--state-hover);
}

.topbar__logo {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--accent);
  font-weight: 700;
  font-size: 18px;
  text-decoration: none;
  white-space: nowrap;
}

.topbar__logo-img {
  width: 22px;
  height: 22px;
  object-fit: contain;
  border-radius: var(--radius-s);
}

.topbar__search {
  flex: 1;
  max-width: 420px;
  position: relative;
}

.topbar__search-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-4);
}

.topbar__search-input {
  width: 100%;
  height: 36px;
  padding: 0 12px 0 36px;
  border: 1px solid var(--border);
  border-radius: var(--radius-full);
  background: var(--bg-muted);
  font-family: var(--font-body);
  font-size: 13px;
  color: var(--text-2);
  outline: none;
  transition: border-color var(--duration-medium-1) var(--ease-standard),
              box-shadow var(--duration-medium-1) var(--ease-standard),
              background var(--duration-medium-1) var(--ease-standard);
}

.topbar__search-input:focus {
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-glow);
  background: var(--bg-card);
}

.topbar__search-input::placeholder {
  color: var(--text-4);
}

.topbar__right {
  display: flex;
  align-items: center;
  gap: 4px;
}

.topbar__icon-btn {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-s);
  color: var(--text-2);
  text-decoration: none;
  overflow: hidden;
  transition: color var(--duration-medium-1) var(--ease-standard);
}

.topbar__icon-btn::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: transparent;
  transition: background var(--duration-medium-1) var(--ease-standard);
  pointer-events: none;
}

.topbar__icon-btn:hover {
  color: var(--accent);
}

.topbar__icon-btn:hover::after {
  background: var(--state-hover);
}

.topbar__badge {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 8px;
  height: 8px;
  background: var(--accent);
  border-radius: 50%;
  border: 2px solid var(--bg-card);
}

.topbar__avatar {
  display: flex;
  text-decoration: none;
  margin-left: 4px;
}

/* ---- Sidebar ---- */
.sidebar-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  z-index: 90;
}

.sidebar {
  position: fixed;
  top: var(--topbar-h);
  left: 0;
  width: var(--sidebar-w);
  height: calc(100vh - var(--topbar-h));
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 12px 0;
  z-index: 80;
  overflow-y: auto;
}

.sidebar__nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 0 8px;
}

.sidebar__item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-m);
  color: var(--text-2);
  text-decoration: none;
  font-size: 14px;
  font-family: var(--font-body);
  position: relative;
  overflow: hidden;
  transition: color var(--duration-medium-1) var(--ease-standard);
  cursor: pointer;
  background: none;
  border: none;
  width: 100%;
  text-align: left;
}

.sidebar__item::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: transparent;
  transition: background var(--duration-medium-1) var(--ease-standard);
  pointer-events: none;
}

.sidebar__item:hover {
  color: var(--text-1);
}

.sidebar__item:hover::after {
  background: var(--state-hover);
}

.sidebar__item.router-link-exact-active {
  color: var(--accent);
  font-weight: 600;
}

.sidebar__item.router-link-exact-active::after {
  background: var(--state-focus);
}

.sidebar__item--danger:hover {
  color: #c47878;
}

.sidebar__item--danger:hover::after {
  background: rgba(196, 120, 120, 0.08);
}

.sidebar__divider {
  height: 1px;
  background: var(--border);
  margin: 8px 12px;
}

.sidebar__footer {
  padding: 0 8px;
}

/* ---- Main Content ---- */
.main-content {
  margin-top: var(--topbar-h);
  margin-left: var(--sidebar-w);
  display: flex;
  min-height: calc(100vh - var(--topbar-h));
}

.main-content__inner {
  flex: 1;
  padding: 24px;
  max-width: 780px;
}

.rightbar {
  width: var(--rightbar-w);
  padding: 24px 24px 24px 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
  position: sticky;
  top: var(--topbar-h);
  height: fit-content;
  max-height: calc(100vh - var(--topbar-h));
  overflow-y: auto;
}

.rightbar__title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: var(--font-display);
  font-size: 14px;
  font-weight: 600;
  color: var(--text-1);
  margin-bottom: 8px;
}

.rightbar__section {
  padding: 16px;
}

/* ---- Responsive ---- */
@media (max-width: 1200px) {
  .rightbar {
    display: none;
  }
  .main-content__inner {
    max-width: 100%;
  }
}

@media (max-width: 960px) {
  .topbar__inner {
    padding-left: 20px;
  }

  .topbar__menu-btn {
    display: flex;
  }

  .sidebar {
    transform: translateX(-100%);
    transition: transform var(--duration-medium-4) var(--ease-emphasized-decelerate);
    z-index: 95;
    background: var(--bg-card);
  }

  .sidebar--open {
    transform: translateX(0);
  }

  .sidebar-overlay {
    display: block;
  }

  .main-content {
    margin-left: 0;
  }

  .topbar__search {
    max-width: 260px;
  }
}

@media (max-width: 600px) {
  .topbar__search {
    display: none;
  }

  .topbar__logo span {
    display: none;
  }

  .main-content__inner {
    padding: 16px;
  }
}
</style>
