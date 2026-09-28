import axios from 'axios'

const api = axios.create({ baseURL: '/api', timeout: 15000 })

api.interceptors.request.use((cfg) => {
  const token = localStorage.getItem('ml_token')
  if (token) cfg.headers.Authorization = `Bearer ${token}`
  return cfg
})

api.interceptors.response.use(
  (res) => res.data,
  (err) => {
    const msg = err.response?.data?.error || '网络异常，请稍后再试'
    if (err.response?.status === 401 && !err.config?._silent) {
      localStorage.removeItem('ml_token')
      localStorage.removeItem('ml_user')
    }
    return Promise.reject(new Error(msg))
  }
)

export default api
