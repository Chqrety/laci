<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { authService, currentUser } from '@/services/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { toast } from '@/components/ui/toast'

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
    toast({
      title: 'Validasi Gagal',
      description: 'Password baru minimal 6 karakter',
      variant: 'destructive',
    })
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    errorMsg.value = 'Konfirmasi password tidak cocok'
    toast({
      title: 'Validasi Gagal',
      description: 'Konfirmasi password tidak cocok',
      variant: 'destructive',
    })
    return
  }

  isSubmitting.value = true
  try {
    await authService.changePassword(newPassword.value)
    successMsg.value = 'Password berhasil diubah!'
    toast({
      title: 'Berhasil',
      description: 'Password akun kamu berhasil diperbarui!',
      variant: 'success',
    })
    setTimeout(() => {
      showModal.value = false
    }, 1200)
  } catch (err: any) {
    errorMsg.value = err.response?.data?.error || 'Gagal mengubah password'
    toast({
      title: 'Gagal Mengubah Password',
      description: errorMsg.value,
      variant: 'destructive',
    })
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <Dialog :open="showModal" @update:open="(val: boolean) => { if (!val) handleDismiss() }">
    <DialogContent class="max-w-md gap-5 overflow-hidden">
      <!-- Background decoration -->
      <div class="absolute -top-12 -right-12 w-32 h-32 bg-[#0A84DC]/10 rounded-full blur-2xl pointer-events-none" />

      <!-- Header -->
      <DialogHeader class="flex flex-row items-start gap-4 space-y-0 text-left">
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
          <DialogTitle class="text-base font-semibold text-zinc-900 tracking-tight">Saran Keamanan Akun</DialogTitle>
          <DialogDescription class="text-xs text-zinc-500 mt-1 leading-relaxed">
            Kamu saat ini masih menggunakan password default dari sistem. Disarankan untuk mengganti password baru yang kamu ingat.
          </DialogDescription>
        </div>
      </DialogHeader>

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
      <form v-if="!successMsg" @submit.prevent="handleSubmit" novalidate class="flex flex-col gap-3.5">
        <div class="space-y-1.5">
          <Label for="new-password">Password Baru</Label>
          <Input
            id="new-password"
            v-model="newPassword"
            type="password"
            placeholder="Minimal 6 karakter"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="confirm-password">Konfirmasi Password Baru</Label>
          <Input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            placeholder="Ketik ulang password baru"
          />
        </div>

        <!-- Action Buttons -->
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            @click="handleDismiss"
          >
            Nanti Saja (Gunakan Default)
          </Button>
          <Button
            type="submit"
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Ganti Password' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
