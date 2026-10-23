<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { ToastAction, type ToastActionProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<ToastActionProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <ToastAction
    v-bind="props"
    :class="
      cn(
        'inline-flex h-8 shrink-0 items-center justify-center rounded-full border border-slate-200 bg-transparent px-3 text-xs font-semibold transition-colors hover:bg-slate-100 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 disabled:pointer-events-none disabled:opacity-50 cursor-pointer',
        props.class,
      )
    "
  >
    <slot />
  </ToastAction>
</template>
