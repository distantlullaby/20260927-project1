<template>
  <div class="app-shell">
    <header class="topbar">
      <div class="topbar-inner">
        <RouterLink to="/" class="brand" @click="goFeed">
          <span class="brand-mark">
            <svg viewBox="0 0 24 24" width="26" height="26" aria-hidden="true">
              <path d="M4 6.5C4 5.7 4.7 5 5.5 5h13c.8 0 1.5.7 1.5 1.5v11c0 .8-.7 1.5-1.5 1.5h-13C4.7 19 4 18.3 4 17.5v-11z" fill="#fffdf8" stroke="#c2703d" stroke-width="1.4"/>
              <path d="M8 4.5v15M16 4.5v15" stroke="#e0d2b8" stroke-width="1.2"/>
              <circle cx="12" cy="10.5" r="2.1" fill="#d9a441"/>
              <path d="M8.4 15.5c1.1-1.8 6-1.8 7.2 0" fill="none" stroke="#6f9279" stroke-width="1.3" stroke-linecap="round"/>
            </svg>
          </span>
          <span class="brand-name">记忆连接</span>
        </RouterLink>

        <p class="slogan">你帮我再看一眼，我把记忆还给你</p>

        <nav class="nav">
          <template v-if="auth.user">
            <RouterLink to="/" class="nav-link">求看广场</RouterLink>
            <RouterLink to="/profile" class="nav-link nav-profile">
              <span class="coin">{{ auth.user.balance }}</span>
              <span class="nav-avatar">{{ avatarText }}</span>
              <span>{{ auth.user.nickname }}</span>
            </RouterLink>
            <button class="btn-ghost nav-logout" @click="logout">退出</button>
          </template>
          <template v-else>
            <RouterLink to="/login" class="nav-link">登录 / 注册</RouterLink>
          </template>
        </nav>
      </div>
    </header>

    <main class="page">
      <RouterView v-slot="{ Component }">
        <transition name="fade" mode="out-in">
          <component :is="Component" @balance-changed="refreshUser" />
        </transition>
      </RouterView>
    </main>

    <footer class="footer">记忆连接 · 让远方的人，替你再看一眼回不去的地方</footer>

    <transition name="fade">
      <div v-if="toastState.visible" class="toast" :class="toastClass">{{ toastState.message }}</div>
    </transition>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter, RouterView, RouterLink } from 'vue-router'
import { auth, clearAuth, updateUser } from './store/auth.js'
import { toastState, toast } from './store/toast.js'
import { api } from './api/index.js'

const router = useRouter()

const avatarText = computed(() => (auth.user?.nickname || '?').slice(0, 1))
const toastClass = computed(() =>
  toastState.type === 'error' ? 'toast-error' : toastState.type === 'success' ? 'toast-success' : ''
)

function goFeed() {
  router.push({ name: 'feed' })
}

function logout() {
  clearAuth()
  toast('已退出登录', 'success')
  router.push({ name: 'login' })
}

// 发帖 / 追加 / 结算后刷新顶栏的硬币余额
async function refreshUser() {
  if (!auth.token) return
  try {
    const data = await api.get('/me')
    updateUser(data.user)
  } catch (e) {
    /* 忽略后台刷新失败 */
  }
}
</script>

<style scoped>
.app-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.topbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background: rgba(246, 239, 227, 0.88);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--line);
}
.topbar-inner {
  max-width: 960px;
  margin: 0 auto;
  padding: 14px 20px;
  display: flex;
  align-items: center;
  gap: 18px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 9px;
}
.brand-mark {
  display: inline-flex;
}
.brand-name {
  font-size: 20px;
  font-weight: 800;
  letter-spacing: 2px;
  color: var(--accent-deep);
}
.slogan {
  flex: 1;
  font-size: 13px;
  color: var(--ink-soft);
  font-style: italic;
  letter-spacing: 1px;
}
.nav {
  display: flex;
  align-items: center;
  gap: 14px;
}
.nav-link {
  font-size: 14px;
  color: var(--ink);
  font-weight: 600;
  white-space: nowrap;
}
.nav-link:hover {
  color: var(--accent-deep);
}
.nav-profile {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--paper-deep);
  padding: 6px 14px 6px 12px;
  border-radius: 999px;
}
.nav-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--sage), var(--sage-deep));
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
}
.nav-logout {
  padding: 6px 14px;
  font-size: 13px;
}

.page {
  flex: 1;
  width: 100%;
  max-width: 960px;
  margin: 0 auto;
  padding: 26px 20px 40px;
}
.footer {
  text-align: center;
  color: var(--ink-soft);
  font-size: 12px;
  padding: 22px;
  border-top: 1px solid var(--line);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

@media (max-width: 720px) {
  .slogan {
    display: none;
  }
}
</style>
