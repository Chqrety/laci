import { ref } from 'vue'
import Tesseract from 'tesseract.js'

export interface DocumentScanResult {
  rawText: string
  nomorSurat: string
  perihal: string
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

export function useDocumentScanner() {
  const isScanning = ref(false)
  const scanProgress = ref(0)
  const scanResult = ref<DocumentScanResult | null>(null)

  const scanImage = async (file: File | Blob): Promise<DocumentScanResult> => {
    isScanning.value = true
    scanProgress.value = 0
    scanResult.value = null

    try {
      const { data } = await Tesseract.recognize(file, 'ind+eng', {
        logger: (m) => {
          if (m?.status === 'recognizing text' && typeof m.progress === 'number') {
            scanProgress.value = Math.min(100, Math.round(m.progress * 100))
          }
        },
      })

      const text = data?.text || ''
      const doscomParsed = parseDoscomNomor(text)
      const perihal = extractPerihal(text)

      const extracted: DocumentScanResult = {
        rawText: text,
        nomorSurat: doscomParsed.full,
        perihal,
        noUrut: doscomParsed.noUrut,
        prokerCode: doscomParsed.prokerCode,
        bulanRomawi: doscomParsed.bulan,
        tahun: doscomParsed.tahun,
      }

      scanProgress.value = 100
      scanResult.value = extracted
      return extracted
    } finally {
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
