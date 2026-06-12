<template>
  <div class="admin-posts">
    <div class="admin__toolbar card">
      <div style="display: flex; gap: 8px; align-items: center; flex: 1; min-width: 0">
        <input v-model="keyword" type="text" class="input" placeholder="搜索帖子标题..." style="max-width: 240px" @keyup.enter="search" />
        <select v-model="nodeFilter" class="input" style="max-width: 160px" @change="search">
          <option value="">全部节点</option>
          <option v-for="node in nodes" :key="node.id" :value="node.id">{{ node.name }}</option>
        </select>
      </div>
      <div style="display: flex; gap: 8px; align-items: center">
        <span class="text-3" style="font-size: 13px">共 {{ total }} 篇</span>
        <button class="btn btn--outline btn--sm" @click="fetchPosts">
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
        <button class="btn btn--outline btn--sm" style="margin-top: 12px" @click="fetchPosts">重试</button>
      </div>

      <div v-else-if="posts.length === 0" class="empty-state">
        <div class="empty-state__icon"><PhArticle :size="24" /></div>
        <p class="text-3" style="font-size: 14px">暂无帖子</p>
      </div>

      <table v-else class="admin__table">
        <thead>
          <tr>
            <th>标题</th>
            <th>作者</th>
            <th>节点</th>
            <th style="text-align: center">回复</th>
            <th style="text-align: center">赞</th>
            <th>发布时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="post in posts" :key="post.code">
            <td>
              <router-link :to="{ name: 'PostDetail', params: { code: post.code } }" style="font-weight: 500; text-decoration: none; color: inherit">
                {{ truncate(post.title, 40) }}
              </router-link>
            </td>
            <td>
              <router-link :to="{ name: 'UserProfile', params: { username: post.user?.username } }" style="color: var(--text-2); text-decoration: none">
                @{{ post.user?.username || '—' }}
              </router-link>
            </td>
            <td>
              <span v-if="post.node" class="pill pill--lavender">{{ post.node.name }}</span>
              <span v-else class="text-4">—</span>
            </td>
            <td style="text-align: center">{{ post.reply_count || 0 }}</td>
            <td style="text-align: center">{{ post.like_count || 0 }}</td>
            <td class="text-3" style="font-size: 12px">{{ formatDate(post.created_at) }}</td>
            <td>
              <button class="btn btn--sm btn--danger" @click="handleDelete(post)" title="删除">
                <PhTrash :size="13" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="total > pageSize" class="admin__pagination">
        <button class="btn btn--outline btn--sm" :disabled="page <= 1" @click="page--; fetchPosts()">
          <PhCaretLeft :size="14" /> 上一页
        </button>
        <span class="text-3" style="font-size: 13px">{{ page }} / {{ totalPages }}</span>
        <button class="btn btn--outline btn--sm" :disabled="page >= totalPages" @click="page++; fetchPosts()">
          下一页 <PhCaretRight :size="14" />
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useToast } from '../../composables/useToast'
import { adminListPosts, adminDeletePost, listNodes } from '../../api'
import {
  PhArrowClockwise, PhTrash, PhCaretLeft, PhCaretRight,
  PhXCircle, PhCircleNotch, PhArticle
} from '@phosphor-icons/vue'

const posts = ref([])
const nodes = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const error = ref('')
const keyword = ref('')
const nodeFilter = ref('')
const toast = useToast()

const totalPages = computed(() => Math.ceil(total.value / pageSize))

function truncate(str, len) {
  if (!str) return ''
  return str.length > len ? str.slice(0, len) + '...' : str
}

function search() {
  page.value = 1
  fetchPosts()
}

async function fetchNodes() {
  try {
    const res = await listNodes()
    nodes.value = res.data.nodes || res.data || []
  } catch (e) { /* ignore */ }
}

async function fetchPosts() {
  loading.value = true
  error.value = ''
  try {
    const params = { page: page.value, page_size: pageSize }
    if (keyword.value) params.keyword = keyword.value
    if (nodeFilter.value) params.node_id = nodeFilter.value
    const res = await adminListPosts(params)
    posts.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function handleDelete(post) {
  if (!confirm(`确定要删除帖子「${post.title}」吗？`)) return
  try {
    await adminDeletePost(post.code)
    toast.success('已删除')
    if (posts.value.length <= 1 && page.value > 1) page.value--
    fetchPosts()
  } catch (e) {
    toast.error(e.message)
  }
}

function formatDate(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

onMounted(() => { fetchNodes(); fetchPosts() })
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
  .admin__table td:nth-child(3),
  .admin__table th:nth-child(4),
  .admin__table td:nth-child(4),
  .admin__table th:nth-child(5),
  .admin__table td:nth-child(5) { display: none; }
}
</style>