<template>
  <div class="admin-dashboard">
    <div v-if="loading" style="padding: 40px; text-align: center">
      <PhCircleNotch :size="24" class="spin" style="color: var(--text-4)" />
    </div>

    <div v-else-if="error" style="padding: 40px; text-align: center">
      <PhXCircle :size="24" style="color: var(--color-danger)" />
      <p style="margin-top: 8px; font-size: 13px; color: var(--color-danger)">{{ error }}</p>
      <button class="btn btn--outline btn--sm" style="margin-top: 12px" @click="fetchStats">重试</button>
    </div>

    <div v-else class="admin-dashboard__grid">
      <div class="admin-dashboard__card card">
        <div class="admin-dashboard__icon-wrap" style="background: var(--accent-light); color: var(--accent)">
          <PhUsers :size="22" />
        </div>
        <div class="admin-dashboard__stat">
          <div class="admin-dashboard__number">{{ stats.user_count }}</div>
          <div class="admin-dashboard__label">用户总数</div>
        </div>
      </div>
      <div class="admin-dashboard__card card">
        <div class="admin-dashboard__icon-wrap" style="background: #e8f5e9; color: #43a047">
          <PhArticle :size="22" />
        </div>
        <div class="admin-dashboard__stat">
          <div class="admin-dashboard__number">{{ stats.post_count }}</div>
          <div class="admin-dashboard__label">帖子总数</div>
        </div>
      </div>
      <div class="admin-dashboard__card card">
        <div class="admin-dashboard__icon-wrap" style="background: #fff3e0; color: #ef6c00">
          <PhChatCircle :size="22" />
        </div>
        <div class="admin-dashboard__stat">
          <div class="admin-dashboard__number">{{ stats.comment_count }}</div>
          <div class="admin-dashboard__label">评论总数</div>
        </div>
      </div>
      <div class="admin-dashboard__card card">
        <div class="admin-dashboard__icon-wrap" style="background: #f3e5f5; color: #8e24aa">
          <PhFolders :size="22" />
        </div>
        <div class="admin-dashboard__stat">
          <div class="admin-dashboard__number">{{ stats.node_count }}</div>
          <div class="admin-dashboard__label">节点总数</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { adminStats } from '../../api'
import { PhUsers, PhArticle, PhChatCircle, PhFolders, PhCircleNotch, PhXCircle } from '@phosphor-icons/vue'

const stats = reactive({ user_count: 0, post_count: 0, comment_count: 0, node_count: 0 })
const loading = ref(false)
const error = ref('')

async function fetchStats() {
  loading.value = true
  error.value = ''
  try {
    const res = await adminStats()
    Object.assign(stats, res.data)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<style scoped>
.admin-dashboard__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.admin-dashboard__card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 20px;
}

.admin-dashboard__icon-wrap {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-m);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.admin-dashboard__stat {
  min-width: 0;
}

.admin-dashboard__number {
  font-size: 24px;
  font-weight: 700;
  font-family: var(--font-display);
  line-height: 1.2;
}

.admin-dashboard__label {
  font-size: 13px;
  color: var(--text-3);
  margin-top: 2px;
}
</style>