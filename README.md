# Mirante

> Central pessoal de comando — monitoro meus projetos, organizo tarefas e
> conduzo minha busca de carreira num só lugar.
>
> _by Lumni_ — o irmão "solo" do Lumni Console.

**Status:** `v0.7.0` — F0–F5 entregues: Projetos, Monitor (no projeto, com rollups/
pruning), Tarefas, Stacks & Custos, Vagas/CV/CRM com IA (Groq), webhooks de alerta e
observabilidade OTLP — além de prontidão de produção (banco Turso, deploy Fly).
**Não lançado:** **multiusuário** (isolamento por usuário + cadastro com aprovação
do admin), recuperação de senha por e-mail e **importação de projeto via link do
GitHub**. Ver [CHANGELOG.md](CHANGELOG.md).

---

## O que é

Mirante é um app **multiusuário** — cada pessoa tem o seu Mirante privado, no
mesmo deploy de instância única — organizado em quatro áreas, tendo **Projeto**
como espinha de tudo:

| Área | O que faz |
|---|---|
| **Projetos** | Cadastro central (importa de um link de repositório do GitHub). Tudo se vincula a um projeto. |
| **Monitor** | Saúde ao vivo dos serviços de cada projeto, com alertas in-app. |
| **Tarefas** | Atividades vinculáveis a projetos e a vagas. |
| **Vagas & CV** | Aderência (determinística + LLM), CV mestre adaptável, e-mail/carta e CRM de candidaturas. |

O **Início** (`/`) é o painel que reúne tudo: KPIs das quatro áreas, status ao
vivo do monitor, foco do dia (tarefas a vencer) e o snapshot de carreira.

### Contas & acesso

Mirante é **multiusuário por isolamento**: cada conta tem o seu próprio Mirante —
ninguém vê os dados de ninguém. O **cadastro é aberto**, mas toda conta nova nasce
**pendente** e só consegue entrar depois que o **admin** a ativa. A primeira conta
da instância (ou a semeada por `OWNER_*`) é o admin. A gestão de contas (ativar,
desativar, criar, excluir) fica em **Usuários** (`/admin/usuarios`), visível apenas
para o admin. Ver [ADR-0008](docs/adr/0008-multiusuario.md).

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
Tokens em [apps/web/src/app.css](apps/web/src/app.css): role tokens (light em
`:root`, dark em `[data-theme="dark"]`), focus ring de 3px e clamp de
`prefers-reduced-motion`. As fontes (Figtree, Instrument Serif, Geist Mono) são
self-hosted em `apps/web/static/fonts/` (woff2, `font-display: swap`).
Mirante não diverge a paleta da Lumni; a única assinatura é **Glow** (`#5EEAD4`)
nos elementos "ao vivo" (`--color-live*`), e a tag mono da sidebar vira `MIRANTE`.

A camada de UX é construída sobre primitivos reutilizáveis em
[apps/web/src/lib/components/](apps/web/src/lib/components/) (Button, Input,
Textarea, Select, Modal, ConfirmHost, Toaster, EmptyState, Skeleton, StatCard,
StatusBadge, BrandMark, UserMenu) e stores em runes (`toast`, `confirm`,
`session`, `monitor`). O app abre num **dashboard** (`/`) que reúne as quatro
áreas, com identidade do usuário e logout no menu da sidebar (mais o atalho
**Usuários** quando o usuário é admin), navegação em drawer no mobile, feedback por
toasts, confirmações em diálogo e transições de página via View Transitions API.
Tudo respeita `prefers-reduced-motion`.

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
│  └─ adr/                 # ADR-0001..0008
├─ docker-compose.yml · CHANGELOG.md · README.md
```

> Os pacotes de domínio (`projects`, `monitor`, `tasks`, `jobs`, `cv`) e o shared
> kernel `skills` entram em `apps/api/internal/` a partir da F1. As migrações ficam
> sob `apps/api/db/` (não na raiz) por causa do `go:embed`, que não alcança fora do
> módulo.

## Decisões de arquitetura (ADRs)

Registros completos em [`docs/adr/`](docs/adr/) (ADR-0001..0008). Destaques:

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
- **ADR-0008 — Multiusuário.** Isolamento por `user_id` (id do dono no contexto da
  requisição, escopo em todos os repositórios); cadastro aberto com **aprovação do
  admin**; RBAC mínimo (admin/user). Supera o adiamento de RBAC do ADR-0007.

## Desenvolvimento

Pré-requisitos: **Docker** (obrigatório) e, opcionalmente, **Node 20+** para rodar
o frontend fora de container. Go **não** é necessário no host — os alvos Go do
`Makefile` rodam na imagem oficial `golang` via Docker.

```bash
docker compose up --build       # sobe libSQL (sqld) + API + web (dev)
# web:  http://localhost:5173
# api:  http://localhost:8080  (healthz, /api/auth/*)
# login de dev (admin): owner@example.com / change-me-dev  (definido no compose)
```

Em produção (sem `OWNER_EMAIL`) a instância sobe **sem conta**: o **primeiro
cadastro pela UI** vira o **admin** e já entra. Cadastros seguintes nascem
**pendentes** e só logam depois que o admin os ativa em **Usuários**
(`/admin/usuarios`). Para testar o primeiro acesso localmente, comente
`OWNER_EMAIL`/`OWNER_PASSWORD` no `docker-compose.yml` e zere o volume
(`docker compose down -v`).

Esqueceu a senha? A tela **"Esqueci minha senha"** (`/forgot-password`) envia um
link de redefinição por e-mail. No stack de dev (`docker compose`) o e-mail vai
para o **Mailpit** — abra **http://localhost:8025** para lê-lo. Sem nenhum SMTP, o
fluxo ainda funciona: o link é impresso no **log da API** (`reset_url`). Use o
e-mail de uma **conta existente** (no dev, o admin `owner@example.com`) — endereços
sem conta não disparam envio (proteção contra enumeração). O link é de uso único, expira em
`PASSWORD_RESET_TTL` (1h) e, ao ser usado, encerra todas as sessões. Para entrega
real (Gmail/provedor), troque os `SMTP_*` por credenciais de verdade.

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

## Deploy (Fly.io)

Guia completo passo a passo em **[docs/DEPLOY.md](docs/DEPLOY.md)**. Resumo:

A API tem um [`apps/api/fly.toml`](apps/api/fly.toml) pronto. **Uma única máquina**
(`fly scale count 1`) — o hub SSE, o scheduler do Monitor e o compactor de rollups
são in-process e pressupõem um só writer (ADR-0002). O banco é **Turso/libSQL**
hospedado; não há volume.

```bash
cd apps/api
fly launch --no-deploy            # cria o app (ou ajuste app/region no fly.toml)
fly secrets set DATABASE_URL=libsql://<db>.turso.io DATABASE_AUTH_TOKEN=... \
  APP_SECRET_KEY="$(openssl rand -base64 32)" WEB_ORIGIN=https://<web-host>
fly deploy && fly scale count 1
```

O **admin** é criado no **primeiro cadastro (signup)** — sem `OWNER_*` em
produção; cadastros seguintes ficam pendentes até o admin ativá-los. O frontend
(SvelteKit, `adapter-node`) é publicado à parte (Fly, Vercel, etc.), apontando
`API_URL`/`WEB_ORIGIN` para a API.

## Segurança

Segredos só via env; CORS restrito; headers de segurança; validação/sanitização
de toda entrada (inclusive links de vaga); credenciais nunca em texto puro
(cifradas em repouso); contêineres non-root. Ver a seção de riscos no plano.

## Licença

Privado — Lumni.
