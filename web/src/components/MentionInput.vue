<template>
  <div class="mention-input-wrapper" ref="wrapperRef">
    <textarea
      ref="textareaRef"
      :value="modelValue"
      :placeholder="placeholder"
      :rows="rows"
      class="mention-textarea"
      @input="onInput"
      @keydown="onKeydown"
      @blur="onBlur"
    />
    <ul v-if="showDropdown && filteredUsers.length" class="mention-dropdown" ref="dropdownRef">
      <li
        v-for="(user, idx) in filteredUsers"
        :key="user.id"
        :class="['mention-item', { active: idx === activeIndex }]"
        @mousedown.prevent="selectUser(user)"
        @mouseenter="activeIndex = idx"
      >
        <span class="mention-item-name">{{ user.nickname }}</span>
        <span class="mention-item-username">@{{ user.username }}</span>
      </li>
    </ul>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { getMentionCache } from '../api'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  rows: { type: Number, default: 4 },
  nodeId: { type: [Number, String], default: null }
})

const emit = defineEmits(['update:modelValue', 'keydown'])

const textareaRef = ref(null)
const wrapperRef = ref(null)
const dropdownRef = ref(null)
const showDropdown = ref(false)
const activeIndex = ref(0)
const query = ref('')
const cursorPosition = ref(0)
const atStartPos = ref(-1)

// Cached mention users
let cachedUsers = []
let cacheExpiry = 0
const CACHE_TTL = 5 * 60 * 1000 // 5 minutes

const filteredUsers = computed(() => {
  if (!query.value) return cachedUsers.slice(0, 20)
  const q = query.value.toLowerCase()
  return cachedUsers
    .filter(u =>
      u.username.toLowerCase().includes(q) ||
      u.nickname.toLowerCase().includes(q)
    )
    .slice(0, 20)
})

async function loadCache() {
  const now = Date.now()
  if (cachedUsers.length && now < cacheExpiry) return

  try {
    const sources = localStorage.getItem('mention_sources') || 'following,followers,admins,moderators'
    const params = { sources }
    if (props.nodeId) params.node_id = props.nodeId
    const res = await getMentionCache(params)
    cachedUsers = res.data || []
    cacheExpiry = now + CACHE_TTL
  } catch {
    // silently fail, use empty cache
  }
}

function findAtPosition(text, cursor) {
  // Look backwards from cursor to find the @ that started this mention
  let i = cursor - 1
  while (i >= 0) {
    const ch = text[i]
    if (ch === '@' && (i === 0 || /[\s\n]/.test(text[i - 1]))) {
      return i
    }
    if (/[\s\n]/.test(ch)) break
    i--
  }
  return -1
}

function onInput(e) {
  const val = e.target.value
  emit('update:modelValue', val)
  cursorPosition.value = e.target.selectionStart

  const atPos = findAtPosition(val, cursorPosition.value)
  if (atPos >= 0) {
    atStartPos.value = atPos
    query.value = val.slice(atPos + 1, cursorPosition.value)
    showDropdown.value = true
    activeIndex.value = 0
    loadCache()
  } else {
    showDropdown.value = false
    query.value = ''
  }
}

function onKeydown(e) {
  if (!showDropdown.value || !filteredUsers.value.length) {
    emit('keydown', e)
    return
  }

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    activeIndex.value = (activeIndex.value + 1) % filteredUsers.value.length
    scrollToActive()
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    activeIndex.value = (activeIndex.value - 1 + filteredUsers.value.length) % filteredUsers.value.length
    scrollToActive()
  } else if (e.key === 'Enter' || e.key === 'Tab') {
    e.preventDefault()
    selectUser(filteredUsers.value[activeIndex.value])
  } else if (e.key === 'Escape') {
    showDropdown.value = false
  } else {
    // Forward unhandled keys (e.g. Ctrl+Enter) to parent
    emit('keydown', e)
    return
  }
}

function scrollToActive() {
  nextTick(() => {
    const dropdown = dropdownRef.value
    if (!dropdown) return
    const active = dropdown.children[activeIndex.value]
    if (active) active.scrollIntoView({ block: 'nearest' })
  })
}

function selectUser(user) {
  const val = props.modelValue
  const before = val.slice(0, atStartPos.value)
  const after = val.slice(cursorPosition.value)
  const newText = `${before}@${user.username}${after}`
  emit('update:modelValue', newText)
  showDropdown.value = false
  query.value = ''

  nextTick(() => {
    const ta = textareaRef.value
    if (ta) {
      const pos = atStartPos.value + user.username.length + 1
      ta.focus()
      ta.setSelectionRange(pos, pos)
    }
  })
}

function onBlur() {
  // Delay to allow mousedown on dropdown items
  setTimeout(() => {
    showDropdown.value = false
  }, 200)
}

// Invalidate cache when nodeId changes
onMounted(() => {
  cachedUsers = []
  cacheExpiry = 0
})

// Expose loadCache for parent to preload
defineExpose({ loadCache })
</script>

<style scoped>
.mention-input-wrapper {
  position: relative;
}

.mention-textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-m);
  background: var(--bg-card);
  color: var(--text-1);
  font-size: 14px;
  line-height: 1.5;
  resize: vertical;
  font-family: inherit;
  transition: border-color 0.2s;
}

.mention-textarea:focus {
  outline: none;
  border-color: var(--accent);
}

.mention-textarea::placeholder {
  color: var(--text-4);
}

.mention-dropdown {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 100%;
  max-height: 200px;
  overflow-y: auto;
  margin: 0 0 4px;
  padding: 4px 0;
  list-style: none;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-m);
  box-shadow: var(--shadow-4);
  z-index: 100;
}

.mention-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  cursor: pointer;
  transition: background 0.15s;
}

.mention-item:hover,
.mention-item.active {
  background: var(--state-hover);
}

.mention-item-name {
  font-weight: 500;
  color: var(--text-1);
}

.mention-item-username {
  font-size: 12px;
  color: var(--text-4);
}
</style>