import { createRouter, createWebHistory } from 'vue-router'
import { authService } from '@/services/api'
import Login from '@/views/Login.vue'
import KanbanBoard from '@/views/KanbanBoard.vue'
import ProkerManagement from '@/views/ProkerManagement.vue'
import SuratManagement from '@/views/SuratManagement.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: Login,
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      name: 'board',
      component: KanbanBoard,
      meta: { requiresAuth: true },
    },
    {
      path: '/surat-management',
      name: 'SuratManagement',
      component: SuratManagement,
      meta: { requiresAuth: true, role: 'sekre' },
    },
    {
      path: '/proker',
      name: 'proker',
      component: ProkerManagement,
      meta: { requiresAuth: true, role: 'sekre' },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const isAuth = authService.isAuthenticated()
  const role = authService.getRole()

  if (to.meta.requiresAuth && !isAuth) {
    next({ name: 'login' })
  } else if (to.name === 'login' && isAuth) {
    next({ name: 'board' })
  } else if (to.meta.role && to.meta.role !== role) {
    next({ name: 'board' })
  } else {
    next()
  }
})

export default router
