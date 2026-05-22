#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DIST_DIR="$SCRIPT_DIR/dist"
SRC_BACKEND="$SCRIPT_DIR/backend"
SRC_ADMIN="$SCRIPT_DIR/frontend/admin"
SRC_LEARNER="$SCRIPT_DIR/frontend/learner"
ENV_FILE="$SCRIPT_DIR/../.env"

GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"

step() { echo ""; echo "==> $1"; }
info() { echo "    $1"; }

# ──────────────────────────────────────
# 1. Clean & prepare dist/
# ──────────────────────────────────────
step "Preparing dist/"
rm -rf "$DIST_DIR"
mkdir -p "$DIST_DIR"/{backend,nginx,frontend}

# ──────────────────────────────────────
# 2. Build Go backend
# ──────────────────────────────────────
step "Building backend ($GOOS_TARGET/$GOARCH_TARGET)"
(
  cd "$SRC_BACKEND"
  CGO_ENABLED=0 GOOS="$GOOS_TARGET" GOARCH="$GOARCH_TARGET" \
    go build -trimpath -ldflags="-s -w" -o "$DIST_DIR/backend/api" ./cmd/api/
  cp -r db "$DIST_DIR/backend/db"
)
info "backend/api  $(du -sh "$DIST_DIR/backend/api" | cut -f1)"

# ──────────────────────────────────────
# 3. Copy frontend sources for Docker builds
# ──────────────────────────────────────
for APP in admin learner; do
  step "Packaging frontend/$APP source"
  SRC="$SCRIPT_DIR/frontend/$APP"
  DST="$DIST_DIR/frontend/$APP"
  mkdir -p "$DST"

  cp "$SRC/package.json" "$DST/"
  cp "$SRC/package-lock.json" "$DST/" 2>/dev/null || true
  cp "$SRC/next.config.ts" "$DST/"
  cp "$SRC/tsconfig.json" "$DST/"
  cp "$SRC/tailwind.config.js" "$DST/"
  cp "$SRC/postcss.config.mjs" "$DST/"
  cp "$SRC/components.json" "$DST/" 2>/dev/null || true

  for DIR in app components lib; do
    if [ -d "$SRC/$DIR" ]; then
      cp -r "$SRC/$DIR" "$DST/$DIR"
    fi
  done

  if [ -d "$SRC/public" ]; then
    cp -r "$SRC/public" "$DST/public"
  else
    mkdir -p "$DST/public"
  fi
done

# ──────────────────────────────────────
# 4. Copy seed data
# ──────────────────────────────────────
step "Copying seed data"
cp "$SCRIPT_DIR/cfa.json" "$DIST_DIR/backend/cfa.json"

# ──────────────────────────────────────
# 5. Nginx config
# ──────────────────────────────────────
step "Generating nginx/nginx.conf"
mkdir -p "$DIST_DIR/nginx"
cat > "$DIST_DIR/nginx/nginx.conf" <<'NGINX'
worker_processes auto;

events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/json;

    sendfile    on;
    tcp_nopush  on;
    keepalive_timeout 65;
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml text/javascript;

    upstream learner {
        server learner:3000;
    }

    upstream admin {
        server admin:3001;
    }

    upstream api {
        server api:8080;
    }

    # ── Learner (default host :80) ──
    server {
        listen 80;
        server_name _;

        location /api/v1/ {
            proxy_pass         http://api;
            proxy_set_header   Host              $host;
            proxy_set_header   X-Real-IP         $remote_addr;
            proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
            proxy_set_header   X-Forwarded-Proto $scheme;
        }

        location / {
            proxy_pass         http://learner;
            proxy_set_header   Host              $host;
            proxy_set_header   X-Real-IP         $remote_addr;
            proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
            proxy_set_header   X-Forwarded-Proto $scheme;
        }
    }

    # ── Admin (:3001) ──
    server {
        listen 3001;
        server_name _;

        location /api/v1/ {
            proxy_pass         http://api;
            proxy_set_header   Host              $host;
            proxy_set_header   X-Real-IP         $remote_addr;
            proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
            proxy_set_header   X-Forwarded-Proto $scheme;
        }

        location / {
            proxy_pass         http://admin;
            proxy_set_header   Host              $host;
            proxy_set_header   X-Real-IP         $remote_addr;
            proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
            proxy_set_header   X-Forwarded-Proto $scheme;
        }
    }
}
NGINX

# ──────────────────────────────────────
# 6. Backend Dockerfile
# ──────────────────────────────────────
step "Generating backend/Dockerfile"
cat > "$DIST_DIR/backend/Dockerfile" <<'DOCKERFILE'
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY api      /app/api
COPY db       /app/db
COPY cfa.json /app/cfa.json

EXPOSE 8080
ENTRYPOINT ["/app/api"]
DOCKERFILE

# ──────────────────────────────────────
# 7. Frontend Dockerfile (shared)
# ──────────────────────────────────────
step "Generating frontend/Dockerfile"
cat > "$DIST_DIR/frontend/Dockerfile" <<'DOCKERFILE'
FROM node:20-alpine AS builder
WORKDIR /app

ARG APP_DIR
COPY ${APP_DIR}/package.json ${APP_DIR}/package-lock.json* ./
RUN npm install --no-fund --no-audit

COPY ${APP_DIR}/ .
RUN npm run build

# ── runtime ──
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production

ARG APP_PORT=3000
EXPOSE ${APP_PORT}

COPY --from=builder /app/.next   .next
COPY --from=builder /app/public  public
COPY --from=builder /app/package.json .
COPY --from=builder /app/node_modules node_modules
COPY --from=builder /app/next.config.ts .

CMD ["npx", "next", "start"]
DOCKERFILE

# ──────────────────────────────────────
# 8. Nginx Dockerfile
# ──────────────────────────────────────
step "Generating nginx/Dockerfile"
cat > "$DIST_DIR/nginx/Dockerfile" <<'DOCKERFILE'
FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80 3001
DOCKERFILE

# ──────────────────────────────────────
# 9. docker-compose.yml
# ──────────────────────────────────────
step "Generating docker-compose.yml"
cat > "$DIST_DIR/docker-compose.yml" <<'COMPOSE'
services:
  redis:
    image: redis:7.2-alpine
    command: ["redis-server", "--appendonly", "yes"]
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    restart: unless-stopped

  api:
    build:
      context: ./backend
      dockerfile: Dockerfile
    environment:
      GO_API_PORT: "8080"
      SUPABASE_URL: ${SUPABASE_URL}
      SUPABASE_DB_URL: ${SUPABASE_DB_URL}
      NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY: ${NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY}
      SUPABASE_SERVICE_ROLE_KEY: ${SUPABASE_SERVICE_ROLE_KEY}
      REDIS_URL: ${REDIS_URL:-redis://redis:6379/0}
    depends_on:
      - redis
    restart: unless-stopped

  learner:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        APP_DIR: learner
        APP_PORT: "3000"
    environment:
      NEXT_PUBLIC_SUPABASE_URL: ${SUPABASE_URL}
      NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY: ${NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY}
    restart: unless-stopped

  admin:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        APP_DIR: admin
        APP_PORT: "3001"
    environment:
      NEXT_PUBLIC_SUPABASE_URL: ${SUPABASE_URL}
      NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY: ${NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY}
    restart: unless-stopped

  nginx:
    build:
      context: ./nginx
      dockerfile: Dockerfile
    ports:
      - "80:80"
      - "3001:3001"
    depends_on:
      - api
      - learner
      - admin
    restart: unless-stopped

volumes:
  redis-data:
COMPOSE

# ──────────────────────────────────────
# 10. .env.example
# ──────────────────────────────────────
step "Generating .env.example"
cat > "$DIST_DIR/.env.example" <<'ENV'
# ── Supabase ──
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_DB_URL=postgres://postgres:password@db.your-project.supabase.co:5432/postgres
NEXT_PUBLIC_SUPABASE_PUBLISHABLE_KEY=eyJ...
SUPABASE_SERVICE_ROLE_KEY=eyJ...

# ── Cache ──
REDIS_URL=redis://redis:6379/0
ENV

# ──────────────────────────────────────
# Done
# ──────────────────────────────────────
echo ""
echo "========================================="
echo "  Build complete → $DIST_DIR/"
echo "========================================="
echo ""
echo "  Deploy with Docker Compose:"
echo "    cd $DIST_DIR"
echo "    cp .env.example .env   # fill in your credentials"
echo "    docker compose build"
echo "    docker compose up -d"
echo ""
echo "  Endpoints:"
echo "    Learner   http://localhost:80"
echo "    Admin     http://localhost:3001"
echo "    API       http://localhost:80/api/v1/"
echo ""
