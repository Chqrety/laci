<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  defaultValue?: string | number
  class?: HTMLAttributes['class']
}>()

const modelValue = defineModel<string | number>()
</script>

<template>
  <input
    v-model="modelValue"
    :class="
      cn(
        'flex w-full px-3.5 py-2.5 rounded-2xl bg-zinc-50 border border-zinc-200 text-xs text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-[#0A84DC]/30 focus:border-[#0A84DC] disabled:cursor-not-allowed disabled:opacity-50 transition-all duration-180 file:border-0 file:bg-transparent file:text-xs file:font-medium',
        props.class,
      )
    "
  />
</template>
