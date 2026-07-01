import { useEffect, useMemo, useState } from 'react'

import { fetchParticipants } from '@/api/participants'
import { BriefMeetingExpertsList } from '@/components/brief/brief-meeting-experts-list'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import { BriefMeetingConfigRow } from '@/components/brief/brief-meeting-config-row'
import {
  BRIEF_MEETING_CONFIG_LABELS,
  BRIEF_TEMPLATE_SECTIONS,
  briefConfigPanelShell,
} from '@/components/brief/brief-template-sections'
import { normalizeBriefDocument } from '@/lib/brief-template-document'
import { cn } from '@/lib/utils'

import type { BriefTemplateDocument } from '@/types/brief-template'

export function BriefTemplateMeetingConfigPreview({
  document,
  className,
}: {
  document: BriefTemplateDocument
  className?: string
}) {
  const doc = normalizeBriefDocument(document)
  const expertIds = doc.meeting?.participant_ids?.filter(Boolean) ?? []
  const isDeliberation = doc.meeting?.mode === 'deliberation'

  const [participantNames, setParticipantNames] = useState<Map<string, string>>(new Map())

  useEffect(() => {
    let cancelled = false
    fetchParticipants()
      .then((res) => {
        if (cancelled) return
        const map = new Map<string, string>()
        for (const p of res.participants ?? []) {
          map.set(p.id, p.display_name?.trim() || p.id)
        }
        setParticipantNames(map)
      })
      .catch(() => {
        if (!cancelled) setParticipantNames(new Map())
      })
    return () => {
      cancelled = true
    }
  }, [])

  const experts = useMemo(
    () =>
      expertIds.map((id) => ({
        id,
        name: participantNames.get(id) ?? id,
      })),
    [expertIds, participantNames],
  )

  return (
    <aside className={cn('space-y-4 lg:sticky lg:top-20 lg:self-start', className)}>
      <BriefSectionHeading
        title={BRIEF_TEMPLATE_SECTIONS.meeting.title}
        description={BRIEF_TEMPLATE_SECTIONS.meeting.description}
      />
      <div className={briefConfigPanelShell}>
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.mode}
          value={isDeliberation ? '研讨型（deliberation）' : '裁决型（decision）'}
        />
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.confirmation}
          value={
            doc.meeting?.confirmation_mode === 'skip'
              ? '跳过 Principal 确认'
              : '需要 Principal 确认'
          }
        />
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.maxRounds}
          value={`${doc.meeting?.max_rounds ?? 3} 轮`}
        />
        {isDeliberation && (
          <BriefMeetingConfigRow
            label={BRIEF_MEETING_CONFIG_LABELS.minSynthesis}
            value={`第 ${doc.meeting?.min_rounds_before_synthesis ?? 2} 轮起`}
          />
        )}
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.freeDialogue}
          value={`${doc.meeting?.free_dialogue_max_questions ?? 1} 问`}
        />
        <BriefMeetingConfigRow
          label={BRIEF_MEETING_CONFIG_LABELS.experts}
          valueAlign="start"
          value={
            <BriefMeetingExpertsList
              experts={experts}
              emptyLabel="全部专家（默认）"
            />
          }
        />
      </div>
    </aside>
  )
}
