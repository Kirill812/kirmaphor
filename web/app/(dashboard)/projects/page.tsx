'use client'

import Link from 'next/link'
import { Topbar } from '@/components/layout/Topbar'

function QuickStartCard({
  emoji,
  emojiBackground,
  title,
  description,
  href,
}: {
  emoji: string
  emojiBackground: string
  title: string
  description: string
  href: string
}) {
  return (
    <Link
      href={href}
      className="block rounded-xl transition-all duration-150"
      style={{
        padding: 14,
        background: 'rgba(255,255,255,0.02)',
        border: '1px solid rgba(255,255,255,0.06)',
        textDecoration: 'none',
      }}
      onMouseEnter={e => {
        const el = e.currentTarget as HTMLElement
        el.style.borderColor = 'rgba(99,102,241,0.30)'
        el.style.background = 'rgba(99,102,241,0.05)'
      }}
      onMouseLeave={e => {
        const el = e.currentTarget as HTMLElement
        el.style.borderColor = 'rgba(255,255,255,0.06)'
        el.style.background = 'rgba(255,255,255,0.02)'
      }}
    >
      <div
        className="flex items-center justify-center rounded-lg"
        style={{ width: 28, height: 28, background: emojiBackground, fontSize: 14, marginBottom: 8 }}
      >
        {emoji}
      </div>
      <div style={{ fontSize: 12, fontWeight: 600, color: 'rgba(255,255,255,0.8)', marginBottom: 3 }}>
        {title}
      </div>
      <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.3)', lineHeight: 1.4 }}>
        {description}
      </div>
    </Link>
  )
}

export default function ProjectsPage() {
  return (
    <>
      <Topbar
        title="Projects"
        actions={
          <button
            className="rounded-lg font-semibold text-white"
            style={{
              padding: '6px 14px',
              fontSize: 12,
              background: 'var(--dash-gradient)',
              border: 'none',
              cursor: 'pointer',
              boxShadow: '0 2px 12px rgba(99,102,241,0.4)',
            }}
          >
            + New Project
          </button>
        }
      />

      {/* Hero empty state */}
      <div
        className="flex flex-col items-center justify-center flex-1"
        style={{ gap: 20, padding: '40px 24px 20px' }}
      >
        {/* Glow circle */}
        <div
          className="flex items-center justify-center relative"
          style={{
            width: 72, height: 72,
            background: 'linear-gradient(135deg, rgba(99,102,241,0.25), rgba(139,92,246,0.15))',
            border: '1px solid rgba(99,102,241,0.4)',
            borderRadius: '50%',
            fontSize: 28,
            boxShadow: '0 0 40px rgba(99,102,241,0.25), 0 0 80px rgba(99,102,241,0.10)',
          }}
        >
          ⚡
          <span
            className="absolute inset-0 rounded-full pointer-events-none"
            style={{
              margin: '-8px',
              border: '1px solid rgba(99,102,241,0.15)',
              borderRadius: '50%',
            }}
          />
        </div>

        {/* Text */}
        <div className="text-center">
          <h1 style={{ fontSize: 20, fontWeight: 700, color: 'var(--dash-text)', marginBottom: 6 }}>
            Deploy your first project
          </h1>
          <p style={{ fontSize: 13, color: 'rgba(255,255,255,0.4)', maxWidth: 320, lineHeight: 1.6 }}>
            Connect an Ansible playbook, define your inventory, and automate infrastructure in minutes.
          </p>
        </div>

        {/* CTA */}
        <div className="flex flex-col items-center" style={{ gap: 10 }}>
          <button
            className="font-semibold text-white"
            style={{
              padding: '10px 28px',
              borderRadius: 24,
              fontSize: 14,
              background: 'var(--dash-gradient)',
              border: 'none',
              cursor: 'pointer',
              boxShadow: '0 4px 20px rgba(99,102,241,0.5)',
            }}
          >
            + Create Project
          </button>
          <div className="flex items-center" style={{ gap: 16 }}>
            <a
              href="/docs"
              style={{ fontSize: 12, color: 'rgba(99,102,241,0.7)', textDecoration: 'underline' }}
            >
              View examples
            </a>
            <a
              href="/docs"
              style={{ fontSize: 12, color: 'rgba(99,102,241,0.7)', textDecoration: 'underline' }}
            >
              Read the docs
            </a>
          </div>
        </div>
      </div>

      {/* Quick start */}
      <div style={{ padding: '0 24px 24px' }}>
        <div
          style={{
            fontSize: 10, fontWeight: 600, letterSpacing: '0.1em',
            textTransform: 'uppercase', color: 'rgba(255,255,255,0.2)',
            marginBottom: 10,
          }}
        >
          Quick start
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3,1fr)', gap: 10 }}>
          <QuickStartCard
            emoji="📁"
            emojiBackground="rgba(99,102,241,0.12)"
            title="Connect a repository"
            description="Link your Git repo with Ansible playbooks"
            href="/projects/new?step=repository"
          />
          <QuickStartCard
            emoji="🖥️"
            emojiBackground="rgba(34,197,94,0.10)"
            title="Add inventory"
            description="Define your servers and environments"
            href="/projects/new?step=inventory"
          />
          <QuickStartCard
            emoji="▶️"
            emojiBackground="rgba(251,146,60,0.10)"
            title="Run a playbook"
            description="Execute your first automated deployment"
            href="/projects/new?step=run"
          />
        </div>
      </div>
    </>
  )
}
