<template>
  <div class="explore fade-up">
    <h1 class="font-display" style="font-size: 26px; margin-bottom: 20px">探索</h1>

    <div class="explore__grid">
      <router-link
        v-for="(node, i) in nodes"
        :key="node.id"
        :to="{ name: 'NodePosts', params: { id: node.id } }"
        class="category-card card"
        :style="{ animationDelay: (i * 40) + 'ms' }"
      >
        <div class="category-card__bar" :style="{ background: node.color || 'var(--accent)' }"></div>
        <div class="category-card__icon" :style="{ background: (node.color || 'var(--accent)') + '14', color: node.color || 'var(--accent)' }">
          <PhStack :size="20" />
        </div>
        <h3 class="font-display category-card__name">{{ node.name }}</h3>
        <p class="text-3 category-card__desc">{{ node.desc }}</p>
        <div class="category-card__stats text-4">
          <span><PhNote :size="12" /> {{ node.post_count || 0 }} 帖子</span>
        </div>
        <span class="btn btn--outline btn--sm category-card__btn">
          <PhArrowRight :size="14" />
          进入
        </span>
      </router-link>
    </div>

    <div v-if="nodes.length === 0 && !loading" class="explore__empty card">
      <div class="explore__empty-icon">
        <PhCompass :size="32" weight="bold" />
      </div>
      <h3 class="font-display" style="font-size: 16px; margin-bottom: 4px">暂无节点</h3>
      <p class="text-3" style="font-size: 13px">节点正在创建中，请稍后再来</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { listNodes } from '../api'
import {
  PhStack, PhNote, PhArrowRight, PhCompass
} from '@phosphor-icons/vue'

const nodes = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await listNodes()
    nodes.value = res.data?.nodes || []
  } catch (e) {
    console.error('加载节点失败:', e)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.explore__grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.category-card {
  padding: 20px;
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  transition: box-shadow var(--duration-medium-2) var(--ease-standard);
  text-decoration: none;
  color: inherit;
}

.category-card:hover {
  box-shadow: var(--shadow-2);
}

.category-card__bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  border-radius: var(--radius-m) 0 0 var(--radius-m);
  transition: width var(--duration-medium-2) var(--ease-standard);
}

.category-card:hover .category-card__bar {
  width: 4px;
}

.category-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-m);
}

.category-card__name {
  font-size: 15px;
  font-weight: 600;
}

.category-card__desc {
  font-size: 13px;
  line-height: 1.5;
}

.category-card__stats {
  display: flex;
  gap: 12px;
  font-size: 12px;
  margin-top: 4px;
}

.category-card__stats span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.category-card__btn {
  align-self: flex-start;
  margin-top: 4px;
}

.explore__empty {
  padding: 40px;
  text-align: center;
  margin-top: 24px;
}

.explore__empty-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-l);
  background: var(--bg-muted);
  color: var(--text-4);
  margin-bottom: 12px;
}

@media (max-width: 960px) {
  .explore__grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 600px) {
  .explore__grid {
    grid-template-columns: 1fr;
  }
}
</style>
