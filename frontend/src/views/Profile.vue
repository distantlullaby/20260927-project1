<template>
  <div class="profile">
    <!-- 硬币总览 -->
    <section class="coin-panel">
      <div class="coin-card main-coin">
        <p class="coin-label">可用记忆硬币</p>
        <div class="coin-big">{{ me?.user.balance ?? auth.user?.balance ?? 0 }}</div>
        <p class="coin-sub">发布求看时冻结，确认代看后结算给对方</p>
      </div>
      <div class="coin-side">
        <div class="coin-card mini">
          <span class="mini-label">冻结中</span>
          <span class="coin">{{ me?.user.frozen ?? auth.user?.frozen ?? 0 }}</span>
        </div>
        <div class="coin-card mini">
          <span class="mini-label">已发布</span>
          <strong>{{ me?.published ?? 0 }}</strong> 条
        </div>
        <div class="coin-card mini">
          <span class="mini-label">已归还</span>
          <strong>{{ me?.settled ?? 0 }}</strong> 条
        </div>
        <div class="coin-card mini">
          <span class="mini-label">被采纳</span>
          <strong>{{ me?.helped ?? 0 }}</strong> 次
        </div>
      </div>
    </section>

    <!-- 切换标签 -->
    <div class="tabs">
      <button v-for="t in tabs" :key="t.key" class="tab" :class="{ active: tab === t.key }" @click="tab = t.key">
        {{ t.label }}
      </button>
    </div>

    <!-- 硬币流水 -->
    <section v-if="tab === 'ledger' && ledgers.length" class="panel">
      <div v-for="l in ledgers" :key="l.id" class="ledger-row">
        <div class="ledger-icon" :class="l.type">{{ iconFor(l.type) }}</div>
        <div class="ledger-info">
          <p class="ledger-type">{{ l.typeText }}</p>
          <p class="ledger-remark">{{ l.remark }} · {{ l.createdAt }}</p>
        </div>
        <div class="ledger-amount" :class="amountClass(l)">
          {{ formatChange(l.change) }}
          <span class="ledger-balance">余额 {{ l.balanceAfter }} / 冻 {{ l.frozenAfter }}</span>
        </div>
      </div>
    </section>

    <!-- 我发布的 -->
    <section v-else-if="tab === 'mine'" class="panel">
      <div v-if="!myStories.length" class="empty"><span class="empty-emoji">✉️</span>你还没有发布过求看</div>
      <RouterLink v-for="s in myStories" :key="s.id" to="/" class="record-row">
        <div>
          <p class="record-title">{{ s.title }}</p>
          <p class="record-meta">{{ s.city || '未填城市' }} · {{ s.responseCount }} 人回应 · {{ s.createdAt }}</p>
        </div>
        <div class="record-right">
          <span class="coin">{{ s.reward }}</span>
          <span class="badge" :class="s.status === 'open' ? 'badge-open' : 'badge-settled'">
            {{ s.status === 'open' ? '等待中' : '已归还' }}
          </span>
        </div>
      </RouterLink>
    </section>

    <!-- 我代看的 -->
    <section v-else-if="tab === 'helped'" class="panel">
      <div v-if="!myResponses.length" class="empty"><span class="empty-emoji">📷</span>你还没有替谁去拍过现场</div>
      <RouterLink v-for="r in myResponses" :key="r.responseId" to="/" class="record-row">
        <div>
          <p class="record-title">{{ r.title }}</p>
          <p class="record-meta">{{ r.city || '未知城市' }} · {{ r.author.nickname }} 的回忆 · {{ r.createdAt }}</p>
        </div>
        <div class="record-right">
          <span class="coin">{{ r.reward }}</span>
          <span class="badge" :class="r.accepted ? 'badge-settled' : 'badge-open'">
            {{ r.accepted ? '已采纳' : r.status === 'settled' ? '未被采纳' : '待确认' }}
          </span>
        </div>
      </RouterLink>
    </section>

    <section v-else class="panel">
      <div class="empty"><span class="empty-emoji">🪙</span>暂无记录</div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { api } from '../api/index.js'
import { auth } from '../store/auth.js'

const me = ref(null)
const ledgers = ref([])
const myStories = ref([])
const myResponses = ref([])
const tab = ref('ledger')

const tabs = [
  { key: 'ledger', label: '硬币流水' },
  { key: 'mine', label: '我发布的求看' },
  { key: 'helped', label: '我代看的现场' }
]

const icons = {
  register: '🎁',
  freeze: '🔒',
  append: '🔒',
  pay: '📮',
  income: '◉'
}
function iconFor(type) {
  return icons[type] || '·'
}
function formatChange(change) {
  if (change === 0) return '结算'
  return (change > 0 ? '+' : '') + change
}
function amountClass(l) {
  if (l.type === 'income') return 'plus'
  if (l.change < 0) return 'minus'
  return ''
}

onMounted(async () => {
  try {
    const [a, b, c, d] = await Promise.all([
      api.get('/me'),
      api.get('/me/ledgers'),
      api.get('/me/stories'),
      api.get('/me/responses')
    ])
    me.value = a
    ledgers.value = b.ledgers
    myStories.value = c.stories
    myResponses.value = d.responses
  } catch (e) {
    /* 路由守卫已保证登录 */
  }
})
</script>

<style scoped>
.coin-panel {
  display: grid;
  grid-template-columns: 1.3fr 1fr;
  gap: 16px;
}
.coin-card {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 24px;
}
.main-coin {
  background:
    radial-gradient(circle at 85% 15%, rgba(217, 164, 65, 0.25), transparent 55%),
    linear-gradient(140deg, #5a4632, #7a5e3c);
  color: #f8efdd;
  border: none;
}
.coin-label {
  font-size: 14px;
  opacity: 0.85;
  letter-spacing: 1px;
}
.coin-big {
  font-size: 52px;
  font-weight: 800;
  line-height: 1.2;
  margin: 8px 0 6px;
  color: #f3cf7e;
}
.coin-sub {
  font-size: 12.5px;
  opacity: 0.75;
}
.coin-side {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}
.coin-card.mini {
  display: flex;
  flex-direction: column;
  gap: 8px;
  justify-content: center;
}
.mini-label {
  font-size: 12.5px;
  color: var(--ink-soft);
}

.tabs {
  display: flex;
  gap: 8px;
  margin: 24px 0 14px;
}
.tab {
  background: var(--card);
  border: 1.5px solid var(--line);
  color: var(--ink-soft);
  font-size: 13.5px;
  padding: 8px 20px;
}
.tab.active {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}

.panel {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  overflow: hidden;
}

/* 流水 */
.ledger-row {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 15px 20px;
  border-bottom: 1px solid var(--paper-deep);
}
.ledger-row:last-child {
  border-bottom: none;
}
.ledger-icon {
  width: 38px;
  height: 38px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 17px;
  background: var(--paper-deep);
  flex-shrink: 0;
}
.ledger-icon.income {
  background: rgba(217, 164, 65, 0.2);
}
.ledger-icon.freeze,
.ledger-icon.append {
  background: rgba(122, 106, 133, 0.18);
}
.ledger-icon.pay {
  background: rgba(111, 146, 121, 0.2);
}
.ledger-info {
  flex: 1;
  min-width: 0;
}
.ledger-type {
  font-size: 14px;
  font-weight: 700;
}
.ledger-remark {
  font-size: 12.5px;
  color: var(--ink-soft);
  margin-top: 3px;
}
.ledger-amount {
  text-align: right;
  font-weight: 800;
  font-size: 16px;
  color: var(--ink);
}
.ledger-amount.plus {
  color: #9a6b1e;
}
.ledger-amount.minus {
  color: #7a6a85;
}
.ledger-balance {
  display: block;
  font-size: 11.5px;
  font-weight: 400;
  color: #b3a48c;
  margin-top: 3px;
}

/* 记录行 */
.record-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 17px 20px;
  border-bottom: 1px solid var(--paper-deep);
}
.record-row:last-child {
  border-bottom: none;
}
.record-row:hover {
  background: var(--paper);
}
.record-title {
  font-size: 14.5px;
  font-weight: 600;
}
.record-meta {
  font-size: 12.5px;
  color: var(--ink-soft);
  margin-top: 4px;
}
.record-right {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
}

@media (max-width: 640px) {
  .coin-panel {
    grid-template-columns: 1fr;
  }
}
</style>
