<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { userService, currentUserId, type User } from '@/services/api'

const users = ref<User[]>([])
const isLoading = ref(false)
const searchQuery = ref('')
const selectedDivisi = ref('')
const isImportModalOpen = ref(false)
const isAddModalOpen = ref(false)
const isEditModalOpen = ref(false)
const isDetailModalOpen = ref(false)
const selectedUser = ref<User | null>(null)
const importFile = ref<File | null>(null)
const isSubmitting = ref(false)
const toastMsg = ref<{ type: 'success' | 'error'; text: string } | null>(null)

// Pagination state
const currentPage = ref(1)
const itemsPerPage = ref(15)

// Form edit user
const editForm = ref({
  id: '',
  nim: '',
  nama: '',
  email: '',
  divisi: 'humas',
  password: '',
})

// Password visibility state per user id
const visiblePasswords = ref<Record<string, boolean>>({})

// Form tambah user manual
const formUser = ref({
  nim: '',
  nama: '',
  email: '',
  username: '',
  password: '',
  divisi: 'humas',
})

const showToast = (text: string, type: 'success' | 'error' = 'success') => {
  toastMsg.value = { type, text }
  setTimeout(() => {
    toastMsg.value = null
  }, 4000)
}

const togglePasswordVisibility = (userId: string) => {
  visiblePasswords.value[userId] = !visiblePasswords.value[userId]
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

    const matchesDivisi = !selectedDivisi.value || u.divisi === selectedDivisi.value
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

const handleResetPrompt = async (user: User) => {
  try {
    const res = await userService.resetPrompt(user.id)
    user.has_seen_password_prompt = false
    showToast(`Berhasil: Reminder aktif kembali untuk ${user.nama}`)
  } catch (err: any) {
    showToast(err.response?.data?.error || 'Gagal mereset reminder', 'error')
  }
}

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
  if (!formUser.value.nama) {
    showToast('Nama wajib diisi', 'error')
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
      password: '',
      divisi: 'humas',
    }
    showToast(`Pengurus ${created.nama} berhasil ditambahkan`)
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
    password: '',
  }
  isEditModalOpen.value = true
}

const handleUpdateUser = async () => {
  if (!editForm.value.nama) {
    showToast('Nama wajib diisi', 'error')
    return
  }

  isSubmitting.value = true
  try {
    const updated = await userService.update(editForm.value.id, {
      nim: editForm.value.nim,
      nama: editForm.value.nama,
      email: editForm.value.email,
      divisi: editForm.value.divisi,
      password: editForm.value.password ? editForm.value.password : undefined,
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
        alert(`Catatan: ${res.skipped.length} baris dilewati karena duplikat:\n` + res.skipped.join('\n'))
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
        alert(`Catatan: ${res.skipped.length} baris dilewati karena duplikat:\n` + res.skipped.join('\n'))
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
    <!-- Toast notification -->
    <div
      v-if="toastMsg"
      class="fixed bottom-6 right-6 z-50 px-4 py-3 rounded-2xl shadow-xl text-xs font-semibold flex items-center gap-2 border transition-all animate-bounce-in"
      :class="
        toastMsg.type === 'success'
          ? 'bg-emerald-500 text-white border-emerald-400'
          : 'bg-rose-500 text-white border-rose-400'
      "
    >
      <span>{{ toastMsg.text }}</span>
    </div>

    <!-- Header & Actions -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
      <div>
        <h2 class="text-xl font-medium tracking-tight text-slate-900 leading-snug">Kelola Pengurus & Divisi</h2>
        <p class="text-xs text-slate-500 mt-0.5">Kelola data humas/pengurus, import CSV/Spreadsheet, dan monitoring password</p>
      </div>

      <div class="flex items-center gap-2 flex-wrap">
        <button
          type="button"
          @click="isImportModalOpen = true"
          class="px-4 py-2 rounded-full bg-white border border-slate-200 text-slate-700 text-xs font-semibold flex items-center gap-1.5 shadow-xs hover:bg-slate-50 active:scale-[0.98] transition-all cursor-pointer"
        >
          <svg class="w-4 h-4 text-sky-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
          </svg>
          Import Data
        </button>

        <button
          type="button"
          @click="isAddModalOpen = true"
          class="px-4.5 py-2 rounded-full bg-sky-600 hover:bg-sky-700 text-white text-xs font-bold flex items-center gap-1.5 shadow-xs active:scale-[0.98] transition-all cursor-pointer"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          Tambah Pengurus
        </button>
      </div>
    </div>

    <!-- Filter Bar (Konsisten dengan SuratManagement.vue & ProkerManagement.vue) -->
    <div class="flex items-center flex-wrap gap-2">
      <input
        v-model="searchQuery"
        type="search"
        placeholder="Cari nama, NIM, email, username..."
        class="px-4 py-2 text-xs rounded-full bg-white border border-slate-200 text-slate-800 placeholder:text-slate-400 focus:outline-none focus:ring-4 focus:ring-sky-100 focus:border-sky-500 shadow-xs transition-all flex-1 min-w-[200px]"
      />

      <select
        v-model="selectedDivisi"
        class="px-3.5 py-2 text-xs rounded-full bg-white border border-slate-200 text-slate-800 focus:outline-none focus:ring-4 focus:ring-sky-100 focus:border-sky-500 shadow-xs transition-all cursor-pointer"
      >
        <option value="">Semua Divisi</option>
        <option v-for="d in divisiOptions" :key="d" :value="d">
         {{ d.toUpperCase() }}
        </option>
      </select>
    </div>

    <!-- Bulk Selection Bar (Clean iOS HIG: rounded-2xl, border-black/5, shadow-[0_8px_30px_rgb(0,0,0,0.05)]) -->
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
          <button
            type="button"
            @click="selectedUserIds = []"
            class="px-3 py-1.5 rounded-xl text-xs font-medium text-zinc-500 hover:text-zinc-800 hover:bg-black/[0.04] transition-colors"
          >
            Batal Pilih
          </button>
          <button
            type="button"
            @click="handleBatchDelete"
            :disabled="isSubmitting"
            class="px-3.5 py-1.5 rounded-xl text-xs font-semibold bg-rose-50 hover:bg-rose-100 text-rose-700 border border-rose-200 active:scale-95 transition-all flex items-center gap-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            Hapus Terpilih ({{ selectedUserIds.length }})
          </button>
        </div>
      </div>
    </Transition>

    <!-- Table Container (Persis seperti SuratManagement & ProkerManagement: bg-white/85, backdrop-blur-xl, border border-black/5, rounded-2xl) -->
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
              <th scope="col" class="py-3 px-4">Password Akun</th>
              <th scope="col" class="py-3 px-4">Status Password</th>
              <th scope="col" class="py-3 px-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-black/5 text-zinc-800">
            <tr v-if="isLoading" class="animate-pulse">
              <td colspan="7" class="py-8 text-center text-zinc-400">Memuat data pengurus...</td>
            </tr>

            <tr v-else-if="filteredUsers.length === 0">
              <td colspan="7" class="py-8 text-center text-zinc-400">Tidak ada data pengurus ditemukan.</td>
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
                <span class="px-2.5 py-1 rounded-xl text-[10px] font-mono font-semibold uppercase tracking-wider bg-zinc-100 text-zinc-700 border border-zinc-200">
                  {{ u.divisi }}
                </span>
              </td>
              <td class="py-3.5 px-4">
                <div class="flex items-center gap-2">
                  <span class="font-mono text-xs text-zinc-800 select-all">
                    {{ visiblePasswords[u.id] ? (u.display_password || '********') : '••••••••' }}
                  </span>
                  <button
                    type="button"
                    @click="togglePasswordVisibility(u.id)"
                    class="p-1 rounded-lg text-zinc-400 hover:text-zinc-700 hover:bg-zinc-100 transition-colors"
                    :title="visiblePasswords[u.id] ? 'Sembunyikan password' : 'Lihat password'"
                  >
                    <!-- Eye icon -->
                    <svg v-if="!visiblePasswords[u.id]" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <!-- Eye off icon -->
                    <svg v-else class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l18 18" />
                    </svg>
                  </button>
                </div>
              </td>
              <td class="py-3.5 px-4">
                <div class="flex items-center gap-1.5">
                  <span
                    class="px-2 py-0.5 rounded-lg text-[10px] font-semibold"
                    :class="
                      u.is_password_default
                        ? 'bg-amber-50 text-amber-700 border border-amber-200'
                        : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                    "
                  >
                    {{ u.is_password_default ? 'Default' : 'Kustom' }}
                  </span>

                  <span
                    v-if="!u.has_seen_password_prompt"
                    class="text-[10px] text-zinc-400"
                    title="Modal reminder akan muncul saat login berikutnya"
                  >
                    (Reminder Aktif)
                  </span>
                </div>
              </td>
              <td class="py-3.5 px-4 text-right">
                <div class="flex items-center justify-end gap-1">
                  <!-- Tombol Detail Pengurus (Info / User Card Icon) -->
                  <button
                    type="button"
                    @click="openDetailModal(u)"
                    class="p-2 rounded-xl text-zinc-400 hover:text-zinc-800 hover:bg-zinc-100 active:scale-95 transition-all"
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
                  </button>

                  <!-- Tombol Edit Pengurus (Pencil Icon) -->
                  <button
                    type="button"
                    @click="openEditModal(u)"
                    class="p-2 rounded-xl text-zinc-400 hover:text-amber-600 hover:bg-amber-50 active:scale-95 transition-all"
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
                  </button>

                  <!-- Tombol Ingatkan Ganti Password (Bell / Reminder Icon) -->
                  <button
                    type="button"
                    @click="handleResetPrompt(u)"
                    class="p-2 rounded-xl text-zinc-400 hover:text-[#0A84DC] hover:bg-[#0A84DC]/10 active:scale-95 transition-all"
                    title="Ingatkan Ganti Password (Picu modal saat login berikutnya)"
                    aria-label="Ingatkan Ganti Password"
                  >
                    <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        stroke-width="1.8"
                        d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
                      />
                    </svg>
                  </button>

                  <!-- Tombol Hapus Pengurus (Trash Icon) -->
                  <button
                    v-if="u.id !== currentUserId"
                    type="button"
                    @click="handleDeleteUser(u)"
                    class="p-2 rounded-xl text-zinc-400 hover:text-rose-600 hover:bg-rose-50 active:scale-95 transition-all"
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
                  </button>
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
          <button
            type="button"
            @click="currentPage--"
            :disabled="currentPage === 1"
            class="px-3 py-1 rounded-xl border border-black/5 bg-white text-zinc-700 hover:bg-zinc-50 active:scale-95 disabled:opacity-40 transition-all font-medium"
          >
            Sebelumnya
          </button>
          <span class="px-2 font-mono text-[11px] font-semibold text-zinc-700">
            {{ currentPage }} / {{ totalPages }}
          </span>
          <button
            type="button"
            @click="currentPage++"
            :disabled="currentPage === totalPages"
            class="px-3 py-1 rounded-xl border border-black/5 bg-white text-zinc-700 hover:bg-zinc-50 active:scale-95 disabled:opacity-40 transition-all font-medium"
          >
            Selanjutnya
          </button>
        </div>
      </div>
    </div>

    <!-- Modal Tambah Pengurus -->
    <Transition name="modal-fade">
      <div
        v-if="isAddModalOpen"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      >
        <div class="modal-panel bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-black/5 flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold text-zinc-900">Tambah Pengurus Baru</h3>
          <button @click="isAddModalOpen = false" class="text-zinc-400 hover:text-zinc-700 text-sm">✕</button>
        </div>

        <form @submit.prevent="handleCreateUser" class="flex flex-col gap-3">
          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">Nama Lengkap *</label>
            <input
              v-model="formUser.nama"
              required
              placeholder="Contoh: Budi Santoso"
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">NIM (Nomor Induk Mahasiswa)</label>
            <input
              v-model="formUser.nim"
              placeholder="Contoh: A11.2023.12345"
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">Email</label>
            <input
              v-model="formUser.email"
              type="email"
              placeholder="budi@doscom.org"
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-medium text-zinc-700 mb-1">Username (Opsional)</label>
              <input
                v-model="formUser.username"
                placeholder="Auto dari NIM/Nama"
                class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-zinc-700 mb-1">Divisi</label>
              <select
                v-model="formUser.divisi"
                class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
              >
                <option value="humas">Humas</option>
                <option value="sekre">Sekre</option>
                <option value="inti">Pengurus Inti</option>
                <option value="acara">Divisi Acara</option>
                <option value="perlengkapan">Divisi Perlengkapan</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">Password Awal (Opsional)</label>
            <input
              v-model="formUser.password"
              placeholder="Kosongkan untuk acak otomatis"
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
          </div>

          <div class="flex items-center justify-end gap-2 mt-2">
            <button
              type="button"
              @click="isAddModalOpen = false"
              class="px-4 py-2 rounded-2xl text-xs text-zinc-600 hover:bg-zinc-100"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isSubmitting"
              class="px-4 py-2 rounded-2xl text-xs font-semibold text-white bg-[#0A84DC] hover:bg-[#086db8] disabled:opacity-50"
            >
              {{ isSubmitting ? 'Menyimpan...' : 'Simpan Pengurus' }}
            </button>
          </div>
        </form>
      </div>
    </div>
    </Transition>

    <!-- Modal Import Data (CSV atau Google Sheets) -->
    <Transition name="modal-fade">
      <div
        v-if="isImportModalOpen"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      >
        <div class="modal-panel bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-black/5 flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <div>
            <h3 class="text-base font-semibold text-zinc-900">Import Data Pengurus</h3>
            <p class="text-xs text-zinc-500 mt-0.5">Pilih metode import yang kamu inginkan</p>
          </div>
          <button @click="isImportModalOpen = false" class="text-zinc-400 hover:text-zinc-700 text-sm">✕</button>
        </div>

        <!-- Segmented Tab: CSV File vs Google Sheets Link -->
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

        <!-- Format Info Box -->
        <div class="text-xs text-zinc-500 leading-relaxed bg-zinc-50 p-3 rounded-2xl border border-zinc-200">
          <p class="font-medium text-zinc-800 mb-1">Format urutan kolom yang didukung:</p>
          <code class="block font-mono text-[11px] text-zinc-700 bg-white p-2 rounded-xl border border-zinc-200">
            nim,nama,email,divisi
          </code>
          <p class="mt-1 text-[11px] text-zinc-500">
            Password awal di-generate otomatis dan dikirimkan ke email anggota.
          </p>
        </div>

        <!-- Input Tab CSV -->
        <div v-if="importMethod === 'csv'" class="flex flex-col gap-2">
          <label class="block text-xs font-medium text-zinc-700">Pilih File CSV</label>
          <input
            type="file"
            accept=".csv"
            @change="onFileChange"
            class="text-xs file:mr-3 file:py-2 file:px-4 file:rounded-2xl file:border-0 file:text-xs file:font-semibold file:bg-[#0A84DC]/10 file:text-[#0A84DC] hover:file:bg-[#0A84DC]/20 cursor-pointer"
          />
        </div>

        <!-- Input Tab Google Sheets -->
        <div v-else class="flex flex-col gap-2">
          <label class="block text-xs font-medium text-zinc-700">URL / Link Google Sheets</label>
          <input
            v-model="sheetUrl"
            type="url"
            placeholder="https://docs.google.com/spreadsheets/d/.../edit"
            class="w-full px-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC] transition-all"
          />
          <span class="text-[11px] text-zinc-400">
            *Pastikan akses link sheet diset ke: <b>"Siapa saja yang memiliki link dapat melihat"</b>.
          </span>
        </div>

        <!-- Modal Actions -->
        <div class="flex items-center justify-end gap-2 mt-2">
          <button
            type="button"
            @click="isImportModalOpen = false"
            class="px-4 py-2 rounded-2xl text-xs text-zinc-600 hover:bg-zinc-100"
          >
            Batal
          </button>
          <button
            type="button"
            @click="handleImportData"
            :disabled="isSubmitting || (importMethod === 'csv' ? !importFile : !sheetUrl.trim())"
            class="px-5 py-2 rounded-2xl text-xs font-semibold text-white bg-[#0A84DC] hover:bg-[#086db8] active:scale-[0.98] transition-all shadow-[0_8px_20px_rgba(10,132,220,0.25)] disabled:opacity-50"
          >
            {{ isSubmitting ? 'Memproses...' : (importMethod === 'csv' ? 'Mulai Import CSV' : 'Tarik Data Spreadsheet') }}
          </button>
        </div>
      </div>
    </div>
    </Transition>

    <!-- Modal Detail Pengurus -->
    <Transition name="modal-fade">
      <div
        v-if="isDetailModalOpen && selectedUser"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      >
        <div class="modal-panel bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-black/5 flex flex-col gap-5 relative overflow-hidden">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-11 h-11 rounded-2xl bg-[#0A84DC]/10 text-[#0A84DC] flex items-center justify-center font-bold text-sm">
              {{ (selectedUser.nama || selectedUser.username).charAt(0).toUpperCase() }}
            </div>
            <div>
              <h3 class="text-base font-semibold text-zinc-900 leading-tight">Detail Pengurus</h3>
              <p class="text-xs text-zinc-400">@{{ selectedUser.username }}</p>
            </div>
          </div>
          <button @click="isDetailModalOpen = false" class="text-zinc-400 hover:text-zinc-700 text-sm">✕</button>
        </div>

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
            <span class="px-2 py-0.5 rounded-lg text-[10px] font-mono font-semibold uppercase tracking-wider bg-zinc-200/70 text-zinc-800">
              {{ selectedUser.divisi }}
            </span>
          </div>
          <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
            <span class="text-zinc-500">Email</span>
            <span class="text-zinc-800 text-right select-all">{{ selectedUser.email || '-' }}</span>
          </div>
          <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
            <span class="text-zinc-500">Password Sistem</span>
            <span class="font-mono text-xs font-semibold text-zinc-900 select-all">
              {{ selectedUser.display_password || '••••••••' }}
            </span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-zinc-500">Status Password</span>
            <span
              class="px-2 py-0.5 rounded-lg text-[10px] font-semibold"
              :class="
                selectedUser.is_password_default
                  ? 'bg-amber-50 text-amber-700 border border-amber-200'
                  : 'bg-emerald-50 text-emerald-700 border border-emerald-200'
              "
            >
              {{ selectedUser.is_password_default ? 'Default Sistem' : 'Kustom Mandiri' }}
            </span>
          </div>
        </div>

        <div class="flex items-center justify-end gap-2">
          <button
            type="button"
            @click="isDetailModalOpen = false"
            class="px-4 py-2 rounded-2xl text-xs text-zinc-600 hover:bg-zinc-100"
          >
            Tutup
          </button>
          <button
            type="button"
            @click="isDetailModalOpen = false; openEditModal(selectedUser)"
            class="px-4 py-2 rounded-2xl text-xs font-semibold text-white bg-[#0A84DC] hover:bg-[#086db8] flex items-center gap-1.5"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            Edit Pengurus
          </button>
        </div>
      </div>
    </div>
    </Transition>

    <!-- Modal Edit Pengurus -->
    <Transition name="modal-fade">
      <div
        v-if="isEditModalOpen"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      >
        <div class="modal-panel bg-white rounded-3xl max-w-md w-full p-6 shadow-2xl border border-black/5 flex flex-col gap-4">
        <div class="flex items-center justify-between">
          <h3 class="text-base font-semibold text-zinc-900">Edit Data Pengurus</h3>
          <button @click="isEditModalOpen = false" class="text-zinc-400 hover:text-zinc-700 text-sm">✕</button>
        </div>

        <form @submit.prevent="handleUpdateUser" class="flex flex-col gap-3">
          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">Nama Lengkap *</label>
            <input
              v-model="editForm.nama"
              required
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">NIM (Nomor Induk Mahasiswa)</label>
            <input
              v-model="editForm.nim"
              placeholder="Contoh: A11.2023.12345"
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">Email</label>
            <input
              v-model="editForm.email"
              type="email"
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">Divisi</label>
            <select
              v-model="editForm.divisi"
              :disabled="editForm.id === currentUserId"
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC] disabled:opacity-60 disabled:cursor-not-allowed"
            >
              <option value="humas">Humas</option>
              <option value="sekre">Sekre</option>
              <option value="inti">Pengurus Inti</option>
              <option value="acara">Divisi Acara</option>
              <option value="perlengkapan">Divisi Perlengkapan</option>
            </select>
            <span v-if="editForm.id === currentUserId" class="text-[10px] text-zinc-400 mt-1 block">
              *Divisi akun Anda sendiri terkunci sebagai Sekre demi keamanan akses.
            </span>
          </div>

          <div>
            <label class="block text-xs font-medium text-zinc-700 mb-1">Reset Password Baru (Opsional)</label>
            <input
              v-model="editForm.password"
              placeholder="Kosongkan jika tidak ingin mengubah password"
              class="w-full px-3 py-2 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC]"
            />
            <span class="text-[10px] text-zinc-400 mt-1 block">
              *Jika diisi, password akun ini akan otomatis diganti dan statusnya direset menjadi default.
            </span>
          </div>

          <div class="flex items-center justify-end gap-2 mt-2">
            <button
              type="button"
              @click="isEditModalOpen = false"
              class="px-4 py-2 rounded-2xl text-xs text-zinc-600 hover:bg-zinc-100"
            >
              Batal
            </button>
            <button
              type="submit"
              :disabled="isSubmitting"
              class="px-4 py-2 rounded-2xl text-xs font-semibold text-white bg-[#0A84DC] hover:bg-[#086db8] disabled:opacity-50"
            >
              {{ isSubmitting ? 'Menyimpan...' : 'Simpan Perubahan' }}
            </button>
          </div>
        </form>
      </div>
    </div>
    </Transition>

    <!-- Modal Konfirmasi Hapus (Single / Batch) with Smooth Scale Transition -->
    <Transition name="modal-fade">
      <div
        v-if="confirmDeleteModal.isOpen"
        class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm"
      >
        <div class="modal-panel bg-white rounded-3xl max-w-sm w-full p-6 shadow-2xl border border-black/5 flex flex-col gap-4 text-center">
          <!-- Trash warning icon -->
          <div class="w-12 h-12 rounded-2xl bg-rose-50 text-rose-500 mx-auto flex items-center justify-center border border-rose-100">
            <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.8" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </div>

          <div>
            <h3 class="text-base font-bold text-zinc-900 tracking-tight">
              {{ confirmDeleteModal.isBatch ? 'Hapus Pengurus Terpilih?' : 'Hapus Akun Pengurus?' }}
            </h3>
            <p class="text-xs text-zinc-500 mt-1.5 leading-relaxed">
              <template v-if="confirmDeleteModal.isBatch">
                Kamu akan menghapus <b class="text-zinc-900">{{ confirmDeleteModal.count }} pengurus</b> sekaligus. Tindakan ini tidak dapat dibatalkan.
              </template>
              <template v-else>
                Kamu yakin ingin menghapus akun <b class="text-zinc-900">{{ confirmDeleteModal.user?.nama || confirmDeleteModal.user?.username }}</b>?
              </template>
            </p>
          </div>

          <div class="flex items-center justify-center gap-2 pt-1">
            <button
              type="button"
              @click="confirmDeleteModal.isOpen = false"
              class="px-4 py-2.5 rounded-2xl text-xs font-medium text-zinc-600 hover:bg-zinc-100 transition-all flex-1"
            >
              Batal
            </button>
            <button
              type="button"
              @click="executeConfirmedDelete"
              :disabled="isSubmitting"
              class="px-4 py-2.5 rounded-2xl text-xs font-semibold text-white bg-rose-600 hover:bg-rose-700 active:scale-95 transition-all shadow-[0_8px_20px_rgba(225,29,72,0.25)] flex-1 disabled:opacity-50"
            >
              {{ isSubmitting ? 'Menghapus...' : 'Ya, Hapus' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
/* Smooth Slide-Fade for Floating Selection Bar */
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

/* Smooth Scale-Fade for Modal Pop-up */
.modal-pop-enter-active {
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.modal-pop-leave-active {
  transition: all 0.18s ease-in;
}
.modal-pop-enter-from {
  opacity: 0;
  transform: scale(0.95);
}
.modal-pop-leave-to {
  opacity: 0;
  transform: scale(0.96);
}
</style>
