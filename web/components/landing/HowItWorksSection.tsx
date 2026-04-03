// web/components/landing/HowItWorksSection.tsx

const STEPS = [
  {
    n: '1',
    title: 'Connect your inventory',
    desc: 'Add servers, point to your Git repo. SSH keys stored encrypted.',
    code: ['host: prod-01.example.com', 'user: deploy', 'key: ••••••••'],
    colors: ['rgba(6,182,212,0.15)', 'rgba(6,182,212,0.35)', '#06b6d4'],
  },
  {
    n: '2',
    title: 'Run — AI reviews first',
    desc: 'Forge AI scans your playbook before a single task executes.',
    code: ['✦ review passed', '  no issues found', '▶ running deploy.yml'],
    colors: ['rgba(129,140,248,0.15)', 'rgba(129,140,248,0.35)', '#818cf8'],
  },
  {
    n: '3',
    title: 'Watch it live',
    desc: 'Logs stream in real-time. If it fails, AI explains why and suggests the fix.',
    code: ['✓ deployed', '✓ healthy', 'total: 14.2s'],
    colors: ['rgba(74,222,128,0.10)', 'rgba(74,222,128,0.25)', '#4ade80'],
  },
]

export default function HowItWorksSection() {
  return (
    <section
      className="px-6 py-20 lg:px-12"
      style={{ backgroundColor: 'var(--land-bg)', borderTop: '1px solid rgba(255,255,255,0.06)' }}
    >
      <p className="mb-3 text-[11px] uppercase tracking-[0.12em]" style={{ color: 'rgba(6,182,212,0.7)' }}>
        How it works
      </p>
      <h2
        className="mb-16 font-bold tracking-[-0.03em]"
        style={{ fontSize: 'clamp(28px, 3.5vw, 44px)', color: 'rgba(255,255,255,0.9)' }}
      >
        From zero to automated<br />in three steps.
      </h2>

      <div className="relative grid grid-cols-1 gap-10 lg:grid-cols-3 lg:gap-0">
        {/* Connector line (desktop only) */}
        <div
          className="pointer-events-none absolute hidden lg:block"
          style={{
            top: '27px', left: 'calc(16% + 20px)', right: 'calc(16% + 20px)',
            height: '1px',
            background: 'linear-gradient(90deg, transparent, rgba(6,182,212,0.25), transparent)',
          }}
        />

        {STEPS.map(({ n, title, desc, code, colors }) => (
          <div key={n} className="relative z-10 flex flex-col items-center text-center lg:px-8">
            {/* Circle */}
            <div
              className="mb-5 flex h-[54px] w-[54px] items-center justify-center rounded-full text-[16px] font-bold"
              style={{
                background: colors[0],
                border: `1px solid ${colors[1]}`,
                color: colors[2],
              }}
            >
              {n}
            </div>
            <h3 className="mb-2.5 text-[15px] font-semibold" style={{ color: 'rgba(255,255,255,0.88)' }}>
              {title}
            </h3>
            <p className="mb-4 text-[13px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)' }}>
              {desc}
            </p>
            <div
              className="w-full rounded-lg p-3 text-left font-mono text-[10px] leading-[1.8]"
              style={{
                background: 'rgba(255,255,255,0.04)',
                border: '1px solid rgba(255,255,255,0.07)',
                color: 'rgba(255,255,255,0.45)',
              }}
            >
              {code.map((line, i) => <div key={i}>{line}</div>)}
            </div>
          </div>
        ))}
      </div>
    </section>
  )
}
