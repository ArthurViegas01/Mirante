# Changelog

Todas as mudanças relevantes deste projeto são documentadas aqui.

O formato segue [Keep a Changelog](https://keepachangelog.com/pt-BR/1.1.0/)
e o versionamento segue [SemVer](https://semver.org/lang/pt-BR/).

Este arquivo é a **fonte de verdade do histórico** do Mirante.

## [Não lançado]

### Adicionado
- **Recuperação de senha por e-mail.** Novo fluxo "Esqueci minha senha" para o
  dono único: `POST /api/auth/forgot-password` emite um token de uso único (hash
  SHA-256 no banco, validade de 1h, tabela `password_resets` na migração 0015) e
  entrega o link; `POST /api/auth/reset-password` define a nova senha e **revoga
  todas as sessões**. Telas `/forgot-password` e `/reset-password` (cards de
  marca, validação inline) e link na tela de login; o guard do layout libera as
  rotas de recuperação para visitantes deslogados. Envio por SMTP
  (`internal/platform/mailer`, STARTTLS/465, auth PLAIN opcional); sem `SMTP_HOST`
  o link é registrado no log da API (`reset_url`) para uso em dev. O pedido
  responde sempre `200` para não revelar se um e-mail tem conta, com rate limit
  por endereço. O stack de dev (`docker compose`) inclui um **Mailpit**
  (`localhost:8025`) como caixa de captura. Novas envs:
  `SMTP_HOST/PORT/USERNAME/PASSWORD/FROM`, `PASSWORD_RESET_TTL`.
- **Dashboard inicial (`/`).** Painel que reúne as quatro áreas: KPIs (projetos
  ativos, serviços no ar, tarefas abertas/atrasadas, custo mensal, pipeline),
  painel "ao vivo" do Monitor, foco do dia (tarefas a vencer) e snapshot de
  carreira. O pós-login passa a abrir aqui (antes caía em `/projetos`) e a sidebar
  ganhou o item **Início**. Dados via endpoints existentes (sem mudança de API).
- **Logout e identidade do dono na UI.** Novo `UserMenu` no rodapé da sidebar
  (avatar gradiente Beam→Glow, nome, e-mail) com ação de **Sair**
  (`POST /api/auth/logout`, que já existia na API mas não tinha interface). Fecha
  ao clicar fora e no Escape.
- **Primitivos de UX reutilizáveis** em `apps/web/src/lib/components/`: toasts
  (`Toaster` + store `toast`), diálogos (`Modal`, `ConfirmHost` + `confirm.ask()`
  no lugar do `confirm()` nativo), `EmptyState`, `Skeleton`, `StatCard`,
  `Textarea`, `BrandMark`. `Button` ganhou variante `danger` e `full`; `Input`
  ganhou estado de erro, hint, autocomplete e botão de mostrar/ocultar senha.
- **Navegação mobile.** A sidebar vira drawer (hambúrguer na topbar, backdrop,
  fecha ao navegar/Escape); a topbar mostra o título da seção. `aria-current` no
  item ativo e `aria-expanded` nos controles expansíveis.
- **Motion.** Transições de página via View Transitions API e entradas sutis no
  dashboard, todas sob `prefers-reduced-motion`.

### Alterado
- **Login e signup redesenhados** como cards de marca (lockup, validação inline,
  reveal de senha, autocomplete correto). O **login é a porta de entrada única**:
  todo acesso anônimo cai nele, com botão fixo **Criar conta** e link de
  recuperação de senha; o signup virou um "Criar conta" simples (sem o
  enquadramento de "primeiro acesso") e mostra um aviso de **cadastro encerrado**
  quando já existe uma conta (o back-end segue single-owner). Deixou de haver
  redirecionamento automático para `/signup`.
- **Páginas refinadas** (projetos, tarefas, custos, vagas, cv, candidaturas) e os
  componentes `ProjectStacks`/`ProjectCosts`: skeletons no lugar de "Carregando…",
  empty states com CTA, toasts em todas as mutações (inclusive ações de IA) e
  confirmação em diálogo nas ações destrutivas (incluindo a remoção de link, que
  antes não pedia confirmação). Styleguide atualizado como referência viva dos
  novos componentes.

### Corrigido
- **As fontes não carregavam.** Os `.woff2` de Figtree, Instrument Serif e Geist
  Mono não estavam versionados, então todo o `@font-face` caía nas fontes do
  sistema e a tipografia do design system não era aplicada no app no ar. Os
  arquivos foram embarcados em `apps/web/static/fonts/` (`font-display: swap`).
- **Sessão expirada (401) tratada no cliente.** Um 401 em rota protegida agora
  limpa a sessão e redireciona ao login, em vez de falhar em silêncio.

### A fazer
- Deploy real (Fly + Turso) e front hospedado; multi-usuário/RBAC quando houver
  necessidade (ADR-0007). Multi-instância (Redis + leader election) segue adiada
  (ADR-0002).

## [0.7.0] - 2026-06-09

F4 + F5 + prontidão de produção: rollups/pruning do Monitor, observabilidade OTLP,
webhooks de alerta, signup do dono, banco hospedado (Turso/libSQL) e configuração
de deploy (Fly.io).

### Adicionado
- **Observabilidade: OTLP traces (F5, ADR-0007).** O `internal/platform/otel`
  deixou de ser no-op: com `OTEL_EXPORTER_OTLP_ENDPOINT` setada, a API exporta
  traces via **OTLP/HTTP** (provider real do SDK, resource com `service.name`,
  propagação W3C `traceparent`/baggage) e embrulha o handler raiz com `otelhttp`
  (um span server por request, método/status como atributos). Sem endpoint, segue
  no-op (zero overhead em dev/testes); falha no exporter degrada para no-op em vez
  de derrubar o boot. **RBAC fica adiado** enquanto o app for single-user (ADR-0007).
- **Webhooks de alerta (F5).** Cada transição de serviço do Monitor (up/degraded/
  down) pode ser entregue a um endpoint externo: com `ALERT_WEBHOOK_URL` (http/https)
  setada, a API faz `POST` JSON (evento, severidade, título legível, from/to, etc.)
  a cada alerta, sobre a interface `AlertChannel`/`Notifier` que já existia (erros
  isolados, timeout por chamada). A URL é do dono (confiável) — sem SSRF-guard,
  diferente do import de vaga. Vazio = só o alerta in-app (sino). Sem migração.
- **Deploy no Fly.io (config).** `apps/api/fly.toml` pronto para a API: máquina
  única (hub SSE + scheduler + compactor são in-process — ADR-0002), healthcheck,
  `force_https`, banco Turso hospedado (sem volume), segredos via `fly secrets`,
  dono via signup. Seção **Deploy (Fly.io)** no README.
- **Signup do dono + banco hospedado (prontidão para deploy).** O acesso deixa de
  exigir o owner por env: se `OWNER_EMAIL` não for fornecido, a instância sobe sem
  dono e o **primeiro acesso pela própria UI** reivindica a conta do dono
  (`POST /api/auth/signup`, single-user — o cadastro fecha depois com
  `ErrSignupClosed`; criação atômica via `CreateFirst` numa transação). Novo
  `GET /api/auth/status` (`{needs_setup}`) deixa o SPA rotear o visitante anônimo
  para `/signup` (página nova) vs `/login`; o guard do layout foi generalizado. O
  bootstrap por env vira atalho de dev (idempotente, opcional). O banco já fala
  **libSQL/Turso** por scheme (`DATABASE_URL=libsql://…` + `DATABASE_AUTH_TOKEN`) —
  habilitando um deploy hospedado sem owner-por-env.
- **Monitor: rollups horários + pruning (F4, ADR-0006).** O histórico bruto de
  checks (`check_results`, ~1 linha/min por serviço) agora é compactado para não
  crescer sem limite. Um worker horário (na inicialização e a cada hora) agrega os
  checks mais antigos que `MONITOR_RETENTION_DAYS` (env, default 14) em buckets
  horários (`check_rollups`, migração `0014`: `samples`/`ups`/`sum_latency_ms` por
  serviço×hora) e **poda** as linhas brutas, numa única transação. O uptime
  (24h/7d/30d) passa a somar rollups + brutos — disjuntos no tempo, **sem dupla
  contagem nem perda** de amostras; a sparkline continua lendo os brutos recentes.
  Multi-instância segue fora de escopo (ADR-0002: app single-instance).

## [0.6.0] - 2026-06-09

F3 — Vagas, CV & CRM: busca de carreira fim-a-fim com IA (Groq), sobre o kernel
`skills`. Importar/colar vagas e CV, % de aderência, adaptação por vaga, export
PDF/DOCX e pipeline de candidaturas.

### Adicionado
- **Kernel `internal/skills`** (fundação da F3): vocabulário canônico de skills com
  sinônimos/ontologia de categorias em dados Go in-code (sem banco, sem HTTP, sem
  dependências). API determinística — `Normalize(raw)` resolve um token para a skill
  canônica e `Match(texto)` extrai as skills mencionadas em texto livre (boundary-aware,
  trata `C#`/`C++`/`.NET`/multi-palavra). É o piso do cálculo de aderência (jobs/cv).
- **Gateway LLM `internal/llm`** (ADR-0004): provider único por env (default **Groq**,
  OpenAI-compatible), sem failover. `Client` aplica rate-limit por rota e grava todo
  uso num ledger (`llm_usage`, migração `0007`); `CompleteJSON` pede saída JSON e valida
  por unmarshal no tipo do chamador. Provider `mock` para dev sem key (a API sobe e as
  features de LLM degradam). Keys: `LLM_API_KEY`/`GROQ_API_KEY` etc.
- **Domínio `jobs` (Vagas):** CRUD de vagas com REST em `/api/jobs` e migração `0008`
  (`jobs` + `job_skills`). As skills exigidas são **extraídas deterministicamente** da
  descrição via `skills.Match`; `POST /api/jobs/{id}/enrich` usa o LLM para preencher
  empresa/senioridade/modelo/resumo. UI em `/vagas` (lista, cadastro colando a descrição,
  botão "Enriquecer com IA"). O compose repassa `GROQ_API_KEY` (de um `.env` na raiz).
- **Import de vaga por link:** `POST /api/jobs/import` busca a URL (fetch SSRF-guard,
  ADR-0003, UA de browser) e extrai os campos do JSON-LD `JobPosting` que LinkedIn e
  boards embutem (com fallback por LLM quando ausente). Na UI de `/vagas`, colar o link
  preenche o formulário automaticamente (título, empresa, local, descrição, modelo).
- **Perfil (domínio `cv`) + profissão no header de Vagas:** novo domínio `cv` com o
  perfil mestre singleton em `/api/profile` (migração `0009`, upsert). O header de
  `/vagas` mostra a **profissão atual e a almejada** (editável inline) — fundação do
  CV mestre que cresce na sequência (experiências, adaptação por vaga, export PDF/DOCX).
- **CV: skills mestre + aderência vaga↔CV:** o domínio `cv` ganhou as **skills mestre**
  do dono (`/api/profile`, migração `0010`, canonicalizadas via skills kernel) e a
  página **`/cv`** para gerenciá-las. Em `/vagas`, cada vaga mostra o **% de aderência**
  (overlap das skills exigidas com as suas) e destaca as que você tem (verde) vs as
  faltantes — base determinística do match (refino por LLM virá com as experiências).
- **CV mestre: experiências e educação** geríveis na página `/cv` (listas com add/remover),
  persistidas via `PUT /api/cv` (migração `0011`, replace atômico, ids no servidor).
- **Import de CV por texto:** `POST /api/cv/import` usa o LLM para estruturar um CV (ou
  inventário de skills) colado em {identidade, resumo, skills, experiências, educação};
  a `/cv` preenche o formulário p/ revisão. Validado ao vivo com um CV real (Groq).
- **Export do CV em PDF e DOCX:** `GET /api/cv/export?format=pdf|docx` renderiza o CV
  mestre (PDF via `go-pdf/fpdf` pure-Go; **DOCX via OOXML escrito à mão** com
  `archive/zip`, sem dependência nova). Botões na `/cv` salvam e baixam. Inclui o novo
  campo **contato** (migração `0012`), preenchido também pelo import.
- **Adaptação do CV por vaga (LLM):** `POST /api/cv/adapt` — dado o CV mestre + uma vaga,
  o LLM gera um **resumo adaptado** para aquela vaga + uma **análise** (pontos fortes,
  lacunas, dica). Botão "🎯 Adaptar CV" em cada vaga. cv não importa jobs (a vaga chega
  por input — ADR-0001).
- **CRM de candidaturas (`internal/applications`):** novo domínio + REST `/api/applications`
  (migração `0013`); pipeline de status (interesse → aplicado → entrevista → oferta →
  aceito/rejeitado), follow-up (próxima ação + data) e notas. Referencia a vaga por
  `job_id` (soft-link, com snapshot de título/empresa — sem importar `jobs`). UI
  **`/candidaturas`** (pipeline, mudar status inline, editar) + botão **"Acompanhar"** em
  cada vaga.

### Alterado
- **Monitor agora é centrado no projeto:** a aba Monitor saiu da sidebar; o status ao
  vivo de cada serviço (front/back/banco) aparece na seção **Stacks** ao abrir um
  projeto, agora com **sparkline de latência, uptime (24h/7d/30d) e pausar/excluir**
  inline (cards expansíveis). Alertas seguem na central de notificações (sino). As
  rotas `/monitor` foram removidas; os endpoints `/api/services*` permanecem.

### Corrigido
- **Dark mode:** header de tabela (Projetos), colunas do kanban (Tarefas) e hover de
  botões usavam cores sempre-claras; agora usam o role token `--color-surface-sunken`
  e respeitam o tema.
- **Import de vaga por link:** a descrição vinha resumida; agora devolve o texto
  completo da vaga (prompt ajustado + limites de entrada/saída maiores).
- **CV:** editar a profissão pelo header de Vagas não apaga mais as skills mestre
  (`PUT /api/profile` virou update parcial — `skills` opcional).

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

[Não lançado]: https://example.com/mirante/compare/v0.7.0...HEAD
[0.7.0]: https://example.com/mirante/compare/v0.6.0...v0.7.0
[0.6.0]: https://example.com/mirante/compare/v0.5.0...v0.6.0
[0.5.0]: https://example.com/mirante/compare/v0.4.0...v0.5.0
[0.4.0]: https://example.com/mirante/compare/v0.3.0...v0.4.0
[0.3.0]: https://example.com/mirante/compare/v0.2.0...v0.3.0
[0.2.0]: https://example.com/mirante/compare/v0.1.0...v0.2.0
[0.1.0]: https://example.com/mirante/releases/tag/v0.1.0
