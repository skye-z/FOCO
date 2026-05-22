"use client"

import * as React from "react"
import Link from "next/link"
import { FlaskConical, LoaderCircle } from "lucide-react"
import { useRouter } from "next/navigation"
import { authFetch, readBrowserAccessToken, clearStoredSession } from "@/lib/auth-session"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"

type UnitSummary = {
  id?: string
  unit_version_id?: string
  version_id?: string
  title: string
  step_count: number
}

function unitTargetId(unit: UnitSummary) {
  return unit.id || unit.unit_version_id || unit.version_id || ""
}

export default function LabsPage() {
  const router = useRouter()
  const [units, setUnits] = React.useState<UnitSummary[]>([])
  const [loading, setLoading] = React.useState(true)

  React.useEffect(() => {
    const token = readBrowserAccessToken()
    if (!token) {
      clearStoredSession()
      router.replace("/")
      return
    }
    authFetch("/api/v1/learner/interactive-units", {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    })
      .then(async (res) => {
        if (res.ok) {
          try {
            const p = await res.json()
            setUnits(p.data ?? [])
          } catch {
            setUnits([])
          }
        }
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [router])

  if (loading) {
    return (
      <div className="flex min-h-[50vh] items-center justify-center">
        <LoaderCircle className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">交互学习单元</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          通过交互式步骤深入理解核心概念
        </p>
      </div>

      {units.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-dashed py-16 text-muted-foreground">
          <FlaskConical className="mb-3 size-10 opacity-40" />
          <p className="text-sm">暂无已发布的交互学习单元</p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {units.map((unit, index) => {
            const targetId = unitTargetId(unit)
            const card = (
              <Card className="transition-all hover:shadow-md hover:-translate-y-0.5">
                <CardContent className="p-5">
                  <div className="mb-3 flex items-center gap-2">
                    <Badge variant="outline" className="bg-teal-50 text-xs text-teal-700">
                      <FlaskConical className="mr-1 size-3" />
                      交互单元
                    </Badge>
                    <Badge variant="outline" className="text-xs text-muted-foreground">
                      {unit.step_count} 步
                    </Badge>
                  </div>
                  <h3 className="text-sm font-semibold leading-relaxed">{unit.title}</h3>
                  {!targetId && (
                    <p className="mt-2 text-xs text-red-500">targetId 为空，无法跳转</p>
                  )}
                </CardContent>
              </Card>
            )
            if (!targetId) {
              return <div key={`${unit.title}-${index}`}>{card}</div>
            }
            return (
              <Link
                key={targetId}
                href={`/labs/${targetId}`}
                className="block"
              >
                {card}
              </Link>
            )
          })}
        </div>
      )}
    </div>
  )
}
