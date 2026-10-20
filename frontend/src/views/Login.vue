<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authService } from '@/services/api'

const router = useRouter()

const username = ref('')
const password = ref('')
const showPassword = ref(false)
const isLoading = ref(false)
const errorMessage = ref('')

const fillQuickAccount = (roleUser: 'sekre' | 'humas') => {
  username.value = roleUser
  password.value = 'password123'
  errorMessage.value = ''
}

const handleLogin = async () => {
  if (!username.value.trim() || !password.value) {
    errorMessage.value = 'Silakan isi username dan password terlebih dahulu.'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    await authService.login(username.value.trim(), password.value)
    router.push({ name: 'board' })
  } catch (err: any) {
    errorMessage.value =
      err.response?.data?.error || err.message || 'Login gagal. Periksa kembali kredensial Anda.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div
    class="relative min-h-dvh flex flex-col items-center justify-center p-4 sm:p-6 bg-[#F2F2F4] text-zinc-900 overflow-hidden select-none"
  >
    <!-- Subtle Ambient Background Lighting & Dot Grid -->
    <div
      class="pointer-events-none absolute inset-0 bg-[radial-gradient(50%_45%_at_50%_0%,rgba(10,132,220,0.10),transparent_70%),radial-gradient(40%_35%_at_80%_90%,rgba(10,132,220,0.06),transparent_70%)]"
    />
    <div
      class="pointer-events-none absolute inset-0 opacity-[0.35] bg-[radial-gradient(#a1a1aa_1px,transparent_1px)] [background-size:20px_20px] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_50%,#000_60%,transparent_100%)]"
    />

    <!-- Main Glassmorphic Login Container -->
    <section
      aria-label="Panel Masuk LACI"
      class="relative z-10 w-full max-w-[390px] bg-white/85 backdrop-blur-2xl border border-black/[0.06] rounded-[28px] p-6 sm:p-8 shadow-[0_20px_50px_-12px_rgba(0,0,0,0.07),0_0_0_1px_rgba(255,255,255,0.8)_inset] space-y-6"
    >
      <!-- Brand Header -->
      <header class="text-center space-y-3">
        <!-- App Squircle Icon -->
        <div
          class="w-12 h-12 mx-auto rounded-[18px] bg-gradient-to-b from-[#1690E8] to-[#0A84DC] text-white flex items-center justify-center shadow-[0_10px_25px_-5px_rgba(10,132,220,0.45)] ring-1 ring-inset ring-white/30"
        >
          <svg
            class="w-6 h-6"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"
            />
          </svg>
        </div>

        <div class="space-y-1">
          <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-[#0A84DC]/10 border border-[#0A84DC]/15 text-[#0A84DC] text-[10px] font-mono font-semibold uppercase tracking-wider">
            <span class="w-1.5 h-1.5 rounded-full bg-[#0A84DC]" />
            DOSCOM Archive
          </div>
          <h1 class="text-xl font-semibold tracking-tight text-zinc-900">Masuk ke LACI</h1>
          <p class="text-xs text-zinc-500 leading-relaxed">
            Sistem Pelacakan &amp; Pengarsipan Surat Resmi
          </p>
        </div>
      </header>

      <!-- Animated Error Alert -->
      <Transition
        enter-active-class="transition duration-200 ease-out"
        enter-from-class="opacity-0 -translate-y-1 scale-95"
        enter-to-class="opacity-100 translate-y-0 scale-100"
        leave-active-class="transition duration-150 ease-in"
        leave-from-class="opacity-100 translate-y-0 scale-100"
        leave-to-class="opacity-0 -translate-y-1 scale-95"
      >
        <div
          v-if="errorMessage"
          role="alert"
          class="flex items-start gap-2.5 p-3 text-xs bg-rose-50/90 border border-rose-200/80 text-rose-700 rounded-2xl font-medium"
        >
          <svg
            class="w-4 h-4 text-rose-500 shrink-0 mt-0.5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z"
            />
          </svg>
          <span class="flex-1 leading-snug">{{ errorMessage }}</span>
          <button
            type="button"
            @click="errorMessage = ''"
            aria-label="Tutup pesan error"
            class="text-rose-400 hover:text-rose-600 transition-colors cursor-pointer"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      </Transition>

      <!-- iOS Inset Grouped Form -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <div
          class="rounded-2xl border border-black/[0.08] bg-white/95 overflow-hidden divide-y divide-black/[0.06] shadow-2xs transition-all focus-within:border-[#0A84DC]/40 focus-within:ring-3 focus-within:ring-[#0A84DC]/10"
        >
          <!-- Username Field -->
          <div class="group px-3.5 py-2.5 flex items-center gap-3 focus-within:bg-[#0A84DC]/[0.02] transition-colors">
            <div class="w-7 h-7 rounded-xl bg-zinc-100 group-focus-within:bg-[#0A84DC]/10 text-zinc-400 group-focus-within:text-[#0A84DC] flex items-center justify-center shrink-0 transition-colors">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <label
                for="username"
                class="block text-[10px] font-semibold text-zinc-400 group-focus-within:text-[#0A84DC] uppercase tracking-wider transition-colors"
              >
                Username
              </label>
              <input
                id="username"
                v-model="username"
                type="text"
                autocomplete="username"
                required
                placeholder="sekre atau humas"
                class="w-full text-sm font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-300 pt-0.5"
              />
            </div>
          </div>

          <!-- Password Field -->
          <div class="group px-3.5 py-2.5 flex items-center gap-3 focus-within:bg-[#0A84DC]/[0.02] transition-colors">
            <div class="w-7 h-7 rounded-xl bg-zinc-100 group-focus-within:bg-[#0A84DC]/10 text-zinc-400 group-focus-within:text-[#0A84DC] flex items-center justify-center shrink-0 transition-colors">
              <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <div class="flex-1 min-w-0">
              <label
                for="password"
                class="block text-[10px] font-semibold text-zinc-400 group-focus-within:text-[#0A84DC] uppercase tracking-wider transition-colors"
              >
                Password
              </label>
              <input
                id="password"
                v-model="password"
                :type="showPassword ? 'text' : 'password'"
                autocomplete="current-password"
                required
                placeholder="••••••••"
                class="w-full text-sm font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-300 pt-0.5"
              />
            </div>
            <button
              type="button"
              @click="showPassword = !showPassword"
              :aria-label="showPassword ? 'Sembunyikan password' : 'Tampilkan password'"
              class="p-1.5 rounded-xl text-zinc-400 hover:text-zinc-600 hover:bg-black/[0.04] transition-colors cursor-pointer focus:outline-none"
            >
              <svg
                v-if="!showPassword"
                class="w-4 h-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <svg
                v-else
                class="w-4 h-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                stroke-width="2"
              >
                <path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Submit Button -->
        <button
          type="submit"
          :disabled="isLoading"
          class="group w-full py-3 px-4 text-sm font-semibold rounded-2xl bg-[#0A84DC] text-white hover:bg-[#0872be] focus:outline-none focus:ring-3 focus:ring-[#0A84DC]/30 shadow-[0_8px_20px_-4px_rgba(10,132,220,0.4)] disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 active:scale-[0.98] flex items-center justify-center gap-2 cursor-pointer"
        >
          <svg
            v-if="isLoading"
            class="w-4 h-4 animate-spin text-white"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="3" />
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          <span>{{ isLoading ? 'Memverifikasi...' : 'Masuk ke Dashboard' }}</span>
          <svg
            v-if="!isLoading"
            class="w-4 h-4 transition-transform duration-200 group-hover:translate-x-0.5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            stroke-width="2"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 7l5 5m0 0l-5 5m5-5H6" />
          </svg>
        </button>
      </form>

      <!-- Quick Role Switcher / Demo Account Helper -->
      <footer class="pt-3 border-t border-black/[0.06] space-y-2">
        <div class="flex items-center justify-between text-[11px] text-zinc-400">
          <span>Akses cepat akun demo:</span>
          <span class="font-mono text-[10px]">pw: password123</span>
        </div>

        <div class="grid grid-cols-2 gap-2">
          <button
            type="button"
            @click="fillQuickAccount('sekre')"
            class="px-3 py-1.5 rounded-xl border text-xs font-medium transition-all flex items-center justify-center gap-1.5 cursor-pointer active:scale-95"
            :class="
              username === 'sekre'
                ? 'bg-[#0A84DC]/10 border-[#0A84DC]/30 text-[#0A84DC] font-semibold'
                : 'bg-black/[0.02] border-black/[0.06] text-zinc-600 hover:bg-black/[0.05] hover:text-zinc-900'
            "
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="username === 'sekre' ? 'bg-[#0A84DC]' : 'bg-zinc-400'" />
            Sekretaris
          </button>

          <button
            type="button"
            @click="fillQuickAccount('humas')"
            class="px-3 py-1.5 rounded-xl border text-xs font-medium transition-all flex items-center justify-center gap-1.5 cursor-pointer active:scale-95"
            :class="
              username === 'humas'
                ? 'bg-[#0A84DC]/10 border-[#0A84DC]/30 text-[#0A84DC] font-semibold'
                : 'bg-black/[0.02] border-black/[0.06] text-zinc-600 hover:bg-black/[0.05] hover:text-zinc-900'
            "
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="username === 'humas' ? 'bg-[#0A84DC]' : 'bg-zinc-400'" />
            Humas
          </button>
        </div>
      </footer>
    </section>
  </div>
</template>
