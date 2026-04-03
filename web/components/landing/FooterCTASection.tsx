// web/components/landing/FooterCTASection.tsx
import Link from 'next/link'

export default function FooterCTASection() {
  return (
    <section
      className="relative overflow-hidden px-6 py-32 text-center lg:px-12"
      style={{ backgroundColor: 'var(--land-bg)', borderTop: '1px solid rgba(255,255,255,0.06)' }}
    >
      {/* Glow */}
      <div
        className="animate-land-pulse pointer-events-none absolute inset-0 flex items-center justify-center"
        aria-hidden="true"
      >
        <div
          style={{
            width: '600px', height: '600px', borderRadius: '50%',
            background: 'radial-gradient(circle, rgba(6,182,212,0.10) 0%, transparent 65%)',
          }}
        />
      </div>

      <h2
        className="relative z-10 mb-4 font-extrabold tracking-[-0.04em]"
        style={{ fontSize: 'clamp(36px, 5vw, 64px)', color: 'rgba(255,255,255,0.95)' }}
      >
        Automate everything.<br />Understand everything.
      </h2>
      <p
        className="relative z-10 mb-10 text-[17px] leading-relaxed"
        style={{ color: 'rgba(255,255,255,0.38)', maxWidth: '440px', margin: '0 auto 40px' }}
      >
        Your infrastructure, finally thinking with you.
      </p>

      <div className="relative z-10 flex items-center justify-center gap-3">
        <Link
          href="/register"
          className="flex h-12 items-center gap-2 rounded-xl px-8 text-[15px] font-semibold text-white transition-[filter] duration-150 hover:brightness-110"
          style={{
            background: 'linear-gradient(135deg, #06b6d4, #0e7490)',
            boxShadow: '0 0 40px rgba(6,182,212,0.30)',
          }}
        >
          Start free →
        </Link>
        <a
          href="https://github.com/Kirill812/kirmaphor"
          target="_blank"
          rel="noopener noreferrer"
          className="flex h-12 items-center gap-2 rounded-xl border px-8 text-[15px] font-medium transition-[background] duration-150 hover:bg-white/[0.09]"
          style={{
            background: 'rgba(255,255,255,0.06)',
            borderColor: 'rgba(255,255,255,0.10)',
            color: 'rgba(255,255,255,0.75)',
          }}
        >
          View on GitHub
        </a>
      </div>

      <p
        className="relative z-10 mt-6 text-[13px]"
        style={{ color: 'rgba(255,255,255,0.18)' }}
      >
        No credit card. No Kubernetes. Just docker compose up.
      </p>
    </section>
  )
}
