<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  suratService,
  authService,
  userService,
  currentUser,
  currentUsername,
  type Proker,
  type User,
} from '@/services/api'
import { useDocumentScanner } from '@/composables/useDocumentScanner'

const props = defineProps<{
  isOpen: boolean
  prokerList: Proker[]
  statusColumns: string[]
  isHumas?: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created'): void
}>()

const isCreating = ref(false)
const romanMonths = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X', 'XI', 'XII']

const noUrut = ref('01')
const selectedProker = ref('')
const selectedBulan = ref(romanMonths[new Date().getMonth()] || 'X')
const selectedTahun = ref(new Date().getFullYear().toString())
const scannedExtraProker = ref<string | null>(null)

// Daftar proker yang menggabungkan master proker dan proker terdeteksi dari surat
const availableProkers = computed(() => {
  const list = [...props.prokerList]
  if (
    scannedExtraProker.value &&
    !list.some((p) => p.nama_proker.toLowerCase() === scannedExtraProker.value?.toLowerCase())
  ) {
    list.push({ id: 'scanned-proker', nama_proker: scannedExtraProker.value })
  }
  return list
})

// Daftar pengurus terdaftar untuk pilihan PIC (khusus Sekre)
const userList = ref<User[]>([])
const isLoadingUsers = ref(false)

const loadUsers = async () => {
  if (props.isHumas) return
  isLoadingUsers.value = true
  try {
    userList.value = await userService.getAll()
  } catch (err) {
    console.error('Gagal memuat daftar pengurus untuk PIC', err)
  } finally {
    isLoadingUsers.value = false
  }
}

const { isScanning, scanProgress, scanImage } = useDocumentScanner()
const fileInputRef = ref<HTMLInputElement | null>(null)
const scanFeedback = ref('')

const createForm = ref({
  nomor_surat: '',
  perihal: '',
  status_saat_ini: 'Standby',
  pic_nama: '',
  catatan: '',
})

const getLoggedInPicName = () => {
  return (
    currentUser.value?.nama ||
    currentUser.value?.username ||
    currentUsername.value ||
    (typeof window !== 'undefined' ? localStorage.getItem('username') : '') ||
    ''
  )
}

const resetForm = () => {
  noUrut.value = '01'
  scannedExtraProker.value = null
  const firstProker = props.prokerList[0]
  selectedProker.value = firstProker ? firstProker.nama_proker : ''
  const monthIdx = new Date().getMonth()
  selectedBulan.value = (monthIdx >= 0 && monthIdx < romanMonths.length && romanMonths[monthIdx]) ? romanMonths[monthIdx]! : 'X'
  selectedTahun.value = new Date().getFullYear().toString()
  scanFeedback.value = ''

  createForm.value = {
    nomor_surat: '',
    perihal: '',
    status_saat_ini: 'Standby',
    pic_nama: getLoggedInPicName(),
    catatan: '',
  }

  if (!createForm.value.pic_nama && authService.isAuthenticated()) {
    authService.getProfile().then((profile) => {
      if (profile?.nama) createForm.value.pic_nama = profile.nama
      else if (profile?.username) createForm.value.pic_nama = profile.username
    }).catch(() => {})
  }
}

const handleFileChange = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  scanFeedback.value = ''
  try {
    const result = await scanImage(file)

    if (result.perihal) {
      createForm.value.perihal = result.perihal.replace(/^[:;.\s-]+/, '').trim()
    }

    if (result.noUrut) {
      noUrut.value = result.noUrut
    }

    if (result.prokerCode) {
      const codeClean = result.prokerCode.toLowerCase().replace(/[^a-z0-9]/g, '')
      const matched = props.prokerList.find((p) => {
        const pClean = p.nama_proker.toLowerCase().replace(/[^a-z0-9]/g, '')
        return pClean === codeClean || codeClean.includes(pClean) || pClean.includes(codeClean)
      })

      if (matched) {
        selectedProker.value = matched.nama_proker
      } else {
        scannedExtraProker.value = result.prokerCode
        selectedProker.value = result.prokerCode
      }
    }

    if (result.bulanRomawi && romanMonths.includes(result.bulanRomawi.toUpperCase())) {
      selectedBulan.value = result.bulanRomawi.toUpperCase()
    }

    if (result.tahun && /^\d{4}$/.test(result.tahun)) {
      selectedTahun.value = result.tahun
    }

    if (result.perihal || result.noUrut || result.prokerCode) {
      scanFeedback.value = 'Data surat DOSCOM berhasil diekstrak dan diisi ke form.'
    } else {
      scanFeedback.value = 'Teks terbaca namun struktur nomor surat DOSCOM tidak terdeteksi.'
    }
  } catch (err: any) {
    scanFeedback.value = 'Gagal memindai: ' + (err?.message || 'Error OCR')
  } finally {
    if (fileInputRef.value) fileInputRef.value.value = ''
  }
}

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      resetForm()
      loadUsers()
    }
  }
)

const handleCreateSurat = async () => {
  if (!noUrut.value.trim()) {
    alert('Nomor urut surat wajib diisi.')
    return
  }
  if (!selectedProker.value) {
    alert('Silakan pilih Program Kerja (Proker).')
    return
  }
  if (!selectedTahun.value.trim()) {
    alert('Tahun surat wajib diisi.')
    return
  }

  const prokerCode = selectedProker.value.toUpperCase().replace(/\s+/g, '-')
  const paddedNo = noUrut.value.trim().padStart(2, '0')
  createForm.value.nomor_surat = `${paddedNo}/DOSCOM/${prokerCode}/${selectedBulan.value}/${selectedTahun.value.trim()}`

  if (props.isHumas) {
    createForm.value.pic_nama = getLoggedInPicName()
  }

  isCreating.value = true
  try {
    await suratService.create(createForm.value)
    emit('created')
    emit('close')
  } catch (err: any) {
    alert(err.response?.data?.error || 'Gagal membuat surat')
  } finally {
    isCreating.value = false
  }
}
</script>

<template>
  <Transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0 scale-95"
    enter-to-class="opacity-100 scale-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="opacity-100 scale-100"
    leave-to-class="opacity-0 scale-95"
  >
    <div
      v-if="isOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 overflow-y-auto"
    >
      <div class="bg-white rounded-3xl border border-black/10 max-w-xl w-full p-6 sm:p-7 space-y-5 shadow-xl">
        <header>
          <h3 class="text-lg font-semibold text-zinc-900 tracking-tight">Tambah Surat Baru</h3>
          <p class="text-xs text-zinc-500 mt-0.5">Lengkapi parameter nomor surat dan informasi dokumen resmi.</p>
        </header>

        <!-- OCR Auto-Fill Area -->
        <div class="p-3 bg-zinc-50 border border-dashed border-zinc-300 rounded-2xl flex flex-col gap-2">
          <div class="flex items-center justify-between gap-3">
            <div class="flex items-center gap-2.5">
              <div class="w-8 h-8 rounded-xl bg-[#0A84DC]/10 text-[#0A84DC] flex items-center justify-center shrink-0">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </div>
              <div>
                <p class="text-xs font-semibold text-zinc-800">Auto-Fill via Scan Dokumen</p>
                <p class="text-[10px] text-zinc-500">Pindai foto fisik surat untuk mengisi No. Surat & Perihal</p>
              </div>
            </div>
            <label
              class="px-3 py-1.5 text-xs font-semibold rounded-xl bg-white border border-zinc-200 text-zinc-700 hover:bg-zinc-100 shadow-2xs cursor-pointer inline-flex items-center gap-1.5 transition active:scale-95 shrink-0"
              :class="{ 'opacity-50 pointer-events-none': isScanning }"
            >
              <input
                ref="fileInputRef"
                type="file"
                accept="image/*"
                capture="environment"
                class="hidden"
                :disabled="isScanning"
                @change="handleFileChange"
              />
              <span>{{ isScanning ? 'Memindai...' : 'Ambil Foto / Scan' }}</span>
            </label>
          </div>

          <!-- Progress Bar -->
          <div v-if="isScanning" class="space-y-1 pt-1">
            <div class="flex justify-between text-[10px] text-zinc-500 font-medium">
              <span>Membaca teks dokumen (OCR)...</span>
              <span>{{ scanProgress }}%</span>
            </div>
            <div class="w-full bg-zinc-200 rounded-full h-1.5 overflow-hidden">
              <div
                class="bg-[#0A84DC] h-full transition-all duration-200 ease-out rounded-full"
                :style="{ width: `${scanProgress}%` }"
              ></div>
            </div>
          </div>

          <p v-else-if="scanFeedback" class="text-[11px] text-emerald-600 font-medium px-0.5">
            {{ scanFeedback }}
          </p>
        </div>

        <form @submit.prevent="handleCreateSurat" class="space-y-4">
          <!-- Inset Grouped Form Container -->
          <div class="rounded-2xl border border-black/10 bg-white overflow-hidden divide-y divide-black/5 shadow-2xs">
            <!-- 1. Nomor Surat Group (Format Standar DOSCOM) -->
            <div class="p-3.5 space-y-2 focus-within:bg-zinc-50/50 transition-colors">
              <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">
                Nomor Surat
              </label>

              <div class="flex flex-row flex-wrap items-center gap-2">
                <input
                  v-model="noUrut"
                  type="text"
                  required
                  placeholder="01"
                  class="w-16 sm:w-20 px-2 py-1.5 text-xs font-mono font-semibold text-center border border-black/10 rounded-xl bg-black/[0.02] focus:outline-none focus:ring-1 focus:ring-[#0A84DC]"
                />

                <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

                <input
                  type="text"
                  value="DOSCOM"
                  disabled
                  class="w-20 sm:w-24 px-2 py-1.5 text-xs font-mono font-semibold text-center border border-black/10 rounded-xl bg-black/[0.04] text-zinc-600 cursor-not-allowed select-none"
                />

                <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

                <select
                  v-model="selectedProker"
                  required
                  class="flex-1 min-w-[120px] px-2.5 py-1.5 text-xs font-medium border border-black/10 rounded-xl bg-white focus:outline-none focus:ring-1 focus:ring-[#0A84DC] truncate"
                >
                  <option value="" disabled>-- Pilih Proker --</option>
                  <option v-for="proker in availableProkers" :key="proker.id" :value="proker.nama_proker">
                    {{ proker.nama_proker }}
                  </option>
                </select>

                <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

                <select
                  v-model="selectedBulan"
                  required
                  class="w-16 sm:w-20 px-2 py-1.5 text-xs font-mono font-semibold text-center border border-black/10 rounded-xl bg-white focus:outline-none focus:ring-1 focus:ring-[#0A84DC]"
                >
                  <option v-for="bulan in romanMonths" :key="bulan" :value="bulan">
                    {{ bulan }}
                  </option>
                </select>

                <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

                <input
                  v-model="selectedTahun"
                  type="number"
                  required
                  placeholder="2026"
                  class="min-w-[4rem] flex-1 sm:w-20 px-2 py-1.5 text-xs font-mono font-semibold text-center border border-black/10 rounded-xl bg-white focus:outline-none focus:ring-1 focus:ring-[#0A84DC]"
                />
              </div>

              <p class="text-[11px] text-zinc-500">
                Pratinjau: <span class="font-mono text-[#0A84DC] font-semibold">{{ (noUrut || '01') }}/DOSCOM/{{ (selectedProker || 'PROKER').toUpperCase().replace(/\s+/g, '-') }}/{{ selectedBulan }}/{{ selectedTahun }}</span>
              </p>
            </div>

            <!-- 2. Perihal -->
            <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
              <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Perihal</label>
              <input
                v-model="createForm.perihal"
                type="text"
                required
                placeholder="Undangan Kerjasama Dies Natalis"
                class="w-full text-xs font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 pt-0.5"
              />
            </div>

            <!-- 3. Nama PIC -->
            <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
              <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Nama PIC</label>
              <!-- Akun Humas: Terkunci otomatis ke nama pengurus yang sedang login -->
              <div v-if="isHumas" class="text-xs font-semibold text-zinc-700 pt-0.5">
                {{ createForm.pic_nama || getLoggedInPicName() || '-' }}
                <span class="text-[10px] text-zinc-400 font-normal ml-1">(Otomatis Akun Humas)</span>
              </div>
              <!-- Akun Sekre: Dropdown daftar pengurus terdaftar -->
              <div v-else>
                <select
                  v-model="createForm.pic_nama"
                  required
                  class="w-full text-xs font-semibold text-zinc-900 bg-transparent focus:outline-none pt-0.5 cursor-pointer"
                >
                  <option value="" disabled>-- Pilih Pengurus PIC --</option>
                  <option v-for="user in userList" :key="user.id" :value="user.nama || user.username">
                    {{ user.nama }} ({{ user.divisi || user.role || 'Pengurus' }})
                  </option>
                  <option
                    v-if="createForm.pic_nama && !userList.some((u) => (u.nama || u.username) === createForm.pic_nama)"
                    :value="createForm.pic_nama"
                  >
                    {{ createForm.pic_nama }}
                  </option>
                </select>
              </div>
            </div>

            <!-- 4. Status Awal -->
            <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
              <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Status Awal</label>
              <select
                v-if="!isHumas"
                v-model="createForm.status_saat_ini"
                class="w-full text-xs font-semibold text-zinc-900 bg-transparent focus:outline-none pt-0.5"
              >
                <option v-for="col in statusColumns" :key="col" :value="col">
                  {{ col }}
                </option>
              </select>
              <div v-else class="text-xs font-semibold text-zinc-700 pt-0.5">
                Standby (Otomatis)
              </div>
            </div>

            <!-- 5. Catatan Awal -->
            <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
              <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Catatan Awal (Opsional)</label>
              <textarea
                v-model="createForm.catatan"
                rows="2"
                placeholder="Keterangan instruksi..."
                class="w-full text-xs font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 resize-none pt-0.5"
              ></textarea>
            </div>
          </div>

          <footer class="flex justify-end gap-2.5 pt-2">
            <button
              type="button"
              @click="emit('close')"
              class="px-4 py-2 text-xs font-semibold border border-black/10 rounded-2xl text-zinc-700 hover:bg-black/5 transition-all duration-200 active:scale-95 cursor-pointer"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isCreating"
              class="px-5 py-2 text-xs font-semibold bg-[#0A84DC] text-white rounded-2xl hover:bg-[#0872be] shadow-md disabled:opacity-50 transition-all duration-200 active:scale-95 cursor-pointer"
            >
              {{ isCreating ? 'Menyimpan...' : 'Simpan Surat' }}
            </button>
          </footer>
        </form>
      </div>
    </div>
  </Transition>
</template>
