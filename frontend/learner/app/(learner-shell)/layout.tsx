"use client"

import * as React from "react"
import { useRouter, usePathname } from "next/navigation"
import Link from "next/link"
import {
  Home,
  Dumbbell,
  BookX,
  FlaskConical,
  LogOut,
  User,
  Menu,
  X,
} from "lucide-react"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import {
  readBrowserAccessToken,
  clearStoredSession,
  readStoredSession,
} from "@/lib/auth-session"
import { cn } from "@/lib/utils"

const NAV_ITEMS = [
  { label: "首页", href: "/home", icon: Home },
  { label: "练习", href: "/practice/setup", icon: Dumbbell },
  { label: "学习", href: "/labs", icon: FlaskConical },
  { label: "错题本", href: "/wrong-book", icon: BookX },
] as const

function isActive(pathname: string, href: string) {
  if (href === "/home") return pathname === "/home" || pathname === "/"
  return pathname.startsWith(href)
}

export default function LearnerShellLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const router = useRouter()
  const pathname = usePathname()
  const [checked, setChecked] = React.useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false)

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

  const session = readStoredSession()
  const displayName = session?.user?.email?.split("@")[0] ?? "学习者"

  const sidebar = (
    <aside className="flex h-full w-64 flex-col bg-sidebar text-sidebar-foreground">
      <div className="flex items-center gap-2.5 px-6 py-5">
        <span className="font-heading text-xl font-bold tracking-tight text-sidebar-accent-foreground">
          FOCO
        </span>
      </div>

      <Separator className="mx-4 w-auto bg-sidebar-border" />

      <nav className="flex flex-1 flex-col gap-1 px-3 pt-4">
        {NAV_ITEMS.map((item) => {
          const active = isActive(pathname, item.href)
          return (
            <Link
              key={item.href}
              href={item.href}
              onClick={() => setMobileMenuOpen(false)}
            >
              <span
                className={cn(
                  "flex items-center gap-3 rounded-full px-4 py-2.5 text-sm font-medium transition-colors",
                  active
                    ? "bg-sidebar-primary text-sidebar-primary-foreground"
                    : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                )}
              >
                <item.icon className="size-[18px]" />
                {item.label}
              </span>
            </Link>
          )
        })}
      </nav>

      <Separator className="mx-4 w-auto bg-sidebar-border" />

      <div className="flex items-center gap-2 px-5 py-5">
        <Link
          href="/profile"
          onClick={() => setMobileMenuOpen(false)}
          className={cn(
            "flex min-w-0 flex-1 items-center gap-3 rounded-full px-2 py-1.5 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
            pathname.startsWith("/profile") && "bg-sidebar-accent text-sidebar-accent-foreground"
          )}
        >
          <div className="flex size-9 items-center justify-center rounded-full bg-sidebar-accent text-sidebar-accent-foreground">
            <User className="size-4" />
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-sidebar-accent-foreground">
              {displayName}
            </p>
            <p className="truncate text-xs text-sidebar-foreground/70">用户中心</p>
          </div>
        </Link>
        <button
          onClick={() => {
            clearStoredSession()
            router.replace("/")
          }}
          className="flex size-8 items-center justify-center rounded-full text-sidebar-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
          aria-label="退出登录"
        >
          <LogOut className="size-4" />
        </button>
      </div>
    </aside>
  )

  return (
    <div className="flex min-h-screen bg-[var(--surface)]">
      {/* Desktop sidebar */}
      <div className="fixed inset-y-0 left-0 z-30 hidden md:flex">{sidebar}</div>

      {/* Mobile overlay */}
      {mobileMenuOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 md:hidden"
          onClick={() => setMobileMenuOpen(false)}
        />
      )}

      {/* Mobile drawer */}
      <div
        className={cn(
          "fixed inset-y-0 left-0 z-50 transition-transform duration-200 md:hidden",
          mobileMenuOpen ? "translate-x-0" : "-translate-x-full"
        )}
      >
        {sidebar}
      </div>

      {/* Main content */}
      <div className="flex min-h-screen flex-1 flex-col md:pl-64">
        {/* Mobile header bar */}
        <header className="sticky top-0 z-20 flex h-12 items-center gap-3 border-b border-[var(--outline-variant)] bg-[var(--surface-container-lowest)] px-4 md:hidden">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setMobileMenuOpen(true)}
            aria-label="打开菜单"
          >
            <Menu className="size-5" />
          </Button>
          <span className="font-heading text-lg font-bold tracking-tight text-[var(--primary)]">
            FOCO
          </span>
        </header>

        <main className="flex-1 px-4 py-6 md:px-6">{children}</main>
      </div>
    </div>
  )
}
