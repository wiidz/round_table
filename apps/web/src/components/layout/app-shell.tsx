import { NavLink, Outlet } from 'react-router-dom'

import { ThemeToggle } from '@/components/theme-toggle'
import { domainNavLabel } from '@/lib/ui-labels'
import { cn } from '@/lib/utils'

const navItems = [
  { to: '/', label: '概览', end: true },
  { to: '/chat', label: '聊天', end: false },
  { to: '/meetings', label: '会议', end: false },
  { to: '/brief-templates', label: '简报模板', end: false },
  { to: '/principals', label: domainNavLabel('principal'), end: false },
  { to: '/participants', label: domainNavLabel('participant'), end: false },
  { to: '/settings', label: '设置', end: false },
]

export function AppShell() {
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
                <p className="text-xs text-text-tertiary">委托人工作台</p>
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
      <main className="mx-auto max-w-6xl px-4 py-8 sm:px-6">
        <Outlet />
      </main>
    </div>
  )
}
