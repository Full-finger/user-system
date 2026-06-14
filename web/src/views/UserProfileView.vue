<template>
  <div class="user-profile fade-up">
    <template v-if="profile">
      <!-- Header（含座右铭 + 紧凑统计） -->
      <div class="card profile-header fade-up" :style="coverStyle">
        <div v-if="profile.cover_theme" class="profile-header__cover"></div>
        <div class="profile-header__body">
          <div class="profile-header__left">
            <UserAvatar :avatar-url="profile.avatar_url" :name="profile.nickname || profile.username" size="lg" />
            <div class="profile-header__info">
              <h1 class="font-display profile-header__name">{{ profile.nickname || profile.username }}</h1>
              <div class="profile-header__meta">
                <span class="pill pill--accent">{{ roleLabel(profile.role) }}</span>
                <span class="text-3" style="font-size: 13px">
                  <PhClock :size="12" style="vertical-align: -1px" />
                  注册于 {{ formatDate(profile.created_at) }}
                </span>
              </div>
              <p v-if="profile.motto" class="profile-header__motto text-3">{{ profile.motto }}</p>
            </div>
          </div>
          <div class="profile-header__right">
            <div class="profile-header__stats">
              <div class="profile-header__stat">
                <div class="profile-header__stat-value font-display">{{ profile.post_count }}</div>
                <div class="profile-header__stat-label text-3">发帖</div>
              </div>
              <div class="profile-header__stat">
                <div class="profile-header__stat-value font-display">{{ profile.like_count }}</div>
                <div class="profile-header__stat-label text-3">获赞</div>
              </div>
              <div class="profile-header__stat">
                <div class="profile-header__stat-value font-display">{{ profile.follower_count }}</div>
                <div class="profile-header__stat-label text-3">粉丝</div>
              </div>
              <div class="profile-header__stat">
                <div class="profile-header__stat-value font-display">{{ profile.following_count }}</div>
                <div class="profile-header__stat-label text-3">关注</div>
              </div>
            </div>
            <!-- 看自己：进入个人中心；看别人：关注按钮 -->
            <router-link
              v-if="auth.isLoggedIn && auth.user?.id === profile.id"
              to="/profile"
              class="btn btn--outline btn--sm"
            >
              <PhGearSix :size="14" />
              进入个人中心
            </router-link>
            <button
              v-else-if="auth.isLoggedIn && auth.user?.id !== profile.id"
              class="btn btn--sm"
              :class="followed ? 'btn--primary' : 'btn--outline'"
              @click="handleFollow"
            >
              <PhUserPlus :size="14" :weight="followed ? 'fill' : 'regular'" />
              {{ followed ? '已关注' : '关注' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Moderated Nodes -->
      <div v-if="profile.moderated_nodes?.length" class="card fade-up" style="animation-delay: 40ms; padding: 16px 20px; display: flex; align-items: center; gap: 10px; flex-wrap: wrap">
        <span class="text-3" style="font-size: 13px; display: flex; align-items: center; gap: 4px"><PhShieldStar :size="14" /> 版主节点</span>
        <router-link
          v-for="n in profile.moderated_nodes" :key="n.id"
          :to="{ name: 'NodePosts', params: { id: n.id } }"
          class="pill"
          :style="{ background: (n.color || '#9b8ec4') + '18', color: n.color || '#9b8ec4' }"
          style="text-decoration: none; cursor: pointer"
        >{{ n.name }}</router-link>
      </div>

      <!-- Posts -->
      <div v-if="userPosts.length > 0">
        <div v-for="(p, i) in userPosts" :key="p.code"
          class="post-card card fade-up"
          :style="{ animationDelay: ((i + 4) * 40) + 'ms' }"
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
      <div v-else-if="!postsLoading" class="empty-state card fade-up" style="padding: 32px">
        <p class="text-3" style="font-size: 14px">暂无帖子</p>
      </div>
      <div v-if="userPosts.length > 0" style="padding: 12px 0; text-align: center">
        <button v-if="hasMore" class="btn btn--outline btn--sm" @click="loadMore" :disabled="loadingMore">
          <PhCircleNotch v-if="loadingMore" :size="14" class="spin" />
          <span v-else>加载更多</span>
        </button>
        <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
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
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'
import { roleLabel } from '../utils/role'
import { getUserProfile, listUserPosts, toggleFollow } from '../api'
import { formatTime } from '../utils/format'
import { getCoverCSS } from '../config/coverThemes'
import UserAvatar from '../components/UserAvatar.vue'
import {
  PhClock, PhUserPlus, PhThumbsUp,
  PhChatCircle, PhMagnifyingGlass, PhCircleNotch, PhShieldStar, PhGearSix
} from '@phosphor-icons/vue'

const route = useRoute()
const auth = useAuthStore()
const toast = useToast()

const profile = ref(null)
const loading = ref(true)
const followed = ref(false)

// 封面背景 CSS 变量，供 cover/body 渐变使用
const coverStyle = computed(() => {
  if (!profile.value?.cover_theme) return {}
  return { '--cover-bg': getCoverCSS(profile.value.cover_theme) }
})

// Posts state
const userPosts = ref([])
const page = ref(1)
const hasMore = ref(true)
const postsLoading = ref(true)
const loadingMore = ref(false)
const total = ref(0)

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
  loadPosts(true)
}

async function loadPosts(reset = true) {
  if (reset) {
    page.value = 1
    hasMore.value = true
    postsLoading.value = true
  } else {
    loadingMore.value = true
  }

  try {
    const params = { page: page.value, page_size: PAGE_SIZE }
    const res = await listUserPosts(route.params.username, params)
    const list = res.data?.list || []
    if (reset) userPosts.value = list; else userPosts.value.push(...list)
    total.value = res.data?.total || 0
    hasMore.value = userPosts.value.length < total.value
  } catch (e) { toast.error(e.message) }
  finally {
    postsLoading.value = false
    loadingMore.value = false
  }
}

function loadMore() {
  page.value++
  loadPosts(false)
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
  display: flex; flex-direction: column; padding: 0; overflow: hidden;
}
.profile-header__cover {
  height: 100px; width: 100%; flex-shrink: 0;
  background: var(--cover-bg, transparent);
}
.profile-header__body {
  display: flex; align-items: center; justify-content: space-between; gap: 24px; padding: 24px;
  background-image: linear-gradient(to bottom, transparent, var(--bg-card)), var(--cover-bg, none);
}
.profile-header__left { display: flex; align-items: center; gap: 16px; min-width: 0; flex: 1; }
.profile-header__info { min-width: 0; }
.profile-header__name { font-size: 22px; margin-bottom: 4px; }
.profile-header__meta { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.profile-header__motto {
  margin-top: 6px; font-size: 13px; line-height: 1.5;
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
}
.profile-header__right {
  display: flex; flex-direction: column; align-items: flex-end; gap: 14px; flex-shrink: 0;
}

.profile-header__stats { display: flex; gap: 20px; }
.profile-header__stat {
  display: flex; flex-direction: column; align-items: center; gap: 2px; min-width: 48px;
}
.profile-header__stat-value { font-size: 18px; font-weight: 700; }
.profile-header__stat-label { font-size: 12px; }

.user-profile {
  display: flex; flex-direction: column; gap: 16px;
}

@media (max-width: 720px) {
  .profile-header__body { flex-direction: column; gap: 18px; align-items: flex-start; }
  .profile-header__right { align-items: flex-start; width: 100%; }
  .profile-header__stats { width: 100%; justify-content: space-between; gap: 8px; }
}

@media (max-width: 600px) {
  .profile-header__stats { gap: 6px; }
  .profile-header__stat { min-width: 0; flex: 1; }
  .profile-header__stat-value { font-size: 16px; }
}
</style>