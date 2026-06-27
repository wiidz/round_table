import { useCallback, useEffect, useState } from 'react'

export type ThemeMode = 'light' | 'dark'

const STORAGE_KEY = 'roundtable-theme'

function readStoredTheme(): ThemeMode | null {
  const value = localStorage.getItem(STORAGE_KEY)
  if (value === 'light' || value === 'dark') return value
  return null
}

function resolveTheme(stored: ThemeMode | null): ThemeMode {
  if (stored) return stored
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function applyTheme(mode: ThemeMode) {
  document.documentElement.classList.toggle('dark', mode === 'dark')
}

export function useTheme() {
  const [theme, setTheme] = useState<ThemeMode>(() => resolveTheme(readStoredTheme()))

  useEffect(() => {
    applyTheme(theme)
    localStorage.setItem(STORAGE_KEY, theme)
  }, [theme])

  const toggleTheme = useCallback(() => {
    setTheme((current) => (current === 'dark' ? 'light' : 'dark'))
  }, [])

  return { theme, setTheme, toggleTheme, isDark: theme === 'dark' }
}
