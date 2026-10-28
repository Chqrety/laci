<script setup lang="ts">
import { ref } from 'vue'
import { userService, type User } from '@/services/api'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { toast } from '@/components/ui/toast'

defineProps<{
  open: boolean
  divisiOptions: { value: string; label: string }[]
}>()

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
  (e: 'created', user: User): void
}>()

const isSubmitting = ref(false)
const formUser = ref({
  nim: '',
  nama: '',
  email: '',
  username: '',
  divisi: 'humas',
  is_koor: false,
})

const handleCreateUser = async () => {
  if (!formUser.value.nama?.trim()) {
    toast({ title: 'Nama pengurus wajib diisi', variant: 'destructive' })
    return
  }

  isSubmitting.value = true
  try {
    const created = await userService.create(formUser.value)
    formUser.value = {
      nim: '',
      nama: '',
      email: '',
      username: '',
      divisi: 'humas',
      is_koor: false,
    }
    toast({
      title: `Pengurus ${created.nama} berhasil ditambahkan`,
      description: 'Undangan aktivasi telah dikirim ke email.',
      variant: 'success',
    })
    emit('created', created)
    emit('update:open', false)
  } catch (err: any) {
    toast({
      title: 'Gagal menambah pengurus',
      description: err.response?.data?.error || err.message,
      variant: 'destructive',
    })
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="(val: boolean) => emit('update:open', val)">
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Tambah Pengurus Baru</DialogTitle>
        <DialogDescription>Pengurus baru akan menerima email aktivasi untuk membuat kata sandi mandiri.</DialogDescription>
      </DialogHeader>

      <form @submit.prevent="handleCreateUser" novalidate class="flex flex-col gap-3">
        <div class="space-y-1">
          <Label for="add-user-nama">Nama Lengkap *</Label>
          <Input
            id="add-user-nama"
            v-model="formUser.nama"
            required
            placeholder="Contoh: Budi Santoso"
          />
        </div>

        <div class="space-y-1">
          <Label for="add-user-nim">NIM (Nomor Induk Mahasiswa)</Label>
          <Input
            id="add-user-nim"
            v-model="formUser.nim"
            placeholder="Contoh: A11.2023.12345"
          />
        </div>

        <div class="space-y-1">
          <Label for="add-user-email">Email *</Label>
          <Input
            id="add-user-email"
            v-model="formUser.email"
            type="email"
            required
            placeholder="budi@email.com"
          />
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div class="space-y-1">
            <Label for="add-user-username">Username (Opsional)</Label>
            <Input
              id="add-user-username"
              v-model="formUser.username"
              placeholder="Auto dari NIM/Nama"
            />
          </div>
          <div class="space-y-1">
            <Label for="add-user-divisi">Divisi</Label>
            <Select v-model="formUser.divisi">
              <SelectTrigger id="add-user-divisi">
                <SelectValue placeholder="Pilih Divisi" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem
                  v-for="d in divisiOptions"
                  :key="d.value"
                  :value="d.value"
                >
                  {{ d.label }}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <!-- Toggle Mandat Koordinator -->
        <div class="flex items-center gap-2.5 p-2.5 rounded-xl bg-zinc-50 border border-zinc-200 mt-1">
          <input
            id="add-user-is-koor"
            v-model="formUser.is_koor"
            type="checkbox"
            class="w-4 h-4 rounded text-[#0A84DC] focus:ring-[#0A84DC]/30 border-zinc-300 cursor-pointer accent-[#0A84DC]"
          />
          <Label for="add-user-is-koor" class="text-xs font-medium text-zinc-800 cursor-pointer select-none">
            Mandat Koordinator (Bisa menugaskan &amp; mengganti PIC surat)
          </Label>
        </div>

        <DialogFooter class="pt-2">
          <Button
            type="button"
            variant="outline"
            size="sm"
            @click="emit('update:open', false)"
          >
            Batal
          </Button>
          <Button
            type="submit"
            size="sm"
            :disabled="isSubmitting"
          >
            {{ isSubmitting ? 'Menyimpan...' : 'Kirim Undangan Aktivasi' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
