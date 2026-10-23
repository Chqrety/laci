<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { authService, currentUser } from '@/services/api'

const showModal = ref(false)
const newPassword = ref('')
const confirmPassword = ref('')
const errorMsg = ref('')
const isSubmitting = ref(false)
const successMsg = ref('')

onMounted(async () => {
  if (!authService.isAuthenticated()) return

  try {
    const user = await authService.getProfile()
    // Jika belum pernah melihat prompt saran ganti password, tampilkan modal 1x
    if (!user.has_seen_password_prompt) {
      showModal.value = true
    }
  } catch {
    // Abaikan jika token expired atau error network
  }
})

const handleDismiss = async () => {
  try {
    await authService.dismissPasswordPrompt()
  } catch {
    // silent
  } finally {
    showModal.value = false
  }
}

const handleSubmit = async () => {
  errorMsg.value = ''
  if (!newPassword.value || newPassword.value.length < 6) {
    errorMsg.value = 'Password baru minimal 6 karakter'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    errorMsg.value = 'Konfirmasi password tidak cocok'
    return
  }

  isSubmitting.value = true
  try {
    await authService.changePassword(newPassword.value)
    successMsg.value = 'Password berhasil diubah!'
    setTimeout(() => {
      showModal.value = false
    }, 1200)
  } catch (err: any) {
    errorMsg.value = err.response?.data?.error || 'Gagal mengubah password'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <Transition name="modal-fade">
    <div
      v-if="showModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
    >
      <div
        class="modal-panel bg-white rounded-3xl max-w-md w-full p-6 sm:p-7 shadow-2xl border border-black/5 flex flex-col gap-5 relative overflow-hidden"
      >
      <!-- Background decoration -->
      <div class="absolute -top-12 -right-12 w-32 h-32 bg-[#0A84DC]/10 rounded-full blur-2xl pointer-events-none" />

      <!-- Header -->
      <div class="flex items-start gap-4">
        <div class="w-12 h-12 rounded-2xl bg-[#0A84DC]/10 text-[#0A84DC] flex items-center justify-center shrink-0">
          <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.8"
              d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
            />
          </svg>
        </div>
        <div>
          <h3 class="text-base font-semibold text-zinc-900 tracking-tight">Saran Keamanan Akun</h3>
          <p class="text-xs text-zinc-500 mt-1 leading-relaxed">
            Kamu saat ini masih menggunakan password default dari sistem. Disarankan untuk mengganti password baru yang kamu ingat.
          </p>
        </div>
      </div>

      <div
        v-if="successMsg"
        class="p-3 rounded-2xl bg-emerald-50 text-emerald-700 text-xs font-medium border border-emerald-200"
      >
        {{ successMsg }}
      </div>

      <div
        v-if="errorMsg"
        class="p-3 rounded-2xl bg-rose-50 text-rose-700 text-xs font-medium border border-rose-200"
      >
        {{ errorMsg }}
      </div>

      <!-- Form -->
      <form v-if="!successMsg" @submit.prevent="handleSubmit" class="flex flex-col gap-3.5">
        <div>
          <label class="block text-xs font-medium text-zinc-700 mb-1">Password Baru</label>
          <input
            v-model="newPassword"
            type="password"
            placeholder="Minimal 6 karakter"
            class="w-full px-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-sm focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC] transition-all"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-zinc-700 mb-1">Konfirmasi Password Baru</label>
          <input
            v-model="confirmPassword"
            type="password"
            placeholder="Ketik ulang password baru"
            class="w-full px-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-sm focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC] transition-all"
          />
        </div>

        <!-- Action Buttons -->
        <div class="flex items-center justify-end gap-2.5 mt-2">
          <button
            type="button"
            @click="handleDismiss"
            class="px-4 py-2.5 rounded-2xl text-xs font-medium text-zinc-600 hover:text-zinc-900 hover:bg-zinc-100 transition-all"
          >
            Nanti Saja (Gunakan Default)
          </button>
          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-5 py-2.5 rounded-2xl text-xs font-semibold text-white bg-[#0A84DC] hover:bg-[#086db8] active:scale-[0.98] transition-all shadow-[0_8px_20px_rgba(10,132,220,0.25)] disabled:opacity-50"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Ganti Password' }}
          </button>
        </div>
      </form>
    </div>
  </div>
  </Transition>
</template>
