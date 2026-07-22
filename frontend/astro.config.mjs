// @ts-check
import { defineConfig } from 'astro/config';

import tailwindcss from '@tailwindcss/vite';

import node from '@astrojs/node';
import cxCommons from '@vdatacloud/cx-commons';

// https://astro.build/config
export default defineConfig({
  output: 'server',

  server: {
    port: 4321,
    host: true,
  },

  integrations: [cxCommons()],

  vite: {
    plugins: [tailwindcss()]
  },

  adapter: node({
    mode: 'standalone'
  })
});