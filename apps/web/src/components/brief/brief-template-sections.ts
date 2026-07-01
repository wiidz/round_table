/** Layout / style tokens for brief template forms and previews */

/** 子字段标签（无左侧竖线） */
export const briefFieldLabelClass = 'text-[13px] font-semibold text-text-primary'

/** 子字段说明标签 */
export const briefFieldCaptionClass =
  'text-[11px] font-medium uppercase tracking-[0.06em] text-text-tertiary'

/** 大板块间距（不用分隔线，靠留白区分） */
export const briefSectionStackClass = 'space-y-8'

export const briefTemplateLeftColumnClass =
  'min-w-0 space-y-8 lg:border-r lg:border-black/[0.06] lg:pr-5'

export const briefTemplateRightColumnClass = ''

export const briefScopeBlockShell = 'rounded-lg px-5 py-4'

export const briefScopeBlockTone = {
  inScope: 'bg-brand-soft/35',
  outOfScope: 'bg-danger-soft/45',
  done: 'bg-success-soft/40',
} as const

export const briefScopeIconShell =
  'flex size-7 shrink-0 items-center justify-center rounded-full bg-surface/80'

export const briefAgendaItemShell = 'flex gap-3 rounded-xs bg-black/[0.025] px-3 py-3'

export const briefConfigPanelShell = 'space-y-3'

/** 会议配置键值行：标签列固定约 4 汉字宽（4rem），label / value 垂直居中 */
export const briefMeetingConfigRowGrid =
  'grid grid-cols-1 gap-x-3 gap-y-1 sm:grid-cols-[4rem_minmax(0,1fr)] sm:items-center'

export const briefMeetingConfigLabelClass = briefFieldCaptionClass
