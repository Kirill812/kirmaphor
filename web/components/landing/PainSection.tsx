// web/components/landing/PainSection.tsx

const PAINS = [
  {
    competitor: 'AWX / TOWER',
    headline: 'Requires Kubernetes to run a cron job',
    body: 'Full-stack deployment just to schedule an Ansible playbook. Overkill for 90% of teams.',
  },
  {
    competitor: 'ANSIBLE TOWER',
    headline: '$13,000/year before you write a line',
    body: 'Enterprise pricing locks out every team under 500 engineers. Pay to see if it works.',
  },
  {
    competitor: 'SEMAPHORE UI',
    headline: '2019 UI, zero AI, zero insight',
    body: "Runs your playbooks. That's it. When a task fails at 2am you're on your own.",
  },
  {
    competitor: 'RUNDECK',
    headline: 'Acquired, rebranded, forgotten',
    body: 'Absorbed into PagerDuty. No roadmap, no momentum, support routes to a webinar.',
  },
]

export default function PainSection() {
  return (
    <section
      id="features"
      className="px-6 py-20 lg:px-12"
      style={{
        backgroundColor: 'var(--land-bg)',
        borderTop: '1px solid rgba(255,255,255,0.06)',
      }}
    >
      <p className="mb-3 text-[11px] uppercase tracking-[0.12em]" style={{ color: 'rgba(6,182,212,0.7)' }}>
        Why teams switch
      </p>
      <h2
        className="mb-3 font-bold tracking-[-0.03em]"
        style={{ fontSize: 'clamp(28px, 3.5vw, 44px)', color: 'rgba(255,255,255,0.9)' }}
      >
        Tired of the alternatives?
      </h2>
      <p className="mb-12 text-[15px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)', maxWidth: '480px' }}>
        Every existing tool makes you choose between power and simplicity. Kirmaphore doesn&apos;t.
      </p>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {PAINS.map(({ competitor, headline, body }) => (
          <div
            key={competitor}
            className="rounded-xl p-5"
            style={{
              background: 'rgba(255,255,255,0.03)',
              border: '1px solid rgba(255,255,255,0.07)',
            }}
          >
            <p className="mb-2 text-[9px] uppercase tracking-[0.08em]" style={{ color: 'rgba(255,255,255,0.2)' }}>
              {competitor}
            </p>
            <p className="mb-2 text-[14px] font-semibold" style={{ color: 'rgba(255,255,255,0.82)' }}>
              {headline}
            </p>
            <p className="text-[12px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)' }}>
              {body}
            </p>
          </div>
        ))}
      </div>
    </section>
  )
}
