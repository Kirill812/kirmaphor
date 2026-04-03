'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname, useRouter } from 'next/navigation'
import {
  FolderOpen, Zap, LayoutGrid, Lock, Users,
  Settings, ChevronLeft, ChevronRight,
} from 'lucide-react'
import { useAuth } from '@/lib/auth-store'

const NAV_MAIN = [
  { href: '/projects',  label: 'Projects',   icon: FolderOpen, badge: null },
  { href: '/runs',      label: 'Runs',        icon: Zap,        badge: 'live' as const },
  { href: '/templates', label: 'Templates',   icon: LayoutGrid, badge: null },
  { href: '/secrets',   label: 'Secrets',     icon: Lock,       badge: null },
]

const NAV_TEAM = [
  { href: '/members', label: 'Members', icon: Users, badge: null },
]

export function Sidebar() {
  const [collapsed, setCollapsed] = useState(false)
  const pathname = usePathname()
  const { user, logout } = useAuth()
  const router = useRouter()

  const width = collapsed ? 56 : 220

  const handleLogout = () => {
    logout()
    router.push('/login')
  }

  const isActive = (href: string) => pathname === href || pathname.startsWith(href + '/')

  return (
    <aside
      className="flex flex-col shrink-0 transition-all duration-200"
      style={{
        width,
        minHeight: '100vh',
        backgroundColor: 'var(--dash-surface)',
        borderRight: '1px solid var(--dash-border)',
      }}
    >
      {/* Workspace switcher */}
      <div
        style={{
          padding: collapsed ? '14px 8px' : '14px 12px',
          borderBottom: '1px solid rgba(255,255,255,0.05)',
        }}
      >
        <div
          className="flex items-center gap-2 rounded-lg"
          style={{
            padding: collapsed ? '6px' : '6px 8px',
            background: 'rgba(255,255,255,0.04)',
            justifyContent: collapsed ? 'center' : undefined,
          }}
          title={collapsed ? 'My Workspace' : undefined}
        >
          <div
            className="shrink-0 flex items-center justify-center rounded-md text-white font-bold text-xs"
            style={{
              width: 22, height: 22,
              background: 'var(--dash-gradient)',
              borderRadius: 5,
            }}
          >
            K
          </div>
          {!collapsed && (
            <>
              <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--dash-text)', flex: 1 }}>
                My Workspace
              </span>
              <span style={{ fontSize: 10, color: 'var(--dash-text-3)' }}>⌄</span>
            </>
          )}
        </div>
      </div>

      {/* Nav */}
      <nav className="flex flex-col flex-1 gap-px" style={{ padding: collapsed ? '8px 4px' : '8px' }}>
        {NAV_MAIN.map(({ href, label, icon: Icon, badge }) => {
          const active = isActive(href)
          return (
            <Link
              key={href}
              href={href}
              className="flex items-center rounded-md relative"
              aria-current={active ? 'page' : undefined}
              style={{
                gap: collapsed ? 0 : 8,
                padding: collapsed ? '7px 0' : '7px 8px',
                justifyContent: collapsed ? 'center' : undefined,
                background: active ? 'rgba(99,102,241,0.15)' : 'transparent',
                transition: 'background 0.1s',
              }}
              title={collapsed ? label : undefined}
            >
              {active && (
                <span
                  className="absolute"
                  style={{
                    left: 0, top: '50%', transform: 'translateY(-50%)',
                    width: 2, height: 18,
                    background: 'var(--dash-indigo)',
                    borderRadius: '0 2px 2px 0',
                  }}
                />
              )}
              <Icon
                size={16}
                strokeWidth={1.75}
                style={{ color: active ? '#818cf8' : 'rgba(255,255,255,0.35)', flexShrink: 0 }}
              />
              {!collapsed && (
                <>
                  <span
                    style={{
                      fontSize: 13,
                      flex: 1,
                      color: active ? '#e0e0ff' : 'rgba(255,255,255,0.4)',
                      fontWeight: active ? 500 : 400,
                    }}
                  >
                    {label}
                  </span>
                  {badge === 'live' && <span className="dash-live-dot" />}
                </>
              )}
            </Link>
          )
        })}

        {/* Team section */}
        {!collapsed && (
          <span
            style={{
              fontSize: 9, fontWeight: 600, letterSpacing: '0.08em',
              textTransform: 'uppercase', color: 'var(--dash-text-3)',
              padding: '8px 8px 4px',
            }}
          >
            Team
          </span>
        )}
        {NAV_TEAM.map(({ href, label, icon: Icon }) => {
          const active = isActive(href)
          return (
            <Link
              key={href}
              href={href}
              className="flex items-center rounded-md relative"
              aria-current={active ? 'page' : undefined}
              style={{
                gap: collapsed ? 0 : 8,
                padding: collapsed ? '7px 0' : '7px 8px',
                justifyContent: collapsed ? 'center' : undefined,
                background: active ? 'rgba(99,102,241,0.15)' : 'transparent',
              }}
              title={collapsed ? label : undefined}
            >
              {active && (
                <span
                  className="absolute"
                  style={{
                    left: 0, top: '50%', transform: 'translateY(-50%)',
                    width: 2, height: 18,
                    background: 'var(--dash-indigo)',
                    borderRadius: '0 2px 2px 0',
                  }}
                />
              )}
              <Icon
                size={16}
                strokeWidth={1.75}
                style={{ color: active ? '#818cf8' : 'rgba(255,255,255,0.35)', flexShrink: 0 }}
              />
              {!collapsed && (
                <span
                  style={{
                    fontSize: 13, flex: 1,
                    color: active ? '#e0e0ff' : 'rgba(255,255,255,0.4)',
                    fontWeight: active ? 500 : 400,
                  }}
                >
                  {label}
                </span>
              )}
            </Link>
          )
        })}
      </nav>

      {/* Collapse button */}
      <button
        onClick={() => setCollapsed(c => !c)}
        className="flex items-center gap-2 cursor-pointer rounded-md"
        style={{
          margin: collapsed ? '0 4px 8px' : '0 8px 8px',
          padding: '5px 8px',
          border: '1px solid rgba(255,255,255,0.06)',
          background: 'transparent',
          color: 'rgba(255,255,255,0.25)',
          fontSize: 11,
          justifyContent: collapsed ? 'center' : undefined,
        }}
      >
        {collapsed
          ? <ChevronRight size={12} strokeWidth={1.75} />
          : <><ChevronLeft size={12} strokeWidth={1.75} />Collapse</>
        }
      </button>

      {/* Bottom: settings + user */}
      <div style={{ borderTop: '1px solid rgba(255,255,255,0.05)', padding: collapsed ? '10px 4px' : '10px 12px' }}>
        <Link
          href="/settings"
          className="flex items-center rounded-md"
          style={{
            gap: collapsed ? 0 : 8,
            padding: collapsed ? '5px 0' : '5px 8px',
            justifyContent: collapsed ? 'center' : undefined,
            marginBottom: 4,
          }}
          title={collapsed ? 'Settings' : undefined}
        >
          <Settings size={16} strokeWidth={1.75} style={{ color: 'rgba(255,255,255,0.3)', flexShrink: 0 }} />
          {!collapsed && (
            <span style={{ fontSize: 13, color: 'rgba(255,255,255,0.3)' }}>Settings</span>
          )}
        </Link>

        <div
          className="flex items-center cursor-pointer rounded-md"
          style={{
            gap: collapsed ? 0 : 8,
            padding: collapsed ? '4px 0' : '4px 8px',
            justifyContent: collapsed ? 'center' : undefined,
          }}
          onClick={handleLogout}
          title={collapsed ? `${user?.display_name} — Sign out` : undefined}
        >
          <div
            className="shrink-0 flex items-center justify-center rounded-full text-white font-bold"
            style={{
              width: 26, height: 26,
              background: 'var(--dash-gradient)',
              fontSize: 11,
            }}
          >
            {user?.display_name?.[0]?.toUpperCase() ?? '?'}
          </div>
          {!collapsed && (
            <>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 12, color: 'var(--dash-text)', fontWeight: 500 }}>
                  {user?.display_name}
                </div>
                <div style={{ fontSize: 10, color: 'var(--dash-text-3)' }}>Free plan</div>
              </div>
              <span style={{ fontSize: 11, color: 'var(--dash-text-3)', flexShrink: 0 }}>
                Sign out
              </span>
            </>
          )}
        </div>
      </div>
    </aside>
  )
}
