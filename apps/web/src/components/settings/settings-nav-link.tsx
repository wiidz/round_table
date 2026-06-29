import { Link, type LinkProps } from 'react-router-dom'

import { primeSettingsNav, type SettingsNavState } from '@/lib/settings-nav'

interface SettingsNavLinkProps extends LinkProps {
  nav: SettingsNavState
}

/** 跳转设置页并定位到指定 Tab / 子节 */
export function SettingsNavLink({ nav, onClick, ...props }: SettingsNavLinkProps) {
  return (
    <Link
      {...props}
      onClick={(event) => {
        primeSettingsNav(nav)
        onClick?.(event)
      }}
    />
  )
}
