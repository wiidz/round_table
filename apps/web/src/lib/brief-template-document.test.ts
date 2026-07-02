import { describe, expect, it } from 'vitest'

import {
  briefTemplateHasSubstantiveContent,
  emptyBriefDocument,
  normalizeBriefDocument,
} from '@/lib/brief-template-document'

describe('briefTemplateHasSubstantiveContent', () => {
  it('rejects title-only template', () => {
    const doc = emptyBriefDocument('仅名称')
    expect(briefTemplateHasSubstantiveContent(doc)).toBe(false)
  })

  it('accepts template with any brief field', () => {
    const doc = emptyBriefDocument('有内容')
    doc.brief.goal = '形成共识'
    expect(briefTemplateHasSubstantiveContent(doc)).toBe(true)
  })

  it('accepts template with partial meeting config', () => {
    const doc = emptyBriefDocument('有配置')
    doc.meeting = { mode: 'decision' }
    expect(briefTemplateHasSubstantiveContent(doc)).toBe(true)
  })

  it('normalize does not inject meeting defaults', () => {
    const normalized = normalizeBriefDocument(emptyBriefDocument('空模板'))
    expect(normalized.meeting).toBeUndefined()
    expect(normalized.brief.agenda).toEqual([])
  })
})
