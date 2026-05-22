"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { readBrowserAccessToken, clearStoredSession } from "@/lib/auth-session"

export default function FocusLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const [checked, setChecked] = React.useState(false)

  React.useEffect(() => {
    const token = readBrowserAccessToken()
    if (!token) {
      clearStoredSession()
      router.replace("/")
      return
    }
    setChecked(true)
  }, [router])

  if (!checked) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-[var(--surface)]">
        <span className="text-sm text-[var(--text-muted)]">正在验证登录状态...</span>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[var(--surface)]">
      {children}
    </div>
  )
}
