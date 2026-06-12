<template>
  <div class="create-post">
    <div class="create-post__header fade-up">
      <button class="btn btn--ghost btn--sm" @click="goBack">
        <PhArrowLeft :size="16" /> 返回
      </button>
      <h1 class="font-display create-post__title">发布新帖</h1>
      <div style="width: 60px"></div>
    </div>

    <div class="card create-post__form fade-up" style="animation-delay: 40ms">
      <div class="auth-form__group">
        <label class="auth-form__label">选择节点</label>
        <select v-model="form.node_id" class="input create-post__select">
          <option :value="null" disabled>请选择节点</option>
          <option v-for="node in nodes" :key="node.id" :value="node.id">{{ node.name }}</option>
        </select>
      </div>

      <div class="auth-form__group">
        <label class="auth-form__label">标题</label>
        <input v-model="form.title" class="input" placeholder="输入帖子标题..." />
      </div>

      <div class="auth-form__group">
        <label class="auth-form__label">内容（输入 @ 提及他人）</label>
        <MentionInput
          v-model="form.content"
          placeholder="写下你的想法..."
          :rows="8"
          :node-id="form.node_id"
          ref="mentionInputRef"
        />
      </div>

      <div class="create-post__actions">
        <button class="btn btn--outline" @click="goBack">取消</button>
        <button
          class="btn btn--primary"
          @click="handleSubmit"
          :disabled="!form.node_id || !form.title || !form.content || submitting"
        >
          <PhCircleNotch v-if="submitting" :size="16" class="spin" />
          <PhPaperPlaneTilt v-else :size="16" />
          <span>发布</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useToast } from '../composables/useToast'
import { createPost, listNodes } from '../api'
import MentionInput from '../components/MentionInput.vue'
import { PhArrowLeft, PhCircleNotch, PhPaperPlaneTilt } from '@phosphor-icons/vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const toast = useToast()

const nodes = ref([])
const submitting = ref(false)
const mentionInputRef = ref(null)

const form = reactive({
  node_id: route.query.node_id ? Number(route.query.node_id) : null,
  title: '',
  content: ''
})

function goBack() {
  router.back()
}

async function handleSubmit() {
  if (!form.node_id || !form.title || !form.content) return
  submitting.value = true
  try {
    const res = await createPost({
      node_id: form.node_id,
      title: form.title,
      content: form.content
    })
    toast.success('发布成功')
    const code = res.data?.code
    if (code) {
      router.push({ name: 'PostDetail', params: { code } })
    } else {
      router.push({ name: 'Home' })
    }
  } catch (e) {
    toast.error(e.message || '发帖失败')
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  if (!auth.isLoggedIn) {
    router.push({ name: 'Login', query: { redirect: route.fullPath } })
    return
  }
  try {
    const res = await listNodes()
    nodes.value = res.data?.nodes || []
  } catch (e) {
    toast.error(e.message)
  }
  // Preload mention cache
  if (mentionInputRef.value) {
    mentionInputRef.value.loadCache()
  }
})
</script>

<style scoped>
.create-post {
  max-width: 680px;
  margin: 0 auto;
  padding: var(--space-4, 1rem);
}

.create-post__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-6, 1.5rem);
}

.create-post__title {
  font-size: 20px;
}

.create-post__form {
  padding: var(--space-6, 1.5rem);
}

.create-post__select {
  height: 38px;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='%237e7290' viewBox='0 0 256 256'%3E%3Cpath d='M213.66,101.66l-80,80a8,8,0,0,1-11.32,0l-80-80A8,8,0,0,1,53.66,90.34L128,164.69l74.34-74.35a8,8,0,0,1,11.32,11.32Z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: 32px;
}

.create-post__actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3, 0.75rem);
  margin-top: var(--space-4, 1rem);
}
</style>