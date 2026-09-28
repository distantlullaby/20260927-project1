<script setup>
import { ref, watch } from 'vue'
import api from '../api'
import { auth } from '../auth'
import PhotoUploader from './PhotoUploader.vue'

const emit = defineEmits(['close', 'created'])
const form = ref({
  title: '', city: '', location: '', memory_text: '',
  old_photos: [], bounty: 10
})
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    const { story } = await api.post('/stories', form.value)
    await auth.refresh()
    emit('created', story)
  } catch (e) {
    alert(e.message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="modal-mask" @click.self="emit('close')">
    <div class="modal">
      <h3>发布求看 · 寄出回忆那一面</h3>
      <form @submit.prevent="submit">
        <div class="form-row">
          <label>标题</label>
          <input class="input" v-model="form.title" required maxlength="100"
                 placeholder="例如：帮我看看巷口的老锅盔摊还在吗" />
        </div>
        <div class="form-row two">
          <div>
            <label>城市</label>
            <input class="input" v-model="form.city" required maxlength="50" placeholder="成都" />
          </div>
          <div>
            <label>具体地点</label>
            <input class="input" v-model="form.location" required maxlength="200"
                   placeholder="锦江区镗钯街尽头的老巷口" />
          </div>
        </div>
        <div class="form-row">
          <label>过去的回忆</label>
          <textarea class="textarea" v-model="form.memory_text" required
                    placeholder="写下你记忆里的那个角落、那个人、那个味道……"></textarea>
        </div>
        <div class="form-row">
          <label>老照片（可选，多张）</label>
          <PhotoUploader v-model="form.old_photos" hint="贴几张当年的照片，帮代看人找到它" />
        </div>
        <div class="form-row">
          <label>悬赏记忆硬币（发布即冻结）</label>
          <div class="bounty-row">
            <input class="input" type="number" min="1" max="9999" v-model.number="form.bounty" required />
            <span class="muted">当前可用 <b class="coin">{{ auth.user?.balance ?? 0 }}</b></span>
          </div>
        </div>
        <div class="modal-actions">
          <button type="button" class="btn-ghost" @click="emit('close')">取消</button>
          <button class="btn-primary" :disabled="loading">
            {{ loading ? '发布中…' : '冻结硬币并发布' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.two { display: grid; grid-template-columns: 1fr 2fr; gap: 10px; }
.bounty-row { display: flex; align-items: center; gap: 12px; }
.bounty-row .input { max-width: 130px; }
</style>
