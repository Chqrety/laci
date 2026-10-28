<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { userService, currentUserId, currentRole, type User } from '@/services/api'
import { ConfirmDialog } from '@/components/common'
import {
  UserAddModal,
  UserEditModal,
  UserDetailModal,
  UserImportModal,
  UserTransferModal,
} from '@/components/user'
import { usePagination } from '@/composables/usePagination'
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
import { toast } from '@/components/ui/toast'

const users = ref<User[]>([])
const isLoading = ref(false)
const searchQuery = ref('')
const selectedDivisi = ref('')
const selectedUserIds = ref<string[]>([])
const isSubmitting = ref(false)

// Modals visibility state
const isAddModalOpen = ref(false)
const isEditModalOpen = ref(false)
const isDetailModalOpen = ref(false)
const isImportModalOpen = ref(false)
const isTransferModalOpen = ref(false)
const selectedUser = ref<User | null>(null)

// Current role check
const isKetum = computed(() => currentRole.value === 'ketum')

const fetchUsers = async () => {
  isLoading.value = true
  try {
    users.value = await userService.getAll()
  } catch (err: any) {
    toast({
      title: 'Gagal memuat data pengurus',
      description: err.response?.data?.error || err.message,
      variant: 'destructive',
    })
  } finally {
    isLoading.value = false
  }
}

onMounted(fetchUsers)

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

// Unified Pagination Composable
const {
  currentPage,
  itemsPerPage,
  totalPages,
  paginatedItems: paginatedUsers,
  nextPage,
  prevPage,
  resetPage,
} = usePagination(filteredUsers, 15)

watch([searchQuery, selectedDivisi], () => {
  resetPage()
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

// Delete confirmation state & actions
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
      toast({ title: `Berhasil menghapus ${count} pengurus`, variant: 'success' })
      selectedUserIds.value = []
    } else if (confirmDeleteModal.value.user) {
      const u = confirmDeleteModal.value.user
      await userService.delete(u.id)
      users.value = users.value.filter((item) => item.id !== u.id)
      selectedUserIds.value = selectedUserIds.value.filter((id) => id !== u.id)
      toast({ title: `Pengurus ${u.nama || u.username} berhasil dihapus`, variant: 'success' })
    }
    confirmDeleteModal.value.isOpen = false
  } catch (err: any) {
    toast({
      title: 'Gagal menghapus pengurus',
      description: err.response?.data?.error || err.message,
      variant: 'destructive',
    })
  } finally {
    isSubmitting.value = false
  }
}

// Modal open triggers
const openDetailModal = (user: User) => {
  selectedUser.value = user
  isDetailModalOpen.value = true
}

const openEditModal = (user: User) => {
  selectedUser.value = user
  isEditModalOpen.value = true
}

const handleUserCreated = async () => {
  await fetchUsers()
}

const handleUserUpdated = async () => {
  await fetchUsers()
}
</script>

<template>
  <div class="flex-1 flex flex-col gap-5">
    <!-- Header & Actions (Floema Architectural Responsive Header) -->
    <header class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5 flex-wrap">
          <h2 class="text-xl sm:text-2xl font-medium tracking-tight text-slate-900 leading-snug">Kelola Pengurus &amp; Divisi</h2>
          <Badge variant="muted" size="pill" class="shrink-0">
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
          variant="orange-outline"
          size="sm"
          class="w-full sm:w-auto"
          @click="isTransferModalOpen = true"
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
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <Select v-model="selectedDivisi">
          <SelectTrigger variant="white" class="w-full sm:w-44 lg:w-48">
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
            @click="openBatchDeleteConfirm"
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
      <div class="overflow-x-auto no-scrollbar">
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
                <div class="font-semibold text-zinc-900 flex items-center gap-1.5 flex-wrap">
                  <span>{{ u.nama || u.username }}</span>
                  <Badge
                    v-if="u.id === currentUserId"
                    variant="info"
                    size="xs"
                  >
                    Anda
                  </Badge>
                  <Badge
                    v-if="u.is_koor"
                    variant="orange"
                    size="xs"
                  >
                    Koor
                  </Badge>
                </div>
                <div class="text-[11px] text-zinc-400">@{{ u.username }} &bull; {{ u.email || 'Tanpa email' }}</div>
              </td>
              <td class="py-3.5 px-4 font-mono text-zinc-600">
                {{ u.nim || '-' }}
              </td>
              <td class="py-3.5 px-4">
                <Badge
                  :variant="u.divisi === 'ketum' ? 'orange' : 'secondary'"
                  size="xs"
                  class="font-mono uppercase tracking-wider"
                >
                  {{ u.divisi }}
                </Badge>
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
                    @click="openSingleDeleteConfirm(u)"
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
            @click="prevPage"
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
            @click="nextPage"
            :disabled="currentPage === totalPages"
          >
            Selanjutnya
          </Button>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <UserAddModal
      v-model:open="isAddModalOpen"
      :divisi-options="assignableDivisiOptions"
      @created="handleUserCreated"
    />

    <UserEditModal
      v-model:open="isEditModalOpen"
      :user="selectedUser"
      :current-user-id="currentUserId"
      :divisi-options="assignableDivisiOptions"
      @updated="handleUserUpdated"
    />

    <UserDetailModal
      v-model:open="isDetailModalOpen"
      :user="selectedUser"
      @edit="openEditModal"
    />

    <UserImportModal
      v-model:open="isImportModalOpen"
      @imported="fetchUsers"
    />

    <UserTransferModal
      v-model:open="isTransferModalOpen"
      :eligible-successors="eligibleSuccessors"
      @transferred="fetchUsers"
    />

    <!-- Modal Konfirmasi Hapus (Reusable ConfirmDialog) -->
    <ConfirmDialog
      :open="confirmDeleteModal.isOpen"
      :title="confirmDeleteModal.isBatch ? 'Hapus Pengurus Terpilih?' : 'Hapus Akun Pengurus?'"
      confirm-text="Ya, Hapus"
      loading-text="Menghapus..."
      :is-loading="isSubmitting"
      @update:open="(val: boolean) => { if (!val) confirmDeleteModal.isOpen = false }"
      @cancel="confirmDeleteModal.isOpen = false"
      @confirm="executeConfirmedDelete"
    >
      <template #description>
        <template v-if="confirmDeleteModal.isBatch">
          Kamu akan menghapus <b class="text-zinc-900">{{ confirmDeleteModal.count }} pengurus</b> sekaligus. Tindakan ini tidak dapat dibatalkan.
        </template>
        <template v-else>
          Kamu yakin ingin menghapus akun <b class="text-zinc-900">{{ confirmDeleteModal.user?.nama || confirmDeleteModal.user?.username }}</b>?
        </template>
      </template>
    </ConfirmDialog>
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
