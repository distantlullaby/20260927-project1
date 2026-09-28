<template>
  <div class="feed">
    <!-- 头部横幅 -->
    <section class="hero">
      <div class="hero-text">
        <h1>那些回不去的角落，<br />总有人刚好路过</h1>
        <p class="hero-sub">
          离城的人留下老照片与回忆，还在城里的人拍下今天的现场。
          一枚记忆硬币，让远方的牵挂被认真对待。
        </p>
        <div class="hero-actions">
          <button v-if="auth.user" class="btn-primary" @click="showCreate = true">
            ＋ 发布求看心愿
          </button>
          <RouterLink v-else to="/login" class="btn-primary">登录后发布求看</RouterLink>
          <span v-if="auth.user" class="hero-balance">
            可用 <span class="coin">{{ auth.user.balance }}</span>
            · 冻结 <span class="coin frozen">{{ auth.user.frozen }}</span>
          </span>
        </div>
      </div>
      <div class="hero-art" aria-hidden="true">
        <svg viewBox="0 0 180 140" width="100%">
          <rect x="10" y="14" width="160" height="112" rx="8" fill="#fffdf8" stroke="#d9c7a5" stroke-width="2"/>
          <path d="M90 14v112" stroke="#e0d2b8" stroke-width="1.4" stroke-dasharray="5 4"/>
          <rect x="30" y="78" width="34" height="34" fill="#c98d5e"/>
          <polygon points="24,78 47,58 70,78" fill="#a9673d"/>
          <circle cx="60" cy="48" r="13" fill="#93bd72"/>
          <path d="M110 52h36M110 66h36M110 80h24" stroke="#cbb89a" stroke-width="3.5" stroke-linecap="round"/>
          <circle cx="140" cy="100" r="11" fill="none" stroke="#d9a441" stroke-width="2"/>
          <text x="118" y="104" font-size="11" fill="#9a6b1e" text-anchor="middle">硬币</text>
        </svg>
      </div>
    </section>

    <!-- 筛选 -->
    <div class="filters">
      <button
        v-for="f in filters"
        :key="f.key"
        class="filter-chip"
        :class="{ active: scope === f.key }"
        @click="changeScope(f.key)"
      >
        {{ f.label }}
      </button>
      <div class="filters-right">
        <select v-model="statusFilter" @change="loadStories">
          <option value="">全部状态</option>
          <option value="open">等待代看</option>
          <option value="settled">记忆已归还</option>
        </select>
      </div>
    </div>

    <!-- 卡片列表 -->
    <div v-if="loading" class="loading">明信片正在从远方寄来…</div>
    <div v-else-if="!stories.length" class="empty">
      <span class="empty-emoji">🪶</span>
      <p>{{ emptyText }}</p>
    </div>

    <div v-else class="card-list">
      <PostcardCard
        v-for="s in stories"
        :key="s.id"
        :story="s"
        @respond="onRespond"
        @append="onAppend"
        @accept="onAccept"
        @preview="(url) => (previewUrl = url)"
      />
    </div>

    <!-- 弹窗 -->
    <CreateStoryModal
      v-if="showCreate && auth.user"
      :balance="auth.user.balance"
      @close="showCreate = false"
      @created="onCreated"
    />
    <RespondModal
      v-if="respondTarget"
      :story="respondTarget"
      @close="respondTarget = null"
      @done="onResponseDone"
    />
    <AppendRewardModal
      v-if="appendTarget"
      :story="appendTarget"
      :balance="auth.user.balance"
      @close="appendTarget = null"
      @done="onAppendDone"
    />
    <Lightbox :url="previewUrl" @close="previewUrl = ''" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import PostcardCard from '../components/PostcardCard.vue'
import CreateStoryModal from '../components/CreateStoryModal.vue'
import RespondModal from '../components/RespondModal.vue'
import AppendRewardModal from '../components/AppendRewardModal.vue'
import Lightbox from '../components/Lightbox.vue'
import { api } from '../api/index.js'
import { auth, updateUser } from '../store/auth.js'
import { toast } from '../store/toast.js'

const emit = defineEmits(['balance-changed'])

const stories = ref([])
const loading = ref(false)
const scope = ref('all')
const statusFilter = ref('')

const showCreate = ref(false)
const respondTarget = ref(null)
const appendTarget = ref(null)
const previewUrl = ref('')

const filters = [
  { key: 'all', label: '全部求看' },
  { key: 'mine', label: '我发布的' },
  { key: 'responded', label: '我代看的' }
]

const emptyText = computed(() => {
  if (scope.value === 'mine') return '你还没有发布过求看心愿，说出一个想再看一眼的地方吧'
  if (scope.value === 'responded') return '你还没有替谁去拍过，去广场上看看谁在等你'
  return '这里还很安静，来发出第一份求看心愿吧'
})

async function loadStories() {
  loading.value = true
  try {
    const params = new URLSearchParams()
    if (scope.value !== 'all') params.set('scope', scope.value)
    if (statusFilter.value) params.set('status', statusFilter.value)
    const qs = params.toString() ? '?' + params.toString() : ''
    const data = await api.get('/stories' + qs)
    stories.value = data.stories
  } catch (e) {
    if (scope.value !== 'all') {
      toast(e.message, 'error')
    }
  } finally {
    loading.value = false
  }
}

function changeScope(key) {
  if (key !== 'all' && !auth.user) {
    toast('请先登录', 'error')
    return
  }
  scope.value = key
  loadStories()
}

function onRespond(story) {
  respondTarget.value = story
}

function onAppend(story) {
  appendTarget.value = story
}

function onCreated(data) {
  showCreate.value = false
  if (data.balance !== undefined && auth.user) {
    updateUser({ ...auth.user, balance: data.balance, frozen: data.frozen })
  }
  emit('balance-changed')
  scope.value = 'all'
  statusFilter.value = ''
  loadStories()
}

function onResponseDone(updatedStory) {
  respondTarget.value = null
  replaceStory(updatedStory)
  loadStories()
}

function onAppendDone(data) {
  appendTarget.value = null
  if (data.balance !== undefined && auth.user) {
    updateUser({ ...auth.user, balance: data.balance, frozen: data.frozen })
  }
  emit('balance-changed')
  replaceStory(data.story)
  toast(`总悬赏已提高到 ${data.story.reward} 枚`, 'success')
}

async function onAccept(response) {
  const story = stories.value.find((s) => s.id === response.storyId || s.responses.some((r) => r.id === response.id))
  if (!story) return
  if (!confirm(`确认采纳 ${response.user.nickname} 的现场吗？${story.reward} 枚硬币将立即结算给对方，不可撤回。`)) return

  try {
    const data = await api.post(`/stories/${story.id}/responses/${response.id}/accept`, {})
    if (data.balance !== undefined && auth.user) {
      updateUser({ ...auth.user, balance: data.balance, frozen: data.frozen })
    }
    emit('balance-changed')
    replaceStory(data.story)
    toast(`记忆已归还，${story.reward} 枚硬币已送达对方`, 'success')
  } catch (e) {
    toast(e.message, 'error')
  }
}

function replaceStory(updated) {
  if (!updated) return
  const idx = stories.value.findIndex((s) => s.id === updated.id)
  if (idx !== -1) stories.value[idx] = updated
}

onMounted(loadStories)
</script>

<style scoped>
.hero {
  display: grid;
  grid-template-columns: 1fr 240px;
  gap: 24px;
  align-items: center;
  background: linear-gradient(120deg, rgba(255, 253, 248, 0.9), rgba(239, 228, 208, 0.7));
  border: 1px solid var(--line);
  border-radius: 20px;
  padding: 32px 34px;
  box-shadow: var(--shadow);
}
.hero-text h1 {
  font-size: 27px;
  line-height: 1.5;
  letter-spacing: 1px;
  color: var(--ink);
}
.hero-sub {
  margin-top: 14px;
  color: var(--ink-soft);
  font-size: 14px;
  line-height: 1.9;
  max-width: 520px;
}
.hero-actions {
  margin-top: 22px;
  display: flex;
  align-items: center;
  gap: 18px;
}
.hero-actions .btn-primary {
  padding: 12px 26px;
  font-size: 15px;
}
.hero-balance {
  font-size: 13.5px;
  color: var(--ink-soft);
}
.coin.frozen {
  color: #7a6a85;
}
.coin.frozen::before {
  color: #a793b5;
}

.filters {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 26px 0 18px;
}
.filter-chip {
  background: var(--card);
  border: 1.5px solid var(--line);
  color: var(--ink-soft);
  border-radius: 999px;
  padding: 7px 18px;
  font-size: 13.5px;
}
.filter-chip.active {
  background: var(--ink);
  border-color: var(--ink);
  color: var(--paper);
}
.filters-right {
  margin-left: auto;
}
.filters-right select {
  width: auto;
  padding: 8px 14px;
  font-size: 13px;
}

.loading {
  text-align: center;
  color: var(--ink-soft);
  padding: 60px;
}
.card-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

@media (max-width: 680px) {
  .hero {
    grid-template-columns: 1fr;
    padding: 24px;
  }
  .hero-art {
    display: none;
  }
  .hero-text h1 {
    font-size: 22px;
  }
}
</style>
