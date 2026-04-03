// web/components/landing/HeroSection.tsx
import Link from 'next/link'
import React from 'react'

export default function HeroSection() {
  return (
    <section
      className="relative flex flex-col items-center justify-center overflow-hidden px-6 pt-32 pb-20"
      style={{ minHeight: '100vh', backgroundColor: 'var(--land-bg)' }}
    >
      {/* Glow blobs */}
      <div
        className="animate-land-pulse pointer-events-none absolute"
        style={{
          top: '25%', left: '50%', transform: 'translate(-50%, -50%)',
          width: '700px', height: '700px', borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(6,182,212,0.10) 0%, transparent 65%)',
        }}
      />
      <div
        className="animate-land-pulse pointer-events-none absolute"
        style={{
          top: '20%', left: '-10%',
          width: '400px', height: '400px', borderRadius: '50%',
          background: 'radial-gradient(circle, rgba(129,140,248,0.07) 0%, transparent 70%)',
          animationDelay: '2s',
        }}
      />

      {/* Badge */}
      <div
        className="relative z-10 mb-7 flex items-center gap-2 rounded-full px-4 py-1.5 text-[11px] tracking-widest"
        style={{
          background: 'rgba(6,182,212,0.10)',
          border: '1px solid rgba(6,182,212,0.25)',
          color: '#06b6d4',
        }}
      >
        <span
          className="animate-land-pulse h-1.5 w-1.5 rounded-full"
          style={{ background: '#06b6d4', animationDuration: '2s' }}
        />
        Now in open beta · AI-native · Self-hostable
      </div>

      {/* Headline */}
      <h1
        className="relative z-10 mb-5 text-center font-extrabold leading-[1.02] tracking-[-0.04em]"
        style={{ fontSize: 'clamp(48px, 7vw, 88px)', color: 'rgba(255,255,255,0.95)' }}
      >
        Ansible automation<br />
        <span
          style={{
            background: 'linear-gradient(135deg, #06b6d4, #818cf8)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            backgroundClip: 'text',
          }}
        >
          that thinks with you.
        </span>
      </h1>

      {/* Sub */}
      <p
        className="relative z-10 mb-9 text-center text-[17px] leading-relaxed"
        style={{ color: 'rgba(255,255,255,0.42)', maxWidth: '520px' }}
      >
        Run playbooks, review them with AI, stream live logs —<br />
        without the complexity of AWX or the price of Ansible Tower.
      </p>

      {/* CTAs */}
      <div className="relative z-10 mb-16 flex items-center gap-3">
        <Link
          href="/register"
          className="flex h-11 items-center gap-2 rounded-xl px-6 text-[14px] font-semibold text-white transition-[filter] duration-150 hover:brightness-110"
          style={{
            background: 'linear-gradient(135deg, #06b6d4, #0e7490)',
            boxShadow: '0 0 32px rgba(6,182,212,0.25)',
          }}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
          </svg>
          Start free
        </Link>
        <a
          href="https://github.com/Kirill812/kirmaphor"
          target="_blank"
          rel="noopener noreferrer"
          className="flex h-11 items-center gap-2 rounded-xl border px-6 text-[14px] font-medium text-white/75 transition-[background] duration-150 hover:bg-white/[0.09]"
          style={{ background: 'rgba(255,255,255,0.06)', borderColor: 'rgba(255,255,255,0.10)' }}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
          </svg>
          View on GitHub
        </a>
      </div>

      {/* Live product preview */}
      <div
        className="relative z-10 w-full overflow-hidden rounded-2xl"
        style={{
          maxWidth: '900px',
          background: 'rgba(255,255,255,0.03)',
          border: '1px solid rgba(255,255,255,0.08)',
          boxShadow: '0 40px 120px rgba(0,0,0,0.6), 0 0 60px rgba(6,182,212,0.06)',
        }}
      >
        {/* Browser chrome */}
        <div
          className="flex items-center gap-2 px-4"
          style={{
            height: '36px',
            background: 'rgba(255,255,255,0.04)',
            borderBottom: '1px solid rgba(255,255,255,0.06)',
          }}
        >
          <div className="h-2.5 w-2.5 rounded-full bg-[#ff5f57]" />
          <div className="h-2.5 w-2.5 rounded-full bg-[#ffbd2e]" />
          <div className="h-2.5 w-2.5 rounded-full bg-[#28ca42]" />
          <span className="flex-1 text-center text-[11px]" style={{ color: 'rgba(255,255,255,0.18)' }}>
            kirmaphore.io/projects/production
          </span>
        </div>

        {/* 3-column dashboard */}
        <div className="grid" style={{ gridTemplateColumns: '180px 1fr 260px', height: '340px' }}>
          {/* Sidebar */}
          <div className="p-4" style={{ borderRight: '1px solid rgba(255,255,255,0.06)' }}>
            <SidebarSection label="Projects">
              <SidebarItem label="production" active running />
              <SidebarItem label="staging" />
              <SidebarItem label="dev-cluster" />
            </SidebarSection>
            <SidebarSection label="Templates">
              <SidebarItem label="deploy.yml" />
              <SidebarItem label="nginx-setup.yml" />
            </SidebarSection>
            <SidebarSection label="Scheduled">
              <SidebarItem label="cert-renew" />
            </SidebarSection>
          </div>

          {/* Log stream */}
          <div className="p-5 overflow-hidden">
            <div className="flex items-center justify-between mb-4">
              <span className="text-[13px] font-semibold" style={{ color: 'rgba(255,255,255,0.85)' }}>
                ▶ Run #847 — deploy.yml
              </span>
              <span
                className="text-[10px] rounded-full px-2 py-0.5"
                style={{
                  background: 'rgba(6,182,212,0.15)',
                  border: '1px solid rgba(6,182,212,0.3)',
                  color: '#06b6d4',
                }}
              >
                ⟳ Running
              </span>
            </div>
            <div className="font-mono text-[11px] leading-[1.8]">
              <LogLine color="rgba(255,255,255,0.35)">PLAY [Deploy application] ***</LogLine>
              <LogLine color="#4ade80">✓ TASK [Pull latest image] — ok (2.1s)</LogLine>
              <LogLine color="#4ade80">✓ TASK [Stop old containers] — ok (0.4s)</LogLine>
              <LogLine color="#4ade80">✓ TASK [Start new containers] — ok (1.8s)</LogLine>
              <LogLine color="#4ade80">✓ TASK [Run migrations] — ok (3.2s)</LogLine>
              <div className="flex items-center gap-1" style={{ color: '#06b6d4' }}>
                <span>⟳ TASK [Health check] — running</span>
                <span
                  className="animate-land-blink inline-block"
                  style={{ width: '7px', height: '13px', background: '#06b6d4', borderRadius: '1px' }}
                />
              </div>
              <LogLine color="rgba(255,255,255,0.3)">{'  '}waiting for /health to return 200...</LogLine>
              <div className="mt-2" style={{ color: '#818cf8' }}>
                ✦ Forge AI — analysing run pattern…
              </div>
              <LogLine color="rgba(129,140,248,0.7)">{'  '}health check timeout detected in similar runs</LogLine>
            </div>
          </div>

          {/* Forge AI panel */}
          <div
            className="p-4"
            style={{
              borderLeft: '1px solid rgba(255,255,255,0.06)',
              background: 'rgba(129,140,248,0.03)',
            }}
          >
            <div className="flex items-center gap-2 mb-4">
              <div
                className="h-4 w-4 rounded"
                style={{ background: 'linear-gradient(135deg, #818cf8, #06b6d4)' }}
              />
              <span className="text-[11px] font-semibold" style={{ color: '#818cf8' }}>Forge AI</span>
            </div>
            <div
              className="rounded-lg p-3 mb-2 text-[10px] leading-[1.6]"
              style={{
                background: 'rgba(129,140,248,0.08)',
                border: '1px solid rgba(129,140,248,0.15)',
              }}
            >
              <div className="font-semibold mb-1.5" style={{ color: '#818cf8' }}>⚠ Pattern detected</div>
              <div style={{ color: 'rgba(255,255,255,0.45)' }}>
                Health check timed out in 3 of last 5 deploys. Container needs 8s warm-up but timeout is 5s.
              </div>
            </div>
            <div
              className="rounded-lg p-3 text-[10px] leading-[1.6]"
              style={{
                background: 'rgba(6,182,212,0.06)',
                border: '1px solid rgba(6,182,212,0.15)',
              }}
            >
              <div className="font-semibold mb-1.5" style={{ color: '#06b6d4' }}>Suggested fix</div>
              <pre className="font-mono" style={{ color: 'rgba(255,255,255,0.55)' }}>
{`wait_for:
  port: 8080
  delay: 8
  timeout: 30`}
              </pre>
            </div>
          </div>
        </div>
      </div>

      <p className="relative z-10 mt-4 text-[12px]" style={{ color: 'rgba(255,255,255,0.18)' }}>
        Live product · no illustrations
      </p>
    </section>
  )
}

function SidebarSection({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="mb-4">
      <div
        className="mb-1.5 text-[9px] uppercase tracking-widest"
        style={{ color: 'rgba(255,255,255,0.25)', letterSpacing: '0.1em' }}
      >
        {label}
      </div>
      {children}
    </div>
  )
}

function SidebarItem({ label, active, running }: { label: string; active?: boolean; running?: boolean }) {
  return (
    <div
      className="flex items-center gap-2 rounded-md px-2 py-1.5 text-[11px] mb-0.5"
      style={{
        background: active ? 'rgba(6,182,212,0.10)' : 'transparent',
        color: active ? '#06b6d4' : 'rgba(255,255,255,0.4)',
      }}
    >
      <span
        className={running ? 'animate-land-pulse' : ''}
        style={{
          display: 'inline-block',
          width: '6px', height: '6px', borderRadius: '50%', flexShrink: 0,
          background: running ? '#06b6d4' : active ? '#4ade80' : 'rgba(255,255,255,0.2)',
        }}
      />
      {label}
    </div>
  )
}

function LogLine({ color, children }: { color: string; children: React.ReactNode }) {
  return <div style={{ color }}>{children}</div>
}
