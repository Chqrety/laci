import { ref } from 'vue'


export interface DocumentScanResult {
  rawText: string
  nomorSurat: string
  perihal: string
  namaKegiatan: string
  tujuanSurat: string
  noUrut: string
  prokerCode: string
  bulanRomawi: string
  tahun: string
}

export function parseDoscomNomor(text: string) {
  if (!text) return { noUrut: '', prokerCode: '', bulan: '', tahun: '', full: '' }

  // Normalisasi karakter slash dan variasinya
  const normalized = text
    .replace(/[|\\]/g, '/')
    .replace(/\s*\/\s*/g, '/')

  // 1. Pola standar nomor surat DOSCOM: [no]/DOSCOM/[proker]/[romawi]/[tahun]
  const pattern = /(\d{1,4})\/D[O0]SC[O0]M\/([^\/\r\n]+)\/([IVXLCDMivxlcdm]+)\/(\d{4})/i
  const match = normalized.match(pattern)

  if (match) {
    const no = (match[1] ?? '').padStart(2, '0')
    const proker = (match[2] ?? '').trim().toUpperCase()
    const bulan = (match[3] ?? '').toUpperCase()
    const tahun = match[4] ?? ''
    return {
      noUrut: no,
      prokerCode: proker,
      bulan,
      tahun,
      full: `${no}/DOSCOM/${proker}/${bulan}/${tahun}`,
    }
  }

  // 2. Fallback: Cari di baris label Nomor / No
  const labelMatch = normalized.match(/(?:Nomor|No|N0|Nomer)\.?\s*[:;=.-]?\s*([^\r\n]+)/i)
  if (labelMatch) {
    const line = (labelMatch[1] ?? '').replace(/\s+(?:Lampiran|Lamp|Hal|Perihal|Sifat|Kepada)\b.*$/i, '').trim()
    const parts = line.split('/').map((s) => s.trim())

    let no = ''
    let proker = ''
    let bulan = ''
    let tahun = ''

    if (parts.length >= 3) {
      const numMatch = parts[0]?.match(/\d+/)
      if (numMatch) no = numMatch[0].padStart(2, '0')

      if (parts[1]?.toUpperCase() === 'DOSCOM') {
        proker = parts[2]?.toUpperCase() || ''
      } else {
        proker = parts[1]?.toUpperCase() || ''
      }

      for (let i = 2; i < parts.length; i++) {
        const pt = parts[i] || ''
        if (/^[IVXLCDM]+$/i.test(pt)) bulan = pt.toUpperCase()
        if (/^20\d{2}$/.test(pt)) tahun = pt
      }

      return {
        noUrut: no,
        prokerCode: proker,
        bulan,
        tahun,
        full: line,
      }
    }
  }

  return { noUrut: '', prokerCode: '', bulan: '', tahun: '', full: '' }
}

export function extractPerihal(text: string): string {
  if (!text) return ''
  const lines = text.split(/\r?\n/).map((l) => l.trim()).filter(Boolean)
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    if (!line) continue

    const labelMatch = line.match(/^(?:Perihal|Hal|HAL|PERIHAL)\.?\s*[:;=.-]?\s*(.*)$/i)
    if (labelMatch) {
      let candidate = (labelMatch[1] ?? '').trim()
      if (!candidate && i + 1 < lines.length) {
        candidate = (lines[i + 1] ?? '').trim()
      }
      if (candidate) {
        candidate = candidate.replace(/\s+(?:Kepada|Yth|Lampiran|Sifat)\b.*$/i, '').trim()
        candidate = candidate.replace(/^[:;=.\s-]+/, '').trim()
        if (candidate) return candidate
      }
    }
  }
  return ''
}

export function extractNamaKegiatan(text: string): string {
  if (!text) return ''
  const match = text.match(/kegiatan\s+(.+?)\s+yang\s+(?:akan\s+)?diselenggarakan/is)
  return match?.[1] ? match[1].replace(/\s+/g, ' ').trim() : ''
}

export function extractTujuanSurat(text: string): string {
  if (!text) return ''
  const match = text.match(/Yth\.?\s*(.+?)(?:\r?\n|$)/i)
  return match?.[1] ? match[1].trim() : ''
}

/**
 * Preprocessing foto kamera HP sebelum OCR:
 * 1. Downscale resolusi raksasa (12-48MP) ke max 1500px agar Tesseract WASM ringan dan super cepat.
 * 2. Grayscale + kontras agar teks hitam pada kertas putih terbaca optimal.
 */
async function preprocessImageForOcr(file: File | Blob, maxDimension = 1500): Promise<Blob> {
  return new Promise((resolve) => {
    const img = new Image()
    const url = URL.createObjectURL(file)
    img.onload = () => {
      URL.revokeObjectURL(url)
      let { width, height } = img

      if (width > maxDimension || height > maxDimension) {
        if (width > height) {
          height = Math.round((height * maxDimension) / width)
          width = maxDimension
        } else {
          width = Math.round((width * maxDimension) / height)
          height = maxDimension
        }
      }

      const canvas = document.createElement('canvas')
      canvas.width = width
      canvas.height = height
      const ctx = canvas.getContext('2d')
      if (!ctx) {
        resolve(file)
        return
      }

      ctx.drawImage(img, 0, 0, width, height)
      const imgData = ctx.getImageData(0, 0, width, height)
      const d = imgData.data
      for (let i = 0; i < d.length; i += 4) {
        const r = d[i] ?? 0
        const g = d[i + 1] ?? 0
        const b = d[i + 2] ?? 0
        const gray = 0.299 * r + 0.587 * g + 0.114 * b
        const contrast = gray < 130 ? Math.max(0, gray - 25) : Math.min(255, gray + 25)
        d[i] = contrast
        d[i + 1] = contrast
        d[i + 2] = contrast
      }
      ctx.putImageData(imgData, 0, 0)

      canvas.toBlob((blob) => resolve(blob || file), 'image/jpeg', 0.85)
    }
    img.onerror = () => {
      URL.revokeObjectURL(url)
      resolve(file)
    }
    img.src = url
  })
}

export function useDocumentScanner() {
  const isScanning = ref(false)
  const scanProgress = ref(0)
  const scanResult = ref<DocumentScanResult | null>(null)

  const scanImage = async (file: File | Blob): Promise<DocumentScanResult> => {
    isScanning.value = true
    scanProgress.value = 0
    scanResult.value = null

    let worker: any = null
    try {
      const optimizedFile = await preprocessImageForOcr(file)
      const { createWorker } = await import('tesseract.js')
      worker = await createWorker('ind', 1, {
        logger: (m: { status?: string; progress?: number }) => {
          if (m?.status === 'recognizing text' && typeof m.progress === 'number') {
            scanProgress.value = Math.min(100, Math.round(m.progress * 100))
          }
        },
      })

      const { data } = await worker.recognize(optimizedFile)
      const text = data?.text || ''
      const doscomParsed = parseDoscomNomor(text)
      const perihal = extractPerihal(text)
      const namaKegiatan = extractNamaKegiatan(text)
      const tujuanSurat = extractTujuanSurat(text)

      const extracted: DocumentScanResult = {
        rawText: text,
        nomorSurat: doscomParsed.full,
        perihal,
        namaKegiatan,
        tujuanSurat,
        noUrut: doscomParsed.noUrut,
        prokerCode: doscomParsed.prokerCode,
        bulanRomawi: doscomParsed.bulan,
        tahun: doscomParsed.tahun,
      }

      scanProgress.value = 100
      scanResult.value = extracted
      return extracted
    } finally {
      if (worker) {
        await worker.terminate()
      }
      isScanning.value = false
    }
  }

  return {
    isScanning,
    scanProgress,
    scanResult,
    scanImage,
  }
}
