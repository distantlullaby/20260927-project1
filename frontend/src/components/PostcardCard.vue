<template>
  <article class="pc" :class="{ 'is-expanded': expanded, 'is-settled': story.status === 'settled' }">
    <!-- 卡片正面：求看需求摘要 -->
    <header class="pc-front" @click="expanded = !expanded">
      <div class="pc-front-main">
        <div class="pc-tags">
          <span class="badge" :class="story.status === 'open' ? 'badge-open' : 'badge-settled'">
            {{ story.status === 'open' ? '等待代看' : '记忆已归还' }}
          </span>
          <span v-if="story.city" class="pc-city">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="currentColor"><path d="M12 2a7 7 0 0 0-7 7c0 5 7 13 7 13s7-8 7-13a7 7 0 0 0-7-7zm0 9.5A2.5 2.5 0 1 1 12 6.5a2.5 2.5 0 0 1 0 5z"/></svg>
            {{ story.city }}
          </span>
        </div>
        <h3 class="pc-title">{{ story.title }}</h3>
        <p v-if="story.location" class="pc-location">{{ story.location }}</p>
        <div class="pc-front-foot">
          <div class="pc-author">
            <span class="pc-avatar">{{ (story.author.nickname || '?').slice(0, 1) }}</span>
            <span>{{ story.author.nickname }}</span>
            <span class="pc-date">· {{ story.createdAt }}</span>
          </div>
          <div class="pc-front-meta">
            <span v-if="story.responses.length" class="pc-resp-count">{{ story.responses.length }} 人替 TA 去拍</span>
            <span class="coin pc-reward">{{ story.reward }}</span>
          </div>
        </div>
      </div>
      <div class="pc-cover" v-if="story.oldPhotos && story.oldPhotos.length">
        <img :src="story.oldPhotos[0]" alt="老照片" />
        <div class="pc-cover-stamp">过去</div>
      </div>
      <span class="pc-toggle">{{ expanded ? '收起' : '展开明信片' }}
        <svg viewBox="0 0 24 24" width="14" height="14" :style="{ transform: expanded ? 'rotate(180deg)' : '' }"><path d="M7 10l5 5 5-5" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      </span>
    </header>

    <!-- 展开：双面明信片 -->
    <transition name="expand">
      <div v-if="expanded" class="pc-body">
        <div class="pc-side-switch">
          <button :class="{ active: side === 'past' }" @click="side = 'past'">
            <span class="side-dot past-dot"></span>过去 · 回忆与老照片
          </button>
          <button :class="{ active: side === 'present' }" @click="side = 'present'">
            <span class="side-dot present-dot"></span>当下 · 现场与新照
            <span v-if="story.responses.length" class="side-count">{{ story.responses.length }}</span>
          </button>
        </div>

        <!-- 过去面 -->
        <section v-show="side === 'past'" class="pc-face face-past">
          <p class="pc-memory">{{ story.memory }}</p>
          <div v-if="story.oldPhotos && story.oldPhotos.length" class="pc-photos">
            <img v-for="p in story.oldPhotos" :key="p" :src="p" alt="过去的老照片" @click="$emit('preview', p)" />
          </div>
          <div v-if="story.wish" class="pc-wish">
            <span class="wish-label">想请你拍</span>
            <p>{{ story.wish }}</p>
          </div>
        </section>

        <!-- 当下面 -->
        <section v-show="side === 'present'" class="pc-face face-present">
          <div v-if="!story.responses.length" class="no-response">
            <span class="empty-emoji">📷</span>
            <p>还没有人替 TA 去拍这个角落</p>
            <p v-if="!isOwner" class="no-response-sub">如果你刚好路过，帮 TA 看一眼吧</p>
          </div>

          <div v-else class="resp-list">
            <div v-for="r in story.responses" :key="r.id" class="resp-item" :class="{ accepted: r.accepted }">
              <div class="resp-head">
                <span class="pc-avatar sage">{{ (r.user.nickname || '?').slice(0, 1) }}</span>
                <strong>{{ r.user.nickname }}</strong>
                <span class="pc-date">{{ r.createdAt }}</span>
                <span v-if="r.accepted" class="badge badge-settled accepted-badge">已被采纳 · 记忆归还</span>
              </div>
              <p class="resp-message">{{ r.message }}</p>
              <div v-if="r.newPhotos && r.newPhotos.length" class="pc-photos new">
                <img v-for="p in r.newPhotos" :key="p" :src="p" alt="现场新照片" @click="$emit('preview', p)" />
              </div>
              <div v-if="canAccept(r)" class="resp-actions">
                <button class="btn-sage" :disabled="accepting === r.id" @click="$emit('accept', r)">
                  {{ accepting === r.id ? '结算中…' : `确认是 TA，支付 ${story.reward} 硬币` }}
                </button>
              </div>
            </div>
          </div>
        </section>

        <!-- 底部操作条 -->
        <footer class="pc-actions">
          <template v-if="!auth.user">
            <RouterLink to="/login" class="btn-primary action-login">登录后替他去拍</RouterLink>
          </template>
          <template v-else>
            <button v-if="isOwner && story.status === 'open'" class="btn-ghost" @click="$emit('append', story)">
              追加悬赏
            </button>
            <button
              v-if="!isOwner && story.status === 'open' && !story.hasMine"
              class="btn-sage"
              @click="$emit('respond', story)"
            >
              替他去拍
            </button>
            <span v-if="!isOwner && story.status === 'open' && story.hasMine" class="mine-hint">
              你已上传现场，等待 TA 确认
            </span>
            <span v-if="isOwner && story.status === 'settled'" class="mine-hint settled-hint">
              这份记忆已经归还，硬币已结算
            </span>
            <span v-if="!isOwner && story.status === 'settled'" class="mine-hint">
              悬赏已结算，感谢所有路过的人
            </span>
          </template>
        </footer>
      </div>
    </transition>
  </article>
</template>

<script setup>
import { ref, computed } from 'vue'
import { RouterLink } from 'vue-router'
import { auth } from '../store/auth.js'

const props = defineProps({
  story: { type: Object, required: true }
})
defineEmits(['respond', 'append', 'accept', 'preview'])

const expanded = ref(false)
const side = ref('past')
const accepting = ref(null)

const isOwner = computed(() => auth.user && props.story.author.id === auth.user.id)

function canAccept(r) {
  return isOwner.value && props.story.status === 'open' && r.user.id !== auth.user.id
}

defineExpose({ setAccepting: (v) => (accepting.value = v), resetSide: () => {} })
</script>

<style scoped>
.pc {
  background: var(--card);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  overflow: hidden;
  transition: box-shadow 0.2s ease;
}
.pc.is-expanded {
  box-shadow: 0 16px 40px rgba(94, 70, 40, 0.18);
}

/* 正面 */
.pc-front {
  display: grid;
  grid-template-columns: 1fr 132px;
  grid-template-areas: 'main cover';
  gap: 16px;
  padding: 20px 20px;
  cursor: pointer;
  position: relative;
}
.pc-front-main {
  grid-area: main;
  min-width: 0;
}
.pc-cover {
  grid-area: cover;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
  height: 132px;
  box-shadow: 0 4px 12px rgba(94, 70, 40, 0.18);
  transform: rotate(1.5deg);
}
.pc-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.pc-cover-stamp {
  position: absolute;
  left: 6px;
  bottom: 6px;
  background: rgba(74, 60, 44, 0.72);
  color: #f6efe3;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 999px;
  letter-spacing: 2px;
}
.pc-tags {
  display: flex;
  align-items: center;
  gap: 10px;
}
.pc-city {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  font-size: 12px;
  color: var(--ink-soft);
}
.pc-title {
  font-size: 18px;
  margin: 9px 0 5px;
  color: var(--ink);
}
.pc-location {
  font-size: 13px;
  color: var(--ink-soft);
}
.pc-front-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 14px;
  gap: 10px;
}
.pc-author {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 13px;
  color: var(--ink-soft);
}
.pc-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--accent), var(--accent-deep));
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
}
.pc-avatar.sage {
  background: linear-gradient(135deg, var(--sage), var(--sage-deep));
}
.pc-date {
  color: #b3a48c;
  font-size: 12px;
}
.pc-front-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}
.pc-resp-count {
  font-size: 12px;
  color: var(--sage-deep);
  font-weight: 600;
}
.pc-reward {
  font-size: 16px;
  background: rgba(217, 164, 65, 0.14);
  padding: 4px 12px;
  border-radius: 999px;
}
.pc-toggle {
  position: absolute;
  top: 18px;
  right: 20px;
  font-size: 12px;
  color: var(--accent-deep);
  display: flex;
  align-items: center;
  gap: 2px;
}
.pc.has-cover .pc-toggle {
  right: 150px;
}

/* 展开区域 */
.pc-body {
  border-top: 1px dashed var(--line);
  padding: 0 20px 18px;
}
.pc-side-switch {
  display: flex;
  gap: 8px;
  padding: 14px 0;
}
.pc-side-switch button {
  flex: 1;
  background: var(--paper);
  color: var(--ink-soft);
  border-radius: 10px;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
.pc-side-switch button.active {
  background: var(--accent);
  color: #fff;
}
.side-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.past-dot {
  background: #c9a86a;
}
.present-dot {
  background: var(--sage);
}
.pc-side-switch button.active .past-dot,
.pc-side-switch button.active .present-dot {
  background: #fff;
}
.side-count {
  background: rgba(255, 255, 255, 0.25);
  border-radius: 999px;
  font-size: 11px;
  padding: 0 7px;
}

.pc-face {
  padding: 4px 2px 8px;
}
.face-past {
  background: linear-gradient(180deg, rgba(217, 196, 150, 0.16), rgba(217, 196, 150, 0.04));
  border-radius: 12px;
  padding: 18px;
}
.face-present {
  padding-top: 8px;
}
.pc-memory {
  line-height: 1.9;
  font-size: 14.5px;
  color: #5d4d38;
  white-space: pre-wrap;
}
.pc-photos {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 14px;
}
.pc-photos img {
  width: 150px;
  height: 112px;
  object-fit: cover;
  border-radius: 10px;
  cursor: zoom-in;
  box-shadow: 0 3px 8px rgba(94, 70, 40, 0.16);
  transition: transform 0.15s ease;
}
.pc-photos img:hover {
  transform: scale(1.03);
}
.pc-photos.new img {
  border: 2px solid rgba(111, 146, 121, 0.35);
}
.pc-wish {
  margin-top: 16px;
  background: rgba(255, 253, 248, 0.7);
  border-left: 3px solid var(--accent);
  border-radius: 0 10px 10px 0;
  padding: 12px 14px;
}
.wish-label {
  font-size: 12px;
  font-weight: 700;
  color: var(--accent-deep);
  letter-spacing: 1px;
}
.pc-wish p {
  margin-top: 4px;
  font-size: 14px;
  line-height: 1.7;
}

/* 回应 */
.no-response {
  text-align: center;
  padding: 34px 16px;
  color: var(--ink-soft);
}
.no-response p {
  font-size: 14px;
}
.no-response-sub {
  margin-top: 6px;
  font-size: 12.5px;
  color: #b3a48c;
}
.resp-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.resp-item {
  background: linear-gradient(180deg, rgba(111, 146, 121, 0.08), rgba(111, 146, 121, 0.02));
  border: 1px solid rgba(111, 146, 121, 0.2);
  border-radius: 12px;
  padding: 15px 16px;
}
.resp-item.accepted {
  border-color: rgba(217, 164, 65, 0.5);
  background: linear-gradient(180deg, rgba(217, 164, 65, 0.1), rgba(217, 164, 65, 0.03));
}
.resp-head {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13.5px;
}
.accepted-badge {
  margin-left: auto;
}
.resp-message {
  margin: 10px 0 0;
  font-size: 14px;
  line-height: 1.85;
  color: #445a4c;
  white-space: pre-wrap;
}
.resp-actions {
  margin-top: 13px;
}

.pc-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  padding-top: 14px;
  margin-top: 4px;
  border-top: 1px dashed var(--line);
}
.action-login {
  padding: 9px 22px;
}
.mine-hint {
  font-size: 13px;
  color: var(--ink-soft);
}
.settled-hint {
  color: #8a6320;
}

.expand-enter-active,
.expand-leave-active {
  transition: opacity 0.22s ease;
}
.expand-enter-from,
.expand-leave-to {
  opacity: 0;
}

@media (max-width: 560px) {
  .pc-front {
    grid-template-columns: 1fr;
    grid-template-areas: 'cover' 'main';
  }
  .pc-cover {
    height: 170px;
  }
  .pc-toggle {
    top: auto;
    bottom: 16px;
  }
}
</style>
