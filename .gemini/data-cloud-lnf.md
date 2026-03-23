# DataCloud Look-and-Feel (LNF) Standards

These standards define the **visual and interactive identity** for DataCloud-powered frontends (Astro/Tailwind).

## 1. Color Palette (Tailwind Named Colors)

Always use named colors to support theme mutability and dark mode.

| Role | Palette | Usage |
| :--- | :--- | :--- |
| **Primary** | `blue-600` | Navigation, primary actions, active tabs. |
| **Success** | `green-600` | Completed milestones, Active status, Nominal health. |
| **Dispute** | `red-600` | Raise dispute, error states, blocked actions. |
| **Neutral** | `slate` | Backgrounds, borders, secondary text. |
| **Oversight** | `indigo-900` | Central Bank / Admin branding blocks. |

### Dark Mode Mapping
- **Light:** `bg-slate-50`, `text-slate-900`, `border-slate-200`.
- **Dark:** `bg-slate-950`, `text-slate-100`, `border-slate-800`.
- **Primary Text:** Always ensure high contrast (`dark:text-white`).

## 2. Typography & Weight

- **Headings:** `font-black` (900 weight) with `tracking-tight` for a modern, high-assurance feel.
- **Sub-headings:** `font-bold` (700 weight).
- **Body:** `font-medium` (500 weight) for improved readability on high-DPI displays.
- **Mono:** `font-mono` for IDs and Ledger hashes (e.g., `text-[10px] uppercase tracking-widest`).

## 3. Component Standards

### Cards (Escrow/Wallet)
- **Border:** `border border-slate-200 dark:border-slate-800`.
- **Shadow:** `shadow-sm hover:shadow-md transition-all duration-300`.
- **Radius:** `rounded-2xl` (for pages/containers) or `rounded-xl` (for cards).
- **Padding:** Minimum `p-6`.

### Interactive Elements
- **Buttons:** `rounded-xl` or `rounded-lg`.
- **Interactions:** Use `active:scale-95 transition-all` for tactile feedback.
- **States:** Always handle `:disabled` with `opacity-30 pointer-events-none`.

### Observability Patterns
- **Graphs:** SVG-based line/area graphs using `stroke-width="3"` and linear gradients for fill.
- **Indicators:** Ping animations (`animate-ping`) for "Live" or "Nominal" system status.

## 4. Branding
- **Logo:** Always use `favicon.svg`. 
- **Logo Size:** `w-10 h-10` in navigation bars.
- **Signage:** Footer MUST include "Built with Daml 3.x + Astro + Go • Powered by DataCloud".
