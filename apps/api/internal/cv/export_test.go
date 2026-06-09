package cv

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func sampleCV() *Profile {
	return &Profile{
		Nome: "Arthur Viegas", Titulo: "Full-Stack Engineer", TituloAlvo: "Staff Engineer",
		Contato: "arthur@example.com · Porto Alegre · github.com/x",
		Resumo:  "Engenheiro de software com foco em IA e back-ends cloud-native.",
		Skills:  []string{"Go", "React", "PostgreSQL", "Docker"},
		Experiences: []Experience{
			{Empresa: "Acme", Cargo: "Backend", Inicio: "2022", Fim: "atual", Descricao: "Go e PostgreSQL.\nSegunda linha."},
		},
		Education: []Education{{Instituicao: "PUCRS", Curso: "Eng. de Software", Inicio: "2021", Fim: "2025"}},
	}
}

func TestRenderPDF(t *testing.T) {
	data, err := RenderPDF(sampleCV())
	require.NoError(t, err)
	require.NotEmpty(t, data)
	require.True(t, bytes.HasPrefix(data, []byte("%PDF")), "deve começar com %PDF")
}

func TestRenderDOCX(t *testing.T) {
	data, err := RenderDOCX(sampleCV())
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(data, []byte("PK")), "docx é um zip (assinatura PK)")

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	require.True(t, names["word/document.xml"])
	require.True(t, names["[Content_Types].xml"])
	require.True(t, names["_rels/.rels"])
}

func TestRenderEmptyCV(t *testing.T) {
	// An empty CV still produces valid documents (no panic).
	pdf, err := RenderPDF(&Profile{})
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(pdf, []byte("%PDF")))
	docx, err := RenderDOCX(&Profile{})
	require.NoError(t, err)
	require.True(t, bytes.HasPrefix(docx, []byte("PK")))
}

func TestCVFilename(t *testing.T) {
	require.Equal(t, "CV-Arthur-Viegas.pdf", cvFilename(&Profile{Nome: "Arthur Viegas"}, "pdf"))
	require.Equal(t, "CV.docx", cvFilename(&Profile{}, "docx"))
}
