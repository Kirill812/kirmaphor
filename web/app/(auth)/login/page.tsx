'use client'

import { useState } from 'react'
import Link from 'next/link'
import PasskeyLoginForm from '@/components/auth/PasskeyLoginForm'
import EmailLoginForm from '@/components/auth/EmailLoginForm'
import AuthDivider from '@/components/auth/AuthDivider'

type View = 'passkey' | 'email'

export default function LoginPage() {
  const [view, setView] = useState<View>('passkey')

  return (
    <div className="space-y-6">
      {/* Heading */}
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold text-white">Welcome back</h1>
        <p className="text-sm text-gray-400">Sign in to your account</p>
      </div>

      {view === 'passkey' ? (
        <>
          <PasskeyLoginForm onSwitchToEmail={() => setView('email')} />
          <AuthDivider />
        </>
      ) : (
        <EmailLoginForm onSwitchToPasskey={() => setView('passkey')} />
      )}

      {/* Bottom link */}
      <p className="text-sm text-gray-400 text-center">
        Don&apos;t have an account?{' '}
        <Link href="/register" className="text-cyan-400 hover:text-cyan-300 transition-colors">
          Create one →
        </Link>
      </p>
    </div>
  )
}
