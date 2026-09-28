<script setup>
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import { auth } from '../auth'

const router = useRouter()
const route = useRoute()
const mode = ref('login')
const username = ref('')
const password = ref('')
const nickname = ref('')
const loading = ref(false)

async function submit() {
  if (!username.value || !password.value) return
  loading.value = true
  try {
    const url = mode.value === 'login' ? '/auth/login' : '/auth/register'
    const payload = { username: username.value, password: password.value }
    if (mode.value === 'register') payload.nickname = nickname.value
    const { token, user } = await api.post(url, payload)
    auth.setSession(token, user)
    router.push(route.query.redirect || '/')
  } catch (e) {
    alert(e.message)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-wrap">
    <div class="auth-card card">
      <div class="auth-head">
        <div class="stamp">记忆硬币 · ◉100</div>
        <h2>{{ mode === 'login' ? '欢迎回来' : '寄出你的第一张明信片' }}</h2>
        <p class="muted">
          {{ mode === 'login'
            ? '登录后查看远方寄来的当下'
            : '注册即赠送 100 枚记忆硬币，发布求看、替人代看，让记忆流动起来' }}
        </p>
      </div>

      <form @submit.prevent="submit">
        <div class="form-row">
          <label>用户名</label>
          <input class="input" v-model="username" placeholder="给自己取个名字" autocomplete="username" />
        </div>
        <div v-if="mode === 'register'" class="form-row">
          <label>昵称（可选）</label>
          <input class="input" v-model="nickname" placeholder="例如：离家的阿栀" />
        </div>
        <div class="form-row">
          <label>密码</label>
          <input class="input" type="password" v-model="password"
                 placeholder="至少 6 位" autocomplete="current-password" />
        </div>
        <button class="btn-primary auth-btn" :disabled="loading">
          {{ loading ? '请稍候…' : (mode === 'login' ? '登录' : '注册并领取硬币') }}
        </button>
      </form>

      <p class="switch">
        {{ mode === 'login' ? '还没有账号？' : '已经有账号了？' }}
        <a href="javascript:void(0)" @click="mode = mode === 'login' ? 'register' : 'login'">
          {{ mode === 'login' ? '注册新账号' : '去登录' }}
        </a>
      </p>

      <div class="demo muted">
        演示账号：alice / bob / carol，密码均为 123456
      </div>
    </div>
  </div>
</template>

<style scoped>
.auth-wrap {
  min-height: 70vh;
  display: flex;
  align-items: center;
  justify-content: center;
}
.auth-card {
  width: 100%;
  max-width: 420px;
  padding: 34px 32px 26px;
  background:
    linear-gradient(180deg, rgba(201,162,79,.08), transparent 120px),
    var(--card);
}
.auth-head { text-align: center; margin-bottom: 22px; }
.auth-head h2 { font-size: 22px; margin: 14px 0 8px; }
.stamp {
  display: inline-block;
  border: 1.5px dashed var(--gold);
  color: #9a7a2e;
  border-radius: 8px;
  padding: 3px 12px;
  font-size: 12px;
  letter-spacing: 1px;
}
.auth-btn { width: 100%; padding: 12px; font-size: 15px; margin-top: 6px; }
.switch { text-align: center; margin-top: 16px; font-size: 14px; }
.demo {
  margin-top: 18px;
  text-align: center;
  font-size: 12px;
  padding-top: 14px;
  border-top: 1px dashed var(--line);
}
</style>
