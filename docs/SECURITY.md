# Segurança — Mirante

Postura de segurança do Mirante (Go + SvelteKit, libSQL/Turso, deploy de instância
única no Railway). Este documento descreve os controles em vigor e os riscos
deliberadamente aceitos. Substitui o antigo plano de ação, cujos achados já foram
implementados (ver `CHANGELOG.md` → Segurança).

## Controles em vigor

| Controle | Onde |
|---|---|
| Senhas com Argon2id em parâmetros OWASP (m=64MiB, t=3, p=2) e comparação constant-time | `internal/platform/auth/password.go` (`HashPassword`/`VerifyPassword`) |
| Login constant-time: e-mail inexistente roda um verify Argon2id "dummy" pré-computado, sem oracle de tempo de enumeração | `internal/platform/auth/service.go` (`Service.dummyHash`, `Login`) |
| Sessão server-side: o cookie guarda só um token CSPRNG de 256 bits; o banco guarda apenas o SHA-256 do token | `internal/platform/auth/token.go`, `auth/store.go` |
| Cookie de sessão HttpOnly + Secure (prod) + SameSite=Strict | `internal/platform/httpserver/auth_handler.go` (`sessionCookie`) |
| CSRF por sessão: checagem de Origin + `X-CSRF-Token` em métodos inseguros, compare constant-time | `httpserver/auth_handler.go` (`CSRF`) |
| CORS restrito à WebOrigin exata, com credenciais e `Vary: Origin` | `httpserver/middleware.go` (`CORS`) |
| Security headers: CSP `default-src 'none'`, X-Frame-Options DENY, nosniff, Referrer-Policy no-referrer, HSTS em prod | `httpserver/middleware.go` (`SecurityHeaders`) |
| SQL 100% parameterizado (`database/sql` com `?`), sem concatenação de input | `internal/**/sqlite.go`, `auth/store.go` |
| Fetcher externo com guarda anti-SSRF: allowlist http/https, bloqueio de IP privado/loopback/link-local/CGNAT/NAT64, **IP pinning anti DNS-rebind** (resolução única dentro do `DialContext`; o IP validado é o discado), sem follow de redirect, cap de corpo e `Timeout` | `internal/platform/httpx/httpx.go` ([ADR-0003](adr/0003-two-trust-domain-fetch.md)) |
| Rate limit: login por e-mail (5/15min), signup por IP (5/h), global por IP (240/min); todos com teto de cardinalidade (memória limitada sob chaves atacante-controladas) | `auth/service.go`, `internal/platform/ratelimit/ratelimit.go` |
| IP do cliente para rate-limit só confia no header do edge quando `TRUSTED_PROXY=true` (prod: `X-Envoy-External-Address` no Railway); exposição direta cai no `RemoteAddr` | `httpserver/middleware.go` (`clientIP`), `cmd/server/main.go` |
| Timeouts do servidor: `ReadHeaderTimeout`, `ReadTimeout`, `IdleTimeout` (`WriteTimeout` omitido de propósito pelo SSE de longa duração) | `cmd/server/main.go` |
| Body com tamanho máximo (`MaxBytesReader`) e `DisallowUnknownFields` em todos os handlers | `internal/platform/respond/respond.go`, handlers |
| LLM: saída forçada a JSON e revalidada por unmarshal em tipo Go; sem tools/execução; rate-limit por rota + ledger de uso; conteúdo do usuário marcado como DADO; campos extraídos com cap de tamanho antes de persistir | `internal/llm/*`, `jobs/import.go`, `jobs/service.go` (`clampJobFields`), `cv/service.go` |
| Owner semeável no deploy (`OWNER_EMAIL` + `OWNER_PASSWORD_HASH`), fechando a janela do primeiro signup; utilitário `cmd/hashpw` gera o hash offline | `auth/service.go` (`Bootstrap`), `cmd/hashpw` |
| Container distroless non-root, binário estático sem CGO | `apps/api/Dockerfile` |
| Segredos só via env (Railway variables / `.env` local gitignored); nunca versionados | `.gitignore`, `railway.toml` |
| Frontend sem injeção de HTML: nenhum `{@html}`/`innerHTML` (Svelte 5 auto-escapa) | `apps/web/src` |

## Deploy seguro (Railway)

Ver [DEPLOY.md](DEPLOY.md). Pontos de segurança no provisionamento:

1. **Semear o dono antes do 1º deploy.** Gerar o hash com `go run ./cmd/hashpw` e
   configurar `OWNER_EMAIL` + `OWNER_PASSWORD_HASH`. Assim o admin nasce no boot e o
   cadastro público já sobe fechado (ninguém reivindica a instância no primeiro signup).
2. **`TRUSTED_PROXY=true`.** Necessário atrás do edge do Railway para o rate-limit
   por IP enxergar o IP real do cliente (`X-Envoy-External-Address`). **Nunca** ligar
   numa API exposta direto: o header seria forjável. O app emite um aviso no boot se
   estiver em produção sem a flag.
3. **Segredos só via variáveis do Railway** (`DATABASE_AUTH_TOKEN`, `GROQ_API_KEY`,
   etc.). Nada de segredo real em arquivo versionado.

## Riscos aceitos / fora de escopo

- **Chave do LLM no `.env` local.** Em dev, a `GROQ_API_KEY` fica num `.env` na
  raiz, **gitignored** (nunca versionado, histórico verificado). Em produção, só via
  variáveis do Railway. Se o `.env` for compartilhado/copiado para fora da máquina,
  rotacionar a chave no console da Groq.
- **Oracle de tempo no login (parcial).** O caminho de e-mail inexistente foi
  igualado (verify dummy), mas o limite de tentativas responde `429` antes do custo
  Argon2id, então um atacante ainda distingue "e-mail rate-limited" de "credencial
  inválida" pelo status. Aceito: o app é de baixo volume e o e-mail do dono é
  essencialmente público.
- **Casing de e-mail no login.** O limiter normaliza o e-mail para minúsculas, mas
  o lookup no banco é case-sensitive (SQLite). É um detalhe de UX pré-existente, não
  de segurança; tratar junto com a normalização de e-mail se/quando for mexido.
- **Slow-read de resposta (slowloris de leitura).** `WriteTimeout` é omitido de
  propósito para não matar o SSE; mitigado pelo cap de 32 clientes do hub SSE e
  pela evicção de consumidor lento.
- **Instância única / sem multi-tenant RLS.** O isolamento é por `user_id` em cada
  repositório (ADR-0008); não há RLS de banco. RBAC fino adiado (ADR-0007).
- **Prompt-injection no conteúdo de vaga/CV.** Risco residual aceito: a saída do LLM
  nunca é executada nem renderizada como HTML, é revalidada em tipo Go e tem cap de
  tamanho. Pior caso é dado adulterado que o próprio dono revisa.

## Não se aplica

- **SQL injection / XSS / path traversal:** queries parameterizadas; sem `{@html}` no
  front (Svelte auto-escapa); sem upload/download por caminho controlado pelo usuário.
- **Webhook de entrada:** não há; `ALERT_WEBHOOK_URL` é saída para endpoint do dono.
