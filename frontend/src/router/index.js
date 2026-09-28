import { createRouter, createWebHistory } from 'vue-router'
import { auth } from '../store/auth.js'
import Feed from '../views/Feed.vue'
import Profile from '../views/Profile.vue'
import Login from '../views/Login.vue'

const routes = [
  { path: '/', name: 'feed', component: Feed },
  { path: '/profile', name: 'profile', component: Profile, meta: { requiresAuth: true } },
  { path: '/login', name: 'login', component: Login }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 }
  }
})

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !auth.token) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  return true
})

export default router
