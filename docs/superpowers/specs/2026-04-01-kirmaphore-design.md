# Kirmaphore — Design Spec
**Date:** 2026-04-01  
**Status:** Approved  
**Author:** kgory + Claude

---

## 1. Overview

**Kirmaphore** is an AI-native, open-core Ansible automation platform for DevOps and Platform Engineering teams in mid-to-large enterprises.

**Positioning:** Not "yet another AWX clone" — an intelligent infrastructure automation hub where AI (Forge) is central, not an add-on.

**Tagline:** *"Your infrastructure understands you."*

### Key Differentiators vs AWX / Semaphore

| Area | AWX | Semaphore | Kirmaphore |
|---|---|---|---|
| AI Agent | Lightspeed (paid add-on) | None | Forge — built-in, agentic |
| Workflow Editor | Dated UI | None | Visual DAG, drag-and-drop |
| Developer Experience | Complex | Basic | Web IDE + Git UI (Monaco) |
| Observability | Basic logs | Basic logs | Config drift, live graph, cloud topology |
| Auth | LDAP/SAML | Local/OIDC | Passkeys + OIDC + SAML + LDAP |
| Cloud Inventory | Limited | None | Native AWS / Azure / GCP |
| Open-core | AWX free / Tower paid | Fully free | Engine free / AI+Enterprise paid |

---

## 2. Target Audience

**Primary:** DevOps and Platform Engineering teams in medium and large companies.

**Needs:**
- Enterprise-grade RBAC, SSO, audit logs, compliance
- Ansible-first with deep integration (inventories, roles, collections, execution environments)
- Secure secrets management
- AI that saves real engineering time

---

## 3. Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      KIRMAPHORE                         │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │  Web IDE  │  │ Workflow │  │ Observ.  │  │ RBAC & │ │
│  │ + Git UI  │  │ Designer │  │Dashboard │  │  Auth  │ │
│  └──────────┘  └──────────┘  └──────────┘  └────────┘ │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │                  FORGE AI AGENT                  │   │
│  │  error analysis · code gen · natural language   │   │
│  │  drift explanation · run suggestions            │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │              EXECUTION ENGINE                    │   │
│  │   Job Queue → Runner Pods → Ansible Process     │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐             │
│  │Inventory │  │ Secrets  │  │   Git    │             │
│  │ Manager  │  │ Manager  │  │ Gateway  │             │
│  └──────────┘  └──────────┘  └──────────┘             │
└─────────────────────────────────────────────────────────┘
```

### Tech Stack

| Layer | Technology | Rationale |
|---|---|---|
| Backend | Go | Performance, single binary deployment |
| Frontend | Next.js 15 + Tailwind + shadcn/ui | Modern enterprise DX |
| Database | PostgreSQL | Reliability, JSONB for dynamic data |
| Queue | NATS JetStream | Lightweight, embeddable, Go-native |
| Execution | Docker / Kubernetes pods | Isolated runner environments |
| AI (Forge) | LLM-agnostic (OpenAI / Anthropic / Ollama) | Self-hosted option for enterprise |
| Web IDE | Monaco Editor + Ansible LSP | Autocomplete, validation |
| Auth | WebAuthn + OIDC + LDAP + local | Enterprise SSO from day one |
| i18n | next-intl (EN, RU, DE, ZH initial) | Localization from day one |

---

## 4. Security, Auth & Profile

### User Profile Model
*(patterns from KirFlow + PostCash)*

**profiles table:**
```
id, email, display_name, avatar_url
blocked_at, created_at, onboarded
session_timeout_minutes
settings (JSONB)
```

**passkey_credentials table:**
```
credential_id, public_key, counter
transports, device_name, last_used_at, user_id
```

**user_sessions table:**
```
session_token_hash, device_fingerprint
geo_city, geo_country, ip_address
user_agent, is_current
```

**user_known_devices table:**
```
device_fingerprint, device_label
first_seen_at, last_seen_at
```

### RBAC Roles

| Role | Capabilities |
|---|---|
| `owner` | Everything + billing + org deletion |
| `admin` | Manage projects, users, secrets |
| `engineer` | Run jobs, edit playbooks |
| `viewer` | Read-only: logs and dashboard |
| Custom roles (Enterprise) | Granular permission sets via Role Constructor |

**Role Constructor (Enterprise):** Visual editor for creating custom roles with granular permission toggles per resource type (projects, inventories, secrets, playbooks, runners, cloud credentials). Supports role inheritance and per-project scope overrides.

### Authentication Methods

| Method | Tier |
|---|---|
| Passkeys (WebAuthn) — primary | Free |
| Local email/password + TOTP | Free |
| OIDC / OAuth2 | Enterprise |
| LDAP / Active Directory | Enterprise |
| SAML 2.0 | Enterprise |
| Admin-enforced MFA policy | Enterprise |

**Passkeys implementation:** Full WebAuthn stack from PostCash — CBOR decoder, AuthData parser, COSE P-256 key import, origin validation, RP ID derivation.

### Secure Session (from PostCash)

Sensitive/destructive operations require re-authentication within 5 minutes:
- Project deletion
- Secret rotation
- Role changes
- Cloud credential update

Returns `403 secure_session_required` if session is stale.

### Secrets Encryption (two-layer)

- **Layer 1:** Client-side AES-256-GCM encryption before transmission
- **Layer 2:** Server-side master key encryption at rest
- Applies to: SSH keys, Ansible Vault passwords, API tokens, cloud credentials

### Device Fingerprinting

- FingerprintJS integration for anomaly detection
- Login anomalies trigger admin alert
- Users can view and revoke active sessions and known devices

---

## 5. Native Cloud Support

### Three Integration Levels

1. **Dynamic Inventory** — automatic host discovery
   - AWS: EC2 Auto Discovery by tags/regions/VPC
   - Azure: VMSS, Resource Groups
   - GCP: GCE Instance Groups, labels

2. **Cloud Credentials** — stored via two-layer encryption
   - AWS: IAM Role / Access Key + Secret
   - Azure: Service Principal / Managed Identity
   - GCP: Service Account JSON key

3. **Cloud Execution** — runner pods deployable in EKS / AKS / GKE for reduced latency

### Forge AI + Cloud

Natural language queries across cloud inventory:
> "Show all hosts in eu-west-1 where no playbook ran in the last 7 days"

### Community vs Enterprise Cloud Features

| Feature | Community | Enterprise |
|---|---|---|
| Dynamic inventory | Read-only | Full (auto-sync) |
| Cloud credentials | Manual | Managed Identity / AssumeRole |
| Runner in cloud | ❌ | ✅ |

---

## 6. Core Modules

### 6.1 Workflow Designer (Visual DAG)

Drag-and-drop job chain editor — primary differentiator from AWX and Semaphore.

**Node types:**
- Job Template
- Approval Gate (manual confirmation before prod)
- Condition (based on previous job output)
- Delay
- Webhook

**Branch types:** `on_success` / `on_failure` / `always`

Workflows stored as YAML → committed to Git automatically.

### 6.2 Web IDE + Git UI

Monaco Editor (VS Code engine) with:
- Ansible Language Server: module autocomplete, variable hints, role resolution
- Built-in Git: commit, push, diff, PR review — GitHub-like UX
- Lint on save (ansible-lint)
- Dry-run preview via `--check` directly from editor

### 6.3 Observability Dashboard

| Feature | Description |
|---|---|
| Live log stream | WebSocket real-time playbook output |
| Config Drift | Periodic `--check` runs, per-host drift visualization |
| Host Graph | Dependency graph of hosts affected by last run |
| Run History | Timeline with filters: status, project, cloud tags |
| Cloud Topology | VM status from AWS/Azure/GCP on dashboard |

### 6.4 Forge AI Agent

Autonomous agent with full platform context access — not a chatbot.

| Capability | Example |
|---|---|
| Error analysis | "Playbook failed on task 7. Forge: variable `db_host` undefined in group_vars/prod" |
| Fix suggestion | Proposes patch + "Apply & Retry" button |
| Code generation | "Write a playbook to install nginx with SSL on Ubuntu 22.04" |
| Natural language query | "Show hosts with no successful deploy in 7 days" |
| Drift explanation | "3 hosts drifting: openssh updated to X outside our playbook" |

**LLM backends:** OpenAI, Anthropic, Ollama (self-hosted for enterprise, zero data egress).

---

## 7. Open-Core Tiers

| Feature | Community | Enterprise |
|---|---|---|
| Execution engine | ✅ | ✅ |
| Workflow Designer | ✅ | ✅ |
| Web IDE + Git UI | ✅ | ✅ |
| Passkeys + local auth | ✅ | ✅ |
| Basic RBAC (4 roles) | ✅ | ✅ |
| Localization (EN/RU/DE/ZH) | ✅ | ✅ |
| Forge AI | 100 req/day | Unlimited + self-hosted LLM |
| Cloud integrations | Read-only inventory | Full dynamic + cloud execution |
| SSO / SAML / LDAP | ❌ | ✅ |
| Audit logs | ❌ | ✅ |
| External Secrets (Vault/KMS) | ❌ | ✅ |
| MFA enforce policy | ❌ | ✅ |
| Secure session | ❌ | ✅ |
| Role Constructor (custom roles) | ❌ | ✅ |
| SLA + support | ❌ | ✅ |

---

## 8. MVP Scope (Phase 1)

Minimum viable product to validate core value proposition:

1. **Execution Engine** — run Ansible playbooks, job queue, Docker runners
2. **Basic RBAC + Auth** — Passkeys + local login, 4 roles
3. **Inventory Manager** — static + AWS/Azure/GCP dynamic (read-only)
4. **Web IDE** — Monaco + Ansible LSP + Git integration
5. **Workflow Designer** — basic DAG (Job Template nodes, success/failure branches)
6. **Observability** — live log stream + run history
7. **Forge AI (limited)** — error analysis + fix suggestion on failed runs
8. **Secrets Manager** — two-layer encrypted storage

**Out of scope for MVP:** SAML, audit logs, External Vault, cloud execution runners, Approval Gate node, full NL query.

---

## 9. Semaphore Implementation References

**Principle:** Where Semaphore has battle-tested implementation, adopt patterns directly. Build on proven foundations, not from scratch.

### Execution Engine → adopt from `services/tasks/`

- **TaskPool** (`TaskPool.go`): channel-based queue (`register`, `logger` 10K buffer, `queueEvents` pub/sub). Use this model directly — not goroutine-per-task.
- **TaskRunner** (`TaskRunner.go`): wraps Task + Template + Inventory + Repository + Environment together. Job interface: `Run(username, version, alias) error`, `Kill()`, `IsKilled()`.
- **Pluggable state stores**: `TaskStateStore` interface — memory for single-node, Redis for HA.
- **Batch DB writes**: 500 log records per 500ms write cycle — adopt for Kirmaphore log persistence.

### Database Models → adopt from `db/`

Reuse schema patterns from Semaphore directly (adapted to PostgreSQL):

```
Task       — ID, TemplateID, ProjectID, Status, RunnerID, Playbook,
             GitBranch, CommitHash, InventoryID, Params (JSONB)
Schedule   — CronFormat, RunAt, Active, DeleteAfterRun, TaskParamsID
Inventory  — Type (static|static-yaml|file|cloud-source), SSHKeyID
Runner     — Token, Tag, MaxParallelTasks, Touched (heartbeat)
AccessKey  — Type (ssh|login_password|string), StorageID, SourceStorageType
Repository — GitURL, GitBranch, SSHKeyID, Type (git|ssh|https|file|local)
```

### Secrets Architecture → adopt from `services/server/access_key_encryption_svc.go`

Strategy pattern for storage backends — adopt directly:
- `LocalAccessKeyDeserializer` — encrypted at rest in DB
- `VaultAccessKeyDeserializer` — HashiCorp Vault
- Extend with: AWS KMS, Azure Key Vault, GCP Secret Manager

### Git Integration → adopt from `db/Repository.go`

- Per-repo SSH key (AccessKey) reference
- Per-task GitBranch override
- CommitHash stored in Task after checkout
- Support: SSH, HTTPS, local paths

### Scheduling → adopt from `services/schedules/`

- Cron + run_at types
- `ValidateCronFormat()` utility
- SchedulePool with refresh on update
- `DeleteAfterRun` flag

### API Patterns → adopt from `api/`

```go
// Middleware chain — adopt directly
ProjectMiddleware → load project → resolve role → check permissions → context

// Helpers — adopt directly
helpers.GetFromContext(r, key)
helpers.GetIntParam("id", w, r)
helpers.WriteJSON(w, status, obj)
helpers.Bind(w, r, &obj)
```

### Notifications → adopt from `services/tasks/alert.go`

- `embed.FS` template-driven email alerts
- Chat hooks pattern (Slack/Telegram/Teams) — extend with Forge AI summaries
- Per-user alert preferences on User model

### What Kirmaphore adds ON TOP of these patterns

| Semaphore foundation | Kirmaphore addition |
|---|---|
| TaskPool + TaskRunner | Forge AI hooks into task lifecycle (on_failure analysis) |
| AccessKey storage | + Two-layer client+server encryption (from PostCash) |
| Static/file inventories | + Native AWS/Azure/GCP dynamic sources |
| Basic permissions | + Role Constructor, org/team scoping |
| Session/local auth | + Passkeys (WebAuthn), OIDC, SAML |
| Template-driven notifications | + Workflow Webhook nodes, per-task NL summaries from Forge |
| Cron scheduling | + Visual Workflow DAG with branching |

---

## 10. Research: Semaphore Gaps (input data)

Top user requests from semaphoreui/semaphore that Kirmaphore addresses:

| Semaphore Issue | Kirmaphore Solution |
|---|---|
| #2334 Visual workflow templates | Workflow Designer (6.1) |
| #891 Granular RBAC | 4-role RBAC + org/team (Sec. 4) |
| #2248 External secrets vault | Enterprise tier secrets (Sec. 7) |
| #1825 Outbound event webhooks | Webhook node in Workflow Designer |
| #973 Full OIDC/SSO | Enterprise auth stack (Sec. 4) |
| #2594 Per-task notifications | Notification node in workflows |
| Dynamic cloud inventories | Native AWS/Azure/GCP (Sec. 5) |
