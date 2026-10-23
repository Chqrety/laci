<script setup lang="ts">
import { useToast } from './use-toast'
import Toast from './Toast.vue'
import ToastTitle from './ToastTitle.vue'
import ToastDescription from './ToastDescription.vue'
import ToastClose from './ToastClose.vue'
import ToastViewport from './ToastViewport.vue'
import ToastProvider from './ToastProvider.vue'

const { toasts, dismiss } = useToast()
</script>

<template>
  <ToastProvider>
    <Toast
      v-for="toast in toasts"
      :key="toast.id"
      :variant="toast.variant"
      :open="toast.open"
      :duration="toast.duration ?? 4000"
      @update:open="(val: boolean) => {
        if (!val) {
          dismiss(toast.id)
        }
      }"
    >
      <div class="flex items-start gap-3 w-full pr-2">
        <!-- Visual Icon Cue -->
        <div
          v-if="toast.variant === 'success'"
          class="shrink-0 w-6 h-6 rounded-full bg-emerald-100/90 text-emerald-700 flex items-center justify-center mt-0.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </div>

        <div
          v-else-if="toast.variant === 'destructive'"
          class="shrink-0 w-6 h-6 rounded-full bg-rose-100/90 text-rose-700 flex items-center justify-center mt-0.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>

        <div
          v-else
          class="shrink-0 w-6 h-6 rounded-full bg-sky-100/90 text-sky-700 flex items-center justify-center mt-0.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>

        <div class="grid gap-0.5 flex-1 min-w-0">
          <ToastTitle v-if="toast.title">
            {{ toast.title }}
          </ToastTitle>
          <ToastDescription v-if="toast.description">
            {{ toast.description }}
          </ToastDescription>
        </div>
      </div>
      <ToastClose @click="dismiss(toast.id)" />
    </Toast>
    <ToastViewport />
  </ToastProvider>
</template>
