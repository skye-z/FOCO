# API 设计

## 约定

- 认证：`Authorization: Bearer <supabase_access_token>`
- 响应包裹：`{ "data": ..., "meta": {}, "error": null }`
- 错误包裹：`{ "error": "message" }`
- GET 用查询参数，写操作用 JSON body。

## 角色

- 学习者：学习首页、诊断、练习、错题本、画像、交互单元。
- 教研 / 内容编辑：题库、知识点、知识图谱、交互单元管理。
- 管理员：用户、平台设置、平台概览。

## 认证与授权

- 登录由 Supabase 负责。
- 前端保存 access token，API 通过 middleware 验证。
- 路由层通过当前用户 claims 读取 `user_id` 与 `roles`。
- 管理端操作依赖管理员角色或服务端 Supabase 管理能力。

## 错误规范

| HTTP | 场景 |
| --- | --- |
| 400 | 请求参数不合法、题型 / JSON 格式错误 |
| 401 | 未登录或 token 无效 |
| 404 | 资源不存在 |
| 409 | 状态冲突，例如已发布版本只读 |
| 500 | 其他服务端错误 |

## Rate limiting

- 当前代码没有独立限流中间件。
- 交付设计建议按用户与接口分层限流：
  - 登录与刷新：更严格。
  - 练习提交：中等限流，避免重复提交风暴。
  - 读取类接口：宽松限流。

## 公共接口

| Method | Path | 作用 |
| --- | --- | --- |
| GET | `/api/v1/health` | 健康检查 |
| POST | `/api/v1/seed/admin` | 初始化默认管理员 |
| GET | `/api/v1/public/settings` | 公共注册开关 |

## 学习者接口

| Method | Path | 作用 |
| --- | --- | --- |
| GET | `/api/v1/me` | 当前用户信息与 active enrollment |
| POST | `/api/v1/learner/bootstrap` | 首次学习者初始化 |
| GET | `/api/v1/learner/exams` | 可选考试列表 |
| POST | `/api/v1/learner/exam-enrollments` | 报名 / 切换考试 |
| GET | `/api/v1/learner/diagnostic/current` | 当前诊断状态 |
| POST | `/api/v1/learner/diagnostic/restart` | 重开诊断 |
| POST | `/api/v1/learner/diagnostic/{attemptId}/submit` | 提交诊断答案 |
| GET | `/api/v1/learner/home` | 今日路径、周节奏、推荐、进度 |
| GET | `/api/v1/learner/recommendations` | 推荐任务列表 |
| GET | `/api/v1/learner/exams/{examId}/content-tree` | 学习内容树 |
| GET | `/api/v1/learner/profile` | 用户画像、记录、知识图谱 |
| GET | `/api/v1/learner/wrong-book` | 错题本 |
| POST | `/api/v1/learner/practice-sessions` | 创建练习 session |
| GET | `/api/v1/learner/practice-sessions/{sessionId}` | 获取 session 详情 |
| POST | `/api/v1/learner/practice-sessions/{sessionId}/items/{itemId}/submit` | 提交单题答案 |
| GET | `/api/v1/learner/practice-sessions/{sessionId}/summary` | 练习结果汇总 |
| GET | `/api/v1/learner/interactive-units` | 交互单元列表 |
| GET | `/api/v1/learner/interactive-units/{unitVersionId}` | 交互单元详情 |
| POST | `/api/v1/learner/interactive-units/{unitVersionId}/attempts` | 开始交互单元 |
| POST | `/api/v1/learner/interactive-unit-attempts/{attemptId}/steps/{stepId}/actions` | 提交步骤动作 |
| POST | `/api/v1/learner/interactive-unit-attempts/{attemptId}/complete` | 完成交互单元 |

## 管理端接口

| Method | Path | 作用 |
| --- | --- | --- |
| GET | `/api/v1/admin/stats` | 平台概览 |
| GET | `/api/v1/admin/settings` | 管理设置 |
| PATCH | `/api/v1/admin/settings` | 保存设置 |
| GET | `/api/v1/admin/users` | 用户列表 |
| POST | `/api/v1/admin/users/{userId}/roles` | 补角色 |
| POST | `/api/v1/admin/users/{userId}/disable` | 禁用用户 |
| POST | `/api/v1/admin/users/{userId}/reset-password` | 重置密码 |
| GET | `/api/v1/admin/exam-tree` | 题库树 |
| GET | `/api/v1/admin/content-package/export` | 导出内容包 |
| POST | `/api/v1/admin/content-package/import` | 导入内容包 |
| GET | `/api/v1/admin/knowledge-points` | 知识点列表 |
| GET | `/api/v1/admin/knowledge-graph` | 知识图谱 |
| GET | `/api/v1/admin/questions` | 题目列表 |
| GET | `/api/v1/admin/questions/{questionId}/versions` | 题目版本列表 |
| POST | `/api/v1/admin/exams` | 新建考试 |
| PATCH | `/api/v1/admin/exams/{examId}` | 重命名考试 |
| DELETE | `/api/v1/admin/exams/{examId}` | 删除考试 |
| POST | `/api/v1/admin/subjects` | 新建科目 |
| PATCH | `/api/v1/admin/subjects/{subjectId}` | 重命名科目 |
| DELETE | `/api/v1/admin/subjects/{subjectId}` | 删除科目 |
| POST | `/api/v1/admin/chapters` | 新建章节 |
| PATCH | `/api/v1/admin/chapters/{chapterId}` | 重命名章节 |
| DELETE | `/api/v1/admin/chapters/{chapterId}` | 删除章节 |
| POST | `/api/v1/admin/knowledge-points` | 新建知识点 |
| GET | `/api/v1/admin/knowledge-point-edges` | 知识点边 |
| POST | `/api/v1/admin/knowledge-point-edges` | 新建知识点边 |
| POST | `/api/v1/admin/questions` | 新建题目 |
| DELETE | `/api/v1/admin/questions/{questionId}` | 删除题目 |
| POST | `/api/v1/admin/questions/{questionId}/versions` | 新建题目版本 |
| GET | `/api/v1/admin/question-versions/{versionId}` | 版本详情 |
| PATCH | `/api/v1/admin/question-versions/{versionId}` | 保存版本草稿 |
| POST | `/api/v1/admin/question-versions/{versionId}/restore` | 还原为新草稿 |
| POST | `/api/v1/admin/question-versions/{versionId}/publish` | 发布版本 |
| GET | `/api/v1/admin/interactive-units` | 交互单元列表 |
| POST | `/api/v1/admin/interactive-units` | 新建交互单元 |
| GET | `/api/v1/admin/interactive-units/{unitId}/versions` | 版本列表 |
| POST | `/api/v1/admin/interactive-units/{unitId}/versions` | 新建版本 |
| GET | `/api/v1/admin/interactive-unit-versions/{versionId}` | 版本详情 |
| PATCH | `/api/v1/admin/interactive-unit-versions/{versionId}` | 更新版本 |
| POST | `/api/v1/admin/interactive-unit-versions/{versionId}/publish` | 发布交互版本 |
| DELETE | `/api/v1/admin/interactive-units/{unitId}` | 删除单元 |

## 代表性请求

### 创建练习 session

```json
{
  "exam_id": "73000000-0000-0000-0000-000000000101",
  "mode": "intelligent",
  "question_types": ["single_choice", "multiple_choice"],
  "difficulty": "medium",
  "count": 15,
  "subject_ids": [],
  "chapter_ids": [],
  "knowledge_point_ids": []
}
```

### 提交单题答案

```json
{
  "answer": ["A"],
  "duration_seconds": 38
}
```

### 交互单元步骤动作

```json
{
  "selected": ["A", "C"],
  "reason": "当前步骤判断前置条件"
}
```

