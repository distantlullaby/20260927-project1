<script setup>
import { onMounted, ref, inject } from 'vue'
import api from '../api'
import { auth } from '../auth'
import PostcardCard from '../components/PostcardCard.vue'
import RespondModal from '../components/RespondModal.vue'
import AppendModal from '../components/AppendModal.vue'

const data = ref(null)
const tab = ref('stories')
const loading = ref(true)
const showToast = inject('toast')
const respondTarget = ref(null)
const appendTarget = ref(null)

const ledgerTypeMap = {
  register: { label: '注册赠送', cls: 'in' },
  freeze:   { label: '发布冻结', cls: 'out' },
  add:      { label: '追加冻结', cls: 'out' },
  unfreeze: { label: '冻结退回', cls: 'in' },
  reward:   { label: '代看收入', cls: 'in' }
}

async function load() {
  loading.value = true
  try {
    data.value = await api.get('/profile')
  } finally {
    loading.value = false
  }
}

function fmtTime(t) {
  return new Date(t).toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit'
  })
}

async function acceptResponse(resp) {
  if (!confirm('确认采纳这张现场明信片？悬赏硬币将立即结算给对方，不可撤销。')) return
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
  <div v-if="loading" class="empty">加载中…</div>
  <template v-else-if="data">
    <!-- 硬币钱包 -->
    <section class="wallet card">
      <div class="wallet-main">
        <div class="wallet-id">
          <span class="avatar-lg">{{ (data.user.nickname || '?')[0] }}</span>
          <div>
            <h2>{{ data.user.nickname }}</h2>
            <p class="muted">@{{ data.user.username }}<template v-if="data.user.bio"> · {{ data.user.bio }}</template></p>
          </div>
        </div>
      </div>
      <div class="coin-stats">
        <div class="coin-stat available">
          <span class="stat-label">可用硬币</span>
          <span class="stat-num">◉ {{ data.user.balance }}</span>
        </div>
        <div class="coin-stat frozen">
          <span class="stat-label">冻结中（求看悬赏）</span>
          <span class="stat-num">◉ {{ data.frozen }}</span>
        </div>
        <div class="coin-stat earned">
          <span class="stat-label">代看累计收入</span>
          <span class="stat-num">◉ {{ data.earned }}</span>
        </div>
      </div>
    </section>

    <!-- Tab -->
    <div class="tabs">
      <button :class="{ active: tab === 'stories' }" @click="tab = 'stories'">
        我发布的求看（{{ data.my_stories.length }}）
      </button>
      <button :class="{ active: tab === 'responses' }" @click="tab = 'responses'">
        我代看的记录（{{ data.my_responses.length }}）
      </button>
      <button :class="{ active: tab === 'ledger' }" @click="tab = 'ledger'">
        硬币流水（{{ data.coin_ledgers.length }}）
      </button>
    </div>

    <!-- 我的求看 -->
    <div v-if="tab === 'stories'">
      <div v-if="data.my_stories.length === 0" class="empty">
        <div class="big-font">🗺️</div>你还没有发布过求看
      </div>
      <PostcardCard v-for="s in data.my_stories" :key="s.id" :story="s"
                    @respond="respondTarget = $event"
                    @accept="acceptResponse($event)"
                    @append="appendTarget = $event" />
    </div>

    <!-- 我的代看 -->
    <div v-if="tab === 'responses'" class="resp-list">
      <div v-if="data.my_responses.length === 0" class="empty">
        <div class="big-font">📷</div>去广场替别人看看远方的角落吧
      </div>
      <div v-for="r in data.my_responses" :key="r.id" class="resp-item card">
        <div class="resp-top">
          <span class="tag">📍 {{ r.story?.city }}</span>
          <span :class="['tag', r.status === 'accepted' ? 'tag-done' : r.status === 'rejected' ? 'tag-reject' : 'tag-open']">
            {{ r.status === 'accepted' ? '已采纳 · 硬币已到账' : r.status === 'rejected' ? '未采纳' : '等待确认' }}
          </span>
        </div>
        <h4>{{ r.story?.title }}</h4>
        <p class="muted resp-now">{{ r.now_text }}</p>
        <div v-if="r.message" class="resp-msg">寄语：{{ r.message }}</div>
        <div v-if="r.new_photos" class="resp-photos">
          <img v-for="(p, i) in r.new_photos.split(',')" :key="i" :src="p" alt="现场照" />
        </div>
        <p class="muted resp-time">{{ fmtTime(r.created_at) }}</p>
      </div>
    </div>

    <!-- 硬币流水 -->
    <div v-if="tab === 'ledger'" class="ledger card">
      <div v-if="data.coin_ledgers.length === 0" class="empty">暂无流水</div>
      <div v-for="l in data.coin_ledgers" :key="l.id" class="ledger-row">
        <div class="ledger-icon" :class="ledgerTypeMap[l.type]?.cls">
          {{ l.change > 0 ? '＋' : '－' }}
        </div>
        <div class="ledger-info">
          <span class="ledger-type">{{ ledgerTypeMap[l.type]?.label || l.type }}</span>
          <span class="muted">{{ l.remark }}</span>
          <span class="muted ledger-time">{{ fmtTime(l.created_at) }}</span>
        </div>
        <div class="ledger-amount" :class="l.change > 0 ? 'plus' : 'minus'">
          {{ l.change > 0 ? '+' : '' }}{{ l.change }}
          <span class="muted ledger-balance">余额 {{ l.balance_after }}</span>
        </div>
      </div>
    </div>

    <RespondModal v-if="respondTarget"
                  :story="respondTarget"
                  @close="respondTarget = null"
                  @done="() => { respondTarget = null; load(); showToast('现场明信片已寄出，等待发起人确认') }" />
    <AppendModal v-if="appendTarget"
                 :story="appendTarget"
                 @close="appendTarget = null"
                 @done="() => { appendTarget = null; load(); showToast('已追加悬赏并冻结') }" />
  </template>
</template>

<style scoped>
.wallet {
  padding: 24px 26px;
  margin-bottom: 22px;
  background:
    radial-gradient(400px 200px at 90% -50%, rgba(201,162,79,.18), transparent),
    var(--card);
}
.wallet-id { display: flex; align-items: center; gap: 14px; }
.avatar-lg {
  width: 54px;
  height: 54px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--teal), #46685f);
  color: #fff;
  font-size: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.wallet-id h2 { font-size: 20px; }
.coin-stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-top: 20px;
}
.coin-stat {
  border-radius: 12px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.coin-stat.available { background: linear-gradient(135deg, #f5e6c8, #ecd9ad); border: 1px solid #e0c78f; }
.coin-stat.frozen { background: #eef0f2; border: 1px solid #d9dde1; }
.coin-stat.earned { background: linear-gradient(135deg, #e2efe9, #d2e5dc); border: 1px solid #bcd8cb; }
.stat-label { font-size: 12px; color: var(--ink-soft); }
.stat-num { font-size: 22px; font-weight: 700; color: var(--ink); }

.tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 18px;
  border-bottom: 1px solid var(--line);
}
.tabs button {
  background: none;
  border-radius: 0;
  padding: 10px 16px;
  color: var(--ink-soft);
  font-size: 14px;
  border-bottom: 2px solid transparent;
  margin-bottom: -1px;
}
.tabs button.active {
  color: var(--accent);
  font-weight: 600;
  border-bottom-color: var(--accent);
}

.resp-item { padding: 18px 20px; margin-bottom: 14px; }
.resp-top { display: flex; gap: 6px; margin-bottom: 8px; }
.resp-item h4 { font-size: 16px; margin-bottom: 6px; }
.resp-now { font-size: 14px; }
.resp-msg {
  margin-top: 8px;
  font-size: 13px;
  background: var(--paper-warm);
  border-left: 3px solid var(--gold);
  padding: 8px 12px;
  border-radius: 0 8px 8px 0;
}
.resp-photos { display: flex; gap: 8px; margin-top: 10px; flex-wrap: wrap; }
.resp-photos img { width: 100px; height: 75px; object-fit: cover; border-radius: 6px; border: 1px solid var(--line); }
.resp-time { margin-top: 8px; font-size: 12px; }
.tag-reject { background: rgba(150,90,80,.12); color: #95504a; }

.ledger { padding: 6px 20px; }
.ledger-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 0;
  border-bottom: 1px dashed var(--line);
}
.ledger-row:last-child { border-bottom: none; }
.ledger-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
  flex-shrink: 0;
}
.ledger-icon.in { background: rgba(91,131,120,.16); color: var(--teal); }
.ledger-icon.out { background: rgba(194,104,63,.14); color: var(--accent); }
.ledger-info { flex: 1; display: flex; flex-direction: column; gap: 1px; }
.ledger-type { font-size: 14px; font-weight: 600; }
.ledger-time { font-size: 11px; }
.ledger-amount { font-size: 17px; font-weight: 700; white-space: nowrap; }
.ledger-amount.plus { color: var(--teal); }
.ledger-amount.minus { color: var(--accent); }
.ledger-balance { display: block; font-size: 11px; font-weight: 400; text-align: right; }

@media (max-width: 680px) {
  .coin-stats { grid-template-columns: 1fr; }
  .tabs { overflow-x: auto; }
}
</style>
