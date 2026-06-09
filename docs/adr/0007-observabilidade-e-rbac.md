# ADR-0007 — Observabilidade (OTLP traces) e adiamento de RBAC

- **Status:** aceito
- **Data:** 2026-06-09

## Contexto

A F5 fecha o roadmap com "polish": observabilidade (o `internal/platform/otel`
era um stub no-op desde a F0), webhooks de alerta (já entregue) e **RBAC**. O
Mirante é **single-user**: existe um único dono, criado por signup; não há outros
usuários a quem atribuir papéis.

## Decisão

1. **Tracing OTLP/HTTP, opt-in por env.** `otel.Init` instala um
   `TracerProvider` real com exporter `otlptracehttp` **somente** quando
   `OTEL_EXPORTER_OTLP_ENDPOINT` está setado; sem ele, mantém o provider no-op
   (spans são livres mas não exportados — dev e testes não pagam nada). O
   exporter lê as variáveis padrão `OTEL_EXPORTER_OTLP_*` (endpoint/headers/TLS),
   então funciona com qualquer coletor compatível, sem vendor lock.
2. **Propagação W3C + um span por request.** Propagador composto
   (`TraceContext` + `Baggage`); o handler raiz é embrulhado por `otelhttp` (span
   server por request, método/status como atributos). Falha ao montar o exporter
   **degrada para no-op** em vez de derrubar o boot — telemetria nunca bloqueia o
   serviço. Spans finos (DB, checks do Monitor) ficam para quando houver
   necessidade real; o helper `otel.Tracer()` já permite adicioná-los sem mexer
   na composição.
3. **RBAC adiado até multi-usuário.** Enquanto o app for single-user, RBAC é um
   no-op (um só principal, o dono, com acesso total). Introduzi-lo agora seria
   complexidade especulativa. Fica adiado junto com o cadastro de múltiplos
   usuários; quando isso existir, um ADR próprio definirá papéis e enforcement.

## Consequências

- Em produção, basta apontar `OTEL_EXPORTER_OTLP_ENDPOINT` para um coletor
  (Grafana/Tempo, Honeycomb, Jaeger via OTLP, etc.) para ter traces de request
  ponta a ponta; sem isso, zero overhead.
- A F5 entrega observabilidade + webhooks; RBAC sai do escopo do v0.7.0 por
  decisão consciente (não por esquecimento), documentado aqui.
- Métricas e logs correlacionados (trace_id no slog) são evolução natural, fora
  do escopo deste ADR.
