import { LoginGate } from "@/components/auth/login-gate"

export default function LoginPage() {
  return (
    <LoginGate
      supabaseUrl={process.env.NEXT_PUBLIC_SUPABASE_URL ?? process.env.SUPABASE_URL ?? ""}
      publishableKey={
        process.env.NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY ??
        process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY ??
        ""
      }
    />
  )
}
