<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { suratService, STATUS_COLUMNS, type Surat } from '@/services/api'
import CardDetail from '@/components/molecules/CardDetail.vue'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { toast } from '@/components/ui/toast'

const suratList = ref<Surat[]>([])
const isLoading = ref(false)
const errorMessage = ref('')
const deletingId = ref<string | null>(null)

// Delete confirm dialog
const isDeleteModalOpen = ref(false)
const suratToDelete = ref<Surat | null>(null)
const isDeleting = ref(false)

// Detail modal
const selectedSurat = ref<Surat | null>(null)
const isDetailOpen = ref(false)

// Filter & Pagination state
const searchQuery = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const itemsPerPage = ref(20)

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
      !q ||
      s.nomor_surat.toLowerCase().includes(q) ||
      s.perihal.toLowerCase().includes(q) ||
      (s.tujuan && s.tujuan.toLowerCase().includes(q))
    const matchStatus = !sf || sf === 'all' || s.status_saat_ini.toLowerCase() === sf
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

const getStatusBadgeVariant = (status: string) => {
  const s = status.toLowerCase()
  if (s.includes('selesai')) return 'emerald'
  if (s.includes('cetak')) return 'default'
  if (s.includes('lapis') || s.includes('ketum')) return 'amber'
  if (s.includes('pembina') || s.includes('koormawa') || s.includes('tertinggi')) return 'purple'
  return 'secondary'
}

const confirmDeleteSurat = (surat: Surat) => {
  suratToDelete.value = surat
  isDeleteModalOpen.value = true
}

const executeDeleteSurat = async () => {
  if (!suratToDelete.value) return

  const surat = suratToDelete.value
  deletingId.value = surat.id
  isDeleting.value = true
  try {
    await suratService.delete(surat.id)
    if (selectedSurat.value?.id === surat.id) closeDetail()
    toast({
      title: 'Berhasil',
      description: `Surat "${surat.nomor_surat}" berhasil dihapus`,
      variant: 'success',
    })
    isDeleteModalOpen.value = false
    suratToDelete.value = null
    await fetchSurat()
  } catch (err: any) {
    toast({
      title: 'Gagal Menghapus Surat',
      description: err.response?.data?.error || err.message || 'Gagal menghapus surat',
      variant: 'destructive',
    })
  } finally {
    deletingId.value = null
    isDeleting.value = false
  }
}

onMounted(fetchSurat)
</script>

<template>
  <div class="max-w-6xl mx-auto space-y-6 w-full">
    <!-- Top Bar -->
    <header class="flex flex-col md:flex-row md:items-center md:justify-between gap-4 pb-2 px-1">
      <div>
        <div class="flex items-center gap-2.5">
          <h2 class="text-xl font-medium tracking-tight text-slate-900 leading-snug">Surat</h2>
          <Badge variant="outline" class="font-normal text-[11px] px-2.5 py-0.5 rounded-full text-slate-500 bg-slate-100/70 border-slate-200/80">
            {{ filteredSurat.length }} {{ filteredSurat.length === suratList.length ? 'surat' : `dari ${suratList.length}` }}
          </Badge>
        </div>
        <p class="text-xs text-slate-500 mt-0.5">Kelola data arsip dan lacak seluruh status surat keluar DOSCOM.</p>
      </div>

      <!-- Filter Bar: Responsive adaptif untuk mobile, tablet, dan desktop -->
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2 w-full md:w-auto">
        <!-- Input Pencarian dengan ikon search & tombol clear -->
        <div class="relative w-full sm:w-60 lg:w-72">
          <svg
            class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-4.35-4.35m0 0A7.5 7.5 0 105.65 5.65a7.5 7.5 0 0010.7 10.7z" />
          </svg>
          <Input
            v-model="searchQuery"
            type="search"
            placeholder="Cari nomor atau perihal..."
            class="pl-9 pr-8 h-9 rounded-xl border-slate-200/90 bg-white text-xs shadow-2xs focus-visible:ring-sky-500/20"
          />
          <button
            v-if="searchQuery"
            type="button"
            @click="searchQuery = ''"
            title="Hapus pencarian"
            aria-label="Hapus pencarian"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 p-0.5 rounded-sm"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Kelompok Filter Status & Reload (sejajar di mobile, compact di desktop) -->
        <div class="flex items-center gap-2 w-full sm:w-auto">
          <Select v-model="statusFilter">
            <SelectTrigger class="h-9 w-full sm:w-44 lg:w-48 rounded-xl border-slate-200/90 bg-white text-xs shadow-2xs">
              <SelectValue placeholder="Semua Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">Semua Status</SelectItem>
              <SelectItem v-for="status in STATUS_COLUMNS" :key="status" :value="status">
                {{ status }}
              </SelectItem>
            </SelectContent>
          </Select>

          <Button
            type="button"
            variant="outline"
            size="icon"
            @click="fetchSurat"
            :disabled="isLoading"
            title="Muat Ulang"
            aria-label="Muat Ulang"
            class="h-9 w-9 shrink-0 rounded-xl border-slate-200/90 bg-white shadow-2xs text-slate-600 hover:text-slate-900 active:scale-95 transition-all"
          >
            <svg class="w-4 h-4" :class="{ 'animate-spin': isLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </Button>
        </div>
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
    <div class="bg-white/95 border border-slate-200/90 rounded-3xl shadow-xs overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs divide-y divide-black/5 min-w-[760px]">
          <thead class="bg-black/[0.02] text-zinc-500 font-semibold">
            <tr>
              <th scope="col" class="px-5 py-4 w-14 text-center">No</th>
              <th scope="col" class="px-5 py-4 w-52 font-mono">Nomor Surat</th>
              <th scope="col" class="px-5 py-4">Perihal</th>
              <th scope="col" class="px-5 py-4 w-40">PIC</th>
              <th scope="col" class="px-5 py-4 w-36">Status</th>
              <th scope="col" class="px-5 py-4 w-28 text-right">Aksi</th>
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
                <div v-if="surat.tujuan" class="text-[11px] font-medium text-[#0A84DC] mt-0.5 line-clamp-1">
                  Tujuan: {{ surat.tujuan }}
                </div>
                <div v-if="surat.catatan" class="text-[11px] font-normal text-zinc-500 mt-0.5 line-clamp-1">
                  {{ surat.catatan }}
                </div>
              </td>
              <td class="px-5 py-4 text-zinc-700 font-medium whitespace-nowrap">
                {{ surat.pic_nama }}
              </td>
              <td class="px-5 py-4 whitespace-nowrap">
                <Badge
                  :variant="getStatusBadgeVariant(surat.status_saat_ini)"
                  size="sm"
                >
                  {{ surat.status_saat_ini }}
                </Badge>
              </td>
              <td class="px-5 py-3 text-right whitespace-nowrap">
                <div class="inline-flex items-center justify-end gap-1">
                  <!-- Tombol Detail Surat (Eye Icon) -->
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    @click="openDetail(surat)"
                    class="h-8 w-8 rounded-xl text-zinc-400 hover:text-[#0A84DC] hover:bg-[#0A84DC]/10"
                    title="Lihat Detail Surat"
                    aria-label="Lihat Detail Surat"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.8"
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                      />
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.8"
                        d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                      />
                    </svg>
                  </Button>

                  <!-- Tombol Hapus Surat (Trash Icon) -->
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    @click="confirmDeleteSurat(surat)"
                    :disabled="deletingId === surat.id"
                    class="h-8 w-8 rounded-xl text-zinc-400 hover:text-rose-600 hover:bg-rose-50 disabled:opacity-50"
                    title="Hapus Surat"
                    aria-label="Hapus Surat"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.8"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </Button>
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
          <Button
            type="button"
            variant="outline"
            size="sm"
            @click="currentPage--"
            :disabled="currentPage <= 1"
          >
            Previous
          </Button>

          <span class="text-xs font-medium text-zinc-600 px-1">
            Halaman <span class="font-semibold text-zinc-900">{{ currentPage }}</span> dari <span class="font-semibold text-zinc-900">{{ totalPages }}</span>
          </span>

          <Button
            type="button"
            variant="outline"
            size="sm"
            @click="currentPage++"
            :disabled="currentPage >= totalPages"
          >
            Next
          </Button>
        </div>
      </footer>
    </div>

    <!-- Modal Konfirmasi Hapus Surat -->
    <Dialog :open="isDeleteModalOpen" @update:open="(val: boolean) => { if (!val) isDeleteModalOpen = false }">
      <DialogContent class="max-w-sm text-center">
        <div class="w-12 h-12 rounded-2xl bg-rose-50 text-rose-500 mx-auto flex items-center justify-center border border-rose-100">
          <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </div>
        <DialogHeader class="text-center sm:text-center">
          <DialogTitle class="text-center sm:text-center">Hapus Surat Resmi?</DialogTitle>
          <DialogDescription class="text-center sm:text-center mt-1 leading-relaxed">
            Yakin ingin menghapus surat <b class="text-zinc-900 font-mono">{{ suratToDelete?.nomor_surat }}</b> beserta seluruh log riwayatnya?
          </DialogDescription>
        </DialogHeader>
        <DialogFooter class="flex sm:justify-center gap-2 pt-2">
          <Button type="button" variant="outline" size="sm" class="flex-1" @click="isDeleteModalOpen = false">
            Batal
          </Button>
          <Button type="button" variant="destructive" size="sm" class="flex-1" :disabled="isDeleting" @click="executeDeleteSurat">
            {{ isDeleting ? 'Menghapus...' : 'Ya, Hapus' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Modal Detail Component Integration -->
    <Teleport to="body">
      <Transition name="drawer">
        <CardDetail
          v-if="isDetailOpen && selectedSurat"
          :surat="selectedSurat"
          :status-options="STATUS_COLUMNS"
          @close="closeDetail"
          @updated="fetchSurat"
        />
      </Transition>
    </Teleport>
  </div>
</template>
