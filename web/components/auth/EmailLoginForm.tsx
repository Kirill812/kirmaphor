// web/components/auth/EmailLoginForm.tsx
'use client'

import { useState } from 'react'
import { useAuth } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'

interface Props {
  onSwitchToPasskey: () => void
}

export default function EmailLoginForm({ onSwitchToPasskey }: Props) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await login(email, password)
      router.push('/dashboard')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Sign in failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-4 animate-slide-in"
    >
      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="login-email">Email</label>
        <input
          id="login-email"
          type="email"
          required
          value={email}
          onChange={e => setEmail(e.target.value)}
          placeholder="you@example.com"
          className="w-full h-11 rounded-lg px-3 text-white placeholder-gray-500 outline-none transition-[border-color,box-shadow] duration-150 border focus:border-cyan-500/50 focus:shadow-[0_0_0_3px_rgba(6,182,212,0.15)]"
          style={{
            backgroundColor: 'rgba(255,255,255,0.05)',
            borderColor: 'rgba(255,255,255,0.08)',
          }}
        />
      </div>

      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="login-password">Password</label>
        <input
          id="login-password"
          type="password"
          required
          value={password}
          onChange={e => setPassword(e.target.value)}
          placeholder="••••••••"
          className="w-full h-11 rounded-lg px-3 text-white placeholder-gray-500 outline-none transition-[border-color,box-shadow] duration-150 border focus:border-cyan-500/50 focus:shadow-[0_0_0_3px_rgba(6,182,212,0.15)]"
          style={{
            backgroundColor: 'rgba(255,255,255,0.05)',
            borderColor: 'rgba(255,255,255,0.08)',
          }}
        />
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading}
        className="w-full h-12 rounded-xl font-semibold text-white transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
        style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
      >
        {loading ? 'Signing in…' : 'Sign in'}
      </button>

      <button
        type="button"
        onClick={onSwitchToPasskey}
        className="w-full text-sm text-gray-400 hover:text-white transition-colors duration-150 text-center"
      >
        ← Back to passkey
      </button>
    </form>
  )
}
