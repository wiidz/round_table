import { Link } from 'react-router-dom'
import { Sparkles } from 'lucide-react'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export function HomePage() {
  return (
    <div className="space-y-8">
      <section className="space-y-3">
        <p className="text-xs font-medium uppercase tracking-[0.18em] text-text-tertiary">
          Build AI Teams, not AI Agents
        </p>
        <h1 className="max-w-2xl text-[28px] font-semibold leading-tight tracking-tight">
          多智能体会议引擎 · Principal 工作台
        </h1>
        <p className="max-w-2xl text-[15px] leading-relaxed text-text-secondary">
          在 Web 端审阅 Confirmation Brief、阅读 Minutes 与 Artifacts，与 Discord
          入口共享同一 Meeting Engine。
        </p>
      </section>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>会议列表</CardTitle>
            <CardDescription>查看历史 Meeting 与当前状态</CardDescription>
          </CardHeader>
          <CardContent>
            <Button asChild variant="outline">
              <Link to="/meetings">进入会议</Link>
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>确认关</CardTitle>
            <CardDescription>逐项审阅 Brief、批准或驳回</CardDescription>
          </CardHeader>
          <CardContent>
            <Button disabled variant="outline">
              待 API 联调
            </Button>
          </CardContent>
        </Card>

        <Card className="border-ai/20 ds-ai-surface">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Sparkles className="size-4 text-ai" />
              AI 产出
            </CardTitle>
            <CardDescription>
              Participant / Moderator 发言与纪要统一用 AI 紫标识
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button variant="ai" disabled>
              阅读 Artifacts
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
