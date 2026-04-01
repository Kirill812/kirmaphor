# Auth UI — Registration & Login Design Spec
**Date:** 2026-04-01  
**Status:** Approved  
**Author:** kgory + Claude

---

## 1. Overview

Replace the current minimal login card with a full-featured, premium auth experience matching the requesty.ai aesthetic: deep navy backgrounds, cyan accents, split layout, Geist typography.

Scope: `/register` (new) + `/login` (restyled). Both share one `AuthLayout`.

---

## 2. Visual Design

### Color Tokens (extend globals.css)

```css
--auth-bg-deep:   #060A14;   /* left panel */
--auth-bg-form:   #0a0e14;   /* right panel */
--auth-border:    rgba(6,182,212,0.15);  /* cyan subtle border */
--auth-glow:      rgba(6,182,212,0.12);  /* glow blob */
```

### Typography
- Font: Geist Sans (already loaded)
- Heading: `text-2xl font-semibold text-white`
- Subtext: `text-sm text-gray-400`

### Primary Button (Passkey)
```
background: linear-gradient(135deg, #06b6d4, #0e7490)
border-radius: 0.75rem (rounded-xl)
height: 48px
font-weight: 600
color: white
icon: fingerprint SVG (left of text)
```

### Inputs
```
background: rgba(255,255,255,0.05)
border: 1px solid rgba(255,255,255,0.08)
color: white
placeholder: gray-500
border-radius: 0.5rem
height: 44px
focus-ring: --auth-border (cyan)
```

---

## 3. Layout — `AuthLayout`

```
┌──────────────────────────────────────────────────────┐
│  LEFT 40%                │  RIGHT 60%                │
│  bg: #060A14             │  bg: #0a0e14              │
│                          │                           │
│  [Logo + wordmark]       │  [Form, max-w-400px]      │
│                          │                           │
│  "Your infrastructure    │                           │
│   understands you."      │                           │
│                          │                           │
│  [Animated cyan glow]    │                           │
└──────────────────────────────────────────────────────┘
```

**Mobile** (< lg): left panel hidden, right panel full-screen. Compact logo shown above form.

**Files:**
- `app/(auth)/layout.tsx` — AuthLayout wrapping both pages
- `components/auth/AuthBranding.tsx` — left panel (logo + tagline + glow animation)

---

## 4. Left Panel — `AuthBranding`

- Logo: SVG wordmark "Kirmaphore" in white, top-left, `text-xl font-bold`
- Tagline: `"Your infrastructure understands you."` — `text-gray-400 text-sm`, bottom-left area
- Decoration: one radial glow blob, centered, `rgba(6,182,212,0.12)` → transparent, 500px diameter, slow pulse animation (3s ease-in-out infinite, opacity 0.6 → 1.0)
- No other content

---

## 5. Register Page — `/register`

**Default view** (passkey-first):

```
"Create your account"          ← text-2xl font-semibold text-white
"Start automating infrastructure today."  ← text-sm text-gray-400

[🔑 Create account with Passkey]   ← primary cyan gradient button

────────── or ──────────

[Continue with email →]            ← ghost link, text-gray-400, hover text-white
```

**Email view** (shown after clicking "Continue with email"):

- Slide-in from right (CSS transition: `translateX(20px) → 0, opacity 0 → 1`, 200ms)
- Fields: Display name, Email, Password (all with label above)
- Submit button: same cyan gradient, text "Create account"
- Below: `← Back to passkey` link
- No passkey button visible in this view

**Bottom of both views:**
```
Already have an account?  Sign in →   ← link to /login
```

**Passkey flow:**
1. Click button → call `POST /api/auth/passkey/register/begin` → get challenge
2. Browser WebAuthn prompt (`navigator.credentials.create(...)`)
3. POST result to `/api/auth/passkey/register/finish`
4. On success → store token in localStorage → redirect to `/`
5. On error → inline red message below button

**Email flow:**
1. Fill form → `POST /api/auth/register` with `{display_name, email, password}`
2. On success → store token → redirect to `/`
3. On error → inline message below submit button

---

## 6. Login Page — `/login` (restyled)

Same `AuthLayout`. Right panel form:

```
"Welcome back"                ← text-2xl font-semibold
"Sign in to your account"     ← text-sm text-gray-400

[🔑 Sign in with Passkey]     ← primary cyan gradient button

────────── or ──────────

[Continue with email →]       ← ghost link
```

Email view: Email + Password fields, "Sign in" button.

Bottom: `Don't have an account? Create one →` link to `/register`.

---

## 7. Shared Components

| Component | Purpose |
|---|---|
| `AuthLayout` | Split layout, mobile collapse |
| `AuthBranding` | Left panel with glow |
| `PasskeyRegisterForm` | Passkey register flow |
| `PasskeyLoginForm` | Passkey login flow |
| `EmailRegisterForm` | Email+password register |
| `EmailLoginForm` | Email+password login |
| `AuthDivider` | "or" divider line |

Each form component is self-contained: handles its own API calls, loading state, error display.

---

## 8. Animations

| Element | Animation |
|---|---|
| Glow blob | `opacity: 0.6 → 1.0`, 3s ease-in-out infinite |
| Email form slide-in | `translateX(20px)→0 + opacity 0→1`, 200ms ease-out |
| Passkey button hover | `brightness(1.1)`, 150ms |
| Input focus | cyan border `rgba(6,182,212,0.5)`, box-shadow glow |

---

## 9. Error States

- Inline, below the triggering element
- `text-red-400 text-sm`
- No toasts, no modals
- Passkey errors: browser already shows native error, show fallback message "Passkey failed — try email instead"

---

## 10. Out of Scope

- TOTP / MFA
- OAuth (Google, GitHub) — future Enterprise feature
- Email verification flow
- Password strength meter
