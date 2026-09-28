<script setup>
import { ref, inject } from 'vue'
import api from '../api'

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  hint: { type: String, default: '' }
})
const emit = defineEmits(['update:modelValue'])
const uploading = ref(0)
const showToast = inject('toast')
const fileInput = ref(null)

async function pick(e) {
  const files = Array.from(e.target.files || [])
  for (const f of files) {
    uploading.value++
    try {
      const fd = new FormData()
      fd.append('file', f)
      const { url } = await api.post('/upload', fd, {
        headers: { 'Content-Type': 'multipart/form-data' }
      })
      emit('update:modelValue', [...props.modelValue, url])
    } catch (err) {
      showToast(err.message, 'error')
    } finally {
      uploading.value--
    }
  }
  e.target.value = ''
}

function remove(url) {
  emit('update:modelValue', props.modelValue.filter((u) => u !== url))
}
</script>

<template>
  <div>
    <div class="up-row">
      <button type="button" class="btn-ghost btn-sm" :disabled="uploading > 0"
              @click="fileInput.click()">
        {{ uploading > 0 ? `上传中 ${uploading}…` : '📷 上传照片' }}
      </button>
      <span v-if="hint" class="muted">{{ hint }}</span>
    </div>
    <input ref="fileInput" type="file" accept="image/*" multiple hidden @change="pick" />
    <div v-if="modelValue.length" class="previews">
      <div v-for="(u, i) in modelValue" :key="i" class="preview">
        <img :src="u" alt="预览" />
        <button type="button" class="rm" @click="remove(u)">×</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.up-row { display: flex; align-items: center; gap: 10px; }
.previews {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}
.preview { position: relative; }
.preview img {
  width: 84px;
  height: 84px;
  object-fit: cover;
  border-radius: 8px;
  border: 1px solid var(--line);
}
.rm {
  position: absolute;
  top: -7px;
  right: -7px;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  padding: 0;
  background: var(--ink);
  color: #fff;
  font-size: 13px;
  line-height: 1;
}
</style>
