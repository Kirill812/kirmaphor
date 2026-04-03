'use client'

import { useState } from 'react'

const COMMUNITY_FEATURES = [
  'All core automation features',
  'Unlimited projects',
  'Docker Compose deploy',
  'Community support',
  'MIT license — own your data',
]

const PRO_FEATURES = [
  'Everything in Community',
  'Managed cloud hosting',
  'Forge AI (unlimited)',
  'Priority support',
  'SSO / SAML',
]

export default function PricingSection() {
  const [email, setEmail] = useState('')
  const [submitted, setSubmitted] = useState(false)

  return (
    <section
      className="px-6 py-20 lg:px-12"
      style={{ backgroundColor: 'var(--land-bg)', borderTop: '1px solid rgba(255,255,255,0.06)' }}
    >
      <p className="mb-3 text-[11px] uppercase tracking-[0.12em]" style={{ color: 'rgba(6,182,212,0.7)' }}>
        Open source
      </p>
      <h2
        className="mb-3 font-bold tracking-[-0.03em]"
        style={{ fontSize: 'clamp(28px, 3.5vw, 44px)', color: 'rgba(255,255,255,0.9)' }}
      >
        Self-host for free.<br />Scale with cloud.
      </h2>
      <p className="mb-12 text-[15px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)', maxWidth: '480px' }}>
        Start on your own server today. Cloud hosting coming soon.
      </p>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 lg:max-w-3xl">
        {/* Community */}
        <div
          className="rounded-2xl p-8"
          style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.10)' }}
        >
          <p className="mb-1 text-[13px] font-semibold" style={{ color: 'rgba(255,255,255,0.88)' }}>Community</p>
          <p className="mb-6 text-[12px]" style={{ color: 'rgba(255,255,255,0.35)' }}>Free forever · MIT license</p>
          <ul className="mb-8 space-y-2.5">
            {COMMUNITY_FEATURES.map(f => (
              <li key={f} className="flex items-start gap-2 text-[13px]" style={{ color: 'rgba(255,255,255,0.55)' }}>
                <span style={{ color: '#4ade80', marginTop: '1px' }}>✓</span>
                {f}
              </li>
            ))}
          </ul>
          <a
            href="https://github.com/Kirill812/kirmaphor"
            target="_blank"
            rel="noopener noreferrer"
            className="flex h-11 w-full items-center justify-center rounded-xl text-[14px] font-medium transition-[background] duration-150 hover:bg-white/[0.09]"
            style={{
              background: 'rgba(255,255,255,0.06)',
              border: '1px solid rgba(255,255,255,0.10)',
              color: 'rgba(255,255,255,0.75)',
            }}
          >
            Deploy yourself →
          </a>
        </div>

        {/* Pro */}
        <div
          className="relative rounded-2xl p-8 overflow-hidden"
          style={{
            background: 'rgba(6,182,212,0.04)',
            border: '1px solid rgba(6,182,212,0.3)',
            boxShadow: '0 0 60px rgba(6,182,212,0.06)',
          }}
        >
          <p className="mb-1 text-[13px] font-semibold" style={{ color: 'rgba(255,255,255,0.88)' }}>Pro</p>
          <p className="mb-6 text-[12px]" style={{ color: 'rgba(6,182,212,0.7)' }}>Cloud-hosted · coming soon</p>
          <ul className="mb-8 space-y-2.5">
            {PRO_FEATURES.map(f => (
              <li key={f} className="flex items-start gap-2 text-[13px]" style={{ color: 'rgba(255,255,255,0.55)' }}>
                <span style={{ color: '#06b6d4', marginTop: '1px' }}>✓</span>
                {f}
              </li>
            ))}
          </ul>
          {submitted ? (
            <p className="text-center text-[13px]" style={{ color: '#4ade80' }}>
              ✓ You&apos;re on the list!
            </p>
          ) : (
            <div className="flex gap-2">
              <input
                type="email"
                placeholder="you@company.com"
                value={email}
                onChange={e => setEmail(e.target.value)}
                className="flex-1 rounded-xl px-3 text-[13px] text-white placeholder-gray-500 outline-none h-11"
                style={{
                  background: 'rgba(255,255,255,0.06)',
                  border: '1px solid rgba(255,255,255,0.10)',
                }}
                onKeyDown={e => { if (e.key === 'Enter' && email) setSubmitted(true) }}
              />
              <button
                onClick={() => { if (email) setSubmitted(true) }}
                className="h-11 rounded-xl px-4 text-[13px] font-semibold text-white transition-[filter] duration-150 hover:brightness-110"
                style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
              >
                Join waitlist →
              </button>
            </div>
          )}
        </div>
      </div>

      {/* GitHub stars badge */}
      <div className="mt-10 flex items-center gap-3">
        <a
          href="https://github.com/Kirill812/kirmaphor"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-2 rounded-full px-4 py-2 text-[12px] transition-[background] duration-150 hover:bg-white/[0.06]"
          style={{
            background: 'rgba(255,255,255,0.04)',
            border: '1px solid rgba(255,255,255,0.08)',
            color: 'rgba(255,255,255,0.5)',
          }}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
          </svg>
          Star on GitHub
        </a>
      </div>
    </section>
  )
}
