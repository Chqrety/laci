<script setup lang="ts">
import type { User } from '@/services/api'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

defineProps<{
  open: boolean
  user: User | null
}>()

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
  (e: 'edit', user: User): void
}>()
</script>

<template>
  <Dialog :open="open && !!user" @update:open="(val: boolean) => emit('update:open', val)">
    <DialogContent v-if="user">
      <DialogHeader>
        <div class="flex items-center gap-3">
          <div class="w-11 h-11 rounded-2xl bg-[#0A84DC]/10 text-[#0A84DC] flex items-center justify-center font-bold text-sm">
            {{ (user.nama || user.username).charAt(0).toUpperCase() }}
          </div>
          <div>
            <DialogTitle>Detail Pengurus</DialogTitle>
            <DialogDescription>@{{ user.username }}</DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <div class="grid grid-cols-1 gap-2.5 bg-zinc-50 p-4 rounded-2xl border border-zinc-200 text-xs">
        <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
          <span class="text-zinc-500">Nama Lengkap</span>
          <span class="font-semibold text-zinc-900 text-right">{{ user.nama || '-' }}</span>
        </div>
        <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
          <span class="text-zinc-500">NIM</span>
          <span class="font-mono text-zinc-800 text-right">{{ user.nim || '-' }}</span>
        </div>
        <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
          <span class="text-zinc-500">Divisi</span>
          <Badge variant="mono" size="sm">
            {{ user.divisi }}
          </Badge>
        </div>
        <div class="flex items-center justify-between pb-2 border-b border-zinc-200/60">
          <span class="text-zinc-500">Email</span>
          <span class="text-zinc-800 text-right select-all">{{ user.email || '-' }}</span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-zinc-500">Status Akun</span>
          <Badge
            :variant="user.is_active ? 'emerald' : 'amber'"
            size="sm"
          >
            {{ user.is_active ? 'Aktif' : 'Menunggu Aktivasi' }}
          </Badge>
        </div>
      </div>

      <DialogFooter class="pt-2">
        <Button
          type="button"
          variant="outline"
          size="sm"
          @click="emit('update:open', false)"
        >
          Tutup
        </Button>
        <Button
          type="button"
          size="sm"
          @click="emit('edit', user); emit('update:open', false)"
        >
          <svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
          </svg>
          Edit Pengurus
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
