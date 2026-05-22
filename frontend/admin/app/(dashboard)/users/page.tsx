"use client"

import * as React from "react"
import { LoaderCircle, Search, ShieldCheck, UserRound } from "lucide-react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"

import { authFetch, clearStoredSession, readBrowserAccessToken } from "@/lib/auth-session"
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

type AdminUser = {
  id: string
  email: string
  display_name: string
  status: string
  created_at: string
  roles: string[]
}

type RoleOption = "admin" | "editor" | "learner"

function roleTone(role: string) {
  switch (role) {
    case "admin":
      return "bg-emerald-50 text-emerald-700"
    case "editor":
      return "bg-blue-50 text-blue-700"
    case "learner":
      return "bg-violet-50 text-violet-700"
    default:
      return "bg-muted text-muted-foreground"
  }
}

function roleLabel(role: string) {
  switch (role) {
    case "admin": return "管理员"
    case "editor": return "编辑者"
    case "learner": return "学习者"
    default: return role
  }
}

function showUserActionError(title: string, description?: string) {
  toast.error(title, {
    description: description ?? "请稍后重试，或刷新页面后再次操作。",
  })
}

export default function UsersPage() {
  const router = useRouter()
  const [users, setUsers] = React.useState<AdminUser[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState("")
  const [search, setSearch] = React.useState("")
  const [roleFilter, setRoleFilter] = React.useState("all")
  const [roleDialogUser, setRoleDialogUser] = React.useState<AdminUser | null>(null)
  const [roleToGrant, setRoleToGrant] = React.useState<RoleOption>("editor")
  const [resetDialogUser, setResetDialogUser] = React.useState<AdminUser | null>(null)
  const [newPassword, setNewPassword] = React.useState("")
  const [disableUser, setDisableUser] = React.useState<AdminUser | null>(null)
  const [submitting, setSubmitting] = React.useState(false)

  const loadUsers = React.useCallback(async () => {
    const token = readBrowserAccessToken()
    if (!token) {
      clearStoredSession()
      router.replace("/")
      return
    }

    try {
      const res = await authFetch("/api/v1/admin/users", {
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
      })

      if (!res.ok) {
        if (res.status === 401) {
          clearStoredSession()
          router.replace("/")
          return
        }
        throw new Error("load users failed")
      }

      const payload = await res.json()
      setUsers((payload.data ?? []) as AdminUser[])
      setError("")
    } catch {
      setError("用户列表加载失败，请刷新后重试。")
    } finally {
      setLoading(false)
    }
  }, [router])

  React.useEffect(() => {
    void loadUsers()
  }, [loadUsers])

  const filteredUsers = React.useMemo(() => {
    const keyword = search.trim().toLowerCase()
    return users.filter((user) => {
      const matchesKeyword =
        keyword === "" ||
        user.display_name.toLowerCase().includes(keyword) ||
        user.email.toLowerCase().includes(keyword) ||
        user.id.toLowerCase().includes(keyword)

      const matchesRole =
        roleFilter === "all"
          ? true
          : roleFilter === "none"
            ? user.roles.length === 0
            : user.roles.includes(roleFilter)

      return matchesKeyword && matchesRole
    })
  }, [users, search, roleFilter])

  const adminCount = users.filter((user) => user.roles.includes("admin")).length
  const editorCount = users.filter((user) => user.roles.includes("editor")).length
  const activeCount = users.filter((user) => user.status === "active").length

  async function postUserAction(path: string, body?: object) {
    const token = readBrowserAccessToken()
    if (!token) {
      clearStoredSession()
      router.replace("/")
      return false
    }

    const res = await authFetch(path, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: body ? JSON.stringify(body) : JSON.stringify({}),
    })

    if (!res.ok) {
      const message = await res.text().catch(() => "")
      throw new Error(message || "request failed")
    }

    return true
  }

  async function handleGrantRole() {
    if (!roleDialogUser) return
    setSubmitting(true)
    try {
      await postUserAction(`/api/v1/admin/users/${roleDialogUser.id}/roles`, {
        role: roleToGrant,
      })
      toast.success("角色已补充", {
        description: `已为 ${roleDialogUser.display_name || roleDialogUser.email} 添加 ${roleLabel(roleToGrant)} 角色。`,
      })
      setRoleDialogUser(null)
      await loadUsers()
    } catch (error) {
      showUserActionError("补角色失败", error instanceof Error ? error.message : undefined)
    } finally {
      setSubmitting(false)
    }
  }

  async function handleResetPassword() {
    if (!resetDialogUser) return
    if (newPassword.trim().length < 8) {
      showUserActionError("重置密码失败", "临时密码至少需要 8 个字符。")
      return
    }

    setSubmitting(true)
    try {
      await postUserAction(`/api/v1/admin/users/${resetDialogUser.id}/reset-password`, {
        new_password: newPassword,
      })
      toast.success("密码已重置", {
        description: `已为 ${resetDialogUser.display_name || resetDialogUser.email} 设置新的临时密码。`,
      })
      setResetDialogUser(null)
      setNewPassword("")
    } catch (error) {
      showUserActionError("重置密码失败", error instanceof Error ? error.message : undefined)
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDisableUser() {
    if (!disableUser) return
    setSubmitting(true)
    try {
      await postUserAction(`/api/v1/admin/users/${disableUser.id}/disable`)
      toast.success("用户已禁用", {
        description: `${disableUser.display_name || disableUser.email} 已被禁用。`,
      })
      setDisableUser(null)
      await loadUsers()
    } catch (error) {
      showUserActionError("禁用用户失败", error instanceof Error ? error.message : undefined)
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <LoaderCircle className="size-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <>
      <main className="mx-auto max-w-6xl px-6 py-8">
        <div className="mb-8 flex flex-col gap-2">
          <h1 className="text-2xl font-bold tracking-tight">用户管理</h1>
          <p className="text-sm text-muted-foreground">
            支持搜索、按角色筛选，并执行后台用户操作。
          </p>
        </div>

        <div className="mb-6 grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="flex items-center gap-2 text-sm text-muted-foreground">
                <ShieldCheck className="size-4" />
                管理员
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{adminCount}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="flex items-center gap-2 text-sm text-muted-foreground">
                <UserRound className="size-4" />
                编辑者
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{editorCount}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">活跃用户</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold">{activeCount}</div>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader className="gap-4">
            <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
              <CardTitle>用户列表</CardTitle>
              <div className="flex flex-col gap-3 md:flex-row md:items-center">
                <div className="relative w-full md:w-72">
                  <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    value={search}
                    onChange={(event) => setSearch(event.target.value)}
                    placeholder="搜索姓名、邮箱或用户 ID"
                    className="pl-9"
                  />
                </div>
                <Select value={roleFilter} onValueChange={(value) => setRoleFilter(value ?? "all")}>
                  <SelectTrigger className="w-full md:w-44">
                    {roleFilter === "all"
                      ? "全部角色"
                      : roleFilter === "none"
                        ? "无角色"
                        : roleLabel(roleFilter)}
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">全部角色</SelectItem>
                    <SelectItem value="admin">管理员</SelectItem>
                    <SelectItem value="editor">编辑者</SelectItem>
                    <SelectItem value="learner">学习者</SelectItem>
                    <SelectItem value="none">无角色</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            {error ? (
              <div className="rounded-lg border border-destructive/30 bg-destructive/5 px-4 py-3 text-sm text-destructive">
                {error}
              </div>
            ) : filteredUsers.length === 0 ? (
              <div className="rounded-lg border border-dashed px-4 py-10 text-center text-sm text-muted-foreground">
                没有符合条件的用户
              </div>
            ) : (
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>用户</TableHead>
                    <TableHead>角色</TableHead>
                    <TableHead>状态</TableHead>
                    <TableHead>创建时间</TableHead>
                    <TableHead>操作</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filteredUsers.map((user) => (
                    <TableRow key={user.id}>
                      <TableCell className="align-top">
                        <div className="flex flex-col gap-1">
                          <span className="font-medium text-foreground">
                            {user.display_name || "未命名用户"}
                          </span>
                          <span className="text-sm text-muted-foreground">{user.email || "无邮箱"}</span>
                        </div>
                      </TableCell>
                      <TableCell className="align-top">
                        <div className="flex flex-wrap gap-1.5">
                          {user.roles.length === 0 ? (
                            <Badge variant="outline">无角色</Badge>
                          ) : (
                            user.roles.map((role) => (
                              <Badge key={role} variant="outline" className={roleTone(role)}>
                                {roleLabel(role)}
                              </Badge>
                            ))
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="align-top">
                        <Badge
                          variant="outline"
                          className={
                            user.status === "active"
                              ? "bg-emerald-50 text-emerald-700"
                              : "bg-muted text-muted-foreground"
                          }
                        >
                          {user.status}
                        </Badge>
                      </TableCell>
                      <TableCell className="align-top text-muted-foreground">
                        {new Date(user.created_at).toLocaleString("zh-CN")}
                      </TableCell>
                      <TableCell className="align-top">
                        <div className="flex flex-wrap gap-2">
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              setRoleDialogUser(user)
                              setRoleToGrant(user.roles.includes("editor") ? "learner" : "editor")
                            }}
                          >
                            补角色
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => {
                              setResetDialogUser(user)
                              setNewPassword("")
                            }}
                          >
                            重置密码
                          </Button>
                          <Button
                            type="button"
                            variant="destructive"
                            size="sm"
                            onClick={() => setDisableUser(user)}
                            disabled={user.status === "disabled"}
                          >
                            禁用用户
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </main>

      <Dialog
        open={roleDialogUser !== null}
        onOpenChange={(open) => {
          if (!open) setRoleDialogUser(null)
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>补角色</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="rounded-lg bg-muted/50 px-3 py-2 text-sm text-muted-foreground">
              目标用户：{roleDialogUser?.display_name || roleDialogUser?.email}
            </div>
            <div className="space-y-2">
              <Label>角色</Label>
              <Select value={roleToGrant} onValueChange={(value) => setRoleToGrant((value as RoleOption) ?? "editor")}>
                <SelectTrigger>{roleLabel(roleToGrant)}</SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">管理员</SelectItem>
                  <SelectItem value="editor">编辑者</SelectItem>
                  <SelectItem value="learner">学习者</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setRoleDialogUser(null)}>
              取消
            </Button>
            <Button type="button" onClick={handleGrantRole} disabled={submitting}>
              {submitting ? "提交中..." : "确认补角色"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={resetDialogUser !== null}
        onOpenChange={(open) => {
          if (!open) {
            setResetDialogUser(null)
            setNewPassword("")
          }
        }}
      >
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>重置密码</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="rounded-lg bg-muted/50 px-3 py-2 text-sm text-muted-foreground">
              目标用户：{resetDialogUser?.display_name || resetDialogUser?.email}
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-password">临时密码</Label>
              <Input
                id="new-password"
                type="text"
                value={newPassword}
                onChange={(event) => setNewPassword(event.target.value)}
                placeholder="至少 8 位，建议包含大小写和数字"
              />
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setResetDialogUser(null)}>
              取消
            </Button>
            <Button type="button" onClick={handleResetPassword} disabled={submitting || newPassword.trim().length < 8}>
              {submitting ? "提交中..." : "确认重置密码"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={disableUser !== null}
        onOpenChange={(open) => {
          if (!open) setDisableUser(null)
        }}
      >
        <AlertDialogContent size="sm">
          <AlertDialogHeader>
            <AlertDialogTitle>确认禁用用户</AlertDialogTitle>
            <AlertDialogDescription>
              <span className="text-foreground">
                确定要禁用「{disableUser?.display_name || disableUser?.email}」吗？
              </span>{" "}
              禁用后该用户将无法继续正常登录。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction type="button" variant="destructive" onClick={handleDisableUser} disabled={submitting}>
              {submitting ? "禁用中..." : "确认禁用"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
