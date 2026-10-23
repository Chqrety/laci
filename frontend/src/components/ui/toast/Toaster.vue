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
      <div class="grid gap-1">
        <ToastTitle v-if="toast.title">
          {{ toast.title }}
        </ToastTitle>
        <ToastDescription v-if="toast.description">
          {{ toast.description }}
        </ToastDescription>
      </div>
      <ToastClose @click="dismiss(toast.id)" />
    </Toast>
    <ToastViewport />
  </ToastProvider>
</template>
