import { defineConfig } from 'astro/config';
import tailwind from '@tailwindcss/vite';
import react from '@astrojs/react';

const isProd = process.env.NODE_ENV === 'production';

export default defineConfig({
  site: 'https://yuler.github.io/airvoice',
  base: isProd ? '/airvoice' : '/',
  trailingSlash: 'always',
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'zh'],
    routing: {
      prefixDefaultLocale: false,
    },
  },
  markdown: {
    shikiConfig: {
      themes: {
        light: 'github-light',
        dark: 'vesper',
      },
    },
  },
  integrations: [react()],
  vite: {
    plugins: [tailwind()],
  },
});
