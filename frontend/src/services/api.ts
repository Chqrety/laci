import axios from 'axios'
import { ref } from 'vue'

export const currentToken = ref<string | null>(typeof window !== 'undefined' ? localStorage.getItem('token') : null)
export const currentRole = ref<string | null>(typeof window !== 'undefined' ? localStorage.getItem('role') : null)

export const API_BASE_URL = import.meta.env.VITE_API_URL.trim().replace(/\/+$/, '')

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
  status_saat_ini: string
  pic_nama: string
  catatan?: string
}

export interface UpdateStatusPayload {
  status_saat_ini: string
  catatan?: string
  catatan_log?: string
}

export interface LoginResponse {
  token: string
  role: 'sekre' | 'humas'
}

export const authService = {
  async login(username: string, password: string): Promise<LoginResponse> {
    const res = await api.post<LoginResponse>('/login', { username, password })
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('role', res.data.role)
    currentToken.value = res.data.token
    currentRole.value = res.data.role
    return res.data
  },

  logout() {
    localStorage.removeItem('token')
    localStorage.removeItem('role')
    currentToken.value = null
    currentRole.value = null
    window.location.href = '/login'
  },

  getToken(): string | null {
    return currentToken.value
  },

  getRole(): string | null {
    return currentRole.value
  },

  isAuthenticated(): boolean {
    return !!currentToken.value
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
    const response = await api.put(`/surat/${id}/status`, payload)
    return response.data
  },

  async uploadArsip(
    id: string,
    file: Blob,
    filename = 'arsip.jpg',
  ): Promise<{ message: string; arsip_url: string; surat: Surat }> {
    const formData = new FormData()
    formData.append('file', file, filename)
    const response = await api.post(`/surat/${id}/arsip`, formData, {
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
  async getAll(): Promise<Proker[]> {
    const response = await api.get<Proker[]>('/proker')
    return response.data
  },

  async create(nama_proker: string, deskripsi = ''): Promise<Proker> {
    const response = await api.post<Proker>('/proker', { nama_proker, deskripsi })
    return response.data
  },
}
