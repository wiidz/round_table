import { useEffect, useMemo, useState } from 'react'

import { fetchParticipants } from '@/api/participants'
import { BriefMeetingExpertsList } from '@/components/brief/brief-meeting-experts-list'
import { BriefSectionHeading } from '@/components/brief/brief-section-heading'
import { BriefMeetingConfigRow } from '@/components/brief/brief-meeting-config-row'
import { briefConfigPanelShell } from '@/components/brief/brief-template-sections'
import { useI18n } from '@/hooks/use-i18n'
import { getBriefMeetingConfigLabels, getBriefSections } from '@/lib/i18n/brief-sections'
import { meetingFieldIsSet, normalizeBriefDocument } from '@/lib/brief-template-document'
import { cn } from '@/lib/utils'

import type { BriefTemplateDocument } from '@/types/brief-template'

export function BriefTemplateMeetingConfigPreview({
  document,
  className,
}: {
  document: BriefTemplateDocument
  className?: string
}) {
  const { t, locale } = useI18n()
  const sections = getBriefSections(locale)
  const configLabels = getBriefMeetingConfigLabels(locale)
  const doc = normalizeBriefDocument(document)
  const meeting = doc.meeting
  const defer = t('brief.config.deferToMeeting')
  const expertIds = meeting?.participant_ids?.filter(Boolean) ?? []
  const isDeliberation = meeting?.mode === 'deliberation'

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
        title={sections.meeting.title}
        description={sections.meeting.description}
      />
      <div className={briefConfigPanelShell}>
        <BriefMeetingConfigRow
          label={configLabels.mode}
          value={
            !meetingFieldIsSet(meeting, 'mode')
              ? defer
              : isDeliberation
                ? t('brief.config.modeDeliberation')
                : t('brief.config.modeDecision')
          }
        />
        <BriefMeetingConfigRow
          label={configLabels.confirmation}
          value={
            !meetingFieldIsSet(meeting, 'confirmation_mode')
              ? defer
              : meeting?.confirmation_mode === 'skip'
                ? t('brief.config.confirmationSkipPrincipal')
                : t('brief.config.confirmationRequired')
          }
        />
        <BriefMeetingConfigRow
          label={configLabels.maxRounds}
          value={
            !meetingFieldIsSet(meeting, 'max_rounds')
              ? defer
              : t('brief.config.maxRoundsValue', { n: meeting!.max_rounds! })
          }
        />
        {(isDeliberation || meetingFieldIsSet(meeting, 'min_rounds_before_synthesis')) && (
          <BriefMeetingConfigRow
            label={configLabels.minSynthesis}
            value={
              !meetingFieldIsSet(meeting, 'min_rounds_before_synthesis')
                ? defer
                : t('brief.config.minSynthesisFrom', {
                    n: meeting!.min_rounds_before_synthesis!,
                  })
            }
          />
        )}
        <BriefMeetingConfigRow
          label={configLabels.freeDialogue}
          value={
            !meetingFieldIsSet(meeting, 'free_dialogue_max_questions')
              ? defer
              : t('brief.config.freeDialogueQuestions', {
                  n: meeting!.free_dialogue_max_questions!,
                })
          }
        />
        <BriefMeetingConfigRow
          label={configLabels.experts}
          valueAlign="start"
          value={
            !meetingFieldIsSet(meeting, 'participant_ids') ? (
              defer
            ) : (
              <BriefMeetingExpertsList experts={experts} emptyLabel={defer} />
            )
          }
        />
      </div>
    </aside>
  )
}
