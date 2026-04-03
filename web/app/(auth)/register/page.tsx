// web/app/(auth)/register/page.tsx
'use client'

import { useState } from 'react'
import Link from 'next/link'
import PasskeyRegisterForm from '@/components/auth/PasskeyRegisterForm'
import EmailRegisterForm from '@/components/auth/EmailRegisterForm'
import AuthDivider from '@/components/auth/AuthDivider'

type View = 'passkey' | 'email'

export default function RegisterPage() {
  const [view, setView] = useState<View>('passkey')

  return (
    <div className="space-y-6">
      {/* Heading */}
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold text-white">Create your account</h1>
        <p className="text-sm text-gray-400">Start automating infrastructure today.</p>
      </div>

      {view === 'passkey' ? (
        <>
          <PasskeyRegisterForm onSwitchToEmail={() => setView('email')} />
          <AuthDivider />
        </>
      ) : (
        <EmailRegisterForm onSwitchToPasskey={() => setView('passkey')} />
      )}

      {/* Bottom link */}
      <p className="text-sm text-gray-400 text-center">
        Already have an account?{' '}
        <Link href="/login" className="text-cyan-400 hover:text-cyan-300 transition-colors">
          Sign in →
        </Link>
      </p>
    </div>
  )
}
