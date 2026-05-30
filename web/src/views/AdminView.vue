<template>
  <div class="admin fade-up">
    <h1 class="font-display" style="font-size: 26px; margin-bottom: 4px">管理后台</h1>
    <p class="text-3" style="font-size: 14px; margin-bottom: 20px">用户管理</p>

    <!-- 顶部操作栏 -->
    <div class="admin__toolbar card">
      <div class="admin__toolbar-left">
        <span class="text-3" style="font-size: 13px">
          共 {{ total }} 个用户
        </span>
      </div>
      <button class="btn btn--outline btn--sm" @click="fetchUsers">
        <PhArrowClockwise :size="14" />
        刷新
      </button>
    </div>

    <!-- 用户表格 -->
    <div class="card admin__table-wrap">
      <div v-if="loading" style="padding: 40px; text-align: center">
        <PhCircleNotch :size="24" class="spin" style="color: var(--text-4)" />
        <p class="text-3" style="margin-top: 8px; font-size: 13px">加载中...</p>
      </div>

      <div v-else-if="error" style="padding: 40px; text-align: center">
        <PhXCircle :size="24" style="color: #c47878" />
        <p style="margin-top: 8px; font-size: 13px; color: #c47878">{{ error }}</p>
        <button class="btn btn--outline btn--sm" style="margin-top: 12px" @click="fetchUsers">重试</button>
      </div>

      <table v-else class="admin__table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户名</th>
            <th>邮箱</th>
            <th>角色</th>
            <th>注册时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td class="text-3 font-mono" style="font-size: 12px">{{ user.id }}</td>
            <td>
              <div style="display: flex; align-items: center; gap: 8px">
                <div class="avatar avatar--sm" style="width: 28px; height: 28px; font-size: 11px">
                  {{ user.username[0].toUpperCase() }}
                </div>
                <span style="font-weight: 500">{{ user.username }}</span>
              </div>
            </td>
            <td class="text-3" style="font-size: 13px">{{ user.email || '—' }}</td>
            <td>
              <span class="pill" :class="user.role === 'admin' ? 'pill--accent' : 'pill--lavender'">
                {{ user.role }}
              </span>
            </td>
            <td class="text-3" style="font-size: 13px">{{ formatDate(user.created_at) }}</td>
            <td>
              <div style="display: flex; gap: 4px">
                <button
                  class="btn btn--outline btn--sm"
                  @click="openEditModal(user)"
                  title="编辑"
                >
                  <PhPencil :size="13" />
                </button>
                <button
                  class="btn btn--sm"
                  :class="user.role === 'admin' ? '' : 'btn--danger'"
                  :disabled="user.role === 'admin'"
                  :title="user.role === 'admin' ? '不能删除管理员' : '删除'"
                  @click="handleDelete(user)"
                >
                  <PhTrash :size="13" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页 -->
      <div v-if="total > pageSize" class="admin__pagination">
        <button
          class="btn btn--outline btn--sm"
          :disabled="page <= 1"
          @click="page--; fetchUsers()"
        >
          <PhCaretLeft :size="14" /> 上一页
        </button>
        <span class="text-3" style="font-size: 13px">{{ page }} / {{ totalPages }}</span>
        <button
          class="btn btn--outline btn--sm"
          :disabled="page >= totalPages"
          @click="page++; fetchUsers()"
        >
          下一页 <PhCaretRight :size="14" />
        </button>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <Transition name="fade">
      <div v-if="editModal.show" class="modal-overlay" @click.self="editModal.show = false">
        <div class="modal-panel card">
          <div class="modal-panel__header">
            <h3 class="font-display" style="font-size: 18px">编辑用户</h3>
            <button class="topbar__icon-btn" @click="editModal.show = false">
              <PhX :size="18" />
            </button>
          </div>
          <div class="modal-panel__body">
            <div class="auth-form__group">
              <label class="auth-form__label">用户名</label>
              <input class="input" :value="editModal.user?.username" disabled />
            </div>
            <div class="auth-form__group">
              <label class="auth-form__label">新密码（留空则不修改）</label>
              <input v-model="editForm.password" type="password" class="input" placeholder="至少 6 位" />
            </div>
            <div class="auth-form__group">
              <label class="auth-form__label">角色</label>
              <select v-model="editForm.role" class="input" style="cursor: pointer">
                <option value="user">user</option>
                <option value="admin">admin</option>
              </select>
            </div>
            <div v-if="editError" class="auth-form__error">
              <PhXCircle :size="14" />
              {{ editError }}
            </div>
          </div>
          <div class="modal-panel__footer">
            <button class="btn btn--outline" @click="editModal.show = false">取消</button>
            <button class="btn btn--primary" @click="handleEdit" :disabled="editLoading">
              <PhCircleNotch v-if="editLoading" :size="16" class="spin" />
              {{ editLoading ? '保存中...' : '保存' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useToast } from '../composables/useToast'
import { listUsers, updateUser, deleteUser } from '../api'
import {
  PhArrowClockwise, PhPencil, PhTrash, PhCaretLeft, PhCaretRight,
  PhX, PhXCircle, PhCircleNotch
} from '@phosphor-icons/vue'

const users = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const error = ref('')

const editModal = reactive({ show: false, user: null })
const editForm = reactive({ password: '', role: 'user' })
const editError = ref('')
const editLoading = ref(false)
const toast = useToast()

const totalPages = computed(() => Math.ceil(total.value / pageSize))

async function fetchUsers() {
  loading.value = true
  error.value = ''
  try {
    const res = await listUsers({ page: page.value, page_size: pageSize })
    users.value = res.data.list || []
    total.value = res.data.total || 0
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function openEditModal(user) {
  editModal.user = user
  editForm.password = ''
  editForm.role = user.role
  editError.value = ''
  editModal.show = true
}

async function handleEdit() {
  editError.value = ''
  const data = {}
  if (editForm.password) {
    if (editForm.password.length < 6) {
      editError.value = '密码至少 6 位'
      return
    }
    data.password = editForm.password
  }
  if (editForm.role !== editModal.user.role) {
    data.role = editForm.role
  }
  if (Object.keys(data).length === 0) {
    editModal.show = false
    return
  }

  editLoading.value = true
  try {
    await updateUser(editModal.user.id, data)
    editModal.show = false
    fetchUsers()
  } catch (e) {
    editError.value = e.message
  } finally {
    editLoading.value = false
  }
}

async function handleDelete(user) {
  if (!confirm(`确定要删除用户 "${user.username}" 吗？（软删除）`)) return
  try {
    await deleteUser(user.id)
    toast.success('已删除')
    fetchUsers()
  } catch (e) {
    toast.error(e.message)
  }
}

function formatDate(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

onMounted(fetchUsers)
</script>

<style scoped>
.admin {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.admin__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
}

.admin__toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.admin__table-wrap {
  overflow-x: auto;
}

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
  position: relative;
  overflow: hidden;
  transition: background var(--duration-medium-1) var(--ease-standard);
}

.admin__table tbody tr:hover {
  background: var(--state-hover);
}

.admin__table tbody tr:last-child td {
  border-bottom: none;
}

.admin__pagination {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  padding: 16px;
  border-top: 1px solid var(--border);
}

/* ---- Modal ---- */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 150;
}

[data-theme="dark"] .modal-overlay {
  background: rgba(0, 0, 0, 0.5);
}

.modal-panel {
  width: 100%;
  max-width: 440px;
  background: var(--bg-card);
  box-shadow: var(--shadow-4);
  backdrop-filter: blur(24px) saturate(1.4);
}

[data-theme="dark"] .modal-panel {
  background: rgba(26, 23, 30, 0.92);
}

.modal-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 24px 0;
}

.modal-panel__body {
  padding: 20px 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-panel__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 16px 24px 20px;
  border-top: 1px solid var(--border);
}

.auth-form__group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.auth-form__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-2);
}

.auth-form__error {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: #c47878;
}

/* topbar__icon-btn inherited from layout */
.topbar__icon-btn {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-s);
  color: var(--text-2);
  background: none;
  border: none;
  cursor: pointer;
  overflow: hidden;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spin {
  animation: spin 0.8s linear infinite;
}

@media (max-width: 600px) {
  .admin__table th:nth-child(4),
  .admin__table td:nth-child(4),
  .admin__table th:nth-child(5),
  .admin__table td:nth-child(5) {
    display: none;
  }
}
</style>
