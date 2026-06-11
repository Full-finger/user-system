<template>
  <div class="profile fade-up">
    <template v-if="auth.user">
      <!-- Profile Header -->
      <div class="card profile-header">
          <div class="profile-header__left">
            <div class="avatar avatar--lg">
              {{ (auth.user.nickname || auth.user.username)[0].toUpperCase() }}
            </div>
            <div class="profile-header__info">
              <h1 class="font-display profile-header__name">
                {{ auth.user.nickname || auth.user.username }}
              </h1>
              <div class="profile-header__meta">
                <span class="pill pill--accent">{{ roleLabel(auth.user.role) }}</span>
                <span class="text-3" style="font-size: 13px">
                  <PhClock :size="12" style="vertical-align: -1px" />
                  注册于 {{ formatDate(auth.user.created_at) }}
                </span>
              </div>
            </div>
          </div>
        <router-link to="/settings" class="btn btn--outline btn--sm">
          <PhGearSix :size="14" />
          设置
        </router-link>
      </div>

      <!-- Stats Cards -->
      <div class="profile__stats">
        <div class="card profile-stat" v-for="stat in stats" :key="stat.label">
          <div class="profile-stat__icon" :style="{ color: stat.color }">
            <component :is="stat.icon" :size="20" />
          </div>
          <div class="profile-stat__value font-display">{{ stat.value }}</div>
          <div class="profile-stat__label text-3">{{ stat.label }}</div>
        </div>
      </div>

      <!-- Info Section -->
      <div class="card profile-info">
        <h2 class="font-display profile-info__title">账户信息</h2>
        <div class="profile-info__row">
          <span class="text-3">用户名</span>
          <span>@{{ auth.user.username }}</span>
        </div>
        <div class="profile-info__row">
          <span class="text-3">昵称</span>
          <span>{{ auth.user.nickname || auth.user.username }}</span>
        </div>
        <div class="profile-info__row">
          <span class="text-3">角色</span>
          <span class="pill pill--accent">{{ roleLabel(auth.user.role) }}</span>
        </div>
        <div class="profile-info__row">
          <span class="text-3">邮箱</span>
          <span>{{ auth.user.email || '未绑定' }}</span>
        </div>
        <div class="profile-info__row">
          <span class="text-3">注册时间</span>
          <span>{{ formatDate(auth.user.created_at) }}</span>
        </div>
      </div>

      <!-- Placeholder sections -->
      <div class="card profile-placeholder">
        <div class="profile-placeholder__icon">
          <PhChartBar :size="28" weight="bold" />
        </div>
        <h3 class="font-display">贡献热力图</h3>
        <p class="text-3">功能开发中...</p>
      </div>
    </template>

    <template v-else>
      <div class="profile__loading">
        <div class="skeleton" style="height: 80px; margin-bottom: 16px"></div>
        <div class="skeleton" style="height: 60px; width: 60%; margin-bottom: 16px"></div>
        <div class="skeleton" style="height: 120px"></div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { roleLabel } from '../utils/role'
import { PhClock, PhGearSix, PhNote, PhChatCircle, PhHeart, PhChartBar } from '@phosphor-icons/vue'

const auth = useAuthStore()

const stats = [
  { label: '发帖数', value: '—', icon: PhNote, color: '#9b8ec4' },
  { label: '回复数', value: '—', icon: PhChatCircle, color: '#7ba4d4' },
  { label: '获赞数', value: '—', icon: PhHeart, color: '#c47a99' },
]

function formatDate(dateStr) {
  if (!dateStr) return '—'
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}
</script>

<style scoped>
.profile {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.profile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px;
}

.profile-header__left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.profile-header__name {
  font-size: 22px;
  margin-bottom: 4px;
}

.profile-header__meta {
  display: flex;
  align-items: center;
  gap: 10px;
}

.profile__stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.profile-stat {
  padding: 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.profile-stat__icon {
  margin-bottom: 4px;
}

.profile-stat__value {
  font-size: 24px;
  font-weight: 700;
}

.profile-stat__label {
  font-size: 12px;
}

.profile-info {
  padding: 24px;
}

.profile-info__title {
  font-size: 15px;
  margin-bottom: 16px;
}

.profile-info__row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid var(--border);
  font-size: 14px;
}

.profile-info__row:last-child {
  border-bottom: none;
}

.profile-placeholder {
  padding: 32px;
  text-align: center;
}

.profile-placeholder__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-l);
  background: var(--bg-muted);
  color: var(--text-4);
  margin-bottom: 12px;
}

.profile__loading {
  padding: 20px;
}

@media (max-width: 600px) {
  .profile-header {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .profile__stats {
    grid-template-columns: 1fr;
  }
}
</style>
