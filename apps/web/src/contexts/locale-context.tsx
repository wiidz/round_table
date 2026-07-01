import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'

import { fetchSettings } from '@/api/settings'
import { defaultLocale, localeFromSettingsFields, type AppLocale } from '@/lib/locale'
import type { SettingsFieldState } from '@/types/settings'

interface LocaleContextValue {
  locale: AppLocale
  loading: boolean
  applyLocaleFromFields: (fields: SettingsFieldState[]) => void
  reloadLocale: () => Promise<void>
}

const LocaleContext = createContext<LocaleContextValue | null>(null)

export function LocaleProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<AppLocale>(defaultLocale)
  const [loading, setLoading] = useState(true)

  const applyLocaleFromFields = useCallback((fields: SettingsFieldState[]) => {
    setLocale(localeFromSettingsFields(fields))
  }, [])

  const reloadLocale = useCallback(async () => {
    const data = await fetchSettings()
    applyLocaleFromFields(data.fields ?? [])
  }, [applyLocaleFromFields])

  useEffect(() => {
    let cancelled = false
    fetchSettings()
      .then((data) => {
        if (!cancelled) applyLocaleFromFields(data.fields ?? [])
      })
      .catch(() => {
        // keep default locale
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [applyLocaleFromFields])

  useEffect(() => {
    document.documentElement.lang = locale === 'en' ? 'en' : 'zh-CN'
  }, [locale])

  const value = useMemo(
    () => ({ locale, loading, applyLocaleFromFields, reloadLocale }),
    [locale, loading, applyLocaleFromFields, reloadLocale],
  )

  return <LocaleContext.Provider value={value}>{children}</LocaleContext.Provider>
}

export function useLocale() {
  const ctx = useContext(LocaleContext)
  if (!ctx) {
    throw new Error('useLocale must be used within LocaleProvider')
  }
  return ctx
}
