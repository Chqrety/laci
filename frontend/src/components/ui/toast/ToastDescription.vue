<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { ToastDescription, type ToastDescriptionProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<ToastDescriptionProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <ToastDescription :class="cn('text-[11px] opacity-90 mt-0.5 leading-relaxed', props.class)">
    <slot />
  </ToastDescription>
</template>
