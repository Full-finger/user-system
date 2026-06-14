<template>
  <div class="messages fade-up">
    <h1 class="font-display" style="font-size: 26px; margin-bottom: 20px">私信</h1>

    <div class="card placeholder-card">
      <div class="placeholder-card__icon">
        <PhEnvelopeSimple :size="40" weight="bold" />
      </div>
      <h2 class="font-display placeholder-card__title">私信功能开发中</h2>
      <p class="text-3 placeholder-card__desc">
        实时私信、群组聊天、消息通知等功能正在开发中，敬请期待。
      </p>
    </div>

    <!-- 静态演示对话列表样式 -->
    <div class="messages__panel">
      <div class="messages__list card">
        <div
          v-for="(chat, i) in sampleChats"
          :key="i"
          class="chat-item"
          :class="{ 'chat-item--active': activeChat === i }"
          @click="activeChat = i"
        >
          <UserAvatar :name="chat.name" size="sm" />
          <div class="chat-item__content">
            <div class="chat-item__top">
              <span class="chat-item__name">{{ chat.name }}</span>
              <span class="text-4" style="font-size: 11px">{{ chat.time }}</span>
            </div>
            <p class="chat-item__preview text-3">{{ chat.preview }}</p>
          </div>
          <div v-if="chat.unread" class="chat-item__dot"></div>
        </div>
      </div>

      <div class="messages__chat card">
        <div class="messages__chat-header">
          <UserAvatar :name="sampleChats[activeChat]?.name" size="sm" />
          <span class="font-display" style="font-size: 14px; font-weight: 600">
            {{ sampleChats[activeChat]?.name }}
          </span>
        </div>
        <div class="messages__chat-body">
          <p class="text-4" style="text-align: center; padding: 40px 0">
            <PhChatsCircle :size="32" weight="bold" style="display: block; margin: 0 auto 8px" />
            消息功能开发中...
          </p>
        </div>
        <div class="messages__chat-input">
          <input class="input" placeholder="输入消息..." disabled />
          <button class="btn btn--primary btn--sm" disabled>
            <PhPaperPlaneRight :size="14" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { PhEnvelopeSimple, PhChatsCircle, PhPaperPlaneRight } from '@phosphor-icons/vue'
import UserAvatar from '../components/UserAvatar.vue'

const activeChat = ref(0)

const sampleChats = [
  { name: 'rustacean', preview: 'Hey! I saw your post about...', time: '2 分钟前', unread: true },
  { name: 'vuefan', preview: 'Thanks for the help!', time: '1 小时前', unread: false },
  { name: 'architect_wang', preview: '关于微服务架构的建议...', time: '昨天', unread: false },
]
</script>

<style scoped>
.placeholder-card {
  padding: 40px 32px;
  text-align: center;
  margin-bottom: 16px;
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

.messages__panel {
  display: flex;
  gap: 12px;
  height: 420px;
}

.messages__list {
  width: 260px;
  padding: 8px;
  overflow-y: auto;
  flex-shrink: 0;
}

.chat-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-m);
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: background var(--duration-medium-1) var(--ease-standard);
}

.chat-item::after {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: transparent;
  transition: background var(--duration-medium-1) var(--ease-standard);
  pointer-events: none;
}

.chat-item:hover::after {
  background: var(--state-hover);
}

.chat-item--active::after {
  background: var(--state-focus);
}

.chat-item__content {
  flex: 1;
  min-width: 0;
}

.chat-item__top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2px;
}

.chat-item__name {
  font-size: 13px;
  font-weight: 600;
  font-family: var(--font-display);
}

.chat-item__preview {
  font-size: 12px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.chat-item__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  flex-shrink: 0;
}

.messages__chat {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.messages__chat-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}

.messages__chat-body {
  flex: 1;
  padding: 16px;
}

.messages__chat-input {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid var(--border);
}

.messages__chat-input .input {
  flex: 1;
  height: 36px;
}

@media (max-width: 700px) {
  .messages__list {
    width: 100%;
  }

  .messages__chat {
    display: none;
  }

  .messages__panel {
    height: auto;
    flex-direction: column;
  }
}
</style>
