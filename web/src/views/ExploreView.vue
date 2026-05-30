<template>
  <div class="explore fade-up">
    <h1 class="font-display" style="font-size: 26px; margin-bottom: 20px">探索</h1>

    <div class="explore__grid">
      <div
        v-for="(cat, i) in categories"
        :key="cat.name"
        class="category-card card"
        :style="{ animationDelay: (i * 40) + 'ms' }"
      >
        <div class="category-card__bar" :style="{ background: cat.color }"></div>
        <div class="category-card__icon" :style="{ background: cat.color + '14', color: cat.color }">
          <component :is="cat.icon" :size="20" />
        </div>
        <h3 class="font-display category-card__name">{{ cat.name }}</h3>
        <p class="text-3 category-card__desc">{{ cat.desc }}</p>
        <div class="category-card__stats text-4">
          <span><PhNote :size="12" /> 开发中</span>
        </div>
        <button class="btn btn--outline btn--sm category-card__btn">
          <PhArrowRight :size="14" />
          进入
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import {
  PhCode, PhRocketLaunch, PhQuestion, PhGift,
  PhMegaphone, PhChatsCircle, PhBriefcase, PhBug,
  PhNote, PhArrowRight
} from '@phosphor-icons/vue'

const categories = [
  { name: '技术讨论', desc: '编程语言、框架、工具的深度技术讨论', color: '#9b8ec4', icon: PhCode },
  { name: '项目展示', desc: '分享你的开源项目和作品', color: '#6db89a', icon: PhRocketLaunch },
  { name: '新手求助', desc: '遇到问题？在这里寻求帮助', color: '#7ba4d4', icon: PhQuestion },
  { name: '资源分享', desc: '优质教程、工具、资源推荐', color: '#d4a07a', icon: PhGift },
  { name: '公告/官方', desc: '社区公告和官方信息', color: '#d4b85a', icon: PhMegaphone },
  { name: '闲聊灌水', desc: '轻松闲聊，分享日常', color: '#c47a99', icon: PhChatsCircle },
  { name: '求职招聘', desc: '工作机会和求职信息', color: '#8bb8a8', icon: PhBriefcase },
  { name: 'Bug 反馈', desc: '反馈社区平台的 Bug 和建议', color: '#c4987a', icon: PhBug },
]
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
