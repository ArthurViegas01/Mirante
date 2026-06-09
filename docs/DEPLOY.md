# Deploy — Mirante (Fly.io + Turso)

Guia para colocar o Mirante no ar. Topologia alvo:

```
  navegador ──▶ web (SvelteKit, adapter-node)  ──/api──▶  API (Go, Fly, 1 máquina)
                                                              │
                                                              ▼
                                                   Turso / libSQL (hospedado)
```

- **API** (`apps/api`): um binário Go (distroless) numa **única máquina** Fly.
  O hub SSE, o scheduler do Monitor e o compactor de rollups são in-process e
  assumem um só writer — **não escale horizontalmente** (ADR-0002).
- **Banco**: Turso/libSQL gerenciado (a API fala `libsql://` por scheme).
- **Web**: build estático/Node (adapter-node) publicado à parte (Fly, Vercel,
  Netlify…), apontando o proxy `/api` para a URL da API.
- **Dono**: criado no **primeiro acesso (signup)** — sem `OWNER_*` em produção.

## Pré-requisitos

- Contas e CLIs: [`flyctl`](https://fly.io/docs/flyctl/) e
  [`turso`](https://docs.turso.tech/cli/installation), ambos autenticados
  (`fly auth login`, `turso auth login`).
- Docker local (o build da imagem roda no Fly, mas é útil para validar).

## 1. Banco (Turso)

```bash
turso db create mirante-prod
turso db show mirante-prod --url          # → libsql://mirante-prod-<org>.turso.io
turso db tokens create mirante-prod       # → token (guarde; vira secret)
```

> Backups: o Turso versiona/branch-eia o banco; veja `turso db` para snapshots.

## 2. API (Fly)

A config já existe em [`apps/api/fly.toml`](../apps/api/fly.toml) (máquina única,
healthcheck `/healthz`, `force_https`, sem volume). A partir de `apps/api`:

```bash
cd apps/api
fly launch --no-deploy            # cria o app; confirme app name/região do fly.toml

# Segredos (NUNCA no fly.toml):
fly secrets set \
  DATABASE_URL="libsql://mirante-prod-<org>.turso.io" \
  DATABASE_AUTH_TOKEN="<token do passo 1>" \
  APP_SECRET_KEY="$(openssl rand -base64 32)" \
  WEB_ORIGIN="https://<host-do-front>" \
  GROQ_API_KEY="<opcional, p/ features de IA>"

fly deploy
fly scale count 1                 # garante exatamente UMA máquina (ADR-0002)
```

Variáveis não-secretas (`APP_ENV=production`, `HTTP_ADDR`, `OTEL_SERVICE_NAME`,
`MONITOR_RETENTION_DAYS`) já vêm do `[env]` do `fly.toml`. Para enviar traces,
adicione `OTEL_EXPORTER_OTLP_ENDPOINT` (secret ou env) apontando para um coletor.

## 3. Web (front)

O frontend usa `@sveltejs/adapter-node` (gera um servidor Node). Publique-o onde
preferir; em runtime ele precisa:

- `API_URL` → URL pública da API (ex.: `https://mirante-api.fly.dev`) para o proxy
  `/api`.
- O domínio do front deve bater com o `WEB_ORIGIN` setado na API (CORS + CSRF).

```bash
cd apps/web
npm ci
npm run build        # saída do adapter-node em build/
node build           # ou rode num container/host à sua escolha
```

## 4. Primeiro acesso (claim do dono)

1. Abra a URL do front. Sem dono, o app vai direto para **`/signup`**.
2. Crie a conta (e-mail + senha ≥ 8). Você vira o **dono**; o cadastro fecha.
3. A partir daí, `/login` é a porta de entrada.

## 5. Verificação

```bash
curl -s https://<api-host>/healthz                 # 200
curl -s https://<api-host>/api/auth/status         # {"needs_setup":true} antes do signup
```

- Faça login, crie um **serviço** no Monitor e confirme o status ao vivo (SSE).
- Webhook (opcional): setar `ALERT_WEBHOOK_URL` (secret) → cada transição
  (up/degraded/down) faz `POST` JSON no endpoint.
- Traces (opcional): com `OTEL_EXPORTER_OTLP_ENDPOINT`, confira spans no coletor;
  o log de boot mostra `otel tracing enabled`.

## Operação

- **Logs**: `fly logs`. O boot mostra migrações aplicadas, estado do dono, LLM,
  webhook e OTel.
- **Migrações**: aplicadas automaticamente no boot (goose embarcado) — deploy é
  suficiente.
- **Não escalar**: mantenha `count 1` (ADR-0002). Mais carga → máquina maior
  (ajuste `[[vm]]` no `fly.toml`), não mais máquinas.
- **Rotação de segredos**: `fly secrets set ...` (redeploya). Trocar
  `APP_SECRET_KEY` invalida segredos cifrados em repouso, se houver.
- **Rollback**: `fly releases` + `fly deploy --image <release anterior>`.
