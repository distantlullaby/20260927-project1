<template>
  <Teleport to="body">
    <div class="modal-mask" @click.self="$emit('close')">
      <div class="modal">
        <h3>替 TA 去拍 · {{ story.title }}</h3>
        <p class="modal-sub">
          {{ story.city }} · {{ story.location }}<br />
          TA 的心愿：{{ story.wish || '拍下这个地方现在的样子' }}
        </p>

        <label class="label">现场新照片 *（至少 1 张）</label>
        <PhotoUploader v-model="photos" :max="4" add-text="上传现场照" />

        <label class="label">给 TA 的寄语 *</label>
        <textarea v-model.trim="message" maxlength="2000" placeholder="告诉 TA 你看到的一切：树还在不在、街变成了什么样、今天天气好不好"></textarea>

        <p v-if="error" class="form-error">{{ error }}</p>

        <div class="modal-actions">
          <button class="btn-subtle" @click="$emit('close')">取消</button>
          <button class="btn-sage" :disabled="submitting" @click="submit">
            {{ submitting ? '提交中…' : '把当下寄给 TA' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'
import PhotoUploader from './PhotoUploader.vue'
import { api } from '../api/index.js'
import { toast } from '../store/toast.js'

const props = defineProps({ story: { type: Object, required: true } })
const emit = defineEmits(['close', 'done'])

const photos = ref([])
const message = ref('')
const submitting = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  if (photos.value.length === 0) return (error.value = '请至少上传一张现场新照片')
  if (!message.value) return (error.value = '请写下给 TA 的寄语')

  submitting.value = true
  try {
    const data = await api.post(`/stories/${props.story.id}/responses`, {
      newPhotos: photos.value,
      message: message.value
    })
    toast('现场已寄出，等待 TA 确认', 'success')
    emit('done', data.story)
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.modal-sub {
  line-height: 1.8;
}
.form-error {
  color: var(--danger);
  font-size: 13px;
  margin-top: 12px;
}
</style>
