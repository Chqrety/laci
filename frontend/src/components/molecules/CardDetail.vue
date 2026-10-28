<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import {
  api,
  getBackendAssetUrl,
  suratService,
  userService,
  currentRole,
  isKoor,
  type Surat,
  type User,
} from '@/services/api'
import { CardCameraCapture, CardLogsTimeline } from '@/components/card-detail'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { toast } from '@/components/ui/toast'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const props = defineProps<{
  surat: Surat
  statusOptions: readonly string[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'updated'): void
}>()

const userRole = computed(() => currentRole.value)
const isSekre = computed(() => userRole.value === 'sekre' || userRole.value === 'ketum')
const canAssignPic = computed(() => isSekre.value || isKoor.value)

const isCompletedStatus = computed(() => {
  const s = props.surat.status_saat_ini?.toLowerCase() || ''
  return s === 'selesai' || s === 'arsip'
})

const canUploadArsip = computed(() => isSekre.value || isCompletedStatus.value)

// Status, PIC, & Note update state
const selectedStatus = ref(props.surat.status_saat_ini)
const selectedPic = ref(props.surat.pic_nama || '')
const newNote = ref(props.surat.catatan || '')
const isUpdatingStatus = ref(false)
const updateMessage = ref('')
const userList = ref<User[]>([])

const loadPicUsers = async () => {
  if (canAssignPic.value) {
    try {
      userList.value = await userService.getAll()
    } catch {}
  }
}

watch(
  () => props.surat,
  (newVal) => {
    if (newVal) {
      selectedStatus.value = newVal.status_saat_ini
      selectedPic.value = newVal.pic_nama || ''
      newNote.value = newVal.catatan || ''
      hasImageError.value = false
    }
  },
)

// Document upload & native camera state
const isCameraOpen = ref(false)
const useNativeCamera = ref(typeof window !== 'undefined' && window.isSecureContext === false)
const hasImageError = ref(false)
const isUploadingDokumen = ref(false)
const uploadDokumenMessage = ref('')
const uploadDokumenError = ref('')

const fileArsipUrl = computed(() => {
  return props.surat.arsip_url || null
})

const isImage = computed(() => {
  if (!fileArsipUrl.value) return false
  const cleanUrl = fileArsipUrl.value.split('?')[0]?.toLowerCase() || ''
  return /\.(jpe?g|png|webp|gif|svg|bmp)$/i.test(cleanUrl) || fileArsipUrl.value.startsWith('data:image/')
})

const isPdf = computed(() => {
  if (!fileArsipUrl.value) return false
  const url = fileArsipUrl.value.toLowerCase()
  return url.includes('.pdf') || url.includes('format=pdf') || url.includes('/export')
})

const handleImageError = (event: Event) => {
  hasImageError.value = true
  console.error('[CardDetail] Gagal memuat pratinjau arsip gambar:', {
    rawUrl: fileArsipUrl.value,
    resolvedUrl: getBackendAssetUrl(fileArsipUrl.value),
    event,
  })
}

const openCamera = () => {
  if (!canUploadArsip.value) {
    uploadDokumenError.value = 'Pengunggahan arsip fisik hanya dapat dilakukan saat surat berstatus Selesai.'
    return
  }
  uploadDokumenMessage.value = ''
  uploadDokumenError.value = ''
  isCameraOpen.value = true
}

const handleFileUpload = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file) return

  uploadDokumenMessage.value = ''
  uploadDokumenError.value = ''
  isUploadingDokumen.value = true

  const formData = new FormData()
  formData.append('file', file)

  try {
    await api.post(`/letters/${props.surat.id}/upload-arsip`, formData)
    uploadDokumenMessage.value = 'Arsip berhasil diunggah'
    toast({
      title: 'Berhasil',
      description: 'Berkas arsip surat berhasil diperbarui.',
      variant: 'success',
    })
    emit('updated')
  } catch (err: any) {
    const errorMsg = err.response?.data?.error || err.message || 'Gagal mengunggah berkas arsip'
    uploadDokumenError.value = errorMsg
    toast({
      title: 'Gagal Mengunggah',
      description: errorMsg,
      variant: 'destructive',
    })
  } finally {
    isUploadingDokumen.value = false
    target.value = ''
  }
}

const handleClosePanel = () => {
  isCameraOpen.value = false
  document.body.style.overflow = ''
  emit('close')
}

const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    handleClosePanel()
  }
}

onMounted(() => {
  document.body.style.overflow = 'hidden'
  window.addEventListener('keydown', handleKeyDown)
  loadPicUsers()
})

onBeforeUnmount(() => {
  document.body.style.overflow = ''
  window.removeEventListener('keydown', handleKeyDown)
})

const handleUpdateStatusAndNote = async () => {
  isUpdatingStatus.value = true
  updateMessage.value = ''
  try {
    await suratService.updateStatus(props.surat.id, {
      status_saat_ini: selectedStatus.value,
      catatan: newNote.value,
      catatan_log: newNote.value,
      pic_nama: canAssignPic.value ? selectedPic.value : undefined,
    })
    updateMessage.value = 'Data surat berhasil diperbarui'
    toast({
      title: 'Berhasil',
      description: updateMessage.value,
      variant: 'success',
    })
    emit('updated')
  } catch (err: any) {
    toast({
      title: 'Gagal',
      description: err.response?.data?.error || err.message || 'Gagal memperbarui data surat',
      variant: 'destructive',
    })
  } finally {
    isUpdatingStatus.value = false
  }
}
</script>

<template>
  <aside
    aria-label="Detail Surat"
    class="fixed inset-0 z-[100] flex justify-end bg-black/60 backdrop-blur-xs transition-opacity"
  >
    <!-- Backdrop close click area -->
    <div class="fixed inset-0" @click="handleClosePanel" aria-hidden="true"></div>

    <!-- iOS Slide-over Sheet (rounded-3xl, pure solid white, no bleed) -->
    <div class="drawer-panel relative w-full max-w-lg bg-white h-full shadow-2xl flex flex-col z-10 border-l border-black/10 overflow-hidden sm:rounded-l-3xl">
      <!-- Header -->
      <header class="p-6 border-b border-black/5 flex items-start justify-between bg-white">
        <div class="space-y-1.5">
          <Badge variant="mono" size="sm">
            {{ props.surat.nomor_surat }}
          </Badge>
          <h2 class="text-base font-semibold text-zinc-900 leading-snug">{{ props.surat.perihal }}</h2>
          <p v-if="props.surat.tujuan" class="text-xs text-zinc-600 font-medium">Tujuan: <span class="text-zinc-900 font-semibold">{{ props.surat.tujuan }}</span></p>
        </div>

        <Button
          type="button"
          variant="ghost"
          size="icon"
          @click="handleClosePanel"
          aria-label="Tutup Panel"
          class="rounded-2xl text-zinc-400 hover:text-zinc-700"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </Button>
      </header>

      <!-- Scrollable Body Content -->
      <div class="flex-1 overflow-y-auto no-scrollbar p-6 space-y-6">
        <!-- 1. Properties Section (Editable for Sekre & Humas) -->
        <section class="space-y-2.5">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-500 px-1">Status &amp; Penugasan</h3>

          <form @submit.prevent="handleUpdateStatusAndNote" novalidate class="space-y-3">
            <div class="rounded-2xl border border-black/10 bg-white overflow-hidden divide-y divide-black/5 shadow-2xs">
              <!-- Status Input (Dapat diubah oleh semua pengurus pelaksana & sekre) -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <Label for="update-status" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Status Saat Ini</Label>
                <Select
                  v-model="selectedStatus"
                >
                  <SelectTrigger id="update-status" class="mt-1">
                    <SelectValue :placeholder="selectedStatus || 'Pilih Status'" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="opt in props.statusOptions" :key="opt" :value="opt">
                      {{ opt }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <!-- PIC Assignment (Select untuk Koordinator / Sekre, Read-only untuk staf Humas) -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <Label for="update-pic" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider">Penanggung Jawab (PIC)</Label>
                <div v-if="canAssignPic" class="mt-1">
                  <Select v-model="selectedPic">
                    <SelectTrigger id="update-pic">
                      <SelectValue :placeholder="selectedPic || '-- Pilih Pengurus PIC --'" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="user in userList" :key="user.id" :value="user.nama || user.username">
                        {{ user.nama }} ({{ user.divisi || user.role || 'Pengurus' }}{{ user.is_koor ? ' - Koor' : '' }})
                      </SelectItem>
                      <SelectItem
                        v-if="selectedPic && !userList.some((u) => (u.nama || u.username) === selectedPic)"
                        :value="selectedPic"
                      >
                        {{ selectedPic }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div v-else class="text-xs font-semibold text-zinc-900 pt-0.5 flex items-center justify-between">
                  <span>{{ props.surat.pic_nama }}</span>
                  <span class="text-[10px] text-zinc-400 font-normal italic">Mandat Penugasan Koordinator</span>
                </div>
              </div>

              <!-- Catatan Dokumen (Always editable for both Sekre and Humas) -->
              <div class="px-4 py-2.5 focus-within:bg-zinc-50/50 transition-colors">
                <Label for="update-note" class="block text-[11px] font-semibold text-zinc-500 uppercase tracking-wider mb-1">Catatan Dokumen</Label>
                <Textarea
                  id="update-note"
                  v-model="newNote"
                  rows="2"
                  placeholder="Tambahkan catatan instruksi..."
                  class="text-xs"
                />
              </div>
            </div>

            <div class="flex items-center justify-between pt-1">
              <span v-if="updateMessage" class="text-[11px] text-emerald-600 font-semibold">{{ updateMessage }}</span>
              <Button
                type="submit"
                :disabled="isUpdatingStatus"
                class="ml-auto"
              >
                {{ isUpdatingStatus ? 'Menyimpan...' : 'Simpan Perubahan' }}
              </Button>
            </div>
          </form>
        </section>

        <!-- 2. Arsip Dokumen Section (In-App Camera WebRTC & File Upload) -->
        <section class="space-y-3">
          <div class="flex items-center justify-between px-1">
            <h3 class="text-xs font-semibold uppercase tracking-wider text-zinc-500">Arsip Dokumen</h3>
            <Badge
              v-if="fileArsipUrl"
              variant="emerald"
              size="sm"
            >
              Tersedia
            </Badge>
          </div>

          <!-- Isolated Web Camera Sub-Component -->
          <CardCameraCapture
            v-model:open="isCameraOpen"
            :surat-id="props.surat.id"
            @captured="emit('updated')"
          />

          <!-- Locked State if Status is not Finished / Selesai -->
          <div v-if="!canUploadArsip" class="p-3 rounded-2xl bg-amber-50/60 border border-amber-200/70 flex items-start gap-2.5 text-xs text-amber-900">
            <svg class="w-4 h-4 text-amber-600 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <span class="leading-relaxed">
              Pengunggahan arsip dokumen fisik hanya dapat dilakukan setelah surat selesai ditandatangani dan berstatus <b>Selesai</b>.
            </span>
          </div>

          <!-- Fallback Native Camera Button (Saat WebRTC Gagal/HTTP Network) -->
          <div v-else-if="useNativeCamera" class="flex items-center gap-2">
            <Button as-child size="sm" class="cursor-pointer">
              <label>
                <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                  <path stroke-linecap="round" stroke-linejoin="round" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span>{{ fileArsipUrl ? 'Buka Kamera' : 'Buka Kamera Perangkat' }}</span>
                <input
                  type="file"
                  accept="image/*,application/pdf"
                  capture="environment"
                  class="hidden"
                  @change="handleFileUpload"
                />
              </label>
            </Button>

            <!-- File Upload Fallback -->
            <Button as-child variant="outline" size="sm" class="cursor-pointer">
              <label>
                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 mr-1.5 text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
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
            </Button>
          </div>

          <!-- Buka Kamera Button (When Camera Closed) -->
          <div v-else-if="!isCameraOpen" class="flex items-center gap-2">
            <Button
              type="button"
              size="sm"
              @click="openCamera"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mr-1.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              <span>{{ fileArsipUrl ? 'Buka Kamera' : 'Buka Kamera' }}</span>
            </Button>

            <!-- File Upload Fallback -->
            <Button as-child variant="outline" size="sm" class="cursor-pointer">
              <label>
                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 mr-1.5 text-zinc-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
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
            </Button>
          </div>

          <!-- Status feedback messages -->
          <p v-if="uploadDokumenMessage" class="text-[11px] text-emerald-600 font-semibold px-1">
            {{ uploadDokumenMessage }}
          </p>
          <p v-if="uploadDokumenError" class="text-[11px] text-rose-600 font-medium px-1">
            {{ uploadDokumenError }}
          </p>

          <!-- Pratinjau Dokumen jika ada & kamera tidak aktif -->
          <div v-if="!isCameraOpen || useNativeCamera" class="space-y-3">
            <!-- Google Docs (Editable) Card jika ada gdocs_url -->
            <div v-if="surat.gdocs_url" class="rounded-2xl border border-sky-200/80 bg-sky-50/50 p-3.5 shadow-2xs flex items-center justify-between gap-3">
              <div class="flex items-center gap-3 overflow-hidden">
                <div class="w-10 h-10 rounded-xl bg-sky-100 border border-sky-200 flex items-center justify-center text-sky-600 shrink-0">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <div class="min-w-0 flex-1">
                  <p class="text-xs font-semibold text-slate-800 truncate">Draft Google Docs</p>
                  <p class="text-[11px] text-slate-500 truncate">Dokumen surat yang dihasilkan dari template</p>
                </div>
              </div>

              <Button as-child size="sm" variant="outline" class="shrink-0 bg-white border-sky-200 hover:bg-sky-50 text-sky-700">
                <a :href="surat.gdocs_url" target="_blank" rel="noopener noreferrer" class="flex items-center gap-1.5">
                  <span>Buka Dokumen</span>
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                  </svg>
                </a>
              </Button>
            </div>

            <div v-if="fileArsipUrl" class="space-y-3">
              <div v-if="isImage" class="relative group rounded-2xl border border-black/10 overflow-hidden bg-zinc-950 flex flex-col items-center justify-center min-h-[220px]">
                <img
                  v-if="!hasImageError"
                  :src="getBackendAssetUrl(fileArsipUrl)"
                  alt="Arsip Dokumen"
                  class="w-full h-auto max-h-[360px] object-contain rounded-2xl"
                  @error="handleImageError"
                />
                <div v-else class="p-6 text-center text-xs text-zinc-400 space-y-1">
                  <p class="font-medium text-zinc-300">Pratinjau gambar tidak dapat dimuat</p>
                  <p class="text-[11px] text-zinc-500">Berkas tetap tersimpan dan dapat dibuka melalui link di bawah.</p>
                </div>
              </div>

              <div v-else-if="isPdf" class="rounded-2xl border border-black/10 p-4 bg-zinc-50 flex items-center justify-between">
                <div class="flex items-center space-x-3">
                  <span class="text-xs font-semibold text-zinc-700">Berkas Arsip PDF</span>
                </div>
                <Button as-child variant="link" size="sm" class="text-xs text-[#0A84DC]">
                  <a :href="getBackendAssetUrl(fileArsipUrl)" target="_blank" rel="noopener noreferrer">
                    Lihat PDF &rarr;
                  </a>
                </Button>
              </div>

              <!-- Quick Link Eksternal -->
              <div class="flex items-center justify-end px-1">
                <Button as-child variant="link" size="sm" class="h-auto p-0 text-[11px] text-[#0A84DC]">
                  <a :href="getBackendAssetUrl(fileArsipUrl)" target="_blank" rel="noopener noreferrer">
                    Buka Berkas di Tab Baru &rarr;
                  </a>
                </Button>
              </div>
            </div>
          </div>
        </section>

        <!-- 3. Comments / Logs Timeline Section (Isolated Sub-Component) -->
        <CardLogsTimeline :logs="props.surat.logs" />
      </div>
    </div>
  </aside>
</template>
