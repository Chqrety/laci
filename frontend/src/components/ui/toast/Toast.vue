<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { ToastRoot, type ToastRootEmits, type ToastRootProps, useForwardPropsEmits } from 'radix-vue'
import { type ToastVariants, toastVariants } from '.'
import { cn } from '@/lib/utils'

interface Props extends ToastRootProps {
  class?: HTMLAttributes['class']
  variant?: ToastVariants['variant']
}

const props = withDefaults(defineProps<Props>(), {
  duration: 4000,
})
const emits = defineEmits<ToastRootEmits>()

const forwarded = useForwardPropsEmits(props, emits)
</script>

<template>
  <ToastRoot
    v-bind="forwarded"
    :class="cn(toastVariants({ variant }), props.class)"
  >
    <slot />
  </ToastRoot>
</template>
