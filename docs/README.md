# FOCO 交付说明

这是一个面向 CFA / FRM / CPA / 法考等高知识密度考试的智能备考平台交付包。

## 交付内容

- [ARCHITECTURE.md](./ARCHITECTURE.md)
- [API_DESIGN.md](./API_DESIGN.md)
- [DATABASE_DESIGN.md](./DATABASE_DESIGN.md)
- [LEARNING_INTELLIGENCE_DESIGN.md](./LEARNING_INTELLIGENCE_DESIGN.md)
- [DESIGN_DECISIONS.md](./DESIGN_DECISIONS.md)
- [FOCO.postman_collection.json](./FOCO.postman_collection.json)
- [../.env.example](../.env.example)
- [../test/README.md](../test/README.md)

## 代码结构

- `backend/`：Go API，Supabase Auth + PostgreSQL 数据访问，xorm 仓储实现。
- `frontend/admin/`：教研 / 管理后台。
- `frontend/learner/`：学习端。
- `test/`：交付 E2E 验收脚本。
- `../build.sh`：生成可交付产物到 `dist/`，包含 Docker Compose。
- `../run.sh`：本地开发启动脚本。

## 快速开始

1. 复制 `.env.example` 为 `../.env` 并填入 Supabase、Redis 与测试账号信息。
2. 管理后台里的注册开关、LLM 配置等平台设置存储在数据库 `admin_settings` 表里，不在环境变量里。
3. 本地开发需先确保 `REDIS_URL` 指向可用 Redis，例如 `redis://127.0.0.1:6379/0`。
4. 本地开发：

```bash
./run.sh --t 1 --t 2 --t 3
```

5. 生成交付包：

```bash
./build.sh
```

`build.sh` 会生成 `dist/backend/api`、`dist/frontend/*`、`dist/nginx/nginx.conf` 和 `dist/docker-compose.yml`。交付 compose 内置 `redis:7.2-alpine`，首次 `docker compose up` 会拉取本地 Redis 镜像。

## 缓存说明

- L1：Go 进程内短 TTL 缓存。
- L2：Redis，通过 `REDIS_URL` 连接。
- 失效：按 namespace 版本号统一 bump，不做全表扫 key。
- 管理页设置仍落库在 `admin_settings`，不放环境变量。

## 访问地址

- 学习端：`http://localhost:3000`
- 管理端：`http://localhost:3001`
- API：`http://localhost:8080`

## 验收

交付 E2E 说明见 [test/README.md](./test/README.md)。
