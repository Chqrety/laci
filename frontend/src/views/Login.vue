<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authService } from '@/services/api'

const router = useRouter()

const username = ref('')
const password = ref('')
const isLoading = ref(false)
const errorMessage = ref('')

const handleLogin = async () => {
  if (!username.value || !password.value) {
    errorMessage.value = 'Silakan isi username dan password'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    await authService.login(username.value, password.value)
    router.push({ name: 'board' })
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || err.message || 'Login gagal. Periksa kredensial Anda.'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <main class="min-h-screen flex items-center justify-center p-4 bg-[#F2F2F2] text-zinc-900">
    <!-- iOS Glass Card -->
    <div class="w-full max-w-sm bg-white/80 backdrop-blur-xl border border-black/5 rounded-3xl p-6 sm:p-8 shadow-[0_8px_30px_rgb(0,0,0,0.05)] space-y-6">
      <!-- Header -->
      <header class="text-center space-y-2">
        <div class="w-12 h-12 mx-auto rounded-2xl bg-[#0A84DC] text-white font-mono font-bold flex items-center justify-center text-lg shadow-[0_8px_30px_rgb(0,0,0,0.05)]">
          L
        </div>
        <div>
          <h1 class="text-xl font-semibold tracking-tight text-zinc-900">Masuk ke LACI</h1>
          <p class="text-xs text-zinc-500 mt-0.5">Sistem Pelacakan Surat DOSCOM</p>
        </div>
      </header>

      <!-- Alert -->
      <div
        v-if="errorMessage"
        role="alert"
        class="p-3 text-xs bg-rose-50 border border-rose-200 text-rose-700 rounded-2xl font-medium"
      >
        {{ errorMessage }}
      </div>

      <!-- iOS Inset Grouped Form -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <div class="rounded-2xl border border-black/10 bg-white/90 overflow-hidden divide-y divide-black/5 shadow-2xs">
          <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
            <label for="username" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Username</label>
            <input
              id="username"
              v-model="username"
              type="text"
              autocomplete="username"
              required
              placeholder="sekre / humas"
              class="w-full text-sm font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 pt-0.5"
            />
          </div>

          <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
            <label for="password" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Password</label>
            <input
              id="password"
              v-model="password"
              type="password"
              autocomplete="current-password"
              required
              placeholder="••••••••"
              class="w-full text-sm font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 pt-0.5"
            />
          </div>
        </div>

        <button
          type="submit"
          :disabled="isLoading"
          class="w-full py-3 px-4 text-sm font-semibold rounded-2xl bg-[#0A84DC] text-white hover:bg-[#0872be] focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/40 shadow-[0_8px_30px_rgb(0,0,0,0.05)] disabled:opacity-50 transition-all duration-200 active:scale-95"
        >
          {{ isLoading ? 'Memproses...' : 'Masuk' }}
        </button>
      </form>

      <!-- iOS Style Footer Hint -->
      <footer class="pt-2 text-center text-[11px] text-zinc-500">
        Default akun: <span class="font-mono text-zinc-700 font-semibold">sekre</span> / <span class="font-mono text-zinc-700 font-semibold">password123</span>
      </footer>
    </div>
  </main>
</template>
