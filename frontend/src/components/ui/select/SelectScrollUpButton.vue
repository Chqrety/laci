<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { SelectScrollUpButton, type SelectScrollUpButtonProps, useForwardProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<SelectScrollUpButtonProps & { class?: HTMLAttributes['class'] }>()
const forwarded = useForwardProps(props)
</script>

<template>
  <SelectScrollUpButton
    v-bind="forwarded"
    :class="cn('flex cursor-default items-center justify-center py-1 text-zinc-500', props.class)"
  >
    <slot>
      <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7" />
      </svg>
    </slot>
  </SelectScrollUpButton>
</template>
