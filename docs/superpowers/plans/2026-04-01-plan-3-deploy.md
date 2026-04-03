# Deploy Kirmaphore to Production Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Deploy Kirmaphore (Go API + Next.js frontend + PostgreSQL) to kirmaphore.kirvps.work on peppermint (51.255.193.130) behind the shared Caddy proxy.

**Architecture:** Multi-stage Docker builds for both services; docker-compose.yml in the repo manages the stack. Services join the existing `agent-sandbox-api_sandbox-net` Docker network so Caddy can reach them by container name. Caddy handles TLS (Let's Encrypt) and routes `/api/*` to the Go API and everything else to Next.js.

**Tech Stack:** Docker / Docker Compose, Go 1.25, Node 22, PostgreSQL 16, Caddy 2, Cloudflare DNS API.

---

## File Map

| File | Action | Purpose |
|---|---|---|
| `Dockerfile.api` | Create | Multi-stage Go binary build |
| `Dockerfile.web` | Create | Multi-stage Next.js build |
| `docker-compose.yml` | Create | Stack definition: postgres + api + web |
| `.env.example` | Create | Template for production secrets |
| `.dockerignore` | Create | Exclude node_modules / build artifacts |
| `web/next.config.ts` | Modify | Accept `NEXT_PUBLIC_API_URL` build arg |

---

## Task 1: Dockerfile for Go API

**Files:**
- Create: `Dockerfile.api`

- [ ] **Step 1: Create multi-stage Dockerfile.api**

```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /kirmaphore ./cmd/kirmaphore

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=builder /kirmaphore /app/kirmaphore
COPY migrations/ /app/migrations/

EXPOSE 8080
ENTRYPOINT ["/app/kirmaphore"]
```

- [ ] **Step 2: Verify it builds locally**

```bash
docker build -f Dockerfile.api -t kirmaphore-api:local .
```

Expected: image built, `kirmaphore-api:local` appears in `docker images`.

- [ ] **Step 3: Commit**

```bash
git add Dockerfile.api
git commit -m "build: add multi-stage Dockerfile for Go API"
```

---

## Task 2: Dockerfile for Next.js Frontend

**Files:**
- Create: `Dockerfile.web`
- Modify: `web/next.config.ts`

- [ ] **Step 1: Add build arg support to web/next.config.ts**

Replace the content of `web/next.config.ts` with:

```typescript
import createNextIntlPlugin from 'next-intl/plugin'

const withNextIntl = createNextIntlPlugin('./i18n/request.ts')

const nextConfig = {
  output: 'standalone',
}

export default withNextIntl(nextConfig)
```

Note: `NEXT_PUBLIC_API_URL` is already read from env at build time by Next.js automatically (any `NEXT_PUBLIC_*` env var present during `next build` is baked in). Setting `output: 'standalone'` produces a self-contained server bundle suitable for Docker.

- [ ] **Step 2: Create Dockerfile.web**

```dockerfile
# syntax=docker/dockerfile:1
FROM node:22-alpine AS deps
WORKDIR /app
COPY web/package.json web/package-lock.json* ./
RUN npm ci

FROM node:22-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY web/ .

ARG NEXT_PUBLIC_API_URL=https://kirmaphore.kirvps.work
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL

RUN npm run build

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public

EXPOSE 3000
CMD ["node", "server.js"]
```

- [ ] **Step 3: Create .dockerignore**

```
# .dockerignore
web/node_modules
web/.next
.git
*.md
docs/
```

- [ ] **Step 4: Verify Next.js build works**

```bash
docker build -f Dockerfile.web \
  --build-arg NEXT_PUBLIC_API_URL=https://kirmaphore.kirvps.work \
  -t kirmaphore-web:local .
```

Expected: image built successfully. Next.js standalone build completes.

- [ ] **Step 5: Commit**

```bash
git add Dockerfile.web .dockerignore web/next.config.ts
git commit -m "build: add multi-stage Dockerfile for Next.js frontend (standalone)"
```

---

## Task 3: docker-compose.yml and .env.example

**Files:**
- Create: `docker-compose.yml`
- Create: `.env.example`

- [ ] **Step 1: Create docker-compose.yml**

```yaml
services:
  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: kirmaphore
      POSTGRES_USER: kirmaphore
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - internal

  api:
    build:
      context: .
      dockerfile: Dockerfile.api
    container_name: kirmaphore-api
    restart: unless-stopped
    environment:
      DATABASE_URL: postgres://kirmaphore:${POSTGRES_PASSWORD}@postgres:5432/kirmaphore?sslmode=disable
      MASTER_KEY: ${MASTER_KEY}
      RP_ORIGIN: https://kirmaphore.kirvps.work
      RP_ID: kirmaphore.kirvps.work
      RP_NAME: Kirmaphore
      ALLOWED_ORIGIN: https://kirmaphore.kirvps.work
    depends_on:
      - postgres
    networks:
      - internal
      - sandbox-net

  web:
    build:
      context: .
      dockerfile: Dockerfile.web
      args:
        NEXT_PUBLIC_API_URL: https://kirmaphore.kirvps.work
    container_name: kirmaphore-web
    restart: unless-stopped
    networks:
      - sandbox-net

volumes:
  postgres_data:

networks:
  internal:
  sandbox-net:
    external: true
    name: agent-sandbox-api_sandbox-net
```

- [ ] **Step 2: Create .env.example**

```bash
# Copy to .env and fill in values before deploying
# Generate POSTGRES_PASSWORD: openssl rand -hex 16
POSTGRES_PASSWORD=

# Generate MASTER_KEY: openssl rand -hex 32  (must be 64 hex chars = 32 bytes)
MASTER_KEY=
```

- [ ] **Step 3: Add .env to .gitignore**

```bash
echo '.env' >> .gitignore
```

Verify `.gitignore` now contains `.env`. If `.gitignore` does not exist yet, this creates it.

- [ ] **Step 4: Commit**

```bash
git add docker-compose.yml .env.example .gitignore
git commit -m "deploy: add docker-compose.yml and env template"
```

---

## Task 4: Push to GitHub and Add DNS Record

**Files:** None (git and Cloudflare API operations)

- [ ] **Step 1: Ensure GitHub remote exists**

```bash
git remote -v
```

If no remote named `origin` exists, add it:

```bash
git remote add origin https://github.com/kgory/kirmaphor.git
```

- [ ] **Step 2: Push to GitHub**

```bash
git push -u origin main
```

Expected: all commits pushed successfully.

- [ ] **Step 3: Add DNS A record for kirmaphore.kirvps.work**

```bash
curl -s -X POST \
  "https://api.cloudflare.com/client/v4/zones/03b42c9f0493f2f5ff3d94cab57699a4/dns_records" \
  -H "Authorization: Bearer nHY8aEi7wgiDkbk3bHiKcCkWGKIh_hZKrPpHdxwo" \
  -H "Content-Type: application/json" \
  --data '{
    "type": "A",
    "name": "kirmaphore",
    "content": "51.255.193.130",
    "ttl": 1,
    "proxied": false
  }'
```

Expected: JSON response with `"success": true`.

- [ ] **Step 4: Verify DNS propagation**

```bash
dig kirmaphore.kirvps.work @1.1.1.1 +short
```

Expected: `51.255.193.130`

---

## Task 5: Deploy to Peppermint and Configure Caddy

**Files:** Edited on server via SSH

- [ ] **Step 1: SSH to peppermint and clone/update repo**

```bash
ssh peppermint "
  mkdir -p /opt/kirmaphore &&
  cd /opt/kirmaphore &&
  if [ -d .git ]; then
    git pull
  else
    git clone https://github.com/kgory/kirmaphor.git .
  fi
"
```

Expected: repo cloned or updated to latest main.

- [ ] **Step 2: Create .env on server**

```bash
ssh peppermint "
  POSTGRES_PASSWORD=\$(openssl rand -hex 16) &&
  MASTER_KEY=\$(openssl rand -hex 32) &&
  cat > /opt/kirmaphore/.env <<EOF
POSTGRES_PASSWORD=\$POSTGRES_PASSWORD
MASTER_KEY=\$MASTER_KEY
EOF
  echo 'Generated .env:'
  cat /opt/kirmaphore/.env
"
```

Expected: `.env` file created with random secrets. Copy and store these values.

- [ ] **Step 3: Build and start the stack**

```bash
ssh peppermint "
  cd /opt/kirmaphore &&
  docker compose build --no-cache &&
  docker compose up -d
"
```

Expected: three containers start: `kirmaphore-postgres` (internal), `kirmaphore-api`, `kirmaphore-web`.

- [ ] **Step 4: Verify containers are running**

```bash
ssh peppermint "docker compose -f /opt/kirmaphore/docker-compose.yml ps"
```

Expected: all three services show `running`.

- [ ] **Step 5: Verify API migrations ran**

```bash
ssh peppermint "docker logs kirmaphore-api 2>&1 | head -20"
```

Expected: log lines showing migration applied and `starting on :8080`.

- [ ] **Step 6: Add kirmaphore.kirvps.work to Caddy**

```bash
ssh peppermint "
cat >> /opt/agent-sandbox-api/Caddyfile << 'EOF'

kirmaphore.kirvps.work {
    handle /api/* {
        reverse_proxy kirmaphore-api:8080
    }
    handle {
        reverse_proxy kirmaphore-web:3000
    }
}
EOF
"
```

- [ ] **Step 7: Reload Caddy**

```bash
ssh peppermint "docker exec \$(docker ps -qf name=caddy) caddy reload --config /etc/caddy/Caddyfile"
```

Expected: Caddy reloads config, no errors output.

- [ ] **Step 8: Smoke test**

```bash
curl -s https://kirmaphore.kirvps.work/api/health
```

Expected:
```json
{"status":"ok"}
```

Then open https://kirmaphore.kirvps.work in a browser — the login page should load.

---

## Self-Review

**Spec coverage:**
- ✅ Multi-stage Dockerfile for Go API (Task 1)
- ✅ Multi-stage Dockerfile for Next.js with standalone output (Task 2)
- ✅ docker-compose.yml with postgres, api, web (Task 3)
- ✅ .env.example and .gitignore (Task 3)
- ✅ DNS A record via Cloudflare API (Task 4)
- ✅ Deploy to /opt/kirmaphore/ on peppermint (Task 5)
- ✅ Join agent-sandbox-api_sandbox-net (Task 3 docker-compose)
- ✅ Caddy route added (Task 5)
- ✅ RP_ORIGIN set to https://kirmaphore.kirvps.work for WebAuthn (Task 3 docker-compose)
- ✅ Migrations run on startup (built into Go binary)

**No placeholders found.**

**Type consistency:** No shared types across tasks.
