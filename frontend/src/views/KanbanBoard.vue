<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  fetchSurat as apiFetchSurat,
  fetchProker as apiFetchProker,
  updateSuratStatus,
  authService,
  STATUS_COLUMNS,
  type Surat,
  type Proker,
} from '@/services/api'
import { useKanbanDragDrop } from '@/composables/useKanbanDragDrop'
import CardDetail from '@/components/CardDetail.vue'
import KanbanCard from '@/components/KanbanCard.vue'
import KanbanFilter from '@/components/KanbanFilter.vue'
import KanbanModalForm from '@/components/KanbanModalForm.vue'

const suratList = ref<Surat[]>([])
const filteredSuratList = ref<Surat[]>([])
const prokerList = ref<Proker[]>([])
const isLoading = ref(false)
const errorMessage = ref('')
const selectedSurat = ref<Surat | null>(null)

// Modal tambah surat baru
const isCreateModalOpen = ref(false)

const userRole = computed(() => authService.getRole())
const isSekre = computed(() => userRole.value === 'sekre')
const isHumas = computed(() => userRole.value === 'humas')
const canCreateSurat = computed(() => isSekre.value || isHumas.value)

// Status columns array defined from shared constant
const statusColumns = [...STATUS_COLUMNS]

const loadSurat = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const data = await apiFetchSurat()
    suratList.value = data

    if (selectedSurat.value) {
      const updated = data.find((s) => s.id === selectedSurat.value?.id)
      if (updated) {
        selectedSurat.value = updated
      }
    }
  } catch (err: any) {
    errorMessage.value = err.response?.data?.error || err.message || 'Gagal memuat data surat'
  } finally {
    isLoading.value = false
  }
}

const loadProker = async () => {
  try {
    prokerList.value = await apiFetchProker()
  } catch (err: any) {
    console.error('Gagal memuat daftar proker:', err)
  }
}

const getSuratByStatus = (status: string) => {
  return filteredSuratList.value.filter((s) => s.status_saat_ini.toLowerCase() === status.toLowerCase())
}

const openDetail = (surat: Surat) => {
  if (isDragging.value) return
  selectedSurat.value = surat
}

const closeDetail = () => {
  selectedSurat.value = null
}

// Core status update function
const executeStatusUpdate = async (suratId: string, targetStatus: string) => {
  const surat = suratList.value.find((s) => s.id === suratId)
  if (!surat) return
  if (surat.status_saat_ini.toLowerCase() === targetStatus.toLowerCase()) return

  const previousStatus = surat.status_saat_ini
  surat.status_saat_ini = targetStatus
  filteredSuratList.value = [...filteredSuratList.value]

  try {
    await updateSuratStatus(suratId, {
      status_saat_ini: targetStatus,
    })
    await loadSurat()
  } catch (err: any) {
    surat.status_saat_ini = previousStatus
    filteredSuratList.value = [...filteredSuratList.value]
    alert(err.response?.data?.error || err.message || 'Gagal memindahkan status surat')
  }
}

// Drag & Drop Composable (Desktop DnD, Mobile Touch, & Ambient Edge Zone)
const {
  draggingCardId,
  dragOverColumn,
  ambientDrop,
  isDragging,
  onDragStart,
  onDrag,
  onDragEnd,
  onDragOver,
  onDragEnter,
  onDragLeave,
  onDrop,
  handleTouchStart,
  handleTouchMove,
  handleTouchEnd,
} = useKanbanDragDrop(suratList, statusColumns, executeStatusUpdate)

onMounted(() => {
  loadSurat()
  loadProker()
})
</script>

<template>
  <div class="h-full flex flex-col space-y-4">
    <!-- Top Bar Controls -->
    <header class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 px-1">
      <div>
        <h2 class="text-lg font-semibold tracking-tight text-zinc-900">Alur Tracking Surat</h2>
        <p class="text-xs text-zinc-500 mt-0.5">Tarik dan lepas kartu atau geser horisontal untuk mengelola status surat.</p>
      </div>

      <div class="flex items-center gap-2 w-full sm:w-auto">
        <button
          type="button"
          @click="loadSurat"
          :disabled="isLoading"
          class="flex-1 sm:flex-initial px-4 py-2 text-xs font-semibold rounded-2xl bg-white/80 border border-black/5 text-zinc-700 hover:bg-white shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all active:scale-[0.98] disabled:opacity-50"
        >
          {{ isLoading ? 'Memuat...' : 'Muat Ulang' }}
        </button>

        <button
          type="button"
          @click="isCreateModalOpen = true"
          class="flex-1 sm:flex-initial px-4 py-2 text-xs font-semibold rounded-2xl bg-[#0A84DC] text-white hover:bg-[#0872be] shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all active:scale-[0.98] cursor-pointer"
        >
          + Tambah Surat
        </button>
      </div>
    </header>

    <!-- Advanced Multi-Filter Bar -->
    <KanbanFilter
      :surat-list="suratList"
      :proker-list="prokerList"
      @update:filtered="filteredSuratList = $event"
    />

    <!-- Error Notice -->
    <div
      v-if="errorMessage"
      role="alert"
      class="p-3 text-xs bg-rose-50 border border-rose-200 text-rose-700 rounded-2xl font-medium"
    >
      {{ errorMessage }}
    </div>

    <!-- Edge Ambient Dropzone Visual Overlay -->
    <Transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="ambientDrop.active"
        class="fixed inset-y-0 z-40 pointer-events-none flex items-center select-none"
        :class="ambientDrop.direction === 'left' ? 'left-0 justify-start' : 'right-0 justify-end'"
      >
        <!-- Soft Gradient Ambient Zone -->
        <div
          class="h-full w-28 sm:w-36 flex items-center transition-all duration-300"
          :class="[
            ambientDrop.direction === 'left'
              ? 'bg-gradient-to-r from-[#0A84DC]/40 via-[#0A84DC]/15 to-transparent pl-4 justify-start'
              : 'bg-gradient-to-l from-[#0A84DC]/40 via-[#0A84DC]/15 to-transparent pr-4 justify-end'
          ]"
        >
          <!-- Destination Status Pill (Miring & Elegan) -->
          <div
            class="flex items-center gap-1.5 px-3 py-2 rounded-2xl bg-white/95 border border-[#0A84DC]/30 text-[#0A84DC] shadow-xl transform transition-transform duration-200"
            :class="ambientDrop.direction === 'left' ? '-rotate-6 translate-x-1' : 'rotate-6 -translate-x-1'"
          >
            <svg
              v-if="ambientDrop.direction === 'left'"
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4 animate-pulse shrink-0"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 19l-7-7 7-7" />
            </svg>

            <span class="text-xs font-bold tracking-tight whitespace-nowrap">
              Ke: {{ ambientDrop.targetStatus }}
            </span>

            <svg
              v-if="ambientDrop.direction === 'right'"
              xmlns="http://www.w3.org/2000/svg"
              class="w-4 h-4 animate-pulse shrink-0"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Mobile-First Horizontal Scrolling Kanban Container -->
    <main
      class="flex-1 flex flex-nowrap gap-4 overflow-x-auto snap-x snap-mandatory no-scrollbar pb-6 pt-1 items-start scroll-smooth"
      style="-webkit-overflow-scrolling: touch;"
    >
      <!-- Column Status Section (iOS rounded-3xl, glassmorphic) -->
      <section
        v-for="status in statusColumns"
        :key="status"
        :data-status="status"
        @dragover.prevent="onDragOver($event, status)"
        @dragenter.prevent="onDragEnter(status)"
        @dragleave="onDragLeave(status, $event)"
        @drop="onDrop($event, status)"
        :class="[
          'w-full min-w-full sm:w-[320px] sm:min-w-[320px] shrink-0 snap-center flex flex-col rounded-3xl border transition-all duration-200 max-h-[calc(100vh-12rem)] shadow-[0_8px_30px_rgb(0,0,0,0.05)]',
          dragOverColumn === status
            ? 'bg-[#0A84DC]/10 border-[#0A84DC]/40 ring-2 ring-[#0A84DC]/30'
            : 'bg-white/70 backdrop-blur-md border-black/5'
        ]"
      >
        <!-- Column Header -->
        <header class="p-4 flex items-center justify-between border-b border-black/5 bg-white/40 rounded-t-3xl select-none">
          <div class="flex items-center gap-2">
            <h3 class="text-xs font-semibold text-zinc-900 tracking-tight">{{ status }}</h3>
            <span class="text-[11px] font-mono text-zinc-500 bg-white/90 border border-black/5 px-2 py-0.5 rounded-xl font-semibold shadow-2xs">
              {{ getSuratByStatus(status).length }}
            </span>
          </div>
        </header>

        <!-- Column Card List -->
        <div class="flex-1 p-3 space-y-3 overflow-y-auto min-h-[130px]">
          <!-- Drop Placeholder Indicator -->
          <div
            v-if="dragOverColumn === status && draggingCardId"
            class="h-16 border-2 border-dashed border-[#0A84DC]/40 rounded-2xl bg-[#0A84DC]/5 flex items-center justify-center text-xs text-[#0A84DC] font-semibold transition-all"
          >
            Lepas kartu di sini
          </div>

          <div
            v-if="getSuratByStatus(status).length === 0 && dragOverColumn !== status"
            class="h-24 border border-dashed border-black/10 rounded-2xl flex items-center justify-center text-[11px] text-zinc-400"
          >
            Kosong
          </div>

          <!-- Surat Card (Apple iOS: rounded-2xl, subtle shadow, touch-supported) -->
          <KanbanCard
            v-for="item in getSuratByStatus(status)"
            :key="item.id"
            :surat="item"
            :is-dragging="draggingCardId === item.id"
            @dragstart="onDragStart($event, item.id)"
            @drag="onDrag"
            @dragend="onDragEnd"
            @dragover.prevent="onDragOver($event, status)"
            @drop.prevent="onDrop($event, status)"
            @touchstart="handleTouchStart($event, item.id)"
            @touchmove="handleTouchMove"
            @touchend="handleTouchEnd"
            @touchcancel="handleTouchEnd"
            @contextmenu.prevent="() => {}"
            @click="openDetail(item)"
          />
        </div>
      </section>
    </main>

    <!-- Slide-over Card Detail with Transition -->
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      enter-from-class="opacity-0"
      enter-to-class="opacity-100"
      leave-active-class="transition-opacity duration-150 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <CardDetail
        v-if="selectedSurat"
        :surat="selectedSurat"
        :status-options="statusColumns"
        @close="closeDetail"
        @updated="loadSurat"
      />
    </Transition>

    <!-- Modal Create Surat -->
    <KanbanModalForm
      :is-open="isCreateModalOpen"
      :proker-list="prokerList"
      :status-columns="statusColumns"
      :is-humas="isHumas"
      @close="isCreateModalOpen = false"
      @created="loadSurat"
    />
  </div>
</template>
