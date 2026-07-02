/** Layout / style tokens for brief template forms and previews */

/** 子字段标签（无左侧竖线） */
export const briefFieldLabelClass = 'text-[13px] font-semibold text-text-primary'

/** 子字段说明标签 */
export const briefFieldCaptionClass =
  'text-[11px] font-medium uppercase tracking-[0.06em] text-text-tertiary'

/** 大板块间距（不用分隔线，靠留白区分） */
export const briefSectionStackClass = 'space-y-8'

/** 编辑视图大板块之间的横线分隔 */
export const briefTemplateSectionDividerClass = 'border-t border-black/[0.06]'

export const briefTemplateLeftColumnClass =
  'min-w-0 space-y-8 lg:border-r lg:border-black/[0.06] lg:pr-5'

export const briefTemplateRightColumnClass = ''

/** 简报正文双栏：左侧简报字段 + 右侧会议配置（编辑 / 预览共用） */
export const briefTemplateBodyGridClass =
  'grid gap-8 lg:grid-cols-[minmax(0,1fr)_minmax(0,20rem)] lg:gap-x-6'

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

/** 会议详情 · 配置侧栏板块宽度 */
export const meetingDetailConfigPanelClass = 'w-[18rem] max-w-full'

/** 会议配置键值行：标签列约 8 汉字宽；label 与控件首行顶对齐 */
export const briefMeetingConfigRowGrid =
  'grid grid-cols-1 gap-x-3 gap-y-1 sm:grid-cols-[6rem_minmax(0,1fr)] sm:items-start'

/** 会议详情 · 配置行：标签列约 4 汉字宽 */
export const meetingDetailConfigRowGrid =
  'grid grid-cols-1 gap-x-3 gap-y-1 sm:grid-cols-[4rem_minmax(0,1fr)] sm:items-start'

export const briefMeetingConfigLabelClass = briefFieldCaptionClass
