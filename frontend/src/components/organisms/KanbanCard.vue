<script setup lang="ts">
import type { Surat } from '@/services/api'
import { Badge } from '@/components/ui/badge'

defineProps<{
  surat: Surat
  isDragging?: boolean
}>()

defineEmits<{
  (e: 'click', surat: Surat): void
}>()
</script>

<template>
  <article
    draggable="true"
    style="-webkit-touch-callout: none;"
    :class="[
      'touch-pan-y select-none bg-white p-4 rounded-2xl border border-slate-200/90 shadow-2xs hover:border-sky-400 hover:shadow-xs hover:-translate-y-0.5 cursor-grab active:cursor-grabbing transition-all duration-200 space-y-2.5 active:scale-[0.98]',
      isDragging ? 'opacity-40 ring-2 ring-sky-400/50' : ''
    ]"
    @click="$emit('click', surat)"
  >
    <div class="flex items-start justify-between gap-1">
      <Badge variant="mono" size="sm">
        {{ surat.nomor_surat }}
      </Badge>
      <Badge
        v-if="surat.arsip_url"
        variant="default"
        size="xs"
        title="Arsip telah dipindai"
      >
        Arsip
      </Badge>
    </div>

    <h4 class="text-xs font-semibold text-zinc-900 leading-snug line-clamp-2">
      {{ surat.perihal }}
    </h4>

    <p v-if="surat.tujuan" class="text-xs text-gray-500 line-clamp-2">
      {{ surat.tujuan }}
    </p>

    <div class="flex items-center justify-between pt-1 border-t border-black/5 text-[11px]">
      <span class="inline-flex items-center gap-1.5 text-zinc-500 font-medium">
        <span class="w-1.5 h-1.5 rounded-full bg-[#0A84DC]"></span>
        {{ surat.pic_nama }}
      </span>

      <span v-if="surat.catatan" title="Terdapat catatan" class="text-zinc-400 text-xs">
        💬
      </span>
    </div>
  </article>
</template>
