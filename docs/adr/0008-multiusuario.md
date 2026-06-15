# ADR-0008 — Multiusuário: isolamento por usuário, cadastro com aprovação e RBAC mínimo

- **Status:** aceito (supera o adiamento de RBAC/multiusuário do [ADR-0007](0007-observabilidade-e-rbac.md))
- **Data:** 2026-06-15

## Contexto

O Mirante nasceu **single-user** (um único dono, ver ADR-0007): nenhuma tabela de
domínio tinha noção de proprietário e nenhuma query filtrava por usuário. A
decisão de produto mudou: o app passa a aceitar **vários usuários**, cada um com
seu próprio Mirante privado, mantendo o mesmo deploy de instância única (ADR-0002).

Isso exige: (1) isolar os dados por usuário sem vazamento entre contas — uma
fronteira de **segurança**, não só de UX; (2) um caminho de cadastro que o dono
controle; e (3) um papel de administrador, sem reescrever RBAC completo.

## Decisão

1. **Isolamento por `user_id`, escopo no repositório.** Toda tabela de domínio
   ganha `user_id` (migração 0016; backfill para o dono existente). O id do dono
   da requisição viaja no `context.Context` (pacote `internal/platform/tenant`),
   injetado pelo middleware de auth logo após autenticar. **Cada repositório lê o
   `tenant.UserID(ctx)` e filtra todas as queries** (INSERT carimba, SELECT/
   UPDATE/DELETE filtram). A rede de segurança são **testes de isolamento por
   domínio** (usuário A não vê/edita/exclui dados de B), inclusive via joins.
   - Vocabulário antes global (codinome de projeto, tags, skills de CV) passa a
     ser **único por usuário**. O CV deixa de ser singleton e vira um perfil por
     usuário.
   - `user_id` é coluna simples, **sem FK**: o libSQL roda com `foreign_keys`
     desligado (a FK seria decorativa em produção). A limpeza referencial na
     exclusão de usuário é feita **explicitamente** (purge transacional).

2. **Monitor: separação sistema × usuário.** O scheduler/engine/compactor rodam
   em **nível de sistema** (contexto sem tenant) e usam métodos deliberadamente
   **não escopados** (`ListEnabledServices`, `RecordCheck`, `Compact`) — o
   scheduler precisa enxergar os serviços de todos para checá-los. Ao registrar
   um check, o alerta e o evento de SSE são **carimbados com o dono do serviço**.
   O hub de SSE faz **fan-out por usuário** (entrega só às conexões do dono) e o
   replay é escopado pelo dono da conexão. O histórico de alto volume
   (`check_results`/`check_rollups`) não ganha `user_id` e é isolado **via o
   serviço-pai** nas leituras.

3. **Cadastro aberto com aprovação do admin.** Qualquer um pode se cadastrar, mas
   a conta nasce `pending` e **não loga** até ser ativada. A **primeira** conta
   (bootstrap por env ou primeiro signup) é o **admin** e já entra ativa; as
   demais ficam `pending` (HTTP 202) e o login recusa (403) até a ativação. O
   `users` ganha `role` (`admin`|`user`) e `status` (`pending`|`active`|
   `disabled`) (migração 0017). `status` faz *fail-closed* (default `pending`).

4. **RBAC mínimo.** Só dois papéis. O admin tem uma API e uma tela
   (`/api/admin/users`, `/admin/usuarios`) para listar, criar, ativar, desativar
   e excluir contas — protegidas por `RequireAdmin` (403 para não-admin), com
   guarda contra auto-exclusão/auto-desativação. Excluir um usuário **purga todos
   os seus dados**. Não há papéis por recurso nem compartilhamento entre contas —
   isso fica para um ADR futuro se houver necessidade.

## Consequências

- O isolamento é uma invariante de **segurança**: vive na camada de dados e é
  coberto por testes por domínio + um teste de fan-out por usuário no SSE.
- O deploy segue de **instância única** (ADR-0002): o hub e o scheduler em memória
  pressupõem um só processo; multiusuário não muda isso.
- O modelo é **multi-tenant por isolamento**, não colaborativo: não há dados
  compartilhados entre usuários. Compartilhamento/papéis por recurso seriam um
  novo ADR.
- Sem verificação de e-mail no cadastro (v1): a barreira de abuso é a **aprovação
  do admin**. Verificação por e-mail pode ser adicionada depois.
- `OWNER_EMAIL`/`OWNER_PASSWORD` continuam sendo o atalho de bootstrap do admin;
  o nome da env foi mantido por compatibilidade, mas o papel é "admin".
