import { auth, clearAuth } from '../store/auth.js'

// 统一的 RESTful API 封装，自动携带 token、统一报错
async function request(method, path, body, isForm = false) {
  const headers = {}
  if (auth.token) headers.Authorization = 'Bearer ' + auth.token
  if (body && !isForm) headers['Content-Type'] = 'application/json'

  const res = await fetch('/api' + path, {
    method,
    headers,
    body: body ? (isForm ? body : JSON.stringify(body)) : undefined
  })

  let data = null
  try {
    data = await res.json()
  } catch (e) {
    data = {}
  }

  if (!res.ok) {
    if (res.status === 401) clearAuth()
    throw new Error(data.error || '请求失败（' + res.status + '）')
  }
  return data
}

export const api = {
  get: (p) => request('GET', p),
  post: (p, b) => request('POST', p, b),
  upload: (file) => {
    const fd = new FormData()
    fd.append('file', file)
    return request('POST', '/uploads', fd, true)
  }
}
