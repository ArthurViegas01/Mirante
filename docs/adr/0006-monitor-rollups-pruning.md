# ADR-0006 — Rollups horários e pruning do histórico de checks

- **Status:** aceito
- **Data:** 2026-06-09

## Contexto

O Monitor grava uma linha em `check_results` por checagem (~1/min por serviço).
Sem poda, a tabela cresce sem limite — um serviço a 60s gera ~525k linhas/ano — e
o uptime de longo prazo (24h/7d/30d) faz `COUNT(*)`/`SUM` sobre essa série bruta
inteira. O dono quer manter o uptime de longo prazo barato e a tabela limitada,
**sem** introduzir multi-instância (que ADR-0002 já adia: o app é single-instance).

## Decisão

1. **Rollups horários.** Migração `0014` cria `check_rollups(service_id, bucket,
   samples, ups, sum_latency_ms, PK(service_id,bucket))`, com `bucket` no formato
   `'YYYY-MM-DDTHH'` (UTC) — exatamente `substr(checked_at,1,13)` do timestamp ISO
   canônico. Cada bucket resume uma hora de um serviço: `samples`, `ups`
   (`outcome != 'down'`, i.e. up+degraded) e `sum_latency_ms`.
2. **Compactação transacional + poda.** Um worker horário (espelha o *session
   sweep* no `cmd/server`, mais uma passada no boot) chama `repo.Compact(before)`:
   numa transação, agrega em `check_rollups` (UPSERT somando) as linhas com
   `checked_at < before` e **deleta** essas linhas brutas. `before` é
   `now − MONITOR_RETENTION_DAYS` (default 14) **truncado para a hora**, então
   nenhuma hora fica partida entre rollup e bruto.
3. **Uptime = rollups + brutos, disjuntos.** A janela soma os rollups com
   `bucket >= hora-de-início` **mais** os brutos com `checked_at >= início`. Como a
   compactação **move** (insere rollup + deleta bruto) na mesma transação, cada
   checagem física está num lado só — somar não dupla-conta nem perde amostras. A
   sparkline continua lendo só os brutos recentes (`ListChecks`).
4. **`sum_latency_ms` é reservado.** A compactação é unidirecional (a latência por
   checagem some ao podar); guardar a soma por bucket preserva, sem custo, a média
   de latência em janelas longas — ainda não exposta na API (como o `job_id`
   reservado da F2). Sem isso, seria irrecuperável.

## Consequências

- `check_results` fica limitada à janela de retenção; o histórico antigo vive
  comprimido em ~1 linha/serviço/hora. O uptime de 30d lê poucos rollups + os
  brutos recentes, em vez de dezenas de milhares de linhas.
- Granularidade do histórico antigo cai para a hora (aceitável para uptime de
  longo prazo); os dados recentes (dentro da retenção) seguem em resolução cheia.
- Re-rodar a compactação é seguro (idempotente): o UPSERT soma e, com `before`
  alinhado à hora e o tempo só avançando, uma hora já compactada não tem brutos
  sobreviventes para re-somar.
- Multi-instância continua fora de escopo (ADR-0002). O compactor é in-process,
  single-writer — coerente com o libSQL single-node.
