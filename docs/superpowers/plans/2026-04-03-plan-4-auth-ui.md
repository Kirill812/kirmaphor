# Auth UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a premium split-layout auth UI (passkey-first, requesty.ai aesthetic) for `/login` and `/register`.

**Architecture:** `AuthLayout` wraps both auth pages with a 40/60 split: left branding panel (navy + cyan glow) and right form panel. Each form component (passkey / email) is self-contained with its own API calls, loading state, and inline errors. Login page is redesigned in-place; register page is created new.

**Tech Stack:** Next.js 16 App Router, TypeScript, Tailwind CSS v4, shadcn/ui, `@simplewebauthn/browser` v13, Zustand (auth-store)

---

## ⚠️ Read First

Before writing any code, read the Next.js 16 docs (breaking changes from prior versions):
```bash
ls /Users/kgory/dev/kirmaphor/web/node_modules/next/dist/docs/
```
Also read `web/AGENTS.md` — this Next.js version has breaking API changes.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `web/lib/auth-store.ts` | Modify | Add cookie sync alongside localStorage |
| `web/app/globals.css` | Modify | Add `--auth-*` CSS custom properties |
| `web/components/auth/AuthBranding.tsx` | Create | Left panel: logo, tagline, glow blob |
| `web/components/auth/AuthDivider.tsx` | Create | "or" horizontal divider |
| `web/app/(auth)/layout.tsx` | Create | Split layout wrapper (40/60) |
| `web/components/auth/PasskeyLoginForm.tsx` | Create | Passkey login + email toggle |
| `web/components/auth/EmailLoginForm.tsx` | Create | Email+password login form |
| `web/app/(auth)/login/page.tsx` | Modify | Wire forms into AuthLayout |
| `web/components/auth/PasskeyRegisterForm.tsx` | Create | Passkey register + email toggle |
| `web/components/auth/EmailRegisterForm.tsx` | Create | Email+password register form |
| `web/app/(auth)/register/page.tsx` | Create | Wire forms into AuthLayout |

---

## Task 1: Fix auth-store cookie sync

**Context:** Middleware (`web/middleware.ts`) guards protected routes by reading the `kirmaphore_token` **cookie**. But `auth-store.ts` only persists to `localStorage`. After login, every navigation to a protected route triggers a redirect to `/login`. This must be fixed before building the UI.

**Files:**
- Modify: `web/lib/auth-store.ts`

- [ ] **Step 1: Open the file and understand it**

Read `web/lib/auth-store.ts`. Note that `login` and `setUser` call `localStorage.setItem` but never `document.cookie`.

- [ ] **Step 2: Apply the fix**

Replace the entire file with:

```typescript
// web/lib/auth-store.ts
import { create } from 'zustand'
import { api } from './api'

export interface AuthUser {
  id: string
  email: string
  display_name: string
  avatar_url?: string
}

interface AuthState {
  user: AuthUser | null
  token: string | null
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  setUser: (user: AuthUser, token: string) => void
}

function saveToken(token: string) {
  localStorage.setItem('kirmaphore_token', token)
  document.cookie = `kirmaphore_token=${token}; path=/; SameSite=Lax`
}

function clearToken() {
  localStorage.removeItem('kirmaphore_token')
  document.cookie = 'kirmaphore_token=; path=/; max-age=0'
}

export const useAuth = create<AuthState>((set) => ({
  user: null,
  token: typeof window !== 'undefined' ? localStorage.getItem('kirmaphore_token') : null,

  login: async (email, password) => {
    const res = await api.post<{ token: string; user: AuthUser }>('/api/auth/login', { email, password })
    saveToken(res.token)
    set({ user: res.user, token: res.token })
  },

  logout: () => {
    api.post('/api/auth/logout', {}).catch(() => {})
    clearToken()
    set({ user: null, token: null })
  },

  setUser: (user, token) => {
    saveToken(token)
    set({ user, token })
  },
}))
```

- [ ] **Step 3: Verify TypeScript compiles**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors relating to auth-store.ts

- [ ] **Step 4: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/lib/auth-store.ts
git commit -m "fix: sync auth token to cookie so middleware can read it"
```

---

## Task 2: Auth CSS tokens

**Context:** The spec defines four CSS custom properties for the auth color palette. They live in `globals.css` inside an `:root` block.

**Files:**
- Modify: `web/app/globals.css`

- [ ] **Step 1: Locate the `:root` block**

Open `web/app/globals.css`. Find the existing `:root {` block (it already has `--background`, `--foreground`, etc.).

- [ ] **Step 2: Append auth tokens inside `:root`**

Add these four variables at the end of the `:root` block, before its closing `}`:

```css
  /* Auth UI */
  --auth-bg-deep:   #060A14;
  --auth-bg-form:   #0a0e14;
  --auth-border:    rgba(6,182,212,0.15);
  --auth-glow:      rgba(6,182,212,0.12);
```

- [ ] **Step 3: Verify dev server still starts**

```bash
cd /Users/kgory/dev/kirmaphor/web && npm run dev &
```
Open `http://localhost:3000/login` — should render without errors (still shows old UI). Kill the server (`kill %1`).

- [ ] **Step 4: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/app/globals.css
git commit -m "feat: add auth color tokens to globals.css"
```

---

## Task 3: AuthBranding component

**Context:** Left panel of the split layout. Contains the Kirmaphore wordmark (top-left), tagline (bottom-left), and a radial cyan glow blob (centered, slow pulse animation).

**Files:**
- Create: `web/components/auth/AuthBranding.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/auth/AuthBranding.tsx
export default function AuthBranding() {
  return (
    <div
      className="relative hidden lg:flex flex-col justify-between h-full px-10 py-10 overflow-hidden"
      style={{ backgroundColor: 'var(--auth-bg-deep)' }}
    >
      {/* Glow blob — centered */}
      <div
        className="absolute inset-0 flex items-center justify-center pointer-events-none"
        aria-hidden="true"
      >
        <div
          className="w-[500px] h-[500px] rounded-full animate-glow-pulse"
          style={{
            background: 'radial-gradient(circle, var(--auth-glow) 0%, transparent 70%)',
          }}
        />
      </div>

      {/* Logo — top-left */}
      <div className="relative z-10">
        <span className="text-xl font-bold text-white tracking-tight">Kirmaphore</span>
      </div>

      {/* Tagline — bottom-left */}
      <div className="relative z-10">
        <p className="text-sm text-gray-400">Your infrastructure understands you.</p>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Add the glow-pulse animation to globals.css**

In `web/app/globals.css`, add a keyframes block at the end of the file (outside `:root`):

```css
@keyframes glow-pulse {
  0%, 100% { opacity: 0.6; }
  50%       { opacity: 1.0; }
}

.animate-glow-pulse {
  animation: glow-pulse 3s ease-in-out infinite;
}
```

- [ ] **Step 3: Verify TypeScript**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/auth/AuthBranding.tsx web/app/globals.css
git commit -m "feat: AuthBranding left panel with glow animation"
```

---

## Task 4: AuthDivider component

**Context:** Reusable "or" divider with horizontal lines on each side. Used in both login and register pages.

**Files:**
- Create: `web/components/auth/AuthDivider.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/auth/AuthDivider.tsx
export default function AuthDivider() {
  return (
    <div className="flex items-center gap-3 my-2">
      <div className="flex-1 h-px bg-white/10" />
      <span className="text-xs text-gray-500 select-none">or</span>
      <div className="flex-1 h-px bg-white/10" />
    </div>
  )
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/auth/AuthDivider.tsx
git commit -m "feat: AuthDivider 'or' separator component"
```

---

## Task 5: AuthLayout

**Context:** `app/(auth)/layout.tsx` is the layout shared by `/login` and `/register`. The existing `(auth)` route group already exists (has `login/page.tsx` in it) but has no `layout.tsx` — Next.js falls back to the root layout. We're adding a dedicated layout.

Left: `AuthBranding` (40%, hidden on mobile). Right: form content (60%, full-width on mobile). Form content is centered vertically.

**Files:**
- Create: `web/app/(auth)/layout.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/app/(auth)/layout.tsx
import AuthBranding from '@/components/auth/AuthBranding'

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen flex" style={{ backgroundColor: 'var(--auth-bg-form)' }}>
      {/* Left panel — 40%, hidden on mobile */}
      <div className="lg:w-[40%]">
        <AuthBranding />
      </div>

      {/* Right panel — 60% (full width on mobile) */}
      <div className="flex-1 flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-[400px]">
          {children}
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Verify dev server renders old login inside new layout**

```bash
cd /Users/kgory/dev/kirmaphor/web && npm run dev &
```
Open `http://localhost:3000/login`. You should see the old card form on the right side with a dark background, and no left panel on mobile, or a blank dark left panel on desktop. Kill the server.

- [ ] **Step 3: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/app/(auth)/layout.tsx
git commit -m "feat: AuthLayout split-layout wrapper for auth pages"
```

---

## Task 6: PasskeyLoginForm

**Context:** Passkey-first view for the login page. Shows a single cyan gradient button ("Sign in with Passkey") and a ghost link to toggle to the email form. Handles the full WebAuthn authentication flow using `@simplewebauthn/browser`.

**Files:**
- Create: `web/components/auth/PasskeyLoginForm.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/auth/PasskeyLoginForm.tsx
'use client'

import { useState } from 'react'
import { startAuthentication } from '@simplewebauthn/browser'
import { useAuth } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'
import { AuthUser } from '@/lib/auth-store'

interface Props {
  onSwitchToEmail: () => void
}

export default function PasskeyLoginForm({ onSwitchToEmail }: Props) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { setUser } = useAuth()
  const router = useRouter()

  const handlePasskey = async () => {
    setLoading(true)
    setError('')
    try {
      const options = await api.post<Record<string, unknown>>('/api/auth/passkey/login/begin', {})
      const credential = await startAuthentication({ optionsJSON: options as never })
      const res = await api.post<{ token: string; user: AuthUser }>('/api/auth/passkey/login/finish', credential)
      setUser(res.user, res.token)
      router.push('/')
    } catch {
      setError('Passkey failed — try email instead')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <button
        onClick={handlePasskey}
        disabled={loading}
        className="w-full h-12 rounded-xl font-semibold text-white flex items-center justify-center gap-2 transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
        style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
      >
        {/* Fingerprint icon */}
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M2 12C2 6.5 6.5 2 12 2a10 10 0 0 1 8 4"/>
          <path d="M5 19.5C5.5 18 6 15 6 12c0-1.6.6-3.1 1.8-4.2"/>
          <path d="M17.5 21C17 20 16 19 16 17c0-1.7-.5-3.3-1.5-4.5"/>
          <path d="M22 12a10 10 0 0 1-2 6"/>
          <path d="M9 12c0 1.7.5 3.3 1.5 4.5"/>
          <path d="M12 7a5 5 0 0 1 5 5c0 3.5-1.5 6.5-4 8.5"/>
          <path d="M12 7a5 5 0 0 0-5 5c0 1.5.2 3 .6 4.3"/>
        </svg>
        {loading ? 'Verifying…' : 'Sign in with Passkey'}
      </button>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <button
        type="button"
        onClick={onSwitchToEmail}
        className="w-full text-sm text-gray-400 hover:text-white transition-colors duration-150 text-center"
      >
        Continue with email →
      </button>
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/auth/PasskeyLoginForm.tsx
git commit -m "feat: PasskeyLoginForm with WebAuthn authentication flow"
```

---

## Task 7: EmailLoginForm

**Context:** Email+password form shown after user clicks "Continue with email" on the login page. Slide-in animation, "← Back to passkey" link, inline errors.

**Files:**
- Create: `web/components/auth/EmailLoginForm.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/auth/EmailLoginForm.tsx
'use client'

import { useState } from 'react'
import { useAuth } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'

interface Props {
  onSwitchToPasskey: () => void
}

export default function EmailLoginForm({ onSwitchToPasskey }: Props) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { login } = useAuth()
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      await login(email, password)
      router.push('/')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Sign in failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-4 animate-slide-in"
    >
      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="login-email">Email</label>
        <input
          id="login-email"
          type="email"
          required
          value={email}
          onChange={e => setEmail(e.target.value)}
          placeholder="you@example.com"
          className="w-full h-11 rounded-lg px-3 text-white placeholder-gray-500 outline-none transition-[border-color,box-shadow] duration-150
            border focus:border-cyan-500/50 focus:shadow-[0_0_0_3px_rgba(6,182,212,0.15)]"
          style={{
            backgroundColor: 'rgba(255,255,255,0.05)',
            borderColor: 'rgba(255,255,255,0.08)',
          }}
        />
      </div>

      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="login-password">Password</label>
        <input
          id="login-password"
          type="password"
          required
          value={password}
          onChange={e => setPassword(e.target.value)}
          placeholder="••••••••"
          className="w-full h-11 rounded-lg px-3 text-white placeholder-gray-500 outline-none transition-[border-color,box-shadow] duration-150
            border focus:border-cyan-500/50 focus:shadow-[0_0_0_3px_rgba(6,182,212,0.15)]"
          style={{
            backgroundColor: 'rgba(255,255,255,0.05)',
            borderColor: 'rgba(255,255,255,0.08)',
          }}
        />
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading}
        className="w-full h-12 rounded-xl font-semibold text-white transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
        style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
      >
        {loading ? 'Signing in…' : 'Sign in'}
      </button>

      <button
        type="button"
        onClick={onSwitchToPasskey}
        className="w-full text-sm text-gray-400 hover:text-white transition-colors duration-150 text-center"
      >
        ← Back to passkey
      </button>
    </form>
  )
}
```

- [ ] **Step 2: Add slide-in animation to globals.css**

Append to `web/app/globals.css` (after the glow-pulse keyframes):

```css
@keyframes slide-in {
  from { opacity: 0; transform: translateX(20px); }
  to   { opacity: 1; transform: translateX(0); }
}

.animate-slide-in {
  animation: slide-in 200ms ease-out forwards;
}
```

- [ ] **Step 3: Verify TypeScript**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/auth/EmailLoginForm.tsx web/app/globals.css
git commit -m "feat: EmailLoginForm with slide-in animation"
```

---

## Task 8: Redesign login page

**Context:** Replace the current minimal card with the new passkey-first form. The page manages which sub-form is visible (`passkey` | `email`) and renders the appropriate component. Both views show the "Already have an account?" link at the bottom.

**Files:**
- Modify: `web/app/(auth)/login/page.tsx`

- [ ] **Step 1: Replace the file content**

```tsx
// web/app/(auth)/login/page.tsx
'use client'

import { useState } from 'react'
import Link from 'next/link'
import PasskeyLoginForm from '@/components/auth/PasskeyLoginForm'
import EmailLoginForm from '@/components/auth/EmailLoginForm'
import AuthDivider from '@/components/auth/AuthDivider'

type View = 'passkey' | 'email'

export default function LoginPage() {
  const [view, setView] = useState<View>('passkey')

  return (
    <div className="space-y-6">
      {/* Heading */}
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold text-white">Welcome back</h1>
        <p className="text-sm text-gray-400">Sign in to your account</p>
      </div>

      {view === 'passkey' ? (
        <>
          <PasskeyLoginForm onSwitchToEmail={() => setView('email')} />
          <AuthDivider />
        </>
      ) : (
        <EmailLoginForm onSwitchToPasskey={() => setView('passkey')} />
      )}

      {/* Bottom link */}
      <p className="text-sm text-gray-400 text-center">
        Don&apos;t have an account?{' '}
        <Link href="/register" className="text-cyan-400 hover:text-cyan-300 transition-colors">
          Create one →
        </Link>
      </p>
    </div>
  )
}
```

- [ ] **Step 2: Verify visually**

```bash
cd /Users/kgory/dev/kirmaphor/web && npm run dev &
```
- Open `http://localhost:3000/login`
- Expect: dark navy background, split layout (desktop), "Welcome back" heading, cyan passkey button
- Click "Continue with email →": email form slides in
- Click "← Back to passkey": passkey view returns
- Kill the server.

- [ ] **Step 3: Verify TypeScript**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/app/(auth)/login/page.tsx
git commit -m "feat: redesign login page with passkey-first UI"
```

---

## Task 9: PasskeyRegisterForm

**Context:** Passkey-first view for the register page. Calls the WebAuthn registration flow. On success, calls `setUser` and redirects to `/`.

**Files:**
- Create: `web/components/auth/PasskeyRegisterForm.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/auth/PasskeyRegisterForm.tsx
'use client'

import { useState } from 'react'
import { startRegistration } from '@simplewebauthn/browser'
import { useAuth, AuthUser } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'

interface Props {
  onSwitchToEmail: () => void
}

export default function PasskeyRegisterForm({ onSwitchToEmail }: Props) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const { setUser } = useAuth()
  const router = useRouter()

  const handlePasskey = async () => {
    setLoading(true)
    setError('')
    try {
      const options = await api.post<Record<string, unknown>>('/api/auth/passkey/register/begin', {})
      const credential = await startRegistration({ optionsJSON: options as never })
      const res = await api.post<{ token: string; user: AuthUser }>('/api/auth/passkey/register/finish', credential)
      setUser(res.user, res.token)
      router.push('/')
    } catch {
      setError('Passkey failed — try email instead')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <button
        onClick={handlePasskey}
        disabled={loading}
        className="w-full h-12 rounded-xl font-semibold text-white flex items-center justify-center gap-2 transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
        style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
      >
        {/* Fingerprint icon */}
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
          <path d="M2 12C2 6.5 6.5 2 12 2a10 10 0 0 1 8 4"/>
          <path d="M5 19.5C5.5 18 6 15 6 12c0-1.6.6-3.1 1.8-4.2"/>
          <path d="M17.5 21C17 20 16 19 16 17c0-1.7-.5-3.3-1.5-4.5"/>
          <path d="M22 12a10 10 0 0 1-2 6"/>
          <path d="M9 12c0 1.7.5 3.3 1.5 4.5"/>
          <path d="M12 7a5 5 0 0 1 5 5c0 3.5-1.5 6.5-4 8.5"/>
          <path d="M12 7a5 5 0 0 0-5 5c0 1.5.2 3 .6 4.3"/>
        </svg>
        {loading ? 'Setting up…' : 'Create account with Passkey'}
      </button>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <button
        type="button"
        onClick={onSwitchToEmail}
        className="w-full text-sm text-gray-400 hover:text-white transition-colors duration-150 text-center"
      >
        Continue with email →
      </button>
    </div>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/auth/PasskeyRegisterForm.tsx
git commit -m "feat: PasskeyRegisterForm with WebAuthn registration flow"
```

---

## Task 10: EmailRegisterForm

**Context:** Email view for the register page. Three fields: Display name, Email, Password. Calls `POST /api/auth/register`. On success calls `setUser` and redirects.

**Files:**
- Create: `web/components/auth/EmailRegisterForm.tsx`

- [ ] **Step 1: Create the file**

```tsx
// web/components/auth/EmailRegisterForm.tsx
'use client'

import { useState } from 'react'
import { useAuth, AuthUser } from '@/lib/auth-store'
import { useRouter } from 'next/navigation'
import { api } from '@/lib/api'

interface Props {
  onSwitchToPasskey: () => void
}

const inputClass = `w-full h-11 rounded-lg px-3 text-white placeholder-gray-500 outline-none
  transition-[border-color,box-shadow] duration-150 border
  focus:border-cyan-500/50 focus:shadow-[0_0_0_3px_rgba(6,182,212,0.15)]`

const inputStyle = {
  backgroundColor: 'rgba(255,255,255,0.05)',
  borderColor: 'rgba(255,255,255,0.08)',
}

export default function EmailRegisterForm({ onSwitchToPasskey }: Props) {
  const [displayName, setDisplayName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const { setUser } = useAuth()
  const router = useRouter()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    try {
      const res = await api.post<{ token: string; user: AuthUser }>('/api/auth/register', {
        display_name: displayName,
        email,
        password,
      })
      setUser(res.user, res.token)
      router.push('/')
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 animate-slide-in">
      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="reg-name">Display name</label>
        <input
          id="reg-name"
          type="text"
          required
          value={displayName}
          onChange={e => setDisplayName(e.target.value)}
          placeholder="Ada Lovelace"
          className={inputClass}
          style={inputStyle}
        />
      </div>

      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="reg-email">Email</label>
        <input
          id="reg-email"
          type="email"
          required
          value={email}
          onChange={e => setEmail(e.target.value)}
          placeholder="you@example.com"
          className={inputClass}
          style={inputStyle}
        />
      </div>

      <div className="space-y-1">
        <label className="text-sm text-gray-300" htmlFor="reg-password">Password</label>
        <input
          id="reg-password"
          type="password"
          required
          value={password}
          onChange={e => setPassword(e.target.value)}
          placeholder="••••••••"
          className={inputClass}
          style={inputStyle}
        />
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <button
        type="submit"
        disabled={loading}
        className="w-full h-12 rounded-xl font-semibold text-white transition-[filter] duration-150 hover:brightness-110 disabled:opacity-60"
        style={{ background: 'linear-gradient(135deg, #06b6d4, #0e7490)' }}
      >
        {loading ? 'Creating account…' : 'Create account'}
      </button>

      <button
        type="button"
        onClick={onSwitchToPasskey}
        className="w-full text-sm text-gray-400 hover:text-white transition-colors duration-150 text-center"
      >
        ← Back to passkey
      </button>
    </form>
  )
}
```

- [ ] **Step 2: Verify TypeScript**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/components/auth/EmailRegisterForm.tsx
git commit -m "feat: EmailRegisterForm with display_name, email, password fields"
```

---

## Task 11: Register page

**Context:** New page at `/register`. Same structure as login page: manages `passkey` | `email` view state, renders the appropriate form, shows "Already have an account? Sign in →" at the bottom.

**Files:**
- Create: `web/app/(auth)/register/page.tsx`

- [ ] **Step 1: Create the directory and file**

```bash
mkdir -p /Users/kgory/dev/kirmaphor/web/app/\(auth\)/register
```

```tsx
// web/app/(auth)/register/page.tsx
'use client'

import { useState } from 'react'
import Link from 'next/link'
import PasskeyRegisterForm from '@/components/auth/PasskeyRegisterForm'
import EmailRegisterForm from '@/components/auth/EmailRegisterForm'
import AuthDivider from '@/components/auth/AuthDivider'

type View = 'passkey' | 'email'

export default function RegisterPage() {
  const [view, setView] = useState<View>('passkey')

  return (
    <div className="space-y-6">
      {/* Heading */}
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold text-white">Create your account</h1>
        <p className="text-sm text-gray-400">Start automating infrastructure today.</p>
      </div>

      {view === 'passkey' ? (
        <>
          <PasskeyRegisterForm onSwitchToEmail={() => setView('email')} />
          <AuthDivider />
        </>
      ) : (
        <EmailRegisterForm onSwitchToPasskey={() => setView('passkey')} />
      )}

      {/* Bottom link */}
      <p className="text-sm text-gray-400 text-center">
        Already have an account?{' '}
        <Link href="/login" className="text-cyan-400 hover:text-cyan-300 transition-colors">
          Sign in →
        </Link>
      </p>
    </div>
  )
}
```

- [ ] **Step 2: Full visual verification**

```bash
cd /Users/kgory/dev/kirmaphor/web && npm run dev &
```

Check all states:
- `http://localhost:3000/login` — "Welcome back", passkey button, email toggle works
- `http://localhost:3000/register` — "Create your account", passkey button, email toggle shows 3 fields
- Resize to mobile (< 1024px) — left panel hidden, form full-width

Kill the server.

- [ ] **Step 3: Verify TypeScript**

```bash
cd /Users/kgory/dev/kirmaphor/web && npx tsc --noEmit
```
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add web/app/\(auth\)/register/
git commit -m "feat: register page with passkey-first UI"
```

---

## Task 12: Production build verification

**Context:** Confirm the Next.js production build succeeds before deployment. The Dockerfile runs `npm run build` — any TS or module error would break the image.

- [ ] **Step 1: Run production build**

```bash
cd /Users/kgory/dev/kirmaphor/web && npm run build
```
Expected: `✓ Compiled successfully` with no errors. The `/login` and `/register` routes should appear in the route table.

- [ ] **Step 2: If build fails, fix errors**

Common issues:
- Missing `'use client'` directive on a component that uses hooks
- Import path typo (all auth components live in `@/components/auth/`)
- `@simplewebauthn/browser` types mismatch — if so, cast options as `never` (already done in plan)

- [ ] **Step 3: Final commit**

```bash
cd /Users/kgory/dev/kirmaphor && git add -p
git commit -m "fix: resolve any build errors in auth UI"
```
(Only needed if step 2 required fixes.)

---

## Self-Review Checklist

Spec sections vs. tasks:

| Spec section | Task |
|---|---|
| § 2 Color tokens | Task 2 |
| § 3 AuthLayout (split, mobile) | Task 5 |
| § 4 AuthBranding (logo, tagline, glow) | Task 3 |
| § 5 Register passkey flow | Task 9 |
| § 5 Register email flow | Task 10 |
| § 5 Register page wiring | Task 11 |
| § 6 Login passkey flow | Task 6 |
| § 6 Login email flow | Task 7 |
| § 6 Login page wiring | Task 8 |
| § 7 AuthDivider | Task 4 |
| § 8 Animations (glow, slide-in, hover, focus) | Tasks 3, 7, 6, 7 |
| § 9 Error states (inline, red-400, no toasts) | Tasks 6–10 |
| Middleware cookie bug | Task 1 |
