'use client'

import { useState } from 'react'
import { useAuth } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import type { AuthUser } from '@/lib/auth-store'

interface Props {
  onSwitchToPasskey?: () => void
}

const inputClass = `w-full h-11 rounded-lg px-3 text-white placeholder-gray-500 outline-none transition-[border-color,box-shadow] duration-150 border focus:border-cyan-500/50 focus:shadow-[0_0_0_3px_rgba(6,182,212,0.15)]`

const inputStyle = {
  backgroundColor: 'rgba(255,255,255,0.05)',
  borderColor: 'rgba(255,255,255,0.08)',
}

export default function EmailRegisterForm({ onSwitchToPasskey }: Props) {
  const [displayName, setDisplayName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { setUser } = useAuth()
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await api.post<{ token: string; user: AuthUser }>('/api/auth/register', {
        display_name: displayName,
        email,
        password,
      })
      setUser(res.user, res.token)
      router.push('/dashboard')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 animate-slide-in">
      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="reg-name">Display name</label>
        <input
          id="reg-name"
          type="text"
          required
          value={displayName}
          onChange={e => setDisplayName(e.target.value)}
          placeholder="Ada Lovelace"
          className={inputClass}
          style={inputStyle}
        />
      </div>

      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="reg-email">Email</label>
        <input
          id="reg-email"
          type="email"
          required
          value={email}
          onChange={e => setEmail(e.target.value)}
          placeholder="you@example.com"
          className={inputClass}
          style={inputStyle}
        />
      </div>

      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="reg-password">Password</label>
        <input
          id="reg-password"
          type="password"
          required
          value={password}
          onChange={e => setPassword(e.target.value)}
          placeholder="••••••••"
          className={inputClass}
          style={inputStyle}
        />
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading}
        className="w-full h-12 rounded-xl font-semibold text-white transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
        style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
      >
        {loading ? 'Creating account…' : 'Create account'}
      </button>

      {onSwitchToPasskey && (
        <button
          type="button"
          onClick={onSwitchToPasskey}
          className="w-full text-sm text-gray-400 hover:text-white transition-colors duration-150 text-center"
        >
          ← Back to passkey
        </button>
      )}
    </form>
  )
}
