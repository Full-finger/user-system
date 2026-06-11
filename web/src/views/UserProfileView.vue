<template>
  <div class="user-profile fade-up">
    <template v-if="profile">
      <!-- Header -->
      <div class="card profile-header fade-up">
        <div class="profile-header__left">
          <div class="avatar avatar--lg">{{ (profile.nickname || profile.username || '?')[0].toUpperCase() }}</div>
          <div class="profile-header__info">
            <h1 class="font-display profile-header__name">{{ profile.nickname || profile.username }}</h1>
            <div class="profile-header__meta">
              <span class="pill pill--accent">{{ roleLabel(profile.role) }}</span>
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
        <button class="tab-btn" :class="{ 'tab-btn--active': activeTab === 'posts' }" @click="switchTab('posts')">
          <PhNote :size="14" /> 帖子
        </button>
        <button class="tab-btn" :class="{ 'tab-btn--active': activeTab === 'likes' }" @click="switchTab('likes')">
          <PhHeart :size="14" /> 点赞
        </button>
        <button class="tab-btn" :class="{ 'tab-btn--active': activeTab === 'followers' }" @click="switchTab('followers')">
          <PhUsers :size="14" /> 粉丝
        </button>
        <button class="tab-btn" :class="{ 'tab-btn--active': activeTab === 'followings' }" @click="switchTab('followings')">
          <PhUserPlus :size="14" /> 关注
        </button>
      </div>

      <!-- Posts tab -->
      <div v-if="activeTab === 'posts'">
        <div v-if="userPosts.length > 0">
          <div v-for="(p, i) in userPosts" :key="p.code"
            class="post-card card fade-up"
            :style="{ animationDelay: (i * 40) + 'ms' }"
            @click="$router.push({ name: 'PostDetail', params: { code: p.code } })"
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
        <div v-else-if="!tabLoading.posts" class="empty-state card fade-up" style="padding: 32px">
          <p class="text-3" style="font-size: 14px">暂无帖子</p>
        </div>
        <div v-if="userPosts.length > 0" style="padding: 12px 0; text-align: center">
          <button v-if="tabHasMore.posts" class="btn btn--outline btn--sm" @click="loadTabMore('posts')" :disabled="tabLoadingMore.posts">
            <PhCircleNotch v-if="tabLoadingMore.posts" :size="14" class="spin" />
            <span v-else>加载更多</span>
          </button>
          <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
        </div>
      </div>

      <!-- Likes tab -->
      <div v-if="activeTab === 'likes'">
        <div v-if="userLikes.length > 0">
          <div v-for="(item, i) in userLikes" :key="item.post.code"
            class="post-card card fade-up"
            :style="{ animationDelay: (i * 40) + 'ms' }"
            @click="$router.push({ name: 'PostDetail', params: { code: item.post.code } })"
          >
            <div class="post-card__bar" :style="{ background: item.post.node?.color || 'var(--accent)' }"></div>
            <div class="post-card__content">
              <div class="post-card__top">
                <h3 class="post-card__title">{{ item.post.title }}</h3>
                <span class="pill" :style="{ background: (item.post.node?.color || '#c47a99') + '18', color: item.post.node?.color || '#c47a99' }">{{ item.post.node?.name }}</span>
              </div>
              <div class="post-card__meta text-4">
                <span>{{ item.post.user?.nickname || item.post.user?.username }}</span>
                <span><PhClock :size="12" style="vertical-align: -1px" /> {{ formatTime(item.post.created_at) }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else-if="!tabLoading.likes" class="empty-state card fade-up" style="padding: 32px">
          <p class="text-3" style="font-size: 14px">暂无点赞</p>
        </div>
        <div v-if="userLikes.length > 0" style="padding: 12px 0; text-align: center">
          <button v-if="tabHasMore.likes" class="btn btn--outline btn--sm" @click="loadTabMore('likes')" :disabled="tabLoadingMore.likes">
            <PhCircleNotch v-if="tabLoadingMore.likes" :size="14" class="spin" />
            <span v-else>加载更多</span>
          </button>
          <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
        </div>
      </div>

      <!-- Followers tab -->
      <div v-if="activeTab === 'followers'">
        <div v-if="followers.length > 0" class="user-list">
          <div v-for="f in followers" :key="f.id" class="card user-list__item fade-up" @click="$router.push({ name: 'UserProfile', params: { username: f.user.username } })">
            <div class="avatar avatar--sm">{{ (f.user.nickname || f.user.username || '?')[0].toUpperCase() }}</div>
            <span class="font-display" style="font-size: 14px; font-weight: 500">{{ f.user.nickname || f.user.username }}</span>
          </div>
        </div>
        <div v-else-if="!tabLoading.followers" class="empty-state card fade-up" style="padding: 32px"><p class="text-3" style="font-size: 14px">暂无粉丝</p></div>
        <div v-if="followers.length > 0" style="padding: 12px 0; text-align: center">
          <button v-if="tabHasMore.followers" class="btn btn--outline btn--sm" @click="loadTabMore('followers')" :disabled="tabLoadingMore.followers">
            <PhCircleNotch v-if="tabLoadingMore.followers" :size="14" class="spin" />
            <span v-else>加载更多</span>
          </button>
          <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
        </div>
      </div>

      <!-- Followings tab -->
      <div v-if="activeTab === 'followings'">
        <div v-if="followings.length > 0" class="user-list">
          <div v-for="f in followings" :key="f.id" class="card user-list__item fade-up" @click="$router.push({ name: 'UserProfile', params: { username: f.user.username } })">
            <div class="avatar avatar--sm">{{ (f.user.nickname || f.user.username || '?')[0].toUpperCase() }}</div>
            <span class="font-display" style="font-size: 14px; font-weight: 500">{{ f.user.nickname || f.user.username }}</span>
          </div>
        </div>
        <div v-else-if="!tabLoading.followings" class="empty-state card fade-up" style="padding: 32px"><p class="text-3" style="font-size: 14px">暂无关注</p></div>
        <div v-if="followings.length > 0" style="padding: 12px 0; text-align: center">
          <button v-if="tabHasMore.followings" class="btn btn--outline btn--sm" @click="loadTabMore('followings')" :disabled="tabLoadingMore.followings">
            <PhCircleNotch v-if="tabLoadingMore.followings" :size="14" class="spin" />
            <span v-else>加载更多</span>
          </button>
          <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
        </div>
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
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'
import { roleLabel } from '../utils/role'
import { getUserProfile, listUserPosts, listUserLikes, getFollowers, getFollowings, toggleFollow } from '../api'
import { formatTime } from '../utils/format'
import {
  PhClock, PhNote, PhHeart, PhUserPlus, PhUsers, PhThumbsUp,
  PhChatCircle, PhMagnifyingGlass, PhCircleNotch
} from '@phosphor-icons/vue'

const route = useRoute()
const auth = useAuthStore()
const toast = useToast()

const profile = ref(null)
const loading = ref(true)
const followed = ref(false)
const activeTab = ref('posts')

// Tab data
const userPosts = ref([])
const userLikes = ref([])
const followers = ref([])
const followings = ref([])

// Tab pagination state
const tabPage = reactive({ posts: 1, likes: 1, followers: 1, followings: 1 })
const tabHasMore = reactive({ posts: true, likes: true, followers: true, followings: true })
const tabLoading = reactive({ posts: false, likes: false, followers: false, followings: false })
const tabLoadingMore = reactive({ posts: false, likes: false, followers: false, followings: false })
const tabTotal = reactive({ posts: 0, likes: 0, followers: 0, followings: 0 })

const PAGE_SIZE = 20

async function fetchAll() {
  loading.value = true
  const username = route.params.username
  try {
    const res = await getUserProfile(username)
    profile.value = res.data
    followed.value = res.data?.followed || false
  } catch (e) {
    profile.value = null
    loading.value = false
    return
  }
  loading.value = false
  switchTab('posts')
}

async function loadTab(tab, reset = true) {
  const username = route.params.username
  if (reset) {
    tabPage[tab] = 1
    tabHasMore[tab] = true
    tabLoading[tab] = true
  } else {
    tabLoadingMore[tab] = true
  }

  try {
    const params = { page: tabPage[tab], page_size: PAGE_SIZE }
    if (tab === 'posts') {
      const res = await listUserPosts(username, params)
      const list = res.data?.list || []
      if (reset) userPosts.value = list; else userPosts.value.push(...list)
      tabTotal.posts = res.data?.total || 0
      tabHasMore.posts = userPosts.value.length < tabTotal.posts
    } else if (tab === 'likes') {
      const res = await listUserLikes(username, params)
      const list = res.data?.list || []
      if (reset) userLikes.value = list; else userLikes.value.push(...list)
      tabTotal.likes = res.data?.total || 0
      tabHasMore.likes = userLikes.value.length < tabTotal.likes
    } else if (tab === 'followers') {
      const res = await getFollowers(username, params)
      const list = res.data?.list || []
      if (reset) followers.value = list; else followers.value.push(...list)
      tabTotal.followers = res.data?.total || 0
      tabHasMore.followers = followers.value.length < tabTotal.followers
    } else if (tab === 'followings') {
      const res = await getFollowings(username, params)
      const list = res.data?.list || []
      if (reset) followings.value = list; else followings.value.push(...list)
      tabTotal.followings = res.data?.total || 0
      tabHasMore.followings = followings.value.length < tabTotal.followings
    }
  } catch (e) { toast.error(e.message) }
  finally {
    tabLoading[tab] = false
    tabLoadingMore[tab] = false
  }
}

function switchTab(tab) {
  activeTab.value = tab
  loadTab(tab, true)
}

function loadTabMore(tab) {
  tabPage[tab]++
  loadTab(tab, false)
}

async function handleFollow() {
  if (!auth.isLoggedIn || !profile.value) return
  try {
    const res = await toggleFollow(profile.value.username)
    followed.value = res.data?.followed
    if (followed.value) { profile.value.follower_count++; toast.success('已关注') }
    else { profile.value.follower_count = Math.max(0, profile.value.follower_count - 1); toast.success('已取消关注') }
  } catch (e) { toast.error(e.message) }
}

function formatDate(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}


onMounted(fetchAll)
watch(() => route.params.username, fetchAll)
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