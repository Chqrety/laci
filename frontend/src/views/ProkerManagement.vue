<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { prokerService, type Proker } from '@/services/api'

const prokerList = ref<Proker[]>([])
const isLoading = ref(false)
const errorMessage = ref('')

// Form state
const isAddModalOpen = ref(false)
const isSubmitting = ref(false)
const formError = ref('')
const newProker = ref({
  nama_proker: '',
  deskripsi: '',
})

const fetchProker = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    prokerList.value = await prokerService.getAll()
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || err.message || 'Gagal memuat data program kerja'
  } finally {
    isLoading.value = false
  }
}

const openAddModal = () => {
  newProker.value = {
    nama_proker: '',
    deskripsi: '',
  }
  formError.value = ''
  isAddModalOpen.value = true
}

const handleCreateProker = async () => {
  if (!newProker.value.nama_proker.trim()) {
    formError.value = 'Nama program kerja wajib diisi.'
    return
  }

  isSubmitting.value = true
  formError.value = ''
  try {
    await prokerService.create(newProker.value.nama_proker.trim(), newProker.value.deskripsi.trim())
    isAddModalOpen.value = false
    await fetchProker()
  } catch (err: any) {
    formError.value = err.response?.data?.error || err.message || 'Gagal menyimpan program kerja.'
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  fetchProker()
})
</script>

<template>
  <div class="max-w-5xl mx-auto space-y-6">
    <!-- Top Bar -->
    <header class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-2 px-1">
      <div>
        <h2 class="text-xl font-semibold tracking-tight text-zinc-900">Manajemen Program Kerja</h2>
        <p class="text-xs text-zinc-500 mt-0.5">Kelola data program kerja DOSCOM untuk klasifikasi dan penomoran surat resmi.</p>
      </div>

      <div class="flex items-center gap-2">
        <button
          type="button"
          @click="fetchProker"
          :disabled="isLoading"
          class="px-4 py-2 text-xs font-semibold border border-black/5 rounded-2xl bg-white/80 text-zinc-700 hover:bg-white shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all active:scale-[0.98] disabled:opacity-50"
        >
          {{ isLoading ? 'Memuat...' : 'Muat Ulang' }}
        </button>

        <button
          type="button"
          @click="openAddModal"
          class="px-4.5 py-2 text-xs font-semibold rounded-2xl bg-[#0A84DC] text-white hover:bg-[#0872be] shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all active:scale-[0.98]"
        >
          + Tambah Proker Baru
        </button>
      </div>
    </header>

    <!-- Error notice -->
    <div
      v-if="errorMessage"
      role="alert"
      class="p-3 text-xs bg-rose-50 border border-rose-200 text-rose-700 rounded-2xl font-medium"
    >
      {{ errorMessage }}
    </div>

    <!-- Table Container (Apple iOS HIG: rounded-2xl, soft shadow) -->
    <div class="bg-white/85 backdrop-blur-xl border border-black/5 rounded-2xl shadow-[0_8px_30px_rgb(0,0,0,0.05)] overflow-hidden">
      <table class="w-full text-left text-xs divide-y divide-black/5">
        <thead class="bg-black/[0.02] text-zinc-500 font-semibold">
          <tr>
            <th scope="col" class="px-5 py-4 w-16">No</th>
            <th scope="col" class="px-5 py-4 w-72">Nama Proker</th>
            <th scope="col" class="px-5 py-4">Deskripsi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-black/5 text-zinc-800">
          <tr v-if="isLoading && prokerList.length === 0">
            <td colspan="3" class="px-5 py-12 text-center text-zinc-400">
              Memuat data program kerja...
            </td>
          </tr>

          <tr v-else-if="prokerList.length === 0">
            <td colspan="3" class="px-5 py-12 text-center text-zinc-400">
              Belum ada data program kerja. Klik "+ Tambah Proker Baru" untuk menambahkan.
            </td>
          </tr>

          <tr
            v-for="(proker, index) in prokerList"
            :key="proker.id"
            class="hover:bg-black/[0.02] transition-colors"
          >
            <td class="px-5 py-4 font-mono text-zinc-400">
              {{ index + 1 }}
            </td>
            <td class="px-5 py-4 font-semibold text-zinc-900">
              {{ proker.nama_proker }}
            </td>
            <td class="px-5 py-4 text-zinc-600 leading-relaxed">
              {{ proker.deskripsi || '-' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal Tambah Proker Baru (iOS HIG: rounded-3xl, Inset Grouped, Glassmorphism, Animated) -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isAddModalOpen"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 overflow-y-auto"
      >
        <div class="bg-white rounded-3xl border border-black/10 max-w-md w-full p-6 sm:p-7 space-y-5 shadow-xl">
          <header>
            <h3 class="text-lg font-semibold text-zinc-900 tracking-tight">Tambah Program Kerja</h3>
            <p class="text-xs text-zinc-500 mt-0.5">Masukkan nama program kerja dan keterangan fungsinya.</p>
          </header>

          <form @submit.prevent="handleCreateProker" class="space-y-4">
            <!-- Inset Grouped Inputs -->
            <div class="rounded-2xl border border-black/10 bg-white overflow-hidden divide-y divide-black/5 shadow-2xs">
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <label for="nama_proker" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Nama Proker</label>
                <input
                  id="nama_proker"
                  v-model="newProker.nama_proker"
                  type="text"
                  required
                  placeholder="Dies Natalis, Oprec, Workshop"
                  class="w-full text-xs font-semibold text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 pt-0.5"
                />
              </div>

              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <label for="deskripsi_proker" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Deskripsi (Opsional)</label>
                <textarea
                  id="deskripsi_proker"
                  v-model="newProker.deskripsi"
                  rows="3"
                  placeholder="Keterangan singkat program kerja..."
                  class="w-full text-xs font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 resize-none pt-0.5"
                ></textarea>
              </div>
            </div>

            <p v-if="formError" class="text-xs text-rose-600 bg-rose-50 p-3 rounded-2xl border border-rose-200 font-medium">
              {{ formError }}
            </p>

            <footer class="flex justify-end gap-2.5 pt-2">
              <button
                type="button"
                @click="isAddModalOpen = false"
                class="px-4 py-2 text-xs font-semibold border border-black/10 rounded-2xl text-zinc-700 hover:bg-black/5 transition-all duration-200 active:scale-95"
              >
                Batal
              </button>
              <button
                type="submit"
                :disabled="isSubmitting"
                class="px-5 py-2 text-xs font-semibold bg-[#0A84DC] text-white rounded-2xl hover:bg-[#0872be] shadow-md disabled:opacity-50 transition-all duration-200 active:scale-95"
              >
                {{ isSubmitting ? 'Menyimpan...' : 'Simpan Proker' }}
              </button>
            </footer>
          </form>
        </div>
      </div>
    </Transition>
  </div>
</template>
