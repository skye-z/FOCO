"use client"

import * as React from "react"
import { LogOut } from "lucide-react"
import { usePathname, useRouter } from "next/navigation"
import { clearStoredSession, readBrowserAccessToken } from "@/lib/auth-session"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"

const NAV_ITEMS = [
  { label: "概览", href: "/home" },
  { label: "题库管理", href: "/exams" },
  { label: "用户管理", href: "/users" },
  { label: "设置", href: "/settings" },
] as const

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const pathname = usePathname()
  const router = useRouter()
  const [ready, setReady] = React.useState(false)

  React.useEffect(() => {
    const token = readBrowserAccessToken()
    if (!token) {
      clearStoredSession()
      router.replace("/")
      return
    }
    setReady(true)
  }, [router])

  function handleLogout() {
    clearStoredSession()
    router.replace("/")
  }

  if (!ready) return null

  return (
    <div className="min-h-screen bg-muted/30">
      <header className="sticky top-0 z-40 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="flex h-14 items-center justify-between px-6">
          <div className="flex items-center gap-3">
            <span className="font-heading text-xl font-bold tracking-tight text-primary">
              FOCO
            </span>
            <Badge variant="secondary" className="text-xs">
              管理后台
            </Badge>
          </div>

          <nav className="flex items-center gap-1">
            {NAV_ITEMS.map((item) => {
              const active = pathname === item.href
              return (
                <Button
                  key={item.href}
                  variant="ghost"
                  size="sm"
                  onClick={() => router.push(item.href)}
                  className={
                    active
                      ? "bg-secondary font-medium text-secondary-foreground"
                      : "text-muted-foreground hover:text-foreground"
                  }
                >
                  {item.label}
                </Button>
              )
            })}
          </nav>

          <div className="flex items-center gap-2">
            <Separator orientation="vertical" className="h-6" />
            <button
              onClick={handleLogout}
              className="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            >
              <LogOut className="size-4" />
              退出登录
            </button>
          </div>
        </div>
      </header>
      {children}
    </div>
  )
}
