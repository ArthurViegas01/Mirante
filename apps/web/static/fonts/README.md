# Fontes self-hosted

`app.css` referencia estes arquivos via `@font-face` (servidos em `/fonts/...`).
Eles agora estão versionados aqui (subsets latinos, OFL):

| Arquivo | Família | Peso/estilo | Origem |
|---|---|---|---|
| `figtree-variable.woff2` | Figtree | 300-900, normal (variável) | @fontsource-variable/figtree |
| `instrument-serif-regular.woff2` | Instrument Serif | 400, normal | @fontsource/instrument-serif |
| `instrument-serif-italic.woff2` | Instrument Serif | 400, itálico | @fontsource/instrument-serif |
| `geist-mono-variable.woff2` | Geist Mono | 400-600, normal (variável) | @fontsource-variable/geist-mono |

`font-display: swap` evita FOIT. Se algum arquivo faltar, o app cai nas fallback
stacks (system sans/serif/mono) sem quebrar o layout.
