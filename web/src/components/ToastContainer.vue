<template>
  <Teleport to="body">
    <div class="toast-container">
      <TransitionGroup name="toast">
        <div
          v-for="t in toasts"
          :key="t.id"
          class="toast-item"
          :class="'toast-item--' + t.type"
        >
          {{ t.message }}
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<script setup>
import { useToast } from '../composables/useToast'
const { toasts } = useToast()
</script>

<style scoped>
.toast-container {
  position: fixed;
  bottom: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  flex-direction: column-reverse;
  gap: 8px;
  pointer-events: none;
}

.toast-item {
  pointer-events: auto;
  padding: 10px 16px;
  border-radius: var(--radius-m, 10px);
  background: var(--bg-card, #fff);
  color: var(--text-1, #2a2234);
  font-size: 13px;
  line-height: 1.4;
  box-shadow: var(--shadow-4, 0 12px 32px rgba(0,0,0,0.08));
  border-left: 3px solid var(--border, rgba(0,0,0,0.08));
  max-width: 320px;
  word-break: break-word;
}

.toast-item--error {
  border-left-color: #c47878;
}

.toast-item--success {
  border-left-color: #6db89a;
}

.toast-enter-active {
  transition: all var(--duration-medium-2, 250ms) var(--ease-standard, cubic-bezier(0.2, 0, 0, 1));
}
.toast-leave-active {
  transition: all var(--duration-medium-1, 200ms) var(--ease-standard, cubic-bezier(0.2, 0, 0, 1));
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(40px);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(40px);
}
</style>