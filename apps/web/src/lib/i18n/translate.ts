import type { AppLocale } from '@/lib/locale'

export type MessageTree = { [key: string]: string | MessageTree }

export type MessageVars = Record<string, string | number>

function resolvePath(tree: MessageTree, key: string): string | undefined {
  const parts = key.split('.')
  let node: string | MessageTree | undefined = tree
  for (const part of parts) {
    if (node == null || typeof node === 'string') return undefined
    node = node[part]
  }
  return typeof node === 'string' ? node : undefined
}

function interpolate(template: string, vars?: MessageVars): string {
  if (!vars) return template
  return template.replace(/\{(\w+)\}/g, (_, name: string) => {
    const value = vars[name]
    return value === undefined || value === null ? '' : String(value)
  })
}

export function createTranslator(messages: MessageTree) {
  return function t(key: string, vars?: MessageVars): string {
    const resolved = resolvePath(messages, key)
    if (resolved === undefined) return key
    return interpolate(resolved, vars)
  }
}

export type Translator = ReturnType<typeof createTranslator>

export function localeIntlTag(locale: AppLocale): string {
  return locale === 'en' ? 'en-US' : 'zh-CN'
}
