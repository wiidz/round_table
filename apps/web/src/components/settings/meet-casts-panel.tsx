import { useEffect, useState } from 'react'
import { Plus, RotateCcw, Save, Trash2, Users } from 'lucide-react'
import { toast } from 'sonner'

import { fetchParticipants } from '@/api/participants'
import { resetMeetCasts, saveMeetCasts } from '@/api/settings'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useI18n } from '@/hooks/use-i18n'
import type { Translator } from '@/lib/i18n/translate'
import { hePressable, heSectionDesc, heSectionTitle } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'
import type { ParticipantIndex } from '@/types/participant'
import type { MeetCastConfig, SettingsResponse } from '@/types/settings'

type CastDraft = MeetCastConfig

function newCastDraft(index: number): CastDraft {
  const id = String(index)
  return { id, name_zh: '', name_en: '', participant_ids: [] }
}

function prepareCastsForSave(drafts: CastDraft[], roster: ParticipantIndex[]): MeetCastConfig[] {
  return drafts.map((d) => {
    const id = d.id.trim()
    let nameZh = d.name_zh.trim()
    let nameEn = d.name_en.trim()
    if (!nameZh && !nameEn) {
      const labels = d.participant_ids.map((pid) => {
        const p = roster.find((item) => item.id === pid)
        return (p?.display_name || p?.id || pid).trim()
      })
      nameZh = labels.filter(Boolean).join('+')
      nameEn = nameZh
    } else if (nameZh && !nameEn) {
      nameEn = nameZh
    } else if (!nameZh && nameEn) {
      nameZh = nameEn
    }
    return {
      id,
      name_zh: nameZh,
      name_en: nameEn,
      participant_ids: [...d.participant_ids],
    }
  })
}

function validateCasts(casts: MeetCastConfig[], t: Translator): string | null {
  for (const cast of casts) {
    if (!cast.id) return t('settings.meetCasts.errorEmptyId')
    if (!cast.name_zh && !cast.name_en) {
      return t('settings.meetCasts.errorMissingName', { id: cast.id })
    }
    if (cast.participant_ids.length === 0) {
      return t('settings.meetCasts.errorNoParticipants', { id: cast.id })
    }
  }
  const seen = new Set<string>()
  for (const cast of casts) {
    if (seen.has(cast.id)) return t('settings.meetCasts.errorDuplicateId', { id: cast.id })
    seen.add(cast.id)
  }
  return null
}

export function MeetCastsPanel({
  casts,
  onSaved,
}: {
  casts: MeetCastConfig[]
  onSaved: (resp: SettingsResponse) => void
}) {
  const { t } = useI18n()
  const [drafts, setDrafts] = useState<CastDraft[]>(() => casts.map((c) => ({ ...c, participant_ids: [...c.participant_ids] })))
  const [roster, setRoster] = useState<ParticipantIndex[]>([])
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    setDrafts(casts.map((c) => ({ ...c, participant_ids: [...c.participant_ids] })))
  }, [casts])

  useEffect(() => {
    fetchParticipants()
      .then((resp) => setRoster(resp.participants.filter((p) => p.in_roster !== false)))
      .catch(() => toast.error(t('settings.meetCasts.loadExpertsFailed')))
  }, [t])

  const updateDraft = (index: number, patch: Partial<CastDraft>) => {
    setDrafts((prev) =>
      prev.map((d, i) => {
        if (i !== index) return d
        const next = { ...d, ...patch }
        if ('name_zh' in patch && !('name_en' in patch)) {
          next.name_en = patch.name_zh ?? next.name_en
        }
        return next
      }),
    )
  }

  const toggleParticipant = (castIndex: number, participantId: string) => {
    setDrafts((prev) =>
      prev.map((d, i) => {
        if (i !== castIndex) return d
        const has = d.participant_ids.includes(participantId)
        return {
          ...d,
          participant_ids: has
            ? d.participant_ids.filter((id) => id !== participantId)
            : [...d.participant_ids, participantId],
        }
      }),
    )
  }

  const handleSave = async () => {
    const payload = prepareCastsForSave(drafts, roster)
    const err = validateCasts(payload, t)
    if (err) {
      toast.error(err)
      return
    }
    setSaving(true)
    try {
      const resp = await saveMeetCasts(payload)
      onSaved(resp)
      toast.success(t('settings.meetCasts.saveSuccess'))
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t('common.error.saveFailed'))
    } finally {
      setSaving(false)
    }
  }

  const handleReset = async () => {
    setSaving(true)
    try {
      const resp = await resetMeetCasts()
      onSaved(resp)
      toast.success(t('settings.meetCasts.clearSuccess'))
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t('settings.meetCasts.resetFailed'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <section className="space-y-4">
      <div>
        <h3 className={heSectionTitle}>{t('settings.meetCasts.title')}</h3>
        <p className={cn(heSectionDesc, 'mt-1')}>{t('settings.meetCasts.description')}</p>
      </div>

      <div className="space-y-4">
        {drafts.map((cast, index) => (
          <div
            key={`${cast.id}-${index}`}
            className="rounded-xl border border-black/[0.06] bg-canvas p-4 shadow-[var(--field-inset-shadow)]"
          >
            <div className="mb-3 flex flex-wrap items-center gap-3">
              <Users className="size-4 text-text-tertiary" />
              <Input
                value={cast.id}
                onChange={(e) => updateDraft(index, { id: e.target.value.trim() })}
                placeholder={t('settings.meetCasts.idPlaceholder')}
                className="h-8 w-24 font-mono text-xs"
              />
              <Input
                value={cast.name_zh}
                onChange={(e) => updateDraft(index, { name_zh: e.target.value, name_en: e.target.value })}
                placeholder={t('settings.meetCasts.namePlaceholder')}
                className="h-8 min-w-[12rem] flex-1 text-sm"
                required
              />
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="text-destructive"
                onClick={() => setDrafts((prev) => prev.filter((_, i) => i !== index))}
              >
                <Trash2 className="size-4" />
              </Button>
            </div>
            <div className="flex flex-wrap gap-2">
              {roster.map((p) => {
                const selected = cast.participant_ids.includes(p.id)
                const label = p.display_name || p.id
                return (
                  <button
                    key={p.id}
                    type="button"
                    onClick={() => toggleParticipant(index, p.id)}
                    className={cn(
                      'rounded-full border px-3 py-1 text-xs transition-colors',
                      selected
                        ? 'border-primary bg-primary/10 text-brand'
                        : 'border-black/[0.08] bg-black/[0.03] text-text-secondary hover:bg-black/[0.06]',
                    )}
                  >
                    {label}
                    <span className="ml-1 font-mono text-[10px] opacity-60">{p.id}</span>
                  </button>
                )
              })}
            </div>
          </div>
        ))}
      </div>

      <div className="flex flex-wrap gap-2">
        <Button
          type="button"
          variant="outline"
          className={cn(hePressable, 'gap-2')}
          onClick={() => setDrafts((prev) => [...prev, newCastDraft(prev.length + 1)])}
        >
          <Plus className="size-4" />
          {t('settings.meetCasts.add')}
        </Button>
        <Button type="button" disabled={saving} className={cn(hePressable, 'gap-2')} onClick={handleSave}>
          <Save className="size-4" />
          {saving ? t('common.saving') : t('settings.meetCasts.save')}
        </Button>
        <Button type="button" variant="ghost" disabled={saving} className="gap-2" onClick={handleReset}>
          <RotateCcw className="size-4" />
          {t('settings.meetCasts.clear')}
        </Button>
      </div>
    </section>
  )
}
