import { reactive } from 'vue'
import api from './api'

export const auth = reactive({
  token: localStorage.getItem('ml_token') || '',
  user: JSON.parse(localStorage.getItem('ml_user') || 'null'),

  get isLogin() {
    return !!this.token
  },

  setSession(token, user) {
    this.token = token
    this.user = user
    localStorage.setItem('ml_token', token)
    localStorage.setItem('ml_user', JSON.stringify(user))
  },

  async refresh() {
    if (!this.token) return
    const { user } = await api.get('/me')
    this.user = user
    localStorage.setItem('ml_user', JSON.stringify(user))
  },

  logout() {
    this.token = ''
    this.user = null
    localStorage.removeItem('ml_token')
    localStorage.removeItem('ml_user')
  }
})
