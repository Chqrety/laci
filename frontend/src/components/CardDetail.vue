<script setup lang="ts">
import { ref, computed, nextTick, onBeforeUnmount } from 'vue'
import { api, API_BASE_URL, suratService, authService, type Surat } from '@/services/api'

const props = defineProps<{
  surat: Surat
  statusOptions: string[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'updated'): void
}>()

const userRole = computed(() => authService.getRole())
const isSekre = computed(() => userRole.value === 'sekre')

// Status & Note update state
const selectedStatus = ref(props.surat.status_saat_ini)
const newNote = ref(props.surat.catatan || '')
const isUpdatingStatus = ref(false)
const updateMessage = ref('')

// In-App Web Camera state & upload
const isCameraOpen = ref(false)
const useNativeCamera = ref(typeof window !== 'undefined' && window.isSecureContext === false)
const videoRef = ref<HTMLVideoElement | null>(null)
const mediaStream = ref<MediaStream | null>(null)
const isCapturing = ref(false)
const cameraError = ref('')

const isUploadingDokumen = ref(false)
const uploadDokumenMessage = ref('')
const uploadDokumenError = ref('')

const fileArsipUrl = computed(() => {
  return props.surat.file_arsip || props.surat.arsip_url || null
})

const isPdf = computed(() => {
  if (!fileArsipUrl.value) return false
  const parts = fileArsipUrl.value.split('?')
  const cleanUrl = (parts[0] || '').toLowerCase()
  return cleanUrl.endsWith('.pdf')
})

const getFullAssetUrl = (url?: string | null) => {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  return `${API_BASE_URL}${url}`
}

const openCamera = async () => {
  cameraError.value = ''
  uploadDokumenMessage.value = ''
  uploadDokumenError.value = ''

  try {
    if (!navigator?.mediaDevices?.getUserMedia) {
      throw new Error('WebRTC tidak didukung atau memerlukan HTTPS.')
    }
    isCameraOpen.value = true
    const stream = await navigator.mediaDevices.getUserMedia({
      video: {
        facingMode: 'environment',
      },
    })
    mediaStream.value = stream
    await nextTick()
    if (videoRef.value) {
      videoRef.value.srcObject = stream
      await videoRef.value.play().catch(() => {})
    }
  } catch (err: any) {
    console.error('Camera access error:', err)
    useNativeCamera.value = true
    isCameraOpen.value = false
    cameraError.value = err.name === 'NotAllowedError'
      ? 'Izin kamera ditolak. Silakan gunakan kamera bawaan perangkat.'
      : 'Kamera WebRTC tidak dapat diakses di jaringan ini. Silakan gunakan kamera bawaan perangkat.'
  }
}

const stopCamera = () => {
  if (mediaStream.value) {
    mediaStream.value.getTracks().forEach((track) => track.stop())
    mediaStream.value = null
  }
  if (videoRef.value) {
    videoRef.value.srcObject = null
  }
  isCameraOpen.value = false
}

const captureAndSave = async () => {
  if (!videoRef.value || !mediaStream.value) return
  isCapturing.value = true
  cameraError.value = ''

  try {
    const video = videoRef.value
    const canvas = document.createElement('canvas')
    const width = video.videoWidth || 1280
    const height = video.videoHeight || 720
    canvas.width = width
    canvas.height = height

    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('Gagal menyiapkan canvas gambar')
    ctx.drawImage(video, 0, 0, width, height)

    const blob = await new Promise<Blob | null>((resolve) =>
      canvas.toBlob((b) => resolve(b), 'image/jpeg', 0.9)
    )

    if (!blob) throw new Error('Gagal memproses frame kamera')

    const filename = `arsip-${props.surat.id.slice(0, 8)}-${Date.now()}.jpg`
    const formData = new FormData()
    formData.append('file', blob, filename)

    await api.post(`/surat/${props.surat.id}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })

    stopCamera()
    uploadDokumenMessage.value = 'Foto arsip berhasil dijepret & diunggah'
    emit('updated')
  } catch (err: any) {
    console.error('Capture and upload error:', err)
    cameraError.value = err.response?.data?.error || err.message || 'Gagal menyimpan foto arsip'
  } finally {
    isCapturing.value = false
  }
}

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  isUploadingDokumen.value = true
  uploadDokumenMessage.value = ''
  uploadDokumenError.value = ''

  try {
    const formData = new FormData()
    formData.append('file', file)
    await api.post(`/surat/${props.surat.id}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    uploadDokumenMessage.value = 'Arsip dokumen berhasil diunggah'
    emit('updated')
  } catch (err: any) {
    uploadDokumenError.value = err.response?.data?.error || err.message || 'Gagal mengunggah dokumen'
  } finally {
    isUploadingDokumen.value = false
    target.value = ''
  }
}

const handleClosePanel = () => {
  stopCamera()
  emit('close')
}

onBeforeUnmount(() => {
  stopCamera()
})

// Tracking logs limit toggle
const showAllLogs = ref(false)

const reversedLogs = computed(() => {
  if (!props.surat.logs) return []
  return [...props.surat.logs].reverse()
})

const displayedLogs = computed(() => {
  if (showAllLogs.value) {
    return reversedLogs.value
  }
  return reversedLogs.value.slice(0, 5)
})

const handleUpdateStatusAndNote = async () => {
  isUpdatingStatus.value = true
  updateMessage.value = ''
  try {
    await suratService.updateStatus(props.surat.id, {
      status_saat_ini: isSekre.value ? selectedStatus.value : props.surat.status_saat_ini,
      catatan: newNote.value,
      catatan_log: newNote.value,
    })
    updateMessage.value = isSekre.value ? 'Status & catatan berhasil diperbarui' : 'Catatan berhasil diperbarui'
    emit('updated')
  } catch (err: any) {
    alert(err.response?.data?.error || err.message || 'Gagal memperbarui data surat')
  } finally {
    isUpdatingStatus.value = false
  }
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
</script>

<template>
  <aside
    aria-label="Detail Surat"
    class="fixed inset-0 z-50 flex justify-end bg-black/50 transition-opacity"
  >
    <!-- Backdrop close click area -->
    <div class="fixed inset-0" @click="handleClosePanel" aria-hidden="true"></div>

    <!-- iOS Slide-over Sheet (rounded-3xl, pure solid white, no bleed) -->
    <div class="relative w-full max-w-lg bg-white h-full shadow-2xl flex flex-col z-10 border-l border-black/10 overflow-hidden sm:rounded-l-3xl">
      <!-- Header -->
      <header class="p-6 border-b border-black/5 flex items-start justify-between bg-white">
        <div class="space-y-1.5">
          <span class="inline-block px-2.5 py-0.5 text-[11px] font-mono font-semibold bg-[#0A84DC]/10 text-[#0A84DC] border border-[#0A84DC]/20 rounded-xl">
            {{ props.surat.nomor_surat }}
          </span>
          <h2 class="text-base font-semibold text-zinc-900 leading-snug">{{ props.surat.perihal }}</h2>
          <p class="text-xs text-zinc-500">PIC: <span class="text-zinc-800 font-semibold">{{ props.surat.pic_nama }}</span></p>
        </div>

        <button
          type="button"
          @click="handleClosePanel"
          aria-label="Tutup Panel"
          class="p-2 rounded-2xl text-zinc-400 hover:text-zinc-700 hover:bg-black/5 transition-all duration-200 active:scale-95"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </header>

      <!-- Scrollable Body Content -->
      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        <!-- 1. Properties Section (iOS Inset Grouped - Editable for Sekre & Humas) -->
        <section class="space-y-2.5">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-500 px-1">Status & Catatan</h3>

          <form @submit.prevent="handleUpdateStatusAndNote" class="space-y-3">
            <div class="rounded-2xl border border-black/10 bg-white overflow-hidden divide-y divide-black/5 shadow-2xs">
              <!-- Status Input (Dropdown for Sekre, Read-Only for Humas) -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <label for="update-status" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Status Saat Ini</label>
                <select
                  v-if="isSekre"
                  id="update-status"
                  v-model="selectedStatus"
                  class="w-full text-xs font-semibold text-zinc-900 bg-transparent focus:outline-none pt-0.5"
                >
                  <option v-for="opt in props.statusOptions" :key="opt" :value="opt">
                    {{ opt }}
                  </option>
                </select>
                <div v-else class="text-xs font-semibold text-zinc-900 pt-0.5 flex items-center justify-between">
                  <span>{{ props.surat.status_saat_ini }}</span>
                  <span class="text-[10px] text-zinc-400 font-normal italic">Hanya Sekre yang dapat mengubah status</span>
                </div>
              </div>

              <!-- Catatan Dokumen (Always editable for both Sekre and Humas) -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <label for="update-note" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Catatan Dokumen</label>
                <textarea
                  id="update-note"
                  v-model="newNote"
                  rows="2"
                  placeholder="Tambahkan catatan instruksi..."
                  class="w-full text-xs font-medium text-zinc-900 bg-transparent focus:outline-none placeholder:text-zinc-400 resize-none pt-0.5"
                ></textarea>
              </div>
            </div>

            <div class="flex items-center justify-between pt-1">
              <span v-if="updateMessage" class="text-[11px] text-emerald-600 font-semibold">{{ updateMessage }}</span>
              <button
                type="submit"
                :disabled="isUpdatingStatus"
                class="ml-auto px-4 py-2 text-xs font-semibold bg-[#0A84DC] text-white rounded-2xl hover:bg-[#0872be] shadow-md disabled:opacity-50 transition-all duration-200 active:scale-95"
              >
                {{ isUpdatingStatus ? 'Menyimpan...' : (isSekre ? 'Perbarui Status' : 'Simpan Catatan') }}
              </button>
            </div>
          </form>
        </section>

        <!-- 2. Arsip Dokumen Section (In-App Camera WebRTC) -->
        <section class="space-y-3">
          <div class="flex items-center justify-between px-1">
            <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-500">Arsip Dokumen</h3>
            <span
              v-if="fileArsipUrl"
              class="text-[11px] font-semibold text-emerald-600 bg-emerald-50 border border-emerald-200 px-2 py-0.5 rounded-xl"
            >
              Tersedia
            </span>
          </div>

          <!-- In-App Camera View with Smooth Transition -->
          <transition
            enter-active-class="transition duration-300 ease-out"
            enter-from-class="transform scale-95 opacity-0"
            enter-to-class="transform scale-100 opacity-100"
            leave-active-class="transition duration-200 ease-in"
            leave-from-class="transform scale-100 opacity-100"
            leave-to-class="transform scale-95 opacity-0"
          >
            <div
              v-if="!useNativeCamera && isCameraOpen"
              class="rounded-3xl border border-black/10 bg-zinc-950 p-3.5 shadow-xl space-y-3 text-white overflow-hidden"
            >
              <div class="relative w-full aspect-4/3 rounded-2xl overflow-hidden bg-black flex items-center justify-center">
                <video
                  ref="videoRef"
                  autoplay
                  playsinline
                  muted
                  class="w-full h-full object-cover"
                ></video>

                <div v-if="!mediaStream && !cameraError" class="absolute inset-0 flex items-center justify-center bg-black/70 text-xs text-zinc-300">
                  Memulai kamera belakang...
                </div>

                <div v-if="cameraError" class="absolute inset-0 p-4 flex items-center justify-center text-center bg-black/85 text-xs text-rose-400 font-medium">
                  {{ cameraError }}
                </div>
              </div>

              <!-- Camera Controls -->
              <div class="flex items-center justify-between px-1 pt-0.5">
                <button
                  type="button"
                  @click="stopCamera"
                  class="px-3.5 py-2 text-xs font-semibold rounded-xl bg-zinc-800 hover:bg-zinc-700 text-zinc-300 border border-zinc-700 transition-all duration-200 active:scale-95"
                >
                  Tutup
                </button>

                <!-- iOS Round / Rounded-xl Camera Shutter Button -->
                <button
                  type="button"
                  @click="captureAndSave"
                  :disabled="isCapturing || !mediaStream"
                  class="inline-flex items-center gap-2 px-5 py-2 text-xs font-semibold rounded-xl bg-[#0A84DC] hover:bg-[#0872be] text-white shadow-md transition-all duration-200 active:scale-95 disabled:opacity-50"
                >
                  <span v-if="!isCapturing" class="w-3 h-3 rounded-full bg-white border-2 border-[#0A84DC]"></span>
                  <span>{{ isCapturing ? 'Menyimpan...' : 'Jepret & Simpan' }}</span>
                </button>
              </div>
            </div>
          </transition>

          <!-- Fallback Native Camera Button (Saat WebRTC Gagal/HTTP Network) -->
          <div v-if="useNativeCamera" class="flex items-center gap-2">
            <label class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#0A84DC] hover:bg-[#0872be] text-white text-xs font-semibold rounded-xl shadow-md cursor-pointer transition-all duration-200 active:scale-95">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              <span>{{ fileArsipUrl ? 'Buka Kamera (Foto Ulang)' : 'Buka Kamera Perangkat' }}</span>
              <input
                type="file"
                accept="image/*,application/pdf"
                capture="environment"
                class="hidden"
                @change="handleFileUpload"
              />
            </label>

            <!-- File Upload Fallback -->
            <label class="inline-flex items-center gap-1.5 px-3 py-2.5 bg-white hover:bg-zinc-50 text-zinc-700 text-xs font-semibold rounded-xl border border-black/10 cursor-pointer shadow-2xs transition-all duration-200 active:scale-95">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
              </svg>
              <span>Pilih File</span>
              <input
                type="file"
                accept="image/*,application/pdf"
                class="hidden"
                @change="handleFileUpload"
              />
            </label>
          </div>

          <!-- Buka Kamera Button (When Camera Closed) -->
          <div v-else-if="!isCameraOpen" class="flex items-center gap-2">
            <button
              type="button"
              @click="openCamera"
              class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#0A84DC] hover:bg-[#0872be] text-white text-xs font-semibold rounded-xl shadow-md transition-all duration-200 active:scale-95"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              <span>{{ fileArsipUrl ? 'Buka Kamera (Foto Ulang)' : 'Buka Kamera' }}</span>
            </button>

            <!-- File Upload Fallback -->
            <label class="inline-flex items-center gap-1.5 px-3 py-2.5 bg-white hover:bg-zinc-50 text-zinc-700 text-xs font-semibold rounded-xl border border-black/10 cursor-pointer shadow-2xs transition-all duration-200 active:scale-95">
              <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
              </svg>
              <span>Pilih File</span>
              <input
                type="file"
                accept="image/*,application/pdf"
                class="hidden"
                @change="handleFileUpload"
              />
            </label>
          </div>

          <!-- Status feedback messages -->
          <p v-if="cameraError && useNativeCamera" class="text-[11px] text-amber-600 font-medium px-1">
            {{ cameraError }}
          </p>
          <p v-if="uploadDokumenMessage" class="text-[11px] text-emerald-600 font-semibold px-1">
            {{ uploadDokumenMessage }}
          </p>
          <p v-if="uploadDokumenError" class="text-[11px] text-rose-600 font-medium px-1">
            {{ uploadDokumenError }}
          </p>

          <!-- Pratinjau Dokumen jika ada -->
          <div v-if="fileArsipUrl && (!isCameraOpen || useNativeCamera)" class="space-y-2">
            <!-- PDF Document Button -->
            <div v-if="isPdf" class="rounded-2xl border border-black/10 bg-white p-3.5 shadow-2xs flex items-center justify-between gap-3">
              <div class="flex items-center gap-3 overflow-hidden">
                <div class="w-10 h-10 rounded-xl bg-rose-50 border border-rose-200 flex items-center justify-center text-rose-600 shrink-0">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                  </svg>
                </div>
                <div class="truncate">
                  <p class="text-xs font-semibold text-zinc-900 truncate">Dokumen Arsip (PDF)</p>
                  <p class="text-[10px] text-zinc-500 font-mono truncate">{{ fileArsipUrl }}</p>
                </div>
              </div>

              <a
                :href="getFullAssetUrl(fileArsipUrl)"
                target="_blank"
                rel="noopener noreferrer"
                class="shrink-0 px-3.5 py-1.5 bg-white border border-black/10 text-xs font-semibold text-[#0A84DC] rounded-xl hover:bg-zinc-50 shadow-2xs transition-all duration-200 active:scale-95 flex items-center gap-1.5"
              >
                <span>Lihat Dokumen</span>
                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
              </a>
            </div>

            <!-- Image Document Preview -->
            <div v-else class="rounded-2xl border border-black/10 bg-white p-3.5 shadow-2xs space-y-2.5">
              <div class="aspect-4/3 max-h-56 rounded-xl overflow-hidden bg-zinc-100 border border-black/5 flex items-center justify-center">
                <img
                  :src="getFullAssetUrl(fileArsipUrl)"
                  alt="Pratinjau Arsip Dokumen"
                  class="w-full h-full object-contain"
                />
              </div>
              <a
                :href="getFullAssetUrl(fileArsipUrl)"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-block text-xs text-[#0A84DC] hover:underline font-semibold px-1"
              >
                Buka Arsip Resolusi Penuh &rarr;
              </a>
            </div>
          </div>

          <div v-else-if="!fileArsipUrl && !isCameraOpen" class="rounded-2xl border border-dashed border-zinc-300 bg-white/50 p-5 text-center text-xs text-zinc-500">
            Belum ada arsip dokumen yang diunggah.
          </div>
        </section>

        <!-- 3. Comments / Logs Timeline Section -->
        <section class="space-y-2.5">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-500 px-1">Riwayat Tracking (Log)</h3>

          <div v-if="!props.surat.logs || props.surat.logs.length === 0" class="text-xs text-zinc-500 italic px-1">
            Belum ada catatan riwayat tracking.
          </div>

          <div v-else class="space-y-3">
            <TransitionGroup
              name="log-item"
              tag="ol"
              class="relative border-l-2 border-[#0A84DC]/30 ml-2.5 space-y-4 text-xs"
            >
              <li
                v-for="log in displayedLogs"
                :key="log.id"
                class="ml-4 space-y-1"
              >
                <div class="absolute -left-[5px] mt-1.5 w-2.5 h-2.5 rounded-full border-2 border-white bg-[#0A84DC]"></div>
                <div class="flex items-center justify-between">
                  <span class="font-semibold text-zinc-900">{{ log.status_baru }}</span>
                  <time class="text-[11px] text-zinc-500 font-mono">{{ formatDate(log.waktu_update) }}</time>
                </div>
                <p v-if="log.catatan_log" class="text-zinc-600 bg-white/80 border border-black/5 p-2.5 rounded-2xl text-[11px] leading-relaxed shadow-2xs">
                  {{ log.catatan_log }}
                </p>
              </li>
            </TransitionGroup>

            <div v-if="reversedLogs.length > 5" class="pl-2.5">
              <button
                type="button"
                @click="showAllLogs = !showAllLogs"
                class="text-xs font-medium text-[#0A84DC] hover:text-[#0872be] transition-colors py-1 cursor-pointer active:opacity-70"
              >
                {{ showAllLogs ? 'Sembunyikan' : 'Lihat Semua' }}
              </button>
            </div>
          </div>
        </section>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.log-item-enter-active,
.log-item-leave-active {
  transition: all 0.25s ease;
}
.log-item-enter-from,
.log-item-leave-to {
  opacity: 0;
  transform: translateY(-6px);
}
</style>
