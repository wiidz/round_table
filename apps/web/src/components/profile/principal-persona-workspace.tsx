import { Plus, UserRound } from 'lucide-react'
import type { ReactNode } from 'react'

import {
  SideTabWorkspace,
  SideTabWorkspaceAddTab,
  SideTabWorkspaceNav,
  SideTabWorkspacePanel,
  SideTabWorkspaceTab,
} from '@/components/layout/side-tab-workspace'
import { DiscordRunningRing } from '@/components/settings/discord-running-ring'
import { useI18n } from '@/hooks/use-i18n'
import { heSectionDesc, heSectionTitle, sideTabIconMotion, sideTabLabelMotion } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { PrincipalPersonaMeta } from '@/types/principal'

function PersonaTab({
  persona,
  selected,
  isActive,
  disabled,
  onSelect,
}: {
  persona: PrincipalPersonaMeta
  selected: boolean
  isActive: boolean
  disabled?: boolean
  onSelect: () => void
}) {
  const { t } = useI18n()
  const fallback = (persona.title.trim() || '?').slice(0, 1).toUpperCase()

  return (
    <SideTabWorkspaceTab
      tone="surface"
      title={
        isActive
          ? `${persona.title} (${t('profile.principal.persona.activeBadge')})`
          : persona.title
      }
      aria-label={
        isActive
          ? `${persona.title} (${t('profile.principal.persona.activeBadge')})`
          : persona.title
      }
      selected={selected}
      disabled={disabled}
      onClick={onSelect}
      className={isActive ? '[--dr-ring-tail:#45dea0] [--dr-ring-mid:#5ef0a8] [--dr-ring-head:#7dffc8] [--dr-track:rgba(47,182,124,0.14)]' : undefined}
    >
      <span className="relative shrink-0">
        <span
          className={cn(
            'flex size-10 shrink-0 items-center justify-center overflow-hidden rounded-xs border p-0.5',
            sideTabIconMotion,
            selected
              ? 'border-2 border-primary/50 bg-brand-soft/70'
              : 'border border-black/[0.08] bg-black/[0.02]',
          )}
        >
          <span className="flex size-full items-center justify-center rounded-[3px] bg-black/[0.04] text-sm font-medium text-text-secondary">
            {fallback === '?' ? (
              <UserRound className="size-5 text-text-tertiary" />
            ) : (
              fallback
            )}
          </span>
        </span>
        {isActive && <DiscordRunningRing emphasis={selected} />}
      </span>

      <span className="flex min-w-0 flex-1 flex-col gap-0.5 leading-none pl-0.5">
        <span
          className={cn(
            'max-w-[6rem] truncate text-[13px]',
            sideTabLabelMotion,
            selected ? 'font-bold text-brand' : 'font-medium text-text-secondary',
          )}
        >
          {persona.title}
        </span>
      </span>
    </SideTabWorkspaceTab>
  )
}

export interface PrincipalPersonaWorkspaceProps {
  personas: PrincipalPersonaMeta[]
  /** 当前浏览的档案 Tab */
  selectedPersonaId: string
  /** 系统激活（会议使用）的档案 */
  activePersonaId: string
  disabled?: boolean
  onSelect: (personaId: string) => void
  children: ReactNode
  panelHeader?: ReactNode
  /** 默认标题行右侧操作区（panelHeader 未提供时生效） */
  panelActions?: ReactNode
  onAddPersona?: () => void
}

export function PrincipalPersonaWorkspace({
  personas,
  selectedPersonaId,
  activePersonaId,
  disabled,
  onSelect,
  children,
  panelHeader,
  panelActions,
  onAddPersona,
}: PrincipalPersonaWorkspaceProps) {
  const { t } = useI18n()
  const selectedTitle =
    personas.find((p) => p.id === selectedPersonaId)?.title ??
    t('profile.principal.preferences')

  return (
    <SideTabWorkspace>
      <SideTabWorkspaceNav aria-label={t('profile.principal.persona.navAriaLabel')}>
        {personas.map((persona) => (
          <PersonaTab
            key={persona.id}
            persona={persona}
            selected={persona.id === selectedPersonaId}
            isActive={persona.id === activePersonaId}
            disabled={disabled}
            onSelect={() => onSelect(persona.id)}
          />
        ))}

        {onAddPersona && (
          <SideTabWorkspaceAddTab
            title={t('profile.principal.persona.new')}
            aria-label={t('profile.principal.persona.new')}
            onClick={onAddPersona}
          >
            <span className="flex size-10 shrink-0 items-center justify-center rounded-xs border border-dashed border-black/[0.12] bg-black/[0.02]">
              <Plus className="size-4" />
            </span>
            <span className="truncate font-medium">{t('profile.principal.persona.new')}</span>
          </SideTabWorkspaceAddTab>
        )}
      </SideTabWorkspaceNav>

      <SideTabWorkspacePanel tone="surface">
        <div className="space-y-6 p-5 sm:p-6">
          {panelHeader ?? (
            <header className="flex items-start justify-between gap-4 border-b border-black/[0.05] pb-5">
              <div className="min-w-0 space-y-2">
                <h2 className={heSectionTitle}>{selectedTitle}</h2>
                <p className={heSectionDesc}>{t('profile.principal.persona.switchHint')}</p>
              </div>
              {panelActions ? (
                <div className="flex shrink-0 items-center gap-2">{panelActions}</div>
              ) : null}
            </header>
          )}
          {children}
        </div>
      </SideTabWorkspacePanel>
    </SideTabWorkspace>
  )
}
