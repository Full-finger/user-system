<template>
  <div class="user-profile fade-up">
    <template v-if="profile">
      <!-- Header -->
      <div class="card profile-header fade-up">
        <div class="profile-header__left">
          <div class="avatar avatar--lg">{{ (profile.username || '?')[0].toUpperCase() }}</div>
          <div class="profile-header__info">
            <h1 class="font-display profile-header__name">{{ profile.username }}</h1>
            <div class="profile-header__meta">
              <span class="pill pill--accent">{{ profile.role === 'admin' ? '管理员' : '普通用户' }}</span>
              <span class="text-3" style="font-size: 13px">
                <PhClock :size="12" style="vertical-align: -1px" />
                注册于 {{ formatDate(profile.created_at) }}
              </span>
            </div>
          </div>
        </div>
        <button
          v-if="auth.isLoggedIn && auth.user?.id !== profile.id"
          class="btn btn--sm"
          :class="followed ? 'btn--primary' : 'btn--outline'"
          @click="handleFollow"
        >
          <PhUserPlus :size="14" :weight="followed ? 'fill' : 'regular'" />
          {{ followed ? '已关注' : '关注' }}
        </button>
      </div>

      <!-- Stats -->
      <div class="profile__stats fade-up" style="animation-delay: 40ms">
        <div class="card profile-stat">
          <div class="profile-stat__icon" style="color: #9b8ec4"><PhNote :size="20" /></div>
          <div class="profile-stat__value font-display">{{ profile.post_count }}</div>
          <div class="profile-stat__label text-3">发帖数</div>
        </div>
        <div class="card profile-stat">
          <div class="profile-stat__icon" style="color: #c47a99"><PhHeart :size="20" /></div>
          <div class="profile-stat__value font-display">{{ profile.follower_count }}</div>
          <div class="profile-stat__label text-3">粉丝</div>
        </div>
        <div class="card profile-stat">
          <div class="profile-stat__icon" style="color: #7ba4d4"><PhUserPlus :size="20" /></div>
          <div class="profile-stat__value font-display">{{ profile.following_count }}</div>
          <div class="profile-stat__label text-3">关注</div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="tab-bar fade-up" style="animation-delay: 80ms">
        <button class="tab-btn" :class="{ 'tab-btn--active': activeTab === 'posts' }" @click="activeTab = 'posts'">
          <PhNote :size="14" /> 帖子
        </button>
        <button class="tab-btn" :class="{ 'tab-btn--active': activeTab === 'likes' }" @click="activeTab = 'likes'">
          <PhHeart :size="14" /> 点赞
        </button>
        <button class="tab-btn" :class="{ 'tab-btn--active': activeTab === 'followers' }" @click="activeTab = 'followers'">
          <PhUsers :size="14" /> 粉丝
        </button>
        <button class="tab-btn" :class="{ 'tab-btn--active': activeTab === 'followings' }" @click="activeTab = 'followings'">
          <PhUserPlus :size="14" /> 关注
        </button>
      </div>

      <!-- Posts tab -->
      <div v-if="activeTab === 'posts'">
        <div v-if="userPosts.length > 0">
          <div v-for="(p, i) in userPosts" :key="p.id"
            class="post-card card fade-up"
            :style="{ animationDelay: (i * 40) + 'ms' }"
            @click="$router.push({ name: 'PostDetail', params: { id: p.id } })"
          >
            <div class="post-card__bar" :style="{ background: p.node?.color || 'var(--accent)' }"></div>
            <div class="post-card__vote">
              <span class="font-display" style="font-size: 14px; font-weight: 600">{{ p.like_count }}</span>
              <span style="font-size: 11px" class="text-4"><PhThumbsUp :size="12" /></span>
            </div>
            <div class="post-card__content">
              <div class="post-card__top">
                <h3 class="post-card__title">{{ p.title }}</h3>
                <span class="pill" :style="{ background: (p.node?.color || '#c47a99') + '18', color: p.node?.color || '#c47a99' }">{{ p.node?.name }}</span>
              </div>
              <div class="post-card__meta text-4">
                <span><PhClock :size="12" style="vertical-align: -1px" /> {{ formatTime(p.created_at) }}</span>
                <span><PhChatCircle :size="12" style="vertical-align: -1px" /> {{ p.reply_count }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="empty-state card fade-up" style="padding: 32px">
          <p class="text-3" style="font-size: 14px">暂无帖子</p>
        </div>
      </div>

      <!-- Likes tab -->
      <div v-if="activeTab === 'likes'">
        <div v-if="userLikes.length > 0">
          <div v-for="(item, i) in userLikes" :key="item.post.id"
            class="post-card card fade-up"
            :style="{ animationDelay: (i * 40) + 'ms' }"
            @click="$router.push({ name: 'PostDetail', params: { id: item.post.id } })"
          >
            <div class="post-card__bar" :style="{ background: item.post.node?.color || 'var(--accent)' }"></div>
            <div class="post-card__content">
              <div class="post-card__top">
                <h3 class="post-card__title">{{ item.post.title }}</h3>
                <span class="pill" :style="{ background: (item.post.node?.color || '#c47a99') + '18', color: item.post.node?.color || '#c47a99' }">{{ item.post.node?.name }}</span>
              </div>
              <div class="post-card__meta text-4">
                <span>{{ item.post.user?.username }}</span>
                <span><PhClock :size="12" style="vertical-align: -1px" /> {{ formatTime(item.post.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="empty-state card fade-up" style="padding: 32px">
          <p class="text-3" style="font-size: 14px">暂无点赞</p>
        </div>
      </div>

      <!-- Followers tab -->
      <div v-if="activeTab === 'followers'">
        <div v-if="followers.length > 0" class="user-list">
          <div v-for="f in followers" :key="f.id" class="card user-list__item fade-up" @click="$router.push({ name: 'UserProfile', params: { id: f.user.id } })">
            <div class="avatar avatar--sm">{{ (f.user.username || '?')[0].toUpperCase() }}</div>
            <span class="font-display" style="font-size: 14px; font-weight: 500">{{ f.user.username }}</span>
          </div>
        </div>
        <div v-else class="empty-state card fade-up" style="padding: 32px"><p class="text-3" style="font-size: 14px">暂无粉丝</p></div>
      </div>

      <!-- Followings tab -->
      <div v-if="activeTab === 'followings'">
        <div v-if="followings.length > 0" class="user-list">
          <div v-for="f in followings" :key="f.id" class="card user-list__item fade-up" @click="$router.push({ name: 'UserProfile', params: { id: f.user.id } })">
            <div class="avatar avatar--sm">{{ (f.user.username || '?')[0].toUpperCase() }}</div>
            <span class="font-display" style="font-size: 14px; font-weight: 500">{{ f.user.username }}</span>
          </div>
        </div>
        <div v-else class="empty-state card fade-up" style="padding: 32px"><p class="text-3" style="font-size: 14px">暂无关注</p></div>
      </div>
    </template>

    <!-- Loading -->
    <div v-else-if="loading">
      <div class="skeleton" style="height: 80px; margin-bottom: 16px"></div>
      <div class="skeleton" style="height: 60px; width: 60%; margin-bottom: 16px"></div>
    </div>

    <!-- Not found -->
    <div v-else class="empty-state card fade-up">
      <div class="empty-state__icon"><PhMagnifyingGlass :size="32" weight="bold" /></div>
      <h2 class="font-display" style="font-size: 18px; margin-bottom: 6px">用户不存在</h2>
      <router-link to="/" class="btn btn--outline btn--sm" style="margin-top: 12px">返回首页</router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { getUserProfile, listUserPosts, listUserLikes, getFollowers, getFollowings, toggleFollow } from '../api'
import {
  PhClock, PhNote, PhHeart, PhUserPlus, PhUsers, PhThumbsUp,
  PhChatCircle, PhMagnifyingGlass
} from '@phosphor-icons/vue'

const route = useRoute()
const auth = useAuthStore()

const profile = ref(null)
const loading = ref(true)
const followed = ref(false)
const activeTab = ref('posts')
const userPosts = ref([])
const userLikes = ref([])
const followers = ref([])
const followings = ref([])

async function fetchAll() {
  loading.value = true
  const userId = route.params.id
  try {
    const res = await getUserProfile(userId)
    profile.value = res.data
  } catch (e) {
    profile.value = null
    loading.value = false
    return
  }
  loading.value = false
  activeTab.value = 'posts'
  loadTab()
}

async function loadTab() {
  const userId = route.params.id
  try {
    if (activeTab.value === 'posts') {
      const res = await listUserPosts(userId, { page: 1, page_size: 20 })
      userPosts.value = res.data?.list || []
    } else if (activeTab.value === 'likes') {
      const res = await listUserLikes(userId, { page: 1, page_size: 20 })
      userLikes.value = res.data?.list || []
    } else if (activeTab.value === 'followers') {
      const res = await getFollowers(userId, { page: 1, page_size: 20 })
      followers.value = res.data?.list || []
    } else if (activeTab.value === 'followings') {
      const res = await getFollowings(userId, { page: 1, page_size: 20 })
      followings.value = res.data?.list || []
    }
  } catch (e) { console.error(e) }
}

async function handleFollow() {
  if (!auth.isLoggedIn || !profile.value) return
  try {
    const res = await toggleFollow(profile.value.id)
    followed.value = res.data?.followed
    if (followed.value) profile.value.follower_count++
    else profile.value.follower_count = Math.max(0, profile.value.follower_count - 1)
  } catch (e) { console.error(e) }
}

function formatDate(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr); const diff = Math.floor((Date.now() - d) / 1000)
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前'
  return d.toLocaleDateString('zh-CN')
}

import { watch as vueWatch } from 'vue'
vueWatch(activeTab, loadTab)

onMounted(fetchAll)
watch(() => route.params.id, fetchAll)
</script>

<style scoped>
.profile-header {
  display: flex; align-items: center; justify-content: space-between; padding: 24px;
}
.profile-header__left { display: flex; align-items: center; gap: 16px; }
.profile-header__name { font-size: 22px; margin-bottom: 4px; }
.profile-header__meta { display: flex; align-items: center; gap: 10px; }

.profile__stats {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 16px;
}
.profile-stat {
  padding: 20px; text-align: center;
  display: flex; flex-direction: column; align-items: center; gap: 4px;
}
.profile-stat__icon { margin-bottom: 4px; }
.profile-stat__value { font-size: 24px; font-weight: 700; }
.profile-stat__label { font-size: 12px; }

.user-list { display: flex; flex-direction: column; gap: 8px; }
.user-list__item {
  display: flex; align-items: center; gap: 10px; padding: 12px 16px;
  cursor: pointer; transition: box-shadow var(--duration-medium-2) var(--ease-standard);
}
.user-list__item:hover { box-shadow: var(--shadow-2); }

@media (max-width: 600px) {
  .profile-header { flex-direction: column; gap: 16px; align-items: flex-start; }
  .profile__stats { grid-template-columns: 1fr; }
}
</style>
