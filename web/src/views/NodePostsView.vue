<template>
  <div class="node-posts fade-up">
    <!-- Node header -->
    <div v-if="node" class="card fade-up" style="display: flex; gap: 16px; padding: 24px; align-items: flex-start; position: relative; margin-bottom: 16px">
      <div style="position: absolute; left: 0; top: 0; bottom: 0; width: 4px; border-radius: var(--radius-m) 0 0 var(--radius-m)" :style="{ background: node.color || 'var(--accent)' }"></div>
      <div class="node-header__icon" :style="{ background: (node.color || 'var(--accent)') + '14', color: node.color || 'var(--accent)' }">
        <PhStack :size="24" />
      </div>
      <div style="display: flex; flex-direction: column; gap: 4px">
        <h1 class="font-display" style="font-size: 22px">{{ node.name }}</h1>
        <p class="text-3" style="font-size: 13px">{{ node.desc }}</p>
        <span class="text-4" style="font-size: 12px"><PhNote :size="12" style="vertical-align: -1px" /> {{ node.post_count || 0 }} 帖子</span>
      </div>
    </div>

    <!-- Sort tabs -->
    <div class="tab-bar fade-up" style="animation-delay: 40ms">
      <button class="tab-btn" :class="{ 'tab-btn--active': sort === 'time' }" @click="switchSort('time')">
        <PhClock :size="14" /> 最新
      </button>
      <button class="tab-btn" :class="{ 'tab-btn--active': sort === 'replies' }" @click="switchSort('replies')">
        <PhChatCircle :size="14" /> 最多回复
      </button>
      <div style="flex:1"></div>
      <button v-if="auth.isLoggedIn" class="btn btn--primary btn--sm" @click="showCreateModal = true">
        <PhPencilSimpleLine :size="14" /> 发帖
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" style="padding-top: 8px">
      <div class="skeleton" style="height: 80px; margin-bottom: 8px"></div>
      <div class="skeleton" style="height: 80px; margin-bottom: 8px"></div>
    </div>

    <!-- Posts -->
    <div v-else-if="posts.length > 0">
      <div
        v-for="(post, i) in posts" :key="post.id"
        class="post-card card fade-up"
        :style="{ animationDelay: (80 + i * 40) + 'ms' }"
        @click="$router.push({ name: 'PostDetail', params: { id: post.id } })"
      >
        <div class="post-card__bar" :style="{ background: node?.color || 'var(--accent)' }"></div>
        <div class="post-card__vote">
          <button class="post-card__vote-btn" @click.stop="handleLike(post)">
            <PhThumbsUp :size="16" :weight="likedPosts.has(post.id) ? 'fill' : 'regular'" />
          </button>
          <span class="font-display" style="font-size: 14px; font-weight: 600">{{ post.like_count }}</span>
        </div>
        <div class="post-card__content">
          <div class="post-card__top">
            <h3 class="post-card__title">{{ post.title }}</h3>
          </div>
          <p class="post-card__desc text-3">{{ post.content }}</p>
          <div class="post-card__meta text-4">
            <span class="post-card__author">
              <span class="post-card__online-dot" style="background: var(--text-4)"></span>
              {{ post.user?.username || '匿名' }}
            </span>
            <span><PhClock :size="12" style="vertical-align: -1px" /> {{ formatTime(post.created_at) }}</span>
            <span><PhChatCircle :size="12" style="vertical-align: -1px" /> {{ post.reply_count }}</span>
            <span><PhEye :size="12" style="vertical-align: -1px" /> {{ post.view_count }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-else class="empty-state card fade-up" style="margin-top: 8px; animation-delay: 80ms">
      <div class="empty-state__icon"><PhNote :size="32" weight="bold" /></div>
      <h2 class="font-display" style="font-size: 18px; margin-bottom: 6px">暂无帖子</h2>
      <p class="text-3" style="font-size: 14px">成为第一个发帖的人吧！</p>
    </div>

    <!-- Load more -->
    <div v-if="posts.length > 0" style="padding: 12px 0; text-align: center; margin-top: 8px">
      <button v-if="hasMore" class="btn btn--outline btn--sm" @click="loadMore" :disabled="loadingMore">
        <PhCircleNotch v-if="loadingMore" :size="14" class="spin" />
        <span v-else>加载更多</span>
      </button>
      <p v-else class="text-4" style="font-size: 12px">已经到底了</p>
    </div>

    <!-- Create Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
          <div class="modal-panel card">
            <div class="modal-panel__header">
              <h2 class="font-display" style="font-size: 18px">发布新帖</h2>
              <button class="modal-panel__close" @click="showCreateModal = false"><PhX :size="18" /></button>
            </div>
            <div class="modal-panel__body">
              <div class="auth-form__group">
                <label class="auth-form__label">节点</label>
                <input class="input" :value="node?.name" disabled style="border: none; background: transparent" />
              </div>
              <div class="auth-form__group">
                <label class="auth-form__label">标题</label>
                <input v-model="newPost.title" class="input" placeholder="输入帖子标题..." style="border: none; background: transparent" />
              </div>
              <div class="auth-form__group" style="min-height: 120px">
                <label class="auth-form__label">内容（支持 @username 提及他人）</label>
                <textarea v-model="newPost.content" class="input" placeholder="写下你的想法..." style="border: none; background: transparent; resize: vertical; min-height: 80px"></textarea>
              </div>
            </div>
            <div class="modal-panel__footer">
              <button class="btn btn--outline btn--sm" @click="showCreateModal = false">取消</button>
              <button class="btn btn--primary btn--sm" @click="handleCreatePost" :disabled="!newPost.title || !newPost.content || creating">
                <PhCircleNotch v-if="creating" :size="14" class="spin" />
                <span v-else>发布</span>
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { getNode, getNodePosts, createPost, toggleLikePost } from '../api'
import {
  PhStack, PhNote, PhClock, PhChatCircle, PhEye, PhThumbsUp,
  PhPencilSimpleLine, PhCircleNotch, PhX
} from '@phosphor-icons/vue'

const route = useRoute()
const auth = useAuthStore()

const node = ref(null)
const posts = ref([])
const loading = ref(true)
const loadingMore = ref(false)
const page = ref(1)
const hasMore = ref(true)
const sort = ref('time')
const likedPosts = ref(new Set())

const showCreateModal = ref(false)
const creating = ref(false)
const newPost = reactive({ title: '', content: '' })

async function fetchNode() {
  try { const res = await getNode(route.params.id); node.value = res.data } catch (e) { console.error(e) }
}

async function fetchPosts(reset = true) {
  if (reset) { page.value = 1; posts.value = []; loading.value = true; hasMore.value = true }
  try {
    const res = await getNodePosts(route.params.id, { page: page.value, page_size: 20, sort: sort.value })
    const list = res.data?.list || []
    if (reset) posts.value = list; else posts.value.push(...list)
    hasMore.value = posts.value.length < (res.data?.total || 0)
  } catch (e) { console.error(e) }
  finally { loading.value = false; loadingMore.value = false }
}

function switchSort(s) { sort.value = s; fetchPosts(true) }
function loadMore() { page.value++; loadingMore.value = true; fetchPosts(false) }

async function handleLike(post) {
  if (!auth.isLoggedIn) return
  try {
    const res = await toggleLikePost(post.id)
    if (res.data?.liked) { likedPosts.value.add(post.id); post.like_count++ }
    else { likedPosts.value.delete(post.id); post.like_count = Math.max(0, post.like_count - 1) }
  } catch (e) { console.error(e) }
}

async function handleCreatePost() {
  if (!newPost.title || !newPost.content) return
  creating.value = true
  try {
    await createPost({ node_id: Number(route.params.id), title: newPost.title, content: newPost.content })
    showCreateModal.value = false; newPost.title = ''; newPost.content = ''
    fetchPosts(true); fetchNode()
  } catch (e) { alert(e.message || '发帖失败') }
  finally { creating.value = false }
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

onMounted(() => { fetchNode(); fetchPosts() })
watch(() => route.params.id, () => { fetchNode(); fetchPosts() })
</script>

<style scoped>
.node-header__icon {
  display: flex; align-items: center; justify-content: center;
  width: 48px; height: 48px; border-radius: var(--radius-m); flex-shrink: 0;
}
</style>
