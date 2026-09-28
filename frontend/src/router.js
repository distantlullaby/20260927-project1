import { createRouter, createWebHistory } from 'vue-router'
import Feed from './views/Feed.vue'
import Profile from './views/Profile.vue'
import Login from './views/Login.vue'
import { auth } from './auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: Feed },
    { path: '/login', component: Login },
    { path: '/profile', component: Profile, meta: { requiresAuth: true } }
  ]
})

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !auth.isLogin) {
    return { path: '/login', query: { redirect: to.fullPath } }
  }
})

export default router
