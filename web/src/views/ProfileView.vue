<template>
  <div class="profile fade-up">
    <template v-if="auth.user">
      <!-- Profile Header（含座右铭 + 紧凑统计） -->
      <div class="card profile-header" :style="coverStyle">
        <div v-if="auth.user.cover_theme" class="profile-header__cover"></div>
        <div class="profile-header__body">
          <div class="profile-header__left">
            <UserAvatar :avatar-url="auth.user.avatar_url" :name="auth.user.nickname || auth.user.username" size="lg" />
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
              <p v-if="auth.user.motto" class="profile-header__motto text-3">{{ auth.user.motto }}</p>
            </div>
          </div>
          <div class="profile-header__right">
            <div class="profile-header__stats">
              <div v-for="stat in stats" :key="stat.label" class="profile-header__stat">
                <div class="profile-header__stat-value font-display">{{ stat.value }}</div>
                <div class="profile-header__stat-label text-3">{{ stat.label }}</div>
              </div>
            </div>
            <div class="profile-header__actions">
              <router-link :to="`/users/${auth.user.username}`" class="btn btn--outline btn--sm">
                <PhEye :size="14" />
                查看公开主页
              </router-link>
              <router-link to="/settings" class="btn btn--outline btn--sm">
                <PhGearSix :size="14" />
                设置
              </router-link>
            </div>
          </div>
        </div>
      </div>

      <!-- Tab 导航 -->
      <div class="tab-bar profile-tabs">
        <button
          v-for="tab in tabs" :key="tab.key"
          class="tab-btn" :class="{ 'tab-btn--active': activeTab === tab.key }"
          @click="switchTab(tab.key)"
        >
          {{ tab.label }}
          <span v-if="tab.count != null && tab.count !== ''" class="tab-badge">{{ tab.count }}</span>
        </button>
      </div>

      <!-- Tab 内容 -->
      <!-- 账户信息 -->
      <div v-show="activeTab === 'account'" class="card profile-info">
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
          <span class="text-3">座右铭</span>
          <span>{{ auth.user.motto || '未设置' }}</span>
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

      <!-- 我的点赞 -->
      <div v-show="activeTab === 'likes'">
        <div v-if="myLikes.length > 0">
          <div v-for="item in myLikes" :key="item.post.code"
            class="post-card card fade-up"
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
        <div v-else-if="!likesLoading" class="empty-state card">
          <p class="text-3" style="font-size: 14px">还没有点赞过帖子</p>
        </div>
        <div v-if="myLikes.length > 0" style="padding: 12px 0; text-align: center">
          <button v-if="likesHasMore" class="btn btn--outline btn--sm" @click="loadMoreLikes" :disabled="likesLoadingMore">
            <PhCircleNotch v-if="likesLoadingMore" :size="14" class="spin" />
            <span v-else>加载更多</span>
          </button>
          <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
        </div>
      </div>

      <!-- 我的关注 -->
      <div v-show="activeTab === 'followings'">
        <div v-if="followings.length > 0">
          <div v-for="item in followings" :key="item.id"
            class="user-card card fade-up"
            @click="$router.push({ name: 'UserProfile', params: { username: item.user.username } })"
          >
            <UserAvatar :avatar-url="item.user.avatar_url" :name="item.user.nickname || item.user.username" size="md" />
            <div class="user-card__info">
              <div class="user-card__name">{{ item.user.nickname || item.user.username }}</div>
              <div class="user-card__sub text-4">@{{ item.user.username }}</div>
            </div>
            <button
              v-if="item.user.username !== auth.user?.username"
              class="btn btn--sm"
              :class="item.followed ? 'btn--primary' : 'btn--outline'"
              @click.stop="handleFollowUser(item)"
            >
              <PhUserMinus v-if="item.followed" :size="14" />
              <PhUserPlus v-else :size="14" />
              {{ item.followed ? '已关注' : '关注' }}
            </button>
          </div>
        </div>
        <div v-else-if="!followingsLoading" class="empty-state card">
          <p class="text-3" style="font-size: 14px">还没有关注任何人</p>
        </div>
        <div v-if="followings.length > 0" style="padding: 12px 0; text-align: center">
          <button v-if="followingsHasMore" class="btn btn--outline btn--sm" @click="loadMoreFollowings" :disabled="followingsLoadingMore">
            <PhCircleNotch v-if="followingsLoadingMore" :size="14" class="spin" />
            <span v-else>加载更多</span>
          </button>
          <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
        </div>
      </div>

      <!-- 我的粉丝 -->
      <div v-show="activeTab === 'followers'">
        <div v-if="followers.length > 0">
          <div v-for="item in followers" :key="item.id"
            class="user-card card fade-up"
            @click="$router.push({ name: 'UserProfile', params: { username: item.user.username } })"
          >
            <UserAvatar :avatar-url="item.user.avatar_url" :name="item.user.nickname || item.user.username" size="md" />
            <div class="user-card__info">
              <div class="user-card__name">{{ item.user.nickname || item.user.username }}</div>
              <div class="user-card__sub text-4">@{{ item.user.username }}</div>
            </div>
            <button
              v-if="item.user.username !== auth.user?.username"
              class="btn btn--sm"
              :class="item.followed ? 'btn--primary' : 'btn--outline'"
              @click.stop="handleFollowUser(item)"
            >
              <PhUserMinus v-if="item.followed" :size="14" />
              <PhUserPlus v-else :size="14" />
              {{ item.followed ? '已关注' : '关注' }}
            </button>
          </div>
        </div>
        <div v-else-if="!followersLoading" class="empty-state card">
          <p class="text-3" style="font-size: 14px">还没有粉丝</p>
        </div>
        <div v-if="followers.length > 0" style="padding: 12px 0; text-align: center">
          <button v-if="followersHasMore" class="btn btn--outline btn--sm" @click="loadMoreFollowers" :disabled="followersLoadingMore">
            <PhCircleNotch v-if="followersLoadingMore" :size="14" class="spin" />
            <span v-else>加载更多</span>
          </button>
          <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
        </div>
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
import { ref, computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { roleLabel } from '../utils/role'
import { listUserLikes, getFollowers, getFollowings, toggleFollow } from '../api'
import { formatTime } from '../utils/format'
import { getCoverCSS } from '../config/coverThemes'
import UserAvatar from '../components/UserAvatar.vue'
import { useToast } from '../composables/useToast'
import {
  PhClock, PhGearSix, PhEye,
  PhUserPlus, PhUserMinus, PhCircleNotch
} from '@phosphor-icons/vue'

const auth = useAuthStore()
const toast = useToast()

// 封面背景 CSS 变量，供 cover/body 渐变使用
const coverStyle = computed(() => {
  if (!auth.user?.cover_theme) return {}
  return { '--cover-bg': getCoverCSS(auth.user.cover_theme) }
})

const stats = computed(() => [
  { label: '发帖数', value: auth.user?.post_count ?? '—' },
  { label: '回复数', value: auth.user?.comment_count ?? '—' },
  { label: '获赞数', value: auth.user?.like_count ?? '—' },
])

// ---- Tabs ----
const activeTab = ref('account')
// 记录已加载过的 tab，实现懒加载（切换才请求）
const loadedTabs = new Set(['account'])

const tabs = computed(() => [
  { key: 'account', label: '账户信息', count: null },
  { key: 'likes', label: '我的点赞', count: auth.user?.liked_count },
  { key: 'followings', label: '我的关注', count: auth.user?.following_count },
  { key: 'followers', label: '我的粉丝', count: auth.user?.follower_count },
])

function switchTab(key) {
  if (activeTab.value === key) return
  activeTab.value = key
  if (loadedTabs.has(key)) return
  loadedTabs.add(key)
  if (key === 'likes') loadLikes(true)
  else if (key === 'followings') loadFollowings(true)
  else if (key === 'followers') loadFollowers(true)
}

// ---- My Likes ----
const myLikes = ref([])
const likesPage = ref(1)
const likesHasMore = ref(true)
const likesLoading = ref(true)
const likesLoadingMore = ref(false)
const likesTotal = ref(0)
const LIKES_PAGE_SIZE = 10

async function loadLikes(reset = true) {
  if (!auth.user?.username) return
  if (reset) {
    likesPage.value = 1
    likesHasMore.value = true
    likesLoading.value = true
  } else {
    likesLoadingMore.value = true
  }
  try {
    const params = { page: likesPage.value, page_size: LIKES_PAGE_SIZE }
    const res = await listUserLikes(auth.user.username, params)
    const list = res.data?.list || []
    if (reset) myLikes.value = list; else myLikes.value.push(...list)
    likesTotal.value = res.data?.total || 0
    likesHasMore.value = myLikes.value.length < likesTotal.value
  } catch (e) { toast.error(e.message) }
  finally {
    likesLoading.value = false
    likesLoadingMore.value = false
  }
}

function loadMoreLikes() {
  likesPage.value++
  loadLikes(false)
}

// ---- Followings ----
const followings = ref([])
const followingsPage = ref(1)
const followingsHasMore = ref(true)
const followingsLoading = ref(true)
const followingsLoadingMore = ref(false)
const followingsTotal = ref(0)
const FOLLOWINGS_PAGE_SIZE = 20

async function loadFollowings(reset = true) {
  if (!auth.user?.username) return
  if (reset) {
    followingsPage.value = 1
    followingsHasMore.value = true
    followingsLoading.value = true
  } else {
    followingsLoadingMore.value = true
  }
  try {
    const params = { page: followingsPage.value, page_size: FOLLOWINGS_PAGE_SIZE }
    const res = await getFollowings(auth.user.username, params)
    const list = res.data?.list || []
    if (reset) followings.value = list; else followings.value.push(...list)
    followingsTotal.value = res.data?.total || 0
    followingsHasMore.value = followings.value.length < followingsTotal.value
  } catch (e) { toast.error(e.message) }
  finally {
    followingsLoading.value = false
    followingsLoadingMore.value = false
  }
}

function loadMoreFollowings() {
  followingsPage.value++
  loadFollowings(false)
}

// ---- Followers ----
const followers = ref([])
const followersPage = ref(1)
const followersHasMore = ref(true)
const followersLoading = ref(true)
const followersLoadingMore = ref(false)
const followersTotal = ref(0)
const FOLLOWERS_PAGE_SIZE = 20

async function loadFollowers(reset = true) {
  if (!auth.user?.username) return
  if (reset) {
    followersPage.value = 1
    followersHasMore.value = true
    followersLoading.value = true
  } else {
    followersLoadingMore.value = true
  }
  try {
    const params = { page: followersPage.value, page_size: FOLLOWERS_PAGE_SIZE }
    const res = await getFollowers(auth.user.username, params)
    const list = res.data?.list || []
    if (reset) followers.value = list; else followers.value.push(...list)
    followersTotal.value = res.data?.total || 0
    followersHasMore.value = followers.value.length < followersTotal.value
  } catch (e) { toast.error(e.message) }
  finally {
    followersLoading.value = false
    followersLoadingMore.value = false
  }
}

function loadMoreFollowers() {
  followersPage.value++
  loadFollowers(false)
}

// ---- Follow / Unfollow（列表内按钮） ----
// 关注/取关后行为随当前 tab 不同：
// - 我的关注：取关则从列表移除，徽标 -1
// - 我的粉丝：就地切换 followed 状态（粉丝关系不因对方是否关注而变化），徽标不变
async function handleFollowUser(item) {
  if (!auth.isLoggedIn) return
  try {
    const res = await toggleFollow(item.user.username)
    const followed = res.data?.followed
    item.followed = followed
    toast.success(followed ? '已关注' : '已取消关注')
    // 在「我的关注」tab 取关：从列表移除并更新本地计数
    if (!followed && activeTab.value === 'followings') {
      followings.value = followings.value.filter((f) => f.id !== item.id)
      if (auth.user && auth.user.following_count > 0) {
        auth.user.following_count--
      }
    }
  } catch (e) { toast.error(e.message) }
}

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
  flex-direction: column;
  padding: 0;
  overflow: hidden;
}

.profile-header__cover {
  height: 100px;
  width: 100%;
  flex-shrink: 0;
  background: var(--cover-bg, transparent);
}

.profile-header__body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 24px;
  background-image: linear-gradient(to bottom, transparent, var(--bg-card)), var(--cover-bg, none);
}

.profile-header__left {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
  flex: 1;
}

.profile-header__info {
  min-width: 0;
}

.profile-header__name {
  font-size: 22px;
  margin-bottom: 4px;
}

.profile-header__meta {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.profile-header__motto {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.5;
  /* 限制两行，超出省略 */
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.profile-header__right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 14px;
  flex-shrink: 0;
}

.profile-header__stats {
  display: flex;
  gap: 24px;
}

.profile-header__stat {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  min-width: 56px;
}

.profile-header__stat-value {
  font-size: 20px;
  font-weight: 700;
}

.profile-header__stat-label {
  font-size: 12px;
}

.profile-header__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

/* ---- Tab 徽标 ---- */
.profile-tabs {
  margin-bottom: 0;
}

.tab-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: var(--radius-full);
  background: var(--bg-muted);
  color: var(--text-3);
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-body);
  line-height: 1;
}

.tab-btn--active .tab-badge {
  background: var(--accent-light);
  color: var(--accent);
}

/* ---- 账户信息 ---- */
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

/* ---- 用户卡片（关注/粉丝列表） ---- */
.user-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: box-shadow var(--duration-medium-2) var(--ease-standard);
}

.user-card:hover {
  box-shadow: var(--shadow-2);
}

.user-card__info {
  flex: 1;
  min-width: 0;
}

.user-card__name {
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-card__sub {
  font-size: 12px;
  margin-top: 2px;
}

.profile__loading {
  padding: 20px;
}

@media (max-width: 720px) {
  .profile-header__body {
    flex-direction: column;
    align-items: flex-start;
    gap: 18px;
  }

  .profile-header__right {
    align-items: flex-start;
    width: 100%;
  }

  .profile-header__stats {
    width: 100%;
    justify-content: space-between;
    gap: 12px;
  }

  .profile-header__actions {
    width: 100%;
  }
}

@media (max-width: 600px) {
  .profile-header__stats {
    gap: 8px;
  }

  .profile-header__stat {
    min-width: 0;
    flex: 1;
  }

  .profile-header__stat-value {
    font-size: 17px;
  }
}
</style>