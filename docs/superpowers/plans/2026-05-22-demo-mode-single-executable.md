# Demo Mode Single Executable Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 给 FOCO 增加 demo 模式：后端同进程托管 learner `/`、admin `/admin` 和 API `/api/v1`，并支持打包成单个 Go 可执行文件。

**Architecture:** 保留现有开发/正式交付链路，新增一个 demo-only 构建路径。两套 Next 前端在 demo 构建中导出为静态资源，Go 后端通过 `go:embed` 在 `-tags demo` 下嵌入资源并负责静态资源解析、SPA fallback 和公开运行时配置；前端通过 route helper 和扩展后的 `public/settings` 完成静态友好路由与运行时配置解耦。

**Tech Stack:** Go 1.25、`net/http`、`go:embed`、Next.js 15 App Router、React 19、Vitest、PowerShell/bash build scripts

---

## File Structure / Responsibility Map

### Backend config and public settings

- Modify: `D:\Project\AI\FOCO\backend\cmd\api\main.go`  
  读取 `FOCO_DEMO_MODE`，把 demo 模式和公开 Supabase 配置传入依赖层。
- Modify: `D:\Project\AI\FOCO\backend\internal\app\dependencies.go`  
  扩展 `app.Config`，把 demo 公开配置注入 platform service / router。
- Modify: `D:\Project\AI\FOCO\backend\internal\domain\platform\model.go`  
  扩展 `PublicSettings`，增加 `demo_mode`、`supabase_url`、`supabase_publishable_key`。
- Modify: `D:\Project\AI\FOCO\backend\internal\domain\platform\service.go`  
  让 `GetPublicSettings` 合并 DB 内的 `registration_open` 和环境里的 demo/public auth 配置。
- Modify: `D:\Project\AI\FOCO\backend\internal\domain\platform\service_test.go`  
  为新的公开字段和缓存行为补测试。

### Backend demo static site hosting

- Create: `D:\Project\AI\FOCO\backend\internal\demo\web\embed_demo.go`  
  `//go:build demo`，嵌入 admin / learner 静态资源。
- Create: `D:\Project\AI\FOCO\backend\internal\demo\web\embed_stub.go`  
  非 demo build tag 下提供空实现，保持普通构建不受影响。
- Create: `D:\Project\AI\FOCO\backend\internal\demo\web\site.go`  
  负责按 URL 查找嵌入资源、设置缓存头、做 fallback 判定。
- Modify: `D:\Project\AI\FOCO\backend\internal\http\router.go`  
  接入 `/admin`、`/` 的静态站点托管，同时保护 `/api/v1/**` 优先级。
- Create: `D:\Project\AI\FOCO\backend\internal\http\demo_router_test.go`  
  覆盖 admin/learner 命中、SPA fallback、资源 404、API 不被覆盖。

### Admin frontend runtime config and routing

- Create: `D:\Project\AI\FOCO\frontend\admin\lib\runtime-config.ts`
- Create: `D:\Project\AI\FOCO\frontend\admin\lib\runtime-config.test.ts`
- Create: `D:\Project\AI\FOCO\frontend\admin\lib\routes.ts`
- Create: `D:\Project\AI\FOCO\frontend\admin\lib\routes.test.ts`
- Create: `D:\Project\AI\FOCO\frontend\admin\components\interactive\interactive-unit-editor-page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\interactive-unit-editor\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\components\auth\login-gate.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\lib\auth-session.ts`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\interactive-units\[unitId]\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\exams\components\create-interactive-unit-dialog.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\exams\components\interactive-unit-card.tsx`

### Learner frontend runtime config and routing

- Create: `D:\Project\AI\FOCO\frontend\learner\lib\runtime-config.ts`
- Create: `D:\Project\AI\FOCO\frontend\learner\lib\runtime-config.test.ts`
- Create: `D:\Project\AI\FOCO\frontend\learner\lib\routes.ts`
- Create: `D:\Project\AI\FOCO\frontend\learner\lib\routes.test.ts`
- Create: `D:\Project\AI\FOCO\frontend\learner\components\practice\practice-session-page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\components\practice\practice-complete-page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\components\labs\lab-view-page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\session\page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\complete\page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\labs\view\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\components\auth\login-gate.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\lib\auth-session.ts`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\[sessionId]\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\[sessionId]\complete\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\labs\[unitVersionId]\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\practice\setup\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\home\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\labs\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\profile\page.tsx`

### Demo export config and packaging

- Modify: `D:\Project\AI\FOCO\frontend\admin\next.config.ts`
- Modify: `D:\Project\AI\FOCO\frontend\learner\next.config.ts`
- Modify: `D:\Project\AI\FOCO\.gitignore`
- Create: `D:\Project\AI\FOCO\backend\internal\demo\web\dist\.gitkeep`
- Create: `D:\Project\AI\FOCO\build-demo.ps1`
- Create: `D:\Project\AI\FOCO\build-demo.sh`
- Create: `D:\Project\AI\FOCO\test\scripts\smoke-demo.mjs`
- Modify: `D:\Project\AI\FOCO\test\package.json`
- Modify: `D:\Project\AI\FOCO\README.md`
- Modify: `D:\Project\AI\FOCO\test\README.md`

---

### Task 1: 扩展 backend 公共配置契约

**Files:**
- Modify: `D:\Project\AI\FOCO\backend\cmd\api\main.go`
- Modify: `D:\Project\AI\FOCO\backend\internal\app\dependencies.go`
- Modify: `D:\Project\AI\FOCO\backend\internal\domain\platform\model.go`
- Modify: `D:\Project\AI\FOCO\backend\internal\domain\platform\service.go`
- Modify: `D:\Project\AI\FOCO\backend\internal\domain\platform\service_test.go`

- [ ] **Step 1: 写 platform service 的失败测试，定义新的 public settings 输出**

```go
func TestGetPublicSettingsIncludesDemoAndSupabaseFields(t *testing.T) {
    svc := NewService(nil, nil)
    svc.public = PublicRuntimeConfig{
        DemoMode: true,
        SupabaseURL: "https://demo.supabase.co",
        SupabasePublishableKey: "pk-demo",
    }

    got, err := svc.getPublicSettingsUncached(context.Background())
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if !got.DemoMode || got.SupabaseURL == "" || got.SupabasePublishableKey == "" {
        t.Fatalf("public settings missing runtime fields: %#v", got)
    }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\backend
go test ./internal/domain/platform -run TestGetPublicSettingsIncludesDemoAndSupabaseFields -v
```

Expected: FAIL，报 `PublicSettings` 缺字段或 `Service` 缺运行时配置注入。

- [ ] **Step 3: 最小实现 backend 公开配置链路**

实现要点：

```go
type Config struct {
    SupabaseURL string
    PublishableKey string
    ServiceRoleKey string
    DatabaseURL string
    RedisURL string
    DemoMode bool
}

type PublicSettings struct {
    RegistrationOpen bool   `json:"registration_open"`
    DemoMode         bool   `json:"demo_mode"`
    SupabaseURL      string `json:"supabase_url"`
    SupabasePublishableKey string `json:"supabase_publishable_key"`
}
```

`main.go` 读取：

```go
demoMode := strings.EqualFold(os.Getenv("FOCO_DEMO_MODE"), "true")
```

并把 `DemoMode`、`SupabaseURL`、`PublishableKey` 传给 platform service。

- [ ] **Step 4: 运行 platform 测试确认通过**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\backend
go test ./internal/domain/platform -v
```

Expected: PASS。

- [ ] **Step 5: Commit**

```powershell
Set-Location D:\Project\AI\FOCO
git add backend/cmd/api/main.go backend/internal/app/dependencies.go backend/internal/domain/platform/model.go backend/internal/domain/platform/service.go backend/internal/domain/platform/service_test.go
git commit -m "feat: expose demo runtime settings"
```

---

### Task 2: 为 backend 增加 demo 静态站点托管与路由保护

**Files:**
- Create: `D:\Project\AI\FOCO\backend\internal\demo\web\embed_demo.go`
- Create: `D:\Project\AI\FOCO\backend\internal\demo\web\embed_stub.go`
- Create: `D:\Project\AI\FOCO\backend\internal\demo\web\site.go`
- Modify: `D:\Project\AI\FOCO\backend\internal\http\router.go`
- Create: `D:\Project\AI\FOCO\backend\internal\http\demo_router_test.go`

- [ ] **Step 1: 写 router 失败测试，先锁定 API 优先级和 fallback 规则**

```go
func TestDemoRouterKeepsAPIAndSeparatesAdminLearner(t *testing.T) {
    router := NewRouter(Dependencies{
        SeedChinese: func(http.ResponseWriter, *http.Request) {},
        DemoMode: true,
        DemoSite: fakeDemoSite(),
    })

    cases := []struct{
        path string
        want int
        wantContentType string
    }{
        {"/api/v1/health", http.StatusOK, "application/json"},
        {"/admin/users", http.StatusOK, "text/html"},
        {"/home", http.StatusOK, "text/html"},
        {"/admin/_next/missing.js", http.StatusNotFound, ""},
        {"/_next/missing.js", http.StatusNotFound, ""},
    }
}
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\backend
go test ./internal/http -run TestDemoRouterKeepsAPIAndSeparatesAdminLearner -v
```

Expected: FAIL，router 还没有 demo site 挂载与 fallback 规则。

- [ ] **Step 3: 实现 demo web 包与 router 接线**

实现要点：

```go
//go:build demo
var embedded embed.FS

type Site struct {
    enabled bool
    fs fs.FS
}

func (s Site) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. exact file
    // 2. path/index.html
    // 3. SPA fallback for extensionless route
}
```

router 顺序：

```go
if deps.DemoSite.Enabled() {
    mux.Handle("/admin/", deps.DemoSite.AdminHandler())
    mux.Handle("/", deps.DemoSite.LearnerHandler())
}
```

但 `/api/v1/**` 必须始终先注册。

- [ ] **Step 4: 运行 backend HTTP 测试确认通过**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\backend
go test ./internal/http -v
```

Expected: PASS。

- [ ] **Step 5: Commit**

```powershell
Set-Location D:\Project\AI\FOCO
git add backend/internal/demo/web backend/internal/http/router.go backend/internal/http/demo_router_test.go
git commit -m "feat: serve embedded demo frontend from backend"
```

---

### Task 3: Admin 前端切到运行时公开配置

**Files:**
- Create: `D:\Project\AI\FOCO\frontend\admin\lib\runtime-config.ts`
- Create: `D:\Project\AI\FOCO\frontend\admin\lib\runtime-config.test.ts`
- Modify: `D:\Project\AI\FOCO\frontend\admin\lib\auth-session.ts`
- Modify: `D:\Project\AI\FOCO\frontend\admin\components\auth\login-gate.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\page.tsx`

- [ ] **Step 1: 写 runtime-config 失败测试**

```ts
import { describe, expect, it, vi } from "vitest"
import { loadRuntimeConfig } from "./runtime-config"

describe("loadRuntimeConfig", () => {
  it("reads public settings from backend", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        data: {
          demo_mode: true,
          supabase_url: "https://demo.supabase.co",
          supabase_publishable_key: "pk-demo",
          registration_open: true,
        },
      }),
    }))

    await expect(loadRuntimeConfig()).resolves.toMatchObject({
      demoMode: true,
      supabaseUrl: "https://demo.supabase.co",
      publishableKey: "pk-demo",
    })
  })
})
```

- [ ] **Step 2: 运行 admin 测试确认失败**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\admin
cmd /c npm.cmd test -- lib/runtime-config.test.ts
```

Expected: FAIL，缺 `runtime-config.ts`。

- [ ] **Step 3: 最小实现 runtime config 和 auth-session 解耦**

实现要点：

```ts
export async function loadRuntimeConfig() {
  const res = await fetch("/api/v1/public/settings", { cache: "no-store" })
  const payload = await res.json()
  return {
    demoMode: Boolean(payload?.data?.demo_mode),
    supabaseUrl: payload?.data?.supabase_url ?? "",
    publishableKey: payload?.data?.supabase_publishable_key ?? "",
    registrationOpen: Boolean(payload?.data?.registration_open),
  }
}
```

`auth-session.ts` 不再直接读 `process.env`，而是读一个缓存的 runtime config getter。

- [ ] **Step 4: 运行 admin 测试和 typecheck**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\admin
cmd /c npm.cmd test -- lib/runtime-config.test.ts
cmd /c npm.cmd run typecheck
```

Expected: PASS。

- [ ] **Step 5: Commit**

```powershell
Set-Location D:\Project\AI\FOCO
git add frontend/admin/lib/runtime-config.ts frontend/admin/lib/runtime-config.test.ts frontend/admin/lib/auth-session.ts frontend/admin/components/auth/login-gate.tsx frontend/admin/app/page.tsx
git commit -m "feat: load admin auth config from backend runtime settings"
```

---

### Task 4: Admin 路由收口到静态友好编辑页

**Files:**
- Create: `D:\Project\AI\FOCO\frontend\admin\lib\routes.ts`
- Create: `D:\Project\AI\FOCO\frontend\admin\lib\routes.test.ts`
- Create: `D:\Project\AI\FOCO\frontend\admin\components\interactive\interactive-unit-editor-page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\interactive-unit-editor\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\interactive-units\[unitId]\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\exams\components\create-interactive-unit-dialog.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\exams\components\interactive-unit-card.tsx`

- [ ] **Step 1: 写 admin route helper 失败测试**

```ts
import { describe, expect, it } from "vitest"
import { buildInteractiveUnitEditorHref } from "./routes"

describe("buildInteractiveUnitEditorHref", () => {
  it("routes demo navigation through static query page", () => {
    expect(buildInteractiveUnitEditorHref("unit-1", "exam-1", { demoMode: true }))
      .toBe("/interactive-unit-editor?unit_id=unit-1&exam_id=exam-1")
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\admin
cmd /c npm.cmd test -- lib/routes.test.ts
```

Expected: FAIL。

- [ ] **Step 3: 抽取编辑页组件并更新所有 admin 跳转**

实现要点：

```tsx
export function InteractiveUnitEditorPage(props: { unitId: string; examId: string }) {
  // 从 props 取 unitId / examId，不再在组件内部依赖 useParams
}
```

- 旧 `[unitId]` 页面改成 wrapper，读取 params 后传给组件
- 新 `/interactive-unit-editor` 页面读取 query 后传给同一组件
- 所有 `router.push("/interactive-units/...")` 改为 helper

- [ ] **Step 4: 运行 admin 测试和 typecheck**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\admin
cmd /c npm.cmd test -- lib/routes.test.ts
cmd /c npm.cmd run typecheck
```

Expected: PASS。

- [ ] **Step 5: Commit**

```powershell
Set-Location D:\Project\AI\FOCO
git add frontend/admin/lib/routes.ts frontend/admin/lib/routes.test.ts frontend/admin/components/interactive/interactive-unit-editor-page.tsx frontend/admin/app/(dashboard)/interactive-unit-editor/page.tsx frontend/admin/app/(dashboard)/interactive-units/[unitId]/page.tsx frontend/admin/app/(dashboard)/exams/components/create-interactive-unit-dialog.tsx frontend/admin/app/(dashboard)/exams/components/interactive-unit-card.tsx
git commit -m "feat: add static-friendly admin interactive editor route"
```

---

### Task 5: Learner 前端切到运行时公开配置

**Files:**
- Create: `D:\Project\AI\FOCO\frontend\learner\lib\runtime-config.ts`
- Create: `D:\Project\AI\FOCO\frontend\learner\lib\runtime-config.test.ts`
- Modify: `D:\Project\AI\FOCO\frontend\learner\lib\auth-session.ts`
- Modify: `D:\Project\AI\FOCO\frontend\learner\components\auth\login-gate.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\page.tsx`

- [ ] **Step 1: 写 learner runtime-config 失败测试**

```ts
import { describe, expect, it, vi } from "vitest"
import { loadRuntimeConfig } from "./runtime-config"

describe("loadRuntimeConfig", () => {
  it("keeps registration_open and auth fields together", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({
        data: {
          registration_open: false,
          supabase_url: "https://demo.supabase.co",
          supabase_publishable_key: "pk-demo",
        },
      }),
    }))

    await expect(loadRuntimeConfig()).resolves.toMatchObject({
      registrationOpen: false,
      supabaseUrl: "https://demo.supabase.co",
      publishableKey: "pk-demo",
    })
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\learner
cmd /c npm.cmd test -- lib/runtime-config.test.ts
```

Expected: FAIL。

- [ ] **Step 3: 最小实现 runtime config，并让 login gate 统一从接口取配置**

实现要点：

```tsx
const [runtimeConfig, setRuntimeConfig] = React.useState<RuntimeConfig | null>(null)

React.useEffect(() => {
  loadRuntimeConfig().then(setRuntimeConfig).catch(() => setError(...))
}, [])
```

`auth-session.ts` 的 refresh 逻辑也使用 runtime config getter，而非 `process.env`。

- [ ] **Step 4: 运行 learner 测试和 typecheck**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\learner
cmd /c npm.cmd test -- lib/runtime-config.test.ts lib/auth-session.test.ts
cmd /c npm.cmd run typecheck
```

Expected: PASS。

- [ ] **Step 5: Commit**

```powershell
Set-Location D:\Project\AI\FOCO
git add frontend/learner/lib/runtime-config.ts frontend/learner/lib/runtime-config.test.ts frontend/learner/lib/auth-session.ts frontend/learner/components/auth/login-gate.tsx frontend/learner/app/page.tsx
git commit -m "feat: load learner auth config from backend runtime settings"
```

---

### Task 6: Learner 路由收口到静态友好 query 页面

**Files:**
- Create: `D:\Project\AI\FOCO\frontend\learner\lib\routes.ts`
- Create: `D:\Project\AI\FOCO\frontend\learner\lib\routes.test.ts`
- Create: `D:\Project\AI\FOCO\frontend\learner\components\practice\practice-session-page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\components\practice\practice-complete-page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\components\labs\lab-view-page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\session\page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\complete\page.tsx`
- Create: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\labs\view\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\[sessionId]\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\[sessionId]\complete\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\labs\[unitVersionId]\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\practice\setup\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\home\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\labs\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\profile\page.tsx`

- [ ] **Step 1: 写 learner route helper 失败测试**

```ts
import { describe, expect, it } from "vitest"
import { buildPracticeSessionHref, buildLabViewHref } from "./routes"

describe("learner routes", () => {
  it("builds static-friendly practice hrefs in demo mode", () => {
    expect(buildPracticeSessionHref("session-1", { demoMode: true }))
      .toBe("/practice/session?session_id=session-1")
  })

  it("builds static-friendly lab hrefs in demo mode", () => {
    expect(buildLabViewHref("unit-v1", { demoMode: true }))
      .toBe("/labs/view?unit_version_id=unit-v1")
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\learner
cmd /c npm.cmd test -- lib/routes.test.ts
```

Expected: FAIL。

- [ ] **Step 3: 抽取 practice/lab 组件并更新所有 learner 跳转**

实现要点：

```tsx
export function PracticeSessionPage({ sessionId }: { sessionId: string }) { ... }
export function PracticeCompletePage({ sessionId }: { sessionId: string }) { ... }
export function LabViewPage({ unitVersionId }: { unitVersionId: string }) { ... }
```

- 旧动态页改为 wrapper
- 新静态 query 页读取 `useSearchParams()`
- 所有 `router.push(`/practice/${sessionId}`)` / `href={`/labs/${id}`}` 改走 helper

- [ ] **Step 4: 运行 learner 测试和 typecheck**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\learner
cmd /c npm.cmd test -- lib/routes.test.ts
cmd /c npm.cmd run typecheck
```

Expected: PASS。

- [ ] **Step 5: Commit**

```powershell
Set-Location D:\Project\AI\FOCO
git add frontend/learner/lib/routes.ts frontend/learner/lib/routes.test.ts frontend/learner/components/practice frontend/learner/components/labs frontend/learner/app/(focus)/practice/session/page.tsx frontend/learner/app/(focus)/practice/complete/page.tsx frontend/learner/app/(learner-shell)/labs/view/page.tsx frontend/learner/app/(focus)/practice/[sessionId]/page.tsx frontend/learner/app/(focus)/practice/[sessionId]/complete/page.tsx frontend/learner/app/(learner-shell)/labs/[unitVersionId]/page.tsx frontend/learner/app/(learner-shell)/practice/setup/page.tsx frontend/learner/app/(learner-shell)/home/page.tsx frontend/learner/app/(learner-shell)/labs/page.tsx frontend/learner/app/(learner-shell)/profile/page.tsx
git commit -m "feat: add static-friendly learner demo routes"
```

---

### Task 7: 配置 demo 静态导出并让旧动态页 export-safe

**Files:**
- Modify: `D:\Project\AI\FOCO\frontend\admin\next.config.ts`
- Modify: `D:\Project\AI\FOCO\frontend\learner\next.config.ts`
- Modify: `D:\Project\AI\FOCO\frontend\admin\app\(dashboard)\interactive-units\[unitId]\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\[sessionId]\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(focus)\practice\[sessionId]\complete\page.tsx`
- Modify: `D:\Project\AI\FOCO\frontend\learner\app\(learner-shell)\labs\[unitVersionId]\page.tsx`

- [ ] **Step 1: 先运行 demo 构建，记录当前失败**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\admin
$env:FOCO_DEMO_BUILD='1'
cmd /c npm.cmd run build

Set-Location D:\Project\AI\FOCO\frontend\learner
$env:FOCO_DEMO_BUILD='1'
cmd /c npm.cmd run build
Remove-Item Env:FOCO_DEMO_BUILD
```

Expected: FAIL，报动态段无法 export、或 `/admin` 资源路径不正确。

- [ ] **Step 2: 写最小 export-safe 改造**

实现要点：

```ts
const isDemoBuild = process.env.FOCO_DEMO_BUILD === "1"

const nextConfig: NextConfig = {
  reactStrictMode: true,
  output: isDemoBuild ? "export" : undefined,
  basePath: isDemoBuild && isAdminApp ? "/admin" : undefined,
}
```

旧动态页 wrapper 增加最小静态导出兼容：

```ts
export const dynamicParams = false
export function generateStaticParams() { return [] }
```

如果 Next 版本要求至少一个 param，再退化为单个占位 param，并在组件内避免 demo 导航使用它。

- [ ] **Step 3: 重新运行 demo 构建直到成功**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\admin
$env:FOCO_DEMO_BUILD='1'
cmd /c npm.cmd run build

Set-Location D:\Project\AI\FOCO\frontend\learner
cmd /c npm.cmd run build
Remove-Item Env:FOCO_DEMO_BUILD
```

Expected: 两套前端都成功输出静态站点。

- [ ] **Step 4: 运行两端单元测试与 typecheck**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\frontend\admin
cmd /c npm.cmd test
cmd /c npm.cmd run typecheck

Set-Location D:\Project\AI\FOCO\frontend\learner
cmd /c npm.cmd test
cmd /c npm.cmd run typecheck
```

Expected: PASS。

- [ ] **Step 5: Commit**

```powershell
Set-Location D:\Project\AI\FOCO
git add frontend/admin/next.config.ts frontend/learner/next.config.ts frontend/admin/app/(dashboard)/interactive-units/[unitId]/page.tsx frontend/learner/app/(focus)/practice/[sessionId]/page.tsx frontend/learner/app/(focus)/practice/[sessionId]/complete/page.tsx frontend/learner/app/(learner-shell)/labs/[unitVersionId]/page.tsx
git commit -m "feat: enable demo static exports"
```

---

### Task 8: 增加 demo 打包脚本、smoke 验证与文档

**Files:**
- Modify: `D:\Project\AI\FOCO\.gitignore`
- Create: `D:\Project\AI\FOCO\backend\internal\demo\web\dist\.gitkeep`
- Create: `D:\Project\AI\FOCO\build-demo.ps1`
- Create: `D:\Project\AI\FOCO\build-demo.sh`
- Create: `D:\Project\AI\FOCO\test\scripts\smoke-demo.mjs`
- Modify: `D:\Project\AI\FOCO\test\package.json`
- Modify: `D:\Project\AI\FOCO\README.md`
- Modify: `D:\Project\AI\FOCO\test\README.md`

- [ ] **Step 1: 先写 smoke 脚本，让它在产物不存在时失败**

```js
import assert from "node:assert/strict"
import { spawn } from "node:child_process"

// 1. 启动 foco-demo(.exe)
// 2. 轮询 GET /, /admin, /api/v1/health
// 3. 断言状态码 200
```

- [ ] **Step 2: 运行 smoke，确认当前失败**

Run:

```powershell
Set-Location D:\Project\AI\FOCO\test
node .\scripts\smoke-demo.mjs
```

Expected: FAIL，提示找不到 `foco-demo.exe` 或启动失败。

- [ ] **Step 3: 实现 build-demo 脚本与生成目录管理**

脚本职责：

```text
1. 清理/创建 backend/internal/demo/web/dist
2. 设置 FOCO_DEMO_BUILD=1
3. 构建 learner 静态产物并复制到 dist/learner
4. 构建 admin 静态产物并复制到 dist/admin
5. go build -tags demo -o dist/foco-demo(.exe) ./backend/cmd/api
```

`.gitignore` 忽略 demo 生成物，但保留 `.gitkeep`。

- [ ] **Step 4: 运行完整 demo build + smoke**

Run:

```powershell
Set-Location D:\Project\AI\FOCO
.\build-demo.ps1

Set-Location D:\Project\AI\FOCO\test
node .\scripts\smoke-demo.mjs
```

Expected:

- `D:\Project\AI\FOCO\dist\foco-demo.exe` 存在
- `/`、`/admin`、`/api/v1/health` 全部返回 200

- [ ] **Step 5: 更新 README / test 文档并提交**

文档至少覆盖：

- demo 模式依赖的环境变量
- `build-demo.ps1` / `build-demo.sh` 用法
- 单文件运行方式
- smoke 验证方式

```powershell
Set-Location D:\Project\AI\FOCO
git add .gitignore backend/internal/demo/web/dist/.gitkeep build-demo.ps1 build-demo.sh test/scripts/smoke-demo.mjs test/package.json README.md test/README.md
git commit -m "build: add demo single executable packaging flow"
```

---

## Final Verification Checklist

- [ ] Backend platform tests

```powershell
Set-Location D:\Project\AI\FOCO\backend
go test ./internal/domain/platform -v
```

- [ ] Backend router tests

```powershell
Set-Location D:\Project\AI\FOCO\backend
go test ./internal/http -v
```

- [ ] Admin tests + typecheck

```powershell
Set-Location D:\Project\AI\FOCO\frontend\admin
cmd /c npm.cmd test
cmd /c npm.cmd run typecheck
```

- [ ] Learner tests + typecheck

```powershell
Set-Location D:\Project\AI\FOCO\frontend\learner
cmd /c npm.cmd test
cmd /c npm.cmd run typecheck
```

- [ ] Demo build

```powershell
Set-Location D:\Project\AI\FOCO
.\build-demo.ps1
```

- [ ] Demo smoke

```powershell
Set-Location D:\Project\AI\FOCO\test
node .\scripts\smoke-demo.mjs
```

- [ ] 手工验证刷新：
  - [ ] `/home`
  - [ ] `/practice/setup`
  - [ ] `/practice/session?session_id=<id>`
  - [ ] `/labs/view?unit_version_id=<id>`
  - [ ] `/admin/home`
  - [ ] `/admin/users`
  - [ ] `/admin/interactive-unit-editor?unit_id=<id>&exam_id=<id>`
