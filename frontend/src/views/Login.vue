<template>
  <div class="auth-wrap">
    <div class="auth-card">
      <div class="auth-hero">
        <div class="postcard-art">
          <svg viewBox="0 0 200 150" width="100%" aria-hidden="true">
            <rect x="8" y="8" width="184" height="134" rx="8" fill="#fffdf8" stroke="#e0d2b8" stroke-width="2" />
            <path d="M100 8v134" stroke="#e0d2b8" stroke-width="1.5" stroke-dasharray="5 4" />
            <circle cx="60" cy="58" r="26" fill="#93bd72" opacity="0.85" />
            <rect x="56" y="78" width="8" height="34" fill="#8a5a3b" />
            <path d="M120 50h48M120 66h48M120 82h34" stroke="#cbb89a" stroke-width="4" stroke-linecap="round" />
            <rect x="150" y="98" width="26" height="20" rx="2" fill="none" stroke="#c2703d" stroke-width="2" stroke-dasharray="3 2" />
          </svg>
        </div>
        <h2 class="auth-title">{{ isLogin ? '欢迎回来' : '加入记忆连接' }}</h2>
        <p class="auth-desc">
          {{ isLogin ? '登录后，看看有没有人替你拍下了远方的角落。' : '注册即赠送 100 枚记忆硬币，发布你的第一个求看心愿。' }}
        </p>
      </div>

      <form class="auth-form" @submit.prevent="submit">
        <label class="label">用户名</label>
        <input v-model.trim="form.username" placeholder="2 位以上，登录用" autocomplete="username" />

        <template v-if="!isLogin">
          <label class="label">昵称（选填）</label>
          <input v-model.trim="form.nickname" placeholder="别人会怎么称呼你" maxlength="20" />
        </template>

        <label class="label">密码</label>
        <input v-model="form.password" type="password" placeholder="至少 6 位" autocomplete="current-password" />

        <p v-if="error" class="form-error">{{ error }}</p>

        <button class="btn-primary auth-submit" type="submit" :disabled="loading">
          {{ loading ? '请稍候…' : isLogin ? '登录' : '注册并领取 100 硬币' }}
        </button>
      </form>

      <button class="switch-mode" @click="switchMode">
        {{ isLogin ? '还没有账号？立即注册' : '已有账号？去登录' }}
      </button>

      <div class="demo-hint">
        演示账号：alin / xiaoman / oldchen，密码均为 123456
      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '../api/index.js'
import { setAuth } from '../store/auth.js'
import { toast } from '../store/toast.js'

const router = useRouter()
const route = useRoute()

const isLogin = ref(true)
const loading = ref(false)
const error = ref('')
const form = reactive({ username: '', nickname: '', password: '' })

function switchMode() {
  isLogin.value = !isLogin.value
  error.value = ''
}

async function submit() {
  error.value = ''
  if (form.username.length < 2 || form.password.length < 6) {
    error.value = '用户名至少 2 位，密码至少 6 位'
    return
  }
  loading.value = true
  try {
    const payload = isLogin.value
      ? { username: form.username, password: form.password }
      : { username: form.username, password: form.password, nickname: form.nickname }
    const data = await api.post(isLogin.value ? '/auth/login' : '/auth/register', payload)
    setAuth(data.token, data.user)
    toast(isLogin.value ? '登录成功' : `注册成功，已到账 ${data.user.balance} 枚记忆硬币`, 'success')
    router.replace(route.query.redirect || { name: 'feed' })
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.auth-wrap {
  display: flex;
  justify-content: center;
  padding: 30px 0;
}
.auth-card {
  width: 100%;
  max-width: 420px;
  background: var(--card);
  border-radius: 20px;
  box-shadow: var(--shadow);
  padding: 32px 32px 24px;
  border: 1px solid var(--line);
}
.postcard-art {
  margin-bottom: 18px;
}
.auth-title {
  font-size: 24px;
  color: var(--accent-deep);
  letter-spacing: 1px;
}
.auth-desc {
  color: var(--ink-soft);
  font-size: 13px;
  margin-top: 8px;
  line-height: 1.7;
}
.auth-form {
  margin-top: 8px;
}
.form-error {
  color: var(--danger);
  font-size: 13px;
  margin-top: 12px;
}
.auth-submit {
  width: 100%;
  margin-top: 22px;
  padding: 12px;
  font-size: 15px;
}
.switch-mode {
  display: block;
  margin: 16px auto 0;
  background: none;
  color: var(--sage-deep);
  font-size: 13px;
  padding: 4px;
}
.demo-hint {
  margin-top: 16px;
  text-align: center;
  font-size: 12px;
  color: var(--ink-soft);
  background: var(--paper-deep);
  border-radius: 10px;
  padding: 9px;
}
</style>
