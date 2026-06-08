# ADR-0003 — Dois domínios de confiança para fetch de saída

- **Status:** aceito
- **Data:** 2026-06-07

## Contexto

O backend faz requisições HTTP de saída em **dois contextos com confiança oposta**:

- **Monitor:** sonda a infraestrutura do próprio dono (inclusive IPs privados).
- **Links de vaga:** busca URLs **externas, influenciadas por terceiros** (o dono
  cola um link de recrutador). Esse é um vetor clássico de **SSRF**.

Usar um único cliente HTTP para os dois seria um buraco de segurança.

## Decisão

`platform/httpx` expõe um `SafeFetcher` **parametrizado por política** (a partir
da F1, primeiro consumidor):

- **Política Monitor:** permite IP privado/loopback; valida o esquema; registra o
  IP resolvido; sem redirects; timeout rígido; cap de tamanho do corpo.
- **Política JobLink:** **bloqueia** IP privado/link-local/metadata (ex.:
  `169.254.169.254`); fixa o IP validado via `DialContext` (anti DNS-rebind);
  re-valida esquema+IP a cada redirect (limite de saltos); rejeita não-HTTP(S)
  pós-redirect; `io.LimitReader` + read deadline.

## Consequências

- Duas políticas, testadas por branch (aceita privado vs. rejeita privado, rebind,
  redirect, scheme pós-redirect, cap de corpo).
- A regra do monitor "não bloquear IP privado" **nunca** vaza para o fetcher de
  vaga.
