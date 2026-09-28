<script setup>
import { computed, ref, inject } from 'vue'
import { auth } from '../auth'

const props = defineProps({
  story: { type: Object, required: true }
})
const emit = defineEmits(['respond', 'accept', 'append'])

const open = ref(false)
const showLightbox = inject('lightbox')

const oldPhotos = computed(() => splitCsv(props.story.old_photos))
const accepted = computed(() =>
  props.story.responses?.find((r) => r.id === props.story.accepted_response_id) ||
  props.story.responses?.find((r) => r.status === 'accepted')
)
const pending = computed(() =>
  (props.story.responses || []).filter((r) => r.status === 'pending')
)
const isOwner = computed(() => props.story.user_id === auth.user?.id)
const iResponded = computed(() =>
  (props.story.responses || []).some((r) => r.user_id === auth.user?.id)
)

function splitCsv(s) {
  return (s || '').split(',').map((x) => x.trim()).filter(Boolean)
}

function fmtTime(t) {
  return new Date(t).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' })
}
</script>

<template>
  <article class="postcard card">
    <!-- 卡片头部：邮戳风格 -->
    <div class="pc-head">
      <div class="pc-meta">
        <span class="tag">📍 {{ story.city }} · {{ story.location }}</span>
        <span :class="['tag', story.status === 'open' ? 'tag-open' : 'tag-done']">
          {{ story.status === 'open' ? '求看中' : '记忆已归还' }}
        </span>
      </div>
      <div class="pc-bounty coin">{{ story.bounty }}</div>
    </div>

    <h3 class="pc-title">{{ story.title }}</h3>

    <div class="pc-author">
      <span class="avatar">{{ (story.user?.nickname || '?')[0] }}</span>
      <span>{{ story.user?.nickname }}</span>
      <span class="muted">· {{ fmtTime(story.created_at) }}</span>
    </div>

    <p class="pc-teaser">{{ story.memory_text }}</p>

    <div v-if="oldPhotos.length" class="pc-thumb-row">
      <img v-for="(p, i) in oldPhotos.slice(0, 3)" :key="i" :src="p"
           :alt="'老照片' + (i + 1)" @click="showLightbox(p)" />
    </div>

    <div class="pc-actions">
      <button class="btn-ghost btn-sm" @click="open = !open">
        {{ open ? '合上明信片' : (story.status === 'fulfilled' ? '展开双面明信片' : '展开回忆 / 去现场看看') }}
        <span>{{ open ? '▴' : '▾' }}</span>
      </button>
      <template v-if="open && story.status === 'open'">
        <button v-if="!auth.isLogin" class="btn-primary btn-sm"
                @click="$router.push('/login')">登录后代看</button>
        <button v-else-if="isOwner" class="btn-ghost btn-sm" @click="emit('append', story)">
          ＋ 追加悬赏
        </button>
        <button v-else-if="!iResponded" class="btn-teal btn-sm" @click="emit('respond', story)">
          📷 替他去拍
        </button>
        <span v-else class="muted">你已提交现场记录，等待确认</span>
      </template>
    </div>

    <!-- 双面明信片 -->
    <transition name="expand">
      <div v-if="open" class="pc-sides">
        <div class="divider">
          <span>双面明信片</span>
        </div>

        <div class="sides">
          <!-- 正面：过去回忆 -->
          <section class="side side-old">
            <div class="side-label">过去回忆</div>
            <p class="memory-text">{{ story.memory_text }}</p>
            <div v-if="oldPhotos.length" class="photo-grid">
              <img v-for="(p, i) in oldPhotos" :key="i" :src="p"
                   alt="老照片" @click="showLightbox(p)" />
            </div>
          </section>

          <!-- 背面：当下现场 -->
          <section class="side side-now">
            <div class="side-label">当下现场</div>

            <div v-if="accepted" class="now-block accepted">
              <div class="now-head">
                <span class="avatar sm">{{ (accepted.user?.nickname || '?')[0] }}</span>
                <strong>{{ accepted.user?.nickname }}</strong>
                <span class="tag tag-done">已采纳</span>
              </div>
              <p class="now-text">{{ accepted.now_text }}</p>
              <div v-if="splitCsv(accepted.new_photos).length" class="photo-grid">
                <img v-for="(p, i) in splitCsv(accepted.new_photos)" :key="i"
                     :src="p" alt="现场新照" @click="showLightbox(p)" />
              </div>
              <div v-if="accepted.message" class="message-box">
                <span class="msg-label">寄语</span>
                {{ accepted.message }}
              </div>
              <div class="settled">◉ {{ story.bounty }} 枚记忆硬币已结算给 {{ accepted.user?.nickname }}</div>
            </div>

            <template v-else-if="pending.length">
              <div v-for="r in pending" :key="r.id" class="now-block">
                <div class="now-head">
                  <span class="avatar sm">{{ (r.user?.nickname || '?')[0] }}</span>
                  <strong>{{ r.user?.nickname }}</strong>
                  <span class="tag tag-open">待确认</span>
                </div>
                <p class="now-text">{{ r.now_text }}</p>
                <div v-if="splitCsv(r.new_photos).length" class="photo-grid">
                  <img v-for="(p, i) in splitCsv(r.new_photos)" :key="i"
                       :src="p" alt="现场新照" @click="showLightbox(p)" />
                </div>
                <div v-if="r.message" class="message-box">
                  <span class="msg-label">寄语</span>{{ r.message }}
                </div>
                <button v-if="isOwner" class="btn-primary btn-sm" @click="emit('accept', r)">
                  确认采纳 · 结算 {{ story.bounty }} ◉
                </button>
              </div>
            </template>

            <div v-else class="now-empty">
              <div class="stamp-dashed">等一张来自现场的新照片</div>
              <p class="muted">还没有人替他去看这个角落</p>
              <button v-if="auth.isLogin && !isOwner" class="btn-teal btn-sm"
                      style="margin-top:10px" @click="emit('respond', story)">
                📷 替他去拍
              </button>
            </div>
          </section>
        </div>
      </div>
    </transition>
  </article>
</template>

<style scoped>
.postcard {
  padding: 20px 22px;
  margin-bottom: 18px;
  position: relative;
  background:
    linear-gradient(90deg, rgba(194,104,63,.05), transparent 60%),
    var(--card);
}
.pc-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 10px;
}
.pc-meta { display: flex; flex-wrap: wrap; gap: 6px; }
.pc-bounty { font-size: 15px; }

.pc-title {
  font-size: 18px;
  margin: 12px 0 8px;
  font-weight: 700;
}
.pc-author {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 13px;
  margin-bottom: 10px;
}
.avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: var(--teal);
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
}
.avatar.sm { width: 22px; height: 22px; font-size: 11px; }

.pc-teaser {
  color: var(--ink-soft);
  font-size: 14px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.pc-thumb-row {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}
.pc-thumb-row img {
  width: 88px;
  height: 62px;
  object-fit: cover;
  border-radius: 6px;
  border: 1px solid var(--line);
  cursor: zoom-in;
  filter: sepia(.25);
}
.pc-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 14px;
  flex-wrap: wrap;
}

/* 双面明信片 */
.divider {
  text-align: center;
  margin: 18px 0 14px;
  position: relative;
}
.divider::before {
  content: "";
  position: absolute;
  left: 0; right: 0; top: 50%;
  border-top: 1px dashed var(--line);
}
.divider span {
  position: relative;
  background: var(--card);
  padding: 0 14px;
  font-size: 12px;
  color: var(--ink-soft);
  letter-spacing: 3px;
}
.sides {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}
.side {
  border-radius: 12px;
  padding: 16px;
}
.side-old {
  background: linear-gradient(160deg, #f3e9d2, #ece0c6);
  border: 1px solid #e0d2b4;
}
.side-now {
  background: linear-gradient(160deg, #e9f0ec, #dfe9e3);
  border: 1px solid #cddcd2;
}
.side-label {
  font-size: 12px;
  letter-spacing: 4px;
  color: var(--ink-soft);
  margin-bottom: 10px;
}
.memory-text, .now-text {
  font-size: 14px;
  white-space: pre-wrap;
}
.side-old .photo-grid img { filter: sepia(.5) contrast(.95); }

.now-block {
  padding-bottom: 14px;
  margin-bottom: 14px;
  border-bottom: 1px dashed #bccdc4;
}
.now-block:last-child { border-bottom: none; margin-bottom: 0; padding-bottom: 0; }
.now-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-size: 13px;
}
.message-box {
  margin-top: 10px;
  background: rgba(255,253,248,.8);
  border-left: 3px solid var(--gold);
  padding: 9px 12px;
  border-radius: 0 8px 8px 0;
  font-size: 13px;
  color: var(--ink);
}
.msg-label {
  display: block;
  font-size: 11px;
  color: var(--ink-soft);
  letter-spacing: 2px;
  margin-bottom: 2px;
}
.settled {
  margin-top: 10px;
  font-size: 13px;
  color: #9a7a2e;
  font-weight: 600;
}
.now-empty { text-align: center; padding: 18px 0; }
.stamp-dashed {
  display: inline-block;
  border: 1.5px dashed var(--teal);
  color: var(--teal);
  border-radius: 8px;
  padding: 4px 14px;
  font-size: 12px;
  margin-bottom: 10px;
}

.expand-enter-active, .expand-leave-active { transition: all .3s ease; overflow: hidden; }
.expand-enter-from, .expand-leave-to { opacity: 0; max-height: 0; }
.expand-enter-to, .expand-leave-from { opacity: 1; }

@media (max-width: 680px) {
  .sides { grid-template-columns: 1fr; }
}
</style>
