<template>
  <div class="home">
    <div class="home__header fade-up">
      <h1 class="font-display home__title">发现</h1>
      <button class="btn btn--primary" @click="showCreateModal = true" v-if="auth.isLoggedIn">
        <PhPencilSimpleLine :size="16" />
        发帖
      </button>
    </div>

    <div class="tab-bar fade-up" style="animation-delay: 40ms">
      <button
        v-for="tab in tabs" :key="tab.key"
        class="tab-btn" :class="{ 'tab-btn--active': activeTab === tab.key }"
        @click="switchTab(tab.key)"
      >
        <component :is="tab.icon" :size="14" />
        {{ tab.label }}
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" style="padding-top: 8px">
      <div class="skeleton" style="height: 80px; margin-bottom: 8px"></div>
      <div class="skeleton" style="height: 80px; margin-bottom: 8px"></div>
      <div class="skeleton" style="height: 80px"></div>
    </div>

    <!-- Posts -->
    <div v-else-if="posts.length > 0">
      <div
        v-for="(post, i) in posts" :key="post.code"
        class="post-card card fade-up"
        :style="{ animationDelay: (80 + i * 40) + 'ms' }"
        @click="$router.push({ name: 'PostDetail', params: { code: post.code } })"
      >
        <div class="post-card__bar" :style="{ background: post.node?.color || 'var(--accent)' }"></div>
        <div class="post-card__vote">
          <button class="post-card__vote-btn" @click.stop="handleLike(post)">
            <PhThumbsUp :size="16" :weight="likedPosts.has(post.code) ? 'fill' : 'regular'" />
          </button>
          <span class="font-display" style="font-size: 14px; font-weight: 600">{{ post.like_count }}</span>
        </div>
        <div class="post-card__content">
          <div class="post-card__top">
            <h3 class="post-card__title">{{ post.title }}</h3>
            <span class="pill" :style="{ background: (post.node?.color || '#c47a99') + '18', color: post.node?.color || '#c47a99' }">
              {{ post.node?.name || '未知' }}
            </span>
          </div>
          <p class="post-card__desc text-3">{{ post.content }}</p>
          <div class="post-card__meta text-4">
            <span class="post-card__author">
              <span class="post-card__online-dot" style="background: var(--text-4)"></span>
              {{ post.user?.nickname || post.user?.username || '匿名' }}
            </span>
            <span><PhClock :size="12" style="vertical-align: -1px" /> {{ formatTime(post.created_at) }}</span>
            <span><PhChatCircle :size="12" style="vertical-align: -1px" /> {{ post.reply_count }}</span>
            <span><PhEye :size="12" style="vertical-align: -1px" /> {{ formatCount(post.view_count) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-else class="empty-state card fade-up" style="animation-delay: 80ms">
      <div class="empty-state__icon"><PhNote :size="32" weight="bold" /></div>
      <h2 class="font-display" style="font-size: 18px; margin-bottom: 6px">还没有帖子</h2>
      <p class="text-3" style="font-size: 14px">成为第一个发帖的人吧！</p>
      <button v-if="auth.isLoggedIn" class="btn btn--primary btn--sm" style="margin-top: 12px" @click="showCreateModal = true">
        <PhPencilSimpleLine :size="14" /> 发帖
      </button>
    </div>

    <!-- Load more -->
    <div v-if="posts.length > 0" style="padding: 12px 0; text-align: center">
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
                <label class="auth-form__label">选择节点</label>
                <select v-model="newPost.node_id" class="input" style="height: 38px">
                  <option :value="null" disabled>请选择节点</option>
                  <option v-for="node in allNodes" :key="node.id" :value="node.id">{{ node.name }}</option>
                </select>
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
              <button class="btn btn--primary btn--sm" @click="handleCreatePost" :disabled="!newPost.node_id || !newPost.title || !newPost.content || creating">
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
import { ref, reactive, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'
import { listPosts, listFeed, createPost, toggleLikePost, listNodes } from '../api'
import { formatTime, formatCount } from '../utils/format'
import {
  PhPencilSimpleLine, PhThumbsUp, PhClock, PhChatCircle, PhEye,
  PhHouse, PhCompass, PhNote, PhCircleNotch, PhX
} from '@phosphor-icons/vue'

const auth = useAuthStore()
const toast = useToast()
const activeTab = ref('all')
const tabs = [
  { key: 'all', label: '全部', icon: PhHouse },
  { key: 'feed', label: '关注', icon: PhCompass },
]

const posts = ref([])
const loading = ref(true)
const loadingMore = ref(false)
const page = ref(1)
const hasMore = ref(true)
const likedPosts = ref(new Set())
const allNodes = ref([])

const showCreateModal = ref(false)
const creating = ref(false)
const newPost = reactive({ node_id: null, title: '', content: '' })

async function fetchPosts(reset = true) {
  if (reset) { page.value = 1; posts.value = []; loading.value = true; hasMore.value = true }
  try {
    const params = { page: page.value, page_size: 20 }
    const res = activeTab.value === 'feed' ? await listFeed(params) : await listPosts(params)
    const list = res.data?.list || []
    if (reset) {
      posts.value = list
      likedPosts.value = new Set(list.filter(p => p.liked).map(p => p.code))
    } else {
      posts.value.push(...list)
      list.forEach(p => { if (p.liked) likedPosts.value.add(p.code) })
    }
    hasMore.value = posts.value.length < (res.data?.total || 0)
  } catch (e) { toast.error(e.message) }
  finally { loading.value = false; loadingMore.value = false }
}

function switchTab(key) { activeTab.value = key; fetchPosts(true) }
function loadMore() { page.value++; loadingMore.value = true; fetchPosts(false) }

async function handleLike(post) {
  if (!auth.isLoggedIn) return
  try {
    const res = await toggleLikePost(post.code)
    if (res.data?.liked) { likedPosts.value.add(post.code); post.like_count++ }
    else { likedPosts.value.delete(post.code); post.like_count = Math.max(0, post.like_count - 1) }
  } catch (e) { toast.error(e.message) }
}

async function handleCreatePost() {
  if (!newPost.node_id || !newPost.title || !newPost.content) return
  creating.value = true
  try {
    await createPost({ node_id: newPost.node_id, title: newPost.title, content: newPost.content })
    showCreateModal.value = false
    newPost.node_id = null; newPost.title = ''; newPost.content = ''
    fetchPosts(true)
    toast.success('发布成功')
  } catch (e) { toast.error(e.message || '发帖失败') }
  finally { creating.value = false }
}


onMounted(async () => {
  await fetchPosts()
  try { const res = await listNodes(); allNodes.value = res.data?.nodes || [] } catch (e) { toast.error(e.message) }
})
</script>

<style scoped>
.home__header {
  display: flex; align-items: center; justify-content: space-between; margin-bottom: 20px;
}
.home__title { font-size: 26px; }

select.input {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='%237e7290' viewBox='0 0 256 256'%3E%3Cpath d='M213.66,101.66l-80,80a8,8,0,0,1-11.32,0l-80-80A8,8,0,0,1,53.66,90.34L128,164.69l74.34-74.35a8,8,0,0,1,11.32,11.32Z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
}
</style>
