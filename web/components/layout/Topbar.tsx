import { ReactNode } from 'react'

interface TopbarProps {
  title: string
  actions?: ReactNode
}

export function Topbar({ title, actions }: TopbarProps) {
  return (
    <div
      className="flex items-center justify-between px-6 shrink-0"
      style={{
        height: '48px',
        borderBottom: '1px solid rgba(255,255,255,0.05)',
        backgroundColor: 'rgba(99,102,241,0.02)',
      }}
    >
      <span style={{ fontSize: '15px', fontWeight: 600, color: 'var(--dash-text)' }}>
        {title}
      </span>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  )
}
