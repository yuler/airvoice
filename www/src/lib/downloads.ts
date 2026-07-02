const TAG = 'v0.3.1';
const VERSION = TAG.replace(/^v/, '');
const BASE = `https://github.com/yuler/airvoice/releases/download/${TAG}`;
export const RELEASES = 'https://github.com/yuler/airvoice/releases/latest';
export const ANDROID_APK_URL = `${BASE}/airvoice-android-${VERSION}.apk`;
export const REPO_URL = 'https://github.com/yuler/airvoice';

export interface DownloadUrls {
  cli: string;
  desktop: string;
  mobile: string;
}

export type MobilePlatform = 'android' | 'ios' | 'other';

export interface MobileDownloadInfo {
  platform: MobilePlatform;
  apkUrl: string;
  iosDocsPath: string;
}

function detectPlatform(): {
  isWindows: boolean;
  isMac: boolean;
  isLinux: boolean;
  isAndroid: boolean;
  isIOS: boolean;
  isARM: boolean;
} {
  if (typeof navigator === 'undefined') {
    return {
      isWindows: false,
      isMac: false,
      isLinux: false,
      isAndroid: false,
      isIOS: false,
      isARM: false,
    };
  }

  const ua = navigator.userAgent;
  const platform = (navigator as Navigator & { userAgentData?: { platform: string } })
    .userAgentData?.platform ?? navigator.platform ?? '';

  return {
    isWindows: /Win/i.test(platform) || /Windows/i.test(ua),
    isMac: /Mac/i.test(platform) || /Macintosh/i.test(ua),
    isLinux: /Linux/i.test(platform) && !/Android/i.test(ua),
    isAndroid: /Android/i.test(ua),
    isIOS: /iPhone|iPad|iPod/i.test(ua),
    isARM: /arm64|aarch64/i.test(platform) || /arm64/i.test(ua),
  };
}

export function getMobilePlatform(): MobilePlatform {
  const { isAndroid, isIOS } = detectPlatform();
  if (isAndroid) return 'android';
  if (isIOS) return 'ios';
  return 'other';
}

export function getMobileDownloadInfo(lang: 'en' | 'zh'): MobileDownloadInfo {
  return {
    platform: getMobilePlatform(),
    apkUrl: ANDROID_APK_URL,
    iosDocsPath: lang === 'zh' ? 'zh/docs/quick-start/#ios' : 'docs/quick-start/#ios',
  };
}

export function getDownloadUrls(): DownloadUrls {
  const { isWindows, isMac, isLinux, isAndroid, isIOS, isARM } = detectPlatform();

  let cli = RELEASES;
  let desktop = RELEASES;
  let mobile = RELEASES;

  if (isWindows) {
    cli = `${BASE}/airvoice-cli-windows-amd64.exe`;
    desktop = `${BASE}/Airvoice-Desktop-${VERSION}-Windows.zip`;
  } else if (isMac) {
    cli = isARM
      ? `${BASE}/airvoice-cli-darwin-arm64`
      : `${BASE}/airvoice-cli-darwin-amd64`;
    desktop = `${BASE}/Airvoice-Desktop-${VERSION}-macOS.zip`;
  } else if (isLinux) {
    cli = `${BASE}/airvoice-cli-linux-amd64`;
    desktop = `${BASE}/Airvoice-Desktop-${VERSION}-Linux.tar.gz`;
  }

  if (isAndroid) {
    mobile = ANDROID_APK_URL;
  } else if (isIOS) {
    mobile = getMobileDownloadInfo('en').iosDocsPath;
  }

  return { cli, desktop, mobile };
}
