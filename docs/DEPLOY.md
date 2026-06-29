# Deploy — Mirante (Railway + Turso)

Guia para colocar o Mirante no ar. Topologia alvo:

```
  navegador ──▶ web (SvelteKit, adapter-node)  ──/api──▶  API (Go, Railway, 1 réplica)
                                                              │
                                                              ▼
                                                   Turso / libSQL (hospedado)
```

- **API** (`apps/api`): um binário Go (distroless) numa **única réplica** no Railway.
  O hub SSE, o scheduler do Monitor e o compactor de rollups são in-process e
  assumem um só writer, então **não escale horizontalmente** (ADR-0002).
- **Banco**: Turso/libSQL gerenciado (a API fala `libsql://` por scheme).
- **Web**: build Node (adapter-node) publicado à parte, apontando o proxy `/api`
  para a URL da API.
- **Dono**: semeado no boot via `OWNER_EMAIL` + `OWNER_PASSWORD_HASH` (recomendado),
  fechando a janela do primeiro signup (F3).

## Pré-requisitos

- Conta no [Railway](https://railway.app) e a CLI [`turso`](https://docs.turso.tech/cli/installation)
  autenticada (`turso auth login`).
- Docker local é útil para validar o build, mas o Railway constrói a imagem pelo
  `Dockerfile` do projeto.

## 1. Banco (Turso)

```bash
turso db create mirante-prod
turso db show mirante-prod --url          # → libsql://mirante-prod-<org>.turso.io
turso db tokens create mirante-prod       # → token (guarde; vira secret)
```

> Backups: o Turso versiona/branch-eia o banco; veja `turso db` para snapshots.

## 2. Hash do dono (offline)

Gere o hash Argon2id da senha do dono **antes** do deploy (nunca commite a saída):

```bash
cd apps/api
go run ./cmd/hashpw           # digita a senha; imprime $argon2id$v=19$m=65536,t=3,p=2$...
```

## 3. API (Railway)

Crie um serviço a partir do repositório e, em **Settings**, defina
**Root Directory = `apps/api`** (o `railway.toml` aponta o `Dockerfile`). Configure
as variáveis no painel:

```
# Secrets (aba Secrets)
DATABASE_URL         libsql://mirante-prod-<org>.turso.io
DATABASE_AUTH_TOKEN  <token do passo 1>
WEB_ORIGIN           https://<host-do-front>
GROQ_API_KEY         <opcional, p/ features de IA>
OWNER_EMAIL          voce@exemplo.com
OWNER_PASSWORD_HASH  <saída do passo 2>

# Variables (aba Variables, não-secretas)
APP_ENV                 production
TRUSTED_PROXY           true        # IP real do cliente atrás do edge do Railway
OTEL_SERVICE_NAME       mirante-api
MONITOR_RETENTION_DAYS  14
```

Notas:

- `TRUSTED_PROXY=true` faz o rate-limit por IP ler o IP real do cliente do header
  do edge do Railway (`X-Envoy-External-Address`). Sem isso, atrás do proxy, todos
  os requests cairiam no mesmo bucket. **Nunca** ligue isso numa API exposta
  direto (o header seria forjável, F4).
- Semeie `OWNER_EMAIL` + `OWNER_PASSWORD_HASH` antes do primeiro deploy: o dono
  nasce no boot e o cadastro público já sobe fechado (F3).
- A API escuta na porta `$PORT` injetada pelo Railway (`config.go` → `httpAddr()`).
- Para traces, adicione `OTEL_EXPORTER_OTLP_ENDPOINT` apontando para um coletor.

## 4. Web (front)

Crie um segundo serviço com **Root Directory = `apps/web`** (usa
`@sveltejs/adapter-node`). Em runtime ele precisa:

- `API_URL` → URL pública da API (ex.: `https://mirante-api.up.railway.app`) para
  o proxy `/api`.
- `ORIGIN` → URL pública do front (adapter-node, checagem de origem).
- O domínio do front deve bater com o `WEB_ORIGIN` setado na API (CORS + CSRF).

## 5. Primeiro acesso

Com o dono semeado (passo 2/3), abra a URL do front e entre direto em **`/login`**
com `OWNER_EMAIL` e a senha cujo hash você gerou. Cadastros de terceiros nascem
**pendentes** e só logam depois que o dono os ativa em **Usuários**
(`/admin/usuarios`).

> Sem `OWNER_*` (apenas fora de produção), o app vai para `/signup` e o primeiro
> cadastro vira o dono. Em produção, sempre semeie o dono.

## 6. Verificação

```bash
curl -s https://<api-host>/healthz                 # 200
curl -s https://<api-host>/api/auth/status         # {"needs_setup":false} com dono semeado
```

- Faça login, crie um **serviço** no Monitor e confirme o status ao vivo (SSE).
- Webhook (opcional): setar `ALERT_WEBHOOK_URL` → cada transição
  (up/degraded/down) faz `POST` JSON no endpoint.
- Traces (opcional): com `OTEL_EXPORTER_OTLP_ENDPOINT`, confira spans no coletor.

## Operação

- **Logs**: painel do Railway (ou `railway logs`). O boot mostra migrações
  aplicadas, estado do dono, LLM, webhook e OTel.
- **Migrações**: aplicadas automaticamente no boot (goose embarcado), deploy basta.
- **Não escalar**: mantenha **1 réplica** (ADR-0002). Mais carga: instância maior,
  não mais réplicas.
- **Rotação de segredos**: troque a variável no painel (redeploya). Para girar a
  senha do dono, gere um novo hash (passo 2) e atualize `OWNER_PASSWORD_HASH`.
- **Rollback**: redeploy de um deployment anterior pelo painel do Railway.

## Alternativa: Fly.io (legado)

Há um [`apps/api/fly.toml`](../apps/api/fly.toml) pronto (máquina única,
`fly scale count 1`). Lá, use `TRUSTED_PROXY_HEADER=Fly-Client-IP` e
`fly secrets set` para os mesmos segredos (incluindo `OWNER_EMAIL` +
`OWNER_PASSWORD_HASH`).
