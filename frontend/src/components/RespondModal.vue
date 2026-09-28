<script setup>
import { ref } from 'vue'
import api from '../api'
import PhotoUploader from './PhotoUploader.vue'

const props = defineProps({ story: { type: Object, required: true } })
const emit = defineEmits(['close', 'done'])
const form = ref({ now_text: '', new_photos: [], message: '' })
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    await api.post(`/stories/${props.story.id}/responses`, form.value)
    emit('done')
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
      <h3>替他去拍 · 还回当下这一面</h3>
      <div class="target muted">
        正在代看：<b>{{ story.title }}</b>（{{ story.city }} · {{ story.location }}）
      </div>
      <form @submit.prevent="submit" style="margin-top:14px">
        <div class="form-row">
          <label>你在现场看到的样子</label>
          <textarea class="textarea" v-model="form.now_text" required
                    placeholder="它还在吗？变成了什么样？尽量描述发起人记忆里的细节……"></textarea>
        </div>
        <div class="form-row">
          <label>现场新照片（可选，多张）</label>
          <PhotoUploader v-model="form.new_photos" hint="拍下此刻的它" />
        </div>
        <div class="form-row">
          <label>给 TA 的寄语（可选）</label>
          <textarea class="textarea" v-model="form.message" style="min-height:70px"
                    placeholder="想对远方的 TA 说的话……"></textarea>
        </div>
        <div class="modal-actions">
          <button type="button" class="btn-ghost" @click="emit('close')">取消</button>
          <button class="btn-teal" :disabled="loading">
            {{ loading ? '提交中…' : '寄出当下现场' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
.target { font-size: 13px; }
</style>
