import { describe, expect, it } from 'vitest'

import { getTranslator } from '@/lib/i18n'
import { normalizeLocale, localeFromSettingsFields } from '@/lib/locale'

describe('normalizeLocale', () => {
  it('maps en variants to en', () => {
    expect(normalizeLocale('en')).toBe('en')
    expect(normalizeLocale('en-US')).toBe('en')
  })

  it('defaults to zh', () => {
    expect(normalizeLocale('zh')).toBe('zh')
    expect(normalizeLocale('zh-CN')).toBe('zh')
    expect(normalizeLocale('')).toBe('zh')
  })
})

describe('localeFromSettingsFields', () => {
  it('reads ROUND_TABLE_LOCALE from settings fields', () => {
    expect(
      localeFromSettingsFields([
        {
          key: 'ROUND_TABLE_LOCALE',
          value: 'en',
          label: '',
          group: '',
          configured: true,
          secret: false,
          editable: true,
        },
      ]),
    ).toBe('en')
  })
})

describe('getTranslator', () => {
  it('returns single-language titles', () => {
    const tZh = getTranslator('zh')
    const tEn = getTranslator('en')
    expect(tZh('domain.participant')).toBe('专家')
    expect(tEn('domain.participant')).toBe('Participant')
    expect(tZh('brief.pageTitle')).toBe('简报模板')
    expect(tEn('brief.pageTitle')).toBe('Brief Templates')
  })

  it('interpolates variables', () => {
    const t = getTranslator('en')
    expect(t('common.error.requestFailed', { status: 404, message: 'not found' })).toBe(
      'Request failed (404): not found',
    )
  })
})
