# Dashboard UI & Onboarding Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the bare white dashboard with a premium dark UI (Vibrant Pro: indigo/violet) including a collapsible sidebar, hero empty state for projects, and quick-start cards.

**Architecture:** CSS custom properties drive the entire color system. The dashboard layout switches to a dark `#0d0d1a` background. Sidebar becomes a client component with `useState` for collapsed/expanded. A new `/projects` page replaces `/dashboard` as the home screen.

**Tech Stack:** Next.js 16 App Router, TypeScript, Tailwind CSS v4 (inline styles for non-standard values), Lucide React icons, Zustand (auth-store for user data in sidebar).

---

## File Map

| Action | File |
|---|---|
| Modify | `web/app/globals.css` |
| Modify | `web/app/(dashboard)/layout.tsx` |
| Rewrite | `web/components/layout/Sidebar.tsx` |
| Create | `web/components/layout/Topbar.tsx` |
| Modify | `web/app/(dashboard)/dashboard/page.tsx` |
| Create | `web/app/(dashboard)/projects/page.tsx` |
| Modify | `web/middleware.ts` |

---

## Task 1: CSS Design Tokens

**Files:**
- Modify: `web/app/globals.css`

Add the `--dash-*` variables after the existing `:root` block. These are used via inline styles throughout the dashboard (Tailwind v4 doesn't have direct arbitrary CSS variable support in the way we need).

- [ ] **Step 1: Add dash tokens to globals.css**

Open `web/app/globals.css`. After the closing `}` of the `.dark` block (around line 100), append:

```css
/* ── Dashboard Vibrant Pro design tokens ── */
:root {
  --dash-bg:         #0d0d1a;
  --dash-surface:    rgba(99, 102, 241, 0.04);
  --dash-border:     rgba(99, 102, 241, 0.10);
  --dash-border-hi:  rgba(99, 102, 241, 0.25);

  --dash-indigo:     #6366f1;
  --dash-violet:     #8b5cf6;
  --dash-gradient:   linear-gradient(135deg, #6366f1, #8b5cf6);

  --dash-text:       #ffffff;
  --dash-text-2:     rgba(255, 255, 255, 0.50);
  --dash-text-3:     rgba(255, 255, 255, 0.25);

  --dash-green:      #22c55e;
  --dash-yellow:     #f59e0b;
  --dash-red:        #ef4444;
}

@keyframes dash-pulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 6px rgba(34, 197, 94, 0.6); }
  50%       { opacity: 0.6; box-shadow: 0 0 3px rgba(34, 197, 94, 0.3); }
}
.dash-live-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--dash-green);
  animation: dash-pulse 2s ease-in-out infinite;
}
```

- [ ] **Step 2: Verify build**

```bash
cd web && npm run build 2>&1 | tail -5
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add web/app/globals.css
git commit -m "feat: add dashboard Vibrant Pro CSS design tokens"
```

---

## Task 2: Dashboard Layout

**Files:**
- Modify: `web/app/(dashboard)/layout.tsx`
- Create: `web/components/layout/Topbar.tsx`

The layout switches to dark background and removes the old Header. Topbar is now a slot passed per-page via a shared context — but for simplicity, Topbar is rendered directly in layout and pages set the title via a `<title>` pattern. Actually simpler: Topbar accepts `title` and `actions` props, but since layout.tsx is a Server Component it can't pass dynamic props. **Solution:** each page renders its own `<Topbar>` inside the content area. Layout just provides the dark shell.

- [ ] **Step 1: Rewrite layout.tsx**

Replace entire `web/app/(dashboard)/layout.tsx` with:

```tsx
import { Sidebar } from '@/components/layout/Sidebar'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="flex min-h-screen"
      style={{ backgroundColor: 'var(--dash-bg)', color: 'var(--dash-text)' }}
    >
      <Sidebar />
      <div className="flex-1 flex flex-col min-w-0">
        {children}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Create Topbar component**

Create `web/components/layout/Topbar.tsx`:

```tsx
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
```

- [ ] **Step 3: Build check**

```bash
cd web && npm run build 2>&1 | tail -5
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add web/app/(dashboard)/layout.tsx web/components/layout/Topbar.tsx
git commit -m "feat: dark dashboard layout shell + Topbar component"
```

---

## Task 3: Sidebar — Hybrid Collapsible

**Files:**
- Rewrite: `web/components/layout/Sidebar.tsx`

Full rewrite. Uses `useState` for `collapsed`. Nav links updated to new routes. User block at bottom pulls from `useAuth`.

- [ ] **Step 1: Rewrite Sidebar.tsx**

Replace entire `web/components/layout/Sidebar.tsx` with:

```tsx
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
  { href: '/projects',  label: 'Projects',   icon: FolderOpen, badge: 'count' as const },
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
          className="flex items-center gap-2 cursor-pointer rounded-lg"
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
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontSize: 12, color: 'var(--dash-text)', fontWeight: 500 }}>
                {user?.display_name}
              </div>
              <div style={{ fontSize: 10, color: 'var(--dash-text-3)' }}>Free plan</div>
            </div>
          )}
        </div>
      </div>
    </aside>
  )
}
```

- [ ] **Step 2: Build check**

```bash
cd web && npm run build 2>&1 | grep -E 'error|Error|✓|✗' | head -20
```

Expected: `✓ Compiled successfully` or similar, no TypeScript errors.

- [ ] **Step 3: Commit**

```bash
git add web/components/layout/Sidebar.tsx
git commit -m "feat: hybrid collapsible sidebar with Vibrant Pro design"
```

---

## Task 4: Projects Page — Hero Empty State

**Files:**
- Create: `web/app/(dashboard)/projects/page.tsx`
- Modify: `web/app/(dashboard)/dashboard/page.tsx`
- Modify: `web/middleware.ts`

- [ ] **Step 1: Create projects page**

Create `web/app/(dashboard)/projects/page.tsx`:

```tsx
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
              href="https://kirmaphore.kirvps.work/docs"
              style={{ fontSize: 12, color: 'rgba(99,102,241,0.7)', textDecoration: 'underline' }}
            >
              View examples
            </a>
            <a
              href="https://kirmaphore.kirvps.work/docs"
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
```

- [ ] **Step 2: Update dashboard/page.tsx to redirect to /projects**

Replace `web/app/(dashboard)/dashboard/page.tsx` with:

```tsx
import { redirect } from 'next/navigation'

export default function DashboardPage() {
  redirect('/projects')
}
```

- [ ] **Step 3: Add /projects to middleware protected routes**

Open `web/middleware.ts`. The current `isPublic` check protects everything except the listed paths. Since `/projects` is not in the public list, it's already protected — no change needed. Verify:

```bash
grep -n 'isPublic\|projects' web/middleware.ts
```

Expected: `/projects` does NOT appear in the `isPublic` block (confirming it's protected by default).

- [ ] **Step 4: Build check**

```bash
cd web && npm run build 2>&1 | grep -E 'error|Error|✓|Route' | head -30
```

Expected: `/projects` appears in routes list, no errors.

- [ ] **Step 5: Commit**

```bash
git add web/app/(dashboard)/projects/page.tsx web/app/(dashboard)/dashboard/page.tsx
git commit -m "feat: projects page with hero empty state and quick-start cards"
```

---

## Task 5: Wire Up & Deploy

**Files:**
- No new files — deployment only

- [ ] **Step 1: Final local build**

```bash
cd web && npm run build 2>&1 | tail -10
```

Expected: exit 0, no TypeScript or build errors.

- [ ] **Step 2: Push to main**

```bash
git push
```

- [ ] **Step 3: Deploy to server**

```bash
ssh root@51.255.193.130 "cd /opt/kirmaphore && git pull && docker compose build --no-cache && docker compose up -d"
```

- [ ] **Step 4: Smoke test**

```bash
curl -s -o /dev/null -w "%{http_code}" https://kirmaphore.kirvps.work/api/health
```

Expected: `200`.

Open https://kirmaphore.kirvps.work/login in browser, log in → should land on `/projects` with dark UI and new sidebar.
