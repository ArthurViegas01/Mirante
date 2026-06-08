# Changelog

Todas as mudanças relevantes deste projeto são documentadas aqui.

O formato segue [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/)
e o versionamento segue [SemVer](https://semver.org/lang/pt-BR/).

Este arquivo é a **fonte de verdade do histórico** do Mirante.

## [Não lançado]

### A fazer
- F1 — Projetos + Monitor (a vitrine real-time): scheduler concorrente, estado
  derivado com anti-flap, histórico/uptime, alertas in-app e push SSE.

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

[Não lançado]: https://example.com/mirante/compare/v0.2.0...HEAD
[0.2.0]: https://example.com/mirante/compare/v0.1.0...v0.2.0
[0.1.0]: https://example.com/mirante/releases/tag/v0.1.0
