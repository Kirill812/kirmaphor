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

export const useAuth = create<AuthState>((set) => ({
  user: null,
  token: typeof window !== 'undefined' ? localStorage.getItem('kirmaphore_token') : null,

  login: async (email, password) => {
    const res = await api.post<{ token: string; user: AuthUser }>('/api/auth/login', { email, password })
    localStorage.setItem('kirmaphore_token', res.token)
    set({ user: res.user, token: res.token })
  },

  logout: () => {
    api.post('/api/auth/logout', {}).catch(() => {})
    localStorage.removeItem('kirmaphore_token')
    set({ user: null, token: null })
  },

  setUser: (user, token) => {
    localStorage.setItem('kirmaphore_token', token)
    set({ user, token })
  },
}))
