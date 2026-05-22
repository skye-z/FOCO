export const AUTH_STORAGE_KEY = "foco.auth.session"
export const ACCESS_TOKEN_COOKIE = "foco_access_token"
const REMEMBERED_EMAIL_KEY = "foco.auth.remembered_email"

export type StoredAuthSession = {
  accessToken: string
  refreshToken?: string
  expiresAt?: number
  user?: {
    id?: string
    email?: string
  }
}

export function readRememberedEmail(): string {
  if (typeof window === "undefined") return ""
  return window.localStorage.getItem(REMEMBERED_EMAIL_KEY) ?? ""
}

export function writeRememberedEmail(email: string) {
  if (typeof window === "undefined") return
  window.localStorage.setItem(REMEMBERED_EMAIL_KEY, email)
}

export function readStoredSession(): StoredAuthSession | null {
  if (typeof window === "undefined") {
    return null
  }

  const raw = window.localStorage.getItem(AUTH_STORAGE_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as StoredAuthSession
  } catch {
    return null
  }
}

export function writeStoredSession(session: StoredAuthSession) {
  if (typeof window === "undefined") {
    return
  }

  window.localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(session))
  setBrowserCookie(ACCESS_TOKEN_COOKIE, session.accessToken, 60 * 60 * 24 * 7)
}

export function clearStoredSession() {
  if (typeof window === "undefined") {
    return
  }

  window.localStorage.removeItem(AUTH_STORAGE_KEY)
  setBrowserCookie(ACCESS_TOKEN_COOKIE, "", 0)
}

export function readBrowserAccessToken(): string | null {
  if (typeof document === "undefined") {
    return null
  }

  const match = document.cookie
    .split("; ")
    .find((entry) => entry.startsWith(`${ACCESS_TOKEN_COOKIE}=`))

  if (!match) {
    return readStoredSession()?.accessToken ?? null
  }

  return decodeURIComponent(match.split("=", 2)[1] ?? "")
}

function setBrowserCookie(name: string, value: string, maxAgeSeconds: number) {
  if (typeof document === "undefined") {
    return
  }

  document.cookie = `${name}=${encodeURIComponent(value)}; Path=/; Max-Age=${maxAgeSeconds}; SameSite=Lax`
}

let refreshPromise: Promise<StoredAuthSession | null> | null = null

export function authFetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response> {
  return fetch(input, init).then(async (res) => {
    if (res.status !== 401) return res

    const session = readStoredSession()
    if (!session?.refreshToken) {
      clearStoredSession()
      window.location.href = "/"
      return res
    }

    const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL ?? process.env.SUPABASE_URL ?? ""
    const publishableKey = process.env.NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY ?? process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY ?? ""

    if (!supabaseUrl || !publishableKey) {
      clearStoredSession()
      window.location.href = "/"
      return res
    }

    if (!refreshPromise) {
      refreshPromise = (async () => {
        try {
          const refreshRes = await fetch(`${supabaseUrl}/auth/v1/token?grant_type=refresh_token`, {
            method: "POST",
            headers: { apikey: publishableKey, "Content-Type": "application/json" },
            body: JSON.stringify({ refresh_token: session!.refreshToken }),
          })
          if (!refreshRes.ok) return null
          const payload = await refreshRes.json()
          const newSession: StoredAuthSession = {
            accessToken: payload.access_token,
            refreshToken: payload.refresh_token ?? session!.refreshToken,
            expiresAt: payload.expires_at,
            user: payload.user ? { id: payload.user.id, email: payload.user.email } : session!.user,
          }
          writeStoredSession(newSession)
          return newSession
        } catch {
          return null
        } finally {
          refreshPromise = null
        }
      })()
    }

    const newSession = await refreshPromise
    if (!newSession) {
      clearStoredSession()
      window.location.href = "/"
      return res
    }

    const retryInit: RequestInit = { ...init, headers: { ...init?.headers, Authorization: `Bearer ${newSession.accessToken}` } }
    return fetch(input, retryInit)
  })
}
