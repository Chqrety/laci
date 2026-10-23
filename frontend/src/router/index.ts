import { createRouter, createWebHistory } from 'vue-router'
import { authService, currentRole } from '@/services/api'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresAuth: false },
    },
    {
      path: '/',
      name: 'board',
      component: () => import('@/views/KanbanBoard.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/letters',
      alias: '/surat-management',
      name: 'letters',
      component: () => import('@/views/SuratManagement.vue'),
      meta: { requiresAuth: true, role: 'sekre' },
    },
    {
      path: '/programs',
      alias: '/proker',
      name: 'programs',
      component: () => import('@/views/ProkerManagement.vue'),
      meta: { requiresAuth: true, role: 'sekre' },
    },
    {
      path: '/users',
      name: 'UserManagement',
      component: () => import('@/views/UserManagement.vue'),
      meta: { requiresAuth: true, role: 'sekre' },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/views/Profile.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

router.beforeEach((to, _from, next) => {
  const isAuth = authService.isAuthenticated()
  const role = currentRole.value

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
