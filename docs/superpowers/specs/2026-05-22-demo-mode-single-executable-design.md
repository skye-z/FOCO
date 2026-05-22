# FOCO Demo 模式单可执行文件设计

日期：2026-05-22  
状态：已评审，待用户确认后进入 implementation plan

## 1. 目标

为后端增加一个 **demo 模式**，在该模式下：

- 学生端页面由后端统一托管，入口为 `/`
- 管理端页面由后端统一托管，入口为 `/admin`
- API 继续由后端提供，入口为 `/api/v1`
- 最终交付为 **单个 Go 可执行文件**

非 demo 模式必须继续保持当前开发/构建链路可用，不影响现有 `run.sh`、`run.ps1`、`build.sh` 的常规用途。

## 2. 范围

### In scope

- 新增 demo 模式开关
- learner/admin 两套前端的 demo 静态构建
- Go 后端嵌入静态资源并按路径路由
- 新增运行时公开配置接口，替代 demo 静态站点对构建时环境变量的依赖
- 新增 demo 单文件打包脚本
- 为 demo 路由和打包增加基本测试/验收

### Out of scope

- 替换现有 docker-compose 正式交付方式
- 在单文件内嵌 Node/Next 运行时
- 为正式模式引入 SSR/Node 依赖
- 重做 Supabase 鉴权流程

## 3. 设计原则

1. **双模式共存**：正常模式与 demo 模式分离，避免互相污染
2. **单文件真实成立**：最终运行不依赖 Node、Nginx、额外前端进程
3. **前端改动尽量收口**：通过 route helper 和 runtime config 减少散点判断
4. **后端优先保护 API**：静态站点绝不能覆盖 `/api/v1/**`
5. **build-time 与 run-time 分离**：demo 静态资源在构建时生成，公开配置在运行时注入

## 4. 顶层架构

### 4.1 模式开关

引入 demo 模式开关，建议：

- 运行时环境变量：`FOCO_DEMO_MODE=true`
- 编译时 build tag：`demo`

职责分离：

- `FOCO_DEMO_MODE`：控制运行时是否启用 demo 站点路由和公开配置
- `-tags demo`：控制是否把前端静态资源嵌入到二进制

### 4.2 运行形态

#### 普通模式

- 与当前行为一致
- 前后端仍可独立开发和部署

#### Demo 模式

- Go API 监听单一端口
- 路由同时提供：
  - learner 静态站点
  - admin 静态站点
  - `/api/v1/**`

## 5. URL 方案

### 5.1 对外 URL

- learner：`/`
- admin：`/admin`
- API：`/api/v1`

### 5.2 路由优先级

优先级必须固定为：

1. `/api/v1/**`
2. `/admin` 与 `/admin/**`
3. learner 根站点 `/` 与其余前端路由

## 6. 前端方案

## 6.1 为什么不能直接原样静态导出

当前前端存在 Next App Router 动态段，例如：

- admin：`/interactive-units/[unitId]`
- learner：`/practice/[sessionId]`
- learner：`/practice/[sessionId]/complete`
- learner：`/labs/[unitVersionId]`

在 demo 模式的“纯静态导出 + Go 托管”里，这些路径不应继续依赖动态段。

## 6.2 Demo 友好路由收敛

### Admin

将 demo 访问统一收敛到静态页面：

- `/admin/interactive-unit-editor?unit_id=...&exam_id=...`

现有业务上仍可保留普通模式路径，但 demo 构建和页面跳转必须走 helper 生成的新路径。

### Learner

将 demo 访问统一收敛为静态页面 + query：

- `/practice/session?session_id=...`
- `/practice/complete?session_id=...`
- `/labs/view?unit_version_id=...`

保留以下稳定静态路径：

- `/home`
- `/practice/setup`
- `/profile`
- `/wrong-book`
- `/diagnostic`
- `/onboarding`

## 6.3 Route helper

新增：

- `frontend/admin/lib/routes.ts`
- `frontend/learner/lib/routes.ts`

统一提供：

- URL 生成函数
- query 读取函数
- demo/普通模式兼容封装

目标：

- 业务组件不直接拼接 demo 路径
- 组件不关心参数来自 path 还是 query
- 后续若再调整路径，只修改 helper

## 6.4 Demo 构建配置

### Learner

demo 构建时：

- 启用静态导出
- 继续部署在根路径 `/`

### Admin

demo 构建时：

- 启用静态导出
- 配置 `basePath: "/admin"`

这样 admin 的 HTML、JS、CSS、字体、`_next` 资源都会自然落在 `/admin/**` 下，避免与 learner 冲突。

## 6.5 运行时公开配置

当前登录页从 `process.env` 读取公开 Supabase 参数。  
在 demo 静态导出中，这会固化为构建时值，不适合单文件运行。

因此改为扩展现有公开配置接口：

- `GET /api/v1/public/settings`

返回可公开信息：

- `demo_mode`
- `supabase_url`
- `supabase_publishable_key`
- `registration_open`

前端在启动认证流程前先读取该接口。

要求：

- 只返回前端可公开字段
- 严禁返回 service role key 或其他敏感配置
- 保持向后兼容；已有 `registration_open` 字段继续保留

## 6.6 动态页 export-safe 要求

即使 canonical demo 导航已收敛到 query 路由，仓库中的动态页文件仍可能影响静态导出。

因此 demo 构建必须满足以下至少一种策略：

1. **动态页不参与 demo 构建**；或
2. **动态页具备 export-safe 能力**，例如提供最小 `generateStaticParams` 占位实现，使 `output: "export"` 构建可通过。

设计要求：

- demo 模式下的真实导航只使用静态友好 query 路由
- 旧动态页即使保留，也只能作为普通模式兼容入口
- 不能让现有 `[unitId]`、`[sessionId]`、`[unitVersionId]` 阻塞 demo 导出

## 7. 后端方案

## 7.1 嵌入资源目录

建议新增：

- `backend/internal/demo/web`
- `backend/internal/demo/web/dist/learner`
- `backend/internal/demo/web/dist/admin`

其中：

- `dist/learner`：learner demo 静态导出产物
- `dist/admin`：admin demo 静态导出产物

通过 `go:embed` 仅在 `demo` build tag 下编译进入二进制。

## 7.2 Router 组织

在现有 `NewRouter` 基础上扩展：

### API

保留：

- `/api/v1/**`

扩展：

- `GET /api/v1/public/settings`

### Admin 站点

- `/admin`
- `/admin/**`

行为：

- 命中真实静态文件 => 返回文件
- 命中前端路由 => 返回 admin 入口 HTML

### Learner 站点

- `/`
- 其他非 `/api/v1/**`、非 `/admin/**` 的前端路由

行为：

- 命中真实静态文件 => 返回文件
- 命中前端路由 => 返回 learner 入口 HTML

## 7.3 SPA fallback 规则

必须避免把资源请求错误 fallback 成 HTML。

### 严格按资源处理的路径

这些路径如果找不到，必须返回 404：

- `/_next/**`
- `/admin/_next/**`
- `/favicon.ico`
- `/assets/**`
- `/images/**`
- 任何带扩展名的文件请求（`.js`、`.css`、`.png`、`.svg`、`.woff2` 等）

### 允许 fallback 的路径

仅无扩展名的前端路由允许 fallback，例如：

- `/home`
- `/practice/setup`
- `/practice/session`
- `/labs/view`
- `/admin/users`
- `/admin/settings`
- `/admin/interactive-unit-editor`

## 7.4 Cache-Control

建议：

- HTML：`Cache-Control: no-store`
- 指纹静态资源：`Cache-Control: public, max-age=31536000, immutable`

这样可以避免旧 HTML 卡住，同时保留资源缓存优势。

## 8. Demo 构建与打包

## 8.1 新增独立脚本

建议新增：

- `build-demo.sh`
- `build-demo.ps1`

不要直接复用现有 `build.sh`，避免影响现有 Docker 交付链。

## 8.2 构建流程

1. 构建 learner demo 静态站点
2. 构建 admin demo 静态站点（`basePath=/admin`）
3. 将产物复制到：
   - `backend/internal/demo/web/dist/learner`
   - `backend/internal/demo/web/dist/admin`
4. 执行 `go build -tags demo`
5. 输出单文件：
   - Windows：`foco-demo.exe`
   - Linux：`foco-demo`

## 8.3 对现有链路的影响

- `run.sh` / `run.ps1`：保留现状
- `build.sh`：保留 Docker 交付逻辑
- demo 打包脚本：只服务于“单文件演示版”

## 9. 测试与验收

## 9.1 后端测试

新增 router 相关测试，至少覆盖：

- `/api/v1/**` 不被静态站点覆盖
- `/admin/**` 正确走 admin
- 普通 learner 路由正确 fallback
- `/_next/**` 与 `/admin/_next/**` 不错误 fallback
- 缺失资源返回 404

## 9.2 前端测试

新增最小必要测试：

- route helper 生成 URL 正确
- query 参数解析正确
- runtime config 加载逻辑正确

## 9.3 集成 smoke

至少验证：

1. demo 单文件构建成功
2. 启动 demo 单文件
3. `GET /`
4. `GET /admin`
5. `GET /api/v1/health`

上述请求都必须成功。

## 9.4 手工验收流

### Learner

- 登录页可打开
- 登录成功
- 进入首页
- 进入练习设置
- 创建练习 session
- 打开练习页
- 打开练习完成页
- 打开互动单元页

### Admin

- 登录页可打开
- 登录成功
- 打开首页
- 打开用户页
- 打开设置页
- 打开互动单元编辑页

## 10. 风险与控制

### 风险 1：旧路径散落在业务组件中

症状：

- 页面跳转仍访问旧动态段
- 刷新后 404

控制：

- 路由生成统一通过 helper
- 尽量不在业务组件中直接拼 URL

### 风险 2：admin basePath 资源错配

症状：

- `/admin` HTML 正常，但 JS/CSS 404

控制：

- admin demo 构建强制带 `basePath: "/admin"`
- 实际浏览器验证 `/admin` 深链刷新

### 风险 3：运行时配置改造不完整

症状：

- demo 静态页使用了构建时环境变量
- 运行时更换环境变量不生效

控制：

- 认证入口统一走扩展后的 `public/settings`
- 搜索并清理散落的公开环境变量依赖

### 风险 4：fallback 过宽

症状：

- `.js` / `.css` 请求返回 HTML
- 浏览器 MIME type 报错

控制：

- 仅无扩展名前端路由允许 fallback
- 资源请求严格 404

### 风险 5：单文件体积增长

控制：

- 仅在 `demo` build tag 下嵌入资源
- 生产构建压缩前端产物
- 普通构建链路不受影响

## 11. 实施顺序（高层）

1. 增加 runtime config 与 demo router 壳层
2. 增加前端 route helper
3. 将动态页访问收敛到静态友好 URL
4. 增加 admin/learner demo 构建配置
5. 增加 `go:embed` 静态资源托管
6. 增加 demo 打包脚本
7. 增加测试与 smoke 验证

## 12. 成功标准

当以下条件同时满足时，视为成功：

- 通过单个 Go 可执行文件访问 learner `/`
- 通过单个 Go 可执行文件访问 admin `/admin`
- `/api/v1/**` 保持可用
- learner/admin 深链刷新不 404
- demo 构建不依赖运行时 Node/Nginx
- 普通开发与正式交付链路不被破坏
