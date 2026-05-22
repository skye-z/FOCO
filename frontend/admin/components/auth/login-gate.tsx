"use client"

import * as React from "react"
import {
  ArrowRight,
  Eye,
  EyeOff,
  LoaderCircle,
  Lock,
  UserRound,
} from "lucide-react"
import {
  clearStoredSession,
  readStoredSession,
  readRememberedEmail,
  writeRememberedEmail,
  type StoredAuthSession,
  writeStoredSession,
} from "@/lib/auth-session"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

type LoginGateProps = {
  supabaseUrl: string
  publishableKey: string
}

type TokenResponse = {
  access_token: string
  refresh_token?: string
  expires_at?: number
  session?: {
    access_token: string
    refresh_token?: string
    expires_at?: number
  } | null
  user?: {
    id?: string
    email?: string
  }
}

type GateState = "checking" | "ready" | "submitting"

const copy = {
  checking: "正在检查登录状态",
  title: "登录管理后台",
  usernameLabel: "邮箱",
  usernamePlaceholder: "请输入登录邮箱",
  passwordLabel: "密码",
  passwordPlaceholder: "请输入密码",
  submit: "登录",
  submitting: "登录中",
  loginFailed: "登录失败，请检查邮箱或密码。",
  loginRequestFailed: "登录请求失败，请稍后重试。",
  roleForbidden: "当前账户没有管理端访问权限，请联系管理员。",
  missingAuthConfig: "缺少 Supabase 认证配置，请检查环境变量。",
  hidePassword: "隐藏密码",
  showPassword: "显示密码",
} as const

export function LoginGate({ supabaseUrl, publishableKey }: LoginGateProps) {
  const router = useRouter()
  const authReady = Boolean(supabaseUrl && publishableKey)
  const [state, setState] = React.useState<GateState>("checking")
  const [email, setEmail] = React.useState("")
  const [password, setPassword] = React.useState("")
  const [error, setError] = React.useState("")
  const [passwordVisible, setPasswordVisible] = React.useState(false)

  React.useEffect(() => {
    const remembered = readRememberedEmail()
    if (remembered) setEmail(remembered)
  }, [])

  React.useEffect(() => {
    let cancelled = false

    async function check() {
      if (!authReady) {
        if (!cancelled) {
          setError(copy.missingAuthConfig)
          setState("ready")
        }
        return
      }

      const session = readStoredSession()
      if (!session) {
        if (!cancelled) setState("ready")
        return
      }

      const valid = await validateSession(supabaseUrl, publishableKey, session)
      if (cancelled) return

      if (valid) {
        const me = await fetchMe(valid.accessToken)
        if (cancelled) return

        if (me && (me.user.roles.includes("admin") || me.user.roles.includes("editor"))) {
          writeStoredSession(valid)
          router.replace("/home")
          return
        }

        clearStoredSession()
        if (me) setError(copy.roleForbidden)
      }

      if (!cancelled) setState("ready")
    }

    void check()
    return () => { cancelled = true }
  }, [authReady, publishableKey, supabaseUrl])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setState("submitting")
    setError("")

    if (!authReady) {
      setError(copy.missingAuthConfig)
      setState("ready")
      return
    }

    try {
      const session = await signInWithPassword({ supabaseUrl, publishableKey, email, password })
      if (!session) {
        setError(copy.loginFailed)
        setState("ready")
        return
      }

      const me = await fetchMe(session.accessToken)
      if (!me) {
        setError(copy.loginRequestFailed)
        setState("ready")
        return
      }

      if (!me.user.roles.includes("admin") && !me.user.roles.includes("editor")) {
        setError(copy.roleForbidden)
        setState("ready")
        return
      }

      writeStoredSession(session)
      writeRememberedEmail(email)
      router.replace("/home")
      return
    } catch {
      setError(copy.loginRequestFailed)
      setState("ready")
    }
  }

  if (state === "checking") {
    return (
      <main className="flex min-h-screen items-center justify-center bg-background px-6">
        <div className="grid place-items-center gap-4 text-center">
          <div className="font-heading text-5xl font-bold tracking-tight text-primary">
            FOCO
          </div>
          <div className="flex items-center gap-3 text-muted-foreground">
            <LoaderCircle className="size-5 animate-spin" />
            <span className="text-sm font-medium">{copy.checking}</span>
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="flex min-h-screen items-center justify-center bg-background px-6">
      <Card className="w-full max-w-[480px] rounded-3xl shadow-lg">
        <CardContent className="p-8 md:p-12">
          <div className="mb-10 text-center">
            <div className="mb-2 font-heading text-5xl font-bold tracking-tight text-primary">
              FOCO
            </div>
            <p className="text-base text-muted-foreground">
              {copy.title}
            </p>
          </div>

          <form className="flex flex-col gap-6" onSubmit={handleSubmit}>
            <div className="flex flex-col gap-2">
              <Label htmlFor="username">{copy.usernameLabel}</Label>
              <div className="relative">
                <UserRound className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  id="username"
                  type="email"
                  autoComplete="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder={copy.usernamePlaceholder}
                  className="pl-10"
                  required
                />
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="password">{copy.passwordLabel}</Label>
              </div>
              <div className="relative">
                <Lock className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  id="password"
                  type={passwordVisible ? "text" : "password"}
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder={copy.passwordPlaceholder}
                  className="pl-10 pr-10"
                  required
                />
                <button
                  type="button"
                  onClick={() => setPasswordVisible((v) => !v)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  aria-label={passwordVisible ? copy.hidePassword : copy.showPassword}
                >
                  {passwordVisible ? <Eye className="size-4" /> : <EyeOff className="size-4" />}
                </button>
              </div>
            </div>

            {error ? <p className="text-sm text-destructive">{error}</p> : null}

            <Button
              type="submit"
              disabled={state === "submitting"}
              className="h-12 rounded-full bg-secondary font-semibold text-secondary-foreground hover:bg-primary"
            >
              {state === "submitting" ? (
                <>
                  <LoaderCircle className="size-4 animate-spin" />
                  {copy.submitting}
                </>
              ) : (
                <>
                  {copy.submit}
                  <ArrowRight className="size-4" />
                </>
              )}
            </Button>
          </form>
        </CardContent>
      </Card>
    </main>
  )
}

async function validateSession(supabaseUrl: string, publishableKey: string, session: StoredAuthSession): Promise<StoredAuthSession | null> {
  const resp = await fetch(`${supabaseUrl}/auth/v1/user`, {
    headers: { apikey: publishableKey, Authorization: `Bearer ${session.accessToken}` },
  })
  if (resp.ok) return session
  if (!session.refreshToken) return null

  const refresh = await fetch(`${supabaseUrl}/auth/v1/token?grant_type=refresh_token`, {
    method: "POST",
    headers: { apikey: publishableKey, "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: session.refreshToken }),
  })
  if (!refresh.ok) return null
  return normalizeSession((await refresh.json()) as TokenResponse)
}

async function signInWithPassword(input: { supabaseUrl: string; publishableKey: string; email: string; password: string }): Promise<StoredAuthSession | null> {
  const resp = await fetch(`${input.supabaseUrl}/auth/v1/token?grant_type=password`, {
    method: "POST",
    headers: { apikey: input.publishableKey, "Content-Type": "application/json" },
    body: JSON.stringify({ email: input.email, password: input.password }),
  })
  if (!resp.ok) return null
  return normalizeSession((await resp.json()) as TokenResponse)
}

async function fetchMe(token: string) {
  try {
    const resp = await fetch("/api/v1/me", {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
    })
    if (!resp.ok) return null
    const payload = (await resp.json()) as {
      data: { user: { id: string; email: string; roles: string[] }; active_exam_enrollment: null | Record<string, string> }
    }
    return payload.data
  } catch {
    return null
  }
}

function normalizeSession(payload: TokenResponse): StoredAuthSession {
  return {
    accessToken: payload.access_token,
    refreshToken: payload.refresh_token,
    expiresAt: payload.expires_at,
    user: payload.user ? { id: payload.user.id, email: payload.user.email } : undefined,
  }
}
