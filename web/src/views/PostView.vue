<template>
  <div class="post-detail">
    <!-- Loading -->
    <div v-if="loading">
      <div class="skeleton" style="height: 40px; margin-bottom: 12px; width: 60%"></div>
      <div class="skeleton" style="height: 20px; margin-bottom: 8px"></div>
      <div class="skeleton" style="height: 200px"></div>
    </div>

    <template v-else-if="post">
      <!-- Breadcrumb -->
      <div class="post-detail__breadcrumb fade-up text-3" style="font-size: 13px; margin-bottom: 16px">
        <router-link to="/" class="text-3">发现</router-link>
        <PhCaretRight :size="12" />
        <router-link :to="{ name: 'NodePosts', params: { id: post.node?.id } }" class="text-3" :style="{ color: post.node?.color }">
          {{ post.node?.name }}
        </router-link>
      </div>

      <!-- Post header -->
      <div class="card post-detail__header fade-up" style="animation-delay: 40ms">
        <div style="position: absolute; left: 0; top: 0; bottom: 0; width: 4px; border-radius: var(--radius-m) 0 0 var(--radius-m)" :style="{ background: post.node?.color || 'var(--accent)' }"></div>
        <div class="post-detail__header-top">
          <h1 class="font-display post-detail__title">{{ post.title }}</h1>
          <span class="pill" :style="{ background: (post.node?.color || '#c47a99') + '18', color: post.node?.color || '#c47a99' }">
            {{ post.node?.name }}
          </span>
        </div>
        <div class="post-detail__meta text-4">
          <span class="post-card__author">
            <span class="post-card__online-dot" style="background: var(--text-4)"></span>
            <router-link v-if="post.user" :to="{ name: 'UserProfile', params: { username: post.user.username } }" style="font-weight: 500">{{ post.user.nickname || post.user.username }}</router-link>
            <span v-else>匿名</span>
          </span>
          <span><PhClock :size="12" style="vertical-align: -1px" /> {{ formatTime(post.created_at) }}</span>
          <span><PhEye :size="12" style="vertical-align: -1px" /> {{ post.view_count }}</span>
        </div>

        <!-- Mentions -->
        <div v-if="post.mentions?.length" class="post-detail__mentions" style="margin-top: 12px">
          <span v-for="m in post.mentions" :key="m.id" class="pill pill--lavender" style="font-size: 11px">
            @{{ m.username }}
          </span>
        </div>

        <!-- Actions -->
        <div class="post-detail__actions">
          <button class="post-detail__action-btn" :class="{ 'post-detail__action-btn--active': liked }" @click="handleLike">
            <PhThumbsUp :size="18" :weight="liked ? 'fill' : 'regular'" />
            <span>{{ post.like_count }}</span>
          </button>
          <button v-if="isAuthor || auth.canManagePosts" class="post-detail__action-btn post-detail__action-btn--danger" @click="handleDelete">
            <PhTrash :size="18" />
            <span>删除</span>
          </button>
        </div>
      </div>

      <!-- Post content -->
      <div class="card post-detail__body fade-up" style="animation-delay: 80ms">
        <div class="post-detail__content text-2" v-html="renderContent(post.content)"></div>
      </div>
    </template>

    <!-- Not found -->
    <div v-else class="empty-state card fade-up">
      <div class="empty-state__icon"><PhMagnifyingGlass :size="32" weight="bold" /></div>
      <h2 class="font-display" style="font-size: 18px; margin-bottom: 6px">帖子不存在</h2>
      <p class="text-3" style="font-size: 14px">该帖子可能已被删除</p>
      <router-link to="/" class="btn btn--outline btn--sm" style="margin-top: 12px">返回首页</router-link>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'
import { getPost, toggleLikePost, deletePost } from '../api'
import {
  PhCaretRight, PhClock, PhEye, PhThumbsUp, PhTrash, PhMagnifyingGlass
} from '@phosphor-icons/vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const toast = useToast()

const post = ref(null)
const loading = ref(true)
const liked = ref(false)

const isAuthor = computed(() => auth.user && post.value?.user?.id === auth.user.id)

async function fetchPost() {
  loading.value = true
  try {
    const res = await getPost(route.params.id)
    post.value = res.data
    liked.value = res.data?.liked || false
  } catch (e) {
    toast.error(e.message)
    post.value = null
  } finally {
    loading.value = false
  }
}

async function handleLike() {
  if (!auth.isLoggedIn) return
  try {
    const res = await toggleLikePost(route.params.id)
    liked.value = res.data?.liked
    if (liked.value) post.value.like_count++
    else post.value.like_count = Math.max(0, post.value.like_count - 1)
  } catch (e) { toast.error(e.message) }
}

async function handleDelete() {
  if (!confirm('确定要删除这个帖子吗？')) return
  try {
    await deletePost(route.params.id)
    toast.success('已删除')
    router.push('/')
  } catch (e) { toast.error(e.message || '删除失败') }
}

function renderContent(content) {
  if (!content) return ''
  // Simple rendering: escape HTML, convert @mentions to highlighted spans, preserve newlines
  let html = content
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/@([a-zA-Z0-9_-]{2,50})/g, '<span style="color: var(--accent); font-weight: 500">@$1</span>')
    .replace(/\n/g, '<br>')
  return html
}

function formatTime(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr); const diff = Math.floor((Date.now() - d) / 1000)
  if (diff < 60) return '刚刚'
  if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前'
  if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前'
  if (diff < 604800) return Math.floor(diff / 86400) + ' 天前'
  return d.toLocaleDateString('zh-CN')
}

onMounted(() => fetchPost())
</script>

<style scoped>
.post-detail__breadcrumb {
  display: flex; align-items: center; gap: 6px;
}

.post-detail__header {
  padding: 24px; position: relative; margin-bottom: 12px;
}

.post-detail__header-top {
  display: flex; align-items: flex-start; gap: 10px; margin-bottom: 10px; flex-wrap: wrap;
}

.post-detail__title {
  font-size: 22px; line-height: 1.4; flex: 1; min-width: 0;
}

.post-detail__meta {
  display: flex; align-items: center; gap: 12px; font-size: 13px; margin-bottom: 4px;
}

.post-detail__actions {
  display: flex; align-items: center; gap: 8px; margin-top: 16px; padding-top: 16px; border-top: 1px solid var(--border);
}

.post-detail__action-btn {
  display: flex; align-items: center; gap: 6px; padding: 6px 14px;
  border: 1.5px solid var(--border); border-radius: var(--radius-full);
  background: none; color: var(--text-3); cursor: pointer;
  font-family: var(--font-display); font-size: 13px; font-weight: 500;
  position: relative; overflow: hidden;
  transition: border-color var(--duration-medium-1) var(--ease-standard),
              color var(--duration-medium-1) var(--ease-standard);
}

.post-detail__action-btn::after {
  content: ''; position: absolute; inset: 0; border-radius: inherit;
  background: transparent; transition: background var(--duration-medium-1) var(--ease-standard);
  pointer-events: none;
}

.post-detail__action-btn:hover { border-color: var(--accent); color: var(--accent); }
.post-detail__action-btn:hover::after { background: var(--state-hover); }

.post-detail__action-btn--active {
  border-color: var(--accent); color: var(--accent); background: var(--accent-light);
}

.post-detail__action-btn--danger:hover { border-color: #c47878; color: #c47878; }
.post-detail__action-btn--danger:hover::after { background: rgba(196, 120, 120, 0.08); }

.post-detail__body { padding: 24px; }

.post-detail__content {
  font-size: 15px; line-height: 1.7; white-space: pre-wrap; word-break: break-word;
}

.post-detail__mentions {
  display: flex; gap: 6px; flex-wrap: wrap;
}
</style>
