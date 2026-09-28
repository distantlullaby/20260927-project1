// 全局轻提示，App.vue 监听并渲染
import { reactive } from 'vue'

export const toastState = reactive({
  message: '',
  type: 'info',
  visible: false
})

let timer = null

export function toast(message, type = 'info', duration = 2600) {
  toastState.message = message
  toastState.type = type
  toastState.visible = true
  clearTimeout(timer)
  timer = setTimeout(() => {
    toastState.visible = false
  }, duration)
}
