<template>
  <div class="admin fade-up">
    <h1 class="font-display" style="font-size: 26px; margin-bottom: 4px">管理后台</h1>

    <!-- Tab 导航 -->
    <div class="admin__tabs card">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="admin__tab"
        :class="{ 'admin__tab--active': activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        <component :is="tab.icon" :size="16" />
        <span>{{ tab.label }}</span>
      </button>
    </div>

    <!-- Tab 内容 -->
    <div class="admin__content">
      <AdminDashboard v-if="activeTab === 'dashboard'" key="dashboard" />
      <AdminUsers v-else-if="activeTab === 'users'" key="users" />
      <AdminPosts v-else-if="activeTab === 'posts'" key="posts" />
      <AdminComments v-else-if="activeTab === 'comments'" key="comments" />
      <AdminNodes v-else-if="activeTab === 'nodes'" key="nodes" />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import {
  PhChartBar,
  PhUsers,
  PhArticle,
  PhChatCircle,
  PhFolders
} from '@phosphor-icons/vue'
import AdminDashboard from './admin/AdminDashboard.vue'
import AdminUsers from './admin/AdminUsers.vue'
import AdminPosts from './admin/AdminPosts.vue'
import AdminComments from './admin/AdminComments.vue'
import AdminNodes from './admin/AdminNodes.vue'

const activeTab = ref('dashboard')

const tabs = [
  { key: 'dashboard', label: '概览', icon: PhChartBar },
  { key: 'users', label: '用户管理', icon: PhUsers },
  { key: 'posts', label: '帖子管理', icon: PhArticle },
  { key: 'comments', label: '评论管理', icon: PhChatCircle },
  { key: 'nodes', label: '节点管理', icon: PhFolders }
]
</script>

<style scoped>
.admin {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.admin__tabs {
  display: flex;
  gap: 4px;
  padding: 6px;
  overflow-x: auto;
}

.admin__tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--radius-m);
  border: none;
  background: transparent;
  font-family: var(--font-body);
  font-size: 14px;
  color: var(--text-3);
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--duration-medium-1) var(--ease-standard);
}

.admin__tab:hover {
  background: var(--state-hover);
  color: var(--text-1);
}

.admin__tab--active {
  background: var(--accent-light);
  color: var(--accent);
  font-weight: 500;
}

@media (max-width: 600px) {
  .admin__tab span {
    display: none;
  }
  .admin__tab {
    padding: 8px 12px;
  }
}
</style>
