<script setup>
import { ref, provide } from 'vue'
import { RouterLink, RouterView } from 'vue-router'
import { auth } from './auth'
import { useRouter } from 'vue-router'

const router = useRouter()
const toast = ref(null)
const lightbox = ref('')

function showToast(msg, type = '') {
  toast.value = { msg, type }
  setTimeout(() => (toast.value = null), 2600)
}
provide('toast', showToast)
provide('lightbox', (url) => (lightbox.value = url))

function logout() {
  auth.logout()
  router.push('/')
  showToast('已退出登录')
}
</script>

<template>
  <div class="layout">
    <header class="nav">
      <div class="nav-inner">
        <RouterLink to="/" class="brand">
          <span class="brand-mark">✉</span>
          <span>记忆连接</span>
        </RouterLink>
        <p class="slogan">你帮我再看一眼，我把记忆还给你</p>
        <nav class="nav-links">
          <RouterLink to="/">求看广场</RouterLink>
          <template v-if="auth.isLogin">
            <RouterLink to="/profile">个人中心</RouterLink>
            <span class="nav-coin coin">{{ auth.user?.balance ?? 0 }}</span>
            <button class="btn-ghost btn-sm" @click="logout">退出</button>
          </template>
          <RouterLink v-else to="/login" class="btn-primary btn-login">登录 / 注册</RouterLink>
        </nav>
      </div>
    </header>

    <main class="container">
      <RouterView />
    </main>

    <footer class="footer">
      记忆连接 MemoryLink · 让远方的人，替你再看一眼回忆的角落
    </footer>

    <div v-if="toast" class="toast" :class="toast.type">{{ toast.msg }}</div>

    <div v-if="lightbox" class="lightbox" @click="lightbox = ''">
      <img :src="lightbox" alt="照片大图" />
    </div>
  </div>
</template>

<style scoped>
.layout { min-height: 100%; display: flex; flex-direction: column; }

.nav {
  position: sticky;
  top: 0;
  z-index: 50;
  background: rgba(247, 241, 230, .92);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--line);
}
.nav-inner {
  max-width: 920px;
  margin: 0 auto;
  padding: 12px 20px;
  display: flex;
  align-items: center;
  gap: 18px;
}
.brand {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 20px;
  font-weight: 700;
  color: var(--ink);
}
.brand-mark { color: var(--accent); font-size: 22px; }
.slogan {
  flex: 1;
  font-size: 13px;
  color: var(--ink-soft);
  letter-spacing: .5px;
}
.nav-links { display: flex; align-items: center; gap: 16px; font-size: 14px; }
.nav-links a { color: var(--ink-soft); }
.nav-links a.router-link-exact-active,
.nav-links a:hover { color: var(--accent); }
.btn-sm { padding: 5px 14px; font-size: 13px; }
.btn-login { padding: 7px 18px; font-size: 13px; border-radius: 8px; }
.nav-coin { font-size: 13px; }

.container {
  flex: 1;
  width: 100%;
  max-width: 920px;
  margin: 0 auto;
  padding: 26px 20px 50px;
}

.footer {
  text-align: center;
  padding: 22px;
  font-size: 12px;
  color: var(--ink-soft);
  border-top: 1px solid var(--line);
}

.lightbox {
  position: fixed;
  inset: 0;
  background: rgba(30, 24, 18, .85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  cursor: zoom-out;
  padding: 30px;
}
.lightbox img {
  max-width: 100%;
  max-height: 100%;
  border-radius: 8px;
  box-shadow: 0 20px 60px rgba(0,0,0,.5);
}

@media (max-width: 720px) {
  .slogan { display: none; }
}
</style>
