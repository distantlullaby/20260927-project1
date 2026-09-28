// 极简的登录态管理：token / 用户信息持久化到 localStorage
import { reactive } from 'vue'

const TOKEN_KEY = 'mc_token'
const USER_KEY = 'mc_user'

export const auth = reactive({
  token: localStorage.getItem(TOKEN_KEY) || '',
  user: JSON.parse(localStorage.getItem(USER_KEY) || 'null')
})

export function setAuth(token, user) {
  auth.token = token
  auth.user = user
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function updateUser(user) {
  auth.user = user
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

export function clearAuth() {
  auth.token = ''
  auth.user = null
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}
