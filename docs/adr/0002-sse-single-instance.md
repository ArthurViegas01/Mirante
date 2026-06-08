# ADR-0002 — Push do monitor via SSE + instância única

- **Status:** aceito
- **Data:** 2026-06-07

## Contexto

O Monitor precisa empurrar ao vivo transições de estado (up/degraded/down) e
alertas para o frontend. O fluxo é esmagadoramente server→client e os eventos são
raros e importantes. O app é single-user. O hub e o scheduler são in-memory.

## Decisão

1. **Transporte: SSE (Server-Sent Events).** `EventSource` reconecta sozinho e
   reenvia `Last-Event-ID`; o cookie de sessão acompanha um GET same-origin.
   Implementação em `net/http` puro, sem dependência de WebSocket.
2. **Sequência de eventos durável.** O id de evento é a **row id de
   `status_transitions`/`alerts`** (não um contador in-memory que zera no restart).
   Replay por `Last-Event-ID` cai para a tabela durável quando fora da janela do
   ring in-memory; caso contrário o cliente re-hidrata via `/snapshot`.
3. **Instância única.** O hub e o scheduler in-memory **não** são seguros sob
   escala horizontal (checks/alertas duplicados). Fly roda com `min=max=1` e o
   arquivo libSQL é single-writer/single-node.
4. **O stream carrega apenas transições + alertas.** Sparkline e uptime são
   buscados via REST/history sob demanda — sem firehose de latência.

## Consequências

- Sem escala horizontal no v1 (decisão consciente e documentada). Escalar exigiria
  um outbox/broker (Redis/NATS) e eleição de scheduler — fora de escopo.
- A troca para WebSocket fica isolada atrás da interface `EventSink`; só se
  justifica com necessidade real de canal bidirecional de alta frequência.
