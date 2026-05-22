# FOCO 现场 Demo 脚本（5-10 分钟）

## 0. 开场（30 秒）

- 说明项目目标：高知识密度职业资格考试的智能备考平台
- 强调不是简单题库 CRUD，而是“题库 + 练习闭环 + 学习智能化”

## 1. 架构总览（1 分钟）

- 打开 `docs/ARCHITECTURE.md`
- 讲三层结构：Next.js 双前端、Go API、Supabase PostgreSQL
- 补充 Redis 二级缓存与交付包里的 Docker Compose / Nginx

## 2. 管理端：内容与知识图谱（2 分钟）

- 登录 Admin
- 展示考试树、题目筛选、版本化编辑
- 导入 `cfa.json` 内容包
- 打开知识图谱弹窗，说明知识点依赖关系与任务选择基础

## 3. 学习端：诊断 -> 今日路径 -> 练习 -> 错题本（2-3 分钟）

- Learner 登录
- 进入诊断测评并提交
- 回到首页，展示：
  - 今日学习路径
  - 最近 7 天节奏
  - streak / XP / 推荐理由
- 进入练习，提交答案，展示即时反馈、解析、正确答案、耗时
- 跳到错题本，说明错误题自动归档

## 4. 交互式学习单元（1-2 分钟）

- 打开 Labs
- 进入一个交互单元
- 演示 1-2 个步骤的即时反馈
- 完成后展示总结 / concept card

## 5. 数据与扩展性（1 分钟）

- 打开 `docs/DATABASE_DESIGN.md`
- 解释：
  - 题库版本化
  - `practice_sessions` / `practice_session_items`
  - `streaks`
  - 交互单元相关表

## 6. 加分项与工程质量（1 分钟）

- 打开 `docs/SCORECARD.md`
- 强调前三个加分项：
  1. 三个进阶方向全部实现
  2. 知识图谱可视化 + 推荐解释字段
  3. 性能优化 / 安全设计 / CI-CD / 监控设计
- 展示根级命令：
  - `npm run lint`
  - `npm run test`
  - `npm run build`

## 7. 收尾（30 秒）

- 总结：该项目已经形成可演示、可回归、可扩展的学习闭环
- 如果面试官追问，可继续展开：
  - 为什么做 session snapshot
  - 为什么用 namespace cache invalidation
  - 推荐理由如何保证可解释
  - 为什么交互单元用 JSON schema + 版本化
