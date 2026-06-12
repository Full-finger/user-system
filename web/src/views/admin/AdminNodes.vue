<template>
  <div class="admin-nodes">
    <div class="admin__toolbar card">
      <span class="text-3" style="font-size: 13px">共 {{ nodes.length }} 个节点</span>
      <div style="display: flex; gap: 8px">
        <button class="btn btn--outline btn--sm" @click="fetchNodes">
          <PhArrowClockwise :size="14" /> 刷新
        </button>
        <button class="btn btn--primary btn--sm" @click="openCreateModal">
          <PhPlus :size="14" /> 新建节点
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
        <button class="btn btn--outline btn--sm" style="margin-top: 12px" @click="fetchNodes">重试</button>
      </div>

      <div v-else-if="nodes.length === 0" class="empty-state">
        <div class="empty-state__icon"><PhFolders :size="24" /></div>
        <p class="text-3" style="font-size: 14px">暂无节点</p>
      </div>

      <table v-else class="admin__table">
        <thead>
          <tr>
            <th>节点</th>
            <th>Slug</th>
            <th>描述</th>
            <th style="text-align: center">帖子数</th>
            <th style="text-align: center">排序</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="node in nodes" :key="node.id">
            <td>
              <div style="display: flex; align-items: center; gap: 8px">
                <span v-if="node.color" class="node-dot" :style="{ background: node.color }"></span>
                <span style="font-weight: 500">{{ node.name }}</span>
              </div>
            </td>
            <td class="font-mono text-3" style="font-size: 12px">{{ node.slug }}</td>
            <td class="text-3" style="font-size: 13px; max-width: 200px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">
              {{ node.desc || '—' }}
            </td>
            <td style="text-align: center">{{ node.post_count || 0 }}</td>
            <td style="text-align: center">{{ node.sort_order || 0 }}</td>
            <td>
              <div style="display: flex; gap: 4px">
                <button class="btn btn--outline btn--sm" @click="openEditModal(node)" title="编辑">
                  <PhPencil :size="13" />
                </button>
                <button class="btn btn--sm btn--danger" @click="handleDelete(node)" title="删除">
                  <PhTrash :size="13" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 创建/编辑弹窗 -->
    <Teleport to="body">
      <Transition name="fade">
        <div v-if="modal.show" class="modal-overlay" @click.self="modal.show = false">
          <div class="modal-panel card">
            <div class="modal-panel__header">
              <h3 class="font-display" style="font-size: 18px">{{ modal.editing ? '编辑节点' : '新建节点' }}</h3>
              <button class="modal-panel__close" @click="modal.show = false">
                <PhX :size="18" />
              </button>
            </div>
            <div class="modal-panel__body">
              <div class="admin__field">
                <label class="admin__label">名称 *</label>
                <input v-model="form.name" type="text" class="input" placeholder="节点名称" />
              </div>
              <div class="admin__field">
                <label class="admin__label">Slug *</label>
                <input v-model="form.slug" type="text" class="input" placeholder="url-friendly 标识" :disabled="modal.editing" />
              </div>
              <div class="admin__field">
                <label class="admin__label">描述</label>
                <textarea v-model="form.desc" class="input" rows="2" placeholder="节点描述（可选）"></textarea>
              </div>
              <div style="display: flex; gap: 12px">
                <div class="admin__field" style="flex: 1">
                  <label class="admin__label">颜色</label>
                  <input v-model="form.color" type="color" class="input" style="height: 38px; padding: 4px" />
                </div>
                <div class="admin__field" style="flex: 1">
                  <label class="admin__label">图标</label>
                  <input v-model="form.icon" type="text" class="input" placeholder="图标名（可选）" />
                </div>
                <div class="admin__field" style="width: 100px">
                  <label class="admin__label">排序</label>
                  <input v-model.number="form.sort_order" type="number" class="input" placeholder="0" />
                </div>
              </div>
              <div v-if="formError" class="admin__error">
                <PhXCircle :size="14" />
                {{ formError }}
              </div>
            </div>
            <div class="modal-panel__footer">
              <button class="btn btn--outline" @click="modal.show = false">取消</button>
              <button class="btn btn--primary" @click="handleSubmit" :disabled="formLoading">
                <PhCircleNotch v-if="formLoading" :size="16" class="spin" />
                {{ formLoading ? '保存中...' : '保存' }}
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
import { useToast } from '../../composables/useToast'
import { listNodes, createNode, updateNode, deleteNode } from '../../api'
import {
  PhArrowClockwise, PhPlus, PhPencil, PhTrash,
  PhX, PhXCircle, PhCircleNotch, PhFolders
} from '@phosphor-icons/vue'

const nodes = ref([])
const loading = ref(false)
const error = ref('')
const toast = useToast()

const modal = reactive({ show: false, editing: false, nodeId: null })
const form = reactive({ name: '', slug: '', desc: '', color: '#6c5ce7', icon: '', sort_order: 0 })
const formError = ref('')
const formLoading = ref(false)

async function fetchNodes() {
  loading.value = true
  error.value = ''
  try {
    const res = await listNodes()
    nodes.value = res.data.nodes || res.data || []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  modal.editing = false
  modal.nodeId = null
  Object.assign(form, { name: '', slug: '', desc: '', color: '#6c5ce7', icon: '', sort_order: 0 })
  formError.value = ''
  modal.show = true
}

function openEditModal(node) {
  modal.editing = true
  modal.nodeId = node.id
  Object.assign(form, {
    name: node.name,
    slug: node.slug,
    desc: node.desc || '',
    color: node.color || '#6c5ce7',
    icon: node.icon || '',
    sort_order: node.sort_order || 0
  })
  formError.value = ''
  modal.show = true
}

async function handleSubmit() {
  formError.value = ''
  if (!form.name || !form.slug) {
    formError.value = '名称和 Slug 不能为空'
    return
  }

  formLoading.value = true
  try {
    const data = { ...form }
    if (modal.editing) {
      await updateNode(modal.nodeId, data)
      toast.success('已保存')
    } else {
      await createNode(data)
      toast.success('已创建')
    }
    modal.show = false
    fetchNodes()
  } catch (e) {
    formError.value = e.message
  } finally {
    formLoading.value = false
  }
}

async function handleDelete(node) {
  if (!confirm(`确定要删除节点「${node.name}」吗？`)) return
  try {
    await deleteNode(node.id)
    toast.success('已删除')
    fetchNodes()
  } catch (e) {
    toast.error(e.message)
  }
}

onMounted(fetchNodes)
</script>

<style scoped>
.admin__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
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

.admin__field { display: flex; flex-direction: column; gap: 6px; }

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

.node-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

@media (max-width: 600px) {
  .admin__table th:nth-child(3),
  .admin__table td:nth-child(3),
  .admin__table th:nth-child(4),
  .admin__table td:nth-child(4) { display: none; }
}
</style>