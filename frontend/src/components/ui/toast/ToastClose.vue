<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { ToastClose, type ToastCloseProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<ToastCloseProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <ToastClose
    v-bind="props"
    :class="
      cn(
        'absolute right-2 top-2 rounded-xl p-1 text-zinc-400 opacity-70 transition-opacity hover:text-zinc-900 hover:opacity-100 focus:opacity-100 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 cursor-pointer',
        props.class,
      )
    "
  >
    <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
    </svg>
  </ToastClose>
</template>
