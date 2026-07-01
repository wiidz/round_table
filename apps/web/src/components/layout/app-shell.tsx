import { NavLink, Outlet } from 'react-router-dom'

import { ThemeToggle } from '@/components/theme-toggle'
import { useI18n } from '@/hooks/use-i18n'
import { cn } from '@/lib/utils'

export function AppShell() {
  const i18n = useI18n()

  const navItems = [
    { to: '/', label: i18n.navLabel('overview'), end: true },
    { to: '/chat', label: i18n.navLabel('chat'), end: false },
    { to: '/meetings', label: i18n.navLabel('meetings'), end: false },
    { to: '/brief-templates', label: i18n.navLabel('briefTemplates'), end: false },
    { to: '/principals', label: i18n.domainNavLabel('principal'), end: false },
    { to: '/participants', label: i18n.domainNavLabel('participant'), end: false },
    { to: '/settings', label: i18n.navLabel('settings'), end: false },
  ]

  return (
    <div className="min-h-screen bg-canvas">
      <header className="sticky top-0 z-40 border-b border-border-subtle bg-surface/90 backdrop-blur-sm">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4 sm:px-6">
          <div className="flex items-center gap-8">
            <div className="flex items-center gap-2">
              <span className="inline-flex size-8 items-center justify-center rounded-lg bg-brand-soft text-sm font-semibold text-brand">
                RT
              </span>
              <div>
                <p className="text-sm font-semibold tracking-tight">RoundTable</p>
                <p className="text-xs text-text-tertiary">{i18n.navLabel('workbench')}</p>
              </div>
            </div>
            <nav className="hidden items-center gap-1 sm:flex">
              {navItems.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.end}
                  className={({ isActive }) =>
                    cn(
                      'rounded-md px-3 py-1.5 text-sm text-text-secondary transition-colors hover:text-text-primary',
                      isActive && 'bg-brand-soft text-brand',
                    )
                  }
                >
                  {item.label}
                </NavLink>
              ))}
            </nav>
          </div>
          <ThemeToggle />
        </div>
      </header>

      <div className="px-4 py-8 sm:px-6">
        <Outlet />
      </div>
    </div>
  )
}
