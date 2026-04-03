# Landing Page & Onboarding UI/UX Design Trends — 2026
*Research compiled April 2026 from: Awwwards SOTD winners (Mar–Apr 2026), land-book.com, lapa.ninja, saaslandingpage.com, NN/Group.*

---

## HERO SECTION TRENDS

### 1. Procedural WebGL / Generative Shaders as Hero
**What it looks like:** Full-viewport hero with a real-time WebGL canvas. Not a video loop — it responds to mouse position, scroll depth, or audio. The "content" IS the animation: cymatic wave patterns, Chladni figures, particle fields rendered live in GLSL.

**Visual:** Dark near-black background (#0d0d0d–#1a1a1a), a single high-contrast accent (electric orange, acid green, or cool white), monospace or extended-sans headline floating over the canvas. No stock imagery anywhere.

**Examples:**
- `iyo.ai` — Awwwards SOTD Apr 2, 2026. Chladni pattern WebGL intro, procedural shader footer, dynamic color-switch product slider. Stack: Three.js, WebGL, Webflow.
- `artefakt.mov` — Awwwards SOTD Apr 3, 2026. ASCII post-processing layered over a particle system, mouse interaction drives the render. Stack: WebGL, Three.js.
- `vastspace.com` — Awwwards SOTD Apr 1, 2026. Interactive WebGL interior/exterior fly-through of a space station. Stack: Three.js, Webflow. Color palette: `#2A2C2F` (near-black charcoal) + `#FF5623` (burnt orange).

**CSS/code clues:**
```glsl
/* GLSL fragment shader pattern — procedural noise hero */
uniform float uTime;
uniform vec2 uMouse;
varying vec2 vUv;

void main() {
  vec2 st = vUv - 0.5;
  float d = length(st + uMouse * 0.1);
  float wave = sin(d * 20.0 - uTime * 3.0) * 0.5 + 0.5;
  gl_FragColor = vec4(vec3(wave), 1.0);
}
```
```js
// Three.js setup skeleton
const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
const uniforms = { uTime: { value: 0 }, uMouse: { value: new THREE.Vector2() } };
```

---

### 2. Cinematic Scroll Storytelling (GSAP ScrollTrigger + mask wipes)
**What it looks like:** The page doesn't scroll — it *reveals*. Content panels wipe in using clip-path masks triggered by scroll position. Horizontal wipes, radial reveals, text that assembles letter by letter as you scroll. The nav shrinks/morphs on scroll. Feels like a film cut.

**Examples:**
- `fluid.glass` — Awwwards SOTD Mar 30, 2026. "Mask wipe transition on scroll," a process component that plays through like a timeline, scroll-driven showroom. Elements: GSAP ScrollTrigger, clip-path animation, intro loading animation.
- `vastspace.com` — Scroll moves you through the space station 3D scene.

**CSS/code clues:**
```css
/* Clip-path wipe reveal */
.panel {
  clip-path: inset(0 100% 0 0);
  transition: clip-path 0.8s cubic-bezier(0.77, 0, 0.18, 1);
}
.panel.revealed {
  clip-path: inset(0 0% 0 0);
}
```
```js
// GSAP ScrollTrigger
gsap.to(".panel", {
  clipPath: "inset(0 0% 0 0)",
  scrollTrigger: { trigger: ".section", scrub: 1, start: "top center" }
});
```

---

### 3. Minimal Two-Tone / Brutalist Constraint
**What it looks like:** Exactly two colors — one deep neutral (charcoal, off-black, or warm dark grey) + one saturated accent. No gradients. No shadows. Thick visible borders. Helvetica or a grotesque extended-weight font at 100–180px. Grid lines as decoration. Raw, confident, anti-decorative.

**Visual reference:** `vastspace.com` — `#2A2C2F` + `#FF5623`. Nothing else. No photos on the hero, just type + WebGL.

**CSS/code clues:**
```css
:root {
  --bg: #2A2C2F;
  --accent: #FF5623;
  --text: #F5F0EB;
}
body { background: var(--bg); color: var(--text); }
h1 { font-size: clamp(64px, 10vw, 160px); font-weight: 900; letter-spacing: -0.04em; }
.border-accent { border: 2px solid var(--accent); }
```

---

## COLOR PALETTE TRENDS

### 4. Near-Black + Single Chromatic Accent (the "Charcoal + Pop" palette)
**Dominant across AI/tech SaaS in 2026.** The background is not pure #000000 — it's a warm or cool near-black with slight chroma (`#0E0F11`, `#111318`, `#1A1A1F`). One accent: typically electric blue, acid green, burnt orange, or signal red. No multi-color gradients.

**Real examples:**
- `vastspace.com`: `#2A2C2F` + `#FF5623`
- `sierra.ai`: Deep navy + white (featured saaslandingpage.com March 2026)
- `gradient-labs.ai`: Dark + gradient accent (featured saaslandingpage.com March 2026)
- `plasticity.xyz`: Dark workspace tool aesthetic (featured April 2026)

**CSS/code clues (oklch for perceptual uniformity):**
```css
:root {
  --bg: oklch(0.145 0 0);         /* near-black */
  --accent: oklch(0.65 0.22 145); /* acid green */
  --text: oklch(0.985 0 0);
}
```

### 5. "Gradient Labs" Aurora / Mesh Gradient
**What it looks like:** Soft, blurry blob gradients behind dark glass panels. Multiple hues (purple, teal, pink) bleed into each other like northern lights. NOT the CSS `linear-gradient` — these are radial blobs with heavy `filter: blur()` or SVG `feTurbulence`. The rest of the page is stark and minimal; the aurora is only behind the hero or a specific card.

**Examples:** `gradient-labs.ai`, `durable.com` (both featured on saaslandingpage.com Q1 2026).

**CSS/code clues:**
```css
.aurora-bg {
  position: relative;
  overflow: hidden;
}
.aurora-bg::before {
  content: '';
  position: absolute;
  inset: -50%;
  background:
    radial-gradient(ellipse 60% 40% at 30% 50%, #7c3aed55, transparent),
    radial-gradient(ellipse 50% 50% at 70% 40%, #06b6d455, transparent),
    radial-gradient(ellipse 40% 60% at 50% 80%, #ec489955, transparent);
  filter: blur(60px);
  animation: aurora-shift 12s ease-in-out infinite alternate;
}
@keyframes aurora-shift {
  from { transform: translate(-5%, -5%) scale(1); }
  to   { transform: translate(5%, 5%) scale(1.1); }
}
```

---

## TYPOGRAPHY TRENDS

### 6. Oversized Display + Tight Letter-Spacing ("Death to Safe Type")
**What it looks like:** Headlines at 100px–200px viewport-relative sizes. Letter-spacing of -0.04em to -0.06em (negative tracking). Either: (a) a grotesque extended sans at heavy weight (Neue Haas Grotesk, ABC Diatype, Satoshi Black), or (b) a high-contrast editorial serif (Editorial New, Playfair at 900 weight). Body copy is small and light (14px, weight 300) — extreme contrast between headline and body.

**Key pattern:** The headline is the product. No subhead needed. One 4–8 word statement in 160px that IS the value prop.

**CSS/code clues:**
```css
h1.hero {
  font-size: clamp(56px, 10vw, 180px);
  font-weight: 900;
  letter-spacing: -0.05em;
  line-height: 0.95;
  text-wrap: balance;
}
```

### 7. Variable Fonts + Hover Kinetics
**What it looks like:** Words where individual letters stretch on hover (font-variation-settings animating `wdth` or `wght` axes). A menu item expands its letterforms as you hover. Not a CSS scale transform — the letterform itself morphs.

**CSS/code clues:**
```css
@font-face {
  font-family: 'Variable';
  src: url('font-variable.woff2');
  font-variation-settings: 'wdth' 75, 'wght' 400;
}
.nav-item {
  transition: font-variation-settings 0.3s ease;
}
.nav-item:hover {
  font-variation-settings: 'wdth' 125, 'wght' 700;
}
```

---

## ANIMATION / MOTION TRENDS

### 8. ASCII / Character-Based Rendering
**What it looks like:** Images or 3D renders that display as ASCII character grids. As you interact (mouse hover, scroll), the resolution improves — the ASCII resolves into a real photo/render. It's a loading aesthetic turned into art. Black background, monospace font, green or white characters.

**Examples:** `artefakt.mov` — ASCII post-processing over particles on mouse interaction; ASCII image pixel reveal as the primary hero animation. Both won Awwwards SOTD April 3, 2026.

**Code approach:**
```js
// Canvas-based ASCII renderer
function renderASCII(imageData, density = ' .:-=+*#@') {
  const chars = density.split('');
  // Map pixel brightness → character index
  ctx.font = `${charSize}px monospace`;
  for (let y = 0; y < h; y += charSize) {
    for (let x = 0; x < w; x += charSize) {
      const i = (y * w + x) * 4;
      const brightness = (data[i] + data[i+1] + data[i+2]) / 3;
      const char = chars[Math.floor(brightness / 255 * (chars.length - 1))];
      ctx.fillText(char, x, y);
    }
  }
}
```

### 9. Scroll-Scrubbed 3D / Interactive Exploded View
**What it looks like:** A 3D product model that your scroll position controls. Scroll down = the product slowly rotates, or explodes into its component parts showing internals. Not autoplay — your scroll IS the playhead. Used heavily for physical products (hardware, devices) but also for abstract "platform" visualizations in SaaS.

**Examples:** `iyo.ai` — "Interactive WebGL Exploded View" of the physical product; "3D Product Experience & Configurator." Awwwards SOTD Apr 2, 2026.

**Code approach:**
```js
// Scroll-scrubbed Three.js model rotation
gsap.to(model.rotation, {
  y: Math.PI * 2,
  scrollTrigger: {
    trigger: '#model-section',
    start: 'top top',
    end: 'bottom bottom',
    scrub: true,
  }
});
```

---

## AUTH / ONBOARDING FLOW TRENDS

### 10. No Login Wall — Value-First, Gate-Second
**The dominant pattern validated by NN/Group research (still the gold standard in 2026):** Show the full product experience before asking for an account. Users who hit a login wall before seeing value drop off immediately.

**What it looks like in practice:**
- Homepage → full interactive demo (no auth)
- "Start free" CTA → drops directly into a sandboxed product environment with dummy data
- Account creation prompt appears only when the user tries to *save* or *share*
- The signup form is minimal: email + password only, or SSO first

**Examples from land-book sign-up gallery (2025–2026):**
- `ElevenLabs` — minimal auth, product-first (featured sign-up page, Webflow)
- `Cohere` — clean dashboard-first experience before account gating
- `Coder` — "Request a Demo" as the primary CTA, not signup

**NN/Group principle:** "Login walls are a nuisance on sites that people visit only rarely. Use them only if users benefit significantly." The modern pattern is: Guest → Explore → Prompted save → Account.

### 11. Split-Screen Auth with Live Product Preview
**What it looks like:** The signup/login page is split 50/50. Left panel: a real (or animated) product screenshot — the dashboard, a chart, a workflow — that slowly animates or scrolls. Right panel: the form. The product is always visible during auth. No empty white form on a blank page.

**Examples:** Multiple sign-up page designs in land-book's curated gallery use this pattern (Kinhive, Wittl, Coder).

**CSS skeleton:**
```css
.auth-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  height: 100dvh;
}
.auth-preview {
  background: var(--bg-dark);
  overflow: hidden;
  position: relative;
}
.auth-preview__inner {
  /* Slow upward scroll of a product screenshot */
  animation: slow-scroll 30s linear infinite;
}
@keyframes slow-scroll {
  from { transform: translateY(0); }
  to   { transform: translateY(-50%); }
}
.auth-form-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
}
```

### 12. Checklist-Driven First-Run ("Progress as Motivation")
**What it looks like:** After signup, the user sees a dashboard with a visible checklist: "Complete your profile → Connect your first integration → Invite a teammate." Each step is a large clickable card, not a tooltip. Completing steps fills a progress bar. The CTA for each step is embedded inline — no navigation required.

**CSS/code clues:**
```css
.onboarding-step {
  display: grid;
  grid-template-columns: 40px 1fr auto;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  border: 1px solid var(--border);
  border-radius: 12px;
  transition: border-color 0.2s, background 0.2s;
}
.onboarding-step.completed {
  background: oklch(0.97 0.02 145);
  border-color: oklch(0.6 0.18 145);
}
.onboarding-step .step-number {
  width: 40px; height: 40px;
  border-radius: 50%;
  display: grid; place-items: center;
  background: var(--accent);
  color: white;
  font-weight: 700;
}
```

### 13. Bento Grid Feature Sections
**What it looks like:** Feature sections built as asymmetric card grids (CSS Grid with `grid-column: span 2` on key cards). Each card has a different size — some wide, some tall, some square. Cards have a subtle dark glass background (glassmorphism: `backdrop-filter: blur(16px)` + semi-transparent border). Cards contain mini-animations: a chart that draws itself, a terminal that types, a map that pans.

**Examples:** `sierra.ai`, `llamaindex.ai`, `sanity.io` — all featured on saaslandingpage.com Q1 2026. This replaced the old "3 columns of icons + text" feature layout entirely.

**CSS/code clues:**
```css
.bento-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  grid-auto-rows: 200px;
  gap: 1rem;
}
.bento-card {
  background: oklch(0.18 0 0 / 0.6);
  border: 1px solid oklch(1 0 0 / 0.08);
  border-radius: 20px;
  backdrop-filter: blur(16px);
  -webkit-backdrop-filter: blur(16px);
  padding: 1.5rem;
  overflow: hidden;
}
.bento-card--wide    { grid-column: span 8; }
.bento-card--narrow  { grid-column: span 4; }
.bento-card--tall    { grid-row: span 2; }
.bento-card--feature { grid-column: span 6; grid-row: span 2; }
```

---

## SUMMARY TABLE

| Trend | Where | Key CSS Technique | Real Sites |
|---|---|---|---|
| Procedural WebGL Hero | Hero section | Three.js + GLSL shader | iyo.ai, artefakt.mov, vastspace.com |
| Scroll-Scrubbed 3D | Product sections | GSAP ScrollTrigger + Three.js | iyo.ai, fluid.glass |
| Mask Wipe Transitions | Page transitions | `clip-path` animated on scroll | fluid.glass, vastspace.com |
| ASCII Render Effect | Hero / gallery | Canvas + monospace character mapping | artefakt.mov |
| Charcoal + Single Accent | Full palette | `oklch()` color tokens | vastspace.com, sierra.ai |
| Aurora Mesh Gradient | Hero bg | `radial-gradient` blobs + `filter:blur` | gradient-labs.ai, durable.com |
| 160px Tight-Tracked Headline | Hero type | `clamp()` + `letter-spacing: -0.05em` | Most Awwwards SOTD 2026 |
| Variable Font Hover | Nav / headings | `font-variation-settings` transition | artefakt.mov |
| No Login Wall / Value-First | Auth flow | Guest-mode → prompted save | ElevenLabs, Cohere |
| Split-Screen Auth | Sign-up page | `grid-template-columns: 1fr 1fr` | Kinhive, Wittl, Coder |
| Checklist First-Run | Onboarding | Progress bar + inline CTAs | Standard SaaS pattern |
| Bento Grid Features | Feature section | CSS Grid `span` + glassmorphism | sierra.ai, llamaindex.ai, sanity.io |

---

*Sources: Awwwards.com SOTD March 30 – April 3 2026; land-book.com/design/sign-up-page; saaslandingpage.com (featured April 2026: Plasticity, Gradient Labs, Sierra, Durable, Playground); lapa.ninja homepage (2025–2026 featured); NN/Group "Login Walls Stop Users in Their Tracks."*

---

## Part 2: Competitor Analysis (Live fetches, 2026-04-03)

### Semaphore CI — semaphore.io

**Headline/tagline:** "The best CI/CD tool for high-performance engineering teams."
Meta description: "A CI/CD platform for the entire software delivery workflow—fast pipelines, secure builds, and predictable costs from commit to production."

**Color scheme / vibe:** WordPress-based site (wp-content paths visible in HTML). Standard light-mode marketing layout. Green brand accent. FAQ accordion-heavy homepage built for SEO, not experience. Feels like mid-tier SaaS from 2021.

**Target audience feel:** Engineering managers at mid-size tech companies. "High-performance teams" aspiration but generic execution.

**Weaknesses:**
- WordPress backend signals the product team didn't build the marketing site
- FAQ-heavy homepage = SEO over user experience
- No dark mode, no cinematic hero, no interactive product demo
- Messaging competes on speed/cost — commodity positioning with no differentiation
- The name collision with semaphoreui.com (the Ansible tool) creates persistent brand confusion in search

---

### AWX / Red Hat Ansible Automation Platform

**Headline/tagline:** "Red Hat Ansible Automation Platform — Turn automation into your strategic advantage." / "Implement enterprise-wide automation."

**Color scheme / vibe:** Red Hat red + white + grey. PatternFly enterprise design system. Dense navigation mega-menu. Built for procurement committees, not individual engineers. Every page routes to "Start a trial / Buy a learning subscription / Contact sales."

**Target audience feel:** CIO/CISO buyers at Fortune 500. Zero developer empathy.

**Weaknesses:**
- AWX (the open-source upstream) has essentially no marketing site — redirects to GitHub
- PatternFly design system is functional but has no personality or delight
- Bureaucratic UX: 3 clicks to find what the product actually does
- Enterprise PDF-brochure energy: icon grids, bullet lists, no visual product demo
- Pricing requires "Contact sales" — instant turnoff for self-hosted adopters
- No dark mode, no scroll animation, no interactive preview

---

### Semaphore UI — semaphoreui.com

**Headline/tagline:** "Modern UI and powerful API for Ansible, Terraform, OpenTofu, PowerShell and other DevOps tools."
Hero cycles: "Modern UI for Ansible / Terraform / OpenTofu / PowerShell / Terragrunt..."

**Color scheme / vibe:** Light blue + white. Material Design aesthetic (Material Icons literally visible in HTML as "keyboard_arrow_right"). Clean but generic. Well-maintained open-source project page energy, not a premium product. Stats: 13K+ GitHub Stars, 2M+ Docker Hub Pulls, 20K+ Installations.

**Target audience feel:** Individual DevOps engineers and homelab enthusiasts. Community-first with Pro/Enterprise tier being added.

**Weaknesses:**
- Rotating hero headline is a 2019 pattern — signals indecision about positioning
- "Effortlessly manage the tasks" — generic DevOps copy, no emotional hook
- Material Icons in visible HTML = dated component library, no design system investment
- Light mode default with no premium dark variant
- Comparison pages (vs AWX, vs Rundeck, vs GitLab etc.) = reactive, not confident
- Emoji in nav items (✨) undercuts enterprise credibility they're chasing
- Social proof is quantity-based (GitHub stars), not quality-based (which teams use it)

---

### Rundeck — rundeck.com / pagerduty.com/platform/rundeck/

**Headline/tagline:** "Rundeck Runbook Automation — Job scheduling and infrastructure management. Self-service for DevOps and platform engineering teams."

**Color scheme / vibe:** Absorbed into PagerDuty's brand (navy/green/white). The rundeck.com domain still exists but feels like a neglected side project. PagerDuty's main site treats Rundeck as a "platform" add-on feature, not a standalone product. The pagerduty.com/platform/rundeck/ URL returns a 404.

**Target audience feel:** SRE/ops teams at enterprises already using PagerDuty for incident management.

**Weaknesses:**
- Product identity is diluted — is it Rundeck or PagerDuty Runbook Automation?
- "Download OSS" vs "Try Commercial" = confusing dual positioning
- No product screenshots or live demo on the homepage
- Generic tagline: "Job scheduling and infrastructure management" — could describe cron + bash
- Primary above-fold CTA is a webinar registration — extremely low urgency
- rundeck.com has almost no content; footer links to PagerDuty California Privacy Notice

---

## Part 3: Reference Site Deep-Dive (Live fetches, 2026-04-03)

### requesty.ai — The target aesthetic

**What it is:** AI gateway/router between apps and 400+ LLM providers.

**Aesthetic:** Dark canvas. Hero features the actual product dashboard — real analytics, model cost breakdown (Opus 42.6%, GPT 25.5%, Gemini 19.6%, DeepSeek 12.7%, Llama 8%). Data is real-looking and readable in 3 seconds. Code snippet with Python/Node/cURL tab switcher shows the 3-line integration. Ambient blue-purple glow around the dashboard card.

**Key design moves:**
- Product UI IS the hero — not above it, not beside it
- Multi-model cost chart communicates ROI instantly (show savings, don't describe them)
- Code tabs with language switcher — developer empathy baked into the hero
- "Plug & Play — Integrate in a minute / Integrate in just 3 lines of code"
- PII scrubbing demo with before/after output — feature demo in 3 lines
- Headline specificity: "400+ LLM providers", "143.2K requests", "$1,247"

**Color:** Near-black base, subtle blue-purple ambient glow, white text, chart accent colors (each model gets a distinct hue). No decoration outside of the product data itself.

---

### linear.app — The benchmark for dark developer SaaS

**Headline:** "The product development system for teams and agents"
**Sub:** "Purpose-built for planning and building products. Designed for the AI era."

**Aesthetic:** Three themes available (dark / light / glass) via `data-theme` attribute. Default is dark. Typography uses CSS custom properties for 4 levels of text opacity (primary, secondary, tertiary, quaternary). Tight letter-spacing throughout.

**Key design moves:**
- "Issue tracking is dead / linear.app/next" — provocative, category-defining positioning
- Hero shows actual AI agent activity in a real issue thread — animated, not illustrated
- "A new species of product tool" — positions as category inventor, not feature-competitor
- Section header: "Make product operations self-driving" — bold, specific, confident
- No pricing in hero — the product sells on desirability, not price comparison
- `["dark","light","glass"]` theme options signal investment in the product

**What makes it premium:** Restraint. One accent color (Linear purple). No decorative elements. Every pixel is either content or intentional negative space.

---

### vercel.com — AI Cloud developer platform

**Title tag:** "Vercel: Build and deploy the best web experiences with the AI Cloud"

**Aesthetic:** System-adaptive (`zeit-theme` auto dark/light). Geist typeface (their own open-source font — owning your typeface is a premium brand signal). Clean, technically dense but visually calm.

**Key design moves:**
- Own typeface signals brand investment in craft
- "AI Cloud" positioning is 2026-current — not "CI/CD" or "hosting"
- Code blocks protected from 1Password DOM manipulation — cares about dev UX at the detail level
- Framework logos as trust signals — ecosystem, not just product
- Scroll-driven hero transforms (CSS `animation-timeline: scroll()`)

**Color:** White/dark adaptive. Minimal brand color. Blue only for interactive elements.

---

### resend.com — Email for developers

**Title:** "Resend · Email for developers"
**Schema description:** "The best way to reach humans" / "Email API for developers"

**Aesthetic:** Default dark mode baked into localStorage (`theme: "dark"` default, not system preference). Pure black (`#000`) background, white text, maximum contrast. Minimal nav: logo + docs + pricing + login + "Get started." YC-backed startup clarity — no wasted space.

**Key design moves:**
- Brand description is exactly 3 words: "Email for developers" — zero hedging
- Dark-first as default (not system-adaptive) — alignment with developer preference
- Code-first CTAs: copy the npm install command
- No hero illustration — the product interface IS the visual
- YC + founder names as trust signals — who built it matters to developers

**Color:** Pure black, white, single red/orange brand accent for the logo. No gradients, no decoration.

---

## Part 4: Competitive Gap — What Kirmaphore Can Own

| Competitor failure | Kirmaphore opportunity |
|---|---|
| WordPress / Drupal / PatternFly enterprise sites | Feels built by the product team, for engineers |
| Feature-list homepages | Show the product running — terminal output, job history, live logs |
| Light mode defaults | Dark-first by default (like Resend), with optional light |
| "Contact sales" as primary CTA | Self-serve install command as hero CTA |
| Rotating hero headlines (SemaphoreUI) | Single, declarative, confident statement |
| Generic "effortlessly manage your infra" copy | Specific: playbook name, environment, execution time |
| Enterprise procurement feel (Red Hat) | Individual engineer feel — they'll advocate internally |
| 3-column icon feature grids | Bento cards with mini product UI previews (live log tail, RBAC matrix) |
| Quantity-based social proof (GitHub stars) | Quality-based: named teams, specific use cases |
| No product demo in hero | Animated terminal / job execution as the hero itself |

**The white space:** Every competitor is either (a) an enterprise procurement site or (b) a community open-source project page. Nobody in this space has a Linear-quality, Vercel-quality, Resend-quality landing page. That gap is the opportunity.

*Sources: Live HTTP fetches 2026-04-03 — semaphore.io, semaphoreui.com, rundeck.com, redhat.com/en/technologies/management/ansible, requesty.ai, linear.app, vercel.com, resend.com.*
