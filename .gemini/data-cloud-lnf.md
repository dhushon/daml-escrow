# Data Cloud — Look & Feel Standards

This document describes the design system as it actually exists in `frontend/src`,
not an aspirational spec. Where the running code and this document disagree, the
code is currently right and this file is stale, please open an issue rather than
silently diverging further.

Last audited against: `frontend/src/styles/global.css`, `layouts/Layout.astro`,
`components/EscrowCard.astro`, `components/Footer.astro`,
`components/LifecycleTracker.astro`, `components/UploadZone.astro`,
`components/RefreshSlider.astro`, `components/InviteModal.astro`,
`components/OperationalVelocity.astro`.

---

## 1. Stack

Tailwind **v4**, CSS-first configuration. There is no `tailwind.config.js` /
`.mjs` in this project, everything lives in `frontend/src/styles/global.css`
under an `@theme` block. If you're used to v3's JS config file, don't go
looking for it, it isn't there.

```css
@import "tailwindcss";
@theme { ... }
```

Framework: Astro, `.astro` single-file components (frontmatter script + markup
+ optional scoped `<style>`). Font: **Plus Jakarta Sans**, loaded via Google
Fonts `@import` at the top of `global.css`, set as `--font-sans` in the theme.
Don't introduce another font stack in a component-scoped `<style>` block, see
§6 for why that matters.

---

## 2. Color

### 2.1 Brand scale

Defined in `@theme` as `--color-brand-*`. Currently sparse:

| Token | Hex |
|---|---|
| `brand-50` | `#eff6ff` |
| `brand-500` | `#3b82f6` |
| `brand-600` | `#2563eb` |
| `brand-700` | `#1d4ed8` |

**Use `brand-*`, not `blue-*`, for anything that represents the product's
primary color.** `brand-500` currently equals Tailwind's stock `blue-500`,
that's intentional (nothing needed to change visually when the token was
introduced) but it's a coincidence you shouldn't rely on. Several components
still reach for raw `blue-*` classes, see §7, "Known drift."

### 2.2 Neutral scale

Standard Tailwind `slate-50` through `slate-950`, redeclared explicitly in
`@theme` rather than inherited, so don't assume Tailwind's defaults apply if
you ever strip the `@theme` block down. Slate is the only neutral in use;
don't introduce `gray-*` or `zinc-*`.

### 2.3 Status colors

There is currently **no single source of truth** for escrow-status colors.
`EscrowCard.astro` and `OperationalVelocity.astro` each pick their own colors
for the same underlying states, and they disagree (`OperationalVelocity` uses
plain `green-500` and inline hex `#8b5cf6` for concepts `EscrowCard` renders
in `emerald-500` and never uses violet for at all). The mapping below is the
one `EscrowCard.astro` follows most closely; treat it as the standard going
forward and migrate the outliers.

| Status | Color | Typical use |
|---|---|---|
| Draft | `slate-400` / `slate-500` | Neutral, not yet active |
| Funded | `brand-500` | Asset locked, awaiting activation |
| Active | `emerald-500` | Escrow live |
| Proposed | `amber-500` | Mediated settlement awaiting ratification |
| Disputed | `rose-500` | Adjudication in progress |
| Settled | `purple-500` | Terms satisfied, awaiting disbursement |

These are now also available as `status-*` theme tokens (`bg-status-active`,
`text-status-disputed`, etc.), see the updated `global.css`. Prefer the named
token over hand-picking `emerald-500` inline, it's one place to change the
mapping later instead of every component that renders a status.

---

## 3. Typography

- Headings: `font-black`, `tracking-tight`.
- Sub-headings / card titles: `font-bold`.
- Body copy: `font-medium`, `text-slate-600` (light) / `text-slate-400` (dark).
- IDs, hashes, addresses, anything ledger-native: `font-mono`.

### Eyebrow labels

A small-caps-style micro-label precedes most section and card headers
throughout the product (`EscrowCard`, `OperationalVelocity`, `RefreshSlider`
all use it, with very slightly different sizing each time). Standardize on:

```html
<div class="text-[10px] font-black uppercase tracking-widest text-slate-400">
  Label Text
</div>
```

Or use the new `.label-eyebrow` utility added to `global.css` (§ below),
which is exactly the above. When the label needs a color instead of neutral
(e.g. marking which pillar a card belongs to), swap `text-slate-400` for
`text-brand-600` or the relevant `status-*` token, keep the size/weight/
tracking identical.

---

## 4. Shape, elevation, motion

- Containers: `rounded-3xl`. Cards inside containers: `rounded-2xl`.
  Pills/badges: `rounded-full`.
- Padding floor: `p-6` for a card, don't go tighter than that on-screen. (Print
  layouts are the one exception, see §8.)
- Shadows: `shadow-sm` at rest, `hover:shadow-md` or `hover:shadow-lg` on
  interactive cards. Don't hand-write box-shadow values for ordinary
  elevation, use the `.interactive-card` utility class for hover-lift cards.
- Transitions: `transition-all` with Tailwind's default duration is fine for
  most things. For the signature "slide up on mount" and "fade in" effects,
  use the `animate-slide-up` / `animate-fade-in` theme animations already
  defined, don't write new `@keyframes` per component (LifecycleTracker did
  this, see §7).

---

## 5. Glassmorphism utilities

Defined once in `global.css`, reused everywhere: `.glass-card`, `.glass-nav`,
`.glow-blue`, `.glow-green`. These exist specifically so components don't
each reinvent `backdrop-filter: blur(...)` with slightly different values.
If a new component needs a translucent, blurred surface, use one of these
rather than writing a new backdrop-filter rule.

---

## 6. Dark mode

Every component supports dark mode via Tailwind's `dark:` variant,
`dark:bg-slate-900`, `dark:text-white`, `dark:border-slate-800`, and so on,
applied alongside the light-mode class on the same element. This is
class-based dark mode (a `.dark` ancestor class, see `.dark .glass-card` in
`global.css`), not the `prefers-color-scheme` media query. A component that
skips `dark:` variants entirely (LifecycleTracker did, forcing a dark
background unconditionally) will look broken or inconsistent depending on
which mode the rest of the page is in. Every new component needs `dark:`
coverage on every color utility, not just the container background.

---

## 7. Known drift (fix opportunistically, don't let it spread)

- **`LifecycleTracker.astro`** — hand-rolled `<style>` block, hardcoded hex,
  Inter font, no dark mode, custom `@keyframes pulse` duplicating
  `animate-pulse-slow`. Rewritten version included alongside this doc.
- **`RefreshSlider.astro`, `InviteModal.astro`, `UploadZone.astro`,
  `OperationalVelocity.astro`** — all use raw `blue-*` / `green-*` Tailwind
  classes instead of `brand-*` / `status-*` tokens. Not visually broken today
  (the values happen to match), but it means the token layer isn't actually
  doing its job. Migrate opportunistically when touching these files; not
  urgent enough to warrant a dedicated pass on its own.

---

## 8. Print output

Astro pages in this project don't currently need print styles, but if you
ever build a printable artifact off this system (a cut sheet, an exportable
statement, anything meant to become a PDF), one thing will bite you:
**`md:` breakpoint grids collapse to a single column when rendered by most
print engines**, because print engines don't compute the page's content
width the same way a browser viewport does, even though the physical page is
wide enough for the grid to fit. Pair every `md:grid-cols-*` used in a
printable context with an explicit `print:grid-cols-*` of the same value,
don't rely on the breakpoint alone to survive into print.
