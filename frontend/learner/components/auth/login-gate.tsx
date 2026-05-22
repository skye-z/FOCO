"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
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
  mergeSessionWithActiveEnrollment,
  readStoredSession,
  type StoredAuthSession,
  writeStoredSession,
} from "@/lib/auth-session"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { toast } from "sonner"

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
type FormMode = "login" | "register"

const copy = {
  checking: "\u6b63\u5728\u68c0\u67e5\u767b\u5f55\u72b6\u6001",
  title: "\u767b\u5f55\u60a8\u7684\u8d26\u6237\uff0c\u7ee7\u7eed\u5b66\u4e60\u4e4b\u65c5",
  registerTitle: "\u521b\u5efa learner \u8d26\u6237\uff0c\u5f00\u59cb\u5b66\u4e60\u4e4b\u65c5",
  displayNameLabel: "\u6635\u79f0",
  displayNamePlaceholder: "\u8bf7\u8f93\u5165\u6635\u79f0",
  usernameLabel: "\u90ae\u7bb1",
  usernamePlaceholder: "\u8bf7\u8f93\u5165\u767b\u5f55\u90ae\u7bb1",
  passwordLabel: "\u5bc6\u7801",
  passwordPlaceholder: "\u8bf7\u8f93\u5165\u5bc6\u7801",
  forgotPassword: "\u5fd8\u8bb0\u5bc6\u7801\uff1f",
  submit: "\u767b\u5f55",
  submitting: "\u767b\u5f55\u4e2d",
  noAccount: "\u8fd8\u6ca1\u6709\u8d26\u6237\uff1f",
  createAccount: "\u521b\u5efa\u8d26\u6237",
  backToLogin: "\u8fd4\u56de\u767b\u5f55",
  loginFailed: "\u767b\u5f55\u5931\u8d25\uff0c\u8bf7\u68c0\u67e5\u90ae\u7bb1\u6216\u5bc6\u7801\u3002",
  loginRequestFailed: "\u767b\u5f55\u8bf7\u6c42\u5931\u8d25\uff0c\u8bf7\u7a0d\u540e\u91cd\u8bd5\u3002",
  signupFailed: "\u6ce8\u518c\u5931\u8d25\uff0c\u8bf7\u68c0\u67e5\u8f93\u5165\u6216\u7a0d\u540e\u91cd\u8bd5\u3002",
  signupClosed: "\u5f53\u524d\u6682\u672a\u5f00\u653e\u6ce8\u518c\uff0c\u8bf7\u8054\u7cfb\u7ba1\u7406\u5458\u5f00\u901a\u8d26\u53f7\u3002",
  signupSuccessVerifyEmail: "\u6ce8\u518c\u6210\u529f\uff0c\u8bf7\u5148\u524d\u5f80\u90ae\u7bb1\u5b8c\u6210\u9a8c\u8bc1\uff0c\u518d\u8fd4\u56de\u767b\u5f55\u3002",
  bootstrapFailed: "\u8d26\u6237\u521b\u5efa\u6210\u529f\uff0c\u4f46 learner \u521d\u59cb\u5316\u5931\u8d25\uff0c\u8bf7\u91cd\u8bd5\u3002",
  roleForbidden: "\u5f53\u524d\u8d26\u6237\u6ca1\u6709 learner \u8bbf\u95ee\u6743\u9650\uff0c\u8bf7\u8054\u7cfb\u7ba1\u7406\u5458\u914d\u7f6e\u89d2\u8272\u3002",
  missingAuthConfig: "\u7f3a\u5c11 Supabase \u8ba4\u8bc1\u914d\u7f6e\uff0c\u8bf7\u68c0\u67e5 NEXT_PUBLIC_SUPABASE_URL \u4e0e KEY\u3002",
  hidePassword: "\u9690\u85cf\u5bc6\u7801",
  showPassword: "\u663e\u793a\u5bc6\u7801",
} as const

export function LoginGate({ supabaseUrl, publishableKey }: LoginGateProps) {
  const router = useRouter()
  const authReady = Boolean(supabaseUrl && publishableKey)
  const [state, setState] = React.useState<GateState>("checking")
  const [mode, setMode] = React.useState<FormMode>("login")
  const [displayName, setDisplayName] = React.useState("")
  const [email, setEmail] = React.useState("")
  const [password, setPassword] = React.useState("")
  const [error, setError] = React.useState("")
  const [passwordVisible, setPasswordVisible] = React.useState(false)
  const [registrationOpen, setRegistrationOpen] = React.useState(true)

  React.useEffect(() => {
    let cancelled = false

    async function loadPublicSettings() {
      try {
        const response = await fetch("/api/v1/public/settings", { cache: "no-store" })
        if (!response.ok) return
        const payload = await response.json()
        if (!cancelled) {
          setRegistrationOpen(Boolean(payload?.data?.registration_open))
        }
      } catch {
      }
    }

    void loadPublicSettings()
    return () => {
      cancelled = true
    }
  }, [])

  React.useEffect(() => {
    let cancelled = false

    async function bootstrap() {
      if (!authReady) {
        if (!cancelled) {
          setError(copy.missingAuthConfig)
          setState("ready")
        }
        return
      }

      const session = readStoredSession()

      if (!session) {
        if (!cancelled) {
          setState("ready")
        }
        return
      }

      const valid = await validateSession(supabaseUrl, publishableKey, session)
      if (cancelled) {
        return
      }

      if (valid) {
        const me = await fetchMe(valid.accessToken)
        if (cancelled) {
          return
        }

        if (me) {
          const withLearnerRole = await ensureLearnerBootstrap(valid.accessToken, me, "")
          if (cancelled) {
            return
          }

          if (withLearnerRole?.user.roles.includes("learner")) {
            writeStoredSession(
              mergeSessionWithActiveEnrollment(
                valid,
                withLearnerRole.active_exam_enrollment,
              ),
            )
            router.replace(withLearnerRole.active_exam_enrollment ? "/home" : "/onboarding")
            return
          }

          clearStoredSession()
          setError(copy.roleForbidden)
          setState("ready")
          return
        }
      }

      clearStoredSession()
      setState("ready")
    }

    void bootstrap()

    return () => {
      cancelled = true
    }
  }, [authReady, publishableKey, router, supabaseUrl])

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setState("submitting")
    setError("")

    if (!authReady) {
      setError(copy.missingAuthConfig)
      setState("ready")
      return
    }

    if (mode === "register" && !registrationOpen) {
      setError(copy.signupClosed)
      setState("ready")
      return
    }

    if (mode === "register") {
      const registered = await registerAccount({
        supabaseUrl,
        publishableKey,
        email,
        password,
      })

      if (!registered.ok) {
        setError(registered.error ?? copy.signupFailed)
        setState("ready")
        return
      }

      if (registered.requiresEmailConfirmation) {
        setError(copy.signupSuccessVerifyEmail)
        setMode("login")
        setState("ready")
        return
      }
    }

    try {
      const session = await signInWithPassword({
        supabaseUrl,
        publishableKey,
        email,
        password,
      })

      if (!session) {
        setError(copy.loginFailed)
        setState("ready")
        return
      }

      if (mode === "register") {
        const bootstrapped = await bootstrapLearner(session.accessToken, displayName)
        if (!bootstrapped || !bootstrapped.user.roles.includes("learner")) {
          setError(copy.bootstrapFailed)
          setState("ready")
          return
        }
        writeStoredSession(session)
        router.replace("/onboarding")
        return
      }

      const me = await fetchMe(session.accessToken)
      if (!me) {
        setError(copy.loginRequestFailed)
        setState("ready")
        return
      }

      const withLearnerRole = await ensureLearnerBootstrap(session.accessToken, me, displayName)

      if (!withLearnerRole) {
        setError(copy.bootstrapFailed)
        setState("ready")
        return
      }

      if (!withLearnerRole.user.roles.includes("learner")) {
        setError(copy.roleForbidden)
        setState("ready")
        return
      }

      writeStoredSession(
        mergeSessionWithActiveEnrollment(
          session,
          withLearnerRole.active_exam_enrollment,
        ),
      )
      router.replace(withLearnerRole.active_exam_enrollment ? "/home" : "/onboarding")
    } catch {
      setError(copy.loginRequestFailed)
      setState("ready")
    }
  }

  if (state === "checking") {
    return (
      <main className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[color:var(--surface)] px-6">
        <PlainBackground />
        <div className="relative z-10 grid place-items-center gap-4 text-center">
          <div className="font-heading text-5xl font-bold tracking-tight text-[color:var(--primary-container)]">
            FOCO
          </div>
          <div className="flex items-center gap-3 text-[color:var(--secondary)]">
            <LoaderCircle className="size-5 animate-spin" />
            <span className="text-sm font-medium text-[color:var(--on-surface-variant)]">
              {copy.checking}
            </span>
          </div>
        </div>
      </main>
    )
  }

  return (
    <main className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[color:var(--surface)] px-[var(--margin-mobile)] py-12 md:px-[var(--margin-desktop)]">
      <PlainBackground />

      <div className="relative z-10 flex w-full justify-center">
        <Card className="w-full max-w-[480px] rounded-[24px] border border-[color:var(--surface-container-highest)] bg-[rgba(255,255,255,0.97)] py-0 shadow-[0_8px_32px_rgba(0,0,0,0.08)] backdrop-blur-xl">
          <CardContent className="p-8 md:p-12">
            <div className="mb-10 text-center">
              <div className="mb-2 font-heading text-[48px] font-bold tracking-tight text-[color:var(--primary-container)]">
                FOCO
              </div>
              <p className="text-base text-[color:var(--on-surface-variant)]">
                {mode === "login" ? copy.title : copy.registerTitle}
              </p>
            </div>

            <form className="flex flex-col gap-6" onSubmit={handleSubmit}>
              {mode === "register" ? (
                <div className="flex flex-col gap-2">
                  <Label
                    htmlFor="display-name"
                    className="font-heading text-xs font-semibold uppercase tracking-[0.14em] text-[color:var(--on-surface-variant)]"
                  >
                    {copy.displayNameLabel}
                  </Label>
                  <Input
                    id="display-name"
                    value={displayName}
                    onChange={(event) => setDisplayName(event.target.value)}
                    placeholder={copy.displayNamePlaceholder}
                    className="h-[52px] rounded-full border-[1.5px] border-[color:var(--surface-container-highest)] bg-[color:var(--surface)] px-4 text-base text-[color:var(--on-surface)] placeholder:text-[color:var(--outline)] focus-visible:border-[color:var(--secondary)]"
                    required
                  />
                </div>
              ) : null}
              <div className="flex flex-col gap-2">
                <Label
                  htmlFor="username"
                  className="font-heading text-xs font-semibold uppercase tracking-[0.14em] text-[color:var(--on-surface-variant)]"
                >
                  {copy.usernameLabel}
                </Label>
                <div className="relative">
                  <UserRound className="pointer-events-none absolute left-4 top-1/2 size-5 -translate-y-1/2 text-[color:var(--outline-variant)]" />
                  <Input
                    id="username"
                    type="email"
                    autoComplete="email"
                    value={email}
                    onChange={(event) => setEmail(event.target.value)}
                    placeholder={copy.usernamePlaceholder}
                    className="h-[52px] rounded-full border-[1.5px] border-[color:var(--surface-container-highest)] bg-[color:var(--surface)] pl-12 pr-4 text-base text-[color:var(--on-surface)] placeholder:text-[color:var(--outline)] focus-visible:border-[color:var(--secondary)]"
                    required
                  />
                </div>
              </div>

              <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between gap-3">
                  <Label
                    htmlFor="password"
                    className="font-heading text-xs font-semibold uppercase tracking-[0.14em] text-[color:var(--on-surface-variant)]"
                  >
                    {copy.passwordLabel}
                  </Label>
                  <button
                    type="button"
                    onClick={() => toast.info("请联系管理员重置密码")}
                    className="text-xs font-semibold text-[color:var(--secondary)] transition-colors hover:text-[color:var(--primary-container)]"
                  >
                  {copy.forgotPassword}
                  </button>
                </div>
                <div className="relative">
                  <Lock className="pointer-events-none absolute left-4 top-1/2 size-5 -translate-y-1/2 text-[color:var(--outline-variant)]" />
                  <Input
                    id="password"
                    type={passwordVisible ? "text" : "password"}
                    autoComplete="current-password"
                    value={password}
                    onChange={(event) => setPassword(event.target.value)}
                    placeholder={copy.passwordPlaceholder}
                    className="h-[52px] rounded-full border-[1.5px] border-[color:var(--surface-container-highest)] bg-[color:var(--surface)] pl-12 pr-12 text-base text-[color:var(--on-surface)] placeholder:text-[color:var(--outline)] focus-visible:border-[color:var(--secondary)]"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setPasswordVisible((value) => !value)}
                    className="absolute right-4 top-1/2 -translate-y-1/2 text-[color:var(--outline)] transition-colors hover:text-[color:var(--on-surface)]"
                    aria-label={passwordVisible ? copy.hidePassword : copy.showPassword}
                  >
                    {passwordVisible ? <Eye className="size-5" /> : <EyeOff className="size-5" />}
                  </button>
                </div>
              </div>

            {error ? <p className="text-sm text-[color:var(--error)]">{error}</p> : null}

              <Button
                type="submit"
                disabled={state === "submitting"}
                className="mt-2 h-14 rounded-full bg-[color:var(--secondary)] font-heading text-[15px] font-semibold text-[color:var(--on-secondary)] shadow-sm transition-all duration-200 hover:bg-[color:var(--primary-container)] active:scale-[0.99]"
              >
                {state === "submitting" ? (
                  <>
                    <LoaderCircle className="size-4 animate-spin" />
                    {copy.submitting}
                  </>
                ) : (
                  <>
                    {copy.submit}
                    <ArrowRight className="size-5" />
                  </>
                )}
              </Button>
            </form>

            <div className="mt-10 text-center">
              <p className="text-base text-[color:var(--on-surface-variant)]">
                {mode === "login" ? copy.noAccount : copy.submit}
                <button
                  type="button"
                  onClick={() => {
                    if (mode === "login" && !registrationOpen) {
                      setError(copy.signupClosed)
                      return
                    }
                    setError("")
                    setMode((value) => (value === "login" ? "register" : "login"))
                  }}
                  className="ml-1 font-heading text-[15px] font-bold text-[color:var(--secondary)] transition-colors hover:text-[color:var(--primary-container)] disabled:cursor-not-allowed disabled:opacity-50"
                  disabled={mode === "login" && !registrationOpen}
                >
                  {mode === "login" ? copy.createAccount : copy.backToLogin}
                </button>
              </p>
            </div>
          </CardContent>
        </Card>
      </div>
    </main>
  )
}

function PlainBackground() {
  return (
    <div className="absolute inset-0 z-0 overflow-hidden" aria-hidden="true">
      <div className="absolute inset-0 bg-[color:var(--surface)]" />
      <div className="absolute -left-24 top-[-80px] h-72 w-72 rounded-full bg-[rgba(232,226,216,0.65)] blur-3xl" />
      <div className="absolute right-[-100px] top-[10%] h-80 w-80 rounded-full bg-[rgba(6,56,32,0.14)] blur-3xl" />
      <div className="absolute bottom-[-100px] left-[18%] h-72 w-72 rounded-full bg-[rgba(0,109,56,0.12)] blur-3xl" />
    </div>
  )
}

async function validateSession(
  supabaseUrl: string,
  publishableKey: string,
  session: StoredAuthSession,
): Promise<StoredAuthSession | null> {
  const userResponse = await fetch(`${supabaseUrl}/auth/v1/user`, {
    method: "GET",
    headers: {
      apikey: publishableKey,
      Authorization: `Bearer ${session.accessToken}`,
    },
  })

  if (userResponse.ok) {
    return session
  }

  if (!session.refreshToken) {
    return null
  }

  const refreshResponse = await fetch(`${supabaseUrl}/auth/v1/token?grant_type=refresh_token`, {
    method: "POST",
    headers: {
      apikey: publishableKey,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      refresh_token: session.refreshToken,
    }),
  })

  if (!refreshResponse.ok) {
    return null
  }

  const payload = (await refreshResponse.json()) as TokenResponse
  return normalizeSession(payload)
}

async function signInWithPassword(input: {
  supabaseUrl: string
  publishableKey: string
  email: string
  password: string
}): Promise<StoredAuthSession | null> {
  const response = await fetch(`${input.supabaseUrl}/auth/v1/token?grant_type=password`, {
    method: "POST",
    headers: {
      apikey: input.publishableKey,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      email: input.email,
      password: input.password,
    }),
  })

  if (!response.ok) {
    return null
  }

  const payload = (await response.json()) as TokenResponse
  return normalizeSession(payload)
}

async function registerAccount(input: {
  supabaseUrl: string
  publishableKey: string
  email: string
  password: string
}) {
  const response = await fetch(`${input.supabaseUrl}/auth/v1/signup`, {
    method: "POST",
    headers: {
      apikey: input.publishableKey,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      email: input.email,
      password: input.password,
    }),
  })

  const payload = (await response.json().catch(() => ({}))) as {
    code?: number
    error_code?: string
    msg?: string
    user?: { id?: string }
    session?: TokenResponse["session"]
  }

  if (!response.ok) {
    return {
      ok: false,
      error: mapSignupError(payload.error_code, payload.msg),
    }
  }

  return {
    ok: true,
    requiresEmailConfirmation: !payload.session,
  }
}

async function bootstrapLearner(token: string, displayName: string) {
  const baseUrl = "/api/v1"
  const response = await fetch(`${baseUrl}/learner/bootstrap`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      display_name: displayName,
    }),
  })

  if (!response.ok) {
    return null
  }

  const payload = (await response.json()) as {
    data: {
      user: {
        id: string
        email: string
        roles: string[]
      }
    }
  }

  return payload.data
}

async function fetchMe(token: string) {
  const baseUrl = "/api/v1"

  try {
    const response = await fetch(`${baseUrl}/me`, {
      method: "GET",
      headers: {
        Authorization: `Bearer ${token}`,
      },
      cache: "no-store",
    })

    if (!response.ok) {
      return null
    }

    const payload = (await response.json()) as {
      data: {
        user: {
          id: string
          email: string
          roles: string[]
        }
        active_exam_enrollment: null | {
          id: string
          exam_id: string
          exam_code: string
          exam_name: string
          status: string
        }
      }
    }

    return payload.data
  } catch {
    return null
  }
}

type MeData = NonNullable<Awaited<ReturnType<typeof fetchMe>>>

async function ensureLearnerBootstrap(
  token: string,
  me: MeData,
  displayName: string,
): Promise<MeData | null> {
  if (me.user.roles.includes("learner")) {
    return me
  }

  const fallbackName =
    displayName.trim() ||
    me.user.email?.split("@")[0] ||
    "Learner"

  const bootstrapped = await bootstrapLearner(token, fallbackName)
  if (!bootstrapped) {
    return null
  }

  if (bootstrapped.user.roles.includes("learner")) {
    return {
      user: bootstrapped.user,
      active_exam_enrollment: me.active_exam_enrollment,
    }
  }

  return fetchMe(token)
}

function mapSignupError(errorCode?: string, message?: string) {
  switch (errorCode) {
    case "email_address_invalid":
      return "\u90ae\u7bb1\u5730\u5740\u65e0\u6548\uff0c\u8bf7\u4f7f\u7528\u771f\u5b9e\u53ef\u7528\u7684\u90ae\u7bb1\u3002"
    case "user_already_exists":
      return "\u8fd9\u4e2a\u90ae\u7bb1\u5df2\u7ecf\u6ce8\u518c\uff0c\u8bf7\u76f4\u63a5\u767b\u5f55\u3002"
    case "weak_password":
      return "\u5bc6\u7801\u5f3a\u5ea6\u4e0d\u8db3\uff0c\u8bf7\u4f7f\u7528\u66f4\u5f3a\u7684\u5bc6\u7801\u3002"
    default:
      return message || copy.signupFailed
  }
}

function normalizeSession(payload: TokenResponse): StoredAuthSession {
  return {
    accessToken: payload.access_token,
    refreshToken: payload.refresh_token,
    expiresAt: payload.expires_at,
    user: payload.user
      ? {
          id: payload.user.id,
          email: payload.user.email,
        }
      : undefined,
  }
}
