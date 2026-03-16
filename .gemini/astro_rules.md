# Astro Frontend Rules

## Component Structure

- **Frontmatter Fence:** Keep server-side JS (imports, data fetching) strictly inside `---`.
- **Type Safety:** Define `interface Props` for every component.
- **Styles:** Use `<style>` tags (scoped by default). Use `is:global` ONLY for resets or theming.

## Fetching Data from Go Backend

- **Server-Side Fetching:** Prefer fetching data in the frontmatter `--- await fetch() ---` so HTML renders with data populated.
- **Client-Side Fetching:** If fetching in a `<script>`, use strict async/await and handle loading states visually.

## Performance & UX

- **Images:** Use `<Image />` from `astro:assets` for local images; explicit width/height for remote.
- **View Transitions:** If modifying navigation, ensure `<ViewTransitions />` is preserved in the Layout.
- **Accessibility:**
  - All interactive elements must have `:focus-visible` styles.
  - Use semantic HTML (`<main>`, `<article>`, `<nav>`) over `<div>` soup.

## Anti-Patterns (Do Not Do)

- Do NOT use `useEffect` patterns; use standard JS event listeners in `<script>`.
- Do NOT import React/Vue unless explicitly requested.
