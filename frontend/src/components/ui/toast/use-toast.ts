import { computed, ref } from 'vue'

export interface ToastProps {
  id: string
  title?: string
  description?: string
  variant?: 'default' | 'destructive' | 'success'
  duration?: number
  open?: boolean
  onOpenChange?: (open: boolean) => void
}

const TOAST_REMOVE_DELAY = 300
const DEFAULT_DURATION = 4000 // 4 detik agar user nyaman membaca notifikasi

type ToasterToast = ToastProps & {
  id: string
  title?: string
  description?: string
  open: boolean
}

let count = 0

function genId() {
  count = (count + 1) % Number.MAX_VALUE
  return count.toString()
}

const toasts = ref<ToasterToast[]>([])
const toastTimers = new Map<string, ReturnType<typeof setTimeout>>()

function dismissToast(toastId?: string) {
  if (toastId) {
    if (toastTimers.has(toastId)) {
      clearTimeout(toastTimers.get(toastId)!)
      toastTimers.delete(toastId)
    }
    toasts.value = toasts.value.map((t) => (t.id === toastId ? { ...t, open: false } : t))
    setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== toastId)
    }, TOAST_REMOVE_DELAY)
  } else {
    toastTimers.forEach((timer) => clearTimeout(timer))
    toastTimers.clear()
    toasts.value = []
  }
}

function resolveVariant(toast: Partial<ToastProps>): 'default' | 'destructive' | 'success' {
  if (toast.variant) return toast.variant
  const text = `${toast.title || ''} ${toast.description || ''}`.toLowerCase()
  if (
    text.includes('gagal') ||
    text.includes('error') ||
    text.includes('wajib') ||
    text.includes('tidak cocok') ||
    text.includes('salah')
  ) {
    return 'destructive'
  }
  if (text.includes('berhasil') || text.includes('sukses') || text.includes('success')) {
    return 'success'
  }
  return 'default'
}

function addToast(toast: Omit<ToasterToast, 'id' | 'open'>) {
  const id = genId()
  const duration = toast.duration ?? DEFAULT_DURATION
  const variant = resolveVariant(toast)

  // Bersihkan semua timer aktif dari toast lama tanpa menunda penghapusan
  toastTimers.forEach((timer) => clearTimeout(timer))
  toastTimers.clear()

  const dismiss = () => dismissToast(id)

  const newToast: ToasterToast = {
    ...toast,
    id,
    variant,
    open: true,
    duration,
    onOpenChange: (open: boolean) => {
      if (!open) dismiss()
    },
  }

  // Gantikan langsung dengan toast baru (anti-stacking)
  toasts.value = [newToast]

  // Timer otomatis pasti dismiss setelah durasi 4000ms penuh
  if (duration > 0) {
    const timer = setTimeout(() => {
      dismissToast(id)
    }, duration)
    toastTimers.set(id, timer)
  }

  return {
    id,
    dismiss,
    update: (props: ToasterToast) => {
      toasts.value = toasts.value.map((t) => (t.id === id ? { ...t, ...props } : t))
    },
  }
}

export function useToast() {
  return {
    toasts: computed(() => toasts.value),
    toast: addToast,
    dismiss: dismissToast,
  }
}

export const toast = addToast
