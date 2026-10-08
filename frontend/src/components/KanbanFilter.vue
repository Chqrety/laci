<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { authService, type Surat, type Proker } from '@/services/api'

const props = defineProps<{
  suratList: Surat[]
  prokerList: Proker[]
}>()

const emit = defineEmits<{
  (e: 'update:filtered', value: Surat[]): void
}>()

// Filter states
const filterTahun = ref('')
const filterProker = ref('')
const filterOnlyMe = ref(false)
const searchQuery = ref('')

const clearFilters = () => {
  filterTahun.value = ''
  filterProker.value = ''
  filterOnlyMe.value = false
  searchQuery.value = ''
}

const isFilterActive = computed(() => {
  return !!(
    filterTahun.value ||
    filterProker.value ||
    filterOnlyMe.value ||
    searchQuery.value.trim()
  )
})

const availableYears = computed(() => {
  const years = new Set<string>()
  const currentYear = new Date().getFullYear().toString()
  years.add(currentYear)
  years.add('2025')
  years.add('2026')

  props.suratList.forEach((s) => {
    const match = s.nomor_surat.match(/\b(20\d{2})\b/)
    if (match && match[1]) {
      years.add(match[1])
    }
  })
  return Array.from(years).sort().reverse()
})

const filteredSuratList = computed(() => {
  return props.suratList.filter((s) => {
    // 1. Filter Tahun
    if (filterTahun.value) {
      const match = s.nomor_surat.match(/\b(20\d{2})\b/)
      const year = match ? match[1] : ''
      if (year !== filterTahun.value && !s.nomor_surat.includes(filterTahun.value)) {
        return false
      }
    }

    // 2. Filter Proker
    if (filterProker.value) {
      const prokerCode = filterProker.value.toUpperCase().replace(/\s+/g, '-')
      const prokerLower = filterProker.value.toLowerCase()
      const matchesProker =
        (s as any).proker_id === filterProker.value ||
        s.nomor_surat.toUpperCase().includes(`/${prokerCode}/`) ||
        s.perihal.toLowerCase().includes(prokerLower)
      if (!matchesProker) {
        return false
      }
    }

    // 3. Filter Hanya Surat Saya (Assigned to Me)
    if (filterOnlyMe.value) {
      const currentRole = authService.getRole() || ''
      const currentUname = (authService.getUsername() || currentRole).trim().toLowerCase()
      const currentRoleVal = currentRole.trim().toLowerCase()
      const currentUid = (authService.getUser().id || '').trim().toLowerCase()
      const pic = (s.pic_nama || '').trim().toLowerCase()

      const matchesMe =
        (currentUname && (pic === currentUname || pic.includes(currentUname))) ||
        (currentRoleVal && (pic === currentRoleVal || pic.includes(currentRoleVal))) ||
        (currentUid && (s as any).user_id === currentUid)

      if (!matchesMe) {
        return false
      }
    }

    // 4. Search query
    if (searchQuery.value.trim()) {
      const q = searchQuery.value.trim().toLowerCase()
      const matchSearch =
        (s.pic_nama && s.pic_nama.toLowerCase().includes(q)) ||
        (s.perihal && s.perihal.toLowerCase().includes(q)) ||
        (s.nomor_surat && s.nomor_surat.toLowerCase().includes(q))
      if (!matchSearch) {
        return false
      }
    }

    return true
  })
})

watch(
  filteredSuratList,
  (val) => {
    emit('update:filtered', val)
  },
  { immediate: true, deep: true }
)
</script>

<template>
  <div
    class="bg-white/80 backdrop-blur-md border border-black/5 rounded-2xl p-3 shadow-[0_8px_30px_rgb(0,0,0,0.04)]"
  >
    <div class="flex flex-col lg:flex-row items-center gap-4 w-full">
      <!-- Search Input -->
      <div class="flex-1 w-full relative">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Cari PIC, perihal, nomor..."
          class="w-full pl-8 pr-3 py-1.5 text-xs rounded-xl bg-black/[0.03] border border-black/5 text-zinc-800 placeholder-zinc-400 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:bg-white transition-all"
        />
        <svg
          class="w-3.5 h-3.5 text-zinc-400 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
      </div>

      <!-- Dropdown Tahun -->
      <div class="flex-1 w-full">
        <select
          v-model="filterTahun"
          class="w-full px-2.5 py-1.5 text-xs rounded-xl bg-black/[0.03] border border-black/5 text-zinc-800 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:bg-white transition-all cursor-pointer"
        >
          <option value="">Semua Tahun</option>
          <option v-for="yr in availableYears" :key="yr" :value="yr">Tahun {{ yr }}</option>
        </select>
      </div>

      <!-- Dropdown Proker -->
      <div class="flex-1 w-full">
        <select
          v-model="filterProker"
          class="w-full px-2.5 py-1.5 text-xs rounded-xl bg-black/[0.03] border border-black/5 text-zinc-800 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:bg-white transition-all cursor-pointer"
        >
          <option value="">Semua Proker</option>
          <option v-for="proker in prokerList" :key="proker.id" :value="proker.nama_proker">
            {{ proker.nama_proker }}
          </option>
        </select>
      </div>

      <!-- Toggle & Clear Filter Container -->
      <div class="flex items-center gap-3 shrink-0">
        <!-- iOS Segmented Control: Semua Surat (Group) vs Hanya Surat Saya (User) -->
        <div
          role="group"
          aria-label="Filter Kepemilikan Surat"
          class="relative inline-flex items-center w-24 h-8 p-1 rounded-full bg-gray-200 select-none"
        >
          <!-- Background Sliding Indicator -->
          <span
            class="absolute top-1 bottom-1 left-1 w-[calc(50%-4px)] rounded-full bg-white shadow-sm transition-transform duration-300 ease-in-out pointer-events-none"
            :class="filterOnlyMe ? 'translate-x-full' : 'translate-x-0'"
          />

          <!-- Left Button: Semua Surat -->
          <button
            type="button"
            @click="filterOnlyMe = false"
            title="Semua Surat"
            class="relative z-10 w-1/2 h-full flex items-center justify-center rounded-full transition-colors duration-200 cursor-pointer focus:outline-none"
            :class="!filterOnlyMe ? 'text-[#0A84DC]' : 'text-zinc-400 hover:text-zinc-600'"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          </button>

          <!-- Right Button: Hanya Surat Saya -->
          <button
            type="button"
            @click="filterOnlyMe = true"
            title="Hanya Surat Saya"
            class="relative z-10 w-1/2 h-full flex items-center justify-center rounded-full transition-colors duration-200 cursor-pointer focus:outline-none"
            :class="filterOnlyMe ? 'text-[#0A84DC]' : 'text-zinc-400 hover:text-zinc-600'"
          >
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </button>
        </div>

        <!-- Clear Filter -->
        <button
          type="button"
          @click="clearFilters"
          :disabled="!isFilterActive"
          class="px-3 py-1.5 text-xs font-semibold rounded-xl border transition-all flex items-center gap-1 active:scale-95 disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer"
          :class="isFilterActive ? 'bg-rose-50 text-rose-600 hover:bg-rose-100 border-rose-200/80 shadow-2xs' : 'bg-black/[0.03] text-zinc-400 border-black/5'"
        >
          <svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
          Clear Filter
        </button>
      </div>
    </div>
  </div>
</template>
