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

export default function PasskeyRegisterForm({ onSwitchToEmail }: Props) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { setUser } = useAuth()
  const router = useRouter()

  const handlePasskey = async () => {
    setLoading(true)
    setError('')
    try {
      const options = await api.post<Record<string, unknown>>('/api/auth/passkey/register/begin', {})
      const credential = await startRegistration({ optionsJSON: options as never })
      const res = await api.post<{ token: string; user: AuthUser }>('/api/auth/passkey/register/finish', credential)
      setUser(res.user, res.token)
      router.push('/dashboard')
    } catch {
      setError('Passkey failed — try email instead')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <button
        onClick={handlePasskey}
        disabled={loading}
        className="w-full h-12 rounded-xl font-semibold text-white flex items-center justify-center gap-2 transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
        style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
      >
        {/* Fingerprint icon */}
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M2 12C2 6.5 6.5 2 12 2a10 10 0 0 1 8 4"/>
          <path d="M5 19.5C5.5 18 6 15 6 12c0-1.6.6-3.1 1.8-4.2"/>
          <path d="M17.5 21C17 20 16 19 16 17c0-1.7-.5-3.3-1.5-4.5"/>
          <path d="M22 12a10 10 0 0 1-2 6"/>
          <path d="M9 12c0 1.7.5 3.3 1.5 4.5"/>
          <path d="M12 7a5 5 0 0 1 5 5c0 3.5-1.5 6.5-4 8.5"/>
          <path d="M12 7a5 5 0 0 0-5 5c0 1.5.2 3 .6 4.3"/>
        </svg>
        {loading ? 'Setting up…' : 'Create account with Passkey'}
      </button>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <button
        type="button"
        onClick={onSwitchToEmail}
        className="w-full text-sm text-gray-400 hover:text-white transition-colors duration-150 text-center"
      >
        Continue with email →
      </button>
    </div>
  )
}
