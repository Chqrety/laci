<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, nextTick } from 'vue'
import { suratService, prokerService, authService, type Surat, type Proker } from '@/services/api'
import CardDetail from '@/components/CardDetail.vue'

const suratList = ref<Surat[]>([])
const prokerList = ref<Proker[]>([])
const isLoading = ref(false)
const errorMessage = ref('')
const selectedSurat = ref<Surat | null>(null)

// Drag and drop state (Desktop & Touch)
const draggingCardId = ref<string | null>(null)
const dragOverColumn = ref<string | null>(null)

// Ambient Edge Dropzone state
const ambientDrop = ref<{
  active: boolean
  direction: 'left' | 'right' | null
  targetStatus: string | null
}>({
  active: false,
  direction: null,
  targetStatus: null,
})

const resetAmbientDrop = () => {
  if (ambientDrop.value.active) {
    ambientDrop.value = {
      active: false,
      direction: null,
      targetStatus: null,
    }
  }
}

// Modal tambah surat baru
const isCreateModalOpen = ref(false)
const isCreating = ref(false)

const romanMonths = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X', 'XI', 'XII']

// Komponen Nomor Surat terpisah
const noUrut = ref('001')
const selectedProker = ref('')
const selectedBulan = ref(romanMonths[new Date().getMonth()] || 'X')
const selectedTahun = ref(new Date().getFullYear().toString())

const createForm = ref({
  nomor_surat: '',
  perihal: '',
  status_saat_ini: 'Standby',
  pic_nama: '',
  catatan: '',
})

const userRole = computed(() => authService.getRole())
const isSekre = computed(() => userRole.value === 'sekre')
const isHumas = computed(() => userRole.value === 'humas')
const canCreateSurat = computed(() => isSekre.value || isHumas.value)

// Status columns strictly defined
const statusColumns = [
  'Standby',
  'Cetak',
  'TTD Lapis 1',
  'TTD Ketum',
  'TTD Pembina',
  'Paraf Koormawa',
  'TTD Tertinggi',
  'Selesai',
]

const fetchSurat = async () => {
  isLoading.value = true
  errorMessage.value = ''
  try {
    const data = await suratService.getAll()
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

const fetchProker = async () => {
  try {
    prokerList.value = await prokerService.getAll()
  } catch (err: any) {
    console.error('Gagal memuat daftar proker:', err)
  }
}

const getSuratByStatus = (status: string) => {
  return suratList.value.filter((s) => s.status_saat_ini.toLowerCase() === status.toLowerCase())
}

const openDetail = (surat: Surat) => {
  if (isDragging.value) return
  selectedSurat.value = surat
}

const closeDetail = () => {
  selectedSurat.value = null
}

const scrollToColumn = (status: string) => {
  nextTick(() => {
    const colEl = document.querySelector(`[data-status="${status}"]`) as HTMLElement | null
    if (colEl) {
      colEl.scrollIntoView({ behavior: 'smooth', block: 'nearest', inline: 'center' })
    }
  })
}

const checkEdgeDetection = (clientX: number) => {
  if (!draggingCardId.value) {
    resetAmbientDrop()
    return
  }

  const card = suratList.value.find((s) => s.id === draggingCardId.value)
  if (!card) {
    resetAmbientDrop()
    return
  }

  const currentIndex = statusColumns.findIndex(
    (col) => col.toLowerCase() === card.status_saat_ini.toLowerCase()
  )
  if (currentIndex === -1) {
    resetAmbientDrop()
    return
  }

  const screenWidth = window.innerWidth
  const edgeThreshold = 55

  if (clientX < edgeThreshold) {
    if (currentIndex > 0) {
      ambientDrop.value = {
        active: true,
        direction: 'left',
        targetStatus: statusColumns[currentIndex - 1] ?? null,
      }
      return
    }
  } else if (clientX > screenWidth - edgeThreshold) {
    if (currentIndex < statusColumns.length - 1) {
      ambientDrop.value = {
        active: true,
        direction: 'right',
        targetStatus: statusColumns[currentIndex + 1] ?? null,
      }
      return
    }
  }

  resetAmbientDrop()
}

// Core status update function
const executeStatusUpdate = async (suratId: string, targetStatus: string) => {
  const surat = suratList.value.find((s) => s.id === suratId)
  if (!surat) return
  if (surat.status_saat_ini.toLowerCase() === targetStatus.toLowerCase()) return

  const previousStatus = surat.status_saat_ini
  surat.status_saat_ini = targetStatus

  try {
    await suratService.updateStatus(suratId, {
      status_saat_ini: targetStatus,
    })
    await fetchSurat()
  } catch (err: any) {
    surat.status_saat_ini = previousStatus
    alert(err.response?.data?.error || err.message || 'Gagal memindahkan status surat')
  }
}

// Native HTML5 Desktop Drag and Drop Handlers
const onDragStart = (event: DragEvent, suratId: string) => {
  draggingCardId.value = suratId
  if (event.dataTransfer) {
    event.dataTransfer.setData('text/plain', suratId)
    event.dataTransfer.effectAllowed = 'move'
  }
}

const onDrag = (event: DragEvent) => {
  if (event.clientX === 0 && event.clientY === 0) return
  checkEdgeDetection(event.clientX)
}

const onDragEnd = async () => {
  if (ambientDrop.value.active && ambientDrop.value.targetStatus && draggingCardId.value) {
    const cardId = draggingCardId.value
    const targetStatus = ambientDrop.value.targetStatus
    resetAmbientDrop()
    draggingCardId.value = null
    dragOverColumn.value = null
    await executeStatusUpdate(cardId, targetStatus)
    scrollToColumn(targetStatus)
    return
  }
  resetAmbientDrop()
  draggingCardId.value = null
  dragOverColumn.value = null
}

const onDragOver = (event: DragEvent, status: string) => {
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move'
  }
  dragOverColumn.value = status
  checkEdgeDetection(event.clientX)
}

const onDragEnter = (status: string) => {
  dragOverColumn.value = status
}

const onDragLeave = (status: string, event: DragEvent) => {
  const currentTarget = event.currentTarget as HTMLElement
  const relatedTarget = event.relatedTarget as HTMLElement | null
  if (!relatedTarget || !currentTarget.contains(relatedTarget)) {
    if (dragOverColumn.value === status) {
      dragOverColumn.value = null
    }
  }
}

const onDrop = async (event: DragEvent, targetStatus: string) => {
  dragOverColumn.value = null
  const suratId = event.dataTransfer?.getData('text/plain') || draggingCardId.value
  const isAmbient = ambientDrop.value.active && ambientDrop.value.targetStatus
  const finalTarget = isAmbient ? ambientDrop.value.targetStatus! : targetStatus

  resetAmbientDrop()
  draggingCardId.value = null
  if (!suratId) return

  await executeStatusUpdate(suratId, finalTarget)
  if (isAmbient) {
    scrollToColumn(finalTarget)
  }
}

// Mobile Touch Drag State (Long Press to Drag & Ghost Card)
const isDragging = ref(false)
let dragTimer: ReturnType<typeof setTimeout> | null = null
let touchGhostElement: HTMLElement | null = null
let touchOffsetX = 0
let touchOffsetY = 0
let touchSuratId: string | null = null
let touchCardElement: HTMLElement | null = null

const cleanupGhostElement = () => {
  if (touchGhostElement) {
    touchGhostElement.remove()
    touchGhostElement = null
  }
}

onBeforeUnmount(() => {
  if (dragTimer) clearTimeout(dragTimer)
  cleanupGhostElement()
})

const handleTouchStart = (event: TouchEvent, suratId: string) => {
  if (dragTimer) {
    clearTimeout(dragTimer)
    dragTimer = null
  }
  cleanupGhostElement()
  isDragging.value = false

  const touch = event.touches[0]
  if (!touch) return

  const target = (event.currentTarget || event.target) as HTMLElement
  const cardElement = (target.closest('article') || target) as HTMLElement
  const rect = cardElement.getBoundingClientRect()

  touchOffsetX = touch.clientX - rect.left
  touchOffsetY = touch.clientY - rect.top
  touchSuratId = suratId
  touchCardElement = cardElement

  // Mulai timer 300ms untuk Long Press to Drag
  dragTimer = setTimeout(() => {
    isDragging.value = true
    draggingCardId.value = suratId

    // Haptic feedback (getar) jika didukung perangkat
    if (typeof navigator !== 'undefined' && typeof navigator.vibrate === 'function') {
      try {
        navigator.vibrate(50)
      } catch (_) {}
    }

    // Ciptakan ghost element melayang setelah 300ms
    if (touchCardElement) {
      const clone = touchCardElement.cloneNode(true) as HTMLElement
      clone.id = 'touch-drag-ghost'
      clone.classList.add('will-change-transform', 'transform-gpu')
      clone.style.position = 'fixed'
      clone.style.left = `${touch.clientX - touchOffsetX}px`
      clone.style.top = `${touch.clientY - touchOffsetY}px`
      clone.style.width = `${rect.width}px`
      clone.style.zIndex = '9999'
      clone.style.opacity = '0.85'
      clone.style.pointerEvents = 'none'
      clone.style.transform = 'scale(1.04) rotate(1.5deg)'
      clone.style.willChange = 'transform, left, top'
      clone.style.boxShadow = '0 20px 25px -5px rgba(0, 0, 0, 0.25), 0 8px 10px -6px rgba(0, 0, 0, 0.25)'
      clone.style.transition = 'transform 0.1s ease, box-shadow 0.1s ease'
      document.body.appendChild(clone)
      touchGhostElement = clone
    }
  }, 300)
}

const handleTouchMove = (event: TouchEvent) => {
  // Jika jari bergerak SEBELUM 300ms (isDragging masih false), batalkan timer agar scroll vertikal berjalan natural
  if (!isDragging.value) {
    if (dragTimer) {
      clearTimeout(dragTimer)
      dragTimer = null
    }
    return
  }

  // Jika jari bergerak SETELAH timer selesai (isDragging true), blokir scroll dan update ghost element
  if (event.cancelable) {
    event.preventDefault()
  }

  const touch = event.touches[0]
  if (!touch) return

  // Update koordinat ghost element agar mengikuti posisi jari
  if (touchGhostElement) {
    touchGhostElement.style.left = `${touch.clientX - touchOffsetX}px`
    touchGhostElement.style.top = `${touch.clientY - touchOffsetY}px`
  }

  checkEdgeDetection(touch.clientX)

  // Deteksi kolom status di bawah jari
  const elementUnderTouch = document.elementFromPoint(touch.clientX, touch.clientY)
  const columnElement = elementUnderTouch?.closest('[data-status]')
  if (columnElement) {
    const status = columnElement.getAttribute('data-status')
    if (status) {
      dragOverColumn.value = status
    }
  } else {
    dragOverColumn.value = null
  }
}

const handleTouchEnd = async (event?: TouchEvent) => {
  if (dragTimer) {
    clearTimeout(dragTimer)
    dragTimer = null
  }

  const wasDragging = isDragging.value
  const cardId = draggingCardId.value || touchSuratId

  cleanupGhostElement()

  if (wasDragging && cardId) {
    const isAmbient = ambientDrop.value.active && ambientDrop.value.targetStatus

    if (isAmbient) {
      const targetStatus = ambientDrop.value.targetStatus!
      resetAmbientDrop()
      draggingCardId.value = null
      dragOverColumn.value = null
      isDragging.value = false
      touchSuratId = null
      touchCardElement = null

      await executeStatusUpdate(cardId, targetStatus)
      scrollToColumn(targetStatus)
      return
    }

    let targetCol = dragOverColumn.value
    if (!targetCol && event && event.changedTouches && event.changedTouches[0]) {
      const endTouch = event.changedTouches[0]
      const el = document.elementFromPoint(endTouch.clientX, endTouch.clientY)
      const colEl = el?.closest('[data-status]')
      if (colEl) {
        targetCol = colEl.getAttribute('data-status')
      }
    }

    if (targetCol) {
      const targetStatus = targetCol
      resetAmbientDrop()
      dragOverColumn.value = null
      draggingCardId.value = null
      isDragging.value = false
      touchSuratId = null
      touchCardElement = null

      await executeStatusUpdate(cardId, targetStatus)
      scrollToColumn(targetStatus)
      return
    }
  }

  resetAmbientDrop()
  draggingCardId.value = null
  dragOverColumn.value = null
  isDragging.value = false
  touchSuratId = null
  touchCardElement = null
}

const openCreateModal = () => {
  noUrut.value = '001'
  const firstProker = prokerList.value[0]
  selectedProker.value = firstProker ? firstProker.nama_proker : ''
  const monthIdx = new Date().getMonth()
  selectedBulan.value = (monthIdx >= 0 && monthIdx < romanMonths.length && romanMonths[monthIdx]) ? romanMonths[monthIdx]! : 'X'
  selectedTahun.value = new Date().getFullYear().toString()

  createForm.value = {
    nomor_surat: '',
    perihal: '',
    status_saat_ini: 'Standby',
    pic_nama: '',
    catatan: '',
  }
  isCreateModalOpen.value = true
}

const handleCreateSurat = async () => {
  if (!noUrut.value.trim()) {
    alert('Nomor urut surat wajib diisi.')
    return
  }
  if (!selectedProker.value) {
    alert('Silakan pilih Program Kerja (Proker).')
    return
  }
  if (!selectedTahun.value.trim()) {
    alert('Tahun surat wajib diisi.')
    return
  }

  const prokerCode = selectedProker.value.toUpperCase().replace(/\s+/g, '-')
  const paddedNo = noUrut.value.trim().padStart(3, '0')
  createForm.value.nomor_surat = `${paddedNo}/DOSCOM/${prokerCode}/${selectedBulan.value}/${selectedTahun.value.trim()}`

  isCreating.value = true
  try {
    await suratService.create(createForm.value)
    isCreateModalOpen.value = false
    await fetchSurat()
  } catch (err: any) {
    alert(err.response?.data?.error || 'Gagal membuat surat')
  } finally {
    isCreating.value = false
  }
}

onMounted(() => {
  fetchSurat()
  fetchProker()
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
          @click="fetchSurat"
          :disabled="isLoading"
          class="flex-1 sm:flex-initial px-4 py-2 text-xs font-semibold rounded-2xl bg-white/80 border border-black/5 text-zinc-700 hover:bg-white shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all active:scale-[0.98] disabled:opacity-50"
        >
          {{ isLoading ? 'Memuat...' : 'Muat Ulang' }}
        </button>

        <button
          type="button"
          @click="openCreateModal"
          class="flex-1 sm:flex-initial px-4 py-2 text-xs font-semibold rounded-2xl bg-[#0A84DC] text-white hover:bg-[#0872be] shadow-[0_8px_30px_rgb(0,0,0,0.05)] transition-all active:scale-[0.98]"
        >
          + Tambah Surat
        </button>
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
          <article
            v-for="item in getSuratByStatus(status)"
            :key="item.id"
            v-memo="[item.id, item.nomor_surat, item.perihal, item.status_saat_ini, item.pic_nama, item.arsip_url, item.catatan, draggingCardId === item.id]"
            draggable="true"
            @dragstart="onDragStart($event, item.id)"
            @drag="onDrag"
            @dragend="onDragEnd"
            @dragover.prevent="onDragOver($event, status)"
            @drop.prevent="onDrop($event, status)"
            @touchstart="handleTouchStart($event, item.id)"
            @touchmove="handleTouchMove"
            @touchend="handleTouchEnd"
            @touchcancel="handleTouchEnd"
            @contextmenu.prevent
            @click="openDetail(item)"
            :class="[
              'touch-pan-y select-none bg-white p-4 rounded-2xl border border-black/5 shadow-[0_8px_30px_rgb(0,0,0,0.05)] hover:border-[#0A84DC]/30 hover:shadow-md cursor-grab active:cursor-grabbing transition-all space-y-2.5 active:scale-[0.98]',
              draggingCardId === item.id ? 'opacity-40 ring-2 ring-[#0A84DC]/50' : ''
            ]"
          >
            <div class="flex items-start justify-between gap-1">
              <span class="font-mono text-[11px] font-semibold text-zinc-700 bg-black/[0.04] px-2 py-0.5 rounded-lg">
                {{ item.nomor_surat }}
              </span>
              <span
                v-if="item.arsip_url"
                title="Arsip telah dipindai"
                class="text-[10px] text-[#0A84DC] bg-[#0A84DC]/10 border border-[#0A84DC]/20 px-2 py-0.5 rounded-lg font-semibold"
              >
                Arsip
              </span>
            </div>

            <h4 class="text-xs font-semibold text-zinc-900 leading-snug line-clamp-2">
              {{ item.perihal }}
            </h4>

            <div class="flex items-center justify-between pt-1 border-t border-black/5 text-[11px]">
              <span class="inline-flex items-center gap-1.5 text-zinc-500 font-medium">
                <span class="w-1.5 h-1.5 rounded-full bg-[#0A84DC]"></span>
                {{ item.pic_nama }}
              </span>

              <span v-if="item.catatan" title="Terdapat catatan" class="text-zinc-400 text-xs">
                💬
              </span>
            </div>
          </article>
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
        @updated="fetchSurat"
      />
    </Transition>

    <!-- Modal Create Surat (iOS HIG: rounded-3xl, Inset Grouped, Glassmorphism, Animated) -->
    <Transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-150 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <div
        v-if="isCreateModalOpen"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4 overflow-y-auto"
      >
        <div class="bg-white rounded-3xl border border-black/10 max-w-xl w-full p-6 sm:p-7 space-y-5 shadow-xl">
          <header>
            <h3 class="text-lg font-semibold text-zinc-900 tracking-tight">Tambah Surat Baru</h3>
            <p class="text-xs text-zinc-500 mt-0.5">Lengkapi parameter nomor surat dan informasi dokumen resmi.</p>
          </header>

          <form @submit.prevent="handleCreateSurat" class="space-y-4">
            <!-- Inset Grouped Form Container -->
            <div class="rounded-2xl border border-black/10 bg-white overflow-hidden divide-y divide-black/5 shadow-2xs">
              <!-- 1. Nomor Surat Group -->
              <div class="p-3.5 space-y-2 focus-within:bg-zinc-50/50 transition-colors">
                <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Nomor Surat</label>
                <div class="flex flex-row flex-wrap items-center gap-2">
                  <input
                    v-model="noUrut"
                    type="text"
                    required
                    placeholder="001"
                    class="w-16 sm:w-20 px-2 py-1.5 text-xs font-mono font-semibold text-center border border-black/10 rounded-xl bg-black/[0.02] focus:outline-none focus:ring-1 focus:ring-[#0A84DC]"
                  />

                  <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

                  <input
                    type="text"
                    value="DOSCOM"
                    disabled
                    class="w-20 sm:w-24 px-2 py-1.5 text-xs font-mono font-semibold text-center border border-black/10 rounded-xl bg-black/[0.04] text-zinc-600 cursor-not-allowed select-none"
                  />

                  <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

                  <select
                    v-model="selectedProker"
                    required
                    class="flex-1 min-w-[120px] px-2.5 py-1.5 text-xs font-medium border border-black/10 rounded-xl bg-white focus:outline-none focus:ring-1 focus:ring-[#0A84DC] truncate"
                  >
                    <option value="" disabled>-- Pilih Proker --</option>
                    <option v-for="proker in prokerList" :key="proker.id" :value="proker.nama_proker">
                      {{ proker.nama_proker }}
                    </option>
                  </select>

                  <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

                  <select
                    v-model="selectedBulan"
                    required
                    class="w-16 sm:w-20 px-2 py-1.5 text-xs font-mono font-semibold text-center border border-black/10 rounded-xl bg-white focus:outline-none focus:ring-1 focus:ring-[#0A84DC]"
                  >
                    <option v-for="bulan in romanMonths" :key="bulan" :value="bulan">
                      {{ bulan }}
                    </option>
                  </select>

                  <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

                  <input
                    v-model="selectedTahun"
                    type="number"
                    required
                    placeholder="2026"
                    class="min-w-[4rem] flex-1 sm:w-20 px-2 py-1.5 text-xs font-mono font-semibold text-center border border-black/10 rounded-xl bg-white focus:outline-none focus:ring-1 focus:ring-[#0A84DC]"
                  />
                </div>

                <p class="text-[11px] text-zinc-500">
                  Pratinjau: <span class="font-mono text-[#0A84DC] font-semibold">{{ (noUrut || '001') }}/DOSCOM/{{ (selectedProker || 'PROKER').toUpperCase().replace(/\s+/g, '-') }}/{{ selectedBulan }}/{{ selectedTahun }}</span>
                </p>
              </div>

              <!-- 2. Perihal -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Perihal</label>
                <input
                  v-model="createForm.perihal"
                  type="text"
                  required
                  placeholder="Undangan Kerjasama Dies Natalis"
                  class="w-full text-xs font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 pt-0.5"
                />
              </div>

              <!-- 3. Nama PIC -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Nama PIC</label>
                <input
                  v-model="createForm.pic_nama"
                  type="text"
                  required
                  placeholder="Fikri"
                  class="w-full text-xs font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 pt-0.5"
                />
              </div>

              <!-- 4. Status Awal -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Status Awal</label>
                <select
                  v-if="!isHumas"
                  v-model="createForm.status_saat_ini"
                  class="w-full text-xs font-semibold text-zinc-900 bg-transparent focus:outline-none pt-0.5"
                >
                  <option v-for="col in statusColumns" :key="col" :value="col">
                    {{ col }}
                  </option>
                </select>
                <div v-else class="text-xs font-semibold text-zinc-700 pt-0.5">
                  Standby (Otomatis)
                </div>
              </div>

              <!-- 5. Catatan Awal -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <label class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Catatan Awal (Opsional)</label>
                <textarea
                  v-model="createForm.catatan"
                  rows="2"
                  placeholder="Keterangan instruksi..."
                  class="w-full text-xs font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 resize-none pt-0.5"
                ></textarea>
              </div>
            </div>

            <footer class="flex justify-end gap-2.5 pt-2">
              <button
                type="button"
                @click="isCreateModalOpen = false"
                class="px-4 py-2 text-xs font-semibold border border-black/10 rounded-2xl text-zinc-700 hover:bg-black/5 transition-all duration-200 active:scale-95"
              >
                Batal
              </button>
              <button
                type="submit"
                :disabled="isCreating"
                class="px-5 py-2 text-xs font-semibold bg-[#0A84DC] text-white rounded-2xl hover:bg-[#0872be] shadow-md disabled:opacity-50 transition-all duration-200 active:scale-95"
              >
                {{ isCreating ? 'Menyimpan...' : 'Simpan Surat' }}
              </button>
            </footer>
          </form>
        </div>
      </div>
    </Transition>
  </div>
</template>
