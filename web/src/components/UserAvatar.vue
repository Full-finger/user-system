<template>
  <div class="avatar" :class="`avatar--${size}`">
    <img
      v-if="avatarUrl && !error"
      :src="avatarUrl"
      :alt="initial"
      class="avatar__img"
      loading="lazy"
      @error="error = true"
    />
    <template v-else>{{ initial }}</template>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  avatarUrl: { type: String, default: '' },
  name: { type: String, default: '' },
  size: { type: String, default: 'sm' }, // sm | md | lg
})

const error = ref(false)
const initial = computed(() => (props.name || '?')[0].toUpperCase())
</script>

<style scoped>
.avatar__img {
  width: 100%;
  height: 100%;
  border-radius: inherit;
  object-fit: cover;
}
</style>