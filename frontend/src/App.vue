<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink, RouterView } from 'vue-router'
import { authService, currentRole } from '@/services/api'
import { Button } from '@/components/ui/button'
import PasswordReminderModal from '@/components/PasswordReminderModal.vue'

const route = useRoute()
const isLoginPage = computed(() => route.name === 'login')
const userRole = computed(() => currentRole.value)

const handleLogout = () => {
  authService.logout()
}
</script>

<template>
  <div class="min-h-screen flex flex-col bg-[#F8FAFC] text-slate-900 antialiased selection:bg-sky-200 relative">
    <!-- Floema Subtle Ambient Grid & Corner Blurs (Unifies workspace with Login) -->
    <div v-if="!isLoginPage" class="pointer-events-none fixed inset-0 overflow-hidden z-0">
      <div class="absolute inset-0 bg-[linear-gradient(to_right,#0284C7_1px,transparent_1px),linear-gradient(to_bottom,#0284C7_1px,transparent_1px)] bg-[size:64px_64px] opacity-[0.025]" />
      <div class="absolute -top-32 -right-32 w-[520px] h-[520px] rounded-full bg-gradient-to-br from-sky-400/8 to-transparent blur-3xl" />
      <div class="absolute -bottom-32 -left-32 w-[520px] h-[520px] rounded-full bg-gradient-to-tr from-orange-400/6 to-transparent blur-3xl" />
    </div>

    <!-- Header with Floema Architectural Styling (Hidden on Login page) -->
    <header
      v-if="!isLoginPage"
      class="backdrop-blur-md bg-white/90 border-b border-slate-200/80 sticky top-0 z-30 shadow-xs transition-all"
    >
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-2.5 sm:py-0 min-h-14 sm:h-15 flex flex-wrap items-center justify-between gap-2.5 sm:gap-4">
        <div class="flex items-center gap-3 sm:gap-6 flex-wrap">
          <RouterLink to="/" class="flex items-center space-x-2.5">
            <div class="w-8 h-8 rounded-full bg-sky-600 text-white font-black flex items-center justify-center text-xs shadow-xs">
              L
            </div>
            <div>
              <h1 class="text-sm font-semibold leading-none tracking-tight text-slate-900">LACI</h1>
              <p class="hidden sm:block text-[10px] text-slate-400 mt-0.5">Pelacakan Surat DOSCOM</p>
            </div>
          </RouterLink>

          <!-- Architectural Pill Navigation Links -->
          <nav class="flex items-center p-1 bg-slate-100/80 rounded-full space-x-1" aria-label="Navigasi Utama">
            <RouterLink
              v-if="userRole"
              to="/"
              class="px-3.5 py-1.5 rounded-full text-xs font-medium text-slate-600 hover:text-slate-900 transition-all"
              active-class="!bg-white !text-sky-700 !font-semibold !shadow-xs border border-slate-200/60"
            >
              Kanban Board
            </RouterLink>
            <RouterLink
              v-if="userRole === 'sekre'"
              to="/surat-management"
              class="px-3.5 py-1.5 rounded-full text-xs font-medium text-slate-600 hover:text-slate-900 transition-all"
              active-class="!bg-white !text-sky-700 !font-semibold !shadow-xs border border-slate-200/60"
            >
              Surat
            </RouterLink>
            <RouterLink
              v-if="userRole === 'sekre'"
              to="/proker"
              class="px-3.5 py-1.5 rounded-full text-xs font-medium text-slate-600 hover:text-slate-900 transition-all"
              active-class="!bg-white !text-sky-700 !font-semibold !shadow-xs border border-slate-200/60"
            >
              Proker
            </RouterLink>
            <RouterLink
              v-if="userRole === 'sekre'"
              to="/users"
              class="px-3.5 py-1.5 rounded-full text-xs font-medium text-slate-600 hover:text-slate-900 transition-all"
              active-class="!bg-white !text-sky-700 !font-semibold !shadow-xs border border-slate-200/60"
            >
              Pengurus
            </RouterLink>
            <RouterLink
              v-if="userRole"
              to="/profile"
              class="px-3.5 py-1.5 rounded-full text-xs font-medium text-slate-600 hover:text-slate-900 transition-all"
              active-class="!bg-white !text-sky-700 !font-semibold !shadow-xs border border-slate-200/60"
            >
              Profil
            </RouterLink>
          </nav>
        </div>

        <!-- User Role Badge & Logout -->
        <div class="flex items-center gap-2 sm:gap-3 shrink-0">
          <span
            v-if="userRole"
            class="px-3 py-1 text-[10px] font-mono uppercase font-bold rounded-full border border-sky-200/80 bg-sky-50 text-sky-700"
          >
            {{ userRole }}
          </span>

          <Button
            variant="ghost"
            size="sm"
            @click="handleLogout"
            class="text-xs text-slate-500 hover:text-slate-900 font-medium"
          >
            Keluar
          </Button>
        </div>
      </div>
      <!-- Fine Top Accent Line -->
      <div class="h-[2px] w-full bg-gradient-to-r from-sky-500 via-sky-400 to-orange-400 opacity-70" />
    </header>

    <!-- Password reminder modal (1x pop-up on first login or after sekre reminder) -->
    <PasswordReminderModal v-if="!isLoginPage" />

    <!-- Main Viewport -->
    <main
      class="flex-1 w-full mx-auto relative z-10"
      :class="isLoginPage ? 'p-0' : 'max-w-7xl p-4 sm:p-6 overflow-hidden flex flex-col'"
    >
      <RouterView v-slot="{ Component, route }">
        <Transition name="page-fade" mode="out-in">
          <KeepAlive include="KanbanBoard">
            <component :is="Component" :key="route.path" />
          </KeepAlive>
        </Transition>
      </RouterView>
    </main>
  </div>
</template>
