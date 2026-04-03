// web/lib/auth-store.ts
import { create } from 'zustand'
import { api } from './api'

export interface AuthUser {
  id: string
  email: string
  display_name: string
  avatar_url?: string
}

interface AuthState {
  user: AuthUser | null
  token: string | null
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  setUser: (user: AuthUser, token: string) => void
}

function saveToken(token: string) {
  localStorage.setItem('kirmaphore_token', token)
  document.cookie = `kirmaphore_token=${token}; path=/; SameSite=Lax`
}

function clearToken() {
  localStorage.removeItem('kirmaphore_token')
  document.cookie = 'kirmaphore_token=; path=/; max-age=0'
}

export const useAuth = create<AuthState>((set) => ({
  user: null,
  token: typeof window !== 'undefined' ? localStorage.getItem('kirmaphore_token') : null,

  login: async (email, password) => {
    const res = await api.post<{ token: string; user: AuthUser }>('/api/auth/login', { email, password })
    saveToken(res.token)
    set({ user: res.user, token: res.token })
  },

  logout: () => {
    api.post('/api/auth/logout', {}).catch(() => {})
    clearToken()
    set({ user: null, token: null })
  },

  setUser: (user, token) => {
    saveToken(token)
    set({ user, token })
  },
}))
