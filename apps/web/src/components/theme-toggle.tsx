import { Moon, Sun } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { useI18n } from '@/hooks/use-i18n'
import { useTheme } from '@/hooks/use-theme'

export function ThemeToggle() {
  const { isDark, toggleTheme } = useTheme()
  const { t } = useI18n()

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleTheme}
      aria-label={isDark ? t('common.theme.toLight') : t('common.theme.toDark')}
    >
      {isDark ? <Sun /> : <Moon />}
    </Button>
  )
}
