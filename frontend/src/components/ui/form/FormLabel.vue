<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { useFormField } from './useFormField'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()

const { error, formItemId } = useFormField()
</script>

<template>
  <Label
    :for="formItemId"
    :class="cn(
      'block text-xs font-medium text-zinc-700 mb-1',
      error && 'text-rose-600',
      props.class
    )"
  >
    <slot />
  </Label>
</template>
