<template>
  <div class="admin-comments">
    <div class="admin__toolbar card">
      <div style="display: flex; gap: 8px; align-items: center; flex: 1; min-width: 0">
        <input v-model="keyword" type="text" class="input" placeholder="搜索评论内容..." style="max-width: 240px" @keyup.enter="search" />
      </div>
      <div style="display: flex; gap: 8px; align-items: center">
        <span class="text-3" style="font-size: 13px">共 {{ total }} 条</span>
        <button class="btn btn--outline btn--sm" @click="fetchComments">
          <PhArrowClockwise :size="14" /> 刷新
        </button>
      </div>
    </div>

    <div class="card admin__table-wrap">
      <div v-if="loading" style="padding: 40px; text-align: center">
        <PhCircleNotch :size="24" class="spin" style="color: var(--text-4)" />
      </div>

      <div v-else-if="error" style="padding: 40px; text-align: center">
        <PhXCircle :size="24" style="color: var(--color-danger)" />
        <p style="margin-top: 8px; font-size: 13px; color: var(--color-danger)">{{ error }}</p>
        <button class="btn btn--outline btn--sm" style="margin-top: 12px" @click="fetchComments">重试</button>
      </div>

      <div v-else-if="comments.length === 0" class="empty-state">
        <div class="empty-state__icon"><PhChatCircle :size="24" /></div>
        <p class="text-3" style="font-size: 14px">暂无评论</p>
      </div>

      <table v-else class="admin__table">
        <thead>
          <tr>
            <th>内容</th>
            <th>作者</th>
            <th style="text-align: center">赞</th>
            <th>发布时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="comment in comments" :key="comment.id">
            <td style="max-width: 360px">
              <div style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">
                {{ truncate(comment.content, 60) }}
              </div>
            </td>
            <td>
              <router-link
                v-if="comment.user?.username"
                :to="{ name: 'UserProfile', params: { username: comment.user.username } }"
                style="color: var(--text-2); text-decoration: none"
              >
                @{{ comment.user.username }}
              </router-link>
              <span v-else class="text-4">—</span>
            </td>
            <td style="text-align: center">{{ comment.like_count || 0 }}</td>
            <td class="text-3" style="font-size: 12px">{{ formatDate(comment.created_at) }}</td>
            <td>
              <button class="btn btn--sm btn--danger" @click="handleDelete(comment)" title="删除">
                <PhTrash :size="13" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="total > pageSize" class="admin__pagination">
        <button class="btn btn--outline btn--sm" :disabled="page <= 1" @click="page--; fetchComments()">
          <PhCaretLeft :size="14" /> 上一页
        </button>
        <span class="text-3" style="font-size: 13px">{{ page }} / {{ totalPages }}</span>
        <button class="btn btn--outline btn--sm" :disabled="page >= totalPages" @click="page++; fetchComments()">
          下一页 <PhCaretRight :size="14" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useToast } from '../../composables/useToast'
import { adminListComments, adminDeleteComment } from '../../api'
import {
  PhArrowClockwise, PhTrash, PhCaretLeft, PhCaretRight,
  PhXCircle, PhCircleNotch, PhChatCircle
} from '@phosphor-icons/vue'

const comments = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const error = ref('')
const keyword = ref('')
const toast = useToast()

const totalPages = computed(() => Math.ceil(total.value / pageSize))

function truncate(str, len) {
  if (!str) return ''
  return str.length > len ? str.slice(0, len) + '...' : str
}

function search() {
  page.value = 1
  fetchComments()
}

async function fetchComments() {
  loading.value = true
  error.value = ''
  try {
    const params = { page: page.value, page_size: pageSize }
    if (keyword.value) params.keyword = keyword.value
    const res = await adminListComments(params)
    comments.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function handleDelete(comment) {
  if (!confirm('确定要删除该评论吗？')) return
  try {
    await adminDeleteComment(comment.id)
    toast.success('已删除')
    if (comments.value.length <= 1 && page.value > 1) page.value--
    fetchComments()
  } catch (e) {
    toast.error(e.message)
  }
}

function formatDate(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

onMounted(fetchComments)
</script>

<style scoped>
.admin__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  gap: 12px;
  flex-wrap: wrap;
}

.admin__table-wrap { overflow-x: auto; }

.admin__table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}

.admin__table th {
  text-align: left;
  padding: 12px 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-3);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border);
  font-family: var(--font-body);
}

.admin__table td {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  vertical-align: middle;
}

.admin__table tbody tr {
  transition: background var(--duration-medium-1) var(--ease-standard);
}

.admin__table tbody tr:hover { background: var(--state-hover); }
.admin__table tbody tr:last-child td { border-bottom: none; }

.admin__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 16px;
  border-top: 1px solid var(--border);
}

@media (max-width: 600px) {
  .admin__table th:nth-child(3),
  .admin__table td:nth-child(3) { display: none; }
}
</style>