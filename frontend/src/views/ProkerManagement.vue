<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { prokerService, instansiService, type Proker, type Instansi } from '@/services/api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
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
import { ConfirmDialog } from '@/components/common'

// Active Tab ('proker' | 'instansi')
const activeTab = ref<'proker' | 'instansi'>('proker')

// Proker State
const prokerList = ref<Proker[]>([])
const isLoadingProker = ref(false)
const prokerErrorMessage = ref('')

const isProkerModalOpen = ref(false)
const isEditingProker = ref(false)
const editingProkerId = ref('')
const isSubmittingProker = ref(false)
const formProkerError = ref('')
const formProker = ref({
  nama_proker: '',
  deskripsi: '',
})

// Unified Delete State
const itemToDelete = ref<{ type: 'proker' | 'instansi'; id: string; name: string } | null>(null)
const isDeleting = ref(false)

// Instansi State
const instansiList = ref<Instansi[]>([])
const isLoadingInstansi = ref(false)
const instansiErrorMessage = ref('')

const isInstansiModalOpen = ref(false)
const isEditingInstansi = ref(false)
const editingInstansiId = ref('')
const isSubmittingInstansi = ref(false)
const formInstansiError = ref('')
const formInstansi = ref({
  nama_instansi: '',
  kategori: 'Internal Kampus',
  keterangan: '',
})

const KATEGORI_OPTIONS = [
  'Fakultas',
  'Universitas',
  'Rektorat',
  'Akademik',
  'Unit Kampus',
  'Sekolah Mitra',
  'Eksternal',
  'Ormawa / Mitra',
  'Internal Kampus',
] as const

// Fetchers
const fetchProker = async () => {
  isLoadingProker.value = true
  prokerErrorMessage.value = ''
  try {
    prokerList.value = await prokerService.getAll()
  } catch (err: any) {
    prokerErrorMessage.value = err.response?.data?.error || err.message || 'Gagal memuat data program kerja'
    toast({
      title: 'Gagal Memuat Data',
      description: prokerErrorMessage.value,
      variant: 'destructive',
    })
  } finally {
    isLoadingProker.value = false
  }
}

const fetchInstansi = async () => {
  isLoadingInstansi.value = true
  instansiErrorMessage.value = ''
  try {
    instansiList.value = await instansiService.getAll()
  } catch (err: any) {
    instansiErrorMessage.value = err.response?.data?.error || err.message || 'Gagal memuat data instansi'
    toast({
      title: 'Gagal Memuat Data',
      description: instansiErrorMessage.value,
      variant: 'destructive',
    })
  } finally {
    isLoadingInstansi.value = false
  }
}

const refreshCurrentTab = () => {
  if (activeTab.value === 'proker') {
    fetchProker()
  } else {
    fetchInstansi()
  }
}

// Proker Handlers
const openProkerModal = (proker?: Proker) => {
  isEditingProker.value = !!proker
  editingProkerId.value = proker?.id || ''
  formProker.value = {
    nama_proker: proker?.nama_proker || '',
    deskripsi: proker?.deskripsi || '',
  }
  formProkerError.value = ''
  isProkerModalOpen.value = true
}

const handleProkerSubmit = async () => {
  const trimmedNama = formProker.value.nama_proker.trim()
  if (!trimmedNama) {
    formProkerError.value = 'Nama program kerja wajib diisi.'
    toast({
      title: 'Validasi Gagal',
      description: formProkerError.value,
      variant: 'destructive',
    })
    return
  }

  isSubmittingProker.value = true
  formProkerError.value = ''
  try {
    if (isEditingProker.value && editingProkerId.value) {
      await prokerService.update(editingProkerId.value, trimmedNama, formProker.value.deskripsi.trim())
      toast({
        title: 'Berhasil',
        description: 'Program kerja berhasil diperbarui',
        variant: 'success',
      })
    } else {
      await prokerService.create(trimmedNama, formProker.value.deskripsi.trim())
      toast({
        title: 'Berhasil',
        description: 'Program kerja berhasil ditambahkan',
        variant: 'success',
      })
    }
    isProkerModalOpen.value = false
    await fetchProker()
  } catch (err: any) {
    formProkerError.value = err.response?.data?.error || err.message || 'Gagal menyimpan program kerja.'
    toast({
      title: 'Gagal Menyimpan Proker',
      description: formProkerError.value,
      variant: 'destructive',
    })
  } finally {
    isSubmittingProker.value = false
  }
}

const confirmDelete = (type: 'proker' | 'instansi', id: string, name: string) => {
  itemToDelete.value = { type, id, name }
}

const executeDelete = async () => {
  if (!itemToDelete.value) return
  isDeleting.value = true
  const { type, id, name } = itemToDelete.value
  try {
    if (type === 'proker') {
      await prokerService.delete(id)
      await fetchProker()
    } else {
      await instansiService.delete(id)
      await fetchInstansi()
    }
    toast({
      title: 'Berhasil',
      description: `"${name}" berhasil dihapus`,
      variant: 'success',
    })
    itemToDelete.value = null
  } catch (err: any) {
    toast({
      title: 'Gagal Menghapus',
      description: err.response?.data?.error || err.message || 'Gagal menghapus data.',
      variant: 'destructive',
    })
  } finally {
    isDeleting.value = false
  }
}

// Instansi Handlers
const openInstansiModal = (ins?: Instansi) => {
  isEditingInstansi.value = !!ins
  editingInstansiId.value = ins?.id || ''
  formInstansi.value = {
    nama_instansi: ins?.nama_instansi || '',
    kategori: ins?.kategori || 'Fakultas',
    keterangan: ins?.keterangan || '',
  }
  formInstansiError.value = ''
  isInstansiModalOpen.value = true
}

const handleInstansiSubmit = async () => {
  const trimmedNama = formInstansi.value.nama_instansi.trim()
  if (!trimmedNama) {
    formInstansiError.value = 'Nama instansi wajib diisi.'
    toast({
      title: 'Validasi Gagal',
      description: formInstansiError.value,
      variant: 'destructive',
    })
    return
  }

  isSubmittingInstansi.value = true
  formInstansiError.value = ''
  try {
    if (isEditingInstansi.value && editingInstansiId.value) {
      await instansiService.update(
        editingInstansiId.value,
        trimmedNama,
        formInstansi.value.kategori,
        formInstansi.value.keterangan.trim(),
      )
      toast({
        title: 'Berhasil',
        description: 'Data instansi berhasil diperbarui',
        variant: 'success',
      })
    } else {
      await instansiService.create(
        trimmedNama,
        formInstansi.value.kategori,
        formInstansi.value.keterangan.trim(),
      )
      toast({
        title: 'Berhasil',
        description: 'Instansi baru berhasil ditambahkan',
        variant: 'success',
      })
    }
    isInstansiModalOpen.value = false
    await fetchInstansi()
  } catch (err: any) {
    formInstansiError.value = err.response?.data?.error || err.message || 'Gagal menyimpan data instansi.'
    toast({
      title: 'Gagal Menyimpan Instansi',
      description: formInstansiError.value,
      variant: 'destructive',
    })
  } finally {
    isSubmittingInstansi.value = false
  }
}

onMounted(() => {
  fetchProker()
  fetchInstansi()
})
</script>

<template>
  <div class="w-full max-w-5xl mx-auto space-y-6">
    <!-- Top Bar: Title & Pill Tab Switcher -->
    <header class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-2 px-1">
      <div class="space-y-1.5">
        <!-- Tab Pill Switcher (Floema Minimalist) -->
        <div class="inline-flex p-1 bg-slate-100/90 border border-slate-200/70 rounded-full text-xs shadow-2xs">
          <button
            type="button"
            @click="activeTab = 'proker'"
            :class="activeTab === 'proker' ? 'bg-white text-slate-900 font-semibold shadow-2xs' : 'text-slate-500 hover:text-slate-800'"
            class="px-4 py-1.5 rounded-full transition-all duration-180 flex items-center gap-2 cursor-pointer"
          >
            <span>Program Kerja</span>
            <span
              :class="activeTab === 'proker' ? 'bg-sky-50 text-sky-700' : 'bg-slate-200/80 text-slate-600'"
              class="text-[10px] px-1.5 py-0.2 rounded-full font-mono font-medium"
            >
              {{ prokerList.length }}
            </span>
          </button>

          <button
            type="button"
            @click="activeTab = 'instansi'"
            :class="activeTab === 'instansi' ? 'bg-white text-slate-900 font-semibold shadow-2xs' : 'text-slate-500 hover:text-slate-800'"
            class="px-4 py-1.5 rounded-full transition-all duration-180 flex items-center gap-2 cursor-pointer"
          >
            <span>Instansi &amp; Mitra</span>
            <span
              :class="activeTab === 'instansi' ? 'bg-sky-50 text-sky-700' : 'bg-slate-200/80 text-slate-600'"
              class="text-[10px] px-1.5 py-0.2 rounded-full font-mono font-medium"
            >
              {{ instansiList.length }}
            </span>
          </button>
        </div>

        <div>
          <h2 class="text-xl font-medium tracking-tight text-slate-900 leading-snug">
            {{ activeTab === 'proker' ? 'Program Kerja' : 'Instansi & Mitra Tujuan' }}
          </h2>
          <p class="text-xs text-slate-500 mt-0.5">
            {{ activeTab === 'proker'
              ? 'Kelola data program kerja DOSCOM untuk penomoran dan klasifikasi surat resmi.'
              : 'Daftar pejabat birokrasi kampus dan mitra eksternal untuk auto-fill tujuan surat.'
            }}
          </p>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center gap-2 w-full sm:w-auto">
        <Button
          type="button"
          variant="outline"
          size="icon"
          @click="refreshCurrentTab"
          :disabled="activeTab === 'proker' ? isLoadingProker : isLoadingInstansi"
          title="Muat Ulang"
          aria-label="Muat Ulang"
          class="h-9 w-9 shrink-0 rounded-xl border-slate-200/90 bg-white shadow-2xs text-slate-600 hover:text-slate-900 active:scale-95 transition-all"
        >
          <svg
            class="w-4 h-4"
            :class="{ 'animate-spin': activeTab === 'proker' ? isLoadingProker : isLoadingInstansi }"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </Button>

        <Button
          type="button"
          size="sm"
          @click="activeTab === 'proker' ? openProkerModal() : openInstansiModal()"
          class="flex-1 sm:flex-initial transition-all duration-180"
        >
          {{ activeTab === 'proker' ? '+ Tambah Proker Baru' : '+ Tambah Instansi Baru' }}
        </Button>
      </div>
    </header>

    <!-- Error notices -->
    <div
      v-if="activeTab === 'proker' && prokerErrorMessage"
      role="alert"
      class="p-3 text-xs bg-rose-50 border border-rose-200 text-rose-700 rounded-2xl font-medium"
    >
      {{ prokerErrorMessage }}
    </div>
    <div
      v-else-if="activeTab === 'instansi' && instansiErrorMessage"
      role="alert"
      class="p-3 text-xs bg-rose-50 border border-rose-200 text-rose-700 rounded-2xl font-medium"
    >
      {{ instansiErrorMessage }}
    </div>

    <!-- Tab Content Transition (Floema page-fade) -->
    <Transition name="page-fade" mode="out-in">
      <!-- TAB 1: Tabel Program Kerja -->
      <div
        v-if="activeTab === 'proker'"
        key="proker"
        class="w-full bg-white/95 border border-slate-200/90 rounded-3xl shadow-xs overflow-hidden"
    >
      <div class="w-full overflow-x-auto">
        <table class="w-full text-left text-xs divide-y divide-black/5 min-w-[640px]">
          <thead class="bg-black/[0.02] text-zinc-500 font-semibold">
            <tr>
              <th scope="col" class="px-5 py-4 w-14">No</th>
              <th scope="col" class="px-5 py-4 w-64">Nama Proker</th>
              <th scope="col" class="px-5 py-4">Deskripsi</th>
              <th scope="col" class="px-5 py-4 text-right w-32">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-black/5 text-zinc-800">
            <tr v-if="isLoadingProker && prokerList.length === 0">
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
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    @click="openProkerModal(proker)"
                    class="h-8 w-8 rounded-xl text-zinc-400 hover:text-[#0A84DC] hover:bg-[#0A84DC]/10"
                    title="Edit Program Kerja"
                    aria-label="Edit Program Kerja"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </Button>

                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    @click="confirmDelete('proker', proker.id, proker.nama_proker)"
                    class="h-8 w-8 rounded-xl text-zinc-400 hover:text-rose-600 hover:bg-rose-50"
                    title="Hapus Program Kerja"
                    aria-label="Hapus Program Kerja"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </Button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

      <!-- TAB 2: Tabel Instansi & Mitra -->
      <div
        v-else
        key="instansi"
        class="w-full bg-white/95 border border-slate-200/90 rounded-3xl shadow-xs overflow-hidden"
    >
      <div class="w-full overflow-x-auto">
        <table class="w-full text-left text-xs divide-y divide-black/5 min-w-[640px]">
          <thead class="bg-black/[0.02] text-zinc-500 font-semibold">
            <tr>
              <th scope="col" class="px-5 py-4 w-14">No</th>
              <th scope="col" class="px-5 py-4 w-64">Nama Instansi / Pejabat</th>
              <th scope="col" class="px-5 py-4 w-36">Kategori</th>
              <th scope="col" class="px-5 py-4">Keterangan</th>
              <th scope="col" class="px-5 py-4 text-right w-32">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-black/5 text-zinc-800">
            <tr v-if="isLoadingInstansi && instansiList.length === 0">
              <td colspan="5" class="px-5 py-12 text-center text-zinc-400">
                Memuat data instansi...
              </td>
            </tr>

            <tr v-else-if="instansiList.length === 0">
              <td colspan="5" class="px-5 py-12 text-center text-zinc-400">
                Belum ada data instansi. Klik "+ Tambah Instansi Baru" untuk menambahkan.
              </td>
            </tr>

            <tr
              v-for="(ins, index) in instansiList"
              :key="ins.id"
              class="hover:bg-black/[0.02] transition-colors"
            >
              <td class="px-5 py-4 font-mono text-zinc-400">
                {{ index + 1 }}
              </td>
              <td class="px-5 py-4 font-semibold text-zinc-900">
                {{ ins.nama_instansi }}
              </td>
              <td class="px-5 py-4">
                <Badge variant="outline" size="xs" class="font-medium bg-slate-50 border-slate-200 text-slate-700">
                  {{ ins.kategori }}
                </Badge>
              </td>
              <td class="px-5 py-4 text-zinc-600 leading-relaxed">
                {{ ins.keterangan || '-' }}
              </td>
              <td class="px-5 py-3 text-right">
                <div class="inline-flex items-center justify-end gap-1">
                  <Button
                    type="button"
                    variant="ghost"
                  size="icon"
                  @click="openInstansiModal(ins)"
                  class="h-8 w-8 rounded-xl text-zinc-400 hover:text-[#0A84DC] hover:bg-[#0A84DC]/10"
                  title="Edit Instansi"
                  aria-label="Edit Instansi"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </Button>

                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  @click="confirmDelete('instansi', ins.id, ins.nama_instansi)"
                  class="h-8 w-8 rounded-xl text-zinc-400 hover:text-rose-600 hover:bg-rose-50"
                  title="Hapus Instansi"
                  aria-label="Hapus Instansi"
                >
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </Button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</Transition>

    <!-- Modal Form (Tambah / Edit Proker) -->
    <Dialog :open="isProkerModalOpen" @update:open="(val: boolean) => { if (!val) isProkerModalOpen = false }">
      <DialogContent class="max-w-md space-y-4">
        <DialogHeader>
          <DialogTitle>
            {{ isEditingProker ? 'Edit Program Kerja' : 'Tambah Program Kerja' }}
          </DialogTitle>
          <DialogDescription>Masukkan nama program kerja dan keterangan fungsinya.</DialogDescription>
        </DialogHeader>

        <form @submit.prevent="handleProkerSubmit" novalidate class="space-y-4">
          <div class="space-y-3">
            <div class="space-y-1.5">
              <Label for="nama_proker">Nama Proker</Label>
              <Input
                id="nama_proker"
                v-model="formProker.nama_proker"
                type="text"
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

          <p v-if="formProkerError" class="text-xs text-rose-600 bg-rose-50 p-3 rounded-2xl border border-rose-200 font-medium">
            {{ formProkerError }}
          </p>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              size="sm"
              @click="isProkerModalOpen = false"
            >
              Batal
            </Button>
            <Button
              type="submit"
              size="sm"
              :disabled="isSubmittingProker"
            >
              {{ isSubmittingProker ? 'Menyimpan...' : (isEditingProker ? 'Perbarui Proker' : 'Simpan Proker') }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Modal Form (Tambah / Edit Instansi) -->
    <Dialog :open="isInstansiModalOpen" @update:open="(val: boolean) => { if (!val) isInstansiModalOpen = false }">
      <DialogContent class="max-w-md space-y-4">
        <DialogHeader>
          <DialogTitle>
            {{ isEditingInstansi ? 'Edit Instansi &amp; Mitra' : 'Tambah Instansi Baru' }}
          </DialogTitle>
          <DialogDescription>Daftarkan instansi atau pejabat kampus yang menjadi tujuan surat.</DialogDescription>
        </DialogHeader>

        <form @submit.prevent="handleInstansiSubmit" novalidate class="space-y-4">
          <div class="space-y-3">
            <div class="space-y-1.5">
              <Label for="nama_instansi">Nama Instansi / Pejabat</Label>
              <Input
                id="nama_instansi"
                v-model="formInstansi.nama_instansi"
                type="text"
                placeholder="Contoh: Tata Usaha Fakultas Teknologi Cerdas"
              />
            </div>

            <div class="space-y-1.5">
              <Label for="kategori_instansi">Kategori</Label>
              <Select v-model="formInstansi.kategori">
                <SelectTrigger id="kategori_instansi">
                  <SelectValue :placeholder="formInstansi.kategori" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="kat in KATEGORI_OPTIONS" :key="kat" :value="kat">
                    {{ kat }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div class="space-y-1.5">
              <Label for="keterangan_instansi">Keterangan / Fungsi Surat</Label>
              <Textarea
                id="keterangan_instansi"
                v-model="formInstansi.keterangan"
                rows="2"
                placeholder="Contoh: Peminjaman ruangan kampus & sarana prasarana..."
              />
            </div>
          </div>

          <p v-if="formInstansiError" class="text-xs text-rose-600 bg-rose-50 p-3 rounded-2xl border border-rose-200 font-medium">
            {{ formInstansiError }}
          </p>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              size="sm"
              @click="isInstansiModalOpen = false"
            >
              Batal
            </Button>
            <Button
              type="submit"
              size="sm"
              :disabled="isSubmittingInstansi"
            >
              {{ isSubmittingInstansi ? 'Menyimpan...' : (isEditingInstansi ? 'Perbarui Instansi' : 'Simpan Instansi') }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Modal Konfirmasi Hapus (Unified) -->
    <ConfirmDialog
      :open="!!itemToDelete"
      :title="itemToDelete?.type === 'proker' ? 'Hapus Program Kerja?' : 'Hapus Instansi / Mitra?'"
      confirm-text="Ya, Hapus"
      loading-text="Menghapus..."
      :is-loading="isDeleting"
      @update:open="(val: boolean) => { if (!val) itemToDelete = null }"
      @cancel="itemToDelete = null"
      @confirm="executeDelete"
    >
      <template #description>
        Yakin ingin menghapus <b class="text-zinc-900">{{ itemToDelete?.name }}</b>? Tindakan ini tidak dapat dibatalkan.
      </template>
    </ConfirmDialog>
  </div>
</template>
