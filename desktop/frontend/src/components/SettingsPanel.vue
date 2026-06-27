<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

interface Settings {
  port: number
  autoStart: boolean
  language: string
}

const settings = ref<Settings>({
  port: 7383,
  autoStart: false,
  language: 'zh-CN',
})

const isOpen = ref(false)
const { t } = useI18n()

async function loadSettings() {
  try {
    settings.value = await window.go.main.App.GetSettings()
  } catch (e) {
    console.error('Failed to load settings:', e)
  }
}

async function saveSettings() {
  try {
    await window.go.main.App.SaveSettings(settings.value)
    isOpen.value = false
  } catch (e) {
    console.error('Failed to save settings:', e)
  }
}

onMounted(loadSettings)
</script>

<template>
  <div>
    <button
      @click="isOpen = true"
      class="px-3 py-1.5 text-sm text-text-secondary hover:text-text-primary border border-border-default rounded-full"
    >
      {{ t('settings.title') }}
    </button>

    <Teleport to="body">
      <div
        v-if="isOpen"
        class="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
        @click.self="isOpen = false"
      >
        <div class="bg-bg-secondary border border-border-default rounded-lg w-80 p-4">
          <h2 class="text-lg font-semibold mb-4">{{ t('settings.title') }}</h2>

          <div class="space-y-4">
            <div>
              <label class="block text-sm text-text-secondary mb-1">{{ t('settings.port') }}</label>
              <input
                v-model.number="settings.port"
                type="number"
                min="1024"
                max="65535"
                class="w-full px-3 py-2 bg-bg-primary border border-border-default rounded-md text-text-primary"
              />
            </div>

            <div class="flex items-center justify-between">
              <label class="text-sm text-text-secondary">{{ t('settings.autoStart') }}</label>
              <button
                @click="settings.autoStart = !settings.autoStart"
                :class="settings.autoStart ? 'bg-accent-blue' : 'bg-border-default'"
                class="w-10 h-6 rounded-full relative transition-colors"
              >
                <span
                  :class="settings.autoStart ? 'translate-x-5' : 'translate-x-1'"
                  class="absolute top-1 w-4 h-4 bg-white rounded-full transition-transform"
                ></span>
              </button>
            </div>
          </div>

          <div class="flex justify-end gap-2 mt-6">
            <button
              @click="isOpen = false"
              class="px-4 py-2 text-sm text-text-secondary hover:text-text-primary"
            >
              {{ t('settings.cancel') }}
            </button>
            <button
              @click="saveSettings"
              class="px-4 py-2 text-sm bg-accent-blue text-white rounded-full"
            >
              {{ t('settings.save') }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
