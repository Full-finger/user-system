<template>
  <div class="home">
    <!-- 发帖按钮区域 -->
    <div class="home__header fade-up">
      <h1 class="font-display home__title">发现</h1>
      <button class="btn btn--primary">
        <PhPencilSimpleLine :size="16" />
        发帖
      </button>
    </div>

    <!-- 版块标签过滤 -->
    <div class="home__tabs fade-up" style="animation-delay: 40ms">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        class="home__tab"
        :class="{ 'home__tab--active': activeTab === tab.key }"
        @click="activeTab = tab.key"
      >
        <component :is="tab.icon" :size="14" />
        {{ tab.label }}
      </button>
    </div>

    <!-- 开发中提示 -->
    <div class="home__dev-hint fade-up" style="animation-delay: 80ms">
      <div class="card home__dev-card">
        <div class="home__dev-icon">
          <PhWrench :size="28" weight="bold" />
        </div>
        <h2 class="font-display" style="font-size: 18px; margin-bottom: 6px">帖子功能开发中</h2>
        <p class="text-3" style="font-size: 14px; line-height: 1.6">
          帖子、话题、投票、评论等社区核心功能正在紧锣密鼓地开发中。<br />
          目前可以使用 <router-link to="/profile">个人中心</router-link> 和
          <router-link to="/admin" v-if="auth.isAdmin">管理后台</router-link>
          <template v-else>登录注册</template>
          功能。
        </p>
      </div>
    </div>

    <!-- 示例帖子卡片（静态演示设计系统） -->
    <div class="home__posts">
      <div
        v-for="(post, i) in samplePosts"
        :key="i"
        class="post-card card fade-up"
        :style="{ animationDelay: (120 + i * 40) + 'ms' }"
      >
        <div class="post-card__bar" :style="{ background: post.color }"></div>

        <div class="post-card__vote">
          <button class="post-card__vote-btn">
            <PhThumbsUp :size="16" />
          </button>
          <span class="font-display" style="font-size: 14px; font-weight: 600">{{ post.votes }}</span>
          <button class="post-card__vote-btn">
            <PhThumbsDown :size="16" />
          </button>
        </div>

        <div class="post-card__content">
          <div class="post-card__top">
            <h3 class="post-card__title font-display">{{ post.title }}</h3>
            <span class="pill" :style="{ background: post.color + '18', color: post.color }">
              {{ post.category }}
            </span>
          </div>
          <p class="post-card__desc text-3">{{ post.desc }}</p>
          <div class="post-card__meta text-4">
            <span class="post-card__author">
              <span class="post-card__online-dot" :style="{ background: post.onlineColor }"></span>
              {{ post.author }}
            </span>
            <span class="pill pill--accent" style="font-size: 11px">{{ post.level }}</span>
            <span><PhClock :size="12" style="vertical-align: -1px" /> {{ post.time }}</span>
            <span><PhChatCircle :size="12" style="vertical-align: -1px" /> {{ post.replies }}</span>
            <span><PhEye :size="12" style="vertical-align: -1px" /> {{ post.views }}</span>
          </div>
          <div class="post-card__tags">
            <span v-for="tag in post.tags" :key="tag" class="pill pill--lavender">{{ tag }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import {
  PhPencilSimpleLine, PhWrench, PhThumbsUp, PhThumbsDown,
  PhClock, PhChatCircle, PhEye,
  PhHouse, PhCompass, PhStack, PhFireSimple, PhStar
} from '@phosphor-icons/vue'

const auth = useAuthStore()

const activeTab = ref('all')
const tabs = [
  { key: 'all', label: '全部', icon: PhHouse },
  { key: 'tech', label: '技术', icon: PhStack },
  { key: 'hot', label: '热门', icon: PhFireSimple },
  { key: 'essence', label: '精华', icon: PhStar },
  { key: 'explore', label: '探索', icon: PhCompass },
]

const samplePosts = [
  {
    title: 'Building a Custom ORM in Rust',
    desc: 'An in-depth guide to building type-safe database abstractions in Rust...',
    category: '技术讨论',
    color: '#9b8ec4',
    votes: 24,
    author: 'rustacean',
    level: 'Lv.3',
    time: '2h ago',
    replies: 24,
    views: '1.2K',
    tags: ['Rust', 'Database', 'Tutorial'],
    onlineColor: '#6db89a',
  },
  {
    title: 'Vue 3 Composition API 最佳实践总结',
    desc: '分享在实际项目中使用 Composition API 的一些心得和踩坑记录...',
    category: '项目展示',
    color: '#6db89a',
    votes: 18,
    author: 'vuefan',
    level: 'Lv.2',
    time: '4h ago',
    replies: 12,
    views: '860',
    tags: ['Vue', 'Frontend'],
    onlineColor: '#a89cb5',
  },
  {
    title: '如何设计一个可扩展的微服务架构？',
    desc: '讨论在业务快速增长阶段，如何合理拆分服务边界...',
    category: '新手求助',
    color: '#7ba4d4',
    votes: 31,
    author: 'architect_wang',
    level: 'Lv.4',
    time: '6h ago',
    replies: 45,
    views: '2.3K',
    tags: ['Microservice', 'Architecture'],
    onlineColor: '#6db89a',
  },
]
</script>

<style scoped>
.home__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.home__title {
  font-size: 26px;
}

.home__tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border);
  overflow-x: auto;
}

.home__tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: none;
  background: none;
  color: var(--text-3);
  font-family: var(--font-body);
  font-size: 13px;
  font-weight: 500;
  border-radius: var(--radius-full);
  cursor: pointer;
  white-space: nowrap;
  position: relative;
  overflow: hidden;
  transition: color var(--duration-medium-1) var(--ease-standard);
}

.home__tab::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: transparent;
  transition: background var(--duration-medium-1) var(--ease-standard);
  pointer-events: none;
}

.home__tab:hover {
  color: var(--text-1);
}

.home__tab:hover::after {
  background: var(--state-hover);
}

.home__tab--active {
  color: var(--accent);
  font-weight: 600;
}

.home__tab--active::after {
  background: var(--accent-light);
}

.home__dev-card {
  padding: 32px;
  text-align: center;
  margin-bottom: 20px;
}

.home__dev-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  border-radius: var(--radius-l);
  background: var(--accent-light);
  color: var(--accent);
  margin-bottom: 16px;
}

/* ---- Post Card ---- */
.post-card {
  display: flex;
  gap: 16px;
  padding: 16px;
  margin-bottom: 8px;
  position: relative;
  transition: box-shadow var(--duration-medium-2) var(--ease-standard);
}

.post-card:hover {
  box-shadow: var(--shadow-2);
}

.post-card__bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 0;
  border-radius: var(--radius-m) 0 0 var(--radius-m);
  transition: width var(--duration-medium-2) var(--ease-standard);
}

.post-card:hover .post-card__bar {
  width: 3px;
}

.post-card__vote {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  padding: 4px 0;
  min-width: 40px;
  color: var(--text-4);
}

.post-card__vote-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: none;
  color: var(--text-4);
  cursor: pointer;
  border-radius: var(--radius-s);
  position: relative;
  overflow: hidden;
  transition: color var(--duration-medium-1) var(--ease-standard);
}

.post-card__vote-btn::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: transparent;
  transition: background var(--duration-medium-1) var(--ease-standard);
  pointer-events: none;
}

.post-card__vote-btn:hover {
  color: var(--accent);
}

.post-card__vote-btn:hover::after {
  background: var(--state-hover);
}

.post-card__content {
  flex: 1;
  min-width: 0;
}

.post-card__top {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.post-card__title {
  font-size: 16px;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.post-card__desc {
  font-size: 13px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 8px;
}

.post-card__meta {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 12px;
  margin-bottom: 8px;
}

.post-card__author {
  display: flex;
  align-items: center;
  gap: 5px;
  font-weight: 500;
}

.post-card__online-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.post-card__tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

@media (max-width: 600px) {
  .post-card {
    flex-direction: column;
    gap: 12px;
  }

  .post-card__vote {
    flex-direction: row;
    min-width: auto;
  }

  .post-card__title {
    font-size: 15px;
  }
}
</style>
