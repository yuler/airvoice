<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetPairingLink, GetQRCode, GetSettings, RefreshPairing } from '../../wailsjs/go/main/App'
import { useConnection } from '../composables/useConnection'
import { useToast } from '../composables/useToast'
import RefreshIcon from './icons/RefreshIcon.vue'

const qrCodeData = ref<string>('')
const pairingLink = ref<string>('')
const pairingToken = ref<string>('')
const isRelayMode = ref(false)
const relayHost = ref('')
const error = ref<string>('')
const refreshing = ref(false)
const copied = ref(false)
let copiedTimer: ReturnType<typeof setTimeout> | null = null

const { status } = useConnection()
const { t } = useI18n()
const { show } = useToast()

const isRelay = computed(() => isRelayMode.value || status.value.mode === 'relay')

const relayBanner = computed(() => {
  if (!isRelay.value) return null
  switch (status.value.state) {
    case 'connecting':
      return { dot: 'bg-status-warning', text: t('status.relayConnecting') }
    case 'disconnected':
      return { dot: 'bg-status-warning', text: t('status.relayReconnecting') }
    case 'error':
      return { dot: 'bg-status-error', text: status.value.error || t('status.relayError') }
    default:
      return null
  }
})

function tokenFromLink(link: string): string {
  try {
    const parsed = JSON.parse(link) as { token?: string }
    return parsed.token || ''
  } catch {
    return ''
  }
}

function hostFromRelayURL(raw: string): string {
  const target = (raw || '').trim() || 'wss://relay.airvoice.yuler.cc/ws'
  try {
    return new URL(target).hostname
  } catch {
    return target
  }
}

async function loadQRCode() {
  error.value = ''
  try {
    const [qr, link, settings] = await Promise.all([
      GetQRCode(),
      GetPairingLink(),
      GetSettings(),
    ])
    qrCodeData.value = qr
    pairingLink.value = link
    isRelayMode.value = settings?.connectionMode === 'relay'
    if (isRelayMode.value) {
      pairingToken.value = ''
      relayHost.value = hostFromRelayURL(settings?.relayUrl || '')
    } else {
      pairingToken.value = tokenFromLink(link)
      relayHost.value = ''
    }
  } catch (e) {
    error.value = t('qr.error')
    console.error(e)
  }
}

async function copyPairingLink() {
  if (!pairingLink.value) return
  try {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      await navigator.clipboard.writeText(pairingLink.value)
    } else {
      const textarea = document.createElement('textarea')
      textarea.value = pairingLink.value
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    copied.value = true
    if (copiedTimer) clearTimeout(copiedTimer)
    copiedTimer = setTimeout(() => {
      copied.value = false
    }, 2000)
    show(t('qr.linkCopied'), 'success')
  } catch (e) {
    show(t('qr.copyError'), 'error')
    console.error(e)
  }
}

async function refreshPairing() {
  if (refreshing.value) return

  refreshing.value = true
  error.value = ''

  try {
    await RefreshPairing()
    await loadQRCode()
    show(t('qr.refreshed'), 'success')
  } catch (e) {
    error.value = t('qr.error')
    show(t('qr.refreshError'), 'error')
    console.error(e)
  } finally {
    refreshing.value = false
  }
}

onMounted(() => {
  loadQRCode()
  const runtime = (window as any).runtime
  if (runtime && runtime.EventsOn) {
    runtime.EventsOn('server_restarted', loadQRCode)
    runtime.EventsOn('pairing_changed', loadQRCode)
  }
})

onUnmounted(() => {
  if (copiedTimer) clearTimeout(copiedTimer)
  const runtime = (window as any).runtime
  if (runtime && runtime.EventsOff) {
    runtime.EventsOff('server_restarted', loadQRCode)
    runtime.EventsOff('pairing_changed', loadQRCode)
  }
})
</script>

<template>
  <div class="flex flex-col items-center gap-4 w-full">
    <div class="text-center flex flex-col items-center gap-1.5">
      <span
        v-if="isRelay"
        class="inline-flex items-center gap-1.5 rounded-xl border border-border-default bg-bg-secondary px-3 py-1 text-xs font-medium text-accent-blue"
      >
        <span class="h-1.5 w-1.5 rounded-full bg-accent-blue" />
        {{ t('qr.relayBadge') }}
      </span>
      <h2 class="text-2xl font-semibold tracking-tight text-primary-text">{{ t('pair.title') }}</h2>
      <p class="text-sm text-secondary-text">{{ t('pair.subtitle') }}</p>
    </div>

    <div
      v-if="relayBanner"
      class="flex w-full items-center justify-center gap-2 rounded-xl border border-border-default bg-bg-secondary px-3 py-2 text-xs text-secondary-text"
    >
      <span class="h-2 w-2 shrink-0 rounded-full" :class="relayBanner.dot" />
      <span class="truncate">{{ relayBanner.text }}</span>
    </div>
    
    <div v-if="error" class="text-status-error text-sm">{{ error }}</div>
    <div v-else-if="qrCodeData" class="w-48 h-48 rounded-xl flex items-center justify-center p-3 overflow-hidden" style="background-color: var(--color-bg-primary); border: 1px solid var(--color-border-default);">
      <img :src="qrCodeData" alt="QR Code" class="block w-full h-full" style="image-rendering: pixelated;" />
    </div>
    <div v-else class="text-secondary-text text-sm">{{ t('qr.loading') }}</div>

    <p
      v-if="isRelay && relayHost"
      class="max-w-full px-4 text-center font-mono text-xs text-muted-text truncate"
    >
      {{ relayHost }}
    </p>
    <p
      v-else-if="pairingToken"
      class="max-w-full px-4 text-center font-mono text-xs text-secondary-text break-all select-all"
      :title="t('qr.tokenHint')"
    >
      {{ pairingToken }}
    </p>
    
    <div class="flex items-center justify-center gap-2">
      <button
        v-if="isRelay"
        type="button"
        :disabled="!pairingLink"
        @click="copyPairingLink"
        class="h-11 px-4 flex items-center justify-center gap-1.5 rounded-full border border-border-default text-xs font-medium text-primary-text hover:bg-bg-secondary transition-colors disabled:opacity-50"
      >
        <svg v-if="copied" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" class="text-status-success">
          <path d="M5 13l4 4L19 7" />
        </svg>
        <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
          <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
        </svg>
        <span>{{ copied ? t('qr.copied') : t('qr.copyLink') }}</span>
      </button>

      <button
        type="button"
        :disabled="refreshing"
        @click="refreshPairing"
        :title="t('qr.refresh')"
        :aria-label="t('qr.refresh')"
        class="w-11 h-11 flex items-center justify-center rounded-full border border-border-default text-primary-text hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <RefreshIcon :class="{ 'animate-spin': refreshing }" />
      </button>
    </div>
  </div>
</template>
