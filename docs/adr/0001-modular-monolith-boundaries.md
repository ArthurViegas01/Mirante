# ADR-0001 — Fronteiras do monólito modular

- **Status:** aceito
- **Data:** 2026-06-07

## Contexto

O backend é um monólito modular em Go com pacotes de domínio (`projects`,
`monitor`, `tasks`, `jobs`, `cv`). A DoD exige baixo acoplamento entre domínios e
Go idiomático. Sem disciplina, domínios passam a se importar mutuamente e o
"monólito modular" vira um monólito emaranhado.

## Decisão

1. **Pacotes de domínio nunca se importam.** Toda referência cross-domain é feita
   por **interfaces (ports)** definidas pelo **consumidor** e implementadas/ligadas
   em `cmd/server` (composition root).
2. **Shared kernel explícito.** Conceitos compartilhados — `skills` (ontologia
   canônica de skills/keywords), `tags`, IDs tipados — vivem em pacotes próprios,
   dependidos "para dentro". `internal/skills` é uma adição consciente à lista do
   briefing, justificada por ser consumido por `jobs` **e** `cv`.
3. **Anatomia de cada domínio:** `types` (entidades), `service` (casos de uso),
   `repo` (interface) + `sqlite/` (implementação), `http` (handlers).
4. **`platform` é a folha de infraestrutura** (config, db, http, auth, otel, etc.),
   importada por todos, importando nenhum domínio.
5. **IDs tipados por domínio** (`type ProjectID string`) para segurança em tempo de
   compilação ao cruzar fronteiras.

## Consequências

- Não há FKs cross-domain garantidas no banco; integridade referencial entre
  domínios é tratada por ports/eventos (ex.: limpeza ao arquivar/excluir projeto).
- Acoplamento de compilação fica baixo e testável; cada domínio é testável isolado
  com fakes das interfaces.
- O custo é mais "fiação" no `cmd/server` — aceitável e explícito.
