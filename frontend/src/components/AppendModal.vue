<script setup>
import { ref } from 'vue'
import api from '../api'
import { auth } from '../auth'

const props = defineProps({ story: { type: Object, required: true } })
const emit = defineEmits(['close', 'done'])
const amount = ref(5)
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    await api.post(`/stories/${props.story.id}/bounty`, { amount: amount.value })
    await auth.refresh()
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
    <div class="modal" style="max-width:420px">
      <h3>追加悬赏</h3>
      <p class="muted">《{{ story.title }}》当前冻结悬赏
        <b class="coin">{{ story.bounty }}</b>
      </p>
      <div class="form-row" style="margin-top:16px">
        <label>追加数量（将继续冻结）</label>
        <input class="input" type="number" min="1" max="9999" v-model.number="amount" />
        <p class="muted" style="margin-top:8px">
          追加后总悬赏 <b class="coin">{{ story.bounty + (amount || 0) }}</b>，
          可用余额 <b class="coin">{{ auth.user?.balance ?? 0 }}</b>
        </p>
      </div>
      <div class="modal-actions">
        <button type="button" class="btn-ghost" @click="emit('close')">取消</button>
        <button class="btn-primary" :disabled="loading || !amount">
          {{ loading ? '处理中…' : '确认追加冻结' }}
        </button>
      </div>
    </div>
  </div>
</template>
