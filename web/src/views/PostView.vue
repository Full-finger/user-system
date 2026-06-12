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
            @{{ m.nickname || m.username }}
          </span>
        </div>

        <!-- Actions -->
        <div class="post-detail__actions">
          <button class="post-detail__action-btn" :class="{ 'post-detail__action-btn--active': liked }" @click="handleLike">
            <PhThumbsUp :size="18" :weight="liked ? 'fill' : 'regular'" />
            <span>{{ post.like_count }}</span>
          </button>
          <button class="post-detail__action-btn" @click="scrollToComment">
            <PhChatCircle :size="18" />
            <span>{{ post.reply_count }}</span>
          </button>
          <button v-if="isAuthor || auth.canManagePosts" class="post-detail__action-btn post-detail__action-btn--danger" @click="handleDelete">
            <PhTrash :size="18" />
            <span>删除</span>
          </button>
        </div>
      </div>

      <!-- Post content -->
      <div class="card post-detail__body fade-up" style="animation-delay: 80ms">
        <div class="post-detail__content text-2" v-html="renderContent(post.content, post.mentions)" @click="handleContentClick"></div>
      </div>

      <!-- Comment section -->
      <div ref="commentSectionRef" class="comment-section fade-up" style="animation-delay: 120ms">
        <h3 class="font-display comment-section__title">
          <PhChatCircleDots :size="20" style="vertical-align: -3px" /> 评论
          <span v-if="commentTotal > 0" class="text-4" style="font-size: 14px; font-weight: 400">({{ commentTotal }})</span>
        </h3>

        <!-- Comment input -->
        <div v-if="auth.isLoggedIn" class="card comment-input">
          <MentionInput
            v-model="newComment"
            :placeholder="replyTarget ? `回复 @${replyTarget.user.nickname || replyTarget.user.username}...` : '写下你的评论...'"
            :rows="3"
            :node-id="post?.node?.id"
            @keydown="handleCommentKeydown"
          />
          <div class="comment-input__footer">
            <span class="text-4" style="font-size: 12px">Ctrl+Enter 发送</span>
            <div style="display: flex; gap: 8px; align-items: center">
              <button v-if="replyTarget" class="btn btn--ghost btn--sm" @click="cancelReply">取消回复</button>
              <button class="btn btn--primary btn--sm" :disabled="!newComment.trim() || submitting" @click="submitComment">
                {{ submitting ? '发送中...' : '发送' }}
              </button>
            </div>
          </div>
        </div>
        <div v-else class="card comment-input comment-input--guest">
          <router-link :to="{ name: 'Login', query: { redirect: route.fullPath } }" class="btn btn--outline btn--sm">登录后评论</router-link>
        </div>

        <!-- Comment list -->
        <div v-if="comments.length" class="comment-list">
          <div v-for="cm in comments" :key="cm.id" class="comment-item card">
            <div class="comment-item__header">
              <router-link :to="{ name: 'UserProfile', params: { username: cm.user.username } }" class="comment-item__author">
                <div class="comment-item__avatar" :style="{ background: getAvatarColor(cm.user.username) }">
                  {{ (cm.user.nickname || cm.user.username).charAt(0).toUpperCase() }}
                </div>
                <span style="font-weight: 500; font-size: 14px">{{ cm.user.nickname || cm.user.username }}</span>
              </router-link>
              <span class="text-4" style="font-size: 12px">{{ formatTime(cm.created_at) }}</span>
            </div>

            <div class="comment-item__body text-2" v-html="renderCommentContent(cm)"></div>

            <div class="comment-item__actions">
              <button class="comment-action" :class="{ 'comment-action--active': cm.liked }" @click="handleCommentLike(cm)">
                <PhThumbsUp :size="14" :weight="cm.liked ? 'fill' : 'regular'" />
                <span v-if="cm.like_count">{{ cm.like_count }}</span>
              </button>
              <button v-if="auth.isLoggedIn" class="comment-action" @click="setReplyTarget(cm)">
                <PhArrowBendUpLeft :size="14" />
                <span>回复</span>
              </button>
            </div>

            <!-- Replies preview -->
            <div v-if="cm.replies?.length || cm.reply_count > 0" class="comment-replies">
              <div v-for="reply in (expandedReplies[cm.id] || cm.replies || [])" :key="reply.id" class="comment-reply-item">
                <div class="comment-item__header" style="margin-bottom: 4px">
                  <router-link :to="{ name: 'UserProfile', params: { username: reply.user.username } }" style="display: flex; align-items: center; gap: 6px; text-decoration: none; color: var(--text-1)">
                    <div class="comment-item__avatar comment-item__avatar--sm" :style="{ background: getAvatarColor(reply.user.username) }">
                      {{ (reply.user.nickname || reply.user.username).charAt(0).toUpperCase() }}
                    </div>
                    <span style="font-weight: 500; font-size: 13px">{{ reply.user.nickname || reply.user.username }}</span>
                  </router-link>
                  <span v-if="reply.reply_to" class="text-4" style="font-size: 12px">
                    回复 <router-link :to="{ name: 'UserProfile', params: { username: reply.reply_to.username } }" style="color: var(--accent)">@{{ reply.reply_to.username }}</router-link>
                  </span>
                  <span class="text-4" style="font-size: 12px">{{ formatTime(reply.created_at) }}</span>
                </div>
                <div class="comment-item__body" style="font-size: 13px; margin-bottom: 4px" v-html="renderCommentContent(reply)"></div>
                <div class="comment-item__actions" style="padding: 2px 0">
                  <button class="comment-action" :class="{ 'comment-action--active': reply.liked }" @click="handleCommentLike(reply)" style="font-size: 12px">
                    <PhThumbsUp :size="12" :weight="reply.liked ? 'fill' : 'regular'" />
                    <span v-if="reply.like_count">{{ reply.like_count }}</span>
                  </button>
                  <button v-if="auth.isLoggedIn" class="comment-action" @click="setReplyTarget(reply, cm)" style="font-size: 12px">
                    <PhArrowBendUpLeft :size="12" />
                  </button>
                </div>
              </div>

              <button v-if="cm.reply_count > (expandedReplies[cm.id]?.length || cm.replies?.length || 0)" class="comment-replies__more text-3" @click="loadMoreReplies(cm)">
                查看更多回复 ({{ cm.reply_count - (expandedReplies[cm.id]?.length || cm.replies?.length || 0) }})
              </button>
            </div>
          </div>

          <!-- Pagination -->
          <div v-if="commentTotal > commentPageSize" class="comment-pagination">
            <button class="btn btn--ghost btn--sm" :disabled="commentPage <= 1" @click="commentPage-- && fetchComments()">上一页</button>
            <span class="text-4" style="font-size: 13px">{{ commentPage }} / {{ Math.ceil(commentTotal / commentPageSize) }}</span>
            <button class="btn btn--ghost btn--sm" :disabled="commentPage * commentPageSize >= commentTotal" @click="commentPage++ && fetchComments()">下一页</button>
          </div>
        </div>

        <div v-else-if="!commentsLoading" class="card empty-state empty-state--compact">
          <p class="text-3" style="font-size: 14px">暂无评论，来说两句吧~</p>
        </div>
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
import { getPost, toggleLikePost, deletePost, listComments, createComment, listReplies, toggleCommentLike } from '../api'
import { renderContent } from '../utils/render'
import MentionInput from '../components/MentionInput.vue'
import {
  PhCaretRight, PhClock, PhEye, PhThumbsUp, PhTrash, PhMagnifyingGlass,
  PhChatCircle, PhChatCircleDots, PhArrowBendUpLeft
} from '@phosphor-icons/vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const toast = useToast()

const post = ref(null)
const loading = ref(true)
const liked = ref(false)
const isAuthor = computed(() => auth.user && post.value?.user?.id === auth.user.id)

// Comment state
const commentSectionRef = ref(null)
const comments = ref([])
const commentsLoading = ref(false)
const commentTotal = ref(0)
const commentPage = ref(1)
const commentPageSize = 20
const expandedReplies = ref({}) // { commentId: [replies] } — 用户点击"查看更多"后缓存的完整回复

const newComment = ref('')
const submitting = ref(false)
const replyTarget = ref(null) // { comment, topLevelComment? }

// Avatar color hash
const avatarColors = ['#9b8ec4', '#6db89a', '#7ba4d4', '#d4a07a', '#c47a99', '#8bb8a8', '#d4b85a', '#c4987a']
function getAvatarColor(name) {
  let hash = 0
  for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
  return avatarColors[Math.abs(hash) % avatarColors.length]
}

function renderCommentContent(cm) {
  return renderContent(cm.content, cm.mentions || [])
}

// Fetch post
async function fetchPost() {
  loading.value = true
  try {
    const res = await getPost(route.params.code)
    post.value = res.data
    liked.value = res.data?.liked || false
  } catch (e) {
    toast.error(e.message)
    post.value = null
  } finally {
    loading.value = false
  }
}

// Fetch comments
async function fetchComments() {
  commentsLoading.value = true
  try {
    const res = await listComments(route.params.code, {
      page: commentPage.value,
      page_size: commentPageSize,
      reply_preview_size: 3
    })
    comments.value = res.data.list || []
    commentTotal.value = res.data.total || 0
    // Build replies map from backend reply data if any
    // (currently replies are loaded separately via loadMoreReplies)
  } catch (e) {
    toast.error(e.message)
  } finally {
    commentsLoading.value = false
  }
}

// Submit comment
async function submitComment() {
  if (!newComment.value.trim() || submitting.value) return
  submitting.value = true
  try {
    const data = { content: newComment.value.trim() }
    if (replyTarget.value) {
      data.parent_id = replyTarget.value.comment.id
    }
    await createComment(route.params.code, data)
    newComment.value = ''
    replyTarget.value = null
    toast.success('评论成功')
    fetchComments()
    if (post.value) post.value.reply_count++
  } catch (e) {
    toast.error(e.message)
  } finally {
    submitting.value = false
  }
}

// Reply target
function setReplyTarget(cm, topLevelComment) {
  replyTarget.value = { comment: cm, user: cm.user, topLevelComment: topLevelComment || cm }
  if (commentSectionRef.value) {
    const textarea = commentSectionRef.value.querySelector('.comment-input textarea')
    if (textarea) textarea.focus()
  }
}

function cancelReply() {
  replyTarget.value = null
}

// Like comment
async function handleCommentLike(cm) {
  if (!auth.isLoggedIn) return
  try {
    const res = await toggleCommentLike(cm.id)
    cm.liked = res.data?.liked
    cm.like_count += cm.liked ? 1 : -1
    cm.like_count = Math.max(0, cm.like_count)
  } catch (e) { toast.error(e.message) }
}

// Load replies
async function loadMoreReplies(cm) {
  try {
    const existing = expandedReplies.value[cm.id]?.length || cm.replies?.length || 0
    const res = await listReplies(cm.id, { page: 1, page_size: existing + 10 })
    expandedReplies.value[cm.id] = res.data.list || []
  } catch (e) { toast.error(e.message) }
}

// Post like
async function handleLike() {
  if (!auth.isLoggedIn) return
  try {
    const res = await toggleLikePost(route.params.code)
    liked.value = res.data?.liked
    if (liked.value) post.value.like_count++
    else post.value.like_count = Math.max(0, post.value.like_count - 1)
  } catch (e) { toast.error(e.message) }
}

// Delete post
async function handleDelete() {
  if (!confirm('确定要删除这个帖子吗？')) return
  try {
    await deletePost(route.params.code)
    toast.success('已删除')
    router.push('/')
  } catch (e) { toast.error(e.message || '删除失败') }
}

function handleContentClick(e) {
  const link = e.target.closest('.mention-link')
  if (link) {
    e.preventDefault()
    router.push({ name: 'UserProfile', params: { username: link.dataset.username } })
  }
}

function handleCommentKeydown(e) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    submitComment()
  }
}

function scrollToComment() {
  if (commentSectionRef.value) {
    commentSectionRef.value.scrollIntoView({ behavior: 'smooth' })
    const textarea = commentSectionRef.value.querySelector('.comment-input textarea')
    if (textarea) textarea.focus()
  }
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

onMounted(() => {
  fetchPost().then(() => fetchComments())
})
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

.post-detail__content :deep(.mention-link) {
  color: var(--accent); font-weight: 500; text-decoration: none;
}
.post-detail__content :deep(.mention-link:hover) {
  text-decoration: underline;
}

.post-detail__mentions {
  display: flex; gap: 6px; flex-wrap: wrap;
}

/* Comment section */
.comment-section {
  margin-top: 24px;
}

.comment-section__title {
  font-size: 18px; margin-bottom: 16px; display: flex; align-items: center; gap: 8px;
}

.comment-input {
  padding: 16px; margin-bottom: 16px;
}

.comment-input textarea {
  width: 100%; border: 1px solid var(--border); border-radius: var(--radius-m);
  padding: 12px; font-size: 14px; resize: vertical; min-height: 60px;
  background: var(--bg-primary); color: var(--text-1);
  font-family: var(--font-body);
}

.comment-input textarea:focus {
  outline: none; border-color: var(--accent);
}

.comment-input__footer {
  display: flex; justify-content: space-between; align-items: center; margin-top: 8px;
}

.comment-input--guest {
  display: flex; justify-content: center; padding: 20px;
}

/* Comment list */
.comment-list { display: flex; flex-direction: column; gap: 8px; }

.comment-item {
  padding: 16px;
}

.comment-item__header {
  display: flex; align-items: center; gap: 8px; margin-bottom: 8px; flex-wrap: wrap;
}

.comment-item__author {
  display: flex; align-items: center; gap: 8px;
  text-decoration: none; color: var(--text-1);
}

.comment-item__avatar {
  width: 28px; height: 28px; border-radius: 50%;
  display: flex; align-items: center; justify-content: center;
  color: white; font-size: 13px; font-weight: 600; flex-shrink: 0;
}

.comment-item__avatar--sm {
  width: 22px; height: 22px; font-size: 11px;
}

.comment-item__body {
  font-size: 14px; line-height: 1.6; word-break: break-word; margin-bottom: 8px;
}

.comment-item__body :deep(.mention-link) {
  color: var(--accent); font-weight: 500; text-decoration: none;
}

.comment-item__actions {
  display: flex; align-items: center; gap: 12px;
}

.comment-action {
  display: flex; align-items: center; gap: 4px;
  background: none; border: none; color: var(--text-4); cursor: pointer;
  font-size: 13px; padding: 2px 4px; border-radius: var(--radius-xs);
  transition: color var(--duration-short-4) var(--ease-standard);
}

.comment-action:hover { color: var(--text-2); }
.comment-action--active { color: var(--accent); }

/* Replies */
.comment-replies {
  margin-top: 8px; padding-top: 8px; padding-left: 20px;
  border-left: 2px solid var(--border);
}

.comment-reply-item {
  padding: 8px 0;
}

.comment-reply-item + .comment-reply-item {
  border-top: 1px solid var(--border);
}

.comment-replies__more {
  background: none; border: none; color: var(--accent); cursor: pointer;
  font-size: 13px; padding: 6px 0; font-family: var(--font-body);
}

.comment-replies__more:hover { text-decoration: underline; }

/* Pagination */
.comment-pagination {
  display: flex; align-items: center; justify-content: center; gap: 12px;
  padding: 16px 0;
}

/* Empty state compact */
.empty-state--compact {
  padding: 24px; text-align: center;
}
</style>