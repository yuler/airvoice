import { ui, defaultLang, showDefaultLang } from './ui';
import type { Locale, UiKey } from './ui';

export type { Locale, UiKey };

export function getLangFromUrl(url: URL): Locale {
  const [, lang] = url.pathname.split('/');
  if (lang in ui) return lang as Locale;
  return defaultLang;
}

export function useTranslations(lang: Locale) {
  return function t(key: UiKey): string {
    return ui[lang][key] ?? ui[defaultLang][key];
  };
}

export type Translator = ReturnType<typeof useTranslations>;

/** Build a locale-aware path. Default locale has no URL prefix when showDefaultLang is false. */
export function useTranslatedPath(lang: Locale) {
  return function translatePath(path: string, locale: Locale = lang): string {
    const normalized = path.startsWith('/') ? path : `/${path}`;
    const withTrailing = normalized.endsWith('/') ? normalized : `${normalized}/`;
    if (!showDefaultLang && locale === defaultLang) {
      return withTrailing.replace(/\/+/g, '/');
    }
    return `/${locale}${withTrailing}`.replace(/\/+/g, '/');
  };
}

/** Strip locale prefix from pathname, returning path without locale (e.g. /docs/foo/). */
export function stripLocaleFromPath(pathname: string): string {
  const parts = pathname.split('/').filter(Boolean);
  if (parts.length > 0 && parts[0] in ui && parts[0] !== defaultLang) {
    parts.shift();
  }
  return parts.length > 0 ? `/${parts.join('/')}/` : '/';
}

export function getAlternateLocale(locale: Locale): Locale {
  return locale === 'zh' ? 'en' : 'zh';
}

/** Path to the same page in the other locale (for language toggle links). */
export function getLangTogglePath(currentPath: string, locale: Locale): string {
  const subPath = stripLocaleFromPath(currentPath);
  const other = getAlternateLocale(locale);
  const translate = useTranslatedPath(other);
  return translate(subPath, other);
}
