package intake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// digestFixture is the text rendering of a real 99Freelas "Novos projetos"
// notification e-mail (11 projects), kept verbatim so the parser is exercised
// against the exact shape it must handle in production.
const digestFixture = `Olá, Arthur Viegas.

Há novos projetos que possam ser do seu interesse:

Web, Mobile & Software

Loja Shopify para mercado dos Estados Unidos

Desenvolvimento Web | Intermediário | Publicado: ontem às 02:04 | Tempo restante: 5 dias e 19 horas | Propostas: 28 | Interessados: 32

Desenvolvimento de loja Shopify para o mercado dos Estados Unidos (EUA). Descrição: Procuro profissional com experiência em Shopify para desenvolver uma lo... Leia mais.

Ver projeto ou Enviar proposta

Sistema inteligente para gestão clínica e neuroreabilitação baseado em inteligência artificial

Criação & Integração com IA | Especialista | Publicado: ontem às 09:46 | Tempo restante: 2 dias e 3 horas | Propostas: 41 | Interessados: 44

Desenvolver e implantar a plataforma FISIO IA CARE, um sistema inteligente para gestão clínica e neuroreabilitação baseado em Inteligência Artificial. Além ... Leia mais.

Ver projeto ou Enviar proposta

Sistema de gestão para oficina mecânica com orçamento via WhatsApp

Desenvolvimento Web | Intermediário | Publicado: ontem às 13:12 | Tempo restante: 29 dias e 6 horas | Propostas: 74 | Interessados: 99

Preciso de um sistema web para gerenciar ordens de serviço de uma oficina mecânica. O diferencial principal é a geração automática de orçamentos em PDF e o e... Leia mais.

Ver projeto ou Enviar proposta

Plataforma web Eco BioLarv para monitoramento do Aedes aegypti

Desenvolvimento Web | Iniciante | Publicado: ontem às 13:40 | Tempo restante: 29 dias e 7 horas | Propostas: 41 | Interessados: 49

Sobre o Projeto O Eco BioLarv é uma solução inteligente, científica e educacional voltada ao monitoramento ambiental e ao combate ao mosquito Aedes aegypti, ... Leia mais.

Ver projeto ou Enviar proposta

Especialista em Pentaho PDI/Kettle - demandas por hora

Banco de Dados | Intermediário | Publicado: ontem às 14:11 | Tempo restante: 29 dias e 7 horas | Propostas: 9 | Interessados: 15

A FLM Tecnologia busca profissional PJ/freelancer com experiência em Pentaho Data Integration (PDI/Kettle) para atendimento sob demanda em projetos de ETL e ... Leia mais.

Habilidades desejadas: Pentaho

Ver projeto ou Enviar proposta

Criar página no Beacons com produtos e pagamento integrado

Desenvolvimento Web | Iniciante | Publicado: ontem às 18:37 | Tempo restante: 29 dias e 12 horas | Propostas: 14 | Interessados: 19

Procuro alguém para criar uma página no Beacons, completa, com meus produtos e pagamento integrado. A página deve ser otimizada para ser usada como link na b... Leia mais.

Ver projeto ou Enviar proposta

Criação de design para área de membros do Curseduca

UX/UI & Web Design | Intermediário | Publicado: ontem às 19:15 | Tempo restante: 2 dias e 12 horas | Propostas: 16 | Interessados: 17

Procuro um designer para criar o design da área de membros do Curseduca. O que preciso que seja criado 1. Dashboard Principal Criação do layout visual da ... Leia mais.

Ver projeto ou Enviar proposta

Painel de arbitragem de futebol ao vivo

Desenvolvimento Web | Intermediário | Publicado: ontem às 20:19 | Tempo restante: 14 dias e 13 horas | Propostas: 25 | Interessados: 32

Preciso de um painel de arbitragem totalmente focado em partidas de futebol. O objetivo central é exibir e atualizar a pontuação em tempo real, de forma está... Leia mais.

Ver projeto ou Enviar proposta

Design de UI/UX para site de software house

UX/UI & Web Design | Intermediário | Publicado: ontem às 21:30 | Tempo restante: 29 dias e 14 horas | Propostas: 16 | Interessados: 22

Preciso criar a UI/UX do site da minha empresa. Quero um projeto que fuja dos padrões de software houses já presentes no mercado, com traços finos, elegantes... Leia mais.

Ver projeto ou Enviar proposta

Desenvolvimento de website enciclopédia sobre cantora

Desenvolvimento Web | Intermediário | Publicado: ontem às 22:05 | Tempo restante: 6 dias e 15 horas | Propostas: 28 | Interessados: 34

Criação e desenvolvimento de um website completo para um projeto expansivo de enciclopédia sobre uma cantora de alcance global. O site deve ser baseado nos r... Leia mais.

Habilidades desejadas: Wordpress

Ver projeto ou Enviar proposta

Sistema web de registro e gestão de notificações de risco

Desenvolvimento Web | Iniciante | Publicado: hoje às 03:06 | Tempo restante: 29 dias e 20 horas | Propostas: 3 | Interessados: 7

Desenvolvimento de sistema web completo para registro, acompanhamento e gestão de notificações de risco. Funcionalidades: - Cadastro de notificações com upl... Leia mais.

Ver projeto ou Enviar proposta`

func TestParseDigest(t *testing.T) {
	got := ParseDigest(digestFixture)

	// 11 projects — the greeting and the "Web, Mobile & Software" section header
	// have no metadata line and must not be parsed as projects.
	require.Len(t, got, 11)

	// Order, titles and the competition signal across every project.
	want := []struct {
		titulo    string
		propostas int
	}{
		{"Loja Shopify para mercado dos Estados Unidos", 28},
		{"Sistema inteligente para gestão clínica e neuroreabilitação baseado em inteligência artificial", 41},
		{"Sistema de gestão para oficina mecânica com orçamento via WhatsApp", 74},
		{"Plataforma web Eco BioLarv para monitoramento do Aedes aegypti", 41},
		{"Especialista em Pentaho PDI/Kettle - demandas por hora", 9},
		{"Criar página no Beacons com produtos e pagamento integrado", 14},
		{"Criação de design para área de membros do Curseduca", 16},
		{"Painel de arbitragem de futebol ao vivo", 25},
		{"Design de UI/UX para site de software house", 16},
		{"Desenvolvimento de website enciclopédia sobre cantora", 28},
		{"Sistema web de registro e gestão de notificações de risco", 3},
	}
	for i, w := range want {
		require.Equal(t, w.titulo, got[i].Titulo, "titulo[%d]", i)
		require.Equal(t, w.propostas, got[i].Propostas, "propostas[%d]", i)
	}
}

func TestParseDigestFields(t *testing.T) {
	got := ParseDigest(digestFixture)
	require.Len(t, got, 11)

	// Full field extraction on the first project.
	shopify := got[0]
	require.Equal(t, "Desenvolvimento Web", shopify.Categoria)
	require.Equal(t, "Intermediário", shopify.Nivel)
	require.Equal(t, "ontem às 02:04", shopify.Publicado)
	require.Equal(t, "5 dias e 19 horas", shopify.TempoRestante)
	require.Equal(t, 5*24+19, shopify.RestanteHoras)
	require.Equal(t, 32, shopify.Interessados)
	require.Contains(t, shopify.Teaser, "loja Shopify para o mercado dos Estados Unidos")
	require.NotContains(t, shopify.Teaser, "Leia mais")
	require.Empty(t, shopify.Skills)

	// Explicit "Habilidades desejadas" is captured when present.
	require.Equal(t, []string{"Pentaho"}, got[4].Skills)
	require.Equal(t, "Banco de Dados", got[4].Categoria)
	require.Equal(t, []string{"Wordpress"}, got[9].Skills)

	// The hot lead: posted today, only 3 proposals, beginner level.
	hot := got[10]
	require.Equal(t, "hoje às 03:06", hot.Publicado)
	require.Equal(t, "Iniciante", hot.Nivel)
	require.Equal(t, 3, hot.Propostas)
	require.Equal(t, 29*24+20, hot.RestanteHoras)
}

func TestParseDigestEmpty(t *testing.T) {
	require.Empty(t, ParseDigest(""))
	require.Empty(t, ParseDigest("Olá, Arthur.\n\nNenhuma novidade hoje."))
}
