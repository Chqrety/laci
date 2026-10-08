<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getBackendAssetUrl, suratService, type Surat } from '@/services/api'

const router = useRouter()

const suratList = ref<Surat[]>([])
const isLoading = ref<boolean>(false)
const errorMessage = ref<string>('')

// Modal state for Update Status
const isUpdateModalOpen = ref<boolean>(false)
const selectedSurat = ref<Surat | null>(null)
const newStatus = ref<string>('')
const isUpdating = ref<boolean>(false)

// Modal state for Create Surat
const isCreateModalOpen = ref<boolean>(false)
const createForm = ref({
  nomor_surat: '',
  perihal: '',
  status_saat_ini: 'Masuk',
  pic_nama: '',
})
const isCreating = ref<boolean>(false)

const statusOptions = ['Masuk', 'Diproses', 'Verifikasi', 'Disetujui', 'Ditolak', 'Selesai']

const fetchSurat = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    suratList.value = await suratService.getAll()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || err.message || 'Gagal memuat data surat'
  } finally {
    isLoading.value = false
  }
}

const openUpdateModal = (surat: Surat) => {
  selectedSurat.value = surat
  newStatus.value = surat.status_saat_ini
  isUpdateModalOpen.value = true
}

const closeUpdateModal = () => {
  isUpdateModalOpen.value = false
  selectedSurat.value = null
  newStatus.value = ''
}

const handleUpdateStatus = async () => {
  if (!selectedSurat.value || !newStatus.value) return
  isUpdating.value = true
  try {
    await suratService.updateStatus(selectedSurat.value.id, {
      status_saat_ini: newStatus.value,
    })
    closeUpdateModal()
    await fetchSurat()
  } catch (err: any) {
    alert(err.response?.data?.error || 'Gagal memperbarui status')
  } finally {
    isUpdating.value = false
  }
}

const openCreateModal = () => {
  createForm.value = {
    nomor_surat: '',
    perihal: '',
    status_saat_ini: 'Masuk',
    pic_nama: '',
  }
  isCreateModalOpen.value = true
}

const closeCreateModal = () => {
  isCreateModalOpen.value = false
}

const handleCreateSurat = async () => {
  isCreating.value = true
  try {
    await suratService.create(createForm.value)
    closeCreateModal()
    await fetchSurat()
  } catch (err: any) {
    alert(err.response?.data?.error || 'Gagal menambahkan surat')
  } finally {
    isCreating.value = false
  }
}

const goToScanner = (suratId: string) => {
  router.push({ name: 'scanner-surat', params: { id: suratId } })
}

const getStatusBadgeClass = (status: string) => {
  const normalized = status.toLowerCase()
  if (normalized.includes('selesai') || normalized.includes('setuju')) {
    return 'bg-emerald-50 text-emerald-800 border-emerald-300'
  }
  if (normalized.includes('proses') || normalized.includes('verifikasi')) {
    return 'bg-sky-50 text-sky-800 border-sky-300'
  }
  if (normalized.includes('tolak')) {
    return 'bg-rose-50 text-rose-800 border-rose-300'
  }
  return 'bg-zinc-100 text-zinc-700 border-zinc-300'
}

const formatDate = (dateString?: string) => {
  if (!dateString) return '-'
  const date = new Date(dateString)
  return new Intl.DateTimeFormat('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

const getLastLogDate = (logs?: Surat['logs']) => {
  if (!logs || logs.length === 0) return '-'
  const last = logs[logs.length - 1]
  return last ? formatDate(last.waktu_update) : '-'
}

onMounted(() => {
  fetchSurat()
})
</script>

<template>
  <section class="space-y-6">
    <!-- Header bar -->
    <header class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-200">
      <div>
        <h2 class="text-xl font-semibold tracking-tight text-zinc-900">Daftar Surat & Dokumen</h2>
        <p class="text-sm text-zinc-500 mt-1">Kelola arsip, pembaruan status log tracking, dan verifikasi dokumen.</p>
      </div>

      <div class="flex items-center gap-3">
        <button
          type="button"
          @click="fetchSurat"
          :disabled="isLoading"
          class="px-3.5 py-1.5 text-sm font-medium border border-zinc-300 rounded-md bg-white text-zinc-700 hover:bg-zinc-50 focus:outline-none focus:ring-2 focus:ring-zinc-400 disabled:opacity-50"
        >
          {{ isLoading ? 'Memuat...' : 'Muat Ulang' }}
        </button>

        <button
          type="button"
          @click="openCreateModal"
          class="px-3.5 py-1.5 text-sm font-medium rounded-md bg-zinc-900 text-white hover:bg-zinc-800 focus:outline-none focus:ring-2 focus:ring-zinc-700"
        >
          + Tambah Surat
        </button>
      </div>
    </header>

    <!-- Error notice -->
    <div v-if="errorMessage" role="alert" class="p-3 border border-rose-200 bg-rose-50 text-rose-700 text-sm rounded-md">
      {{ errorMessage }}
    </div>

    <!-- Data Table -->
    <div class="overflow-x-auto border border-zinc-200 rounded-lg bg-white shadow-xs">
      <table class="w-full text-left text-sm divide-y divide-zinc-200">
        <thead class="bg-zinc-50 text-zinc-600 font-medium">
          <tr>
            <th scope="col" class="px-4 py-3">Nomor Surat</th>
            <th scope="col" class="px-4 py-3">Perihal</th>
            <th scope="col" class="px-4 py-3">PIC</th>
            <th scope="col" class="px-4 py-3">Status Terkini</th>
            <th scope="col" class="px-4 py-3">Arsip Fisik</th>
            <th scope="col" class="px-4 py-3">Log Terakhir</th>
            <th scope="col" class="px-4 py-3 text-right">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-200 text-zinc-800">
          <tr v-if="isLoading && suratList.length === 0">
            <td colspan="7" class="px-4 py-8 text-center text-zinc-500">
              Sedang mengambil data surat...
            </td>
          </tr>

          <tr v-else-if="suratList.length === 0">
            <td colspan="7" class="px-4 py-8 text-center text-zinc-500">
              Belum ada data surat. Silakan buat surat baru.
            </td>
          </tr>

          <tr v-for="item in suratList" :key="item.id" class="hover:bg-zinc-50/70 transition-colors">
            <td class="px-4 py-3 font-mono text-xs font-semibold text-zinc-900">
              {{ item.nomor_surat }}
            </td>
            <td class="px-4 py-3">
              {{ item.perihal }}
            </td>
            <td class="px-4 py-3 text-zinc-600">
              {{ item.pic_nama }}
            </td>
            <td class="px-4 py-3">
              <span
                :class="getStatusBadgeClass(item.status_saat_ini)"
                class="inline-block px-2.5 py-0.5 text-xs font-medium rounded border"
              >
                {{ item.status_saat_ini }}
              </span>
            </td>
            <td class="px-4 py-3">
              <a
                v-if="item.arsip_url"
                :href="getBackendAssetUrl(item.arsip_url)"
                target="_blank"
                rel="noreferrer"
                class="inline-flex items-center text-xs text-blue-600 hover:text-blue-800 underline font-medium"
              >
                Lihat Arsip
              </a>
              <span v-else class="text-xs text-zinc-400 italic">Belum diunggah</span>
            </td>
            <td class="px-4 py-3 text-xs text-zinc-500">
              {{ getLastLogDate(item.logs) }}
            </td>
            <td class="px-4 py-3 text-right space-x-2">
              <button
                type="button"
                @click="openUpdateModal(item)"
                class="px-2.5 py-1 text-xs font-medium border border-zinc-200 rounded text-zinc-700 bg-white hover:bg-zinc-50"
              >
                Ubah Status
              </button>
              <button
                type="button"
                @click="goToScanner(item.id)"
                class="px-2.5 py-1 text-xs font-medium border border-zinc-200 rounded text-zinc-700 bg-white hover:bg-zinc-50"
              >
                {{ item.arsip_url ? 'Pindai Ulang' : 'Pindai Arsip' }}
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal: Update Status -->
    <div
      v-if="isUpdateModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-xs p-4"
    >
      <div class="bg-white rounded-lg border border-zinc-200 max-w-md w-full p-6 space-y-4 shadow-lg">
        <header>
          <h3 class="text-base font-semibold text-zinc-900">Perbarui Status Surat</h3>
          <p class="text-xs text-zinc-500 mt-1">
            Nomor: <strong class="font-mono text-zinc-800">{{ selectedSurat?.nomor_surat }}</strong>
          </p>
        </header>

        <form @submit.prevent="handleUpdateStatus" class="space-y-4">
          <div>
            <label for="status-select" class="block text-xs font-medium text-zinc-700 mb-1">
              Pilih Status Baru
            </label>
            <select
              id="status-select"
              v-model="newStatus"
              class="w-full px-3 py-2 text-sm border border-zinc-300 rounded-md bg-white focus:outline-none focus:ring-2 focus:ring-zinc-400"
              required
            >
              <option v-for="opt in statusOptions" :key="opt" :value="opt">
                {{ opt }}
              </option>
            </select>
          </div>

          <!-- History preview -->
          <div v-if="selectedSurat?.logs && selectedSurat.logs.length > 0" class="border-t border-zinc-200 pt-3">
            <h4 class="text-xs font-semibold text-zinc-600 mb-2">Riwayat Log Tracking:</h4>
            <ul class="space-y-1.5 max-h-36 overflow-y-auto text-xs text-zinc-600">
              <li
                v-for="log in selectedSurat.logs"
                :key="log.id"
                class="flex items-center justify-between bg-zinc-50 px-2 py-1 rounded border border-zinc-100"
              >
                <span>{{ log.status_baru }}</span>
                <time class="text-zinc-400 text-[11px]">{{ formatDate(log.waktu_update) }}</time>
              </li>
            </ul>
          </div>

          <footer class="flex justify-end gap-2 pt-2 border-t border-zinc-200">
            <button
              type="button"
              @click="closeUpdateModal"
              class="px-3 py-1.5 text-xs font-medium border border-zinc-300 rounded-md text-zinc-700 hover:bg-zinc-50"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isUpdating"
              class="px-3.5 py-1.5 text-xs font-medium rounded-md bg-zinc-900 text-white hover:bg-zinc-800 disabled:opacity-50"
            >
              {{ isUpdating ? 'Menyimpan...' : 'Simpan Pembaruan' }}
            </button>
          </footer>
        </form>
      </div>
    </div>

    <!-- Modal: Create Surat -->
    <div
      v-if="isCreateModalOpen"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-xs p-4"
    >
      <div class="bg-white rounded-lg border border-zinc-200 max-w-lg w-full p-6 space-y-4 shadow-lg">
        <header>
          <h3 class="text-base font-semibold text-zinc-900">Tambah Surat Baru</h3>
          <p class="text-xs text-zinc-500 mt-1">Masukkan informasi surat untuk memulai pelacakan.</p>
        </header>

        <form @submit.prevent="handleCreateSurat" class="space-y-3.5">
          <div>
            <label for="create-nomor" class="block text-xs font-medium text-zinc-700 mb-1">
              Nomor Surat
            </label>
            <input
              id="create-nomor"
              v-model="createForm.nomor_surat"
              type="text"
              placeholder="Contoh: 001/DOSCOM/X/2026"
              class="w-full px-3 py-2 text-sm border border-zinc-300 rounded-md focus:outline-none focus:ring-2 focus:ring-zinc-400"
              required
            />
          </div>

          <div>
            <label for="create-perihal" class="block text-xs font-medium text-zinc-700 mb-1">
              Perihal
            </label>
            <input
              id="create-perihal"
              v-model="createForm.perihal"
              type="text"
              placeholder="Contoh: Permohonan Peminjaman Ruangan"
              class="w-full px-3 py-2 text-sm border border-zinc-300 rounded-md focus:outline-none focus:ring-2 focus:ring-zinc-400"
              required
            />
          </div>

          <div>
            <label for="create-pic" class="block text-xs font-medium text-zinc-700 mb-1">
              Nama PIC (Penanggung Jawab)
            </label>
            <input
              id="create-pic"
              v-model="createForm.pic_nama"
              type="text"
              placeholder="Contoh: Fikri"
              class="w-full px-3 py-2 text-sm border border-zinc-300 rounded-md focus:outline-none focus:ring-2 focus:ring-zinc-400"
              required
            />
          </div>

          <div>
            <label for="create-status" class="block text-xs font-medium text-zinc-700 mb-1">
              Status Awal
            </label>
            <select
              id="create-status"
              v-model="createForm.status_saat_ini"
              class="w-full px-3 py-2 text-sm border border-zinc-300 rounded-md bg-white focus:outline-none focus:ring-2 focus:ring-zinc-400"
              required
            >
              <option v-for="opt in statusOptions" :key="opt" :value="opt">
                {{ opt }}
              </option>
            </select>
          </div>

          <footer class="flex justify-end gap-2 pt-3 border-t border-zinc-200">
            <button
              type="button"
              @click="closeCreateModal"
              class="px-3 py-1.5 text-xs font-medium border border-zinc-300 rounded-md text-zinc-700 hover:bg-zinc-50"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isCreating"
              class="px-3.5 py-1.5 text-xs font-medium rounded-md bg-zinc-900 text-white hover:bg-zinc-800 disabled:opacity-50"
            >
              {{ isCreating ? 'Menyimpan...' : 'Simpan Surat' }}
            </button>
          </footer>
        </form>
      </div>
    </div>
  </section>
</template>
