import { useEffect, useMemo, useState } from 'react'

import { fetchParticipants } from '@/api/participants'
import { BriefMeetingExpertsList } from '@/components/brief/brief-meeting-experts-list'
import {
  SearchableSelect,
  type SearchableSelectOption,
} from '@/components/settings/searchable-select'

interface ParticipantMultiSelectProps {
  id?: string
  value: string[]
  disabled?: boolean
  onChange: (ids: string[]) => void
}

export function ParticipantMultiSelect({
  id,
  value,
  disabled,
  onChange,
}: ParticipantMultiSelectProps) {
  const [options, setOptions] = useState<SearchableSelectOption[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    fetchParticipants()
      .then((res) => {
        if (cancelled) return
        setOptions(
          (res.participants ?? []).map((p) => ({
            value: p.id,
            label: p.display_name?.trim() || p.id,
            hint: [p.id, p.expertise].filter(Boolean).join(' · '),
          })),
        )
      })
      .catch(() => {
        if (!cancelled) setOptions([])
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  const placeholder = useMemo(() => {
    if (loading) return '加载专家列表…'
    if (value.length === 0) return '搜索并选择专家（留空则默认全部专家）'
    return '继续搜索添加专家'
  }, [loading, value.length])

  const selectedExperts = useMemo(
    () =>
      value.map((participantId) => {
        const option = options.find((o) => o.value === participantId)
        return {
          id: participantId,
          name: option?.label ?? participantId,
        }
      }),
    [options, value],
  )

  return (
    <div className="space-y-2.5">
      {selectedExperts.length > 0 && (
        <BriefMeetingExpertsList
          experts={selectedExperts}
          removable
          disabled={disabled || loading}
          onRemove={(participantId) =>
            onChange(value.filter((current) => current !== participantId))
          }
        />
      )}
      <SearchableSelect
        id={id}
        multiple
        disabled={disabled || loading}
        value={value}
        options={options}
        placeholder={placeholder}
        searchPlaceholder="搜索专家名称、代号或专长"
        onChange={onChange}
        hideSelectedChips
      />
    </div>
  )
}
