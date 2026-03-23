// @ts-check
import { defineConfig } from 'astro/config';

import tailwindcss from '@tailwindcss/vite';

// https://astro.build/config
export default defineConfig({
  output: 'server',
  server: {
    port: 8080,
    host: true,
  },
  vite: {
    plugins: [tailwindcss()]
  }
});