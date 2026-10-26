<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { userService, currentUserId, currentRole, type User } from '@/services/api'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
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

const users = ref<User[]>([])
const isLoading = ref(false)
const searchQuery = ref('')
const selectedDivisi = ref('')
const isImportModalOpen = ref(false)
const isAddModalOpen = ref(false)
const isEditModalOpen = ref(false)
const isDetailModalOpen = ref(false)
const isTransferModalOpen = ref(false)
const selectedUser = ref<User | null>(null)
const importFile = ref<File | null>(null)
const isSubmitting = ref(false)

// Transfer Ketum State
const transferTargetUserId = ref('')
const transferKetumPassword = ref('')
const isTransferring = ref(false)

// Pagination state
const currentPage = ref(1)
const itemsPerPage = ref(15)

// Check current user role
const isKetum = computed(() => currentRole.value === 'ketum')

// Form edit user
const editForm = ref({
  id: '',
  nim: '',
  nama: '',
  email: '',
  divisi: 'humas',
})

// Form tambah user manual
const formUser = ref({
  nim: '',
  nama: '',
  email: '',
  username: '',
  divisi: 'humas',
})

const showToast = (text: string, type: 'success' | 'error' = 'success') => {
  toast({
    title: text,
    variant: type === 'error' ? 'destructive' : 'success',
  })
}

const fetchUsers = async () => {
  isLoading.value = true
  try {
    users.value = await userService.getAll()
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal memuat data pengurus', 'error')
  } finally {
    isLoading.value = false
  }
}

onMounted(() => {
  fetchUsers()
})

watch([searchQuery, selectedDivisi], () => {
  currentPage.value = 1
})

const filteredUsers = computed(() => {
  return users.value.filter((u) => {
    const q = searchQuery.value.toLowerCase()
    const matchesSearch =
      !q ||
      u.nama?.toLowerCase().includes(q) ||
      u.nim?.toLowerCase().includes(q) ||
      u.username?.toLowerCase().includes(q) ||
      u.email?.toLowerCase().includes(q)

    const matchesDivisi = !selectedDivisi.value || selectedDivisi.value === 'all' || u.divisi === selectedDivisi.value
    return matchesSearch && matchesDivisi
  })
})

const totalPages = computed(() => Math.ceil(filteredUsers.value.length / itemsPerPage.value) || 1)

watch(totalPages, (newTotal) => {
  if (currentPage.value > newTotal) currentPage.value = newTotal
})

const paginatedUsers = computed(() => {
  const start = (currentPage.value - 1) * itemsPerPage.value
  return filteredUsers.value.slice(start, start + itemsPerPage.value)
})

const divisiOptions = computed(() => {
  const set = new Set(users.value.map((u) => u.divisi).filter(Boolean))
  return Array.from(set)
})

const assignableDivisiOptions = computed(() => {
  const list = [
    { value: 'humas', label: 'Humas' },
    { value: 'pengurus', label: 'Pengurus' },
    { value: 'sekre', label: 'Sekretaris' },
    { value: 'tupel', label: 'Ketua Pelaksana' },
  ]
  if (isKetum.value) {
    list.unshift({ value: 'ketum', label: 'Ketua Umum' })
  }
  return list
})

const eligibleSuccessors = computed(() => {
  return users.value.filter((u) => u.id !== currentUserId.value && u.divisi !== 'ketum')
})

const selectedUserIds = ref<string[]>([])

const selectableUsers = computed(() => {
  return filteredUsers.value.filter((u) => u.id !== currentUserId.value)
})

const isAllSelected = computed(() => {
  if (selectableUsers.value.length === 0) return false
  return selectableUsers.value.every((u) => selectedUserIds.value.includes(u.id))
})

const isSomeSelected = computed(() => {
  return selectedUserIds.value.length > 0 && !isAllSelected.value
})

const toggleSelectAll = () => {
  if (isAllSelected.value) {
    selectedUserIds.value = []
  } else {
    selectedUserIds.value = selectableUsers.value.map((u) => u.id)
  }
}

const toggleSelectUser = (id: string) => {
  if (id === currentUserId.value) return
  const index = selectedUserIds.value.indexOf(id)
  if (index > -1) {
    selectedUserIds.value.splice(index, 1)
  } else {
    selectedUserIds.value.push(id)
  }
}

const confirmDeleteModal = ref({
  isOpen: false,
  isBatch: false,
  user: null as User | null,
  count: 0,
})

const openSingleDeleteConfirm = (user: User) => {
  confirmDeleteModal.value = {
    isOpen: true,
    isBatch: false,
    user,
    count: 1,
  }
}

const openBatchDeleteConfirm = () => {
  if (selectedUserIds.value.length === 0) return
  confirmDeleteModal.value = {
    isOpen: true,
    isBatch: true,
    user: null,
    count: selectedUserIds.value.length,
  }
}

const executeConfirmedDelete = async () => {
  isSubmitting.value = true
  try {
    if (confirmDeleteModal.value.isBatch) {
      const count = selectedUserIds.value.length
      await userService.deleteBatch(selectedUserIds.value)
      users.value = users.value.filter((u) => !selectedUserIds.value.includes(u.id))
      showToast(`Berhasil menghapus ${count} pengurus`)
      selectedUserIds.value = []
    } else if (confirmDeleteModal.value.user) {
      const u = confirmDeleteModal.value.user
      await userService.delete(u.id)
      users.value = users.value.filter((item) => item.id !== u.id)
      selectedUserIds.value = selectedUserIds.value.filter((id) => id !== u.id)
      showToast(`Pengurus ${u.nama || u.username} berhasil dihapus`)
    }
    confirmDeleteModal.value.isOpen = false
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal menghapus pengurus', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const handleBatchDelete = () => {
  openBatchDeleteConfirm()
}

const handleDeleteUser = (user: User) => {
  openSingleDeleteConfirm(user)
}

const handleCreateUser = async () => {
  if (!formUser.value.nama?.trim()) {
    showToast('Nama pengurus wajib diisi', 'error')
    return
  }

  isSubmitting.value = true
  try {
    const created = await userService.create(formUser.value)
    users.value.unshift(created)
    isAddModalOpen.value = false
    formUser.value = {
      nim: '',
      nama: '',
      email: '',
      username: '',
      divisi: 'humas',
    }
    showToast(`Pengurus ${created.nama} berhasil ditambahkan. Undangan aktivasi telah dikirim.`)
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal menambah pengurus', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const openDetailModal = (user: User) => {
  selectedUser.value = user
  isDetailModalOpen.value = true
}

const openEditModal = (user: User) => {
  selectedUser.value = user
  editForm.value = {
    id: user.id,
    nim: user.nim || '',
    nama: user.nama || '',
    email: user.email || '',
    divisi: user.divisi || 'humas',
  }
  isEditModalOpen.value = true
}

const handleUpdateUser = async () => {
  if (!editForm.value.nama?.trim()) {
    showToast('Nama pengurus wajib diisi', 'error')
    return
  }

  isSubmitting.value = true
  try {
    const updated = await userService.update(editForm.value.id, {
      nim: editForm.value.nim,
      nama: editForm.value.nama,
      email: editForm.value.email,
      divisi: editForm.value.divisi,
    })

    const idx = users.value.findIndex((u) => u.id === updated.id)
    if (idx !== -1) {
      users.value[idx] = updated
    }
    if (selectedUser.value?.id === updated.id) {
      selectedUser.value = updated
    }
    isEditModalOpen.value = false
    showToast(`Data pengurus ${updated.nama} berhasil diperbarui`)
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal memperbarui pengurus', 'error')
  } finally {
    isSubmitting.value = false
  }
}

const handleTransferKetum = async () => {
  if (!transferTargetUserId.value) {
    showToast('Pilih calon Ketua Umum pengganti', 'error')
    return
  }
  if (!transferKetumPassword.value) {
    showToast('Masukkan kata sandi Ketua Umum untuk verifikasi', 'error')
    return
  }

  isTransferring.value = true
  try {
    const res = await userService.transferKetum(transferTargetUserId.value, transferKetumPassword.value)
    showToast(res.message || 'Serah terima jabatan berhasil dilaksanakan')
    isTransferModalOpen.value = false
    transferTargetUserId.value = ''
    transferKetumPassword.value = ''
    await fetchUsers()
    setTimeout(() => {
      window.location.reload()
    }, 1200)
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal memproses serah terima jabatan', 'error')
  } finally {
    isTransferring.value = false
  }
}

const importMethod = ref<'csv' | 'sheet'>('csv')
const sheetUrl = ref('')

const onFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement
  if (target.files && target.files[0]) {
    importFile.value = target.files[0]
  } else {
    importFile.value = null
  }
}

const handleImportData = async () => {
  if (importMethod.value === 'csv') {
    if (!importFile.value) {
      showToast('Pilih file CSV terlebih dahulu', 'error')
      return
    }

    isSubmitting.value = true
    try {
      const res = await userService.importCSV(importFile.value)
      showToast(res.message)
      if (res.skipped && res.skipped.length > 0) {
        toast({
          title: 'Catatan Duplikasi Import',
          description: `${res.skipped.length} baris dilewati karena duplikat: ${res.skipped.join(', ')}`,
        })
      }
      isImportModalOpen.value = false
      importFile.value = null
      await fetchUsers()
    } catch (err: any) {
      showToast(err.response?.data?.error || 'Gagal mengimpor file CSV', 'error')
    } finally {
      isSubmitting.value = false
    }
  } else {
    if (!sheetUrl.value.trim()) {
      showToast('Masukkan link Google Sheets terlebih dahulu', 'error')
      return
    }

    isSubmitting.value = true
    try {
      const res = await userService.importSpreadsheet(sheetUrl.value.trim())
      showToast(res.message)
      if (res.skipped && res.skipped.length > 0) {
        toast({
          title: 'Catatan Duplikasi Import',
          description: `${res.skipped.length} baris dilewati karena duplikat: ${res.skipped.join(', ')}`,
        })
      }
      isImportModalOpen.value = false
      sheetUrl.value = ''
      await fetchUsers()
    } catch (err: any) {
      showToast(err.response?.data?.error || 'Gagal menarik data dari Google Sheets', 'error')
    } finally {
      isSubmitting.value = false
    }
  }
}
</script>

<template>
  <div class="flex-1 flex flex-col gap-5">
    <!-- Header & Actions (Floema Architectural Responsive Header) -->
    <header class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5 flex-wrap">
          <h2 class="text-xl sm:text-2xl font-medium tracking-tight text-slate-900 leading-snug">Kelola Pengurus &amp; Divisi</h2>
          <Badge variant="outline" class="font-normal text-[11px] px-2.5 py-0.5 rounded-full text-slate-500 bg-slate-100/70 border-slate-200/80 shrink-0">
            {{ filteredUsers.length }} {{ filteredUsers.length === users.length ? 'pengurus' : `dari ${users.length}` }}
          </Badge>
        </div>
        <p class="text-xs text-slate-500 mt-1 leading-relaxed">
          Kelola data delegasi divisi pengurus, undangan aktivasi mandiri, dan suksesi kepengurusan DOSCOM.
        </p>
      </div>

      <!-- Action Buttons: 2-column on mobile, single fluid row on tablet/desktop -->
      <div class="grid grid-cols-2 sm:flex sm:items-center gap-2 sm:gap-2.5 w-full sm:w-auto shrink-0">
        <!-- Tombol Serah Terima Jabatan (Hanya Ketum) -->
        <Button
          v-if="isKetum"
          type="button"
          variant="outline"
          size="sm"
          @click="isTransferModalOpen = true"
          class="rounded-full px-3.5 py-2 text-xs font-semibold border-orange-200/90 bg-orange-50/80 text-orange-800 hover:bg-orange-100/90 active:scale-[0.98] transition-all cursor-pointer shadow-2xs flex items-center justify-center gap-1.5 w-full sm:w-auto"
        >
          <svg class="w-3.5 h-3.5 text-orange-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
          <span class="truncate">Serah Terima</span>
        </Button>

        <!-- Tombol Import Data -->
        <Button
          type="button"
          variant="outline"
          size="sm"
          @click="isImportModalOpen = true"
          class="rounded-full px-3.5 py-2 text-xs font-semibold bg-white border border-slate-200/90 text-slate-700 hover:bg-slate-50 active:scale-[0.98] shadow-2xs transition-all cursor-pointer flex items-center justify-center gap-1.5 w-full sm:w-auto"
          :class="isKetum ? '' : 'col-span-1'"
        >
          <svg class="w-3.5 h-3.5 text-sky-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
          </svg>
          <span class="truncate">Import Data</span>
        </Button>

        <!-- Tombol Tambah Pengurus (Primary CTA) -->
        <Button
          type="button"
          size="sm"
          @click="isAddModalOpen = true"
          class="rounded-full px-4 py-2 text-xs font-bold text-white bg-sky-600 hover:bg-sky-700 active:scale-[0.98] shadow-xs transition-all cursor-pointer flex items-center justify-center gap-1.5 col-span-2 sm:col-span-1 w-full sm:w-auto"
        >
          <svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m8-8H4" />
          </svg>
          <span class="truncate">Tambah Pengurus</span>
        </Button>
      </div>
    </header>

    <!-- Filter Bar: Search, Divisi Filter -->
    <div class="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-2.5">
      <div class="relative flex-1 sm:max-w-xs lg:max-w-sm">
        <svg
          class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400 pointer-events-none"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          stroke-width="2"
        >
          <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35m0 0A7.5 7.5 0 105.65 5.65a7.5 7.5 0 0010.7 10.7z" />
        </svg>
        <Input
          v-model="searchQuery"
          type="search"
          placeholder="Cari nama, NIM, email, username..."
          class="pl-9 pr-8 h-9 rounded-2xl border-slate-200/90 bg-white text-xs shadow-2xs focus-visible:ring-sky-500/20 w-full"
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

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <Select v-model="selectedDivisi">
          <SelectTrigger class="h-9 w-full sm:w-44 lg:w-48 rounded-2xl border-slate-200/90 bg-white text-xs shadow-2xs">
            <SelectValue placeholder="Semua Divisi" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Semua Divisi</SelectItem>
            <SelectItem v-for="d in divisiOptions" :key="d" :value="d">
              {{ d.toUpperCase() }}
            </SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>

    <!-- Bulk Selection Bar -->
    <Transition name="slide-fade">
      <div
        v-if="selectedUserIds.length > 0"
        class="bg-white/85 backdrop-blur-xl border border-black/5 rounded-2xl p-3 px-4 shadow-[0_8px_30px_rgb(0,0,0,0.05)] flex items-center justify-between gap-4 transition-all"
      >
        <div class="flex items-center gap-2.5">
          <span class="px-2 py-0.5 rounded-lg bg-[#0A84DC]/10 text-[#0A84DC] text-xs font-semibold font-mono">
            {{ selectedUserIds.length }}
          </span>
          <span class="text-xs font-medium text-zinc-700">
            Pengurus dipilih
          </span>
        </div>

        <div class="flex items-center gap-2">
          <Button
            type="button"
            variant="ghost"
            size="sm"
            @click="selectedUserIds = []"
          >
            Batal Pilih
          </Button>
          <Button
            type="button"
            variant="destructive"
            size="sm"
            @click="handleBatchDelete"
            :disabled="isSubmitting"
          >
            <svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Hapus Terpilih ({{ selectedUserIds.length }})
          </Button>
        </div>
      </div>
    </Transition>

    <!-- Table Container -->
    <div class="bg-white/95 border border-slate-200/90 rounded-3xl shadow-xs overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs divide-y divide-black/5 min-w-[760px]">
          <thead class="bg-black/[0.02] text-zinc-500 font-semibold">
            <tr>
              <th scope="col" class="py-3 px-4 w-12 text-center">
                <input
                  type="checkbox"
                  :checked="isAllSelected"
                  :indeterminate="isSomeSelected"
                  @change="toggleSelectAll"
                  class="w-4 h-4 rounded text-[#0A84DC] focus:ring-[#0A84DC]/30 border-black/20 cursor-pointer accent-[#0A84DC]"
                  title="Pilih Semua Pengurus"
                />
              </th>
              <th scope="col" class="py-3 px-4">Pengurus</th>
              <th scope="col" class="py-3 px-4">NIM</th>
              <th scope="col" class="py-3 px-4">Divisi</th>
              <th scope="col" class="py-3 px-4">Status Akun</th>
              <th scope="col" class="py-3 px-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-black/5 text-zinc-800">
            <tr v-if="isLoading" class="animate-pulse">
              <td colspan="6" class="py-8 text-center text-zinc-400">Memuat data pengurus...</td>
            </tr>

            <tr v-else-if="filteredUsers.length === 0">
              <td colspan="6" class="py-8 text-center text-zinc-400">Tidak ada data pengurus ditemukan.</td>
            </tr>

            <tr
              v-for="u in paginatedUsers"
              :key="u.id"
              class="hover:bg-zinc-50/80 transition-colors"
              :class="selectedUserIds.includes(u.id) ? 'bg-[#0A84DC]/[0.03]' : ''"
            >
              <td class="py-3.5 px-4 text-center">
                <input
                  v-if="u.id !== currentUserId"
                  type="checkbox"
                  :checked="selectedUserIds.includes(u.id)"
                  @change="toggleSelectUser(u.id)"
                  class="w-4 h-4 rounded text-[#0A84DC] focus:ring-[#0A84DC]/30 border-zinc-300 cursor-pointer accent-[#0A84DC]"
                />
                <span v-else class="text-[10px] font-mono text-zinc-300" title="Akun Anda">&mdash;</span>
              </td>
              <td class="py-3.5 px-4">
                <div class="font-semibold text-zinc-900 flex items-center gap-1.5">
                  <span>{{ u.nama || u.username }}</span>
                  <span
                    v-if="u.id === currentUserId"
                    class="px-1.5 py-0.5 text-[9px] font-semibold rounded-md bg-[#0A84DC]/10 text-[#0A84DC]"
                  >
                    Anda
                  </span>
                </div>
                <div class="text-[11px] text-zinc-400">@{{ u.username }} &bull; {{ u.email || 'Tanpa email' }}</div>
              </td>
              <td class="py-3.5 px-4 font-mono text-zinc-600">
                {{ u.nim || '-' }}
              </td>
              <td class="py-3.5 px-4">
                <span
                  class="px-2.5 py-1 rounded-xl text-[10px] font-mono font-semibold uppercase tracking-wider border"
                  :class="u.divisi === 'ketum' ? 'bg-orange-50 text-orange-700 border-orange-200' : 'bg-zinc-100 text-zinc-700 border-zinc-200'"
                >
                  {{ u.divisi }}
                </span>
              </td>
              <td class="py-3.5 px-4">
                <Badge
                  :variant="u.is_active ? 'emerald' : 'amber'"
                  size="xs"
                >
                  {{ u.is_active ? 'Aktif' : 'Menunggu Aktivasi' }}
                </Badge>
              </td>
              <td class="py-3.5 px-4 text-right">
                <div class="flex items-center justify-end gap-1">
                  <!-- Tombol Detail Pengurus -->
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    @click="openDetailModal(u)"
                    class="rounded-xl text-zinc-400 hover:text-zinc-800"
                    title="Lihat Detail Pengurus"
                    aria-label="Lihat Detail Pengurus"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.8"
                        d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                  </Button>

                  <!-- Tombol Edit Pengurus -->
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    @click="openEditModal(u)"
                    class="rounded-xl text-zinc-400 hover:text-amber-600 hover:bg-amber-50"
                    title="Edit Data Pengurus"
                    aria-label="Edit Data Pengurus"
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

                  <!-- Tombol Hapus Pengurus -->
                  <Button
                    v-if="u.id !== currentUserId && (isKetum || u.divisi !== 'ketum')"
                    type="button"
                    variant="ghost"
                    size="icon"
                    @click="handleDeleteUser(u)"
                    class="rounded-xl text-zinc-400 hover:text-rose-600 hover:bg-rose-50"
                    title="Hapus Pengurus"
                    aria-label="Hapus Pengurus"
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
      <div
        v-if="filteredUsers.length > itemsPerPage"
        class="px-5 py-3 border-t border-black/5 bg-black/[0.01] flex items-center justify-between text-xs text-zinc-500"
      >
        <span>
          Menampilkan {{ (currentPage - 1) * itemsPerPage + 1 }} - {{ Math.min(currentPage * itemsPerPage, filteredUsers.length) }} dari {{ filteredUsers.length }} pengurus
        </span>
        <div class="flex items-center gap-1.5">
          <Button
            type="button"
            variant="outline"
            size="sm"
            @click="currentPage--"
            :disabled="currentPage === 1"
          >
            Sebelumnya
          </Button>
          <span class="px-2 font-mono text-[11px] font-semibold text-zinc-700">
            {{ currentPage }} / {{ totalPages }}
          </span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            @click="currentPage++"
            :disabled="currentPage === totalPages"
          >
            Selanjutnya
          </Button>
        </div>
      </div>
    </div>

    <!-- Modal Tambah Pengurus -->
    <Dialog :open="isAddModalOpen" @update:open="(val: boolean) => { if (!val) isAddModalOpen = false }">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>Tambah Pengurus Baru</DialogTitle>
          <DialogDescription>Pengurus baru akan menerima email aktivasi untuk membuat kata sandi mandiri.</DialogDescription>
        </DialogHeader>

        <form @submit.prevent="handleCreateUser" novalidate class="flex flex-col gap-3">
          <div class="space-y-1">
            <Label for="add-user-nama">Nama Lengkap *</Label>
            <Input
              id="add-user-nama"
              v-model="formUser.nama"
              required
              placeholder="Contoh: Budi Santoso"
            />
          </div>

          <div class="space-y-1">
            <Label for="add-user-nim">NIM (Nomor Induk Mahasiswa)</Label>
            <Input
              id="add-user-nim"
              v-model="formUser.nim"
              placeholder="Contoh: A11.2023.12345"
            />
          </div>

          <div class="space-y-1">
            <Label for="add-user-email">Email *</Label>
            <Input
              id="add-user-email"
              v-model="formUser.email"
              type="email"
              required
              placeholder="budi@doscom.org"
            />
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div class="space-y-1">
              <Label for="add-user-username">Username (Opsional)</Label>
              <Input
                id="add-user-username"
                v-model="formUser.username"
                placeholder="Auto dari NIM/Nama"
              />
            </div>
            <div class="space-y-1">
              <Label for="add-user-divisi">Divisi</Label>
              <Select v-model="formUser.divisi">
                <SelectTrigger id="add-user-divisi">
                  <SelectValue placeholder="Pilih Divisi" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="d in assignableDivisiOptions"
                    :key="d.value"
                    :value="d.value"
                  >
                    {{ d.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter class="pt-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              @click="isAddModalOpen = false"
            >
              Batal
            </Button>
            <Button
              type="submit"
              size="sm"
              :disabled="isSubmitting"
            >
              {{ isSubmitting ? 'Menyimpan...' : 'Kirim Undangan Aktivasi' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Modal Import Data -->
    <Dialog :open="isImportModalOpen" @update:open="(val: boolean) => { if (!val) isImportModalOpen = false }">
      <DialogContent
        class="max-w-md"
        @pointer-down-outside="(e: Event) => e.preventDefault()"
        @interact-outside="(e: Event) => e.preventDefault()"
      >
        <DialogHeader>
          <DialogTitle>Import Data Pengurus</DialogTitle>
          <DialogDescription>Import massal pengurus melalui berkas CSV atau Google Sheets.</DialogDescription>
        </DialogHeader>

        <div class="p-1 bg-zinc-100 rounded-2xl flex items-center gap-1">
          <button
            type="button"
            @click="importMethod = 'csv'"
            class="flex-1 py-1.5 rounded-xl text-xs font-semibold transition-all"
            :class="importMethod === 'csv' ? 'bg-white text-zinc-900 shadow-sm' : 'text-zinc-500 hover:text-zinc-900'"
          >
            Upload File CSV
          </button>
          <button
            type="button"
            @click="importMethod = 'sheet'"
            class="flex-1 py-1.5 rounded-xl text-xs font-semibold transition-all"
            :class="importMethod === 'sheet' ? 'bg-white text-zinc-900 shadow-sm' : 'text-zinc-500 hover:text-zinc-900'"
          >
            Link Google Sheets
          </button>
        </div>

        <div class="text-xs text-zinc-500 leading-relaxed bg-zinc-50 p-3 rounded-2xl border border-zinc-200">
          <p class="font-medium text-zinc-800 mb-1">Format urutan kolom yang didukung:</p>
          <code class="block font-mono text-[11px] text-zinc-700 bg-white p-2 rounded-xl border border-zinc-200">
            nim,nama,email,divisi
          </code>
          <p class="mt-1 text-[11px] text-zinc-500">
            Seluruh akun hasil import akan menerima tautan aktivasi mandiri ke email masing-masing.
          </p>
        </div>

        <div v-if="importMethod === 'csv'" class="flex flex-col gap-2">
          <Label>Pilih File CSV</Label>
          <input
            type="file"
            accept=".csv"
            @change="onFileChange"
            class="text-xs file:mr-3 file:py-2 file:px-4 file:rounded-2xl file:border-0 file:text-xs file:font-semibold file:bg-[#0A84DC]/10 file:text-[#0A84DC] hover:file:bg-[#0A84DC]/20 cursor-pointer"
          />
        </div>

        <div v-else class="flex flex-col gap-2">
          <Label for="sheet-url">URL / Link Google Sheets</Label>
          <Input
            id="sheet-url"
            v-model="sheetUrl"
            type="url"
            placeholder="https://docs.google.com/spreadsheets/d/.../edit"
          />
          <span class="text-[11px] text-zinc-400">
            *Pastikan akses link sheet diset ke: <b>"Siapa saja yang memiliki link dapat melihat"</b>.
          </span>
        </div>

        <DialogFooter class="pt-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            @click="isImportModalOpen = false"
          >
            Batal
          </Button>
          <Button
            type="button"
            size="sm"
            @click="handleImportData"
            :disabled="isSubmitting || (importMethod === 'csv' ? !importFile : !sheetUrl.trim())"
          >
            {{ isSubmitting ? 'Memproses...' : (importMethod === 'csv' ? 'Mulai Import CSV' : 'Tarik Data Spreadsheet') }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Modal Detail Pengurus -->
    <Dialog :open="isDetailModalOpen && !!selectedUser" @update:open="(val: boolean) => { if (!val) isDetailModalOpen = false }">
      <DialogContent class="max-w-md" v-if="selectedUser">
        <DialogHeader>
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-2xl bg-[#0A84DC]/10 text-[#0A84DC] flex items-center justify-center font-bold text-sm">
              {{ (selectedUser.nama || selectedUser.username).charAt(0).toUpperCase() }}
            </div>
            <div>
              <DialogTitle>Detail Pengurus</DialogTitle>
              <DialogDescription>@{{ selectedUser.username }}</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div class="grid grid-cols-1 gap-2.5 bg-zinc-50 p-4 rounded-2xl border border-zinc-200 text-xs">
          <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
            <span class="text-zinc-500">Nama Lengkap</span>
            <span class="font-semibold text-zinc-900 text-right">{{ selectedUser.nama || '-' }}</span>
          </div>
          <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
            <span class="text-zinc-500">NIM</span>
            <span class="font-mono text-zinc-800 text-right">{{ selectedUser.nim || '-' }}</span>
          </div>
          <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
            <span class="text-zinc-500">Divisi</span>
            <Badge variant="mono" size="sm">
              {{ selectedUser.divisi }}
            </Badge>
          </div>
          <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
            <span class="text-zinc-500">Email</span>
            <span class="text-zinc-800 text-right select-all">{{ selectedUser.email || '-' }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-zinc-500">Status Akun</span>
            <Badge
              :variant="selectedUser.is_active ? 'emerald' : 'amber'"
              size="sm"
            >
              {{ selectedUser.is_active ? 'Aktif' : 'Menunggu Aktivasi' }}
            </Badge>
          </div>
        </div>

        <DialogFooter class="pt-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            @click="isDetailModalOpen = false"
          >
            Tutup
          </Button>
          <Button
            type="button"
            size="sm"
            @click="isDetailModalOpen = false; openEditModal(selectedUser)"
          >
            <svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            Edit Pengurus
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- Modal Edit Pengurus -->
    <Dialog :open="isEditModalOpen" @update:open="(val: boolean) => { if (!val) isEditModalOpen = false }">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Data Pengurus</DialogTitle>
          <DialogDescription>Perbarui data profil dan penempatan divisi.</DialogDescription>
        </DialogHeader>

        <form @submit.prevent="handleUpdateUser" novalidate class="flex flex-col gap-3">
          <div class="space-y-1">
            <Label for="edit-user-nama">Nama Lengkap *</Label>
            <Input
              id="edit-user-nama"
              v-model="editForm.nama"
              required
            />
          </div>

          <div class="space-y-1">
            <Label for="edit-user-nim">NIM (Nomor Induk Mahasiswa)</Label>
            <Input
              id="edit-user-nim"
              v-model="editForm.nim"
              placeholder="Contoh: A11.2023.12345"
            />
          </div>

          <div class="space-y-1">
            <Label for="edit-user-email">Email</Label>
            <Input
              id="edit-user-email"
              v-model="editForm.email"
              type="email"
            />
          </div>

          <div class="space-y-1">
            <Label for="edit-user-divisi">Divisi</Label>
            <Select v-model="editForm.divisi" :disabled="editForm.id === currentUserId && editForm.divisi === 'ketum'">
              <SelectTrigger id="edit-user-divisi">
                <SelectValue placeholder="Pilih Divisi" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="d in assignableDivisiOptions"
                  :key="d.value"
                  :value="d.value"
                >
                  {{ d.label }}
                </SelectItem>
              </SelectContent>
            </Select>
            <span v-if="editForm.id === currentUserId && editForm.divisi === 'ketum'" class="text-[10px] text-zinc-400 mt-1 block">
              *Pergantian jabatan Ketua Umum wajib melalui menu Serah Terima Jabatan.
            </span>
          </div>

          <DialogFooter class="pt-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              @click="isEditModalOpen = false"
            >
              Batal
            </Button>
            <Button
              type="submit"
              size="sm"
              :disabled="isSubmitting"
            >
              {{ isSubmitting ? 'Menyimpan...' : 'Simpan Perubahan' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Modal Serah Terima Jabatan (Ketum Transfer) -->
    <Dialog :open="isTransferModalOpen" @update:open="(val: boolean) => { if (!val) isTransferModalOpen = false }">
      <DialogContent class="max-w-md">
        <DialogHeader>
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-2xl bg-orange-100 text-orange-600 flex items-center justify-center shrink-0">
              <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
            </div>
            <div>
              <DialogTitle>Serah Terima Jabatan Ketua Umum</DialogTitle>
              <DialogDescription>Alihkan kepemimpinan DOSCOM kepada Ketua Umum terpilih.</DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form @submit.prevent="handleTransferKetum" novalidate class="flex flex-col gap-3.5 pt-2">
          <div class="p-3 bg-amber-50 border border-amber-200/80 rounded-2xl text-xs text-amber-900 leading-relaxed">
            <p class="font-semibold mb-1">Perhatian Penting:</p>
            Akun Anda akan dialihkan menjadi divisi <b>ex-ketum</b>, dan pengurus yang dipilih akan diangkat menjadi <b>Ketua Umum</b> baru.
          </div>

          <div class="space-y-1">
            <Label for="transfer-target">Pilih Calon Ketua Umum Baru *</Label>
            <Select v-model="transferTargetUserId">
              <SelectTrigger id="transfer-target">
                <SelectValue placeholder="Pilih pengurus penerus" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="cand in eligibleSuccessors"
                  :key="cand.id"
                  :value="cand.id"
                >
                  {{ cand.nama || cand.username }} ({{ cand.divisi.toUpperCase() }})
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div class="space-y-1">
            <Label for="transfer-password">Konfirmasi Kata Sandi Ketua Umum Saat Ini *</Label>
            <Input
              id="transfer-password"
              v-model="transferKetumPassword"
              type="password"
              required
              placeholder="Masukkan kata sandi Anda untuk otentikasi"
            />
          </div>

          <DialogFooter class="pt-3">
            <Button
              type="button"
              variant="outline"
              size="sm"
              @click="isTransferModalOpen = false"
            >
              Batal
            </Button>
            <Button
              type="submit"
              variant="orange"
              size="sm"
              :disabled="isTransferring || !transferTargetUserId || !transferKetumPassword"
            >
              {{ isTransferring ? 'Memproses Suksesi...' : 'Konfirmasi Serah Terima' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <!-- Modal Konfirmasi Hapus -->
    <Dialog :open="confirmDeleteModal.isOpen" @update:open="(val: boolean) => { if (!val) confirmDeleteModal.isOpen = false }">
      <DialogContent class="max-w-sm text-center">
        <div class="w-12 h-12 rounded-2xl bg-rose-50 text-rose-500 mx-auto flex items-center justify-center border border-rose-100">
          <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </div>

        <DialogHeader class="text-center sm:text-center">
          <DialogTitle class="text-center sm:text-center">
            {{ confirmDeleteModal.isBatch ? 'Hapus Pengurus Terpilih?' : 'Hapus Akun Pengurus?' }}
          </DialogTitle>
          <DialogDescription class="text-center sm:text-center mt-1.5 leading-relaxed">
            <template v-if="confirmDeleteModal.isBatch">
              Kamu akan menghapus <b class="text-zinc-900">{{ confirmDeleteModal.count }} pengurus</b> sekaligus. Tindakan ini tidak dapat dibatalkan.
            </template>
            <template v-else>
              Kamu yakin ingin menghapus akun <b class="text-zinc-900">{{ confirmDeleteModal.user?.nama || confirmDeleteModal.user?.username }}</b>?
            </template>
          </DialogDescription>
        </DialogHeader>

        <DialogFooter class="flex sm:justify-center gap-2 pt-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            class="flex-1"
            @click="confirmDeleteModal.isOpen = false"
          >
            Batal
          </Button>
          <Button
            type="button"
            variant="destructive"
            size="sm"
            class="flex-1"
            :disabled="isSubmitting"
            @click="executeConfirmedDelete"
          >
            {{ isSubmitting ? 'Menghapus...' : 'Ya, Hapus' }}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<style scoped>
.slide-fade-enter-active {
  transition: all 0.28s cubic-bezier(0.16, 1, 0.3, 1);
}
.slide-fade-leave-active {
  transition: all 0.2s cubic-bezier(0.7, 0, 0.84, 0);
}
.slide-fade-enter-from {
  opacity: 0;
  transform: translateY(-8px) scale(0.98);
}
.slide-fade-leave-to {
  opacity: 0;
  transform: translateY(-8px) scale(0.98);
}
</style>
