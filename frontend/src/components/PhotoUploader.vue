<template>
  <div class="photo-uploader">
    <div class="thumbs">
      <div v-for="(url, i) in modelValue" :key="url" class="thumb">
        <img :src="url" alt="已上传图片" />
        <button type="button" class="thumb-del" @click="remove(i)">×</button>
      </div>
      <label v-if="!disabled && modelValue.length < max" class="thumb-add">
        <input type="file" accept="image/*" hidden @change="onPick" />
        <span v-if="uploading">上传中…</span>
        <template v-else>
          <span class="plus">＋</span>
          <span class="add-text">{{ addText }}</span>
        </template>
      </label>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { api } from '../api/index.js'

const props = defineProps({
  modelValue: { type: Array, default: () => [] },
  max: { type: Number, default: 4 },
  addText: { type: String, default: '上传照片' },
  disabled: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue'])

const uploading = ref(false)

async function onPick(e) {
  const file = e.target.files[0]
  e.target.value = ''
  if (!file) return
  uploading.value = true
  try {
    const data = await api.upload(file)
    emit('update:modelValue', [...props.modelValue, data.url])
  } catch (err) {
    alert(err.message)
  } finally {
    uploading.value = false
  }
}

function remove(i) {
  emit('update:modelValue', props.modelValue.filter((_, idx) => idx !== i))
}
</script>

<style scoped>
.thumbs {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}
.thumb,
.thumb-add {
  width: 88px;
  height: 88px;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
}
.thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}
.thumb-del {
  position: absolute;
  top: 3px;
  right: 3px;
  width: 20px;
  height: 20px;
  padding: 0;
  border-radius: 50%;
  background: rgba(74, 60, 44, 0.62);
  color: #fff;
  font-size: 14px;
  line-height: 1;
}
.thumb-add {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 3px;
  border: 1.5px dashed var(--line);
  background: var(--paper);
  color: var(--ink-soft);
  cursor: pointer;
  font-size: 12px;
}
.thumb-add:hover {
  border-color: var(--accent);
  color: var(--accent-deep);
}
.plus {
  font-size: 22px;
  line-height: 1;
}
</style>
