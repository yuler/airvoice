<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useToast } from '../composables/useToast'

interface Settings {
  token?: string
  relayKey?: string
  port: number
  autoStart: boolean
  language: string
  connectionMode: 'lan' | 'relay'
  relayUrl: string
  relayAuthKey: string
}

const DEFAULT_RELAY_URL = 'wss://relay.airvoice.yuler.cc/ws'

const settings = ref<Settings>({
  port: 7655,
  autoStart: false,
  language: 'zh-CN',
  connectionMode: 'lan',
  relayUrl: '',
  relayAuthKey: '',
})

const isOpen = ref(false)
const saving = ref(false)
const saveError = ref('')
const { t } = useI18n()
const { show } = useToast()

async function loadSettings() {
  try {
    const loaded = await window.go.main.App.GetSettings()
    settings.value = {
      ...loaded,
      connectionMode: loaded.connectionMode === 'relay' ? 'relay' : 'lan',
      relayUrl: loaded.relayUrl || '',
      relayAuthKey: loaded.relayAuthKey || '',
    }
  } catch (e) {
    console.error('Failed to load settings:', e)
  }
}

async function openPanel() {
  saveError.value = ''
  await loadSettings()
  isOpen.value = true
}

async function saveSettings() {
  if (saving.value) return
  saveError.value = ''
  saving.value = true
  try {
    await window.go.main.App.SaveSettings(settings.value)
    isOpen.value = false
    show(t('settings.saved'), 'success')
  } catch (e: any) {
    saveError.value = (typeof e === 'string' ? e : e?.message) || t('settings.saveError')
    console.error('Failed to save settings:', e)
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>

<template>
  <div>
    <button
      type="button"
      @click="openPanel"
      :title="t('settings.title')"
      :aria-label="t('settings.title')"
      class="w-8 h-8 flex items-center justify-center text-secondary-text hover:text-primary-text transition-colors"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
      </svg>
    </button>

    <Teleport to="body">
      <div
        v-if="isOpen"
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
        @click.self="isOpen = false"
      >
        <div class="bg-bg-secondary border border-border-default rounded-2xl w-full max-w-[288px] max-h-[460px] overflow-y-auto p-4">
          <h2 class="text-base font-semibold text-primary-text mb-4">{{ t('settings.title') }}</h2>

          <div class="space-y-4">
            <div>
              <label class="block text-xs text-secondary-text mb-1.5">{{ t('settings.connectionMode') }}</label>
              <div class="grid grid-cols-2 gap-1 p-1 rounded-xl bg-bg-primary border border-border-default">
                <button
                  type="button"
                  @click="settings.connectionMode = 'lan'"
                  :class="settings.connectionMode === 'lan'
                    ? 'bg-accent-blue text-white'
                    : 'text-secondary-text hover:text-primary-text'"
                  class="py-1.5 text-xs font-medium rounded-lg transition-colors"
                >
                  {{ t('settings.modeLan') }}
                </button>
                <button
                  type="button"
                  @click="settings.connectionMode = 'relay'"
                  :class="settings.connectionMode === 'relay'
                    ? 'bg-accent-blue text-white'
                    : 'text-secondary-text hover:text-primary-text'"
                  class="py-1.5 text-xs font-medium rounded-lg transition-colors"
                >
                  {{ t('settings.modeRelay') }}
                </button>
              </div>
            </div>

            <div v-if="settings.connectionMode === 'lan'">
              <label class="block text-xs text-secondary-text mb-1.5">{{ t('settings.port') }}</label>
              <input
                v-model.number="settings.port"
                type="number"
                min="1024"
                max="65535"
                class="w-full px-3 py-2 text-sm font-mono bg-bg-primary border border-border-default rounded-xl text-primary-text focus:outline-none focus:ring-2 focus:ring-accent-blue"
              />
            </div>

            <template v-else>
              <div>
                <label class="block text-xs text-secondary-text mb-1.5">{{ t('settings.relayUrl') }}</label>
                <input
                  v-model="settings.relayUrl"
                  type="text"
                  :placeholder="DEFAULT_RELAY_URL"
                  class="w-full px-3 py-2 text-xs font-mono bg-bg-primary border border-border-default rounded-xl text-primary-text placeholder:text-muted-text focus:outline-none focus:ring-2 focus:ring-accent-blue"
                />
                <p class="mt-1 text-[11px] text-muted-text">{{ t('settings.relayUrlHint') }}</p>
              </div>

              <div>
                <label class="block text-xs text-secondary-text mb-1.5">{{ t('settings.relayAuthKey') }}</label>
                <input
                  v-model="settings.relayAuthKey"
                  type="password"
                  :placeholder="t('settings.relayAuthKeyPlaceholder')"
                  class="w-full px-3 py-2 text-xs font-mono bg-bg-primary border border-border-default rounded-xl text-primary-text placeholder:text-muted-text focus:outline-none focus:ring-2 focus:ring-accent-blue"
                />
              </div>
            </template>

            <div class="flex items-center justify-between pt-1">
              <label class="text-xs text-secondary-text">{{ t('settings.autoStart') }}</label>
              <button
                type="button"
                @click="settings.autoStart = !settings.autoStart"
                :class="settings.autoStart ? 'bg-accent-blue' : 'bg-border-default'"
                class="w-10 h-6 rounded-full relative transition-colors"
              >
                <span
                  :class="settings.autoStart ? 'translate-x-5' : 'translate-x-1'"
                  class="absolute top-1 left-0 w-4 h-4 bg-white rounded-full transition-transform"
                ></span>
              </button>
            </div>
          </div>

          <div v-if="saveError" class="mt-3 text-xs text-status-error break-words">{{ saveError }}</div>

          <div class="flex justify-end gap-2 mt-5">
            <button
              type="button"
              @click="isOpen = false"
              class="px-4 py-2 text-xs text-secondary-text hover:text-primary-text rounded-full border border-border-default transition-colors"
            >
              {{ t('settings.cancel') }}
            </button>
            <button
              type="button"
              :disabled="saving"
              @click="saveSettings"
              class="px-4 py-2 text-xs font-medium bg-accent-blue text-white rounded-full disabled:opacity-50 transition-opacity"
            >
              {{ t('settings.save') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
