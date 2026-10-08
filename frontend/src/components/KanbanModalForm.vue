<script setup lang="ts">
import { ref, watch } from 'vue'
import { suratService, type Proker } from '@/services/api'

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

const noUrut = ref('001')
const selectedProker = ref('')
const selectedBulan = ref(romanMonths[new Date().getMonth()] || 'X')
const selectedTahun = ref(new Date().getFullYear().toString())

const createForm = ref({
  nomor_surat: '',
  perihal: '',
  status_saat_ini: 'Standby',
  pic_nama: '',
  catatan: '',
})

const resetForm = () => {
  noUrut.value = '001'
  const firstProker = props.prokerList[0]
  selectedProker.value = firstProker ? firstProker.nama_proker : ''
  const monthIdx = new Date().getMonth()
  selectedBulan.value = (monthIdx >= 0 && monthIdx < romanMonths.length && romanMonths[monthIdx]) ? romanMonths[monthIdx]! : 'X'
  selectedTahun.value = new Date().getFullYear().toString()

  createForm.value = {
    nomor_surat: '',
    perihal: '',
    status_saat_ini: 'Standby',
    pic_nama: '',
    catatan: '',
  }
}

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      resetForm()
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
  const paddedNo = noUrut.value.trim().padStart(3, '0')
  createForm.value.nomor_surat = `${paddedNo}/DOSCOM/${prokerCode}/${selectedBulan.value}/${selectedTahun.value.trim()}`

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

        <form @submit.prevent="handleCreateSurat" class="space-y-4">
          <!-- Inset Grouped Form Container -->
          <div class="rounded-2xl border border-black/10 bg-white overflow-hidden divide-y divide-black/5 shadow-2xs">
            <!-- 1. Nomor Surat Group -->
            <div class="p-3.5 space-y-2 focus-within:bg-zinc-50/50 transition-colors">
              <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Nomor Surat</label>
              <div class="flex flex-row flex-wrap items-center gap-2">
                <input
                  v-model="noUrut"
                  type="text"
                  required
                  placeholder="001"
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
                  <option v-for="proker in prokerList" :key="proker.id" :value="proker.nama_proker">
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
                Pratinjau: <span class="font-mono text-[#0A84DC] font-semibold">{{ (noUrut || '001') }}/DOSCOM/{{ (selectedProker || 'PROKER').toUpperCase().replace(/\s+/g, '-') }}/{{ selectedBulan }}/{{ selectedTahun }}</span>
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
              <input
                v-model="createForm.pic_nama"
                type="text"
                required
                placeholder="Fikri"
                class="w-full text-xs font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 pt-0.5"
              />
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
