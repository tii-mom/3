import { ref } from 'vue'

const isDark = ref(false)

// Initialize state if in browser environment
if (typeof document !== 'undefined') {
  isDark.value = document.documentElement.classList.contains('dark')
}

export function useTheme() {
  function toggleTheme() {
    if (typeof document === 'undefined') return
    isDark.value = !isDark.value
    document.documentElement.classList.toggle('dark', isDark.value)
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
  }

  function syncThemeFromDOM() {
    if (typeof document === 'undefined') return
    isDark.value = document.documentElement.classList.contains('dark')
  }

  return {
    isDark,
    toggleTheme,
    syncThemeFromDOM
  }
}
