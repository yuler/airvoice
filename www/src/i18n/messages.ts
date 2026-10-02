import type { Translator } from './utils';
import type { Locale } from './ui';

export function headerMessages(t: Translator, lang: Locale) {
  return {
    docs: t('nav.docs'),
    menu: t('nav.menu'),
    langTheme: t('nav.langTheme'),
    switchLangTitle: lang === 'zh' ? t('nav.switchToEn') : t('nav.switchToZh'),
    toggleThemeTitle: t('nav.toggleTheme'),
  };
}

export function heroMessages(t: Translator) {
  return {
    title: t('hero.title'),
    subtitle: t('hero.subtitle'),
    cta: t('hero.cta'),
    docs: t('hero.docs'),
    tabCli: t('hero.tab.cli'),
    tabDesktop: t('hero.tab.desktop'),
    statusConnected: t('hero.status.connected'),
    statusConnecting: t('hero.status.connecting'),
    statusError: t('hero.status.error'),
    statusOffline: t('hero.status.offline'),
  };
}

export function howItWorksMessages(t: Translator) {
  return {
    title: t('howItWorks.title'),
    subtitle: t('howItWorks.subtitle'),
    diagram: {
      phone: t('howItWorks.diagram.phone'),
      phoneDesc: t('howItWorks.diagram.phoneDesc'),
      bridge: t('howItWorks.diagram.bridge'),
      pc: 'PC',
      pcDesc: t('howItWorks.diagram.pcDesc'),
      sampleText: t('howItWorks.diagram.sampleText'),
      lanOnly: t('howItWorks.diagram.lanOnly'),
      ariaLabel: t('howItWorks.diagram.ariaLabel'),
    },
  };
}

export function getStartedMessages(t: Translator) {
  return {
    title: t('getStarted.title'),
    subtitle: t('getStarted.subtitle'),
    pc: t('getStarted.pc'),
    or: t('getStarted.or'),
    copied: t('getStarted.copied'),
    copyCommand: t('getStarted.copyCommand'),
    downloadCli: t('getStarted.downloadCli'),
    desktopDesc: t('getStarted.desktopDesc'),
    downloadDesktop: t('getStarted.downloadDesktop'),
    downloadDesktopWindows: t('getStarted.downloadDesktopWindows'),
    downloadDesktopMacos: t('getStarted.downloadDesktopMacos'),
    downloadDesktopLinux: t('getStarted.downloadDesktopLinux'),
    downloadOtherDesktop: t('getStarted.downloadOtherDesktop'),
    mobile: t('getStarted.mobile'),
    mobileDesc: t('getStarted.mobileDesc'),
    iosBanner: t('getStarted.iosBanner'),
    androidDesc: t('getStarted.androidDesc'),
    downloadApk: t('getStarted.downloadApk'),
    iosDesc: t('getStarted.iosDesc'),
    iosGuide: t('getStarted.iosGuide'),
  };
}

export function featuresMessages(t: Translator) {
  return {
    title: t('features.title'),
    pairing: { title: t('features.pairing.title'), desc: t('features.pairing.desc') },
    sync: { title: t('features.sync.title'), desc: t('features.sync.desc') },
    secure: { title: t('features.secure.title'), desc: t('features.secure.desc') },
    cli: { title: t('features.cli.title'), desc: t('features.cli.desc') },
    desktop: { title: t('features.desktop.title'), desc: t('features.desktop.desc') },
    step1: t('features.step1'),
    step2: t('features.step2'),
    step3: t('features.step3'),
  };
}

export function worksEverywhereMessages(t: Translator) {
  return { title: t('worksEverywhere.title') };
}

export function footerMessages(t: Translator) {
  return {
    tagline: t('footer.tagline'),
    rights: t('footer.rights'),
    product: t('footer.product'),
    resources: t('footer.resources'),
    downloadCli: t('footer.downloadCli'),
    downloadDesktop: t('footer.downloadDesktop'),
    downloadApk: t('footer.downloadApk'),
    releases: t('footer.releases'),
    documentation: t('footer.documentation'),
    quickStart: t('footer.quickStart'),
    architecture: t('footer.architecture'),
  };
}

export function docsMessages(t: Translator) {
  return {
    tocTitle: t('docs.toc.title'),
    backToTop: t('docs.toc.backToTop'),
    prev: t('docs.pagination.prev'),
    next: t('docs.pagination.next'),
    sidebarToggle: t('docs.sidebar.toggle'),
    overview: t('docs.sidebar.overview'),
    guide: t('docs.sidebar.guide'),
    background: t('docs.sidebar.background'),
    quickStart: t('docs.sidebar.quickStart'),
    development: t('docs.sidebar.development'),
    architecture: t('docs.sidebar.architecture'),
    platformDeps: t('docs.sidebar.platformDeps'),
  };
}

export type HeaderMessages = ReturnType<typeof headerMessages>;
export type HeroMessages = ReturnType<typeof heroMessages>;
export type HowItWorksMessages = ReturnType<typeof howItWorksMessages>;
export type GetStartedMessages = ReturnType<typeof getStartedMessages>;
export type FeaturesMessages = ReturnType<typeof featuresMessages>;
export type WorksEverywhereMessages = ReturnType<typeof worksEverywhereMessages>;
export type FooterMessages = ReturnType<typeof footerMessages>;
export type DocsMessages = ReturnType<typeof docsMessages>;
