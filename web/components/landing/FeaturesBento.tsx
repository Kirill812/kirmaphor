// web/components/landing/FeaturesBento.tsx
import React from 'react'

export default function FeaturesBento() {
  return (
    <section
      className="px-6 py-20 lg:px-12"
      style={{ backgroundColor: 'var(--land-bg)', borderTop: '1px solid rgba(255,255,255,0.06)' }}
    >
      <p className="mb-3 text-[11px] uppercase tracking-[0.12em]" style={{ color: 'rgba(6,182,212,0.7)' }}>
        Features
      </p>
      <h2
        className="mb-3 font-bold tracking-[-0.03em]"
        style={{ fontSize: 'clamp(28px, 3.5vw, 44px)', color: 'rgba(255,255,255,0.9)' }}
      >
        Everything you need.<br />Nothing you don&apos;t.
      </h2>
      <p className="mb-12 text-[15px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)', maxWidth: '460px' }}>
        Built for engineers who automate infrastructure daily — not for a procurement deck.
      </p>

      {/* 12-column asymmetric grid */}
      <div
        className="grid gap-3"
        style={{ gridTemplateColumns: 'repeat(12, 1fr)', gridAutoRows: '160px' }}
      >
        {/* Forge AI — 7 cols × 2 rows */}
        <BentoCard
          style={{ gridColumn: 'span 7', gridRow: 'span 2' }}
          ai
          icon={
            <svg viewBox="0 0 24 24" fill="none" stroke="rgba(129,140,248,0.95)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
              <path d="M12 2v5M12 17v5M2 12h5M17 12h5M5.6 5.6l3.5 3.5M14.9 14.9l3.5 3.5M18.4 5.6l-3.5 3.5M9.1 14.9l-3.5 3.5" />
            </svg>
          }
          title="Forge AI — built into every run"
          desc="Reviews playbooks before execution, explains failures in plain English, generates tasks from natural language."
        >
          <div
            className="mt-3 rounded-lg p-3 font-mono text-[10px] leading-[1.9]"
            style={{ background: 'rgba(0,0,0,0.3)', border: '1px solid rgba(255,255,255,0.06)' }}
          >
            <div style={{ color: 'rgba(255,255,255,0.22)' }}># pre-run review</div>
            <div style={{ color: '#f87171' }}>✗ no_log missing on password task (line 24)</div>
            <div style={{ color: '#4ade80' }}>✓ suggest: add no_log: true</div>
            <div style={{ color: 'rgba(255,255,255,0.15)', marginTop: '2px' }}>──────────────────</div>
            <div style={{ color: '#a78bfa' }}>» &quot;add nginx rate limiting&quot;</div>
            <div style={{ color: 'rgba(255,255,255,0.4)' }}>→ task generated in 1.2s</div>
          </div>
        </BentoCard>

        {/* Live logs — 5 cols × 2 rows */}
        <BentoCard
          style={{ gridColumn: 'span 5', gridRow: 'span 2' }}
          icon={
            <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
              <polyline points="3 12 7 12 9 5 13 19 15 12 21 12" />
            </svg>
          }
          title="Live log streaming"
          desc="WebSocket-powered real-time output. Every task, as it runs — no refresh."
        >
          <div className="mt-3 font-mono text-[10px] leading-[1.9]">
            <div style={{ color: '#4ade80' }}>✓ nginx.conf updated</div>
            <div style={{ color: '#4ade80' }}>✓ ssl cert renewed</div>
            <div style={{ color: '#4ade80' }}>✓ containers restarted</div>
            <div className="flex items-center gap-1" style={{ color: '#06b6d4' }}>
              <span>⟳ health check</span>
              <span
                className="animate-land-blink inline-block rounded-[1px]"
                style={{ width: '6px', height: '11px', background: '#06b6d4' }}
              />
            </div>
          </div>
        </BentoCard>

        {/* Git webhooks — 3 cols */}
        <BentoCard
          style={{ gridColumn: 'span 3' }}
          icon={
            <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
              <circle cx="18" cy="18" r="3" /><circle cx="6" cy="6" r="3" /><path d="M6 21V9a9 9 0 0 0 9 9" />
            </svg>
          }
          title="Git webhooks"
          desc="Push to branch → playbook runs automatically. CI/CD for infrastructure."
        />

        {/* RBAC — 4 cols */}
        <BentoCard
          style={{ gridColumn: 'span 4' }}
          icon={
            <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
              <path d="M12 3L4 7v5c0 5 4 9 8 10 4-1 8-5 8-10V7l-8-4z" />
            </svg>
          }
          title="Enterprise RBAC"
          desc="Roles, permissions, project isolation. Every action auditable."
        />

        {/* Visual DAG — 5 cols */}
        <BentoCard
          style={{ gridColumn: 'span 5' }}
          icon={
            <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
              <circle cx="5" cy="12" r="2.5" /><circle cx="19" cy="5" r="2.5" /><circle cx="19" cy="19" r="2.5" />
              <line x1="7.5" y1="12" x2="16.5" y2="6" /><line x1="7.5" y1="12" x2="16.5" y2="18" />
            </svg>
          }
          title="Visual DAG designer"
          desc="Drag & drop playbook pipelines. Parallel tasks, dependencies, conditions."
        >
          <div className="mt-3 flex items-center gap-1.5 flex-wrap">
            {['provision', 'configure', 'deploy', 'verify'].map((node, i) => (
              <div key={node} className="flex items-center gap-1.5">
                <span
                  className="rounded text-[9px] px-2 py-0.5"
                  style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.1)', color: 'rgba(255,255,255,0.45)' }}
                >
                  {node}
                </span>
                {i < 3 && <span style={{ color: 'rgba(255,255,255,0.18)', fontSize: '10px' }}>→</span>}
              </div>
            ))}
          </div>
        </BentoCard>

        {/* Secrets — 4 cols */}
        <BentoCard
          style={{ gridColumn: 'span 4' }}
          icon={
            <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
              <rect x="4" y="11" width="16" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" />
            </svg>
          }
          title="Secrets vault"
          desc="Encrypted per project. Injected at runtime — never logged."
        />

        {/* Scheduler — 4 cols */}
        <BentoCard
          style={{ gridColumn: 'span 4' }}
          icon={
            <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
              <circle cx="12" cy="12" r="9" /><polyline points="12 7 12 12 15 14" />
            </svg>
          }
          title="Cron scheduler"
          desc="Schedule any playbook. Visual editor, timezone-aware."
        />

        {/* Docker — 4 cols */}
        <BentoCard
          style={{ gridColumn: 'span 4' }}
          icon={
            <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
              <polyline points="21 8 12 3 3 8" /><polyline points="3 8 3 16 12 21 21 16 21 8" /><line x1="12" y1="3" x2="12" y2="21" />
            </svg>
          }
          title="One-command deploy"
          desc={<><code className="text-[10px]" style={{ color: '#06b6d4' }}>docker compose up</code> — running in 60 seconds.</>}
        />
      </div>
    </section>
  )
}

function BentoCard({
  children, icon, title, desc, ai, style,
}: {
  children?: React.ReactNode
  icon: React.ReactNode
  title: string
  desc: React.ReactNode
  ai?: boolean
  style?: React.CSSProperties
}) {
  return (
    <div
      className="relative overflow-hidden rounded-2xl p-[22px] transition-[border-color,background] duration-200"
      style={{
        background: ai
          ? 'linear-gradient(135deg, rgba(129,140,248,0.07), rgba(6,182,212,0.04))'
          : 'rgba(255,255,255,0.03)',
        border: `1px solid ${ai ? 'rgba(129,140,248,0.18)' : 'rgba(255,255,255,0.07)'}`,
        ...style,
      }}
    >
      {/* Icon container */}
      <div
        className="mb-3.5 flex h-9 w-9 items-center justify-center rounded-[9px]"
        style={{
          background: ai ? 'rgba(129,140,248,0.12)' : 'rgba(255,255,255,0.06)',
          border: `1px solid ${ai ? 'rgba(129,140,248,0.25)' : 'rgba(255,255,255,0.10)'}`,
        }}
      >
        {icon}
      </div>
      <p className="mb-1.5 text-[13px] font-semibold" style={{ color: 'rgba(255,255,255,0.88)' }}>
        {title}
      </p>
      <p className="text-[11px] leading-[1.6]" style={{ color: 'rgba(255,255,255,0.35)' }}>
        {desc}
      </p>
      {children}
    </div>
  )
}
