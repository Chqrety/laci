<script setup lang="ts">
import { ref, watch } from 'vue'
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

const props = defineProps<{
  open: boolean
  user: User | null
  currentUserId: string | null
  divisiOptions: { value: string; label: string }[]
}>()

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
  (e: 'updated', user: User): void
}>()

const isSubmitting = ref(false)
const editForm = ref({
  id: '',
  nim: '',
  nama: '',
  email: '',
  divisi: 'humas',
})

watch(
  () => props.user,
  (newUser) => {
    if (newUser) {
      editForm.value = {
        id: newUser.id,
        nim: newUser.nim || '',
        nama: newUser.nama || '',
        email: newUser.email || '',
        divisi: newUser.divisi || 'humas',
      }
    }
  },
  { immediate: true },
)

const handleUpdateUser = async () => {
  if (!editForm.value.nama?.trim()) {
    toast({ title: 'Nama pengurus wajib diisi', variant: 'destructive' })
    return
  }

  isSubmitting.value = true
  try {
    const updated = await userService.update(editForm.value.id, {
      nim: editForm.value.nim,
      nama: editForm.value.nama,
      email: editForm.value.email,
      divisi: editForm.value.divisi,
    })

    toast({
      title: `Data pengurus ${updated.nama} berhasil diperbarui`,
      variant: 'success',
    })
    emit('updated', updated)
    emit('update:open', false)
  } catch (err: any) {
    toast({
      title: 'Gagal memperbarui pengurus',
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
        <DialogTitle>Edit Data Pengurus</DialogTitle>
        <DialogDescription>Perbarui data profil dan penempatan divisi.</DialogDescription>
      </DialogHeader>

      <form @submit.prevent="handleUpdateUser" novalidate class="flex flex-col gap-3">
        <div class="space-y-1">
          <Label for="edit-user-nama">Nama Lengkap *</Label>
          <Input
            id="edit-user-nama"
            v-model="editForm.nama"
            required
          />
        </div>

        <div class="space-y-1">
          <Label for="edit-user-nim">NIM (Nomor Induk Mahasiswa)</Label>
          <Input
            id="edit-user-nim"
            v-model="editForm.nim"
            placeholder="Contoh: A11.2023.12345"
          />
        </div>

        <div class="space-y-1">
          <Label for="edit-user-email">Email</Label>
          <Input
            id="edit-user-email"
            v-model="editForm.email"
            type="email"
          />
        </div>

        <div class="space-y-1">
          <Label for="edit-user-divisi">Divisi</Label>
          <Select
            v-model="editForm.divisi"
            :disabled="editForm.id === currentUserId && editForm.divisi === 'ketum'"
          >
            <SelectTrigger id="edit-user-divisi">
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
          <span
            v-if="editForm.id === currentUserId && editForm.divisi === 'ketum'"
            class="text-[10px] text-zinc-400 mt-1 block"
          >
            *Pergantian jabatan Ketua Umum wajib melalui menu Serah Terima Jabatan.
          </span>
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
            {{ isSubmitting ? 'Menyimpan...' : 'Simpan Perubahan' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>
