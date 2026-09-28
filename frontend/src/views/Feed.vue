<script setup>
import { onMounted, ref, computed, inject } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'
import { auth } from '../auth'
import PostcardCard from '../components/PostcardCard.vue'
import StoryFormModal from '../components/StoryFormModal.vue'
import RespondModal from '../components/RespondModal.vue'
import AppendModal from '../components/AppendModal.vue'

const router = useRouter()
const showToast = inject('toast')

const stories = ref([])
const loading = ref(true)
const filter = ref('all')
const city = ref('')

const showStoryModal = ref(false)
const respondTarget = ref(null)
const appendTarget = ref(null)

const cities = computed(() => [...new Set(stories.value.map((s) => s.city))])
const filtered = computed(() => {
  return stories.value.filter((s) => {
    if (filter.value !== 'all' && s.status !== filter.value) return false
    if (city.value && s.city !== city.value) return false
    return true
  })
})

async function load() {
  loading.value = true
  try {
    const { stories: list } = await api.get('/stories')
    stories.value = list
  } catch (e) {
    showToast(e.message, 'error')
  } finally {
    loading.value = false
  }
}

function publish() {
  if (!auth.isLogin) {
    router.push('/login')
    return
  }
  showStoryModal.value = true
}

async function acceptResponse(resp) {
  if (!confirm(`确认采纳这张现场明信片？${resp.story_id ? '' : ''}悬赏硬币将立即结算给对方，不可撤销。`)) return
  try {
    await api.post(`/responses/${resp.id}/accept`)
    showToast('已确认采纳，记忆硬币已结算给对方')
    await Promise.all([load(), auth.refresh()])
  } catch (e) {
    showToast(e.message, 'error')
  }
}

onMounted(load)
</script>

<template>
  <section>
    <!-- 顶部 Hero -->
    <div class="hero">
      <div class="hero-text">
        <h1>让远方的人，替你再看一眼回忆的角落</h1>
        <p>离开那座城以后，老街还在吗？旧摊子还生火吗？<br />
           发布求看、冻结记忆硬币，等待一张来自当下的双面明信片。</p>
        <button class="btn-primary hero-btn" @click="publish">✉ 发布我的求看</button>
      </div>
      <div class="hero-stamp">
        <div class="stamp-circle">◉</div>
        <p>记忆硬币循环</p>
        <span>注册赠送 → 发布冻结 → 代看上传 → 采纳结算</span>
      </div>
    </div>

    <!-- 筛选 -->
    <div class="filters">
      <div class="filter-tabs">
        <button :class="{ active: filter === 'all' }" @click="filter = 'all'">全部</button>
        <button :class="{ active: filter === 'open' }" @click="filter = 'open'">求看中</button>
        <button :class="{ active: filter === 'fulfilled' }" @click="filter = 'fulfilled'">已归还</button>
      </div>
      <select class="input city-select" v-model="city">
        <option value="">全部城市</option>
        <option v-for="c in cities" :key="c" :value="c">{{ c }}</option>
      </select>
    </div>

    <!-- Feed -->
    <div v-if="loading" class="empty">正在拆阅明信片…</div>
    <div v-else-if="filtered.length === 0" class="empty">
      <div class="big-font">📭</div>
      这里还没有求看，寄出第一张吧
    </div>
    <PostcardCard
      v-for="s in filtered"
      :key="s.id"
      :story="s"
      @respond="respondTarget = $event"
      @accept="acceptResponse($event)"
      @append="appendTarget = $event"
    />
  </section>

  <StoryFormModal v-if="showStoryModal"
                  @close="showStoryModal = false"
                  @created="() => { showStoryModal = false; load(); showToast('求看已发布，悬赏硬币已冻结') }" />
  <RespondModal v-if="respondTarget"
                :story="respondTarget"
                @close="respondTarget = null"
                @done="() => { respondTarget = null; load(); showToast('现场明信片已寄出，等待发起人确认') }" />
  <AppendModal v-if="appendTarget"
               :story="appendTarget"
               @close="appendTarget = null"
               @done="() => { appendTarget = null; load(); showToast('已追加悬赏并冻结') }" />
</template>

<style scoped>
.hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 24px;
  background:
    linear-gradient(135deg, rgba(194,104,63,.10), rgba(201,162,79,.12)),
    var(--card);
  border: 1px solid var(--line);
  border-radius: 18px;
  padding: 30px 32px;
  margin-bottom: 22px;
  box-shadow: var(--shadow);
}
.hero-text h1 { font-size: 24px; margin-bottom: 10px; letter-spacing: 1px; }
.hero-text p { color: var(--ink-soft); font-size: 14px; margin-bottom: 18px; }
.hero-btn { padding: 11px 26px; font-size: 15px; }

.hero-stamp {
  flex-shrink: 0;
  text-align: center;
  border: 2px dashed var(--gold);
  border-radius: 14px;
  padding: 16px 20px;
}
.stamp-circle {
  width: 46px;
  height: 46px;
  margin: 0 auto 6px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 30%, #e6c877, var(--gold));
  color: #fff;
  font-size: 22px;
  line-height: 46px;
  box-shadow: inset 0 -3px 6px rgba(0,0,0,.15);
}
.hero-stamp p { font-weight: 700; font-size: 14px; margin-bottom: 4px; }
.hero-stamp span { font-size: 11px; color: var(--ink-soft); }

.filters {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 18px;
  gap: 12px;
}
.filter-tabs { display: flex; gap: 4px; background: var(--paper-warm); border-radius: 10px; padding: 4px; }
.filter-tabs button {
  background: transparent;
  color: var(--ink-soft);
  padding: 6px 16px;
  font-size: 13px;
  border-radius: 8px;
}
.filter-tabs button.active {
  background: var(--card);
  color: var(--accent);
  font-weight: 600;
  box-shadow: 0 1px 4px rgba(90,70,45,.12);
}
.city-select { width: auto; padding: 7px 12px; font-size: 13px; }

@media (max-width: 680px) {
  .hero { flex-direction: column; text-align: center; }
}
</style>
