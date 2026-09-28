<template>
  <Teleport to="body">
    <div class="modal-mask" @click.self="$emit('close')">
      <div class="modal">
        <h3>发布求看心愿</h3>
        <p class="modal-sub">说出你想再看一眼的地方，悬赏硬币会从余额冻结，确认代看后结算给对方</p>

        <label class="label">标题 *</label>
        <input v-model.trim="form.title" maxlength="100" placeholder="例如：再看一眼巷口的老槐树" />

        <div class="form-row">
          <div>
            <label class="label">城市</label>
            <input v-model.trim="form.city" maxlength="50" placeholder="长沙" />
          </div>
          <div class="reward-field">
            <label class="label">悬赏硬币 *</label>
            <input v-model.number="form.reward" type="number" min="1" :max="balance" />
          </div>
        </div>
        <p class="field-hint">当前可用 <span class="coin">{{ balance }}</span>，发布后冻结，确认时代看人获得</p>

        <label class="label">具体地点</label>
        <input v-model.trim="form.location" maxlength="100" placeholder="街巷、学校、门店的名字" />

        <label class="label">过去回忆 *</label>
        <textarea v-model.trim="form.memory" maxlength="2000" placeholder="写下你关于这个地方的记忆，让路过的人懂你想看什么"></textarea>

        <label class="label">老照片 *（至少 1 张）</label>
        <PhotoUploader v-model="form.oldPhotos" :max="4" add-text="上传老照片" />

        <label class="label">想请对方拍什么</label>
        <textarea v-model.trim="form.wish" maxlength="500" placeholder="例如：拍一下老槐树现在的样子，树下还有没有下棋的老人"></textarea>

        <p v-if="error" class="form-error">{{ error }}</p>

        <div class="modal-actions">
          <button class="btn-subtle" @click="$emit('close')">取消</button>
          <button class="btn-primary" :disabled="submitting" @click="submit">
            {{ submitting ? '发布中…' : `冻结 ${form.reward || 0} 硬币并发布` }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { reactive, ref } from 'vue'
import PhotoUploader from './PhotoUploader.vue'
import { api } from '../api/index.js'
import { toast } from '../store/toast.js'

const props = defineProps({ balance: { type: Number, required: true } })
const emit = defineEmits(['close', 'created'])

const submitting = ref(false)
const error = ref('')
const form = reactive({
  title: '',
  city: '',
  location: '',
  memory: '',
  oldPhotos: [],
  wish: '',
  reward: 10
})

async function submit() {
  error.value = ''
  if (!form.title) return (error.value = '请填写标题')
  if (!form.memory || form.oldPhotos.length === 0) return (error.value = '请写下回忆并至少上传一张老照片')
  if (!form.reward || form.reward < 1) return (error.value = '悬赏至少 1 枚硬币')
  if (form.reward > props.balance) return (error.value = `可用硬币不足（当前 ${props.balance}）`)

  submitting.value = true
  try {
    const data = await api.post('/stories', { ...form })
    toast('求看已发布，悬赏硬币已冻结', 'success')
    emit('created', data)
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.form-row {
  display: flex;
  gap: 14px;
}
.form-row > div {
  flex: 1;
}
.reward-field input {
  font-weight: 700;
  color: #9a6b1e;
}
.field-hint {
  font-size: 12px;
  color: var(--ink-soft);
  margin-top: 6px;
}
.form-error {
  color: var(--danger);
  font-size: 13px;
  margin-top: 12px;
}
</style>
