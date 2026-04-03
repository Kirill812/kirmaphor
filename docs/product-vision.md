# Kirmaphore — Product Vision & Feature Roadmap

> Ansible automation SaaS для команд, которым не нужен Red Hat Enterprise за $17 500/год.

---

## Позиционирование

| | Semaphore UI | AWX / Tower | Rundeck | **Kirmaphore** |
|---|---|---|---|---|
| Модель | Self-hosted | Self-hosted | Self-hosted | **SaaS** |
| Цена | Free / $15 mo | Free / $13K yr | Free / $51K yr | **Usage-based** |
| Целевая аудитория | Solo DevOps | Enterprise | Ops teams | **Solo + small teams** |
| UI качество | Хорошее | Устаревшее | Устаревшее | **Premium 2026** |
| Passkey auth | ❌ | ❌ | ❌ | **✅** |
| AI помощник | ❌ | ❌ | ❌ | **✅ (roadmap)** |
| Onboarding | Ручной | Ручной | Ручной | **Guided** |

**Главная ставка:** DevOps-инженер (соло или команда до 20 человек) хочет автоматизировать инфраструктуру через Ansible, но не хочет тратить неделю на настройку AWX и $13K/год на Tower. Kirmaphore — это Tower-уровень UX за Railway-уровень цены.

---

## Tier 1 — Core MVP (без этого нельзя начать работать)

### 1.1 Projects

**Суть:** Проект = единица изоляции. Внутри — всё нужное для автоматизации одного окружения или продукта.

**Что нужно:**
- Создание / редактирование / архивирование проекта
- Название, описание, иконка/цвет (visual identity)
- Dashboard проекта: последние запуски, статус окружений, быстрые действия

**Чего нет у конкурентов:** У Semaphore нет dashboard проекта — сразу попадаешь в Task Templates. У Kirmaphore — карточка здоровья проекта с одного взгляда.

---

### 1.2 Repository Integration

**Суть:** Playbook-и живут в Git. Kirmaphore клонирует и держит в актуальном состоянии.

**Что нужно:**
- Подключение репозитория (GitHub, GitLab, Gitea, bare Git URL)
- OAuth-авторизация для GitHub/GitLab (не вводить токены руками)
- Выбор ветки / тега / конкретного коммита для запуска
- Auto-sync: webhook от GitHub → автоматически подтягивает изменения
- Список файлов `.yml` / `.yaml` в репозитории для выбора playbook-а

**Почему важно:** В AWX/Semaphore настройка репозитория — 4 экрана с токенами. Нужно сделать это в 2 клика через OAuth.

---

### 1.3 Inventory Management

**Суть:** Список серверов, на которых запускаются playbook-и.

**Что нужно:**
- Статический inventory (INI или YAML, вставить текстом или загрузить файл)
- Динамический inventory (скрипт / плагин — advanced)
- Группы хостов с переменными
- Проверка доступности хостов (ping check перед запуском)
- Визуальное дерево групп и хостов

**Чего нет у конкурентов:** Ping check до запуска — есть в Tower, нет в Semaphore. В Kirmaphore — стандарт.

---

### 1.4 Credentials & Secrets

**Суть:** SSH ключи, пароли, vault passwords, токены API — хранить безопасно, использовать в запусках.

**Что нужно:**
- Типы credentials: SSH Private Key, Username/Password, Vault Password, API Token, AWS/GCP/Azure
- Хранение зашифрованное (AES-256, ключ из env)
- Привязка credential к проекту или глобально
- Никогда не показывать секрет после сохранения — только «Replace»
- Audit trail: кто и когда использовал credential

**Критично:** Без безопасного хранения ключей — нет доверия продукту. Это блокер для любого реального использования.

---

### 1.5 Playbook Runner (Task Execution)

**Суть:** Запустить `ansible-playbook` с нужными параметрами. Это главная кнопка продукта.

**Что нужно:**
- Выбор: playbook + inventory + credentials + extra vars
- Extra vars — key-value редактор + raw YAML режим
- Теги (`--tags`, `--skip-tags`)
- Verbosity level (-v до -vvvv)
- Dry run (`--check` mode)
- Diff mode (`--diff`)
- Limit (запустить только на части хостов)
- Кнопка «Run» — запуск немедленно
- Отмена запущенного задания

**UX:** Форма запуска должна помещаться в один экран без скролла. Самые частые параметры — наверху, advanced — свернуты.

---

### 1.6 Real-Time Run Logs

**Суть:** Смотреть что происходит прямо сейчас. Диагностировать падения.

**Что нужно:**
- Стриминг ANSI-colored вывода в реальном времени (WebSocket)
- Фильтр по хосту / таску
- Поиск по логу (Ctrl+F внутри)
- Collapse/expand отдельных task-ов
- Статусы задач: ✓ ok / ✗ failed / ⟳ changed / ⚡ unreachable
- Итоговая таблица PLAY RECAP (хосты × статусы)
- Скачать полный лог как .txt
- История запусков с быстрым доступом к логу

**Почему важно:** У Semaphore логи работают хорошо. У AWX — перегружены. Kirmaphore берёт Semaphore-подход + добавляет фильтрацию по хосту.

---

## Tier 2 — Необходимо для регулярного использования

### 2.1 Templates (Saved Run Configurations)

**Суть:** Сохранить набор параметров запуска, чтобы не вводить каждый раз.

**Что нужно:**
- Template = playbook + inventory + credentials + extra vars + опции
- Запуск template в один клик с dashboard
- Переопределение отдельных параметров при запуске (не менять template)
- Клонирование template
- Иконка / цвет для визуального различия

---

### 2.2 Schedules (Cron Automation)

**Суть:** Запускать template по расписанию — ночные бэкапы, daily security patches, etc.

**Что нужно:**
- Cron expression + human-readable preview ("каждый день в 3:00 UTC")
- Timezone support
- Enable/disable без удаления
- Следующий запуск — показывать на карточке
- История scheduled runs отдельно от manual

---

### 2.3 Notifications

**Суть:** Знать когда что-то упало — не сидеть смотреть в экран.

**Что нужно:**
- Telegram (самый важный для целевой аудитории)
- Slack
- Email
- Webhook (универсальный)
- Триггеры: on success / on failure / on change / always
- Кастомное сообщение с переменными (имя проекта, статус, ссылка на лог)

**Почему Telegram первым:** Целевая аудитория — русскоязычные DevOps-инженеры. Telegram стандарт де-факто.

---

### 2.4 Environments (Variable Sets)

**Суть:** staging vs production — разные переменные, один playbook.

**Что нужно:**
- Environment = именованный набор extra vars + credential
- Привязка к template: «запустить в staging» / «запустить в production»
- Визуальная метка на запуске (цветная таблетка: staging=blue, prod=red)
- Запрет случайного запуска в prod (confirmation dialog)

---

### 2.5 Team & Access Control

**Суть:** Дать доступ коллеге, не давая ему всё подряд.

**Что нужно:**
- Роли: Owner / Admin / Operator / Viewer
- Operator: может запускать templates, не может редактировать
- Viewer: только смотреть логи
- Invite по email
- Аудит: кто что запустил

---

## Tier 3 — Competitive Advantage (уникальные фичи)

### 3.1 Forge AI — AI-Assisted Playbook Generation

**Суть:** Описываешь что хочешь сделать — получаешь готовый playbook.

**Что нужно:**
- Chat-интерфейс: "Установи nginx на Ubuntu 22.04 с SSL"
- Генерация playbook + объяснение каждого шага
- Запуск прямо из чата (в dry-run сначала)
- Анализ упавшего лога: "Почему упало?" → AI объясняет и предлагает фикс

**Почему это выигрыш:** Ни один конкурент не имеет встроенного AI-ассистента для Ansible. Это снижает порог входа с «знаю Ansible» до «знаю что хочу».

---

### 3.2 Inventory Discovery

**Суть:** Автоматически найти серверы — не вводить IP вручную.

**Что нужно:**
- Import из облака: AWS EC2, Hetzner, DigitalOcean, VK Cloud
- SSH scan подсети (указываешь 192.168.1.0/24 — находит хосты)
- Группировка по тегам из облака

---

### 3.3 Approval Gates

**Суть:** Перед запуском в production — кто-то второй должен подтвердить.

**Что нужно:**
- Template помечается «requires approval»
- При запуске → статус Pending Approval
- Approver получает уведомление + кнопку Approve/Reject
- Таймаут: если не ответили за N минут → auto-reject или auto-approve

---

### 3.4 Run Graph (Workflow Templates)

**Суть:** Запустить несколько playbook-ов последовательно/параллельно с условиями.

**Что нужно:**
- Visual DAG editor: узлы = templates, стрелки = порядок
- Условия: «следующий шаг только если предыдущий succeeded»
- Параллельные ветки
- Fan-out: один playbook → несколько окружений параллельно

**Аналог:** AWX Workflow Templates. Semaphore не имеет. Это enterprise-killer-feature.

---

### 3.5 Public API + CLI

**Суть:** Интеграция в существующие CI/CD пайплайны.

**Что нужно:**
- REST API: запуск template, статус run, логи
- API Keys (не только сессионные токены)
- CLI: `kirmaphore run --template deploy-prod --env production`
- GitHub Actions integration: официальный action

---

## Антифичи (что намеренно НЕ делаем)

- **Terraform/OpenTofu** — фокус на Ansible, не размываем
- **Встроенный редактор playbook-ов** — Git для этого лучше
- **Собственный secrets backend** — интеграция с Vault, не замена
- **On-premise версия** — SaaS only на старте, усложняет поддержку

---

## Приоритизированный roadmap

```
Now (MVP)          → Projects, Repos, Inventory, Secrets, Runner, Logs
Next (v0.4)        → Templates, Schedules, Notifications (Telegram first)
Soon (v0.5)        → Environments, Teams/RBAC, Audit log
Later (v0.6)       → Forge AI, Inventory Discovery
Ambitious (v1.0)   → Approval Gates, Workflow DAG, API + CLI
```

---

## Метрика успеха

**Активация:** Пользователь запустил первый playbook в течение 10 минут после регистрации.
**Retention:** Пользователь вернулся через 7 дней и запустил ещё один run.
**Expansion:** Команда из 2+ человек использует один проект.
