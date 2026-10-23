import axios from 'axios'
import { ref } from 'vue'

export function parseJwtPayload(token: string | null) {
  try { return token ? JSON.parse(atob(token.split('.')[1] || '')) : null } catch { return null }
}

export const currentToken = ref<string | null>(typeof window !== 'undefined' ? localStorage.getItem('token') : null)
export const currentRole = ref<string | null>(typeof window !== 'undefined' ? localStorage.getItem('role') : null)
const initialPayload = parseJwtPayload(currentToken.value)
export const currentUsername = ref<string | null>(
  (typeof window !== 'undefined' ? localStorage.getItem('username') : null) || initialPayload?.username || null,
)
export const currentUserId = ref<string | null>(initialPayload?.sub || null)

export const API_BASE_URL = (import.meta.env.VITE_API_URL || 'http://localhost:8080').trim().replace(/\/+$/, '')

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
      if (window.location.pathname !== '/login') {
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
  file_arsip?: string | null
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
  display_password?: string
  is_password_default: boolean
  has_seen_password_prompt: boolean
}

export interface LoginResponse {
  token: string
  divisi: string
  role?: string
  username: string
  nama?: string
  nim?: string
  email?: string
  is_password_default: boolean
  has_seen_password_prompt: boolean
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
      is_password_default: res.data.is_password_default,
      has_seen_password_prompt: res.data.has_seen_password_prompt,
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

  async dismissPasswordPrompt(): Promise<void> {
    await api.post('/profile/dismiss-prompt')
    if (currentUser.value) {
      currentUser.value.has_seen_password_prompt = true
    }
  },

  async changePassword(newPassword: string, oldPassword?: string): Promise<{ message: string }> {
    const res = await api.put('/profile/change-password', {
      new_password: newPassword,
      old_password: oldPassword,
    })
    if (currentUser.value) {
      currentUser.value.is_password_default = false
      currentUser.value.has_seen_password_prompt = true
    }
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

  async resetPrompt(id: string): Promise<{ message: string }> {
    const res = await api.put(`/users/${id}/reset-prompt`)
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
    const response = await api.get<Surat[]>('/surat')
    return response.data
  },

  async create(data: CreateSuratPayload): Promise<Surat> {
    const response = await api.post<Surat>('/surat', data)
    return response.data
  },

  async updateStatus(
    id: string,
    payload: UpdateStatusPayload,
  ): Promise<{ message: string; surat: Surat; log: LogTracking }> {
    const response = await api.put(`/surat/${id}`, payload)
    return response.data
  },

  async uploadArsip(
    id: string,
    file: Blob,
    filename = 'arsip.jpg',
  ): Promise<{ message: string; arsip_url: string; surat: Surat }> {
    const formData = new FormData()
    formData.append('file', file, filename)
    const response = await api.post(`/surat/${id}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  async uploadDokumen(
    id: string,
    file: File | Blob,
    filename?: string,
  ): Promise<{ message: string; file_arsip: string; arsip_url: string; surat: Surat }> {
    const formData = new FormData()
    if (filename) {
      formData.append('file', file, filename)
    } else {
      formData.append('file', file)
    }
    const response = await api.post(`/surat/${id}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
    return response.data
  },

  async delete(id: string): Promise<{ message: string }> {
    const response = await api.delete(`/surat/${id}`)
    return response.data
  },
}

export const prokerService = {
  async getAll(force = false): Promise<Proker[]> {
    if (!force && prokerCache) return prokerCache
    const response = await api.get<Proker[]>('/proker')
    prokerCache = response.data
    return response.data
  },

  async create(nama_proker: string, deskripsi = ''): Promise<Proker> {
    prokerCache = null
    const response = await api.post<Proker>('/proker', { nama_proker, deskripsi })
    return response.data
  },

  async update(id: string, nama_proker: string, deskripsi = ''): Promise<Proker> {
    prokerCache = null
    const response = await api.put<Proker>(`/proker/${id}`, { nama_proker, deskripsi })
    return response.data
  },

  async delete(id: string): Promise<{ message: string }> {
    prokerCache = null
    const response = await api.delete(`/proker/${id}`)
    return response.data
  },
}

// Modular API Functions
export const fetchSurat = () => suratService.getAll()
export const updateSuratStatus = (id: string, payload: UpdateStatusPayload) => suratService.updateStatus(id, payload)
export const uploadSuratArsip = (id: string, file: Blob, filename?: string) =>
  suratService.uploadArsip(id, file, filename)
export const uploadSuratDokumen = (id: string, file: File | Blob, filename?: string) =>
  suratService.uploadDokumen(id, file, filename)

export const fetchProker = () => prokerService.getAll()

export const STATUS_COLUMNS = [
  'Standby',
  'Cetak',
  'TTD Lapis 1',
  'TTD Ketum',
  'TTD Pembina',
  'Paraf Koormawa',
  'TTD Tertinggi',
  'Selesai',
] as const

export type SuratStatus = (typeof STATUS_COLUMNS)[number]
