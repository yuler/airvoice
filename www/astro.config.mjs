import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { defineConfig } from 'astro/config';
import tailwind from '@tailwindcss/vite';
import react from '@astrojs/react';

function resolveAppVersion() {
  const envVersion = (process.env.AIRVOICE_VERSION || '').trim().replace(/^v/, '');
  if (/^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?$/.test(envVersion)) {
    return envVersion;
  }

  try {
    const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..');
    const out = execFileSync('git', ['tag', '--list', 'v*', '--sort=-v:refname'], {
      cwd: repoRoot,
      encoding: 'utf8',
      stdio: ['ignore', 'pipe', 'ignore'],
    });
    const latest = out
      .split(/\r?\n/)
      .map((line) => line.trim())
      .find((line) => /^v\d+\.\d+\.\d+$/.test(line));
    if (latest) {
      return latest.slice(1);
    }
  } catch {
    // Fallback when built outside a git checkout or without tags
  }

  return '0.0.0';
}

const appVersion = resolveAppVersion();

export default defineConfig({
  site: 'https://airvoice.yuler.cc',
  base: '/',
  trailingSlash: 'always',
  i18n: {
    defaultLocale: 'zh',
    locales: ['zh', 'en'],
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
    define: {
      __APP_VERSION__: JSON.stringify(appVersion),
    },
    plugins: [tailwind()],
    server: {
      fs: {
        allow: ['..'],
      },
    },
  },
});
