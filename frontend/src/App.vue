<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, RouterLink, RouterView } from 'vue-router'
import { authService, currentRole } from '@/services/api'

const route = useRoute()
const isLoginPage = computed(() => route.name === 'login')
const userRole = computed(() => currentRole.value)

const handleLogout = () => {
  authService.logout()
}
</script>

<template>
  <div class="min-h-screen flex flex-col bg-[#F2F2F2] text-zinc-900 antialiased selection:bg-[#0A84DC]/20">
    <!-- Header with iOS Glassmorphism (Hidden on Login page) -->
    <header
      v-if="!isLoginPage"
      class="backdrop-blur-xl bg-white/80 border-b border-black/5 sticky top-0 z-30 shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all"
    >
      <div class="max-w-7xl mx-auto px-4 sm:px-6 py-2.5 sm:py-0 min-h-14 sm:h-15 flex flex-wrap items-center justify-between gap-2.5 sm:gap-4">
        <div class="flex items-center gap-3 sm:gap-6 flex-wrap">
          <RouterLink to="/" class="flex items-center space-x-2.5">
            <div class="w-8 h-8 rounded-2xl bg-[#0A84DC] text-white font-mono font-bold flex items-center justify-center text-xs shadow-[0_8px_30px_rgb(0,0,0,0.05)]">
              L
            </div>
            <div>
              <h1 class="text-sm font-semibold leading-none tracking-tight text-zinc-900">LACI</h1>
              <p class="hidden sm:block text-[10px] text-zinc-500 mt-0.5">Pelacakan Surat DOSCOM</p>
            </div>
          </RouterLink>

          <!-- iOS Segmented-like Navigation Links -->
          <nav class="flex items-center p-1 bg-black/[0.04] rounded-2xl space-x-1" aria-label="Navigasi Utama">
            <RouterLink
              v-if="userRole"
              to="/"
              class="px-3 py-1.5 rounded-xl text-xs font-medium text-zinc-600 hover:text-zinc-900 transition-all"
              active-class="!bg-white !text-[#0A84DC] font-semibold shadow-[0_8px_30px_rgb(0,0,0,0.05)]"
            >
              Kanban Board
            </RouterLink>
            <RouterLink
              v-if="userRole === 'sekre'"
              to="/surat-management"
              class="px-3 py-1.5 rounded-xl text-xs font-medium text-zinc-600 hover:text-zinc-900 transition-all"
              active-class="!bg-white !text-[#0A84DC] font-semibold shadow-[0_8px_30px_rgb(0,0,0,0.05)]"
            >
              Tabel Surat
            </RouterLink>
            <RouterLink
              v-if="userRole === 'sekre'"
              to="/proker"
              class="px-3 py-1.5 rounded-xl text-xs font-medium text-zinc-600 hover:text-zinc-900 transition-all"
              active-class="!bg-white !text-[#0A84DC] font-semibold shadow-[0_8px_30px_rgb(0,0,0,0.05)]"
            >
              Kelola Proker
            </RouterLink>
          </nav>
        </div>

        <!-- User Role Badge & Logout -->
        <div class="flex items-center gap-2 sm:gap-3 shrink-0">
          <span
            v-if="userRole"
            class="px-2.5 sm:px-3 py-1 text-[10px] font-mono uppercase font-semibold rounded-xl border border-[#0A84DC]/20 bg-[#0A84DC]/10 text-[#0A84DC]"
          >
            {{ userRole }}
          </span>

          <button
            type="button"
            @click="handleLogout"
            class="text-xs text-zinc-500 hover:text-zinc-900 px-2.5 sm:px-3 py-1.5 rounded-xl hover:bg-black/[0.05] transition-all font-medium"
          >
            Keluar
          </button>
        </div>
      </div>
    </header>

    <!-- Main Viewport -->
    <main
      class="flex-1 w-full mx-auto"
      :class="isLoginPage ? 'p-0' : 'max-w-7xl p-4 sm:p-6 overflow-hidden flex flex-col'"
    >
      <RouterView />
    </main>
  </div>
</template>
