import { Children, isValidElement, type ReactNode } from 'react'

export interface MarkdownHeading {
  level: number
  text: string
  id: string
}

/** 顶栏 sticky + 留白，与标题 scroll-mt 对齐 */
export const MARKDOWN_HEADING_SCROLL_OFFSET = 88

const FLASH_CLASS = 'md-heading-target-flash'
const HEADING_LABEL_SELECTOR = '[data-md-heading-label]'

/** Strip common inline Markdown from heading text */
export function normalizeHeadingText(raw: string): string {
  return raw
    .replace(/!\[([^\]]*)\]\([^)]+\)/g, '$1')
    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1')
    .replace(/[*_`~]/g, '')
    .trim()
}

export function headingTextFromChildren(children: ReactNode): string {
  let text = ''
  Children.forEach(children, (child) => {
    if (typeof child === 'string' || typeof child === 'number') {
      text += String(child)
      return
    }
    if (isValidElement<{ children?: ReactNode }>(child)) {
      text += headingTextFromChildren(child.props.children)
    }
  })
  return text
}

export function collectMarkdownHeadingsFromDom(
  root: HTMLElement,
): MarkdownHeading[] {
  const nodes = root.querySelectorAll<HTMLElement>('h1[id], h2[id], h3[id]')
  const seen = new Set<string>()

  return Array.from(nodes).flatMap((node) => {
    const id = node.id
    if (!id || seen.has(id)) return []
    seen.add(id)

    return [{
      level: Number(node.tagName.charAt(1)),
      text:
        node.querySelector<HTMLElement>(HEADING_LABEL_SELECTOR)?.textContent?.trim() ??
        node.textContent?.trim() ??
        '',
      id,
    }]
  })
}

function nextHeadingId(index: number): string {
  return `md-h-${index}`
}

export function createHeadingIdRegistry() {
  const collected: MarkdownHeading[] = []

  return {
    reset() {
      collected.length = 0
    },
    register(level: number, rawText: string): string {
      const text = normalizeHeadingText(rawText)
      const id = nextHeadingId(collected.length)
      collected.push({ level, text, id })
      return id
    },
    getCollected(): MarkdownHeading[] {
      return [...collected]
    },
  }
}

export function headingsEqual(
  a: MarkdownHeading[],
  b: MarkdownHeading[],
): boolean {
  if (a.length !== b.length) return false
  return a.every(
    (heading, index) =>
      heading.id === b[index]?.id &&
      heading.text === b[index]?.text &&
      heading.level === b[index]?.level,
  )
}

const HEADING_FLASH_MS = 3000

export function flashHeadingElement(headingEl: HTMLElement): void {
  const target =
    headingEl.querySelector<HTMLElement>(HEADING_LABEL_SELECTOR) ?? headingEl

  target.classList.remove(FLASH_CLASS)
  void target.offsetWidth
  target.classList.add(FLASH_CLASS)

  const remove = () => {
    target.classList.remove(FLASH_CLASS)
    target.removeEventListener('animationend', remove)
  }
  target.addEventListener('animationend', remove)
  window.setTimeout(remove, HEADING_FLASH_MS + 100)
}

export function scrollToHeading(
  id: string,
  options?: { onDone?: () => void },
): void {
  const el = document.getElementById(id)
  if (!el) {
    options?.onDone?.()
    return
  }

  flashHeadingElement(el)

  const targetTop = () =>
    el.getBoundingClientRect().top +
    window.scrollY -
    MARKDOWN_HEADING_SCROLL_OFFSET

  window.scrollTo({ top: Math.max(0, targetTop()), behavior: 'smooth' })

  let scrollEndTimer = 0
  let finished = false

  const finish = () => {
    if (finished) return
    finished = true
    window.removeEventListener('scroll', onScroll)
    window.clearTimeout(scrollEndTimer)

    const target = document.getElementById(id)
    if (target) {
      const top =
        target.getBoundingClientRect().top +
        window.scrollY -
        MARKDOWN_HEADING_SCROLL_OFFSET
      if (Math.abs(window.scrollY - top) > 2) {
        window.scrollTo({ top: Math.max(0, top), behavior: 'auto' })
      }
    }
    options?.onDone?.()
  }

  const onScroll = () => {
    window.clearTimeout(scrollEndTimer)
    scrollEndTimer = window.setTimeout(finish, 120)
  }

  window.addEventListener('scroll', onScroll, { passive: true })
  scrollEndTimer = window.setTimeout(finish, 1000)
}

/** @deprecated Prefer headings collected during Markdown render */
export function extractMarkdownHeadings(content: string): MarkdownHeading[] {
  const registry = createHeadingIdRegistry()
  let inFence = false

  for (const line of content.split('\n')) {
    const fence = line.match(/^(`{3,}|~{3,})/)
    if (fence) {
      inFence = !inFence
      continue
    }
    if (inFence) continue

    const match = line.match(/^(#{1,3})\s+(.+?)\s*$/)
    if (!match) continue

    const level = match[1].length
    const text = normalizeHeadingText(match[2])
    if (!text) continue

    registry.register(level, text)
  }

  return registry.getCollected()
}
