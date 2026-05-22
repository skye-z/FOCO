"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import { cn } from "@/lib/utils"
import { authFetch, readBrowserAccessToken } from "@/lib/auth-session"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { AlertCircle, BookX, Coins, RotateCcw, Star } from "lucide-react"

const API_BASE = "/api/v1"

type SessionSummary = {
  total: number
  correct: number
  wrong: number
  accuracy: number
  xp_earned: number
  coins_earned: number
  duration_minutes: number
}

const DEMO_SUMMARY: SessionSummary = {
  total: 20,
  correct: 17,
  wrong: 3,
  accuracy: 85,
  xp_earned: 120,
  coins_earned: 50,
  duration_minutes: 25,
}

function AccuracyRing({ value, size = 140 }: { value: number; size?: number }) {
  const strokeWidth = 10
  const radius = (size - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (value / 100) * circumference
  const center = size / 2

  return (
    <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={center}
          cy={center}
          r={radius}
          fill="none"
          stroke="var(--surface-container-high)"
          strokeWidth={strokeWidth}
        />
        <circle
          cx={center}
          cy={center}
          r={radius}
          fill="none"
          stroke="var(--secondary-green)"
          strokeWidth={strokeWidth}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          className="transition-[stroke-dashoffset] duration-700 ease-out"
        />
      </svg>
      <div className="absolute flex flex-col items-center">
        <span className="font-heading text-3xl font-bold" style={{ fontFamily: "var(--font-heading)" }}>
          {value}%
        </span>
        <span className="text-xs text-muted-foreground">正确率</span>
      </div>
    </div>
  )
}

function StatCard({
  icon: Icon,
  label,
  value,
  color,
}: {
  icon: React.ElementType
  label: string
  value: string | number
  color: string
}) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-xl border border-border/50 bg-muted/30 p-4">
      <div
        className="flex size-10 items-center justify-center rounded-full"
        style={{ backgroundColor: color + "18", color }}
      >
        <Icon className="size-5" />
      </div>
      <span className="font-heading text-2xl font-bold" style={{ fontFamily: "var(--font-heading)" }}>
        {value}
      </span>
      <span className="text-xs text-muted-foreground">{label}</span>
    </div>
  )
}

export default function PracticeCompletePage({
  params,
}: {
  params: Promise<{ sessionId: string }>
}) {
  const router = useRouter()
  const [summary, setSummary] = useState<SessionSummary>(DEMO_SUMMARY)

  useEffect(() => {
    async function load() {
      const { sessionId } = await params
      const token = readBrowserAccessToken()
      if (!token) return

      try {
        const res = await authFetch(
          `${API_BASE}/learner/practice-sessions/${sessionId}/summary`,
          { headers: { Authorization: `Bearer ${token}` } }
        )
        if (res.ok) {
          const payload = await res.json()
          setSummary(payload.data ?? payload)
        }
      } catch {}
    }
    load()
  }, [params])

  return (
    <div className="relative flex min-h-screen flex-col items-center justify-center overflow-hidden px-4 py-12">
      <div
        className="pointer-events-none absolute -top-40 left-1/2 h-[500px] w-[600px] -translate-x-1/2 rounded-full opacity-30"
        style={{
          background:
            "radial-gradient(ellipse at center, var(--secondary-fixed) 0%, transparent 70%)",
        }}
      />
      <div
        className="pointer-events-none absolute -bottom-32 right-0 h-[300px] w-[400px] rounded-full opacity-20"
        style={{
          background:
            "radial-gradient(ellipse at center, var(--secondary-fixed-dim) 0%, transparent 70%)",
        }}
      />

      <div className="relative z-10 flex w-full max-w-lg flex-col items-center gap-8">
        <div className="flex flex-col items-center gap-2 text-center">
          <div className="flex size-16 items-center justify-center rounded-full bg-secondary-green/15">
            <Star className="size-8 text-secondary-green" />
          </div>
          <h1
            className="font-heading text-3xl font-bold tracking-tight"
            style={{ fontFamily: "var(--font-heading)" }}
          >
            练习完成！
          </h1>
          <p className="text-muted-foreground">
            你已完成本次练习，用时 {summary.duration_minutes} 分钟
          </p>
        </div>

        <Card className="w-full backdrop-blur-md">
          <CardContent className="flex flex-col items-center gap-6 py-6">
            <AccuracyRing value={summary.accuracy} />

            <div className="grid w-full grid-cols-3 gap-3">
              <StatCard
                icon={AlertCircle}
                label="错题数"
                value={summary.wrong}
                color="var(--error)"
              />
              <StatCard
                icon={Star}
                label="获得经验"
                value={`+${summary.xp_earned}`}
                color="var(--secondary-green)"
              />
              <StatCard
                icon={Coins}
                label="获得金币"
                value={`+${summary.coins_earned}`}
                color="#e6a817"
              />
            </div>
          </CardContent>
        </Card>

        <div className="flex w-full flex-col gap-3 sm:flex-row">
          <Button variant="outline" className="flex-1" onClick={() => router.push("/wrong-book")}>
            <BookX className="size-4" />
            查看错题
          </Button>
          <Button variant="secondary" className="flex-1" onClick={() => router.push("/home")}>
            返回首页
          </Button>
          <Button className="flex-1" onClick={() => router.push("/practice/setup")}>
            <RotateCcw className="size-4" />
            再来一轮
          </Button>
        </div>
      </div>
    </div>
  )
}
