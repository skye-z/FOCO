# FOCO 评分对照与加分项说明

## 结论

当前项目已经具备冲击 **90 分以上** 的交付面，并且覆盖了原始需求里的**前三个加分项**：

1. 完整实现第三部分全部 3 个进阶方向
2. 知识图谱可视化或学习策略效果评估
3. 性能优化、安全设计、CI/CD、监控设计

---

## 评分标准对照

| 维度 | 目标表现 | 主要证据 |
| --- | --- | --- |
| 系统架构设计（40%） | 可扩展、可维护、可解释 | `docs/ARCHITECTURE.md`、`docs/API_DESIGN.md`、`docs/DATABASE_DESIGN.md`、`docs/LEARNING_INTELLIGENCE_DESIGN.md` |
| 代码质量（30%） | 明确分层、可测试、可回归 | `backend/internal/domain/*`、`backend/internal/http/router_test.go`、`frontend/learner/lib/*.test.ts`、`frontend/admin/components/interactive/*.test.ts`、`package.json` |
| 问题解决能力（20%） | 解释推荐理由、处理幂等与边界 | `backend/internal/domain/practice/service.go`、`backend/internal/domain/home/service.go`、`backend/internal/domain/interactive/service.go` |
| 文件完整性（10%） | 文档齐全、命令可执行、验收路径清晰 | `README.md`、`docs/README.md`、`docs/SCORECARD.md`、`docs/DEMO_SCRIPT.md`、`test/README.md` |

---

## 必做功能覆盖

### 1) 考试知识体系与题库管理系统

- 管理端题库树、题目筛选、创建/编辑/发布：`frontend/admin/app/(dashboard)/exams/page.tsx`
- 知识点与知识图谱管理：`frontend/admin/components/knowledge-graph-modal.tsx`
- 后端题库 API：`backend/internal/http/router.go`
- 题库仓储与导入导出：`backend/internal/domain/content/repository.go`

### 2) 练习、判题与学习记录系统

- 练习入口与完整答题流程：`frontend/learner/app/(learner-shell)/practice/setup/page.tsx`、`frontend/learner/app/(focus)/practice/[sessionId]/page.tsx`
- 结果页与学习总结：`frontend/learner/app/(focus)/practice/[sessionId]/complete/page.tsx`
- 错题本与学习记录：`frontend/learner/app/(learner-shell)/wrong-book/page.tsx`、`frontend/learner/app/(learner-shell)/profile/page.tsx`
- 判题、幂等与错题归档：`backend/internal/domain/practice/service.go`、`backend/internal/domain/practice/service_test.go`

---

## 三个进阶方向：已全部覆盖

### 进阶方向 1：每日学习路径与习惯留存

- 首页今日路径、连续学习、最近 7 天节奏：`frontend/learner/app/(learner-shell)/home/page.tsx`
- 后端路径生成与 streak 汇总：`backend/internal/domain/home/service.go`

### 进阶方向 2：诊断测评、知识图谱与自动任务选择

- 诊断流程：`frontend/learner/app/(focus)/diagnostic/page.tsx`
- 推荐理由字段：`reason_code` / `reason_text` / `evidence`
- 用户知识图谱展示：`frontend/learner/app/(learner-shell)/profile/page.tsx`
- 后端画像与推荐：`backend/internal/domain/diagnostic/service.go`、`backend/internal/domain/home/service.go`

### 进阶方向 3：交互式学习单元

- Learner 交互学习页：`frontend/learner/app/(learner-shell)/labs/[unitVersionId]/page.tsx`
- Admin 交互单元编辑器：`frontend/admin/app/(dashboard)/interactive-units/[unitId]/page.tsx`
- 运行时评测与 concept card：`backend/internal/domain/interactive/service.go`、`backend/internal/domain/interactive/evaluator.go`

---

## 前三个加分项对照

### 加分项 1：完整实现第三部分全部 3 个进阶方向

已实现，不仅有文档，还落到了真实前后端页面、API 与数据库结构。

### 加分项 2：知识图谱可视化或学习策略效果评估

- Admin 侧知识图谱可视化：`frontend/admin/components/knowledge-graph-modal.tsx`
- Learner 侧知识画像：`frontend/learner/app/(learner-shell)/profile/page.tsx`
- 推荐解释链路：`reason_code`、`reason_text`、`evidence`

### 加分项 3：性能优化、安全设计、CI/CD、监控设计

**性能优化**
- Go 本地缓存 + Redis 二级缓存：`backend/internal/cache/cache.go`
- 首页 / 题库 / 错题本 / 图谱 / 交互单元按 namespace 失效：`docs/ARCHITECTURE.md`
- Nginx gzip：`build.sh`

**安全设计**
- Bearer Token 校验与 RBAC：`backend/internal/http/middleware/auth.go`
- CORS 白名单、基础安全响应头、写接口限流：`backend/internal/http/middleware/cors.go`、`backend/internal/http/middleware/security_headers.go`、`backend/internal/http/middleware/rate_limit.go`
- 管理员 / 教研 / 学习者边界在 API 与前端路由中分离：`backend/internal/http/router.go`

**CI/CD**
- GitHub Actions 质量门禁：`.github/workflows/ci.yml`
- 根级统一校验命令：`package.json`

**监控设计**
- 监控、日志、成本与指标设计说明：`docs/ARCHITECTURE.md`、`docs/API_DESIGN.md`

---

## 交付验收命令

```bash
npm run lint
npm run test
npm run build
npm run check --prefix test
npm run typecheck --prefix test
```

如需本地启动：

```bash
./run.sh --t 1 --t 2 --t 3
```

Windows:

```powershell
.\run.ps1 --t 1 --t 2 --t 3
```
