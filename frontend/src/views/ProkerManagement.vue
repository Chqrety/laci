<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { prokerService, type Proker } from '@/services/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const prokerList = ref<Proker[]>([])
const isLoading = ref(false)
const errorMessage = ref('')

// Modal state (Create & Update)
const isModalOpen = ref(false)
const isEditing = ref(false)
const editingId = ref('')
const isSubmitting = ref(false)
const formError = ref('')
const formProker = ref({
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

const openModal = (proker?: Proker) => {
  isEditing.value = !!proker
  editingId.value = proker?.id || ''
  formProker.value = {
    nama_proker: proker?.nama_proker || '',
    deskripsi: proker?.deskripsi || '',
  }
  formError.value = ''
  isModalOpen.value = true
}

const handleSubmit = async () => {
  const trimmedNama = formProker.value.nama_proker.trim()
  if (!trimmedNama) {
    formError.value = 'Nama program kerja wajib diisi.'
    return
  }

  isSubmitting.value = true
  formError.value = ''
  try {
    if (isEditing.value && editingId.value) {
      await prokerService.update(editingId.value, trimmedNama, formProker.value.deskripsi.trim())
    } else {
      await prokerService.create(trimmedNama, formProker.value.deskripsi.trim())
    }
    isModalOpen.value = false
    await fetchProker()
  } catch (err: any) {
    formError.value = err.response?.data?.error || err.message || 'Gagal menyimpan program kerja.'
  } finally {
    isSubmitting.value = false
  }
}

const handleDeleteProker = async (proker: Proker) => {
  if (!window.confirm(`Hapus program kerja "${proker.nama_proker}"?`)) {
    return
  }

  try {
    await prokerService.delete(proker.id)
    await fetchProker()
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal menghapus program kerja.')
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
        <h2 class="text-xl font-medium tracking-tight text-slate-900 leading-snug">Program Kerja</h2>
        <p class="text-xs text-slate-500 mt-0.5">Kelola data program kerja DOSCOM untuk klasifikasi dan penomoran surat resmi.</p>
      </div>

      <div class="flex items-center gap-2">
        <Button
          type="button"
          variant="outline"
          size="sm"
          @click="fetchProker"
          :disabled="isLoading"
        >
          {{ isLoading ? 'Memuat...' : 'Muat Ulang' }}
        </Button>

        <Button
          type="button"
          size="sm"
          @click="openModal()"
        >
          + Tambah Proker Baru
        </Button>
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
    <div class="bg-white/95 border border-slate-200/90 rounded-3xl shadow-xs overflow-hidden">
      <table class="w-full text-left text-xs divide-y divide-black/5">
        <thead class="bg-black/[0.02] text-zinc-500 font-semibold">
          <tr>
            <th scope="col" class="px-5 py-4 w-14">No</th>
            <th scope="col" class="px-5 py-4 w-64">Nama Proker</th>
            <th scope="col" class="px-5 py-4">Deskripsi</th>
            <th scope="col" class="px-5 py-4 text-right w-36">Aksi</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-black/5 text-zinc-800">
          <tr v-if="isLoading && prokerList.length === 0">
            <td colspan="4" class="px-5 py-12 text-center text-zinc-400">
              Memuat data program kerja...
            </td>
          </tr>

          <tr v-else-if="prokerList.length === 0">
            <td colspan="4" class="px-5 py-12 text-center text-zinc-400">
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
            <td class="px-5 py-3 text-right">
              <div class="inline-flex items-center justify-end gap-1">
                <!-- Tombol Edit Proker (Pencil Icon) -->
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  @click="openModal(proker)"
                  class="h-8 w-8 rounded-xl text-zinc-400 hover:text-[#0A84DC] hover:bg-[#0A84DC]/10"
                  title="Edit Program Kerja"
                  aria-label="Edit Program Kerja"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="1.8"
                      d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                    />
                  </svg>
                </Button>

                <!-- Tombol Hapus Proker (Trash Icon) -->
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  @click="handleDeleteProker(proker)"
                  class="h-8 w-8 rounded-xl text-zinc-400 hover:text-rose-600 hover:bg-rose-50"
                  title="Hapus Program Kerja"
                  aria-label="Hapus Program Kerja"
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

    <!-- Modal Form (Tambah / Edit Proker) -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isModalOpen"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 overflow-y-auto"
      >
        <div class="bg-white rounded-3xl border border-black/10 max-w-md w-full p-6 sm:p-7 space-y-5 shadow-xl">
          <header>
            <h3 class="text-lg font-semibold text-zinc-900 tracking-tight">
              {{ isEditing ? 'Edit Program Kerja' : 'Tambah Program Kerja' }}
            </h3>
            <p class="text-xs text-zinc-500 mt-0.5">Masukkan nama program kerja dan keterangan fungsinya.</p>
          </header>

          <form @submit.prevent="handleSubmit" class="space-y-4">
            <div class="space-y-3">
              <div class="space-y-1.5">
                <Label for="nama_proker">Nama Proker</Label>
                <Input
                  id="nama_proker"
                  v-model="formProker.nama_proker"
                  type="text"
                  required
                  placeholder="Dies Natalis, Oprec, Workshop"
                />
              </div>

              <div class="space-y-1.5">
                <Label for="deskripsi_proker">Deskripsi (Opsional)</Label>
                <Textarea
                  id="deskripsi_proker"
                  v-model="formProker.deskripsi"
                  rows="3"
                  placeholder="Keterangan singkat program kerja..."
                />
              </div>
            </div>

            <p v-if="formError" class="text-xs text-rose-600 bg-rose-50 p-3 rounded-2xl border border-rose-200 font-medium">
              {{ formError }}
            </p>

            <footer class="flex justify-end gap-2.5 pt-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                @click="isModalOpen = false"
              >
                Batal
              </Button>
              <Button
                type="submit"
                size="sm"
                :disabled="isSubmitting"
              >
                {{ isSubmitting ? 'Menyimpan...' : (isEditing ? 'Perbarui Proker' : 'Simpan Proker') }}
              </Button>
            </footer>
          </form>
        </div>
      </div>
    </Transition>
  </div>
</template>
