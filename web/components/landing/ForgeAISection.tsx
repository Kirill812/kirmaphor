// web/components/landing/ForgeAISection.tsx
'use client'

import { useState } from 'react'

const TABS = ['Review', 'Explain', 'Generate'] as const
type Tab = typeof TABS[number]

const TAB_CONTENT: Record<Tab, { lines: { text: string; color: string }[] }> = {
  Review: {
    lines: [
      { text: '$ forge review deploy.yml', color: 'rgba(255,255,255,0.5)' },
      { text: '  Scanning playbook...', color: 'rgba(255,255,255,0.3)' },
      { text: '✗ line 24: no_log not set on password task', color: '#f87171' },
      { text: '  → Secrets may appear in logs. Add no_log: true', color: '#f87171' },
      { text: '✗ line 41: shell module used, consider command', color: '#facc15' },
      { text: '  → shell spawns a subshell, command is safer', color: '#facc15' },
      { text: '✓ No deprecated modules found', color: '#4ade80' },
      { text: '✓ Idempotency check passed', color: '#4ade80' },
      { text: '  2 issues, 0 critical. Fix before running.', color: 'rgba(255,255,255,0.3)' },
    ],
  },
  Explain: {
    lines: [
      { text: '$ forge explain --run 847 --task "health check"', color: 'rgba(255,255,255,0.5)' },
      { text: '  Analysing failure context...', color: 'rgba(255,255,255,0.3)' },
      { text: '  ', color: 'rgba(255,255,255,0.3)' },
      { text: '✦ Root cause identified', color: '#818cf8' },
      { text: '  Container starts but needs 8s to warm up.', color: 'rgba(255,255,255,0.55)' },
      { text: '  wait_for timeout is set to 5s — too short.', color: 'rgba(255,255,255,0.55)' },
      { text: '  ', color: 'rgba(255,255,255,0.3)' },
      { text: '✦ Suggested fix:', color: '#06b6d4' },
      { text: '  wait_for:', color: '#06b6d4' },
      { text: '    delay: 8   # was: 5', color: '#06b6d4' },
      { text: '    timeout: 30', color: '#06b6d4' },
    ],
  },
  Generate: {
    lines: [
      { text: '$ forge generate', color: 'rgba(255,255,255,0.5)' },
      { text: '  Describe what you want:', color: 'rgba(255,255,255,0.3)' },
      { text: '  > add nginx rate limiting to port 80', color: 'rgba(255,255,255,0.7)' },
      { text: '  ', color: 'rgba(255,255,255,0.3)' },
      { text: '  Generating task... done (1.2s)', color: '#4ade80' },
      { text: '  ', color: 'rgba(255,255,255,0.3)' },
      { text: '- name: Apply nginx rate limiting', color: '#a78bfa' },
      { text: '  community.general.ini_file:', color: '#a78bfa' },
      { text: '    path: /etc/nginx/nginx.conf', color: '#a78bfa' },
      { text: '    section: http', color: '#a78bfa' },
      { text: '    option: limit_req_zone', color: '#a78bfa' },
      { text: '    value: "$binary_remote_addr zone=api:10m rate=10r/s"', color: '#a78bfa' },
    ],
  },
}

const CARDS = [
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
        <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
      </svg>
    ),
    title: 'Pre-run review',
    desc: 'Catches security issues, deprecated modules, and logic errors before execution. Like a linter, but it understands your playbook intent.',
    example: 'no_log missing on credential task → flagged with fix suggestion.',
  },
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
        <polyline points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
      </svg>
    ),
    title: 'Failure explanation',
    desc: 'When a task fails, AI reads the error, the task context, and your inventory — and tells you exactly what went wrong.',
    example: 'Health check timeout → root cause + YAML fix in plain English.',
  },
  {
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.8)" strokeWidth="2.75" strokeLinecap="round" strokeLinejoin="round" width="18" height="18">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
      </svg>
    ),
    title: 'Natural language generation',
    desc: 'Describe what you want in plain English. Forge AI writes the Ansible task — correct module, right options, ready to run.',
    example: '"add nginx rate limiting to port 80" → task YAML in 1.2s.',
  },
]

export default function ForgeAISection() {
  const [activeTab, setActiveTab] = useState<Tab>('Review')

  return (
    <section
      id="forge-ai"
      className="px-6 py-20 lg:px-12"
      style={{ backgroundColor: 'var(--land-bg)', borderTop: '1px solid rgba(255,255,255,0.06)' }}
    >
      <p className="mb-3 text-[11px] uppercase tracking-[0.12em]" style={{ color: 'rgba(129,140,248,0.8)' }}>
        Forge AI
      </p>
      <h2
        className="mb-3 font-bold tracking-[-0.03em]"
        style={{ fontSize: 'clamp(28px, 3.5vw, 44px)', color: 'rgba(255,255,255,0.9)' }}
      >
        Your playbooks, reviewed<br />before they run.
      </h2>
      <p className="mb-12 text-[15px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)', maxWidth: '520px' }}>
        Three capabilities, built into every project. No setup, no plugin, no extra cost.
      </p>

      {/* 3 capability cards */}
      <div className="mb-10 grid grid-cols-1 gap-3 lg:grid-cols-3">
        {CARDS.map(({ icon, title, desc, example }) => (
          <div
            key={title}
            className="rounded-2xl p-6"
            style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.07)' }}
          >
            <div
              className="mb-4 flex h-9 w-9 items-center justify-center rounded-[9px]"
              style={{ background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.10)' }}
            >
              {icon}
            </div>
            <h3 className="mb-2 text-[14px] font-semibold" style={{ color: 'rgba(255,255,255,0.88)' }}>{title}</h3>
            <p className="mb-3 text-[12px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)' }}>{desc}</p>
            <p className="text-[11px] italic" style={{ color: 'rgba(255,255,255,0.25)' }}>{example}</p>
          </div>
        ))}
      </div>

      {/* Terminal mockup with tab switcher */}
      <div
        className="overflow-hidden rounded-2xl"
        style={{ background: 'rgba(0,0,0,0.4)', border: '1px solid rgba(255,255,255,0.08)' }}
      >
        {/* Tab bar */}
        <div
          className="flex items-center gap-1 px-4 py-2"
          style={{ borderBottom: '1px solid rgba(255,255,255,0.06)' }}
        >
          <div className="flex gap-1.5 mr-4">
            <div className="h-2.5 w-2.5 rounded-full bg-[#ff5f57]" />
            <div className="h-2.5 w-2.5 rounded-full bg-[#ffbd2e]" />
            <div className="h-2.5 w-2.5 rounded-full bg-[#28ca42]" />
          </div>
          {TABS.map(tab => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className="rounded-md px-3 py-1 text-[12px] font-medium transition-colors duration-150"
              style={{
                background: activeTab === tab ? 'rgba(129,140,248,0.15)' : 'transparent',
                color: activeTab === tab ? '#818cf8' : 'rgba(255,255,255,0.35)',
                border: activeTab === tab ? '1px solid rgba(129,140,248,0.25)' : '1px solid transparent',
              }}
            >
              {tab}
            </button>
          ))}
        </div>
        {/* Terminal output */}
        <div className="p-5 font-mono text-[11px] leading-[1.8]" style={{ minHeight: '200px' }}>
          {TAB_CONTENT[activeTab].lines.map((line, i) => (
            <div key={i} style={{ color: line.color }}>{line.text || '\u00A0'}</div>
          ))}
        </div>
      </div>
    </section>
  )
}
