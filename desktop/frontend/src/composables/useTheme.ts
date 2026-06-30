import { ref, watchEffect } from 'vue'

const theme = ref<'light' | 'dark'>(
  localStorage.getItem('theme') as 'light' | 'dark' ||
  (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
)

function applyTheme() {
  document.documentElement.classList.toggle('dark', theme.value === 'dark')
}

applyTheme()

watchEffect(() => {
  localStorage.setItem('theme', theme.value)
  applyTheme()
})

export function useTheme() {
  function toggle() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
  }

  return { theme, toggle }
}
