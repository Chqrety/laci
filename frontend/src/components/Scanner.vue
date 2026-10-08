<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Cropper } from 'vue-advanced-cropper'
import 'vue-advanced-cropper/dist/style.css'
import { getBackendAssetUrl, suratService, type Surat } from '@/services/api'

const route = useRoute()
const router = useRouter()

const cropperRef = ref<any>(null)
const imageSrc = ref<string | null>(null)
const suratList = ref<Surat[]>([])
const selectedSuratId = ref<string>('')
const isLoadingSurat = ref<boolean>(false)
const isUploading = ref<boolean>(false)
const uploadStatus = ref<{ success: boolean; message: string; arsipUrl?: string } | null>(null)

// If an ID is provided in route params or query
const preselectedId = computed(() => (route.params.id as string) || (route.query.id as string) || '')

const fetchSuratList = async () => {
  isLoadingSurat.value = true
  try {
    suratList.value = await suratService.getAll()
    if (preselectedId.value) {
      selectedSuratId.value = preselectedId.value
    } else if (suratList.value.length > 0 && !selectedSuratId.value) {
      const firstSurat = suratList.value[0]
      if (firstSurat) {
        selectedSuratId.value = firstSurat.id
      }
    }
  } catch (err: any) {
    console.error('Failed to load surat list', err)
  } finally {
    isLoadingSurat.value = false
  }
}

const currentSelectedSurat = computed(() => {
  return suratList.value.find((s) => s.id === selectedSuratId.value)
})

const onFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (target.files && target.files[0]) {
    const file = target.files[0]
    uploadStatus.value = null
    const reader = new FileReader()
    reader.onload = (e) => {
      imageSrc.value = e.target?.result as string
    }
    reader.readAsDataURL(file)
  }
}

const rotateClockwise = () => {
  if (cropperRef.value) {
    cropperRef.value.rotate(90)
  }
}

const flipHorizontal = () => {
  if (cropperRef.value) {
    cropperRef.value.flip(true, false)
  }
}

const resetCrop = () => {
  if (cropperRef.value) {
    cropperRef.value.reset()
  }
}

const clearImage = () => {
  imageSrc.value = null
  uploadStatus.value = null
}

const uploadCroppedImage = async () => {
  if (!cropperRef.value || !selectedSuratId.value) {
    alert('Pilih surat target dan pastikan gambar telah dipilih.')
    return
  }

  const { canvas } = cropperRef.value.getResult()
  if (!canvas) {
    alert('Gagal mengambil hasil crop dari kanvas.')
    return
  }

  isUploading.value = true
  uploadStatus.value = null

  canvas.toBlob(async (blob: Blob | null) => {
    if (!blob) {
      isUploading.value = false
      alert('Gagal mengonversi gambar ke format biner.')
      return
    }

    try {
      const filename = `arsip-${selectedSuratId.value.slice(0, 8)}-${Date.now()}.jpg`
      const res = await suratService.uploadArsip(selectedSuratId.value, blob, filename)

      uploadStatus.value = {
        success: true,
        message: 'Arsip fisik berhasil dipindai dan diunggah.',
        arsipUrl: res.arsip_url,
      }
    } catch (err: any) {
      uploadStatus.value = {
        success: false,
        message: err.response?.data?.error || err.message || 'Gagal mengunggah arsip dokumen.',
      }
    } finally {
      isUploading.value = false
    }
  }, 'image/jpeg', 0.9)
}

onMounted(() => {
  fetchSuratList()
})
</script>

<template>
  <section class="max-w-4xl mx-auto space-y-6">
    <!-- Header -->
    <header class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-zinc-200">
      <div>
        <h2 class="text-xl font-semibold tracking-tight text-zinc-900">Pemindai & Unggah Arsip</h2>
        <p class="text-sm text-zinc-500 mt-1">Unggah atau tangkap foto dokumen surat, sesuaikan area crop, lalu simpan ke arsip sistem.</p>
      </div>

      <button
        type="button"
        @click="router.push({ name: 'dashboard' })"
        class="self-start sm:self-auto px-3 py-1.5 text-sm font-medium border border-zinc-300 rounded-md bg-white text-zinc-700 hover:bg-zinc-50"
      >
        &larr; Kembali ke Dashboard
      </button>
    </header>

    <!-- Form: Select Surat -->
    <div class="bg-white p-4 rounded-lg border border-zinc-200 space-y-3">
      <label for="select-surat" class="block text-sm font-medium text-zinc-800">
        Pilih Surat Target Arsip:
      </label>

      <div class="flex flex-col sm:flex-row gap-3">
        <select
          id="select-surat"
          v-model="selectedSuratId"
          :disabled="isLoadingSurat"
          class="flex-1 px-3 py-2 text-sm border border-zinc-300 rounded-md bg-white focus:outline-none focus:ring-2 focus:ring-zinc-400 disabled:opacity-50"
        >
          <option value="" disabled>-- Pilih Surat yang Akan Diarsip --</option>
          <option v-for="s in suratList" :key="s.id" :value="s.id">
            [{{ s.nomor_surat }}] - {{ s.perihal }} (PIC: {{ s.pic_nama }})
          </option>
        </select>
      </div>

      <div v-if="currentSelectedSurat" class="text-xs text-zinc-500 flex flex-wrap gap-x-4 gap-y-1 pt-1">
        <span>Status: <strong class="text-zinc-700">{{ currentSelectedSurat.status_saat_ini }}</strong></span>
        <span>
          Status Arsip:
          <strong :class="currentSelectedSurat.arsip_url ? 'text-emerald-700' : 'text-zinc-500'">
            {{ currentSelectedSurat.arsip_url ? 'Sudah Ada Arsip' : 'Belum Ada Arsip' }}
          </strong>
        </span>
      </div>
    </div>

    <!-- Upload Input / Capture -->
    <div class="bg-white p-6 rounded-lg border border-zinc-200">
      <div v-if="!imageSrc" class="text-center py-10 border-2 border-dashed border-zinc-200 rounded-lg">
        <p class="text-sm font-medium text-zinc-800 mb-1">Pilih Berkas atau Ambil Foto Dokumen</p>
        <p class="text-xs text-zinc-500 mb-4">Format berkas: JPG, PNG, atau WEBP</p>
        
        <label class="cursor-pointer inline-flex items-center px-4 py-2 text-sm font-medium rounded-md bg-zinc-900 text-white hover:bg-zinc-800 focus:outline-none">
          <span>Pilih / Jepret Dokumen</span>
          <input
            type="file"
            accept="image/*"
            capture="environment"
            @change="onFileSelect"
            class="hidden"
          />
        </label>
      </div>

      <!-- Cropper Viewport -->
      <div v-else class="space-y-4">
        <div class="overflow-hidden rounded-md border border-zinc-200 bg-zinc-900 max-h-[500px] flex items-center justify-center">
          <Cropper
            ref="cropperRef"
            :src="imageSrc"
            :stencil-props="{
              aspectRatio: 0,
            }"
            class="w-full h-[450px]"
          />
        </div>

        <!-- Controls Toolbar -->
        <div class="flex flex-wrap items-center justify-between gap-3 pt-2 border-t border-zinc-100">
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="rotateClockwise"
              class="px-2.5 py-1 text-xs font-medium border border-zinc-200 rounded text-zinc-700 bg-white hover:bg-zinc-50"
            >
              Putar 90°
            </button>
            <button
              type="button"
              @click="flipHorizontal"
              class="px-2.5 py-1 text-xs font-medium border border-zinc-200 rounded text-zinc-700 bg-white hover:bg-zinc-50"
            >
              Balik Horisontal
            </button>
            <button
              type="button"
              @click="resetCrop"
              class="px-2.5 py-1 text-xs font-medium border border-zinc-200 rounded text-zinc-700 bg-white hover:bg-zinc-50"
            >
              Reset Area
            </button>
            <button
              type="button"
              @click="clearImage"
              class="px-2.5 py-1 text-xs font-medium border border-zinc-200 rounded text-rose-600 bg-white hover:bg-rose-50"
            >
              Ganti Gambar
            </button>
          </div>

          <button
            type="button"
            @click="uploadCroppedImage"
            :disabled="isUploading || !selectedSuratId"
            class="px-4 py-2 text-sm font-medium rounded-md bg-zinc-900 text-white hover:bg-zinc-800 disabled:opacity-50"
          >
            {{ isUploading ? 'Mengunggah Arsip...' : 'Crop & Simpan Arsip' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Status Feedback -->
    <div
      v-if="uploadStatus"
      :class="uploadStatus.success ? 'bg-emerald-50 border-emerald-200 text-emerald-800' : 'bg-rose-50 border-rose-200 text-rose-800'"
      class="p-4 border rounded-md text-sm space-y-2"
    >
      <p class="font-medium">{{ uploadStatus.message }}</p>
      <div v-if="uploadStatus.arsipUrl" class="flex items-center gap-4 text-xs">
        <a
          :href="getBackendAssetUrl(uploadStatus.arsipUrl)"
          target="_blank"
          rel="noreferrer"
          class="underline font-semibold"
        >
          Buka Tautan Arsip
        </a>
        <button
          type="button"
          @click="router.push({ name: 'dashboard' })"
          class="underline font-medium"
        >
          Kembali ke Dashboard &rarr;
        </button>
      </div>
    </div>
  </section>
</template>
