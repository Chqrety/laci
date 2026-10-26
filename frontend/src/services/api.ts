import axios from 'axios'
import { ref } from 'vue'

export function parseJwtPayload(token: string | null) {
  try {
    return token ? JSON.parse(atob(token.split('.')[1] || '')) : null
  } catch {
    return null
  }
}

export const currentToken = ref<string | null>(typeof window !== 'undefined' ? localStorage.getItem('token') : null)
export const currentRole = ref<string | null>(typeof window !== 'undefined' ? localStorage.getItem('role') : null)
const initialPayload = parseJwtPayload(currentToken.value)
export const currentUsername = ref<string | null>(
  (typeof window !== 'undefined' ? localStorage.getItem('username') : null) || initialPayload?.username || null,
)
export const currentUserId = ref<string | null>(initialPayload?.sub || null)

export const API_BASE_URL = (import.meta.env.VITE_API_URL || '/api').trim().replace(/\/+$/, '')

export function getBackendAssetUrl(path?: string | null): string {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  return `${API_BASE_URL}/${path.replace(/^\/+/, '')}`
}

export const api = axios.create({
  baseURL: API_BASE_URL,
})

// Request interceptor: attach JWT bearer token if available
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor: handle 401 unauthorized
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('role')
      currentToken.value = null
      currentRole.value = null
      if (window.location.pathname !== '/login' && window.location.pathname !== '/activate' && window.location.pathname !== '/forgot-password' && window.location.pathname !== '/reset-password') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

export interface LogTracking {
  id: string
  surat_id: string
  status_baru: string
  catatan_log?: string
  waktu_update: string
}

export interface Surat {
  id: string
  nomor_surat: string
  perihal: string
  tujuan?: string
  status_saat_ini: string
  pic_nama: string
  arsip_url?: string | null
  gdocs_url?: string | null

  catatan?: string
  logs?: LogTracking[]
}

export interface Proker {
  id: string
  nama_proker: string
  deskripsi?: string
}

export interface CreateSuratPayload {
  nomor_surat: string
  perihal: string
  tujuan?: string
  status_saat_ini: string
  pic_nama: string
  catatan?: string
}

export interface UpdateStatusPayload {
  status_saat_ini: string
  tujuan?: string
  catatan?: string
  catatan_log?: string
}

export interface User {
  id: string
  nim?: string
  nama: string
  email?: string
  username: string
  divisi: string
  role?: string
  is_active?: boolean
}

export interface LoginResponse {
  token: string
  divisi: string
  role?: string
  username: string
  nama?: string
  nim?: string
  email?: string
  is_active?: boolean
}

let prokerCache: Proker[] | null = null
let usersCache: User[] | null = null

export const clearApiCache = () => {
  prokerCache = null
  usersCache = null
}

export const currentUser = ref<User | null>(null)

export const authService = {
  async login(username: string, password: string): Promise<LoginResponse> {
    const res = await api.post<LoginResponse>('/login', { username, password })
    const payload = parseJwtPayload(res.data.token)
    const divisiVal = res.data.divisi || res.data.role || 'humas'

    localStorage.setItem('token', res.data.token)
    localStorage.setItem('role', divisiVal)
    localStorage.setItem('divisi', divisiVal)
    localStorage.setItem('username', res.data.username || username || payload?.username || '')

    currentToken.value = res.data.token
    currentRole.value = divisiVal
    currentUsername.value = res.data.username || username || payload?.username || null
    currentUserId.value = payload?.sub || null
    currentUser.value = {
      id: payload?.sub || '',
      nim: res.data.nim || '',
      nama: res.data.nama || '',
      email: res.data.email || '',
      username: res.data.username || username,
      divisi: divisiVal,
      role: divisiVal,
    }

    return res.data
  },

  logout() {
    clearApiCache()
    localStorage.removeItem('token')
    localStorage.removeItem('role')
    localStorage.removeItem('divisi')
    localStorage.removeItem('username')
    currentToken.value = null
    currentRole.value = null
    currentUsername.value = null
    currentUserId.value = null
    currentUser.value = null
    window.location.href = '/login'
  },

  isAuthenticated(): boolean {
    return !!currentToken.value
  },

  async getProfile(force = false): Promise<User> {
    if (!force && currentUser.value?.nama) {
      return currentUser.value
    }
    const res = await api.get<User>('/profile')
    currentUser.value = res.data
    return res.data
  },

  async changePassword(newPassword: string, oldPassword?: string): Promise<{ message: string }> {
    const res = await api.put('/profile/change-password', {
      new_password: newPassword,
      old_password: oldPassword,
    })
    return res.data
  },

  async changeUsername(newUsername: string): Promise<{ message: string; username: string }> {
    const res = await api.put<{ message: string; username: string }>('/profile/change-username', {
      username: newUsername,
    })
    if (currentUser.value) {
      currentUser.value.username = res.data.username
    }
    localStorage.setItem('username', res.data.username)
    currentUsername.value = res.data.username
    return res.data
  },

  async verifyToken(token: string, type: 'activation' | 'reset' = 'activation'): Promise<{ valid: boolean; email: string; nama: string; username: string }> {
    const res = await api.get('/auth/verify-token', { params: { token, type } })
    return res.data
  },

  async activateAccount(token: string, password: string): Promise<{ message: string }> {
    const res = await api.post('/auth/activate', { token, password })
    return res.data
  },

  async forgotPassword(email: string): Promise<{ message: string }> {
    const res = await api.post('/auth/forgot-password', { email })
    return res.data
  },

  async resetPassword(token: string, password: string): Promise<{ message: string }> {
    const res = await api.post('/auth/reset-password', { token, password })
    return res.data
  },
}

export const userService = {
  async getAll(force = false): Promise<User[]> {
    if (!force && usersCache) return usersCache
    const res = await api.get<User[]>('/users')
    usersCache = res.data
    return res.data
  },

  async create(data: Partial<User> & { password?: string }): Promise<User> {
    usersCache = null
    const res = await api.post<User>('/users', data)
    return res.data
  },

  async update(id: string, data: Partial<User> & { password?: string }): Promise<User> {
    usersCache = null
    const res = await api.put<User>(`/users/${id}`, data)
    return res.data
  },

  async delete(id: string): Promise<{ message: string }> {
    usersCache = null
    const res = await api.delete(`/users/${id}`)
    return res.data
  },

  async deleteBatch(ids: string[]): Promise<{ message: string; deleted: number }> {
    usersCache = null
    const res = await api.delete('/users/batch', { data: { ids } })
    return res.data
  },

  async transferKetum(targetUserId: string, currentPassword: string): Promise<{ message: string; ketum_baru: string }> {
    usersCache = null
    const res = await api.post('/users/transfer-ketum', {
      target_user_id: targetUserId,
      current_password: currentPassword,
    })
    return res.data
  },

  async importCSV(file: File): Promise<{ message: string; imported: User[]; skipped: string[] }> {
    usersCache = null
    const formData = new FormData()
    formData.append('file', file)
    const res = await api.post('/users/import-csv', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return res.data
  },

  async importSpreadsheet(sheetURL: string): Promise<{ message: string; imported: User[]; skipped: string[] }> {
    usersCache = null
    const res = await api.post('/users/import-spreadsheet', { sheet_url: sheetURL })
    return res.data
  },
}

export const suratService = {
  async getAll(): Promise<Surat[]> {
    const response = await api.get<Surat[]>('/letters')
    return response.data
  },

  async create(data: CreateSuratPayload): Promise<Surat> {
    const response = await api.post<Surat>('/letters', data)
    return response.data
  },

  async updateStatus(
    id: string,
    payload: UpdateStatusPayload,
  ): Promise<{ message: string; surat: Surat; log: LogTracking }> {
    const response = await api.put(`/letters/${id}`, payload)
    return response.data
  },

  async delete(id: string): Promise<{ message: string }> {
    const response = await api.delete(`/letters/${id}`)
    return response.data
  },
}

export const prokerService = {
  async getAll(force = false): Promise<Proker[]> {
    if (!force && prokerCache) return prokerCache
    const response = await api.get<Proker[]>('/programs')
    prokerCache = response.data
    return response.data
  },

  async create(nama_proker: string, deskripsi = ''): Promise<Proker> {
    prokerCache = null
    const response = await api.post<Proker>('/programs', { nama_proker, deskripsi })
    return response.data
  },

  async update(id: string, nama_proker: string, deskripsi = ''): Promise<Proker> {
    prokerCache = null
    const response = await api.put<Proker>(`/programs/${id}`, { nama_proker, deskripsi })
    return response.data
  },

  async delete(id: string): Promise<{ message: string }> {
    prokerCache = null
    const response = await api.delete(`/programs/${id}`)
    return response.data
  },
}

export const STATUS_COLUMNS = [
  'Standby',
  'Cetak',
  'TTD Dasar',
  'TTD Ketum',
  'TTD Pembina',
  'Paraf Koormawa',
  'TTD Tertinggi',
  'Selesai',
] as const

export type SuratStatus = (typeof STATUS_COLUMNS)[number]

export interface SuratTemplate {
  id: string
  nama_template: string
  gdoc_id: string
  placeholders: string
  is_active: boolean
}

export interface GenerateSuratPayload {
  template_id: string
  folder_url?: string
  file_name?: string
  nomor_surat: string
  perihal: string
  tujuan?: string
  pic_nama?: string
  proker_id?: string
  replacements: Record<string, string>
}

export const templateService = {
  async getAll(): Promise<SuratTemplate[]> {
    const response = await api.get<SuratTemplate[]>('/letters/templates')
    return response.data
  },

  async sync(
    folderUrl: string,
    gasUrl?: string,
  ): Promise<{ templates: SuratTemplate[]; count: number; message: string }> {
    const response = await api.post('/letters/templates/sync', { folder_url: folderUrl, gas_url: gasUrl })
    return response.data
  },

  async generate(
    payload: GenerateSuratPayload,
  ): Promise<{ surat: Surat; gdocs_url: string; pdf_url: string; doc_id: string; message: string }> {
    const response = await api.post('/letters/generate', payload)
    return response.data
  },

  async getSettings(): Promise<Record<string, string>> {
    const response = await api.get<Record<string, string>>('/settings')
    return response.data
  },

  async updateSettings(settings: Record<string, string>): Promise<void> {
    await api.put('/settings', settings)
  },
}
