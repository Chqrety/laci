<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { suratService, type Surat } from '@/services/api'
import CardDetail from '@/components/CardDetail.vue'

const suratList = ref<Surat[]>([])
const isLoading = ref(false)
const errorMessage = ref('')
const deletingId = ref<string | null>(null)

// Detail modal
const selectedSurat = ref<Surat | null>(null)
const isDetailOpen = ref(false)

// Filter & Pagination state
const searchQuery = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(20)

const statusOptions = [
  'Standby',
  'Cetak',
  'TTD Lapis 1',
  'TTD Ketum',
  'TTD Pembina',
  'Paraf Koormawa',
  'TTD Tertinggi',
  'Selesai',
]

const fetchSurat = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    suratList.value = await suratService.getAll()
    if (selectedSurat.value) {
      selectedSurat.value = suratList.value.find((s) => s.id === selectedSurat.value?.id) || null
    }
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || err.message || 'Gagal memuat data surat'
  } finally {
    isLoading.value = false
  }
}

watch([searchQuery, statusFilter], () => {
  currentPage.value = 1
})

const filteredSurat = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  const sf = statusFilter.value.trim().toLowerCase()

  return suratList.value.filter((s) => {
    const matchSearch =
      !q || s.nomor_surat.toLowerCase().includes(q) || s.perihal.toLowerCase().includes(q)
    const matchStatus = !sf || s.status_saat_ini.toLowerCase() === sf
    return matchSearch && matchStatus
  })
})

const totalPages = computed(() => Math.ceil(filteredSurat.value.length / itemsPerPage.value) || 1)

watch(totalPages, (newTotal) => {
  if (currentPage.value > newTotal) currentPage.value = newTotal
})

const paginatedSurat = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredSurat.value.slice(start, start + itemsPerPage.value)
})

const openDetail = (surat: Surat) => {
  selectedSurat.value = surat
  isDetailOpen.value = true
}

const closeDetail = () => {
  isDetailOpen.value = false
  selectedSurat.value = null
}

const getStatusBadgeClass = (status: string) => {
  const s = status.toLowerCase()
  if (s.includes('selesai')) return 'bg-emerald-50 text-emerald-700 border-emerald-200'
  if (s.includes('cetak')) return 'bg-sky-50 text-sky-700 border-sky-200'
  if (s.includes('lapis') || s.includes('ketum')) return 'bg-amber-50 text-amber-700 border-amber-200'
  if (s.includes('pembina') || s.includes('koormawa') || s.includes('tertinggi')) {
    return 'bg-purple-50 text-purple-700 border-purple-200'
  }
  return 'bg-zinc-100 text-zinc-700 border-zinc-200'
}

const handleDeleteSurat = async (surat: Surat) => {
  if (!window.confirm(`Hapus surat "${surat.nomor_surat}" beserta semua log riwayatnya?`)) return

  deletingId.value = surat.id
  errorMessage.value = ''
  try {
    await suratService.delete(surat.id)
    if (selectedSurat.value?.id === surat.id) closeDetail()
    await fetchSurat()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || err.message || 'Gagal menghapus surat'
  } finally {
    deletingId.value = null
  }
}

onMounted(fetchSurat)
</script>

<template>
  <div class="max-w-6xl mx-auto space-y-6 w-full">
    <!-- Top Bar (Mirip dengan ProkerManagement.vue) -->
    <header class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-2 px-1">
      <div>
        <h2 class="text-xl font-semibold tracking-tight text-zinc-900">Manajemen Surat</h2>
        <p class="text-xs text-zinc-500 mt-0.5">Kelola data arsip dan lacak seluruh status surat keluar DOSCOM.</p>
      </div>

      <!-- Filter Bar di luar card tabel, sejajar dengan judul -->
      <div class="flex items-center flex-wrap gap-2">
        <input
          v-model="searchQuery"
          type="search"
          placeholder="Cari nomor atau perihal..."
          class="px-3.5 py-2 text-xs rounded-2xl bg-white/80 border border-black/5 text-zinc-800 placeholder:text-zinc-400 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all w-48 sm:w-60"
        />

        <select
          v-model="statusFilter"
          class="px-3 py-2 text-xs rounded-2xl bg-white/80 border border-black/5 text-zinc-800 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all cursor-pointer"
        >
          <option value="">Semua Status</option>
          <option v-for="status in statusOptions" :key="status" :value="status">
            {{ status }}
          </option>
        </select>

        <button
          type="button"
          @click="fetchSurat"
          :disabled="isLoading"
          class="px-4 py-2 text-xs font-semibold border border-black/5 rounded-2xl bg-white/80 text-zinc-700 hover:bg-white shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all active:scale-[0.98] disabled:opacity-50 shrink-0"
        >
          {{ isLoading ? 'Memuat...' : 'Muat Ulang' }}
        </button>
      </div>
    </header>

    <!-- Error Notice -->
    <div
      v-if="errorMessage"
      role="alert"
      class="p-3 text-xs bg-rose-50 border border-rose-200 text-rose-700 rounded-2xl font-medium"
    >
      {{ errorMessage }}
    </div>

    <!-- Table Container (Persis seperti ProkerManagement: bg-white/85 backdrop-blur-xl border border-black/5 rounded-2xl) -->
    <div class="bg-white/85 backdrop-blur-xl border border-black/5 rounded-2xl shadow-[0_8px_30px_rgb(0,0,0,0.05)] overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs divide-y divide-black/5 min-w-[760px]">
          <!-- Header Tabel Bersih: Putih menyatu tanpa background abu-abu tebal -->
          <thead class="text-zinc-500 font-semibold">
            <tr>
              <th scope="col" class="px-5 py-4 w-14 text-center">No</th>
              <th scope="col" class="px-5 py-4 w-52 font-mono">Nomor Surat</th>
              <th scope="col" class="px-5 py-4">Perihal</th>
              <th scope="col" class="px-5 py-4 w-40">PIC</th>
              <th scope="col" class="px-5 py-4 w-36">Status</th>
              <th scope="col" class="px-5 py-4 w-44 text-center">Aksi</th>
            </tr>
          </thead>

          <tbody class="divide-y divide-black/5 text-zinc-800">
            <!-- Loading Row -->
            <tr v-if="isLoading && suratList.length === 0">
              <td colspan="6" class="px-5 py-12 text-center text-zinc-400">
                Memuat data surat...
              </td>
            </tr>

            <!-- Empty Row -->
            <tr v-else-if="filteredSurat.length === 0">
              <td colspan="6" class="px-5 py-12 text-center text-zinc-400">
                {{ searchQuery || statusFilter ? 'Tidak ada surat yang sesuai kriteria pencarian.' : 'Belum ada data surat tersedia.' }}
              </td>
            </tr>

            <!-- Data Rows -->
            <tr
              v-for="(surat, index) in paginatedSurat"
              :key="surat.id"
              class="hover:bg-black/[0.02] transition-colors"
            >
              <td class="px-5 py-4 font-mono text-zinc-400 text-center">
                {{ (currentPage - 1) * itemsPerPage + index + 1 }}
              </td>
              <td class="px-5 py-4 font-mono font-medium text-zinc-900 whitespace-nowrap">
                {{ surat.nomor_surat }}
              </td>
              <td class="px-5 py-4 font-semibold text-zinc-900 leading-snug">
                <div>{{ surat.perihal }}</div>
                <div v-if="surat.catatan" class="text-[11px] font-normal text-zinc-500 mt-0.5 line-clamp-1">
                  {{ surat.catatan }}
                </div>
              </td>
              <td class="px-5 py-4 text-zinc-700 font-medium whitespace-nowrap">
                {{ surat.pic_nama }}
              </td>
              <td class="px-5 py-4 whitespace-nowrap">
                <span
                  class="inline-flex items-center px-2.5 py-1 text-[11px] font-semibold rounded-xl border"
                  :class="getStatusBadgeClass(surat.status_saat_ini)"
                >
                  {{ surat.status_saat_ini }}
                </span>
              </td>
              <td class="px-5 py-4 text-center whitespace-nowrap">
                <div class="inline-flex items-center justify-center gap-2">
                  <!-- Tombol Detail (Pill-shape) -->
                  <button
                    type="button"
                    @click="openDetail(surat)"
                    class="px-4 py-1.5 text-xs font-semibold rounded-full border border-black/10 bg-white text-zinc-700 hover:bg-black/5 hover:text-zinc-900 transition-all shadow-2xs inline-flex items-center gap-1 active:scale-95 cursor-pointer"
                    title="Lihat Detail Surat"
                  >
                    <svg class="w-3.5 h-3.5 text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <span>Detail</span>
                  </button>

                  <!-- Tombol Hapus (Pill-shape) -->
                  <button
                    type="button"
                    @click="handleDeleteSurat(surat)"
                    :disabled="deletingId === surat.id"
                    class="px-4 py-1.5 text-xs font-semibold rounded-full border border-rose-200/80 bg-rose-50/80 text-rose-600 hover:bg-rose-100 hover:text-rose-700 disabled:opacity-50 transition-all shadow-2xs inline-flex items-center gap-1 active:scale-95 cursor-pointer"
                    title="Hapus Surat"
                  >
                    <svg class="w-3.5 h-3.5 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                    <span>{{ deletingId === surat.id ? '...' : 'Hapus' }}</span>
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <footer
        v-if="filteredSurat.length > 0"
        class="flex flex-col sm:flex-row items-center justify-between px-5 py-3.5 border-t border-black/5 bg-black/[0.01] text-xs text-zinc-500 gap-3"
      >
        <div>
          Menampilkan
          <span class="font-semibold text-zinc-800">
            {{ (currentPage - 1) * itemsPerPage + 1 }}
          </span>
          -
          <span class="font-semibold text-zinc-800">
            {{ Math.min(currentPage * itemsPerPage, filteredSurat.length) }}
          </span>
          dari
          <span class="font-semibold text-zinc-800">{{ filteredSurat.length }}</span>
          surat
        </div>

        <div class="flex items-center gap-2">
          <button
            type="button"
            @click="currentPage--"
            :disabled="currentPage <= 1"
            class="px-3.5 py-1.5 text-xs font-semibold rounded-2xl border border-black/10 bg-white text-zinc-700 hover:bg-black/5 disabled:opacity-40 disabled:cursor-not-allowed transition-all shadow-2xs active:scale-95 cursor-pointer"
          >
            Previous
          </button>

          <span class="text-xs font-medium text-zinc-600 px-1">
            Halaman <span class="font-semibold text-zinc-900">{{ currentPage }}</span> dari <span class="font-semibold text-zinc-900">{{ totalPages }}</span>
          </span>

          <button
            type="button"
            @click="currentPage++"
            :disabled="currentPage >= totalPages"
            class="px-3.5 py-1.5 text-xs font-semibold rounded-2xl border border-black/10 bg-white text-zinc-700 hover:bg-black/5 disabled:opacity-40 disabled:cursor-not-allowed transition-all shadow-2xs active:scale-95 cursor-pointer"
          >
            Next
          </button>
        </div>
      </footer>
    </div>

    <!-- Modal Detail Component Integration -->
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <CardDetail
        v-if="isDetailOpen && selectedSurat"
        :surat="selectedSurat"
        :status-options="statusOptions"
        @close="closeDetail"
        @updated="fetchSurat"
      />
    </Transition>
  </div>
</template>
