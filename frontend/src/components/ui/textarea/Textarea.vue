<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
  defaultValue?: string | number
  modelValue?: string | number
}>()

const emits = defineEmits<{
  (e: 'update:modelValue', payload: string | number): void
}>()
</script>

<template>
  <textarea
    :value="props.modelValue ?? props.defaultValue"
    :class="
      cn(
        'flex min-h-[60px] w-full rounded-2xl border border-zinc-200 bg-zinc-50 px-3.5 py-2.5 text-xs text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC] disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-180 resize-none',
        props.class,
      )
    "
    @input="emits('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
  />
</template>
