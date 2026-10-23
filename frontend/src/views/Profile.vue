<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as z from 'zod'
import { toTypedSchema } from '@vee-validate/zod'
import { authService, type User } from '@/services/api'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
  FormDescription,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'

const profile = ref<User | null>(null)
const isLoading = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const isSubmitting = ref(false)
const isSubmittingUsername = ref(false)
const msg = ref<{ type: 'success' | 'error'; text: string } | null>(null)

const usernameFormSchema = toTypedSchema(
  z.object({
    username: z
      .string({ required_error: 'Username wajib diisi' })
      .min(1, 'Username wajib diisi')
      .min(3, 'Username minimal 3 karakter')
      .regex(/^\S+$/, 'Username tidak boleh mengandung spasi'),
  }),
)

const fetchProfile = async () => {
  isLoading.value = true
  try {
    profile.value = await authService.getProfile()
  } catch (err: any) {
    msg.value = { type: 'error', text: err.response?.data?.error || 'Gagal memuat profil' }
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchProfile()
})

const onSubmitUsername = async (values: Record<string, any>) => {
  msg.value = null
  isSubmittingUsername.value = true
  try {
    const res = await authService.changeUsername(String(values.username).trim())
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

    <!-- Form Ganti Username (Shadcn-vue + Vee-Validate + Zod) -->
    <div class="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs flex flex-col gap-4">
      <div>
        <h3 class="text-sm font-semibold text-zinc-900">Ubah Username</h3>
        <p class="text-xs text-zinc-400 mt-0.5">Ubah username kamu agar lebih mudah diingat saat login</p>
      </div>

      <Form
        :key="profile?.username || 'user-form'"
        :validation-schema="usernameFormSchema"
        :initial-values="{ username: profile?.username || '' }"
        novalidate
        @submit="onSubmitUsername"
        class="flex flex-col gap-3.5"
        v-slot="{ values }"
      >
        <FormField name="username" v-slot="{ componentField }">
          <FormItem>
            <FormLabel>Username Baru</FormLabel>
            <FormControl>
              <div class="relative">
                <span class="absolute left-3.5 top-1/2 -translate-y-1/2 text-xs text-zinc-400 font-medium z-10 pointer-events-none">@</span>
                <Input
                  type="text"
                  placeholder="username_kamu"
                  class="pl-8"
                  v-bind="componentField"
                />
              </div>
            </FormControl>
            <FormMessage />
            <FormDescription>
              Default akun menggunakan format persis NIM (misal: {{ profile?.nim || 'A11.2022.14000' }}). Anda dapat menggantinya menjadi username personal tanpa spasi.
            </FormDescription>
          </FormItem>
        </FormField>

        <div class="flex justify-end pt-1">
          <Button
            type="submit"
            :disabled="isSubmittingUsername || !values.username || values.username === profile?.username"
          >
            {{ isSubmittingUsername ? 'Menyimpan...' : 'Simpan Username Baru' }}
          </Button>
        </div>
      </Form>
    </div>

    <!-- Form Ganti Password -->
    <div class="bg-white rounded-3xl p-6 border border-slate-200/90 shadow-xs flex flex-col gap-4">
      <div>
        <h3 class="text-sm font-semibold text-zinc-900">Ubah Password</h3>
        <p class="text-xs text-zinc-400 mt-0.5">Ganti password akun kamu kapan saja demi keamanan</p>
      </div>

      <form @submit.prevent="handleChangePassword" class="flex flex-col gap-3.5">
        <div v-if="!profile?.is_password_default" class="space-y-1.5">
          <Label for="old-password">Password Saat Ini</Label>
          <Input
            id="old-password"
            v-model="oldPassword"
            type="password"
            placeholder="Masukkan password saat ini"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="new-password">Password Baru</Label>
          <Input
            id="new-password"
            v-model="newPassword"
            type="password"
            required
            placeholder="Minimal 6 karakter"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="confirm-password">Konfirmasi Password Baru</Label>
          <Input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            required
            placeholder="Ketik ulang password baru"
          />
        </div>

        <div class="flex justify-end pt-2">
          <Button
            type="submit"
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Password Baru' }}
          </Button>
        </div>
      </form>
    </div>
  </div>
</template>
