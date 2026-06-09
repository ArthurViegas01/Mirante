package cv

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"
)

// RenderPDF renders the master CV to a PDF (A4, core Helvetica via fpdf — no CGO,
// no embedded fonts; accents are mapped to the cp1252 codepage).
func RenderPDF(p *Profile) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 16, 18)
	pdf.SetAutoPageBreak(true, 16)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	name := p.Nome
	if name == "" {
		name = "Currículo"
	}
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetTextColor(20, 30, 40)
	pdf.MultiCell(0, 9, tr(name), "", "L", false)

	if line := joinTitulo(p); line != "" {
		pdf.SetFont("Helvetica", "", 12)
		pdf.SetTextColor(90, 100, 110)
		pdf.MultiCell(0, 6, tr(line), "", "L", false)
	}
	if p.Contato != "" {
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(110, 120, 130)
		pdf.MultiCell(0, 5, tr(p.Contato), "", "L", false)
	}

	section := func(title string) {
		pdf.Ln(3)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(20, 120, 110)
		pdf.MultiCell(0, 6, tr(strings.ToUpper(title)), "", "L", false)
		y := pdf.GetY()
		pdf.SetDrawColor(200, 210, 215)
		pdf.Line(18, y, 192, y)
		pdf.Ln(1.5)
	}
	body := func(s string, size float64) {
		pdf.SetFont("Helvetica", "", size)
		pdf.SetTextColor(40, 50, 60)
		pdf.MultiCell(0, 5, tr(s), "", "L", false)
	}
	entry := func(head, period, desc string) {
		pdf.SetFont("Helvetica", "B", 10.5)
		pdf.SetTextColor(20, 30, 40)
		pdf.MultiCell(0, 5.5, tr(head), "", "L", false)
		if period != "" {
			pdf.SetFont("Helvetica", "I", 9)
			pdf.SetTextColor(110, 120, 130)
			pdf.MultiCell(0, 5, tr(period), "", "L", false)
		}
		if desc != "" {
			body(desc, 9.5)
		}
		pdf.Ln(1.5)
	}

	if p.Resumo != "" {
		section("Resumo")
		body(p.Resumo, 10)
	}
	if len(p.Skills) > 0 {
		section("Skills")
		body(strings.Join(p.Skills, " · "), 10)
	}
	if len(p.Experiences) > 0 {
		section("Experiência")
		for _, e := range p.Experiences {
			entry(headOf(e.Cargo, e.Empresa), periodOf(e.Inicio, e.Fim), e.Descricao)
		}
	}
	if len(p.Education) > 0 {
		section("Educação")
		for _, e := range p.Education {
			entry(headOf(e.Curso, e.Instituicao), periodOf(e.Inicio, e.Fim), "")
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- DOCX (minimal OOXML written by hand into a zip) ---

const (
	docxDecl = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`

	docxContentTypes = docxDecl + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
		`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
		`<Default Extension="xml" ContentType="application/xml"/>` +
		`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>` +
		`</Types>`

	docxRels = docxDecl + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
		`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>` +
		`</Relationships>`
)

var xmlEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")

// RenderDOCX renders the master CV to a .docx (OOXML) byte slice.
func RenderDOCX(p *Profile) ([]byte, error) {
	var b strings.Builder
	para := func(text string, halfPt int, bold, italic bool, color string) {
		if strings.TrimSpace(text) == "" {
			return
		}
		b.WriteString(`<w:p><w:pPr><w:spacing w:after="80"/></w:pPr><w:r><w:rPr>`)
		if bold {
			b.WriteString(`<w:b/>`)
		}
		if italic {
			b.WriteString(`<w:i/>`)
		}
		if color != "" {
			b.WriteString(`<w:color w:val="` + color + `"/>`)
		}
		fmt.Fprintf(&b, `<w:sz w:val="%d"/>`, halfPt)
		b.WriteString(`</w:rPr>`)
		for i, part := range strings.Split(text, "\n") {
			if i > 0 {
				b.WriteString(`<w:br/>`)
			}
			b.WriteString(`<w:t xml:space="preserve">` + xmlEscaper.Replace(part) + `</w:t>`)
		}
		b.WriteString(`</w:r></w:p>`)
	}
	heading := func(t string) { para(strings.ToUpper(t), 22, true, false, "147A6E") }

	name := p.Nome
	if name == "" {
		name = "Currículo"
	}
	para(name, 40, true, false, "1E2A33")
	para(joinTitulo(p), 24, false, false, "5A6470")
	para(p.Contato, 18, false, false, "6E7884")

	if p.Resumo != "" {
		heading("Resumo")
		para(p.Resumo, 20, false, false, "")
	}
	if len(p.Skills) > 0 {
		heading("Skills")
		para(strings.Join(p.Skills, " · "), 20, false, false, "")
	}
	if len(p.Experiences) > 0 {
		heading("Experiência")
		for _, e := range p.Experiences {
			para(headOf(e.Cargo, e.Empresa), 21, true, false, "1E2A33")
			para(periodOf(e.Inicio, e.Fim), 18, false, true, "6E7884")
			para(e.Descricao, 19, false, false, "")
		}
	}
	if len(p.Education) > 0 {
		heading("Educação")
		for _, e := range p.Education {
			para(headOf(e.Curso, e.Instituicao), 21, true, false, "1E2A33")
			para(periodOf(e.Inicio, e.Fim), 18, false, true, "6E7884")
		}
	}

	doc := docxDecl +
		`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>` +
		b.String() + `<w:sectPr/></w:body></w:document>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range []struct{ name, content string }{
		{"[Content_Types].xml", docxContentTypes},
		{"_rels/.rels", docxRels},
		{"word/document.xml", doc},
	} {
		w, err := zw.Create(f.name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(f.content)); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// --- shared helpers ---

func joinTitulo(p *Profile) string {
	line := p.Titulo
	if p.TituloAlvo != "" {
		if line != "" {
			line += "  ·  "
		}
		line += "Objetivo: " + p.TituloAlvo
	}
	return line
}

func headOf(primary, secondary string) string {
	if secondary == "" {
		return primary
	}
	if primary == "" {
		return secondary
	}
	return primary + " — " + secondary
}

func periodOf(inicio, fim string) string {
	inicio, fim = strings.TrimSpace(inicio), strings.TrimSpace(fim)
	switch {
	case inicio != "" && fim != "":
		return inicio + " – " + fim
	case inicio != "":
		return inicio
	default:
		return fim
	}
}

// cvFilename builds a download filename like "CV-Arthur-Viegas.pdf".
func cvFilename(p *Profile, ext string) string {
	base := strings.TrimSpace(p.Nome)
	if base == "" {
		return "CV." + ext
	}
	slug := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		case r == ' ':
			return '-'
		default:
			return -1
		}
	}, base)
	if slug == "" {
		slug = "CV"
	}
	return "CV-" + slug + "." + ext
}
