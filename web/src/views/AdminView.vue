<template>
  <div class="admin fade-up">
    <h1 class="font-display" style="font-size: 26px; margin-bottom: 4px">管理后台</h1>
    <p class="text-3" style="font-size: 14px; margin-bottom: 20px">用户管理</p>

    <!-- 顶部操作栏 -->
    <div class="admin__toolbar card">
      <span class="text-3" style="font-size: 13px">共 {{ total }} 个用户</span>
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
        <PhXCircle :size="24" style="color: var(--color-danger)" />
        <p style="margin-top: 8px; font-size: 13px; color: var(--color-danger)">{{ error }}</p>
        <button class="btn btn--outline btn--sm" style="margin-top: 12px" @click="fetchUsers">重试</button>
      </div>

      <div v-else-if="users.length === 0" class="empty-state">
        <div class="empty-state__icon">
          <PhUsers :size="24" />
        </div>
        <p class="text-3" style="font-size: 14px">暂无用户</p>
      </div>

      <table v-else class="admin__table">
        <thead>
          <tr>
            <th>用户</th>
            <th>角色</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <td>
              <div style="display: flex; align-items: center; gap: 10px">
                <div class="avatar avatar--sm">
                  {{ (user.nickname || user.username)[0].toUpperCase() }}
                </div>
                <div style="min-width: 0">
                  <div style="font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">
                    {{ user.nickname || user.username }}
                  </div>
                  <div class="text-4" style="font-size: 11px">@{{ user.username }}</div>
                </div>
              </div>
            </td>
            <td>
              <span class="pill" :class="ADMIN_ROLES.includes(user.role) ? 'pill--accent' : 'pill--lavender'">
                {{ roleLabel(user.role) }}
              </span>
            </td>
            <td>
              <div v-if="user.id === authStore.user?.id" class="text-4" style="font-size: 12px">当前用户</div>
              <div v-else style="display: flex; gap: 4px">
                <button class="btn btn--outline btn--sm" @click="openEditModal(user)" title="编辑">
                  <PhPencil :size="13" />
                </button>
                <button
                  class="btn btn--sm btn--danger"
                  :disabled="ADMIN_ROLES.includes(user.role)"
                  :title="ADMIN_ROLES.includes(user.role) ? '不能删除管理员' : '删除'"
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
        <button class="btn btn--outline btn--sm" :disabled="page <= 1" @click="page--; fetchUsers()">
          <PhCaretLeft :size="14" /> 上一页
        </button>
        <span class="text-3" style="font-size: 13px">{{ page }} / {{ totalPages }}</span>
        <button class="btn btn--outline btn--sm" :disabled="page >= totalPages" @click="page++; fetchUsers()">
          下一页 <PhCaretRight :size="14" />
        </button>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="editModal.show" class="modal-overlay" @click.self="editModal.show = false">
          <div class="modal-panel card">
            <div class="modal-panel__header">
              <h3 class="font-display" style="font-size: 18px">编辑用户</h3>
              <button class="modal-panel__close" @click="editModal.show = false">
                <PhX :size="18" />
              </button>
            </div>
            <div class="modal-panel__body">
              <!-- 只读信息 -->
              <div class="admin__readonly">
                <div class="admin__readonly-row">
                  <span class="text-4">ID</span>
                  <span class="text-3 font-mono" style="font-size: 12px">{{ editModal.user?.id }}</span>
                </div>
                <div class="admin__readonly-row">
                  <span class="text-4">用户名</span>
                  <span style="font-size: 13px">@{{ editModal.user?.username }}</span>
                </div>
                <div v-if="editModal.user?.email" class="admin__readonly-row">
                  <span class="text-4">邮箱</span>
                  <span class="text-3" style="font-size: 13px">{{ editModal.user.email }}</span>
                </div>
                <div class="admin__readonly-row">
                  <span class="text-4">注册时间</span>
                  <span class="text-3" style="font-size: 13px">{{ formatDate(editModal.user?.created_at) }}</span>
                </div>
              </div>

              <!-- 可编辑字段 -->
              <div class="admin__field">
                <label class="admin__label">昵称</label>
                <input v-model="editForm.nickname" type="text" class="input" placeholder="留空则不修改" />
              </div>
              <div class="admin__field">
                <label class="admin__label">角色</label>
                <div class="admin__role-options">
                  <button
                    v-for="r in assignableRoles"
                    :key="r"
                    class="admin__role-option"
                    :class="{ 'admin__role-option--active': editForm.role === r }"
                    @click="editForm.role = r"
                    type="button"
                  >
                    {{ roleLabel(r) }}
                  </button>
                </div>
              </div>

              <div v-if="editError" class="admin__error">
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
    </Teleport>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useToast } from '../composables/useToast'
import { roleLabel, ADMIN_ROLES, ASSIGNABLE_ROLES } from '../utils/role'
import { useAuthStore } from '../stores/auth'
import { listUsers, updateUser, deleteUser } from '../api'
import {
  PhArrowClockwise, PhPencil, PhTrash, PhCaretLeft, PhCaretRight,
  PhX, PhXCircle, PhCircleNotch, PhUsers
} from '@phosphor-icons/vue'

const users = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const loading = ref(false)
const error = ref('')

const editModal = reactive({ show: false, user: null })
const editForm = reactive({ nickname: '', role: 'user' })
const editError = ref('')
const editLoading = ref(false)
const toast = useToast()
const authStore = useAuthStore()

const totalPages = computed(() => Math.ceil(total.value / pageSize))
const assignableRoles = computed(() => ASSIGNABLE_ROLES[authStore.user?.role] || [])

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
  if (user.id === authStore.user?.id) return
  editModal.user = user
  editForm.nickname = ''
  editForm.role = user.role
  editError.value = ''
  editModal.show = true
}

async function handleEdit() {
  editError.value = ''
  const data = {}
  if (editForm.nickname) {
    data.nickname = editForm.nickname
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
    toast.success('已保存')
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
    // 删完后当前页可能为空，回退一页
    if (users.value.length <= 1 && page.value > 1) {
      page.value--
    }
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
  --color-danger: #c47878;
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

/* ---- Edit Modal: read-only info ---- */
.admin__readonly {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px;
  border-radius: var(--radius-m);
  background: var(--bg-muted);
}

.admin__readonly-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

/* ---- Edit Modal: fields ---- */
.admin__field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.admin__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-2);
}

.admin__error {
  --color-danger: #c47878;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--color-danger);
}

/* ---- Role Option Buttons ---- */
.admin__role-options {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.admin__role-option {
  padding: 6px 14px;
  border-radius: var(--radius-full);
  border: 1.5px solid var(--border);
  background: var(--bg-muted);
  font-family: var(--font-body);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--duration-medium-1) var(--ease-standard);
}

.admin__role-option:hover {
  border-color: var(--border-hover);
  background: var(--state-hover);
}

.admin__role-option--active {
  border-color: var(--accent);
  background: var(--accent-light);
  color: var(--accent);
  font-weight: 500;
}

@media (max-width: 600px) {
  .admin__table th:nth-child(2),
  .admin__table td:nth-child(2) {
    display: none;
  }
}
</style>
