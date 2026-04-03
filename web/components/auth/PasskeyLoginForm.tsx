'use client'

import { useState } from 'react'
import { startAuthentication } from '@simplewebauthn/browser'
import { useAuth } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import type { AuthUser } from '@/lib/auth-store'
import { Fingerprint } from 'lucide-react'

interface Props {
  onSwitchToEmail: () => void
}

export default function PasskeyLoginForm({ onSwitchToEmail }: Props) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { setUser } = useAuth()
  const router = useRouter()

  const handlePasskey = async () => {
    setLoading(true)
    setError('')
    try {
      const begin = await api.post<{ session_id: string; options: { publicKey: Record<string, unknown> } }>('/api/auth/passkey/login/begin', {})
      const credential = await startAuthentication({ optionsJSON: begin.options.publicKey as never })
      const res = await api.post<{ token: string; user: AuthUser }>(
        `/api/auth/passkey/login/finish?session_id=${begin.session_id}`,
        credential,
      )
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
        <Fingerprint size={20} strokeWidth={1.75} aria-hidden="true" />
        {loading ? 'Verifying…' : 'Sign in with Passkey'}
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
