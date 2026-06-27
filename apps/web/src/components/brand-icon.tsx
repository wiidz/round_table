/** Platform brand marks — SVG assets in public/brands/ (Simple Icons + Wikimedia Slack). */

const BRAND_ASSETS: Record<string, string> = {
  discord: '/brands/discord.svg',
  telegram: '/brands/telegram.svg',
  slack: '/brands/slack.svg',
  feishu: '/brands/feishu.svg',
  whatsapp: '/brands/whatsapp.svg',
  lark: '/brands/lark.svg',
  imessage: '/brands/imessage.svg',
  deepseek: '/brands/deepseek.svg',
  openai: '/brands/openai.svg',
  anthropic: '/brands/anthropic.svg',
  qwen: '/brands/qwen.svg',
  gemini: '/brands/gemini.svg',
  grok: '/brands/grok.svg',
  ollama: '/brands/ollama.svg',
}

export function hasBrandIcon(id: string) {
  return id in BRAND_ASSETS
}

export function BrandIcon({
  id,
  className = 'size-7',
}: {
  id: string
  className?: string
}) {
  const src = BRAND_ASSETS[id]
  if (!src) return null
  return (
    <img
      src={src}
      alt=""
      draggable={false}
      className={className + ' object-contain'}
    />
  )
}
