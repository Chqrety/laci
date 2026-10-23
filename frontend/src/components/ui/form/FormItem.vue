<script setup lang="ts">
import { type HTMLAttributes, provide, useId } from 'vue'
import { cn } from '@/lib/utils'
import { FORM_ITEM_INJECTION_KEY } from './useFormField'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()

const id = useId()
provide(FORM_ITEM_INJECTION_KEY, id)
</script>

<template>
  <div :class="cn('space-y-1', props.class)">
    <slot />
  </div>
</template>
