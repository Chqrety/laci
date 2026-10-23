<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { authService, type User } from '@/services/api'

const profile = ref<User | null>(null)
const isLoading = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const isSubmitting = ref(false)
const newUsername = ref('')
const isSubmittingUsername = ref(false)
const msg = ref<{ type: 'success' | 'error'; text: string } | null>(null)

const fetchProfile = async () => {
  isLoading.value = true
  try {
    profile.value = await authService.getProfile()
    if (profile.value?.username) {
      newUsername.value = profile.value.username
    }
  } catch (err: any) {
    msg.value = { type: 'error', text: err.response?.data?.error || 'Gagal memuat profil' }
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchProfile()
})

const handleChangeUsername = async () => {
  msg.value = null
  isSubmittingUsername.value = true
  try {
    const res = await authService.changeUsername(newUsername.value.trim())
    msg.value = { type: 'success', text: res.message || 'Username berhasil diperbarui!' }
    if (profile.value) {
      profile.value.username = res.username
    }
  } catch (err: any) {
    msg.value = { type: 'error', text: err.response?.data?.error || 'Gagal mengubah username' }
  } finally {
    isSubmittingUsername.value = false
  }
}

const handleChangePassword = async () => {
  msg.value = null
  if (!newPassword.value || newPassword.value.length < 6) {
    msg.value = { type: 'error', text: 'Password baru minimal 6 karakter' }
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    msg.value = { type: 'error', text: 'Konfirmasi password tidak cocok' }
    return
  }

  isSubmitting.value = true
  try {
    const res = await authService.changePassword(newPassword.value, oldPassword.value)
    msg.value = { type: 'success', text: res.message || 'Password berhasil diperbarui!' }
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    await fetchProfile()
  } catch (err: any) {
    msg.value = { type: 'error', text: err.response?.data?.error || 'Gagal mengubah password' }
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="max-w-2xl mx-auto w-full flex flex-col gap-6 py-4">
    <div>
      <h2 class="text-xl font-medium tracking-tight text-slate-900 leading-snug">Profil Saya</h2>
      <p class="text-xs text-zinc-500 mt-0.5">Informasi akun dan pengaturan keamanan password</p>
    </div>

    <!-- Alert message -->
    <div
      v-if="msg"
      class="p-3.5 rounded-2xl text-xs font-semibold flex items-center justify-between border"
      :class="
        msg.type === 'success'
          ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
          : 'bg-rose-50 text-rose-700 border-rose-200'
      "
    >
      <span>{{ msg.text }}</span>
      <button @click="msg = null" class="opacity-60 hover:opacity-100">✕</button>
    </div>

    <!-- Info User Card -->
    <div class="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs flex flex-col gap-4">
      <div class="flex items-center gap-4">
        <div class="w-14 h-14 rounded-2xl bg-sky-50 text-sky-700 flex items-center justify-center font-bold text-lg border border-sky-100">
          {{ profile?.nama ? profile.nama.charAt(0).toUpperCase() : 'U' }}
        </div>
        <div>
          <h3 class="text-base font-semibold text-zinc-900">{{ profile?.nama || profile?.username }}</h3>
          <p class="text-xs text-zinc-400">@{{ profile?.username }}</p>
        </div>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-3 border-t border-black/5 text-xs">
        <div>
          <span class="text-zinc-400 block text-[11px]">NIM</span>
          <span class="font-mono font-medium text-zinc-800">{{ profile?.nim || '-' }}</span>
        </div>
        <div>
          <span class="text-zinc-400 block text-[11px]">Divisi</span>
          <span class="font-mono uppercase font-semibold text-[#0A84DC]">{{ profile?.divisi }}</span>
        </div>
        <div>
          <span class="text-zinc-400 block text-[11px]">Email</span>
          <span class="text-zinc-800">{{ profile?.email || '-' }}</span>
        </div>
        <div>
          <span class="text-zinc-400 block text-[11px]">Status Password</span>
          <span
            class="inline-block px-2 py-0.5 mt-0.5 rounded-lg text-[10px] font-semibold"
            :class="
              profile?.is_password_default
                ? 'bg-amber-50 text-amber-700 border border-amber-200'
                : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
            "
          >
            {{ profile?.is_password_default ? 'Masih Default' : 'Password Kustom' }}
          </span>
        </div>
      </div>
    </div>

    <!-- Form Ganti Username -->
    <div class="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs flex flex-col gap-4">
      <div>
        <h3 class="text-sm font-semibold text-zinc-900">Ubah Username</h3>
        <p class="text-xs text-zinc-400 mt-0.5">Ubah username kamu agar lebih mudah diingat saat login</p>
      </div>

      <form @submit.prevent="handleChangeUsername" class="flex flex-col gap-3.5">
        <div>
          <label class="block text-xs font-medium text-zinc-700 mb-1">Username Baru</label>
          <div class="relative">
            <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-xs text-zinc-400 font-medium">@</span>
            <input
              v-model="newUsername"
              type="text"
              required
              minlength="3"
              pattern="^\S+$"
              title="Minimal 3 karakter tanpa spasi"
              placeholder="username_kamu"
              class="w-full pl-8 pr-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
          </div>
          <p class="text-[11px] text-zinc-400 mt-1">
            Default akun menggunakan format persis NIM (misal: {{ profile?.nim || 'A11.2022.14000' }}). Anda dapat menggantinya menjadi username personal tanpa spasi.
          </p>
        </div>

        <div class="flex justify-end pt-1">
          <button
            type="submit"
            :disabled="isSubmittingUsername || !newUsername.trim() || newUsername.trim() === profile?.username"
            class="px-6 py-2.5 rounded-full text-xs font-bold text-white bg-sky-600 hover:bg-sky-700 active:scale-[0.98] transition-all shadow-xs disabled:opacity-50 cursor-pointer"
          >
            {{ isSubmittingUsername ? 'Menyimpan...' : 'Simpan Username Baru' }}
          </button>
        </div>
      </form>
    </div>

    <!-- Form Ganti Password -->
    <div class="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs flex flex-col gap-4">
      <div>
        <h3 class="text-sm font-semibold text-zinc-900">Ubah Password</h3>
        <p class="text-xs text-zinc-400 mt-0.5">Ganti password akun kamu kapan saja demi keamanan</p>
      </div>

      <form @submit.prevent="handleChangePassword" class="flex flex-col gap-3.5">
        <div v-if="!profile?.is_password_default">
          <label class="block text-xs font-medium text-zinc-700 mb-1">Password Saat Ini</label>
          <input
            v-model="oldPassword"
            type="password"
            placeholder="Masukkan password saat ini"
            class="w-full px-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-zinc-700 mb-1">Password Baru</label>
          <input
            v-model="newPassword"
            type="password"
            required
            placeholder="Minimal 6 karakter"
            class="w-full px-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-zinc-700 mb-1">Konfirmasi Password Baru</label>
          <input
            v-model="confirmPassword"
            type="password"
            required
            placeholder="Ketik ulang password baru"
            class="w-full px-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
          />
        </div>

        <div class="flex justify-end pt-2">
          <button
            type="submit"
            :disabled="isSubmitting"
            class="px-6 py-2.5 rounded-full text-xs font-bold text-white bg-sky-600 hover:bg-sky-700 active:scale-[0.98] transition-all shadow-xs disabled:opacity-50 cursor-pointer"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Password Baru' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
