# Landing Page Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a premium 2026-style public landing page at `/` for Kirmaphore — live product hero, bento features, Forge AI deep dive, comparison table, and pricing.

**Architecture:** `app/(marketing)/page.tsx` assembles 10 standalone section components from `components/landing/`. The current `/` (dashboard stub) moves to `/dashboard` within the existing `(dashboard)` route group. Middleware is updated to allow public access to `/`. All animations are CSS-first with an IntersectionObserver scroll-reveal hook.

**Tech Stack:** Next.js 16 App Router, TypeScript, Tailwind CSS v4, inline CSS custom properties, SVG icons (stroke-width: 2.75), no animation libraries.

---

## ⚠️ Read First

Before writing any code, read `web/AGENTS.md` — this Next.js version has breaking changes from your training data.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `web/middleware.ts` | Modify | Add `'/'` to PUBLIC_PATHS |
| `web/app/(dashboard)/dashboard/page.tsx` | Create | Move dashboard stub to `/dashboard` |
| `web/app/(dashboard)/page.tsx` | Delete | Replaced by above |
| `web/components/auth/PasskeyLoginForm.tsx` | Modify | Redirect to `/dashboard` after login |
| `web/components/auth/EmailLoginForm.tsx` | Modify | Redirect to `/dashboard` after login |
| `web/components/auth/PasskeyRegisterForm.tsx` | Modify | Redirect to `/dashboard` after register |
| `web/components/auth/EmailRegisterForm.tsx` | Modify | Redirect to `/dashboard` after register |
| `web/app/globals.css` | Modify | Add `--land-*` CSS vars + animations |
| `web/components/landing/LandingNav.tsx` | Create | Sticky nav with logo + links + CTAs |
| `web/components/landing/HeroSection.tsx` | Create | Headline + live product preview |
| `web/components/landing/PainSection.tsx` | Create | 4-card competitor pain grid |
| `web/components/landing/FeaturesBento.tsx` | Create | Asymmetric 12-col bento grid |
| `web/components/landing/HowItWorksSection.tsx` | Create | 3-step horizontal layout |
| `web/components/landing/ForgeAISection.tsx` | Create | 3-card AI capabilities + terminal mockup |
| `web/components/landing/ComparisonSection.tsx` | Create | Feature comparison table |
| `web/components/landing/PricingSection.tsx` | Create | Community vs Pro cards |
| `web/components/landing/FooterCTASection.tsx` | Create | Final CTA block |
| `web/components/landing/LandingFooter.tsx` | Create | Minimal footer |
| `web/components/landing/ScrollReveal.tsx` | Create | IntersectionObserver reveal wrapper |
| `web/app/(marketing)/page.tsx` | Create | Assembles all sections |

---

## Task 1: Route Architecture

**Context:** Currently `app/(dashboard)/page.tsx` serves `/` and is protected. We need `/` to be the public landing. This task makes all routing changes before any landing component exists — no visual breakage possible because the dashboard at `/` is a 4-line stub.

**Files:**
- Modify: `web/middleware.ts`
- Create: `web/app/(dashboard)/dashboard/page.tsx`
- Delete: `web/app/(dashboard)/page.tsx`
- Modify: `web/components/auth/PasskeyLoginForm.tsx`
- Modify: `web/components/auth/EmailLoginForm.tsx`
- Modify: `web/components/auth/PasskeyRegisterForm.tsx`
- Modify: `web/components/auth/EmailRegisterForm.tsx`

- [ ] **Step 1: Update middleware PUBLIC_PATHS**

Read `web/middleware.ts`, then replace:

```typescript
// web/middleware.ts
import { NextRequest, NextResponse } from 'next/server'

const PUBLIC_PATHS = ['/login', '/register', '/']

export function middleware(req: NextRequest) {
  const token = req.cookies.get('kirmaphore_token')?.value
  const isPublic = PUBLIC_PATHS.some(p => req.nextUrl.pathname === p || req.nextUrl.pathname.startsWith(p + '/') && p !== '/')

  if (!token && !isPublic) {
    return NextResponse.redirect(new URL('/login', req.url))
  }
  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}
```

Wait — the original `startsWith` check would make `/` match everything. Use exact match for `/`:

```typescript
// web/middleware.ts
import { NextRequest, NextResponse } from 'next/server'

export function middleware(req: NextRequest) {
  const token = req.cookies.get('kirmaphore_token')?.value
  const path = req.nextUrl.pathname

  const isPublic =
    path === '/' ||
    path === '/login' ||
    path === '/register' ||
    path.startsWith('/login/') ||
    path.startsWith('/register/')

  if (!token && !isPublic) {
    return NextResponse.redirect(new URL('/login', req.url))
  }
  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
}
```

- [ ] **Step 2: Create dashboard page at new route**

Create `web/app/(dashboard)/dashboard/page.tsx`:

```tsx
// web/app/(dashboard)/dashboard/page.tsx
export default function DashboardPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-2">Welcome to Kirmaphore</h1>
      <p className="text-muted-foreground">Select a project to get started.</p>
    </div>
  )
}
```

- [ ] **Step 3: Delete old dashboard page**

```bash
rm /path/to/web/app/\(dashboard\)/page.tsx
```

Use the Bash tool:
```bash
rm /Users/kgory/dev/kirmaphor/web/app/\(dashboard\)/page.tsx
```

- [ ] **Step 4: Update auth redirects — PasskeyLoginForm**

In `web/components/auth/PasskeyLoginForm.tsx`, change `router.push('/')` → `router.push('/dashboard')`.

The line reads:
```typescript
      setUser(res.user, res.token)
      router.push('/')
```
Replace with:
```typescript
      setUser(res.user, res.token)
      router.push('/dashboard')
```

- [ ] **Step 5: Update auth redirects — EmailLoginForm**

In `web/components/auth/EmailLoginForm.tsx`, change:
```typescript
      await login(email, password)
      router.push('/')
```
To:
```typescript
      await login(email, password)
      router.push('/dashboard')
```

- [ ] **Step 6: Update auth redirects — PasskeyRegisterForm**

In `web/components/auth/PasskeyRegisterForm.tsx`, change:
```typescript
      setUser(res.user, res.token)
      router.push('/')
```
To:
```typescript
      setUser(res.user, res.token)
      router.push('/dashboard')
```

- [ ] **Step 7: Update auth redirects — EmailRegisterForm**

In `web/components/auth/EmailRegisterForm.tsx`, change:
```typescript
      setUser(res.user, res.token)
      router.push('/')
```
To:
```typescript
      setUser(res.user, res.token)
      router.push('/dashboard')
```

- [ ] **Step 8: TypeScript check**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 9: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/middleware.ts web/app/\(dashboard\)/dashboard/ web/components/auth/
git commit -m "refactor: move dashboard to /dashboard, open / for public landing"
```

---

## Task 2: Landing CSS Tokens + Animations

**Context:** All landing components use `--land-*` CSS custom properties inline (`style={{ color: 'var(--land-cyan)' }}`). They also rely on three animation classes: `animate-land-pulse` (glow blobs), `animate-land-blink` (log cursor), `animate-land-reveal` (scroll reveals). Add these to `globals.css`.

**Files:**
- Modify: `web/app/globals.css`

- [ ] **Step 1: Add tokens and animations**

Append to the end of `web/app/globals.css` (after the existing `slide-in` keyframes):

```css
/* Landing page */
:root {
  --land-bg:      #060A14;
  --land-cyan:    #06b6d4;
  --land-violet:  #818cf8;
}

@keyframes land-pulse {
  0%, 100% { opacity: 0.7; }
  50%       { opacity: 1.0; }
}

@keyframes land-blink {
  0%, 50%   { opacity: 1; }
  51%, 100% { opacity: 0; }
}

@keyframes land-reveal {
  from { opacity: 0; transform: translateY(16px); }
  to   { opacity: 1; transform: translateY(0); }
}

.animate-land-pulse  { animation: land-pulse 4s ease-in-out infinite; }
.animate-land-blink  { animation: land-blink 1s step-end infinite; }
.animate-land-reveal { animation: land-reveal 400ms ease-out forwards; }

@media (prefers-reduced-motion: reduce) {
  .animate-land-pulse,
  .animate-land-blink,
  .animate-land-reveal { animation: none; opacity: 1; transform: none; }
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/app/globals.css
git commit -m "feat: add landing CSS tokens and animations"
```

---

## Task 3: LandingNav

**Context:** Sticky top nav, `height: 56px`, glassmorphism background. Left: wordmark. Center: 4 links (hidden on mobile). Right: "Sign in" + "Start free →". No client-side interactivity needed — pure server component.

**Files:**
- Create: `web/components/landing/LandingNav.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/landing/LandingNav.tsx
import Link from 'next/link'

export default function LandingNav() {
  return (
    <nav
      className="fixed top-0 left-0 right-0 z-50 flex items-center justify-between px-6 lg:px-12"
      style={{
        height: '56px',
        background: 'rgba(6,10,20,0.8)',
        backdropFilter: 'blur(12px)',
        borderBottom: '1px solid rgba(255,255,255,0.06)',
      }}
    >
      {/* Logo */}
      <span className="text-white font-bold text-[15px] tracking-tight">Kirmaphore</span>

      {/* Center links — hidden on mobile */}
      <div className="hidden lg:flex items-center gap-8">
        {(['Features', 'Pricing', 'Docs'] as const).map(label => (
          <a
            key={label}
            href={`#${label.toLowerCase()}`}
            className="text-[13px] transition-colors duration-150"
            style={{ color: 'rgba(255,255,255,0.45)' }}
            onMouseOver={e => (e.currentTarget.style.color = 'white')}
            onMouseOut={e => (e.currentTarget.style.color = 'rgba(255,255,255,0.45)')}
          >
            {label}
          </a>
        ))}
        <a
          href="https://github.com/Kirill812/kirmaphor"
          target="_blank"
          rel="noopener noreferrer"
          className="text-[13px] transition-colors duration-150"
          style={{ color: 'rgba(255,255,255,0.45)' }}
          onMouseOver={e => (e.currentTarget.style.color = 'white')}
          onMouseOut={e => (e.currentTarget.style.color = 'rgba(255,255,255,0.45)')}
        >
          GitHub
        </a>
      </div>

      {/* CTAs */}
      <div className="flex items-center gap-2">
        <Link
          href="/login"
          className="text-[13px] px-3 py-1.5 transition-colors duration-150"
          style={{ color: 'rgba(255,255,255,0.55)' }}
        >
          Sign in
        </Link>
        <Link
          href="/register"
          className="text-[13px] font-semibold text-white px-4 py-1.5 rounded-lg transition-[filter] duration-150 hover:brightness-110"
          style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
        >
          Start free →
        </Link>
      </div>
    </nav>
  )
}
```

- [ ] **Step 2: TypeScript check**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/LandingNav.tsx
git commit -m "feat: LandingNav sticky navigation"
```

---

## Task 4: HeroSection

**Context:** The most complex component. Full-viewport hero with two glow blobs, badge, headline with gradient accent, sub text, two CTAs, and a live product preview (browser chrome + 3-column dashboard mockup with animated log cursor and Forge AI panel).

**Files:**
- Create: `web/components/landing/HeroSection.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/landing/HeroSection.tsx
import Link from 'next/link'

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
```

- [ ] **Step 2: TypeScript check**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/HeroSection.tsx
git commit -m "feat: HeroSection with live product preview"
```

---

## Task 5: PainSection

**Context:** 4-column grid, each card calls out a competitor's core weakness. Dark cards, minimal typography.

**Files:**
- Create: `web/components/landing/PainSection.tsx`

- [ ] **Step 1: Create the file**

```tsx
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
```

- [ ] **Step 2: TypeScript check**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/PainSection.tsx
git commit -m "feat: PainSection competitor pain cards"
```

---

## Task 6: FeaturesBento

**Context:** Asymmetric 12-column grid. Forge AI card spans 7 columns × 2 rows. Live logs spans 5 × 2. Six smaller cards fill the remaining rows. SVG icons: `stroke-width: 2.75`, `fill: none`, simple paths. Animated log cursor in Live Logs card.

**Files:**
- Create: `web/components/landing/FeaturesBento.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/landing/FeaturesBento.tsx

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
```

- [ ] **Step 2: TypeScript check**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```

- [ ] **Step 3: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/FeaturesBento.tsx
git commit -m "feat: FeaturesBento asymmetric bento grid"
```

---

## Task 7: HowItWorksSection

**Context:** 3 numbered steps in a row with a horizontal connector line. Each step has a numbered circle, title, description, and a small monospace code snippet.

**Files:**
- Create: `web/components/landing/HowItWorksSection.tsx`

- [ ] **Step 1: Create the file**

```tsx
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
```

- [ ] **Step 2: TypeScript check + commit**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/HowItWorksSection.tsx
git commit -m "feat: HowItWorksSection 3-step layout"
```

---

## Task 8: ForgeAISection

**Context:** 3-column card row with one AI capability per card, plus a large terminal mockup below with tab switcher (Review / Explain / Generate) — client component for tab interactivity.

**Files:**
- Create: `web/components/landing/ForgeAISection.tsx`

- [ ] **Step 1: Create the file**

```tsx
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
```

- [ ] **Step 2: TypeScript check + commit**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/ForgeAISection.tsx
git commit -m "feat: ForgeAISection with tabbed terminal mockup"
```

---

## Task 9: ComparisonSection

**Context:** Feature comparison table. Kirmaphore column has a cyan-tinted header and subtle left border highlight. Checkmarks are green, crosses are dim. Server component.

**Files:**
- Create: `web/components/landing/ComparisonSection.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/landing/ComparisonSection.tsx

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
```

- [ ] **Step 2: TypeScript check + commit**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/ComparisonSection.tsx
git commit -m "feat: ComparisonSection feature comparison table"
```

---

## Task 10: PricingSection

**Context:** Community vs Pro cards side by side. Community is plain bordered. Pro has cyan border + glow. Pro CTA is "Join waitlist →" with an email input (client component for input state). GitHub stars badge.

**Files:**
- Create: `web/components/landing/PricingSection.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/landing/PricingSection.tsx
'use client'

import { useState } from 'react'

const COMMUNITY_FEATURES = [
  'All core automation features',
  'Unlimited projects',
  'Docker Compose deploy',
  'Community support',
  'MIT license — own your data',
]

const PRO_FEATURES = [
  'Everything in Community',
  'Managed cloud hosting',
  'Forge AI (unlimited)',
  'Priority support',
  'SSO / SAML',
]

export default function PricingSection() {
  const [email, setEmail] = useState('')
  const [submitted, setSubmitted] = useState(false)

  return (
    <section
      className="px-6 py-20 lg:px-12"
      style={{ backgroundColor: 'var(--land-bg)', borderTop: '1px solid rgba(255,255,255,0.06)' }}
    >
      <p className="mb-3 text-[11px] uppercase tracking-[0.12em]" style={{ color: 'rgba(6,182,212,0.7)' }}>
        Open source
      </p>
      <h2
        className="mb-3 font-bold tracking-[-0.03em]"
        style={{ fontSize: 'clamp(28px, 3.5vw, 44px)', color: 'rgba(255,255,255,0.9)' }}
      >
        Self-host for free.<br />Scale with cloud.
      </h2>
      <p className="mb-12 text-[15px] leading-relaxed" style={{ color: 'rgba(255,255,255,0.38)', maxWidth: '480px' }}>
        Start on your own server today. Cloud hosting coming soon.
      </p>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2 lg:max-w-3xl">
        {/* Community */}
        <div
          className="rounded-2xl p-8"
          style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid rgba(255,255,255,0.10)' }}
        >
          <p className="mb-1 text-[13px] font-semibold" style={{ color: 'rgba(255,255,255,0.88)' }}>Community</p>
          <p className="mb-6 text-[12px]" style={{ color: 'rgba(255,255,255,0.35)' }}>Free forever · MIT license</p>
          <ul className="mb-8 space-y-2.5">
            {COMMUNITY_FEATURES.map(f => (
              <li key={f} className="flex items-start gap-2 text-[13px]" style={{ color: 'rgba(255,255,255,0.55)' }}>
                <span style={{ color: '#4ade80', marginTop: '1px' }}>✓</span>
                {f}
              </li>
            ))}
          </ul>
          <a
            href="https://github.com/Kirill812/kirmaphor"
            target="_blank"
            rel="noopener noreferrer"
            className="flex h-11 w-full items-center justify-center rounded-xl text-[14px] font-medium transition-[background] duration-150 hover:bg-white/[0.09]"
            style={{
              background: 'rgba(255,255,255,0.06)',
              border: '1px solid rgba(255,255,255,0.10)',
              color: 'rgba(255,255,255,0.75)',
            }}
          >
            Deploy yourself →
          </a>
        </div>

        {/* Pro */}
        <div
          className="relative rounded-2xl p-8 overflow-hidden"
          style={{
            background: 'rgba(6,182,212,0.04)',
            border: '1px solid rgba(6,182,212,0.3)',
            boxShadow: '0 0 60px rgba(6,182,212,0.06)',
          }}
        >
          <p className="mb-1 text-[13px] font-semibold" style={{ color: 'rgba(255,255,255,0.88)' }}>Pro</p>
          <p className="mb-6 text-[12px]" style={{ color: 'rgba(6,182,212,0.7)' }}>Cloud-hosted · coming soon</p>
          <ul className="mb-8 space-y-2.5">
            {PRO_FEATURES.map(f => (
              <li key={f} className="flex items-start gap-2 text-[13px]" style={{ color: 'rgba(255,255,255,0.55)' }}>
                <span style={{ color: '#06b6d4', marginTop: '1px' }}>✓</span>
                {f}
              </li>
            ))}
          </ul>
          {submitted ? (
            <p className="text-center text-[13px]" style={{ color: '#4ade80' }}>
              ✓ You&apos;re on the list!
            </p>
          ) : (
            <div className="flex gap-2">
              <input
                type="email"
                placeholder="you@company.com"
                value={email}
                onChange={e => setEmail(e.target.value)}
                className="flex-1 rounded-xl px-3 text-[13px] text-white placeholder-gray-500 outline-none h-11"
                style={{
                  background: 'rgba(255,255,255,0.06)',
                  border: '1px solid rgba(255,255,255,0.10)',
                }}
                onKeyDown={e => { if (e.key === 'Enter' && email) setSubmitted(true) }}
              />
              <button
                onClick={() => { if (email) setSubmitted(true) }}
                className="h-11 rounded-xl px-4 text-[13px] font-semibold text-white transition-[filter] duration-150 hover:brightness-110"
                style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
              >
                Join waitlist →
              </button>
            </div>
          )}
        </div>
      </div>

      {/* GitHub stars badge */}
      <div className="mt-10 flex items-center gap-3">
        <a
          href="https://github.com/Kirill812/kirmaphor"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-2 rounded-full px-4 py-2 text-[12px] transition-[background] duration-150 hover:bg-white/[0.06]"
          style={{
            background: 'rgba(255,255,255,0.04)',
            border: '1px solid rgba(255,255,255,0.08)',
            color: 'rgba(255,255,255,0.5)',
          }}
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23A11.509 11.509 0 0112 5.803c1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576C20.566 21.797 24 17.3 24 12c0-6.627-5.373-12-12-12z" />
          </svg>
          Star on GitHub
        </a>
      </div>
    </section>
  )
}
```

- [ ] **Step 2: TypeScript check + commit**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/PricingSection.tsx
git commit -m "feat: PricingSection community/pro cards with waitlist"
```

---

## Task 11: FooterCTASection + LandingFooter

**Context:** FooterCTASection is the large final push block with a glow blob and two CTAs. LandingFooter is a minimal 3-column bar.

**Files:**
- Create: `web/components/landing/FooterCTASection.tsx`
- Create: `web/components/landing/LandingFooter.tsx`

- [ ] **Step 1: Create FooterCTASection**

```tsx
// web/components/landing/FooterCTASection.tsx
import Link from 'next/link'

export default function FooterCTASection() {
  return (
    <section
      className="relative overflow-hidden px-6 py-32 text-center lg:px-12"
      style={{ backgroundColor: 'var(--land-bg)', borderTop: '1px solid rgba(255,255,255,0.06)' }}
    >
      {/* Glow */}
      <div
        className="animate-land-pulse pointer-events-none absolute inset-0 flex items-center justify-center"
        aria-hidden="true"
      >
        <div
          style={{
            width: '600px', height: '600px', borderRadius: '50%',
            background: 'radial-gradient(circle, rgba(6,182,212,0.10) 0%, transparent 65%)',
          }}
        />
      </div>

      <h2
        className="relative z-10 mb-4 font-extrabold tracking-[-0.04em]"
        style={{ fontSize: 'clamp(36px, 5vw, 64px)', color: 'rgba(255,255,255,0.95)' }}
      >
        Automate everything.<br />Understand everything.
      </h2>
      <p
        className="relative z-10 mb-10 text-[17px] leading-relaxed"
        style={{ color: 'rgba(255,255,255,0.38)', maxWidth: '440px', margin: '0 auto 40px' }}
      >
        Your infrastructure, finally thinking with you.
      </p>

      <div className="relative z-10 flex items-center justify-center gap-3">
        <Link
          href="/register"
          className="flex h-12 items-center gap-2 rounded-xl px-8 text-[15px] font-semibold text-white transition-[filter] duration-150 hover:brightness-110"
          style={{
            background: 'linear-gradient(135deg, #06b6d4, #0e7490)',
            boxShadow: '0 0 40px rgba(6,182,212,0.30)',
          }}
        >
          Start free →
        </Link>
        <a
          href="https://github.com/Kirill812/kirmaphor"
          target="_blank"
          rel="noopener noreferrer"
          className="flex h-12 items-center gap-2 rounded-xl border px-8 text-[15px] font-medium transition-[background] duration-150 hover:bg-white/[0.09]"
          style={{
            background: 'rgba(255,255,255,0.06)',
            borderColor: 'rgba(255,255,255,0.10)',
            color: 'rgba(255,255,255,0.75)',
          }}
        >
          View on GitHub
        </a>
      </div>

      <p
        className="relative z-10 mt-6 text-[13px]"
        style={{ color: 'rgba(255,255,255,0.18)' }}
      >
        No credit card. No Kubernetes. Just docker compose up.
      </p>
    </section>
  )
}
```

- [ ] **Step 2: Create LandingFooter**

```tsx
// web/components/landing/LandingFooter.tsx
export default function LandingFooter() {
  return (
    <footer
      className="flex flex-col items-center justify-between gap-3 px-6 py-6 text-[12px] sm:flex-row lg:px-12"
      style={{
        borderTop: '1px solid rgba(255,255,255,0.06)',
        backgroundColor: 'var(--land-bg)',
        color: 'rgba(255,255,255,0.22)',
      }}
    >
      <span>Kirmaphore · © 2026</span>
      <div className="flex gap-6">
        <a href="#" className="hover:text-white/50 transition-colors">Docs</a>
        <a href="https://github.com/Kirill812/kirmaphor" target="_blank" rel="noopener noreferrer" className="hover:text-white/50 transition-colors">GitHub</a>
        <a href="#" className="hover:text-white/50 transition-colors">Twitter</a>
      </div>
      <span>Built with ♥ for infrastructure engineers</span>
    </footer>
  )
}
```

- [ ] **Step 3: TypeScript check + commit**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/FooterCTASection.tsx web/components/landing/LandingFooter.tsx
git commit -m "feat: FooterCTASection and LandingFooter"
```

---

## Task 12: ScrollReveal + Landing Page Root

**Context:** `ScrollReveal` is a lightweight client wrapper that uses `IntersectionObserver` to add `animate-land-reveal` when an element enters the viewport. The landing page root assembles all sections in order and wraps each in `<ScrollReveal>`.

**Files:**
- Create: `web/components/landing/ScrollReveal.tsx`
- Create: `web/app/(marketing)/page.tsx`

- [ ] **Step 1: Create ScrollReveal**

```tsx
// web/components/landing/ScrollReveal.tsx
'use client'

import { useEffect, useRef } from 'react'

interface Props {
  children: React.ReactNode
  delay?: number
}

export default function ScrollReveal({ children, delay = 0 }: Props) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return

    // Skip animation if user prefers reduced motion
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      el.style.opacity = '1'
      return
    }

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.style.animationDelay = `${delay}ms`
          el.classList.add('animate-land-reveal')
          observer.disconnect()
        }
      },
      { threshold: 0.08 }
    )
    observer.observe(el)
    return () => observer.disconnect()
  }, [delay])

  return (
    <div ref={ref} style={{ opacity: 0 }}>
      {children}
    </div>
  )
}
```

- [ ] **Step 2: Create landing page root**

```tsx
// web/app/(marketing)/page.tsx
import LandingNav from '@/components/landing/LandingNav'
import HeroSection from '@/components/landing/HeroSection'
import PainSection from '@/components/landing/PainSection'
import FeaturesBento from '@/components/landing/FeaturesBento'
import HowItWorksSection from '@/components/landing/HowItWorksSection'
import ForgeAISection from '@/components/landing/ForgeAISection'
import ComparisonSection from '@/components/landing/ComparisonSection'
import PricingSection from '@/components/landing/PricingSection'
import FooterCTASection from '@/components/landing/FooterCTASection'
import LandingFooter from '@/components/landing/LandingFooter'
import ScrollReveal from '@/components/landing/ScrollReveal'

export default function LandingPage() {
  return (
    <main style={{ backgroundColor: '#060A14' }}>
      <LandingNav />
      <HeroSection />
      <ScrollReveal><PainSection /></ScrollReveal>
      <ScrollReveal delay={50}><FeaturesBento /></ScrollReveal>
      <ScrollReveal delay={50}><HowItWorksSection /></ScrollReveal>
      <ScrollReveal delay={50}><ForgeAISection /></ScrollReveal>
      <ScrollReveal delay={50}><ComparisonSection /></ScrollReveal>
      <ScrollReveal delay={50}><PricingSection /></ScrollReveal>
      <ScrollReveal delay={50}><FooterCTASection /></ScrollReveal>
      <LandingFooter />
    </main>
  )
}
```

- [ ] **Step 3: TypeScript check**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/landing/ScrollReveal.tsx web/app/\(marketing\)/page.tsx
git commit -m "feat: landing page root with ScrollReveal"
```

---

## Task 13: Production Build Verification

- [ ] **Step 1: Run production build**

```bash
cd /Users/kgory/dev/kirmaphor/web && npm run build 2>&1
```

Expected output includes:
```
Route (app)
├ ○ /
├ ○ /login
├ ○ /register
└ ○ /dashboard
```
And: `✓ Compiled successfully`

- [ ] **Step 2: Fix any build errors**

Common issues and fixes:

**Missing `'use client'`** on a component using `useState`/`useEffect` → add directive at top.

**`onMouseOver`/`onMouseOut` in server component** (LandingNav uses these) → if build complains, add `'use client'` to LandingNav.tsx.

**`next-intl` error on public route** → the root layout wraps everything in `NextIntlClientProvider`. If `/` fails with next-intl, add a `(marketing)/layout.tsx` that bypasses it:

```tsx
// web/app/(marketing)/layout.tsx
export default function MarketingLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}
```

This overrides the root layout for the `(marketing)` group only — the root layout still runs above it.

**Actually**, in Next.js App Router nested layouts are additive (they don't replace root layout). Root layout always runs. This pattern won't bypass `NextIntlClientProvider`. The simpler fix if next-intl causes issues: wrap next-intl locale detection in a try/catch in the root layout, or handle missing messages gracefully. Most likely it will just work since the provider is lenient about missing translations.

- [ ] **Step 3: Verify landing page visually**

```bash
cd /Users/kgory/dev/kirmaphor/web && npm run dev &
```

Open `http://localhost:3000`:
- Nav is sticky, glassmorphism visible on scroll
- Hero: badge + headline with gradient + live dashboard preview
- Scroll down: pain cards, bento grid, how it works, forge AI section, comparison table, pricing, footer CTA
- Scroll reveals trigger as sections enter viewport
- `http://localhost:3000/login` still works
- `http://localhost:3000/dashboard` requires auth (redirects to login)

Kill dev server.

- [ ] **Step 4: Commit any fixes**

```bash
cd /Users/kgory/dev/kirmaphor && git add -A && git commit -m "fix: resolve landing page build issues"
```
(Only if fixes were needed.)

---

## Self-Review

| Spec section | Task |
|---|---|
| § 2 Route architecture | Task 1 |
| § 3 CSS tokens + animations | Task 2 |
| § 4 NavBar | Task 3 |
| § 5 Hero (badge, headline, CTAs, live preview) | Task 4 |
| § 6 Pain section | Task 5 |
| § 7 Features bento | Task 6 |
| § 8 How it works | Task 7 |
| § 9 Forge AI section | Task 8 |
| § 10 Comparison table | Task 9 |
| § 11 Pricing / open-core | Task 10 |
| § 12 Footer CTA | Task 11 |
| § 13 Footer | Task 11 |
| § 15 Animations (pulse, blink, reveal) | Tasks 2, 4, 12 |
| Icon spec (stroke 2.75, simple paths) | Tasks 4, 6 |
