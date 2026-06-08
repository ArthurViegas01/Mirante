# Mirante

> Central pessoal de comando — monitoro meus projetos, organizo tarefas e
> conduzo minha busca de carreira num só lugar.
>
> _by Lumni_ — o irmão "solo" do Lumni Console.

**Status:** `v0.1.0` — planejamento. Nenhuma feature implementada ainda
(ver [CHANGELOG.md](CHANGELOG.md)).

---

## O que é

Mirante é um app **single-user** organizado em quatro áreas, tendo **Projeto**
como espinha de tudo:

| Área | O que faz |
|---|---|
| **Projetos** | Cadastro central. Tudo se vincula a um projeto. |
| **Monitor** | Saúde ao vivo dos serviços de cada projeto, com alertas in-app. |
| **Tarefas** | Atividades vinculáveis a projetos e a vagas. |
| **Vagas & CV** | Aderência (determinística + LLM), CV mestre adaptável, e-mail/carta e CRM de candidaturas. |

## Stack

Tecnologias deliberadamente novas (regra de aprendizado do projeto).

- **Backend:** Go — monólito modular, domain-driven.
- **Frontend:** SvelteKit + Svelte 5 (runes).
- **Estilo:** CSS moderno nativo (View Transitions, container queries,
  scroll-driven animations, `color-mix`, anchor positioning) + GSAP. Sem Tailwind.
- **Banco:** SQLite via **libSQL/Turso** (single-writer, single-instance).
- **Real-time:** **SSE** (Server-Sent Events) em Go — ver ADR-0002.
- **LLM:** HTTP direto a Anthropic/Groq/OpenAI, multi-provider trocável por env,
  saída estruturada via JSON Schema.
- **Migrações:** goose (embarcadas no binário via `go:embed`).
- **Validação:** go-playground/validator. **Testes:** Go testing + testify +
  testcontainers-go. **Lint:** golangci-lint.
- **Observabilidade:** OpenTelemetry. **Deploy:** Fly.io (single machine).
  **Dev:** Docker Compose.

## Design

Fonte de verdade: [docs/design/lumni-design-system.md](docs/design/lumni-design-system.md).
Tokens em [apps/web/src/app.css](apps/web/src/app.css) — role tokens (light em
`:root`, dark em `[data-theme="dark"]`), focus ring de 3px, clamp de
`prefers-reduced-motion`, fontes Figtree / Instrument Serif / Geist Mono.
Mirante não diverge a paleta da Lumni; a única assinatura é **Glow** (`#5EEAD4`)
nos elementos "ao vivo" (`--color-live*`), e a tag mono da sidebar vira `MIRANTE`.

## Estrutura do monorepo

```
mirante/
├─ apps/
│  ├─ api/                 # Go — módulo github.com/lumni/mirante
│  │  ├─ cmd/server/       # entrypoint (composition root)
│  │  ├─ internal/platform # config, db, migrate, auth, httpserver, otel, … (+ domínios a partir da F1)
│  │  └─ db/migrations/    # SQL goose, embarcadas no binário via go:embed
│  └─ web/                 # SvelteKit + Svelte 5
├─ docs/
│  ├─ design/              # Lumni Design System (fonte de verdade)
│  └─ adr/                 # ADR-0001..0004
├─ docker-compose.yml · CHANGELOG.md · README.md
```

> Os pacotes de domínio (`projects`, `monitor`, `tasks`, `jobs`, `cv`) e o shared
> kernel `skills` entram em `apps/api/internal/` a partir da F1. As migrações ficam
> sob `apps/api/db/` (não na raiz) por causa do `go:embed`, que não alcança fora do
> módulo.

## Decisões de arquitetura (ADRs)

> A serem materializadas em `docs/adr/` na F0.

- **ADR-0001 — Fronteiras do monólito modular.** Pacotes de domínio não importam
  uns aos outros; cross-domain via interfaces (ports) + IDs tipados. Shared
  kernel (`skills`, `tags`, IDs, canonical keywords) explícito.
- **ADR-0002 — SSE + instância única.** Push do monitor por SSE; `min=max=1` no
  Fly (hub e scheduler in-memory não são seguros sob escala horizontal).
- **ADR-0003 — Dois domínios de confiança para fetch.** Alvos do monitor (infra
  própria) permitem IP privado; links de vaga (URL externa) **bloqueiam**
  IP privado/link-local/metadata (anti-SSRF).
- **ADR-0004 — Abstração de LLM.** Provider trocável por env (não failover
  automático entre providers); saída estruturada via JSON Schema; rate limit por
  rota; ledger de uso.

## Desenvolvimento

Pré-requisitos: **Docker** (obrigatório) e, opcionalmente, **Node 20+** para rodar
o frontend fora de container. Go **não** é necessário no host — os alvos Go do
`Makefile` rodam na imagem oficial `golang` via Docker.

```bash
docker compose up --build       # sobe libSQL (sqld) + API + web (dev)
# web:  http://localhost:5173
# api:  http://localhost:8080  (healthz, /api/auth/*)
# login de dev: owner@example.com / change-me-dev  (definido no compose)
```

Sem Docker para o frontend:

```bash
npm --prefix apps/web install
npm --prefix apps/web run dev   # usa API_URL para o proxy /api → :8080
```

Qualidade (via Docker, sem Go no host):

```bash
make api-test     # go test ./...
make api-lint     # golangci-lint
make web-build    # build de produção do SvelteKit
```

> Copie `.env.example` para `.env` para rodar a API fora do compose. Nenhum segredo
> é versionado.

## Segurança

Segredos só via env; CORS restrito; headers de segurança; validação/sanitização
de toda entrada (inclusive links de vaga); credenciais nunca em texto puro
(cifradas em repouso); contêineres non-root. Ver a seção de riscos no plano.

## Licença

Privado — Lumni.
