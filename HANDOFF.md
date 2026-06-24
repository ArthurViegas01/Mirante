# Handoff — Captação de propostas freelance (intake)

_Gerado em 2026-06-24 para troca de máquina. Apague quando não precisar mais._

## O que é
Automação para captar projetos do **99Freelas** e triá-los dentro do Mirante, reusando
a máquina de Vagas & CV. Objetivo final: do e-mail do 99Freelas até uma proposta pronta
para enviar, com o mínimo de trabalho manual.

## Estado atual — ingestão + triagem PRONTAS e verificadas ao vivo
Pipeline completo: **poller IMAP → parse → score → fila `/propostas` (descartar)**.

- `apps/api/internal/intake/`:
  - `email.go`/`intake.go` — `ParseEmail` (decodifica o `.eml`, quoted-printable, HTML→texto)
    e `ParseDigest` (campos + links/id). Fixture real em `testdata/digest.eml`.
  - `score.go` — `Score(projeto, skills) → Triage` (aderência + concorrência + frescor).
  - `item.go`/`repo.go`/`sqlite.go` — staging `intake_items` isolado por usuário (migração
    `db/migrations/0018_intake.sql`), dedup por `(user_id, fonte, fonte_id)`.
  - `service.go` — `Ingest([][]byte)`, `List`, `Get`, `Dismiss`. Skills do CV via port
    `SkillsProvider` (ADR-0001).
  - `runner.go` — `Runner` + interface `MessageSource` (testável com fake).
  - `imap.go` — `IMAPSource` (go-imap v1, read-only).
  - `http.go` — `GET /api/intake?estado=&shortlist=`, `POST /api/intake/{id}/dismiss`.
- Front: `apps/web/src/routes/propostas/+page.svelte` + link "Propostas" na Sidebar.
- Wiring: `apps/api/cmd/server/main.go` (gated em `cfg.IntakeEnabled()`).
- Config: `INTAKE_IMAP_*`, `INTAKE_POLL_INTERVAL`, `INTAKE_MIN_SCORE` (ver `.env.example`).
  No `docker-compose.yml` as vars são repassadas à API.
- **Verificado ao vivo:** `intake poll emails=50 new=50`. Testes: `internal/intake` (16),
  `skills`, `config`, `auth`, `db` verdes; `go build ./...` + `npm run build` limpos.

## Como rodar / verificar
1. `.env` na raiz com `INTAKE_IMAP_USERNAME`, `INTAKE_IMAP_PASSWORD` (senha de app do
   Gmail, 16 chars), `INTAKE_IMAP_MAILBOX=INBOX`, `INTAKE_IMAP_FROM=99freelas.com.br`.
   Para testar rápido: `INTAKE_POLL_INTERVAL=1m`.
2. `docker compose up -d --build` (Go **não** roda no host — build via Docker `golang:1.25-alpine`).
3. Log da API: `intake poller started` → `intake poll new=N`. Abra `/propostas`.
- Qualidade: `make api-test`, `make api-lint`, `make web-build`.

## Pendências / próximos passos
1. **E-mails single-project.** No teste real, `emails=50 new=50` ≈ 1 projeto/e-mail → a
   maioria são notificações **"Novo Projeto: X"** (1 projeto), com layout diferente do
   digest **"Novos projetos"** (vários). O parser foi calibrado no digest. Verificar se
   os single vieram completos (título/score/links); se não, adicionar um segundo caminho
   de parse a partir de um `.eml` desse tipo.
2. **Promover** item → vaga (`jobs`) + candidatura (`applications`), via ports (ADR-0001;
   o composition root orquestra). Botão na tela `/propostas`.
3. **Geração da proposta** (LLM, reusa `cv.Adapt`) + **fetch do brief completo** (atrás do
   login do 99Freelas; cookie de sessão cifrado em repouso) + **faixa de preço por porte**.

## Gotchas
- **Atribuição:** o poller usa `GetByEmail(INTAKE_IMAP_USERNAME)` → fallback admin. Em dev
  havia 2 admins (`owner@example.com` semeado pelo compose + a sua conta); os 50 itens do
  primeiro teste foram para o `owner@example.com` (órfãos, inofensivos).
- **Windows + Docker:** o vite não recebe eventos de arquivo do host → após editar
  `.svelte`, `docker compose restart web` (ou ligar `server.watch.usePolling` no vite.config).
- **Prod (Railway):** push na `main` faz deploy. A migração 0018 roda no Turso (aditiva,
  segura). O poller fica **dormente** até setar as `INTAKE_IMAP_*` nas variáveis do Railway.
