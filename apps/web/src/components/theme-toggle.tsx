import { Moon, Sun } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { useTheme } from '@/hooks/use-theme'

export function ThemeToggle() {
  const { isDark, toggleTheme } = useTheme()

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleTheme}
      aria-label={isDark ? '切换到亮色主题' : '切换到暗色主题'}
    >
      {isDark ? <Sun /> : <Moon />}
    </Button>
  )
}
