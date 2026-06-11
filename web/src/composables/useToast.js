import { ref } from 'vue'

const toasts = ref([])
let nextId = 0

function add(type, message, duration = 3000) {
  const id = nextId++
  toasts.value.push({ id, type, message })
  setTimeout(() => {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }, duration)
}

export function useToast() {
  return {
    toasts,
    error: (msg) => add('error', msg),
    success: (msg) => add('success', msg),
  }
}