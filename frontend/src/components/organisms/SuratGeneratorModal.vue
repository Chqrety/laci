<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  templateService,
  prokerService,
  userService,
  type SuratTemplate,
  type Proker,
  type User,
  type Surat,
} from '@/services/api'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { toast } from '@/components/ui/toast'

const props = defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', surat: Surat): void
  (e: 'openSetup'): void
}>()

const romanMonths = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X', 'XI', 'XII']

const idDateFormatter = new Intl.DateTimeFormat('id-ID', {
  weekday: 'long',
  day: 'numeric',
  month: 'long',
  year: 'numeric',
})

const idDateOnlyFormatter = new Intl.DateTimeFormat('id-ID', {
  day: 'numeric',
  month: 'long',
  year: 'numeric',
})

const STANDARD_COVERED = new Set([
  'nomor_surat',
  'perihal',
  'tujuan',
  'nama_kegiatan',
  'hari_tanggal',
  'tanggal',
  'tanggal_surat',
  'tanggal_pembuatan',
  'hari_tanggal_surat',
  'tanggal_acara',
  'tanggal_pelaksanaan',
  'waktu',
  'tempat',
  'narahubung',
  'nama_narahubung',
  'nohp_narahubung',
  'narahubung_nama',
  'narahubung_nohp',
  'narahubung_hp',
  'nama_cp',
  'nohp_cp',
  'no_hp',
  'nama_ketum',
  'nim_ketum',
  'nama_dasar',
  'nim_dasar',
  'nama_tingkat_1',
  'nim_tingkat_1',
  'nama_lapis_1',
  'nim_lapis_1',
  'nama_lapis1',
  'nim_lapis1',
  'nama_ketupel',
  'nim_ketupel',
])

const templates = ref<SuratTemplate[]>([])
const prokerList = ref<Proker[]>([])
const userList = ref<User[]>([])
const isLoadingData = ref(false)
const isSubmitting = ref(false)

// Template Selection
const selectedTemplateId = ref<string>('')

// 1. Nomor Surat Segmented State (Format Versi Awal)
const noUrut = ref('01')
const selectedProkerCode = ref('NONE') // 'NONE' = Tanpa Proker (Kosong)
const selectedBulan = ref(romanMonths[new Date().getMonth()] || 'X')
const selectedTahun = ref(new Date().getFullYear().toString())

// 2. Data Pokok Surat
const perihal = ref('')
const outputFolderUrl = ref('')

// 3. Detail Acara Kegiatan
const namaKegiatan = ref('')
const tipeTanggal = ref<'single' | 'range'>('single')
const tanggalMulai = ref('')
const tanggalSelesai = ref('')
const jamMulai = ref('')
const jamSelesai = ref('')
const isSampaiSelesai = ref(false)
const tempatAcara = ref('')

const formattedWaktu = computed(() => {
  if (!jamMulai.value) return ''
  if (isSampaiSelesai.value) {
    return `${jamMulai.value} WIB - Selesai`
  }
  if (jamSelesai.value) {
    return `${jamMulai.value} WIB - ${jamSelesai.value} WIB`
  }
  return `${jamMulai.value} WIB`
})

// UX Helper: Buka dialog picker kalender/jam otomatis saat field diklik,
// sementara kursor/fokus tetap aktif sehingga bisa langsung diedit lewat keyboard
const handlePickerClick = (e: MouseEvent) => {
  try {
    const el = (e.currentTarget || e.target) as HTMLInputElement
    if (el && !el.disabled) {
      el.showPicker?.()
    }
  } catch {
    // Graceful fallback jika browser tidak support showPicker()
  }
}

// 4. Narahubung (2 Input Terpisah)
const namaNarahubung = ref('')
const nohpNarahubung = ref('')

// 5. Penandatangan (Ketum di Kiri, Ketupel di Kanan)
const selectedKetumId = ref('')
const selectedKetupelId = ref('')

// Dynamic Extra Placeholders
const dynamicReplacements = ref<Record<string, string>>({})

// Success Result
const generationResult = ref<{
  surat: Surat
  gdocs_url: string
  pdf_url: string
} | null>(null)

const selectedTemplate = computed(() => {
  return templates.value.find((t) => t.id === selectedTemplateId.value) || null
})

// Data Pengurus dengan divisi Ketua Umum (hanya divisi ketum)
const ketumOptions = computed(() => {
  const byDivisi = userList.value.filter((u) => u.divisi?.toLowerCase() === 'ketum')
  if (byDivisi.length > 0) return byDivisi
  return userList.value.filter((u) => u.nama?.toLowerCase().includes('ketua umum'))
})

const ketumUser = computed(() => ketumOptions.value[0] || null)

// Penandatangan Dasar (Ketua Pelaksana / Pembuat): Semua user kecuali Ketum
const dasarUsers = computed(() => {
  const ketumId = selectedKetumId.value
  return userList.value
    .filter((u) => u.divisi?.toLowerCase() !== 'ketum' && u.id !== ketumId)
    .sort((a, b) => {
      const aIsDasar = ['dasar', 'pengurus', 'tupel'].includes(a.divisi?.toLowerCase() || '') ? 1 : 0
      const bIsDasar = ['dasar', 'pengurus', 'tupel'].includes(b.divisi?.toLowerCase() || '') ? 1 : 0
      return bIsDasar - aIsDasar
    })
})
const lapis1Users = dasarUsers

watch(selectedKetumId, (newKetumId) => {
  if (selectedKetupelId.value && selectedKetupelId.value === newKetumId) {
    selectedKetupelId.value = ''
  }
})

// Ekstrak prefix angka dari nama template (contoh: "01_Surat Undangan" -> "01")
const extractPrefixNoUrut = (name: string): string => {
  const match = name.trim().match(/^(\d+)/)
  return match?.[1] ? match[1].padStart(2, '0') : '01'
}

// Format Tanggal Indonesia
const formatIndonesianDate = (dateStr: string): string => {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    return isNaN(d.getTime()) ? dateStr : idDateFormatter.format(d)
  } catch {
    return dateStr
  }
}

// Format Hari/Tanggal (Single vs Range)
const formattedHariTanggal = computed(() => {
  if (tipeTanggal.value === 'single') {
    return formatIndonesianDate(tanggalMulai.value)
  }
  if (tanggalMulai.value && tanggalSelesai.value) {
    return `${formatIndonesianDate(tanggalMulai.value)} s.d. ${formatIndonesianDate(tanggalSelesai.value)}`
  }
  return formatIndonesianDate(tanggalMulai.value)
})

// Format Narahubung: Format 2 Baris (Nama & No.Hp)
const formattedNarahubung = computed(() => {
  const nama = namaNarahubung.value.trim()
  const hp = nohpNarahubung.value.trim()
  if (nama && hp) return `Nama   : ${nama}\nNo.Hp  : ${hp}`
  if (nama) return `Nama   : ${nama}`
  if (hp) return `No.Hp  : ${hp}`
  return ''
})

// Full Nomor Surat Computed: jika Tanpa Proker, segmen proker dihilangkan
const computedNomorSurat = computed(() => {
  const paddedNo = (noUrut.value.trim() || '01').padStart(2, '0')
  const isNoProker = !selectedProkerCode.value || selectedProkerCode.value === 'NONE'

  if (isNoProker) {
    return `${paddedNo}/DOSCOM/${selectedBulan.value}/${selectedTahun.value.trim()}`
  }

  const prokerPart = selectedProkerCode.value.toUpperCase().replace(/\s+/g, '-')
  return `${paddedNo}/DOSCOM/${prokerPart}/${selectedBulan.value}/${selectedTahun.value.trim()}`
})

// Format Tanggal untuk Penamaan File (contoh: 26Okt2026)
const formatFileDate = (date: Date): string => {
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
  const day = String(date.getDate()).padStart(2, '0')
  const month = months[date.getMonth()]
  const year = date.getFullYear()
  return `${day}${month}${year}`
}

// Penamaan File Otomatis: [Template]_[Proker]_[Tanggal] atau [Template]_[Tanggal]
const computedFileName = computed(() => {
  const rawTplName = selectedTemplate.value?.nama_template || 'Surat'
  const cleanTplName = rawTplName.trim().replace(/[\/\\?%*:|"<>]/g, '_')
  const dateStr = formatFileDate(new Date())
  const isNoProker = !selectedProkerCode.value || selectedProkerCode.value === 'NONE'

  if (isNoProker) {
    return `${cleanTplName}_${dateStr}`
  }

  const prokerPart = selectedProkerCode.value.toUpperCase().replace(/\s+/g, '-')
  return `${cleanTplName}_${prokerPart}_${dateStr}`
})

const templatePlaceholders = computed<string[]>(() => {
  if (!selectedTemplate.value) return []
  try {
    const list = JSON.parse(selectedTemplate.value.placeholders)
    return Array.isArray(list) ? list : []
  } catch {
    return []
  }
})

// Placeholder ekstra yang belum ter-cover oleh form standar
const extraPlaceholders = computed<string[]>(() => {
  return templatePlaceholders.value.filter((p) => !STANDARD_COVERED.has(p.toLowerCase()))
})

const loadData = async () => {
  isLoadingData.value = true
  try {
    const [tplData, prokerData, usersData, settings] = await Promise.all([
      templateService.getAll(),
      prokerService.getAll(),
      userService.getAll().catch(() => [] as User[]),
      templateService.getSettings().catch(() => ({} as Record<string, string>)),
    ])

    templates.value = tplData
    prokerList.value = prokerData
    userList.value = usersData

    if (settings.default_output_folder_url) {
      outputFolderUrl.value = settings.default_output_folder_url
    }

    // Auto set Ketum jika ditemukan
    const foundKetum = usersData.find((u) => u.divisi?.toLowerCase() === 'ketum')
    if (foundKetum && !selectedKetumId.value) {
      selectedKetumId.value = foundKetum.id
    }

    // Auto set Dasar jika ada user berdivisi dasar, pengurus, atau tupel
    const foundDasar = usersData.find(
      (u) =>
        ['dasar', 'pengurus', 'tupel'].includes(u.divisi?.toLowerCase() || '') &&
        (!foundKetum || u.id !== foundKetum.id)
    )
    if (foundDasar && !selectedKetupelId.value) {
      selectedKetupelId.value = foundDasar.id
    }

    // Pastikan template terpilih valid di list baru
    const templateExists = templates.value.some((t) => t.id === selectedTemplateId.value)
    if (!templateExists && templates.value.length > 0 && templates.value[0]) {
      selectedTemplateId.value = templates.value[0].id
    }
  } catch (err) {
    console.error('Gagal memuat data generator surat:', err)
  } finally {
    isLoadingData.value = false
  }
}

watch(
  () => props.isOpen,
  (open) => {
    if (open) {
      generationResult.value = null
      loadData()
    }
  }
)

// Saat template berubah, perbarui nomor urut dan perihal otomatis sesuai template yang dipilih
watch(
  selectedTemplate,
  (tpl) => {
    if (!tpl) return
    noUrut.value = extractPrefixNoUrut(tpl.nama_template)
    perihal.value = tpl.nama_template.replace(/^\d+[\s_-]*/, '').replace(/_/g, ' ').trim()
  },
  { immediate: true }
)

const handleSubmit = async () => {
  if (!selectedTemplateId.value) {
    toast({
      title: 'Validasi Gagal',
      description: 'Pilih template surat terlebih dahulu.',
      variant: 'destructive',
    })
    return
  }

  if (!perihal.value.trim()) {
    toast({
      title: 'Validasi Gagal',
      description: 'Perihal surat wajib diisi.',
      variant: 'destructive',
    })
    return
  }

  // Validasi: Ketum dan Penandatangan Dasar tidak boleh orang yang sama
  if (selectedKetumId.value && selectedKetupelId.value && selectedKetumId.value === selectedKetupelId.value) {
    toast({
      title: 'Validasi Tanda Tangan',
      description: 'Ketua Umum dan Penandatangan Dasar tidak boleh orang yang sama.',
      variant: 'destructive',
    })
    return
  }

  isSubmitting.value = true
  try {
    const ketum = userList.value.find((u) => u.id === selectedKetumId.value)
    const ketupel = userList.value.find((u) => u.id === selectedKetupelId.value)
    const todayTanggalSurat = idDateOnlyFormatter.format(new Date())

    const finalReplacements: Record<string, string> = {
      nomor_surat: computedNomorSurat.value,
      tanggal_surat: todayTanggalSurat,
      tanggal_pembuatan: todayTanggalSurat,
      hari_tanggal_surat: idDateFormatter.format(new Date()),
      perihal: perihal.value.trim(),
      nama_kegiatan: namaKegiatan.value.trim(),
      hari_tanggal: formattedHariTanggal.value,
      tanggal_acara: formattedHariTanggal.value,
      tanggal_pelaksanaan: formattedHariTanggal.value,
      tanggal: formattedHariTanggal.value || todayTanggalSurat,
      waktu: formattedWaktu.value,
      waktu_acara: formattedWaktu.value,
      jam_acara: formattedWaktu.value,
      jam_mulai: jamMulai.value,
      jam_selesai: isSampaiSelesai.value ? 'Selesai' : jamSelesai.value,
      tempat: tempatAcara.value.trim(),
      narahubung: formattedNarahubung.value,
      nama_narahubung: namaNarahubung.value.trim(),
      nohp_narahubung: nohpNarahubung.value.trim(),
      narahubung_nama: namaNarahubung.value.trim(),
      narahubung_nohp: nohpNarahubung.value.trim(),
      narahubung_hp: nohpNarahubung.value.trim(),
      nama_cp: namaNarahubung.value.trim(),
      nohp_cp: nohpNarahubung.value.trim(),
      no_hp: nohpNarahubung.value.trim(),
      nama_ketum: ketum?.nama || '',
      nim_ketum: ketum?.nim || '',
      nama_dasar: ketupel?.nama || '',
      nim_dasar: ketupel?.nim || '',
      nama_tingkat_1: ketupel?.nama || '',
      nim_tingkat_1: ketupel?.nim || '',
      nama_lapis_1: ketupel?.nama || '',
      nim_lapis_1: ketupel?.nim || '',
      nama_lapis1: ketupel?.nama || '',
      nim_lapis1: ketupel?.nim || '',
      nama_ketupel: ketupel?.nama || '',
      nim_ketupel: ketupel?.nim || '',
      ...dynamicReplacements.value,
    }

    const isNoProker = !selectedProkerCode.value || selectedProkerCode.value === 'NONE'
    const matchedProker = isNoProker
      ? undefined
      : prokerList.value.find(
          (p) => p.nama_proker.toUpperCase().replace(/\s+/g, '-') === selectedProkerCode.value
        )

    const res = await templateService.generate({
      template_id: selectedTemplateId.value,
      folder_url: outputFolderUrl.value.trim() || undefined,
      nomor_surat: computedNomorSurat.value,
      file_name: computedFileName.value,
      perihal: perihal.value.trim(),
      pic_nama: '-', // PIC kosong/nullable saat pertama dibuat
      proker_id: matchedProker ? matchedProker.id : undefined,
      replacements: finalReplacements,
    })

    generationResult.value = {
      surat: res.surat,
      gdocs_url: res.gdocs_url,
      pdf_url: res.pdf_url,
    }

    emit('created', res.surat)

    toast({
      title: 'Surat Berhasil Dibuat',
      description: `Dokumen "${res.surat.nomor_surat}" berhasil dibuat dan disimpan di Google Drive.`,
      variant: 'success',
    })
  } catch (err: any) {
    toast({
      title: 'Gagal Men-generate Surat',
      description: err.response?.data?.error || err.message || 'Terjadi kesalahan sistem saat memproses dokumen',
      variant: 'destructive',
    })
  } finally {
    isSubmitting.value = false
  }
}

const handleFinish = () => {
  generationResult.value = null
  perihal.value = ''
  namaKegiatan.value = ''
  tanggalMulai.value = ''
  tanggalSelesai.value = ''
  jamMulai.value = ''
  jamSelesai.value = ''
  isSampaiSelesai.value = false
  tempatAcara.value = ''
  namaNarahubung.value = ''
  nohpNarahubung.value = ''
  selectedKetupelId.value = ''
  dynamicReplacements.value = {}
  emit('close')
}

</script>

<template>
  <Dialog :open="isOpen" @update:open="emit('close')">
    <DialogContent v-if="isOpen" class="sm:max-w-3xl rounded-3xl p-0 border border-slate-200 bg-white shadow-xl max-h-[90vh] flex flex-col overflow-hidden">
      <DialogHeader class="px-6 py-4 border-b border-slate-100 shrink-0">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 rounded-full bg-orange-50 border border-orange-200/80 flex items-center justify-center text-orange-600 shrink-0">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
          </div>
          <div>
            <DialogTitle class="text-base font-semibold text-slate-900 tracking-tight">
              Buat Surat Resmi DOSCOM
            </DialogTitle>
            <DialogDescription class="text-xs text-slate-500">
              Template otomatis dengan nomor surat segmented, jadwal kegiatan, dan tanda tangan pengurus.
            </DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <!-- SUCCESS STATE: Link GDocs & PDF Instan -->
      <div v-if="generationResult" class="p-6 sm:p-8 space-y-5 text-center overflow-y-auto overscroll-contain">
        <div class="w-14 h-14 mx-auto rounded-full bg-emerald-50 border border-emerald-200 flex items-center justify-center text-emerald-600 shadow-xs">
          <svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        </div>

        <div class="space-y-1">
          <h3 class="text-base font-bold text-slate-900">Surat Berhasil Digenerate!</h3>
          <p class="text-xs text-slate-600 font-mono font-semibold">{{ generationResult.surat.nomor_surat }}</p>
          <p class="text-xs text-slate-500 font-medium">{{ generationResult.surat.perihal }}</p>
        </div>

        <div class="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
          <Button
            as-child
            class="w-full sm:w-auto px-5 py-2.5 rounded-full text-xs font-bold text-white bg-sky-600 hover:bg-sky-700 shadow-xs active:scale-[0.98] transition-all cursor-pointer"
          >
            <a :href="generationResult.gdocs_url" target="_blank" rel="noopener noreferrer" class="flex items-center justify-center gap-1.5">
              <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
              <span>Buka Google Docs</span>
            </a>
          </Button>

          <Button
            as-child
            variant="outline"
            class="w-full sm:w-auto px-5 py-2.5 rounded-full text-xs font-bold text-slate-700 bg-white border border-slate-200 hover:bg-slate-50 shadow-xs active:scale-[0.98] transition-all cursor-pointer"
          >
            <a :href="generationResult.pdf_url" target="_blank" rel="noopener noreferrer" class="flex items-center justify-center gap-1.5">
              <svg class="w-3.5 h-3.5 text-rose-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
              <span>Download PDF</span>
            </a>
          </Button>
        </div>

        <div class="pt-4 border-t border-black/5 flex justify-center">
          <Button
            type="button"
            variant="secondary"
            @click="handleFinish"
            class="rounded-full px-6 text-xs font-semibold"
          >
            Selesai & Masuk ke Laci
          </Button>
        </div>
      </div>

      <!-- FORM STATE -->
      <form v-else @submit.prevent="handleSubmit" class="flex flex-col flex-1 min-h-0 overflow-hidden">
        <div class="flex-1 overflow-y-auto px-6 py-4 space-y-4 overscroll-contain [transform:translateZ(0)] [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden">
        <!-- ALERT KETUM: Muncul jika data Ketua Umum belum ada di data users -->
        <div
          v-if="!isLoadingData && !ketumUser"
          role="alert"
          class="p-3.5 rounded-2xl bg-amber-50 border border-amber-200 text-amber-900 flex items-start gap-2.5 text-xs"
        >
          <svg class="w-4 h-4 text-amber-600 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div class="space-y-0.5">
            <p class="font-bold">Data Ketua Umum Belum Ditemukan</p>
            <p class="text-[11px] text-amber-800 leading-relaxed">
              Belum ada akun pengurus dengan divisi <b>"Ketua Umum"</b> di sistem. Silakan tambahkan atau ubah divisi akun di menu <b>Pengurus</b> agar tanda tangan Ketua Umum terisi otomatis.
            </p>
          </div>
        </div>

        <!-- BLOK 1: TEMPLATE & NOMOR SURAT SEGMENTED (SEPERTI VERSI AWAL) -->
        <div class="space-y-3 bg-slate-50 p-4 rounded-2xl border border-slate-200">
          <div class="flex items-center justify-between">
            <Label class="text-xs font-semibold text-slate-800">
              1. Pilih Template & Nomor Surat
            </Label>
            <Badge variant="outline" class="text-[10px] bg-white text-slate-600 border-slate-200">
              Segmented Input
            </Badge>
          </div>

          <!-- Pilihan Template -->
          <div class="space-y-1">
            <Select v-model="selectedTemplateId">
              <SelectTrigger class="h-10 text-xs rounded-2xl bg-white border-zinc-200">
                <SelectValue placeholder="Pilih jenis surat..." />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="tpl in templates"
                  :key="tpl.id"
                  :value="tpl.id"
                  class="text-xs cursor-pointer"
                >
                  {{ tpl.nama_template }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          <!-- Format Nomor Surat Terpisah-pisah -->
          <div class="space-y-1.5 pt-1">
            <Label class="text-[11px] font-semibold text-zinc-500 uppercase tracking-wider block">
              Format Nomor Surat
            </Label>

            <div class="flex flex-wrap items-center gap-1.5 p-2 rounded-2xl bg-white border border-slate-200/90 shadow-2xs">
              <!-- No Urut (Auto dari prefix nama template, editable) -->
              <Input
                v-model="noUrut"
                type="text"
                placeholder="01"
                class="w-14 h-8 text-xs font-mono font-bold text-center rounded-xl bg-zinc-50 border-zinc-200"
                title="Nomor Urut Surat (diambil dari awalan nama template)"
              />

              <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

              <!-- Static DOSCOM Badge -->
              <span class="px-2.5 py-1 text-[11px] font-bold text-sky-700 bg-sky-50 rounded-xl border border-sky-200 select-none font-mono">
                DOSCOM
              </span>

              <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

              <!-- Pilihan Proker (Murni dari tabel proker atau Kosong) -->
              <div class="min-w-[130px] flex-1">
                <Select v-model="selectedProkerCode">
                  <SelectTrigger class="h-8 text-xs rounded-xl bg-zinc-50 border-zinc-200">
                    <SelectValue placeholder="Pilih Proker (atau Kosong)..." />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="NONE">Tanpa Proker (Kosong)</SelectItem>
                    <SelectItem
                      v-for="proker in prokerList"
                      :key="proker.id"
                      :value="proker.nama_proker.toUpperCase().replace(/\s+/g, '-')"
                    >
                      {{ proker.nama_proker }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

              <!-- Pilihan Bulan Romawi -->
              <div class="w-16">
                <Select v-model="selectedBulan">
                  <SelectTrigger class="h-8 text-xs font-mono font-semibold rounded-xl bg-zinc-50 border-zinc-200">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="b in romanMonths" :key="b" :value="b">
                      {{ b }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <span class="text-zinc-400 font-semibold text-xs select-none">/</span>

              <!-- Input Tahun -->
              <Input
                v-model="selectedTahun"
                type="text"
                placeholder="2026"
                class="w-18 h-8 text-xs font-mono font-semibold text-center rounded-xl bg-zinc-50 border-zinc-200"
              />
            </div>

            <!-- Live Preview: Nomor & Nama File -->
            <div class="flex flex-col sm:flex-row sm:items-center gap-1.5 sm:gap-4 pt-0.5 text-xs text-slate-500">
              <div class="flex items-center gap-2">
                <span class="text-[11px] text-slate-400">Preview Nomor:</span>
                <span class="font-mono font-bold text-sky-700 bg-sky-50 px-2 py-0.5 rounded-lg border border-sky-100">
                  {{ computedNomorSurat }}
                </span>
              </div>
              <div class="flex items-center gap-2 min-w-0">
                <span class="text-[11px] text-slate-400 shrink-0">Nama File:</span>
                <span
                  class="font-mono text-[11px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-lg border border-emerald-100 truncate"
                  :title="computedFileName"
                >
                  {{ computedFileName }}
                </span>
              </div>
            </div>
          </div>

          <!-- Perihal Surat (Ditujukan Kepada ditiadakan karena sudah statis di template) -->
          <div class="space-y-1 pt-1">
            <Label for="perihal" class="text-xs font-semibold text-slate-800">
              Perihal Surat <span class="text-rose-500">*</span>
            </Label>
            <Input
              id="perihal"
              v-model="perihal"
              type="text"
              required
              placeholder="Contoh: Undangan Menghadiri Acara"
              class="h-9 text-xs rounded-xl bg-white border-zinc-200"
            />
          </div>
        </div>

        <!-- BLOK 2: DETAIL KEGIATAN & WAKTU ACARA -->
        <div class="space-y-3 bg-slate-50 p-4 rounded-2xl border border-slate-200">
          <Label class="text-xs font-semibold text-slate-800 block">
            2. Detail Jadwal Pelaksanaan / Acara / Peminjaman
          </Label>

          <div class="space-y-1">
            <Label for="nama-kegiatan" class="text-xs font-medium text-slate-700">Nama Kegiatan / Agenda</Label>
            <Input
              id="nama-kegiatan"
              v-model="namaKegiatan"
              type="text"
              placeholder="Contoh: Penutupan DOSCOM University 2026"
              class="h-9 text-xs rounded-xl bg-white border-zinc-200"
            />
          </div>

          <!-- Opsi Tanggal: 1 Hari vs Rentang Hari -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <Label class="text-xs font-medium text-slate-700">Hari & Tanggal Pelaksanaan</Label>
              <div class="flex items-center gap-1 bg-white p-0.5 rounded-full border border-slate-200">
                <button
                  type="button"
                  @click="tipeTanggal = 'single'"
                  class="px-2.5 py-0.5 rounded-full text-[10px] font-semibold transition-all cursor-pointer"
                  :class="tipeTanggal === 'single' ? 'bg-sky-600 text-white shadow-2xs' : 'text-slate-500 hover:text-slate-800'"
                >
                  1 Hari
                </button>
                <button
                  type="button"
                  @click="tipeTanggal = 'range'"
                  class="px-2.5 py-0.5 rounded-full text-[10px] font-semibold transition-all cursor-pointer"
                  :class="tipeTanggal === 'range' ? 'bg-sky-600 text-white shadow-2xs' : 'text-slate-500 hover:text-slate-800'"
                >
                  Rentang Hari (s.d.)
                </button>
              </div>
            </div>

            <!-- Single Date Input -->
            <div v-if="tipeTanggal === 'single'" class="space-y-1">
              <Input
                v-model="tanggalMulai"
                type="date"
                class="h-9 text-xs rounded-xl bg-white border-zinc-200 cursor-pointer [&::-webkit-calendar-picker-indicator]:cursor-pointer"
                @click="handlePickerClick"
              />
            </div>

            <!-- Range Date Inputs -->
            <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <div class="space-y-0.5">
                <span class="text-[10px] text-slate-400">Tanggal Mulai:</span>
                <Input
                  v-model="tanggalMulai"
                  type="date"
                  class="h-9 text-xs rounded-xl bg-white border-zinc-200 cursor-pointer [&::-webkit-calendar-picker-indicator]:cursor-pointer"
                  @click="handlePickerClick"
                />
              </div>
              <div class="space-y-0.5">
                <span class="text-[10px] text-slate-400">Tanggal Selesai:</span>
                <Input
                  v-model="tanggalSelesai"
                  type="date"
                  class="h-9 text-xs rounded-xl bg-white border-zinc-200 cursor-pointer [&::-webkit-calendar-picker-indicator]:cursor-pointer"
                  @click="handlePickerClick"
                />
              </div>
            </div>

            <p v-if="formattedHariTanggal" class="text-[11px] text-sky-700 font-medium">
              Format di surat: <span class="font-semibold">{{ formattedHariTanggal }}</span>
            </p>
          </div>

          <!-- Waktu, Tempat, dan Narahubung (2 Input Terpisah) -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <!-- Waktu Acara (Jam Mulai & Jam Selesai / Sampai Selesai) -->
            <div class="space-y-2 col-span-1 sm:col-span-2 p-3 bg-white rounded-xl border border-slate-200">
              <div class="flex items-center justify-between">
                <Label class="text-xs font-medium text-slate-700">Waktu Acara</Label>
                <label class="flex items-center gap-1.5 cursor-pointer text-xs text-slate-600 select-none">
                  <input
                    type="checkbox"
                    v-model="isSampaiSelesai"
                    class="rounded border-slate-300 text-sky-600 focus:ring-sky-500 h-3.5 w-3.5 cursor-pointer"
                  />
                  Sampai Selesai
                </label>
              </div>
              <div class="grid grid-cols-2 gap-3">
                <div class="space-y-1">
                  <span class="text-[11px] text-slate-500 block">Jam Mulai</span>
                  <Input
                    type="time"
                    v-model="jamMulai"
                    class="h-9 text-xs rounded-xl bg-white border-zinc-200 cursor-pointer [&::-webkit-calendar-picker-indicator]:cursor-pointer"
                    @click="handlePickerClick"
                  />
                </div>
                <div class="space-y-1">
                  <span class="text-[11px] text-slate-500 block">Jam Selesai</span>
                  <Input
                    type="time"
                    v-model="jamSelesai"
                    :disabled="isSampaiSelesai"
                    placeholder="Selesai"
                    class="h-9 text-xs rounded-xl bg-white border-zinc-200 disabled:bg-slate-100 disabled:text-slate-400 cursor-pointer [&::-webkit-calendar-picker-indicator]:cursor-pointer"
                    @click="handlePickerClick"
                  />
                </div>
              </div>
              <p v-if="formattedWaktu" class="text-[11px] text-sky-700 font-medium">
                Format di surat: <span class="font-semibold">{{ formattedWaktu }}</span>
              </p>
            </div>

            <div class="space-y-1 col-span-1 sm:col-span-2">
              <Label for="tempat" class="text-xs font-medium text-slate-700">Tempat Acara</Label>
              <Input
                id="tempat"
                v-model="tempatAcara"
                type="text"
                placeholder="Contoh: Gedung H Lt. 7 UDINUS"
                class="h-9 text-xs rounded-xl bg-white border-zinc-200"
              />
            </div>

            <!-- Narahubung Input 1: Nama -->
            <div class="space-y-1">
              <Label for="nama-narahubung" class="text-xs font-medium text-slate-700">Nama Narahubung (CP)</Label>
              <Input
                id="nama-narahubung"
                v-model="namaNarahubung"
                type="text"
                placeholder="Contoh: Saudara Budi"
                class="h-9 text-xs rounded-xl bg-white border-zinc-200"
              />
            </div>

            <!-- Narahubung Input 2: No HP / WA -->
            <div class="space-y-1">
              <Label for="nohp-narahubung" class="text-xs font-medium text-slate-700">No. HP / WhatsApp</Label>
              <Input
                id="nohp-narahubung"
                v-model="nohpNarahubung"
                type="tel"
                placeholder="Contoh: 081234567890"
                class="h-9 text-xs rounded-xl bg-white border-zinc-200"
              />
            </div>
          </div>
        </div>

        <!-- BLOK 3: TANDA TANGAN (KETUM DI KIRI, KETUPEL DI KANAN) -->
        <div class="space-y-3 bg-slate-50 p-4 rounded-2xl border border-slate-200">
          <div class="flex items-center justify-between">
            <Label class="text-xs font-semibold text-slate-800">
              3. Penandatangan Surat (Pengurus)
            </Label>
            <span class="text-[10px] text-slate-400">Pembina & WR III statis di template</span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
            <!-- TTD 1: KETUA UMUM (POSISI KIRI) -->
            <div class="space-y-1">
              <Label class="text-xs font-medium text-slate-700 flex items-center gap-1.5">
                <span>Ketua Umum</span>
                <span v-if="ketumUser" class="text-[10px] text-emerald-600 font-semibold font-mono">(Terdeteksi)</span>
              </Label>
              <Select v-model="selectedKetumId">
                <SelectTrigger class="h-9 text-xs rounded-xl bg-white border-zinc-200">
                  <SelectValue placeholder="Pilih Ketua Umum..." />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="u in ketumOptions"
                    :key="u.id"
                    :value="u.id"
                    class="text-xs cursor-pointer"
                  >
                    {{ u.nama }} ({{ u.nim || u.username }})
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            <!-- TTD 2: KETUA PELAKSANA / PEMBUAT (POSISI KANAN) -->
            <div class="space-y-1">
              <Label class="text-xs font-medium text-slate-700">
                Penandatangan Dasar
              </Label>
              <Select v-model="selectedKetupelId">
                <SelectTrigger class="h-9 text-xs rounded-xl bg-white border-zinc-200">
                  <SelectValue placeholder="Pilih Penandatangan Dasar..." />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem
                    v-for="u in lapis1Users"
                    :key="u.id"
                    :value="u.id"
                    class="text-xs cursor-pointer"
                  >
                    {{ u.nama }} ({{ u.nim || u.username }}){{ ['dasar', 'pengurus', 'tupel'].includes(u.divisi?.toLowerCase() || '') ? ' • Dasar' : '' }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </div>

        <!-- BLOK 4: VARIABEL EKSTRA (JIKA ADA PLACEHOLDER LAIN DI DOCS) -->
        <div v-if="extraPlaceholders.length > 0" class="space-y-2 p-4 bg-slate-50 rounded-2xl border border-slate-200">
          <Label class="text-xs font-semibold text-slate-800 block">
            Variabel Tambahan Template
          </Label>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div v-for="phKey in extraPlaceholders" :key="phKey" class="space-y-1">
              <Label :for="'ph-' + phKey" class="text-[11px] font-medium text-slate-700">
                {{ phKey.replace(/_/g, ' ').toUpperCase() }}
              </Label>
              <Input
                :id="'ph-' + phKey"
                v-model="dynamicReplacements[phKey]"
                type="text"
                :placeholder="'Isi ' + phKey"
                class="h-8 text-xs rounded-xl bg-white border-zinc-200"
              />
            </div>
          </div>
        </div>
      </div>

      <!-- STICKY FOOTER -->
      <DialogFooter class="px-6 py-3.5 border-t border-slate-100 bg-white flex items-center justify-between gap-2 shrink-0">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          @click="emit('close')"
          class="rounded-full px-4 text-xs font-medium text-slate-500 hover:text-slate-800"
        >
          Batal
        </Button>

        <Button
          type="submit"
          :disabled="isSubmitting || templates.length === 0"
          class="px-6 py-2.5 rounded-full text-xs font-bold text-white bg-sky-600 hover:bg-sky-700 active:scale-[0.98] shadow-xs transition-all cursor-pointer disabled:opacity-50"
        >
          <svg
            v-if="isSubmitting"
            class="w-3.5 h-3.5 mr-1.5 animate-spin"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>{{ isSubmitting ? 'Memproses ke GDrive...' : 'Generate Surat (GDocs & PDF) →' }}</span>
        </Button>
      </DialogFooter>
    </form>
    </DialogContent>
  </Dialog>
</template>
