# Changelog

Todas as mudanças relevantes deste projeto são documentadas aqui.

O formato segue [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/)
e o versionamento segue [SemVer](https://semver.org/lang/pt-BR/).

Este arquivo é a **fonte de verdade do histórico** do Mirante.

## [Não lançado]

### A fazer
- F3 — Vagas, CV e CRM (sobre o kernel `internal/skills`).

## [0.5.0] - 2026-06-08

Projetos: Stacks & Custos — o status ao vivo de cada peça do projeto e o custo das
assinaturas, sem sair do projeto. Extensão de Projetos/Monitor (ver ADR-0005).

### Adicionado
- **Stacks:** serviços do Monitor ganham `provider` (texto livre) e `camada`
  (frontend/backend/database/outro) via migração `0005`. A project view tem uma
  seção **Stacks** que agrupa os serviços do projeto por camada e mostra o status ao
  vivo (SSE), com formulário compacto para adicionar serviço; os forms do Monitor
  passam a aceitar provedor/camada. Status = checagem dos próprios endpoints
  (status-de-provedor fica para depois).
- **Custos:** novo domínio `subscriptions` (migração `0006`) — custo recorrente por
  projeto, opcionalmente ligado a um serviço (`service_id`, soft-link sem FK), em
  **BRL/USD sem conversão**, ciclo mensal/anual. REST em `/api/subscriptions`
  (filtro `?project=`). A project view tem a seção **Custos** (total mensalizado por
  moeda, add/editar/remover) e há a página global **`/custos`** agrupando as
  assinaturas por projeto, com totais gerais.

## [0.4.0] - 2026-06-08

F2 — Tarefas: trabalho em quadro kanban, vinculável a projetos.

### Adicionado
- **Tarefas:** domínio completo (CRUD; status kanban `a_fazer`/`fazendo`/`feito`,
  prioridade, prazo, tags) com REST em `/api/tasks` (filtros `?status=` e
  `?project=`) e migração `0004` (`tasks` + `task_tags`; `project_id` FK→projects
  `ON DELETE SET NULL`; `job_id` reservado, FK só na F3). Anatomia espelha
  `projects`.
- **UI de Tarefas:** quadro kanban (3 colunas, mover entre colunas), formulário com
  prioridade/prazo/projeto/tags, destaque de prazo vencido e filtro por projeto.
- **Project view:** seção "Tarefas abertas" com link para o quadro filtrado; a cópia
  do excluir esclarece que as tarefas são desvinculadas (não apagadas).

### Corrigido
- **Guard de auth no SPA:** deslogado vai para `/login` num shell isolado (sem
  sidebar); logado sai de `/login`; o stream SSE do monitor conecta e desconecta
  reativamente conforme a sessão.
- **`api.js`:** corpo não-JSON (página de erro HTML, resposta de proxy) vira uma
  mensagem limpa em vez de estourar `JSON.parse`.

## [0.3.0] - 2026-06-08

F1 — Projetos + Monitor: o cadastro central e a vitrine real-time.

### Adicionado
- **Projetos:** domínio completo (CRUD, links, tags, archive vs delete) com REST
  em `/api/projects` e UI (data table, criar, project view).
- **Monitor:** migração `0003` (services, check_results, alerts, events outbox);
  `Derive()` puro table-tested (anti-flap N, recovery K, degraded vs timeout,
  wrong-code → down); checkers HTTP/TCP/db_ping via `SafeFetcher` (política-monitor);
  scheduler concorrente (registry de `CancelFunc`, worker pool, single-flight,
  reconcile sob demanda, shutdown sem leak via goleak); alert layer com interface
  `Notifier` (canais externos pluggáveis, nenhum no v1); escrita atômica
  check+alert+event.
- **SSE:** hub com sequência durável (id do outbox), replay por `Last-Event-ID`,
  drop de cliente lento, cap de conexões; `GET /api/stream/monitor`.
- **UI do Monitor:** board com status badges ao vivo, sparkline com ponto Glow,
  uptime (24h/7d/30d), centro de notificações e indicador "ao vivo" ligado ao SSE.
- Refactor de composição: `cmd/server` é o composition root e `platform` não
  importa domínios (ADR-0001); pacote `respond`; `/api/auth/me` devolve o CSRF.

### Notas
- Uptime é computado direto de `check_results` no v1; rollups/pruning ficam para a F5.

## [0.2.0] - 2026-06-07

F0 — scaffolding executável ponta a ponta (web ↔ API ↔ libSQL) com autenticação
de sessão real. Nenhuma feature de domínio ainda.

### Adicionado
- Monorepo Go (`apps/api`, módulo `github.com/lumni/mirante`): `cmd/server` +
  `internal/platform/{config,db,migrate,id,otel,logging,validate,ratelimit,auth,httpserver}`.
- Banco via `database/sql` trocável por scheme (`file:` SQLite pure-Go / `libsql://`
  Turso); migrações goose embarcadas via `go:embed` — `0001` users + sessions.
- Autenticação single-user: Argon2id, sessões server-side (cookie opaco, hash no
  banco), bootstrap do owner por env, rate-limit de login, CSRF + checagem de
  Origin, cookie `HttpOnly`/`Secure`(prod)/`SameSite=Strict`.
- Stack HTTP: request-id, recover, security headers, CORS restrito, rate-limit por IP.
- Testes (verdes): Argon2id, token, rate-limiter, migrações up/down e fluxo de auth
  HTTP (login → me → CSRF → logout → 401); integração libSQL com testcontainers-go
  (build tag `integration`).
- Frontend SvelteKit + Svelte 5: shell com sidebar (tag `MIRANTE`, ponto Glow ao
  vivo), alternância de tema, `/login`, `/styleguide` e primitivos (Button, Input,
  StatusBadge) consumindo só role tokens.
- Infra: Dockerfile da API (distroless non-root, sem CGO), `docker-compose`
  (sqld + api + web) validado ponta a ponta, CI (build + vet + test +
  golangci-lint + build do web).
- ADR-0001..0004, `.env.example`, `Makefile` (targets Go via Docker).

### Notas
- `db/migrations` mora em `apps/api/db/migrations` (exigência do `go:embed`).
- `internal/skills` (shared kernel) será criado na F3.

## [0.1.0] - 2026-06-07

Marco de **planejamento**. Nenhuma feature implementada — apenas plano,
estrutura e artefatos de fundação.

### Adicionado
- `docs/design/lumni-design-system.md` reconhecido como fonte de verdade do design.
- Plano de entrega em fases (F0–F5) com objetivo e Definition of Done por fase.
- Árvore de diretórios do monorepo (`apps/api`, `apps/web`, `db`, `docs`) com a
  estrutura interna idiomática do Go (`cmd/server` + `internal/*`).
- Modelagem de dados de todos os módulos (Projetos, Monitor, Tarefas, Vagas & CV/CRM).
- Decisão arquitetural registrada: **SSE** para o push do monitor (ver README).
- `apps/web/src/app.css` — bloco de tokens do Lumni Design System (role tokens com
  light em `:root` e dark em `[data-theme="dark"]`, focus ring de 3px, clamp de
  `prefers-reduced-motion`, fontes Figtree/Instrument Serif/Geist Mono) e a
  assinatura **Glow** (`#5EEAD4`) dos elementos "ao vivo" (`--color-live*`).
- `README.md` (esqueleto) e este `CHANGELOG.md`.

[Não lançado]: https://example.com/mirante/compare/v0.5.0...HEAD
[0.5.0]: https://example.com/mirante/compare/v0.4.0...v0.5.0
[0.4.0]: https://example.com/mirante/compare/v0.3.0...v0.4.0
[0.3.0]: https://example.com/mirante/compare/v0.2.0...v0.3.0
[0.2.0]: https://example.com/mirante/compare/v0.1.0...v0.2.0
[0.1.0]: https://example.com/mirante/releases/tag/v0.1.0
