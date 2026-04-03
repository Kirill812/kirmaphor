'use client'

import { useState } from 'react'
import { startRegistration } from '@simplewebauthn/browser'
import { useAuth } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import type { AuthUser } from '@/lib/auth-store'

interface Props {
  onSwitchToEmail: () => void
}

const inputClass = `w-full h-11 rounded-lg px-3 text-white placeholder-gray-500 outline-none transition-[border-color,box-shadow] duration-150 border focus:border-cyan-500/50 focus:shadow-[0_0_0_3px_rgba(6,182,212,0.15)]`
const inputStyle = { backgroundColor: 'rgba(255,255,255,0.05)', borderColor: 'rgba(255,255,255,0.08)' }

export default function PasskeyRegisterForm({ onSwitchToEmail }: Props) {
  const [displayName, setDisplayName] = useState('')
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { setUser } = useAuth()
  const router = useRouter()

  const handlePasskey = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email || !displayName) return
    setLoading(true)
    setError('')
    try {
      const res = await api.post<{ pending_id: string; options: { publicKey: Record<string, unknown> } }>(
        '/api/auth/passkey/register/begin',
        { email, display_name: displayName },
      )
      const credential = await startRegistration({ optionsJSON: res.options.publicKey as never })
      const authRes = await api.post<{ token: string; user: AuthUser }>(
        `/api/auth/passkey/register/finish?pending_id=${encodeURIComponent(res.pending_id)}`,
        credential,
      )
      setUser(authRes.user, authRes.token)
      router.push('/dashboard')
    } catch {
      setError('Passkey setup failed — try email instead')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handlePasskey} className="space-y-4 animate-slide-in">
      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="pk-name">Display name</label>
        <input
          id="pk-name"
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
        <label className="text-sm text-gray-300" htmlFor="pk-email">Email</label>
        <input
          id="pk-email"
          type="email"
          required
          value={email}
          onChange={e => setEmail(e.target.value)}
          placeholder="you@example.com"
          className={inputClass}
          style={inputStyle}
        />
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading}
        className="w-full h-12 rounded-xl font-semibold text-white flex items-center justify-center gap-2 transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
        style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
      >
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M2 12C2 6.5 6.5 2 12 2a10 10 0 0 1 8 4"/>
          <path d="M5 19.5C5.5 18 6 15 6 12c0-1.6.6-3.1 1.8-4.2"/>
          <path d="M17.5 21C17 20 16 19 16 17c0-1.7-.5-3.3-1.5-4.5"/>
          <path d="M22 12a10 10 0 0 1-2 6"/>
          <path d="M9 12c0 1.7.5 3.3 1.5 4.5"/>
          <path d="M12 7a5 5 0 0 1 5 5c0 3.5-1.5 6.5-4 8.5"/>
          <path d="M12 7a5 5 0 0 0-5 5c0 1.5.2 3 .6 4.3"/>
        </svg>
        {loading ? 'Setting up…' : 'Continue with Passkey'}
      </button>

      <button
        type="button"
        onClick={onSwitchToEmail}
        className="w-full text-sm text-gray-400 hover:text-white transition-colors duration-150 text-center"
      >
        Use password instead →
      </button>
    </form>
  )
}
