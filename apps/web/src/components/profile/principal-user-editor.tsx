import { useCallback, useEffect, useState } from 'react'
import { ArrowLeft, Save } from 'lucide-react'
import { Link } from 'react-router-dom'
import { toast } from 'sonner'

import { fetchPrincipal, savePrincipalUserProfile } from '@/api/principals'
import { ApiError } from '@/api/client'
import {
  ProfilePageHeader,
  ProfileStatePanel,
} from '@/components/profile/profile-page-header'
import { SettingsFieldRow } from '@/components/settings/field-hint-popover'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  heColumnTitleBrand,
  hePanelShell,
  hePressable,
  heSpring,
} from '@/lib/highend-styles'
import { domainNavLabel, domainPageEyebrow } from '@/lib/ui-labels'
import { cn } from '@/lib/utils'
import {
  EMPTY_PRINCIPAL_USER_PROFILE,
  type PrincipalUserProfile,
} from '@/types/principal'

interface PrincipalUserEditorProps {
  id: string
}

function profilesEqual(a: PrincipalUserProfile, b: PrincipalUserProfile): boolean {
  return (
    a.language === b.language &&
    (a.confirmation ?? '') === (b.confirmation ?? '') &&
    (a.context ?? '') === (b.context ?? '')
  )
}

export function PrincipalUserEditor({ id }: PrincipalUserEditorProps) {
  const [displayName, setDisplayName] = useState('')
  const [savedProfile, setSavedProfile] = useState<PrincipalUserProfile>(
    EMPTY_PRINCIPAL_USER_PROFILE,
  )
  const [form, setForm] = useState<PrincipalUserProfile>(EMPTY_PRINCIPAL_USER_PROFILE)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    const data = await fetchPrincipal(id)
    const profile = data.user_profile ?? EMPTY_PRINCIPAL_USER_PROFILE
    setDisplayName(data.display_name ?? '')
    setSavedProfile(profile)
    setForm(profile)
    setError(null)
  }, [id])

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    load()
      .catch((err: unknown) => {
        if (cancelled) return
        if (err instanceof ApiError) {
          setError(`请求失败 (${err.status})：${err.message}`)
        } else if (err instanceof Error) {
          setError(err.message)
        } else {
          setError('无法加载档案')
        }
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [load])

  const dirty = !profilesEqual(form, savedProfile)

  async function handleSave() {
    setSaving(true)
    try {
      const res = await savePrincipalUserProfile(id, form)
      setSavedProfile(res.user_profile)
      setForm(res.user_profile)
      toast.success('已保存偏好档案')
    } catch (err) {
      toast.error(err instanceof Error ? err.message : '保存失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="space-y-8">
      <Link
        to="/principals"
        className={cn(
          'inline-flex items-center gap-1.5 text-sm text-text-secondary',
          'hover:text-brand',
          heSpring,
        )}
      >
        <ArrowLeft className="size-4" />
        返回{domainNavLabel('principal')}列表
      </Link>

      <ProfilePageHeader
        role="principal"
        eyebrow={domainPageEyebrow('principal')}
        title={displayName || id}
        description={
          <>
            {displayName && (
              <span className="mb-1 block font-mono text-xs text-text-tertiary">{id}</span>
            )}
            填写偏好字段后由服务端生成 USER.md。会议议题请使用「简报模板」。
          </>
        }
      />

      {loading && (
        <ProfileStatePanel title="加载中" description="正在读取偏好档案…" />
      )}

      {!loading && error && (
        <ProfileStatePanel variant="danger" title="加载失败" description={error} />
      )}

      {!loading && !error && (
        <div className={cn(hePanelShell, 'flex flex-col gap-6 p-6 sm:p-8')}>
          <div className="space-y-6">
            <p className={heColumnTitleBrand}>偏好</p>
            <SettingsFieldRow
              label="语言"
              htmlFor="user-language"
              hint="Moderator 与 Reception 面向你的默认语言"
            >
              <Input
                id="user-language"
                value={form.language}
                placeholder="zh-CN"
                onChange={(e) => setForm((prev) => ({ ...prev, language: e.target.value }))}
              />
            </SettingsFieldRow>
            <SettingsFieldRow
              label="Confirmation 习惯"
              htmlFor="user-confirmation"
              hint="例如：逐项审阅编号清单、偏好简短摘要"
            >
              <Input
                id="user-confirmation"
                value={form.confirmation ?? ''}
                placeholder="review numbered lists carefully"
                onChange={(e) =>
                  setForm((prev) => ({ ...prev, confirmation: e.target.value }))
                }
              />
            </SettingsFieldRow>
            <SettingsFieldRow
              label="背景与约束"
              htmlFor="user-context"
              hint="行业、团队、项目约束；供 Moderator 长期理解你的语境"
            >
              <Textarea
                id="user-context"
                value={form.context ?? ''}
                rows={5}
                className="min-h-[8rem] font-sans text-sm"
                placeholder="例如：手游研发团队，关注可执行结论与上线风险"
                onChange={(e) => setForm((prev) => ({ ...prev, context: e.target.value }))}
              />
            </SettingsFieldRow>
          </div>

          <div className="flex flex-wrap items-center gap-3 border-t border-border-subtle/80 pt-4">
            <Button
              onClick={handleSave}
              disabled={!dirty || saving}
              className={cn(hePressable, 'gap-2 rounded-full px-5')}
            >
              <Save className="size-4" />
              {saving ? '保存中…' : '保存偏好'}
            </Button>
            {dirty && (
              <span className="text-xs font-medium text-warning">有未保存的修改</span>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
