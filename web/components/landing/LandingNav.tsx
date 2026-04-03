'use client'
import Link from 'next/link'

export default function LandingNav() {
  return (
    <nav
      className="fixed top-0 left-0 right-0 z-50 flex items-center justify-between px-6 lg:px-12"
      style={{
        height: '56px',
        background: 'rgba(6,10,20,0.8)',
        backdropFilter: 'blur(12px)',
        borderBottom: '1px solid rgba(255,255,255,0.06)',
      }}
    >
      {/* Logo */}
      <span className="text-white font-bold text-[15px] tracking-tight">Kirmaphore</span>

      {/* Center links — hidden on mobile */}
      <div className="hidden lg:flex items-center gap-8">
        {(['Features', 'Pricing', 'Docs'] as const).map(label => (
          <a
            key={label}
            href={`#${label.toLowerCase()}`}
            className="text-[13px] transition-colors duration-150"
            style={{ color: 'rgba(255,255,255,0.45)' }}
            onMouseOver={e => (e.currentTarget.style.color = 'white')}
            onMouseOut={e => (e.currentTarget.style.color = 'rgba(255,255,255,0.45)')}
          >
            {label}
          </a>
        ))}
        <a
          href="https://github.com/Kirill812/kirmaphor"
          target="_blank"
          rel="noopener noreferrer"
          className="text-[13px] transition-colors duration-150"
          style={{ color: 'rgba(255,255,255,0.45)' }}
          onMouseOver={e => (e.currentTarget.style.color = 'white')}
          onMouseOut={e => (e.currentTarget.style.color = 'rgba(255,255,255,0.45)')}
        >
          GitHub
        </a>
      </div>

      {/* CTAs */}
      <div className="flex items-center gap-2">
        <Link
          href="/login"
          className="text-[13px] px-3 py-1.5 transition-colors duration-150"
          style={{ color: 'rgba(255,255,255,0.55)' }}
        >
          Sign in
        </Link>
        <Link
          href="/register"
          className="text-[13px] font-semibold text-white px-4 py-1.5 rounded-lg transition-[filter] duration-150 hover:brightness-110"
          style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
        >
          Start free →
        </Link>
      </div>
    </nav>
  )
}
