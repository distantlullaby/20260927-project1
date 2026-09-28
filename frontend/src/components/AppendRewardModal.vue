<template>
  <Teleport to="body">
    <div class="modal-mask" @click.self="$emit('close')">
      <div class="modal modal-sm">
        <h3>追加悬赏</h3>
        <p class="modal-sub">
          《{{ story.title }}》当前悬赏 <span class="coin">{{ story.reward }}</span>，追加部分同样从可用余额冻结
        </p>

        <label class="label">追加硬币数量</label>
        <input v-model.number="amount" type="number" min="1" :max="balance" />
        <p class="field-hint">追加后总悬赏 <span class="coin">{{ (story.reward || 0) + (amount || 0) }}</span></p>

        <p v-if="error" class="form-error">{{ error }}</p>

        <div class="modal-actions">
          <button class="btn-subtle" @click="$emit('close')">取消</button>
          <button class="btn-primary" :disabled="submitting" @click="submit">
            {{ submitting ? '处理中…' : '确认追加冻结' }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'
import { api } from '../api/index.js'
import { toast } from '../store/toast.js'

const props = defineProps({
  story: { type: Object, required: true },
  balance: { type: Number, required: true }
})
const emit = defineEmits(['close', 'done'])

const amount = ref(5)
const submitting = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  if (!amount.value || amount.value < 1) return (error.value = '追加数量至少为 1')
  if (amount.value > props.balance) return (error.value = `可用硬币不足（当前 ${props.balance}）`)

  submitting.value = true
  try {
    const data = await api.post(`/stories/${props.story.id}/reward`, { amount: amount.value })
    toast(`已追加 ${amount.value} 枚硬币`, 'success')
    emit('done', data)
  } catch (e) {
    error.value = e.message
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.modal-sm {
  max-width: 420px;
}
.field-hint {
  font-size: 12px;
  color: var(--ink-soft);
  margin-top: 8px;
}
.form-error {
  color: var(--danger);
  font-size: 13px;
  margin-top: 12px;
}
</style>
