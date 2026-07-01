/** 简报模板主体四板块的标题与说明文案 */
export const BRIEF_TEMPLATE_SECTIONS = {
  topicGoal: {
    title: '主题与目标',
    description: '预填议题与期望结论；主题留空时，可在开会时再指定。',
  },
  agenda: {
    title: '议程',
    description: '按顺序列出本场需要逐项讨论的议题。',
  },
  scope: {
    title: '讨论边界',
    description: '明确可讨论范围、排除项，以及怎样算开完这场会。',
  },
  meeting: {
    title: '会议配置',
    description: '模式、轮次、确认关与专家阵容等运行参数。',
  },
} as const

/** 主题未填写时的展示文案 */
export const BRIEF_TOPIC_EMPTY_COPY = {
  preview: '暂未指定，可在开会时再定',
  placeholder: '留空表示开会时再指定',
} as const

/** 会议配置字段标签（预览 / 编辑共用，右栏宜短） */
export const BRIEF_MEETING_CONFIG_LABELS = {
  mode: '模式',
  confirmation: '确认关',
  maxRounds: '辩论轮次',
  minSynthesis: '合成轮次',
  freeDialogue: '自由对话',
  experts: '专家',
} as const

/** 讨论边界预览空态文案 */
export const BRIEF_SCOPE_EMPTY_COPY = {
  inScope: '未填写讨论范围',
  outOfScope: '未填写排除项',
  doneCriteria: '未填写完成标准',
} as const

/** 子字段标签（无左侧竖线） */
export const briefFieldLabelClass = 'text-[13px] font-semibold text-text-primary'

/** 子字段说明标签 */
export const briefFieldCaptionClass = 'text-[11px] font-medium uppercase tracking-[0.06em] text-text-tertiary'

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

export const briefScopeIconShell = 'flex size-7 shrink-0 items-center justify-center rounded-full bg-surface/80'

export const briefAgendaItemShell = 'flex gap-3 rounded-xs bg-black/[0.025] px-3 py-3'

export const briefConfigPanelShell = 'space-y-3'

/** 会议配置键值行：标签列固定约 4 汉字宽（4rem），label / value 垂直居中 */
export const briefMeetingConfigRowGrid =
  'grid grid-cols-1 gap-x-3 gap-y-1 sm:grid-cols-[4rem_minmax(0,1fr)] sm:items-center'

export const briefMeetingConfigLabelClass = briefFieldCaptionClass
