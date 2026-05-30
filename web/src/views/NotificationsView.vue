<template>
  <div class="notifications fade-up">
    <h1 class="font-display" style="font-size: 26px; margin-bottom: 20px">通知中心</h1>

    <div class="card placeholder-card">
      <div class="placeholder-card__icon">
        <PhBell :size="40" weight="bold" />
      </div>
      <h2 class="font-display placeholder-card__title">通知功能开发中</h2>
      <p class="text-3 placeholder-card__desc">
        回复通知、点赞提醒、@提及、系统公告等功能正在开发中，敬请期待。
      </p>
    </div>

    <!-- 静态演示通知列表样式 -->
    <div class="notifications__list">
      <div class="notif-item card" v-for="(n, i) in sampleNotifs" :key="i">
        <div class="notif-item__bar" v-if="n.unread" style="background: var(--accent)"></div>
        <div class="notif-item__type-icon" :style="{ color: n.iconColor }">
          <component :is="n.icon" :size="16" />
        </div>
        <div class="notif-item__content">
          <p class="notif-item__text" :class="{ 'text-1': n.unread, 'text-3': !n.unread }">
            {{ n.text }}
          </p>
          <span class="notif-item__time text-4">
            <PhClock :size="10" style="vertical-align: -1px" /> {{ n.time }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { PhBell, PhChatCircleText, PhHeart, PhStar, PhClock } from '@phosphor-icons/vue'

const sampleNotifs = [
  { text: '@rustacean 回复了你的帖子 "Building a Custom ORM in Rust"', time: '2 分钟前', icon: PhChatCircleText, iconColor: '#7ba4d4', unread: true },
  { text: '@vuefan 给你的回复点了赞', time: '1 小时前', icon: PhHeart, iconColor: '#c47a99', unread: true },
  { text: '你的帖子被标记为精华', time: '3 小时前', icon: PhStar, iconColor: '#d4b85a', unread: false },
]
</script>

<style scoped>
.notifications {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.placeholder-card {
  padding: 40px 32px;
  text-align: center;
}

.placeholder-card__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  border-radius: var(--radius-l);
  background: var(--accent-light);
  color: var(--accent);
  margin-bottom: 16px;
}

.placeholder-card__title {
  font-size: 18px;
  margin-bottom: 8px;
}

.placeholder-card__desc {
  font-size: 14px;
  line-height: 1.7;
  max-width: 360px;
  margin: 0 auto;
}

.notifications__list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.notif-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  position: relative;
  transition: box-shadow var(--duration-medium-2) var(--ease-standard);
}

.notif-item:hover {
  box-shadow: var(--shadow-2);
}

.notif-item__bar {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  border-radius: var(--radius-m) 0 0 var(--radius-m);
}

.notif-item__type-icon {
  flex-shrink: 0;
  display: flex;
}

.notif-item__content {
  flex: 1;
  min-width: 0;
}

.notif-item__text {
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 2px;
}

.notif-item__time {
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}
</style>
