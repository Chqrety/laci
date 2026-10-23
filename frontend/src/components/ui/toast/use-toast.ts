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

const TOAST_LIMIT = 5
const TOAST_REMOVE_DELAY = 1000

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

function addToast(toast: Omit<ToasterToast, 'id' | 'open'>) {
  const id = genId()

  const update = (props: ToasterToast) => {
    toasts.value = toasts.value.map((t) => (t.id === id ? { ...t, ...props } : t))
  }

  const dismiss = () => {
    toasts.value = toasts.value.map((t) => (t.id === id ? { ...t, open: false } : t))
    setTimeout(() => {
      toasts.value = toasts.value.filter((t) => t.id !== id)
    }, TOAST_REMOVE_DELAY)
  }

  toasts.value = [
    {
      ...toast,
      id,
      open: true,
      onOpenChange: (open: boolean) => {
        if (!open) dismiss()
      },
    },
    ...toasts.value,
  ].slice(0, TOAST_LIMIT)

  return {
    id,
    dismiss,
    update,
  }
}

export function useToast() {
  return {
    toasts: computed(() => toasts.value),
    toast: addToast,
    dismiss: (toastId?: string) => {
      if (toastId) {
        toasts.value = toasts.value.map((t) => (t.id === toastId ? { ...t, open: false } : t))
      } else {
        toasts.value = toasts.value.map((t) => ({ ...t, open: false }))
      }
    },
  }
}

export const toast = addToast
