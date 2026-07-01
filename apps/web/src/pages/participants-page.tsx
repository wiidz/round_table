import { useEffect, useState } from 'react'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'

import {
  createParticipant,
  deleteParticipant,
  fetchParticipants,
  updateParticipant,
} from '@/api/participants'
import { fetchSettings } from '@/api/settings'
import { ApiError } from '@/api/client'
import { ParticipantFormDialog } from '@/components/profile/participant-form-dialog'
import { PageLayout } from '@/components/layout/page-main-layout'
import {
  ParticipantGridCard,
  ParticipantGridSkeleton,
} from '@/components/profile/participant-grid-card'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { domainPageEyebrow, domainPageTitle } from '@/lib/ui-labels'
import { Button } from '@/components/ui/button'
import { hePressable } from '@/lib/highend-styles'
import { cn } from '@/lib/utils'

import type { ParticipantIndex, ParticipantRosterInput } from '@/types/participant'

export function ParticipantsPage() {
  const [participants, setParticipants] = useState<ParticipantIndex[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<'create' | 'edit'>('create')
  const [editing, setEditing] = useState<ParticipantIndex | null>(null)
  const [discordBots, setDiscordBots] = useState<{ id: string; label: string }[]>([])

  useEffect(() => {
    let cancelled = false
    fetchParticipants()
      .then((data) => {
        if (!cancelled) {
          setParticipants(data.participants ?? [])
          setError(null)
        }
      })
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载专家列表')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [])

  useEffect(() => {
    let cancelled = false
    fetchSettings()
      .then((data) => {
        if (cancelled) return
        setDiscordBots(
          (data.discord_bots ?? [])
            .filter((b) => b.deletable && b.configured)
            .map((b) => ({
              id: b.discord_application_id || b.id,
              label:
                b.display_name?.trim() ||
                b.discord_username?.trim() ||
                b.label?.trim() ||
                b.discord_application_id ||
                b.id,
            })),
        )
      })
      .catch(() => {
        if (!cancelled) setDiscordBots([])
      })
    return () => {
      cancelled = true
    }
  }, [])

  function openCreate() {
    setDialogMode('create')
    setEditing(null)
    setDialogOpen(true)
  }

  function openEdit(p: ParticipantIndex) {
    setDialogMode('edit')
    setEditing(p)
    setDialogOpen(true)
  }

  async function handleDelete(p: ParticipantIndex) {
    const name = p.display_name?.trim() || p.id
    if (
      !window.confirm(
        `确定删除专家「${name}」（${p.id}）？\n\n将移除专家配置与档案目录；不会影响已配置的 Discord Bot。`,
      )
    ) {
      return
    }
    try {
      const resp = await deleteParticipant(p.id)
      setParticipants(resp.participants ?? [])
      toast.success('已删除专家')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '删除失败')
    }
  }

  async function handleSubmit(input: ParticipantRosterInput) {
    if (dialogMode === 'create') {
      const resp = await createParticipant(input)
      setParticipants(resp.participants ?? [])
      toast.success('已添加专家')
      return
    }
    if (!editing) return
    const resp = await updateParticipant(editing.id, input)
    setParticipants(resp.participants ?? [])
    toast.success('已更新专家')
  }

  return (
    <PageLayout
      header={
        <div className="flex flex-wrap items-start justify-between gap-4">
          <ProfilePageHeader
            role="participant"
            eyebrow={domainPageEyebrow('participant')}
            title={domainPageTitle('participant')}
            description={
              <>
                管理会议专家（Participant）：代号、名称与 SOUL / AGENTS / TOOLS 档案。代号与名称均不可重复；修改代号会同步重命名档案目录。
              </>
            }
          />
          <Button
            type="button"
            onClick={openCreate}
            className={cn(hePressable, 'shrink-0 gap-2 rounded-xl px-4')}
          >
            <Plus className="size-4" />
            添加专家
          </Button>
        </div>
      }
    >
    <div className="space-y-8">
      {loading && <ParticipantGridSkeleton />}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && participants.length === 0 && (
        <ProfileStatePanel
          title="暂无专家档案"
          description={
            <>
              点击「添加专家」创建第一位专家（Participant），或在{' '}
              <code className="font-mono text-xs">data/profiles/participants/</code>{' '}
              下手动新建目录。
            </>
          }
        />
      )}

      {!loading && !error && participants.length > 0 && (
        <ul className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {participants.map((p) => (
            <li key={p.id} className="min-w-0">
              <ParticipantGridCard
                participant={p}
                onEdit={openEdit}
                onDelete={(target) => void handleDelete(target)}
              />
            </li>
          ))}
        </ul>
      )}

      <ParticipantFormDialog
        open={dialogOpen}
        mode={dialogMode}
        initial={editing}
        peers={participants}
        discordBots={discordBots}
        onClose={() => setDialogOpen(false)}
        onSubmit={handleSubmit}
      />
    </div>
    </PageLayout>
  )
}
