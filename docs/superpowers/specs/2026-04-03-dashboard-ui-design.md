# Dashboard UI & Onboarding Design Spec

## Vision

Kirmaphore должен выглядеть как лучший infra SaaS 2026 года — тёмный, точный, живой. Референсы: Supabase, Raycast, Railway. Пользователь открывает дашборд и сразу понимает что это премиальный продукт.

---

## Design Decisions (утверждены)

| Вопрос | Решение |
|---|---|
| Визуальный стиль | **Vibrant Pro** — индиго/фиолет, тёмный фон |
| Первый логин | **Hero Empty State** — центр экрана, одна кнопка |
| Сайдбар | **Hybrid Collapsible** — 220px с текстом, сворачивается в 56px |

---

## Color System

Заменяет текущие shadcn/ui light-mode токены. Вся палитра строится на двух акцентах.

```css
/* globals.css — добавить в :root */
--dash-bg:        #0d0d1a;   /* основной фон */
--dash-surface:   rgba(99,102,241,0.04);  /* фон сайдбара, карточек */
--dash-border:    rgba(99,102,241,0.10);  /* разделители */
--dash-border-hi: rgba(99,102,241,0.25);  /* hover/active borders */

--dash-indigo:    #6366f1;   /* primary accent */
--dash-violet:    #8b5cf6;   /* secondary accent */
--dash-gradient:  linear-gradient(135deg, #6366f1, #8b5cf6);

--dash-text:      #ffffff;
--dash-text-2:    rgba(255,255,255,0.5);   /* secondary text */
--dash-text-3:    rgba(255,255,255,0.25);  /* muted text / labels */

--dash-green:     #22c55e;   /* success / running */
--dash-yellow:    #f59e0b;   /* warning / pending */
--dash-red:       #ef4444;   /* error / failed */
```

**Типографика:** Inter (уже в Next.js). Размеры: 13px nav labels, 15px page titles, 20px hero title. `font-weight: 600` для заголовков.

**Иконки:** Lucide React, `stroke-width="1.75"`, размер 16px в nav, 20px в topbar.

---

## Layout Architecture

```
┌─────────────────────────────────────────────────────┐
│  Sidebar (220px / 56px collapsed)  │  Main           │
│  ┌──────────────────────────────┐  │  ┌───────────┐  │
│  │  Workspace Switcher          │  │  │  Topbar   │  │
│  ├──────────────────────────────┤  │  ├───────────┤  │
│  │  Nav Items                   │  │  │           │  │
│  │  · Projects  [badge]         │  │  │  Content  │  │
│  │  · Runs      [live dot]      │  │  │  Area     │  │
│  │  · Templates                 │  │  │           │  │
│  │  · Secrets                   │  │  │           │  │
│  │                              │  │  │           │  │
│  │  ── Team ──                  │  │  │           │  │
│  │  · Members                   │  │  └───────────┘  │
│  ├──────────────────────────────┤  │                  │
│  │  [← Collapse]                │  │                  │
│  ├──────────────────────────────┤  │                  │
│  │  Settings                    │  │                  │
│  │  Kirill · Free plan  ···     │  │                  │
│  └──────────────────────────────┘  │                  │
└─────────────────────────────────────────────────────┘
```

### Sidebar — детали

**Expanded (220px):**
- Workspace Switcher: logo (22px gradient square) + workspace name + chevron, background `rgba(255,255,255,0.04)`, border-radius 7px
- Nav item: 16px Lucide icon + label + optional badge, padding `7px 8px`, border-radius 6px
- Active state: `rgba(99,102,241,0.15)` bg + 2px left border `#6366f1`
- Badges: count badge `rgba(99,102,241,0.25)` / `#818cf8` text; live dot `#22c55e` с glow
- Section labels: `text-[9px] uppercase tracking-widest text-white/25`
- Collapse button: border `rgba(255,255,255,0.06)`, текст "Collapse" с иконкой `‹`
- User block: avatar (26px gradient circle) + name + plan + `···` menu

**Collapsed (56px):**
- Только иконки, центрированные
- Workspace: только logo square
- Active: та же левая полоска
- Tooltip при hover: `bg-indigo-600/90`, positioned справа от иконки
- Collapse button → "Expand" кнопка (иконка `›`)
- User: только avatar

**Transition:** `width` CSS transition `200ms ease`, иконки/labels fade через `opacity`.

### Topbar

Высота 48px. `border-bottom: 1px solid rgba(255,255,255,0.05)`. `background: rgba(99,102,241,0.02)`.
- Левая часть: page title `text-[15px] font-semibold`
- Правая часть: кнопки (Docs — secondary, + New Project — primary gradient)

**Button styles:**
- Primary: `background: var(--dash-gradient)`, `border-radius: 7px`, `box-shadow: 0 2px 12px rgba(99,102,241,0.4)`, `font-weight: 600`
- Secondary: `border: 1px solid rgba(255,255,255,0.1)`, transparent bg, `color: rgba(255,255,255,0.5)`

---

## Pages

### Projects Page — Empty State (first login)

Главный UX момент продукта.

```
┌────────────────────────────────────────┐
│                                        │
│                  ⚡                    │  ← 72px glow circle
│           (gradient border,            │     indigo glow shadow
│            double ring)                │
│                                        │
│     Deploy your first project          │  ← 20px font-weight:700
│  Connect Ansible, define inventory,    │  ← 13px text-white/40
│  automate infrastructure in minutes.   │
│                                        │
│     [  + Create Project  ]             │  ← pill button, gradient
│                                        │
│   View examples   ·   Read the docs    │  ← 12px links, indigo/70
│                                        │
│ ─────── Quick start ──────────────    │
│ ┌──────────┐ ┌──────────┐ ┌────────┐  │
│ │ 📁 Repo  │ │ 🖥️ Inv.  │ │ ▶ Run │  │  ← 3-col cards
│ │ Connect  │ │   Add    │ │ First │  │
│ └──────────┘ └──────────┘ └────────┘  │
└────────────────────────────────────────┘
```

**Glow circle:** 72px, `background: linear-gradient(135deg, rgba(99,102,241,0.25), rgba(139,92,246,0.15))`, `border: 1px solid rgba(99,102,241,0.4)`, `box-shadow: 0 0 40px rgba(99,102,241,0.25), 0 0 80px rgba(99,102,241,0.1)`. Pseudo `::after` — второе кольцо `inset: -8px, border-radius: 50%, border: 1px solid rgba(99,102,241,0.15)`.

**Quick Start Cards (3 шт):**
- `background: rgba(255,255,255,0.02)`, `border: 1px solid rgba(255,255,255,0.06)`, `border-radius: 10px`
- Hover: `border-color: rgba(99,102,241,0.3)`, `background: rgba(99,102,241,0.05)`
- 28px icon square с цветным bg (indigo/green/orange tint)
- Title `12px font-weight:600`, Description `11px text-white/30`

### Projects Page — With Data

Когда проекты есть, hero пропадает, появляется grid:
- 3-колонный card grid (responsive: 2 на планшете)
- Project card: название, статус badge (running/idle/error), кол-во runs, последний запуск, hover — subtle lift + border glow

### Runs Page — Empty State

Аналогично: центрированный empty state "No runs yet. Start from a project."

### Settings Page

Стандартные секции: Profile, Security (passkeys list), Workspace, Billing. Left nav внутри страницы.

---

## Onboarding Flow

**Нет wizard и modal tours.** Вместо этого:

1. **Hero Empty State** — первый экран IS онбординг. Одна кнопка.
2. **Quick Start Cards** — три шага без числовых меток, кликабельны.
3. **Sidebar progress badge** (optional v2): маленький индикатор "Getting started 1/3" в sidebar появляется после создания первого проекта, исчезает после первого run.

---

## Micro-animations

- Sidebar collapse/expand: `transition: width 200ms ease`
- Nav hover: `transition: background 100ms`
- Button hover: `filter: brightness(1.1)`, `transition: filter 150ms`
- Quick Start card hover: `border-color` transition `150ms`
- Status live dot: `animation: pulse 2s ease-in-out infinite` с glow
- Page transitions: `animate-in fade-in` (Next.js default) — не нужна библиотека

---

## File Structure

```
web/
  app/
    globals.css                          ← добавить dash-* переменные
    (dashboard)/
      layout.tsx                         ← переписать: тёмный bg, новый sidebar
      dashboard/
        page.tsx                         ← переписать: пустой, redirect → /projects
      projects/
        page.tsx                         ← новый: hero empty state + grid
      runs/
        page.tsx                         ← новый: runs empty state
      settings/
        page.tsx                         ← новый: settings layout
  components/
    layout/
      Sidebar.tsx                        ← переписать: hybrid collapsible
      Header.tsx                         ← удалить (влито в sidebar bottom + topbar)
      Topbar.tsx                         ← новый: page title + actions slot
    dashboard/
      ProjectCard.tsx                    ← новый: card для проекта с data
      EmptyState.tsx                     ← новый: reusable hero empty state
      QuickStartCards.tsx                ← новый: 3 quick start cards
```

---

## Scope

**В этом плане:**
- Новая цветовая система (CSS vars)
- Sidebar (collapsed/expanded)
- Dashboard layout
- Projects page (empty state + quick start)

**Out of scope (следующий план):**
- Projects page с реальными данными (карточки проектов)
- Runs page с реальным списком
- Settings page с настройками passkeys
- Onboarding sidebar progress badge
