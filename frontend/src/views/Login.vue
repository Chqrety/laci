<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { authService, currentRole, type LoginResponse } from '@/services/api'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'

const router = useRouter()

// --- Form & State Management ---
const currentStep = ref<'login' | 'force-password'>('login')

// State: Login
const username = ref('')
const password = ref('')
const showPassword = ref(false)
const isLoading = ref(false)
const errorMessage = ref('')

// State: Force Password Change
const newPassword = ref('')
const confirmPassword = ref('')
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)
const isUpdatingPassword = ref(false)
const passwordSuccessMessage = ref('')

const loggedUsername = ref('')


// Handle Login Submission
const handleLogin = async () => {
  if (!username.value.trim() || !password.value) {
    errorMessage.value = 'Harap lengkapi username atau NIM dan kata sandi.'
    return
  }

  isLoading.value = true
  errorMessage.value = ''

  try {
    const res: LoginResponse = await authService.login(username.value.trim(), password.value)
    loggedUsername.value = res.username || username.value

    if (res.is_password_default) {
      currentStep.value = 'force-password'
      errorMessage.value = ''
    } else {
      router.push({ name: 'board' })
    }
  } catch (err: any) {
    errorMessage.value =
      err.response?.data?.error || err.message || 'Otentikasi gagal. Periksa kembali data login Anda.'
  } finally {
    isLoading.value = false
  }
}

// Password Validation
const isMinLength = computed(() => newPassword.value.length >= 6)
const isPasswordMatched = computed(
  () => newPassword.value.length > 0 && newPassword.value === confirmPassword.value,
)
const canSubmitPasswordChange = computed(
  () => isMinLength.value && isPasswordMatched.value && !isUpdatingPassword.value,
)

// Handle Force Change Password
const handleForceChangePassword = async () => {
  if (!canSubmitPasswordChange.value) return

  isUpdatingPassword.value = true
  errorMessage.value = ''

  try {
    await authService.changePassword(newPassword.value, password.value)
    passwordSuccessMessage.value = 'Kredensial berhasil diperbarui. Mengalihkan ke sistem...'

    setTimeout(() => {
      router.push({ name: 'board' })
    }, 800)
  } catch (err: any) {
    errorMessage.value =
      err.response?.data?.error || err.message || 'Gagal memperbarui kata sandi. Silakan coba kembali.'
  } finally {
    isUpdatingPassword.value = false
  }
}

const handleSkipPasswordChange = async () => {
  try {
    await authService.dismissPasswordPrompt()
  } catch {
    // Ignore error and proceed
  }
  router.push({ name: 'board' })
}
</script>

<template>
  <div
    class="min-h-[100dvh] w-full bg-[#F8FAFC] text-slate-900 antialiased relative overflow-hidden flex flex-col justify-between p-4 sm:p-6 md:p-10 lg:p-12 select-none"
  >
    <!-- ================= FLOEMA BACKGROUND: SUBTLE GEOMETRIC & ORGANIC ACCENTS ================= -->
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <!-- Architectural Grid Lines (Ultra Subtle) -->
      <div
        class="absolute inset-0 bg-[linear-gradient(to_right,#0284C7_1px,transparent_1px),linear-gradient(to_bottom,#0284C7_1px,transparent_1px)] bg-[size:64px_64px] opacity-[0.03]"
      />

      <!-- Subtle Geometric Circle 1 (DOSCOM Blue Soft Glow) -->
      <div
        class="absolute -top-32 -right-32 w-[520px] h-[520px] rounded-full bg-gradient-to-br from-sky-400/10 to-transparent blur-2xl"
      />

      <!-- Subtle Geometric Circle 2 (DOSCOM Orange Accent Glow) -->
      <div
        class="absolute -bottom-40 -left-40 w-[560px] h-[560px] rounded-full bg-gradient-to-tr from-orange-400/8 to-transparent blur-3xl"
      />

      <!-- Architectural Delicate Rings (Floema Motif) -->
      <div
        class="absolute top-1/4 right-8 lg:right-24 w-72 h-72 rounded-full border border-sky-300/25 pointer-events-none"
      />
      <div
        class="absolute top-1/4 right-8 lg:right-24 w-48 h-48 rounded-full border border-dashed border-sky-400/20 pointer-events-none"
      />

      <!-- Subtle Floating Organic Arc (Bottom Right) -->
      <svg
        class="absolute -bottom-20 right-1/4 w-96 h-96 text-orange-500/5"
        viewBox="0 0 200 200"
        fill="currentColor"
      >
        <path
          d="M45.7,-77.9C58.9,-71.1,69.1,-58.5,76.5,-44.6C83.9,-30.7,88.5,-15.3,87.3,-0.7C86.1,13.9,79.1,27.8,70.8,40.1C62.4,52.4,52.8,63,40.7,70.6C28.7,78.2,14.3,82.8,-0.9,84.4C-16.1,86,-32.2,84.5,-45.8,77.5C-59.4,70.4,-70.6,57.7,-77.9,43.2C-85.2,28.7,-88.7,12.3,-87.3,-3.8C-85.9,-19.9,-79.6,-35.8,-69.8,-48.5C-60,-61.2,-46.8,-70.7,-32.8,-77C-18.7,-83.4,-3.8,-86.6,10.5,-85.4C24.8,-84.2,32.5,-84.7,45.7,-77.9Z"
          transform="translate(100 100)"
        />
      </svg>

      <!-- Corner Crosshair Ticks -->
      <span class="absolute top-6 left-6 text-slate-300 font-mono text-xs select-none">+</span>
      <span class="absolute top-6 right-6 text-slate-300 font-mono text-xs select-none">+</span>
      <span class="absolute bottom-6 left-6 text-slate-300 font-mono text-xs select-none">+</span>
      <span class="absolute bottom-6 right-6 text-slate-300 font-mono text-xs select-none">+</span>
    </div>

    <!-- ================= TOP ARCHITECTURAL HEADER ================= -->
    <header class="relative z-10 w-full max-w-5xl mx-auto flex items-center justify-between py-2">
      <!-- DOSCOM Pill Logo Badge -->
      <div class="flex items-center gap-3">
        <img
          src="/logo-doscom.png"
          alt="DOSCOM"
          class="h-9 w-9 rounded-full object-contain shrink-0 shadow-xs"
          @error="(e) => ((e.target as HTMLElement).style.display = 'none')"
        />
        <div>
          <span class="text-xs font-bold tracking-widest uppercase text-slate-900 block leading-tight">
            DOSCOM
          </span>
          <span class="text-[10px] font-medium tracking-wider text-slate-400 block uppercase">
            Sistem Laci Digital
          </span>
        </div>
      </div>

      <!-- Architectural Status Pill -->
      <div class="hidden sm:inline-flex items-center gap-2 px-3.5 py-1.5 rounded-full border border-slate-200/80 bg-white/80 backdrop-blur-sm text-xs font-semibold text-slate-600">
        <span class="w-2 h-2 rounded-full bg-emerald-500" />
        <span class="tracking-wide">Portal Resmi Ormawa</span>
      </div>
    </header>

    <!-- ================= CENTER FRAMED LOGIN CONTAINER (FLOEMA CLEAN FRAMING) ================= -->
    <main class="relative z-10 w-full max-w-xl mx-auto my-auto py-8">
      <!-- Outer Architectural Card -->
      <div
        class="border border-slate-200/90 bg-white/95 backdrop-blur-md rounded-3xl p-6 sm:p-10 md:p-12 shadow-[0_12px_40px_-16px_rgba(2,132,199,0.08)] relative"
      >
        <!-- Fine Accent Bar at Card Top (DOSCOM Blue & Orange) -->
        <div class="absolute top-0 left-10 right-10 h-[2px] bg-gradient-to-r from-sky-500 via-sky-400 to-orange-400 rounded-full" />

        <!-- Feedback Alert -->
        <Transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="opacity-0 -translate-y-2 scale-95"
          enter-to-class="opacity-100 translate-y-0 scale-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 translate-y-0 scale-100"
          leave-to-class="opacity-0 -translate-y-2 scale-95"
        >
          <div
            v-if="errorMessage"
            role="alert"
            class="mb-6 flex items-start gap-3 p-3.5 bg-rose-50 border border-rose-200/80 text-rose-800 rounded-2xl text-xs sm:text-sm font-medium"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-rose-500 shrink-0 mt-1.5" />
            <span class="flex-1 leading-snug">{{ errorMessage }}</span>
            <button
              type="button"
              @click="errorMessage = ''"
              aria-label="Tutup pesan"
              class="text-rose-400 hover:text-rose-700 cursor-pointer"
            >
              &times;
            </button>
          </div>

          <div
            v-else-if="passwordSuccessMessage"
            role="status"
            class="mb-6 flex items-center gap-3 p-3.5 bg-emerald-50 border border-emerald-200/80 text-emerald-800 rounded-2xl text-xs sm:text-sm font-medium"
          >
            <span class="w-2 h-2 rounded-full bg-emerald-500 shrink-0" />
            <span class="flex-1 leading-snug">{{ passwordSuccessMessage }}</span>
          </div>
        </Transition>

        <!-- Dynamic State View: Login vs Force Password -->
        <Transition
          mode="out-in"
          enter-active-class="transition duration-250 ease-out"
          enter-from-class="opacity-0 translate-y-2"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 -translate-y-2"
        >
          <!-- STATE 1: STANDARD FLOEMA LOGIN -->
          <div v-if="currentStep === 'login'" key="floema-login" class="space-y-8">
            <!-- Headline & Framing Description -->
            <div class="space-y-2">
              <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-sky-50 border border-sky-200/60 text-sky-800 text-[11px] font-bold uppercase tracking-wider">
                <span class="w-1.5 h-1.5 rounded-full bg-sky-600" />
                <span>Otentikasi Akun</span>
              </div>
              <h1 class="text-2xl sm:text-3xl font-medium tracking-tight text-slate-900 leading-snug">
                Akses Berkas &amp; Persuratan
              </h1>
              <p class="text-sm text-slate-500 font-normal leading-relaxed">
                Silakan masukkan kredensial resmi pengurus DOSCOM untuk membuka repositori.
              </p>
            </div>

            <!-- Login Form -->
            <form @submit.prevent="handleLogin" class="space-y-5" novalidate>
              <!-- Field: Username / NIM -->
              <div class="space-y-2">
                <Label for="username" class="block text-xs font-bold uppercase tracking-wider text-slate-700">
                  Username atau NIM
                </Label>
                <Input
                  id="username"
                  v-model="username"
                  type="text"
                  autocomplete="username"
                  required
                  placeholder="sekre atau humas"
                  class="text-sm sm:text-base py-3"
                />
              </div>

              <!-- Field: Password -->
              <div class="space-y-2">
                <div class="flex items-center justify-between">
                  <Label for="password" class="block text-xs font-bold uppercase tracking-wider text-slate-700">
                    Kata Sandi
                  </Label>
                  <Button
                    type="button"
                    variant="link"
                    size="sm"
                    @click="showPassword = !showPassword"
                    class="h-auto p-0 text-xs font-bold text-sky-600 hover:text-sky-700"
                  >
                    {{ showPassword ? 'Sembunyikan' : 'Perlihatkan' }}
                  </Button>
                </div>
                <Input
                  id="password"
                  v-model="password"
                  :type="showPassword ? 'text' : 'password'"
                  autocomplete="current-password"
                  required
                  placeholder="Masukkan kata sandi"
                  class="text-sm sm:text-base py-3 tracking-wider"
                />
              </div>

              <!-- Pill Button CTA (Floema Hallmark) -->
              <div class="pt-3">
                <Button
                  type="submit"
                  size="lg"
                  :disabled="isLoading"
                  class="w-full text-sm sm:text-base"
                >
                  <span v-if="isLoading" class="inline-block w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span>{{ isLoading ? 'Memverifikasi...' : 'Masuk ke Repositori' }}</span>
                  <span v-if="!isLoading" class="text-sm tracking-widest">&rarr;</span>
                </Button>
              </div>
            </form>


          </div>

          <!-- STATE 2: FORCE PASSWORD ROTATION (FLOEMA MINIMALIST ARCHITECTURE) -->
          <div v-else key="floema-force-password" class="space-y-8">
            <div class="space-y-2">
              <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-orange-50 border border-orange-200/60 text-orange-800 text-[11px] font-bold uppercase tracking-wider">
                <span class="w-1.5 h-1.5 rounded-full bg-orange-500" />
                <span>Rotasi Keamanan</span>
              </div>
              <h1 class="text-2xl sm:text-3xl font-medium tracking-tight text-slate-900 leading-snug">
                Pembaruan Kata Sandi
              </h1>
              <p class="text-sm text-slate-500 font-normal leading-relaxed">
                Akun pengurus <strong class="text-slate-800">{{ loggedUsername }}</strong> masih memakai kredensial awal. Tentukan sandi baru untuk melanjutkan.
              </p>
            </div>

            <form @submit.prevent="handleForceChangePassword" class="space-y-5" novalidate>
              <!-- Field: New Password -->
              <div class="space-y-2">
                <div class="flex items-center justify-between">
                  <Label for="newPass" class="block text-xs font-bold uppercase tracking-wider text-slate-700">
                    Kata Sandi Baru
                  </Label>
                  <Button
                    type="button"
                    variant="link"
                    size="sm"
                    @click="showNewPassword = !showNewPassword"
                    class="h-auto p-0 text-xs font-bold text-sky-600 hover:text-sky-700"
                  >
                    {{ showNewPassword ? 'Sembunyikan' : 'Perlihatkan' }}
                  </Button>
                </div>
                <Input
                  id="newPass"
                  v-model="newPassword"
                  :type="showNewPassword ? 'text' : 'password'"
                  required
                  placeholder="Minimal 6 karakter"
                  class="text-sm sm:text-base py-3"
                />
              </div>

              <!-- Field: Confirm Password -->
              <div class="space-y-2">
                <div class="flex items-center justify-between">
                  <Label for="confPass" class="block text-xs font-bold uppercase tracking-wider text-slate-700">
                    Konfirmasi Kata Sandi
                  </Label>
                  <Button
                    type="button"
                    variant="link"
                    size="sm"
                    @click="showConfirmPassword = !showConfirmPassword"
                    class="h-auto p-0 text-xs font-bold text-sky-600 hover:text-sky-700"
                  >
                    {{ showConfirmPassword ? 'Sembunyikan' : 'Perlihatkan' }}
                  </Button>
                </div>
                <Input
                  id="confPass"
                  v-model="confirmPassword"
                  :type="showConfirmPassword ? 'text' : 'password'"
                  required
                  placeholder="Ketik ulang kata sandi baru"
                  class="text-sm sm:text-base py-3"
                />
              </div>

              <!-- Criteria Badges -->
              <div class="p-3.5 rounded-2xl bg-slate-50 border border-slate-200/70 flex flex-wrap gap-4 text-xs font-semibold">
                <div class="flex items-center gap-1.5" :class="isMinLength ? 'text-emerald-700' : 'text-slate-400'">
                  <span class="w-1.5 h-1.5 rounded-full" :class="isMinLength ? 'bg-emerald-500' : 'bg-slate-300'" />
                  <span>Min. 6 Karakter</span>
                </div>
                <div class="flex items-center gap-1.5" :class="isPasswordMatched ? 'text-emerald-700' : 'text-slate-400'">
                  <span class="w-1.5 h-1.5 rounded-full" :class="isPasswordMatched ? 'bg-emerald-500' : 'bg-slate-300'" />
                  <span>Sandi Cocok</span>
                </div>
              </div>

              <!-- Submit Button (DOSCOM Orange Accent Pill) -->
              <div class="space-y-3 pt-2">
                <Button
                  type="submit"
                  variant="orange"
                  size="lg"
                  :disabled="!canSubmitPasswordChange"
                  class="w-full text-sm sm:text-base"
                >
                  <span v-if="isUpdatingPassword" class="inline-block w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  <span>{{ isUpdatingPassword ? 'Menyimpan...' : 'Konfirmasi & Masuk' }}</span>
                  <span v-if="!isUpdatingPassword" class="text-sm tracking-widest">&rarr;</span>
                </Button>

                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  @click="handleSkipPasswordChange"
                  class="w-full text-xs font-bold text-slate-400 hover:text-slate-700 tracking-wider uppercase"
                >
                  Lewati untuk sementara &gt;
                </Button>
              </div>
            </form>
          </div>
        </Transition>
      </div>
    </main>

    <!-- ================= FLOEMA ARCHITECTURAL FOOTER ================= -->
    <footer class="relative z-10 w-full max-w-5xl mx-auto flex flex-col sm:flex-row items-center justify-between text-xs font-medium text-slate-400 py-2 border-t border-slate-200/60 gap-2">
      <div>
        <span>&copy; {{ new Date().getFullYear() }} DOSCOM</span>
        <span class="mx-2">&bull;</span>
        <span>Kelompok Studi Linux &amp; Open Source</span>
      </div>
      <div class="flex items-center gap-4 text-[11px] text-slate-400">
        <span>LACI PERSURATAN // PORTAL RESMI</span>
      </div>
    </footer>
  </div>
</template>
