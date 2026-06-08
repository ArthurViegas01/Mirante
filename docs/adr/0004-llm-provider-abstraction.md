# ADR-0004 — Abstração de provedor LLM

- **Status:** aceito
- **Data:** 2026-06-07

## Contexto

A análise de vagas e a geração de CV/carta usam LLM multi-provider
(Anthropic/Groq/OpenAI), trocável por env, com saída estruturada via JSON Schema.
Os dialetos de "structured output" divergem entre provedores.

## Decisão

1. **Provedor único selecionado por env** (`LLM_PROVIDER`). **Sem failover
   automático** entre provedores em tempo de execução — isso trocaria o modelo no
   meio de uma requisição e contradiz a segurança de schema. Retry é **dentro** do
   mesmo provedor.
2. **Saída estruturada via `response_format`/JSON Schema**, com uma camada de
   **adaptação de dialeto por provedor** (ex.: `additionalProperties:false` em todo
   objeto no modo estrito da OpenAI).
3. **Registry de schemas compilado no startup** (validadores reusados, não
   recompilados por requisição) e **re-validação da saída em Go**.
4. **Rate-limit por rota** (cota/custo do dono, não por IP) + **ledger `llm_usage`**
   (provider, modelo, tokens, custo) para limite e observabilidade.
5. Entrada (texto/link da vaga) é tratada como **dado, nunca instrução**
   (anti prompt-injection).

## Consequências

- Trocar de provedor é configuração; o caminho determinístico de aderência produz
  o mesmo score independentemente do provedor.
- Custo é limitado e auditável; nenhum envelope LLM é guardado para sempre
  (política de retenção/redaction).
