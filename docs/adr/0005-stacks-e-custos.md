# ADR-0005 — Stacks rotuladas e custos por assinatura

- **Status:** aceito
- **Data:** 2026-06-08

## Contexto

Um projeto tem um stack distribuído por provedores (ex.: front no Netlify, back no
Railway, banco no Supabase). O dono quer, na própria visão do projeto, **ver o status
ao vivo de cada peça** e **acompanhar o custo das assinaturas** — por projeto e no
geral. O Monitor (F1) já vincula `services` a um projeto e deriva status (up/degraded/
down) via SSE; o que falta é rótulo e superfície. Custo é greenfield.

## Decisão

1. **Status = endpoints próprios.** "Status do stack" é o resultado da checagem do
   próprio Mirante no endpoint publicado de cada peça (reusa o Monitor; ADR-0003,
   política monitor). Status-de-provedor (status pages do Netlify/Railway/Supabase)
   é evolução futura — exigiria fetch externo (política JobLink) e parsing.
2. **Rótulos de stack no serviço.** `services` ganha `provider` (texto livre, ex.:
   "netlify") e `camada` (`frontend`/`backend`/`database`/`outro`), ambos opcionais.
   A project view agrupa os serviços do projeto por camada e mostra o status ao vivo.
   Sem rotas novas — `GET /api/services?project_id=` já serve.
3. **Custos num domínio próprio `subscriptions`.** Custo recorrente é um conceito
   independente do monitoramento: nem toda assinatura tem health check (domínio, SaaS,
   API paga) e nem todo serviço monitorado é pago. Um `subscriptions` separado
   (anatomia padrão de domínio) cobre ambos.
4. **`service_id` é soft-link, sem FK.** Uma assinatura pode apontar para um serviço
   monitorado (para exibir custo ao lado do status), mas **não** há FK
   `subscriptions → services`: são dois domínios distintos (ADR-0001). A FK real é
   apenas para o spine: `subscriptions.project_id → projects` `ON DELETE CASCADE`.
5. **Multi-moeda sem conversão.** `valor_cents` (inteiro) + `moeda` (`BRL`/`USD`) +
   `ciclo` (`mensal`/`anual`). Totais são **mensalizados** (anual ÷ 12) e somados
   **por moeda, separadamente** — nada de cotação/câmbio no v1.

## Consequências

- A project view vira a casa do stack: status por camada + custo por moeda, sem sair
  do projeto. A gestão fina do serviço (thresholds, intervalo) continua no Monitor.
- Dinheiro como inteiro evita erro de ponto flutuante; somar por moeda é honesto sem
  depender de fonte de câmbio (a conversão para BRL fica como evolução).
- `subscriptions` não importa `monitor` nem vice-versa; a correlação custo↔status é
  feita no cliente via `service_id`. Excluir um projeto leva junto seus serviços
  (CASCADE, já existente) e suas assinaturas (CASCADE).
