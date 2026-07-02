const TAG = 'v0.3.1';
const BASE = `https://github.com/yuler/airvoice/releases/download/${TAG}`;
const RELEASES = 'https://github.com/yuler/airvoice/releases/latest';

export interface DownloadUrls {
  cli: string;
  desktop: string;
  mobile: string;
}

export function getDownloadUrls(): DownloadUrls {
  if (typeof navigator === 'undefined') {
    return { cli: RELEASES, desktop: RELEASES, mobile: RELEASES };
  }

  const ua = navigator.userAgent;
  const platform = (navigator as Navigator & { userAgentData?: { platform: string } })
    .userAgentData?.platform ?? navigator.platform ?? '';

  const isWindows = /Win/i.test(platform) || /Windows/i.test(ua);
  const isMac = /Mac/i.test(platform) || /Macintosh/i.test(ua);
  const isLinux = /Linux/i.test(platform) && !/Android/i.test(ua);
  const isAndroid = /Android/i.test(ua);
  const isIOS = /iPhone|iPad|iPod/i.test(ua);
  const isARM = /arm64|aarch64/i.test(platform) || /arm64/i.test(ua);

  let cli = RELEASES;
  let desktop = RELEASES;
  let mobile = RELEASES;

  if (isWindows) {
    cli = `${BASE}/airvoice-cli-windows-amd64.exe`;
    desktop = `${BASE}/Airvoice-Desktop-${TAG}-Windows.zip`;
  } else if (isMac) {
    cli = isARM
      ? `${BASE}/airvoice-cli-darwin-arm64`
      : `${BASE}/airvoice-cli-darwin-amd64`;
    desktop = `${BASE}/Airvoice-Desktop-${TAG}-macOS.zip`;
  } else if (isLinux) {
    cli = `${BASE}/airvoice-cli-linux-amd64`;
    desktop = `${BASE}/Airvoice-Desktop-${TAG}-Linux.tar.gz`;
  }

  if (isAndroid) {
    mobile = `${BASE}/airvoice-android-${TAG}.apk`;
  } else if (isIOS) {
    mobile = 'https://github.com/yuler/airvoice';
  }

  return { cli, desktop, mobile };
}
