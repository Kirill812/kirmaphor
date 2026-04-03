const FEATURES = [
  { label: 'AI playbook review',       k: true,  s: false, a: false, t: false },
  { label: 'AI failure explanation',   k: true,  s: false, a: false, t: false },
  { label: 'AI task generation',       k: true,  s: false, a: false, t: false },
  { label: 'Live WebSocket logs',      k: true,  s: false, a: true,  t: true  },
  { label: 'Passkey / WebAuthn',       k: true,  s: false, a: false, t: false },
  { label: 'Visual DAG designer',      k: true,  s: false, a: false, t: true  },
  { label: 'Self-hostable',            k: true,  s: true,  a: true,  t: false },
  { label: 'Free tier',                k: true,  s: true,  a: true,  t: false },
  { label: 'Docker Compose deploy',    k: true,  s: true,  a: false, t: false },
  { label: 'Price (team of 10)',       k: 'Free', s: 'Free', a: 'Free', t: '~$13K/yr' },
]

function Cell({ value }: { value: boolean | string }) {
  if (typeof value === 'string') {
    return (
      <td className="py-3 px-4 text-center text-[13px]" style={{ color: 'rgba(255,255,255,0.55)' }}>
        {value}
      </td>
    )
  }
  return (
    <td className="py-3 px-4 text-center text-[15px]">
      {value
        ? <span style={{ color: '#4ade80' }}>✓</span>
        : <span style={{ color: 'rgba(255,255,255,0.2)' }}>✗</span>
      }
    </td>
  )
}

export default function ComparisonSection() {
  return (
    <section
      id="pricing"
      className="px-6 py-20 lg:px-12"
      style={{ backgroundColor: 'var(--land-bg)', borderTop: '1px solid rgba(255,255,255,0.06)' }}
    >
      <p className="mb-3 text-[11px] uppercase tracking-[0.12em]" style={{ color: 'rgba(6,182,212,0.7)' }}>
        Compare
      </p>
      <h2
        className="mb-12 font-bold tracking-[-0.03em]"
        style={{ fontSize: 'clamp(28px, 3.5vw, 44px)', color: 'rgba(255,255,255,0.9)' }}
      >
        See how we stack up.
      </h2>

      <div className="overflow-x-auto">
        <table className="w-full text-left" style={{ borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.08)' }}>
              <th className="py-3 px-4 text-[12px] font-normal" style={{ color: 'rgba(255,255,255,0.3)', width: '35%' }}>
                Feature
              </th>
              <th
                className="py-3 px-4 text-center text-[13px] font-semibold"
                style={{
                  color: '#06b6d4',
                  background: 'rgba(6,182,212,0.06)',
                  borderLeft: '2px solid rgba(6,182,212,0.4)',
                }}
              >
                Kirmaphore
              </th>
              <th className="py-3 px-4 text-center text-[12px] font-normal" style={{ color: 'rgba(255,255,255,0.35)' }}>
                Semaphore UI
              </th>
              <th className="py-3 px-4 text-center text-[12px] font-normal" style={{ color: 'rgba(255,255,255,0.35)' }}>
                AWX
              </th>
              <th className="py-3 px-4 text-center text-[12px] font-normal" style={{ color: 'rgba(255,255,255,0.35)' }}>
                Ansible Tower
              </th>
            </tr>
          </thead>
          <tbody>
            {FEATURES.map(({ label, k, s, a, t }) => (
              <tr
                key={label}
                style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}
              >
                <td className="py-3 px-4 text-[13px]" style={{ color: 'rgba(255,255,255,0.6)' }}>
                  {label}
                </td>
                <Cell value={k} />
                <Cell value={s} />
                <Cell value={a} />
                <Cell value={t} />
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
