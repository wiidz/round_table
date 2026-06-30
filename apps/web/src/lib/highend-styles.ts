/** RoundTable Web · High-End Flat（DESIGN.md § High-End Flat） */

import { cn } from '@/lib/utils'

export const heSpring =
  'transition-[color,background-color,box-shadow,transform,opacity] duration-500 ease-[cubic-bezier(0.32,0.72,0,1)] motion-reduce:transition-none'

/** 侧栏 Tab 布局缓动曲线（CRM DetailDialog 同款） */
export const sideTabLayoutEase = 'ease-[cubic-bezier(0.22,1,0.36,1)]'

/** 侧栏 Tab 未激活边框色 */
export const sideTabInactiveBorderClass = 'border-black/[0.04]'

/**
 * 侧栏 Tab 按钮动效：width/margin 等布局属性 200ms 过渡；
 * 背景色不在 transition 内，激活时立刻换底（避免灰底扫过）。
 */
export const sideTabButtonMotion = cn(
  '!shadow-none motion-reduce:transition-none',
  '!transition-[width,margin,min-height,padding,border-color,color]',
  'duration-200',
  sideTabLayoutEase,
)

export const sideTabIconMotion = cn(
  'transition-[transform,color] duration-200 motion-reduce:transition-none',
  sideTabLayoutEase,
)

export const sideTabLabelMotion = cn(
  'transition-[color] duration-200 motion-reduce:transition-none',
  sideTabLayoutEase,
)

export const hePressable = `${heSpring} active:scale-[0.98] motion-reduce:active:scale-100`

/** Chat side rails beside max-w-6xl main (50vw - 50% = gutter to viewport edge). */
export const chatSideRailLeftClass = [
  'absolute right-full top-0 z-20 mr-4 flex h-full min-w-[20rem]',
  'w-[calc(50vw-50%-1rem)]',
].join(' ')

export const chatSideRailRightClass = [
  'absolute left-full top-0 z-20 ml-4 flex h-full min-w-[20rem]',
  'w-[calc(50vw-50%-1rem)]',
].join(' ')

/** Scroll container — inherits global CRM scrollbar */
export const heScrollbar = 'overscroll-contain'

export const hePageTitle =
  'text-balance text-[28px] font-semibold leading-[1.5] tracking-[-0.03em] text-text-primary'

export const hePageDesc = 'mt-1.5 text-[14px] leading-[1.65] text-text-secondary'

export const heEyebrowBrand =
  'inline-flex shrink-0 rounded-full px-3 py-1 text-[10px] font-medium uppercase tracking-[0.18em] text-brand ring-1 ring-primary/20'

export const heEyebrowAI =
  'inline-flex shrink-0 rounded-full px-3 py-1 text-[10px] font-medium uppercase tracking-[0.18em] text-ai/80 ring-1 ring-ai/15'

export const hePanelShell = [
  'overflow-hidden rounded-[1.75rem] border-0 bg-surface',
  'ring-1 ring-[var(--panel-shell-ring)]',
  'shadow-[var(--panel-shell-shadow)]',
  heSpring,
].join(' ')

export const hePanelShellHover = [
  'hover:ring-primary/20 hover:shadow-[var(--panel-hover-shadow)]',
].join(' ')

export const heColumnTitleBrand = [
  'border-l-[3px] border-primary/35 pl-3',
  'text-sm font-normal tracking-[-0.01em] text-text-secondary',
].join(' ')

export const heColumnTitleAI = [
  'border-l-[3px] border-ai/35 pl-3',
  'text-sm font-normal tracking-[-0.01em] text-text-secondary',
].join(' ')

export const heFieldLabel =
  'text-xs font-medium uppercase tracking-[0.12em] text-text-tertiary'

/** Settings / form — section heading (16px semibold, DESIGN.md Section Title) */
export const heSectionTitle =
  'text-base font-semibold tracking-[-0.02em] text-text-primary'

/** Settings / form — section lead (Meta 13px) */
export const heSectionDesc = 'text-[13px] leading-relaxed text-text-secondary'

/** Settings / form — field helper under inputs */
export const heFieldHint = 'text-[13px] leading-relaxed text-text-tertiary'

/** Settings / form — mono meta (env keys, storage paths) */
export const heFieldMeta = 'font-mono text-[11px] leading-relaxed text-text-tertiary/75'

/** Settings / form — embedded edit surface (DESIGN.md High-End Flat field area) */
export const heFormEmbed = [
  'rounded-xl bg-canvas',
  'shadow-[var(--field-inset-shadow)]',
  'ring-1 ring-inset ring-[var(--field-ring)]',
].join(' ')

/** Chat composer · CRM TodoCreateAiInput hero shell */
export const chatComposerOuterClass = [
  'rounded-[1.25rem] bg-[var(--hero-outer-bg)] p-1.5',
  'ring-1 ring-[var(--hero-outer-ring)]',
  heSpring,
  'focus-within:ring-ai/30',
  'focus-within:[box-shadow:var(--ai-glow-shadow)]',
].join(' ')

export const chatComposerInnerClass = [
  'flex items-end gap-3 rounded-[calc(1.25rem-6px)] bg-surface p-3',
  'shadow-[var(--field-inset-shadow)]',
].join(' ')

export const chatComposerTextareaClass = [
  'min-h-[2.75rem] max-h-32 flex-1 resize-none border-0 bg-transparent px-1 py-1',
  'text-[14px] leading-[1.65] text-text-primary shadow-none',
  'outline-none focus:outline-none focus-visible:outline-none',
  'ring-0 ring-transparent focus:ring-0 focus-visible:ring-0',
  'placeholder:text-text-tertiary',
  'disabled:cursor-not-allowed disabled:opacity-60',
].join(' ')

export const chatComposerSendClass = [
  hePressable,
  'inline-flex shrink-0 items-center gap-1.5 rounded-xl px-4 py-2.5',
  'text-sm font-medium text-white bg-brand',
  'shadow-[0_10px_28px_-10px_rgba(232,93,4,0.48)]',
  'hover:bg-[var(--brand-color-hover)]',
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand/35 focus-visible:ring-offset-2 focus-visible:ring-offset-surface',
  'disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none',
].join(' ')

export const heSubsectionTitleNeutral = [
  'border-l-[3px] border-black/[0.08] pl-3',
  'text-[13px] font-medium tracking-[0.06em] text-text-secondary uppercase',
].join(' ')

export const heFieldSurface = [
  'rounded-xl border-0 bg-surface box-border',
  'shadow-[var(--field-inset-shadow)]',
  'ring-1 ring-inset ring-[var(--field-ring)]',
  heSpring,
  'focus-within:ring-2 focus-within:ring-inset focus-within:ring-primary/45',
  'focus-within:shadow-[var(--field-focus-shadow)]',
  'autofill:shadow-[inset_0_0_0_1000px_var(--surface)] autofill:[-webkit-text-fill-color:var(--text-primary)]',
].join(' ')

/** Settings — non-editable field value (readOnly / disabled) */
export const heFieldReadonly = [
  '!rounded-xl !bg-black/[0.05] !text-text-tertiary',
  '!shadow-none !ring-black/[0.02]',
].join(' ')

export const heTextarea = [
  'min-h-[420px] w-full resize-y border-0 bg-transparent p-4',
  'font-mono text-[14px] leading-[1.75] text-text-primary',
  'outline-none focus:outline-none focus-visible:outline-none',
  'ring-0 focus:ring-0 placeholder:text-text-tertiary',
].join(' ')

export const heFilePill = [
  'rounded-full px-3.5 py-1.5 text-left text-[13px] font-medium',
  'bg-black/[0.02] text-text-secondary ring-1 ring-inset ring-black/[0.05]',
  heSpring,
  'hover:bg-brand-soft/60 hover:text-brand hover:ring-primary/25',
].join(' ')

export const heFilePillSelected = [
  'rounded-full px-3.5 py-1.5 text-left text-[13px] font-semibold',
  'bg-brand-soft text-brand',
  'ring-1 ring-inset ring-primary/40 shadow-[var(--field-focus-shadow)]',
].join(' ')

/** 会议侧栏文件项（多行内容时用 rounded-md，避免 pill 过长） */
export const heFileNavItem = [
  'rounded-md px-2.5 py-1.5 text-left text-[13px] font-medium',
  'bg-black/[0.02] text-text-secondary ring-1 ring-inset ring-black/[0.05]',
  heSpring,
  'hover:bg-brand-soft/60 hover:text-brand hover:ring-primary/25',
].join(' ')

export const heFileNavItemSelected = [
  'rounded-md px-2.5 py-1.5 text-left text-[13px] font-semibold',
  'bg-brand-soft text-brand',
  'ring-1 ring-inset ring-primary/40 shadow-[var(--field-focus-shadow)]',
].join(' ')

export const heFileBadge = [
  'rounded-full px-2.5 py-0.5 text-[11px] font-medium',
  'bg-black/[0.03] text-text-secondary ring-1 ring-inset ring-black/[0.05]',
].join(' ')

/** 侧栏主交付物标记（未选中行） */
export const hePrimaryDeliverableBadge = [
  'rounded-full px-2 py-0.5 text-[10px] font-medium',
  'bg-brand-soft text-brand ring-1 ring-inset ring-primary/25',
].join(' ')

/** 侧栏主交付物标记（选中行，叠在 brand-soft pill 上） */
export const hePrimaryDeliverableBadgeOnBrand = [
  'rounded-full px-2 py-0.5 text-[10px] font-semibold',
  'bg-white/80 text-brand shadow-sm ring-1 ring-inset ring-primary/35',
].join(' ')

export const heEmptyPanel = [
  hePanelShell,
  'px-8 py-10 text-center',
].join(' ')

/** 设置页左侧悬浮 Tab 列宽（参考 CRM DetailDialog 侧栏） */
export const SETTINGS_SIDE_TAB_WIDTH = '7.5rem'

/** 设置页左侧 TabsList：浮在面板左缘外 */
export const settingsSideTabListClass = cn(
  'hidden shrink-0 flex-col gap-2 self-start overflow-visible lg:flex',
  'pt-12',
)

/** 设置页左侧 Tab 按钮：未选中内缩，选中向左/右延伸与主面板衔接 */
export function settingsSideTabButtonClass(selected: boolean) {
  return cn(
    sideTabButtonMotion,
    'group flex min-h-[3rem] flex-row items-center rounded-l-lg rounded-r-none',
    'border border-r-0 border-l-[3px] cursor-pointer px-2.5 py-2 text-left',
    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 focus-visible:ring-offset-2',
    selected
      ? cn(
          '-ml-3 relative z-10 min-h-[3.25rem] w-[calc(100%+1.25rem)] pl-3 pr-2',
          'border-0 border-l-[3px] border-l-primary !bg-surface font-semibold text-brand',
        )
      : cn(
          'z-0 ml-3 w-[calc(100%-0.75rem)]',
          sideTabInactiveBorderClass,
          'border-l-transparent bg-black/[0.04] font-medium text-[13px] text-text-secondary',
          'hover:bg-black/[0.06] hover:text-text-primary',
        ),
    '[&_[data-tab-icon]>svg]:size-5 [&_[data-tab-icon]>svg]:shrink-0',
    sideTabIconMotion,
    '[&_[data-tab-icon]>svg]:origin-center',
    selected && '[&_[data-tab-icon]>svg]:scale-110 [&_[data-tab-icon]>svg]:text-brand',
    !selected && '[&_[data-tab-icon]>svg]:text-text-tertiary',
  )
}
