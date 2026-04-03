# Landing Page Design Spec
**Date:** 2026-04-03
**Status:** Approved
**Author:** kgory + Claude

---

## 1. Overview

A premium, 2026-style public landing page at `/` for Kirmaphore — an AI-native Ansible automation platform. Goal: make AWX/Semaphore look dated in comparison. Design references: Linear, Vercel, requesty.ai.

**Target audiences (two layers):**
- Startup DevOps engineer (5–50 people, Ansible already in use, AWX felt like overkill)
- Scaleup SRE (100–500 people, suffering from Ansible Tower/AWX complexity or cost)

---

## 2. Route Architecture Change

The current `/` route is protected by middleware and served by `app/(dashboard)/page.tsx`. Landing page requires:

- **Add** `app/(marketing)/page.tsx` — public landing at `/`
- **Move** dashboard home to `/dashboard` — rename `app/(dashboard)/page.tsx` to `app/(dashboard)/dashboard/page.tsx`
- **Update** `web/middleware.ts`: add `'/'` to `PUBLIC_PATHS`
- **Update** post-login redirect in auth forms: change `router.push('/')` → `router.push('/dashboard')`

---

## 3. Visual Design System

### Colors
```css
--land-bg:       #060A14;   /* page background */
--land-surface:  rgba(255,255,255,0.03);  /* card background */
--land-border:   rgba(255,255,255,0.07);  /* card border */
--land-cyan:     #06b6d4;   /* primary accent */
--land-violet:   #818cf8;   /* AI accent */
--land-text:     rgba(255,255,255,0.88);  /* primary text */
--land-muted:    rgba(255,255,255,0.38);  /* secondary text */
--land-dim:      rgba(255,255,255,0.18);  /* tertiary / dividers */
```

### Typography
- Headings: `font-weight: 700`, `letter-spacing: -0.03em` to `-0.04em`
- Hero headline: `clamp(48px, 7vw, 88px)`, `font-weight: 800`, `letter-spacing: -0.04em`
- Section titles: `clamp(28px, 3.5vw, 44px)`, `font-weight: 700`
- Section labels: `11px`, `letter-spacing: 0.12em`, uppercase, `color: --land-cyan * 0.7`
- Body: `15px`, `line-height: 1.6`
- Monospace: `'SF Mono', monospace` for code/logs

### Icons
SVG line icons. Rules:
- `stroke-width: 2.75` — matches `font-weight: 600` of card titles visually
- `stroke-linecap: round`, `stroke-linejoin: round`
- `fill: none`
- Simple forms — max 2-3 paths per icon, no decorative detail
- Size: `18px` rendered inside `36×36px` container (`border-radius: 9px`)
- Container: `background: rgba(255,255,255,0.06)`, `border: 1px solid rgba(255,255,255,0.10)`
- Forge AI icon only: container uses violet tint (`rgba(129,140,248,0.12)`, border `rgba(129,140,248,0.25)`), stroke `rgba(129,140,248,0.95)`

### Buttons
```
Primary:   gradient(135deg, #06b6d4, #0e7490), border-radius: 10px, h: 44px, font-weight: 600
           box-shadow: 0 0 32px rgba(6,182,212,0.25)
           hover: brightness(1.08)

Secondary: rgba(255,255,255,0.06), border: 1px solid rgba(255,255,255,0.10)
           border-radius: 10px, h: 44px, font-weight: 500
           hover: background rgba(255,255,255,0.09)
```

### Glow blobs
Radial gradients, `pointer-events: none`, `position: absolute`, `animation: pulse 4s ease-in-out infinite` (opacity 0.7→1.0).

---

## 4. NavBar

Sticky top, `height: 56px`, `backdrop-filter: blur(12px)`, `background: rgba(6,10,20,0.8)`, `border-bottom: 1px solid rgba(255,255,255,0.06)`.

**Left:** "Kirmaphore" wordmark, `font-weight: 700`, `font-size: 15px`
**Center:** Links — Features, Pricing, Docs, GitHub (`font-size: 13px`, `color: --land-muted`, hover white)
**Right:** "Sign in" ghost button + "Start free →" primary button (small, `h: 32px`, `font-size: 13px`)

Mobile (`< lg`): center links hidden, hamburger menu (out of scope for now — just hide links).

**File:** `web/components/landing/LandingNav.tsx`

---

## 5. Section 1 — Hero

**File:** `web/components/landing/HeroSection.tsx`

### Layout
Full viewport height (`min-height: 100vh`), flex column, centered. Two glow blobs: center (cyan, 700px) and left (violet, 400px).

### Content (top to bottom, centered)

**Badge:**
```
[● dot] Now in open beta · AI-native · Self-hostable
```
`background: rgba(6,182,212,0.10)`, `border: 1px solid rgba(6,182,212,0.25)`, `border-radius: 100px`, cyan text, animated dot.

**Headline:**
```
Ansible automation
that thinks with you.
```
`clamp(48px, 7vw, 88px)`, `font-weight: 800`. "thinks with you." has gradient text: `linear-gradient(135deg, #06b6d4, #818cf8)`.

**Sub:**
```
Run playbooks, review them with AI, stream live logs —
without the complexity of AWX or the price of Ansible Tower.
```
`17px`, `color: --land-muted`, `max-width: 520px`.

**CTAs:**
```
[⚡ Start free]    [GitHub-icon  View on GitHub]
```
Primary + secondary, side by side, `gap: 12px`.

**Live product preview** (`max-width: 900px`):
- Browser chrome bar (3 dots + fake URL `kirmaphore.io/projects/production`)
- 3-column layout: sidebar (200px) | log stream (flex-1) | Forge AI panel (280px)
- **Sidebar:** project list (production = active+running, staging, dev-cluster), templates list, scheduled jobs
- **Log stream:** `font-family: monospace`, live task output with animated cursor on running task, Forge AI mention in log
- **Forge AI panel:** header with violet icon + "Forge AI" label, "Pattern detected" card (problem explanation), "Suggested fix" card (YAML snippet)
- Subtle glow: `box-shadow: 0 40px 120px rgba(0,0,0,0.6), 0 0 60px rgba(6,182,212,0.06)`

Caption below preview: `"Live product · no illustrations"`, `font-size: 12px`, `--land-dim`.

---

## 6. Section 2 — Pain ("Why teams switch")

**File:** `web/components/landing/PainSection.tsx`

Section label: "Why teams switch"
Title: "Tired of the alternatives?"
Sub: "Every existing tool makes you choose between power and simplicity. Kirmaphore doesn't."

**4-column grid** of pain cards:

| Competitor | Headline | Body |
|---|---|---|
| AWX / TOWER | Requires Kubernetes to run a cron job | Full-stack deployment just to schedule an Ansible playbook. Overkill for 90% of teams. |
| ANSIBLE TOWER | $13,000/year before you write a line | Enterprise pricing locks out every team under 500 engineers. |
| SEMAPHORE UI | 2019 UI, zero AI, zero insight | Runs your playbooks. That's it. When a task fails at 2am you're on your own. |
| RUNDECK | Acquired, rebranded, forgotten | Absorbed into PagerDuty. No roadmap, support routes to a webinar. |

Cards: `background: --land-surface`, `border: 1px solid --land-border`, `border-radius: 12px`, `padding: 20px`. Competitor label: `9px`, `letter-spacing: 0.08em`, `--land-dim`. Headline: `14px`, `font-weight: 600`. Body: `12px`, `--land-muted`.

---

## 7. Section 3 — Features Bento Grid

**File:** `web/components/landing/FeaturesBento.tsx`

Asymmetric 12-column CSS grid, `grid-auto-rows: 160px`, `gap: 12px`.

| Card | Span | Rows | Content |
|---|---|---|---|
| Forge AI | 7 cols | 2 rows | Icon (star/sparkle, violet), title, desc, AI demo terminal (pre-run review + generation example) |
| Live logs | 5 cols | 2 rows | Icon (pulse/waveform), title, desc, animated log mini-stream |
| Git webhooks | 3 cols | 1 row | Icon (git branch), title, desc — "Push to branch → playbook runs automatically" |
| Enterprise RBAC | 4 cols | 1 row | Icon (shield), title, desc |
| Visual DAG | 5 cols | 1 row | Icon (network nodes), title, desc, mini DAG node row |
| Secrets vault | 4 cols | 1 row | Icon (lock), title, desc |
| Cron scheduler | 4 cols | 1 row | Icon (clock — circle + hands), title, desc |
| One-command deploy | 4 cols | 1 row | Icon (box/cube), title, `docker compose up` in code |

Forge AI card: `background: linear-gradient(135deg, rgba(129,140,248,0.07), rgba(6,182,212,0.04))`, `border-color: rgba(129,140,248,0.18)`.

Icon spec: SVG `18px` inside `36×36px` container, `stroke-width: 2.75`, `stroke-linecap: round`, `stroke-linejoin: round`, `fill: none`.

---

## 8. Section 4 — How It Works

**File:** `web/components/landing/HowItWorksSection.tsx`

Section label: "How it works"
Title: "From zero to automated in three steps."

3-column horizontal layout. Connector line between step circles: `height: 1px`, `background: linear-gradient(90deg, transparent, rgba(6,182,212,0.25), transparent)`.

| Step | Title | Body | Code snippet |
|---|---|---|---|
| 1 | Connect your inventory | Add servers, point to your Git repo. SSH keys stored encrypted. | `host: prod.example.com`, `user: deploy`, `key: ••••` |
| 2 | Run — AI reviews first | Forge AI scans your playbook before a single task executes. | `✦ review passed`, `▶ running deploy.yml` |
| 3 | Watch it live | Logs stream in real-time. If it fails, AI explains why and suggests the fix. | `✓ deployed`, `✓ healthy`, `total: 14.2s` |

Step circles: `54×54px`, `border-radius: 50%`, `background: rgba(255,255,255,0.04)`, `border: 1px solid rgba(255,255,255,0.12)`, number in `--land-muted`.

---

## 9. Section 5 — Forge AI Deep Dive

**File:** `web/components/landing/ForgeAISection.tsx`

Section label: "Forge AI"
Title: "Your playbooks, reviewed before they run."
Sub: "Three capabilities, built into every project. No setup, no plugin, no extra cost."

3-column card row, each card highlighting one AI capability:

**Card 1 — Pre-run review**
Icon: magnifying glass. "Catches security issues, deprecated modules, and logic errors before execution."
Example: `no_log` missing on credential task → flagged with explanation.

**Card 2 — Failure explanation**
Icon: lightning/bolt. "When a task fails, AI reads the error, the task context, and your inventory — and tells you exactly what went wrong."
Example: health check timeout → root cause + fix in plain English.

**Card 3 — Natural language generation**
Icon: chat bubble / text cursor. "Describe what you want in plain English. Forge AI writes the Ansible task."
Example: `"add nginx rate limiting to port 80"` → generates task YAML in 1.2s.

Below cards: large interactive terminal mockup showing all three flows in sequence, with tab switcher (Review / Explain / Generate).

---

## 10. Section 6 — Comparison Table

**File:** `web/components/landing/ComparisonSection.tsx`

Section label: "Compare"
Title: "See how we stack up."

Table with Kirmaphore column highlighted (cyan left border):

| Feature | Kirmaphore | Semaphore UI | AWX | Ansible Tower |
|---|---|---|---|---|
| AI playbook review | ✓ | ✗ | ✗ | ✗ |
| AI failure explanation | ✓ | ✗ | ✗ | ✗ |
| AI task generation | ✓ | ✗ | ✗ | ✗ |
| Live WebSocket logs | ✓ | ✗ | ✓ | ✓ |
| Passkey / WebAuthn | ✓ | ✗ | ✗ | ✗ |
| Visual DAG designer | ✓ | ✗ | ✗ | ✓ |
| Self-hostable | ✓ | ✓ | ✓ | ✗ |
| Free tier | ✓ | ✓ | ✓ | ✗ |
| Docker Compose deploy | ✓ | ✓ | ✗ | ✗ |
| Price (team of 10) | Free | Free | Free | ~$13K/yr |

✓ = `#4ade80`, ✗ = `rgba(255,255,255,0.2)`. Kirmaphore column header has subtle cyan background.

---

## 11. Section 7 — Open-Core / Pricing

**File:** `web/components/landing/PricingSection.tsx`

Section label: "Open source"
Title: "Self-host for free. Scale with cloud."

Two cards side by side:

**Community (left)**
- `border: 1px solid rgba(255,255,255,0.1)`
- "Free forever · MIT license"
- Features: All core features, unlimited projects, Docker Compose deploy, community support
- CTA: "Deploy yourself →" → GitHub repo

**Pro (right, highlighted)**
- `border: 1px solid rgba(6,182,212,0.3)`, subtle cyan glow
- "Cloud-hosted · coming soon"
- Features: Everything in Community + Managed hosting, Forge AI (unlimited), priority support, SSO
- CTA: "Join waitlist →" → email capture (simple inline input)

GitHub stars badge between or below cards: `⭐ Star on GitHub` with live count placeholder.

---

## 12. Section 8 — Footer CTA

**File:** `web/components/landing/FooterCTASection.tsx`

Full-width dark block, large centered content:

```
[glow blob, centered, large]

  Automate everything.
  Understand everything.

  Your infrastructure, finally thinking with you.

  [Start free →]   [View on GitHub]
```

Headline: `clamp(36px, 5vw, 64px)`, `font-weight: 800`.
Below CTAs: `"No credit card. No Kubernetes. Just docker compose up."` — `13px`, `--land-dim`.

---

## 13. Footer

Minimal dark footer:
- Left: "Kirmaphore · © 2026"
- Center: Docs · GitHub · Twitter
- Right: "Built with ♥ for infrastructure engineers"

All `12px`, `--land-dim`.

---

## 14. File Structure

```
web/
  app/
    (marketing)/
      page.tsx                         ← landing page root
    (dashboard)/
      dashboard/
        page.tsx                       ← moved from (dashboard)/page.tsx
  components/
    landing/
      LandingNav.tsx
      HeroSection.tsx
      PainSection.tsx
      FeaturesBento.tsx
      HowItWorksSection.tsx
      ForgeAISection.tsx
      ComparisonSection.tsx
      PricingSection.tsx
      FooterCTASection.tsx
      LandingFooter.tsx
  middleware.ts                        ← add '/' to PUBLIC_PATHS
```

Auth forms update: `router.push('/')` → `router.push('/dashboard')` in:
- `web/components/auth/PasskeyLoginForm.tsx`
- `web/components/auth/EmailLoginForm.tsx`
- `web/components/auth/PasskeyRegisterForm.tsx`
- `web/components/auth/EmailRegisterForm.tsx`

---

## 15. Animations

| Element | Animation |
|---|---|
| Glow blobs | `opacity: 0.7→1.0`, 4s ease-in-out infinite |
| Badge dot | `opacity: 0.6→1.0`, 2s infinite |
| Log cursor | `opacity: blink`, 1s step-end |
| Card hover | `border-color`, `background`, 200ms ease |
| Button hover | `brightness(1.08)`, 150ms |
| Scroll reveals | `opacity: 0→1 + translateY(16px→0)`, `animation-timeline: scroll()` or Intersection Observer, 400ms ease-out |

No heavy JS animation libraries. CSS-first, fallback `prefers-reduced-motion: reduce`.

---

## 16. Out of Scope

- Dark/light mode toggle (dark only)
- Blog / changelog section
- Testimonials / social proof logos (no customers yet)
- Pricing calculator
- Mobile hamburger menu (links hidden on mobile, added in future)
- Actual GitHub stars API call (placeholder)
- Email capture backend for Pro waitlist
