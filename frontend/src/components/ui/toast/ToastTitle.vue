<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { ToastTitle, type ToastTitleProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<ToastTitleProps & { class?: HTMLAttributes['class'] }>()
</script>

<template>
  <ToastTitle :class="cn('text-xs font-semibold leading-snug', props.class)">
    <slot />
  </ToastTitle>
</template>
