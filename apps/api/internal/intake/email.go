package intake

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strings"
)

// Fonte99Freelas tags opportunities ingested from the 99Freelas digest feed.
const Fonte99Freelas = "99freelas"

// ParseEmail decodes a raw 99Freelas notification e-mail (RFC 822) and returns the
// projects it lists, each enriched with the source links and a stable project id
// (the trailing number in the URL) for dedup. It pairs the text-field parser
// (ParseDigest over the rendered text) with link extraction over the HTML, zipping
// the two by document order.
func ParseEmail(raw []byte) ([]ParsedProject, error) {
	htmlBody, err := decodeHTML(raw)
	if err != nil {
		return nil, err
	}
	projects := ParseDigest(htmlToText(htmlBody))
	links := projectLinks(htmlBody)

	for i := range projects {
		if i >= len(links) {
			break
		}
		projects[i].Fonte = Fonte99Freelas
		projects[i].FonteID = links[i].id
		projects[i].URL = links[i].verURL
		projects[i].EnviarURL = links[i].bidURL
	}
	return projects, nil
}

// decodeHTML extracts the decoded text/html body from a raw e-mail, honouring the
// Content-Transfer-Encoding (quoted-printable or base64). Charset is assumed UTF-8,
// which is what 99Freelas/SES send. Multipart messages are not handled yet.
func decodeHTML(raw []byte) (string, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("intake: parse e-mail: %w", err)
	}
	if mediaType, _, err := mime.ParseMediaType(msg.Header.Get("Content-Type")); err == nil &&
		strings.HasPrefix(mediaType, "multipart/") {
		return "", fmt.Errorf("intake: multipart e-mail ainda não suportado")
	}

	var body io.Reader = msg.Body
	switch strings.ToLower(strings.TrimSpace(msg.Header.Get("Content-Transfer-Encoding"))) {
	case "quoted-printable":
		body = quotedprintable.NewReader(body)
	case "base64":
		body = base64.NewDecoder(base64.StdEncoding, body)
	}
	b, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("intake: read e-mail body: %w", err)
	}
	return string(b), nil
}

// --- link extraction ---

type projectLink struct {
	id     string
	verURL string // .../project/<slug>-<id>
	bidURL string // .../project/bid/<slug>-<id>
}

var (
	projectHrefRe = regexp.MustCompile(`https://www\.99freelas\.com\.br/project/([^"]+)`)
	idTailRe      = regexp.MustCompile(`(\d+)$`) // trailing id of a slug: "slug-762750" → "762750"
)

// projectLinks returns the projects' links in document order — one entry per
// distinct project id (first appearance wins) — pairing the "Ver projeto" URL with
// the "Enviar proposta" (/project/bid/…) URL by shared id. The non-project
// /email/… links (view-in-browser, unsubscribe) never match the /project/ prefix.
func projectLinks(htmlBody string) []projectLink {
	var order []string
	byID := map[string]*projectLink{}

	for _, m := range projectHrefRe.FindAllStringSubmatch(htmlBody, -1) {
		path := m[1] // "slug-762750" or "bid/slug-762750"
		isBid := strings.HasPrefix(path, "bid/")
		slug := strings.TrimPrefix(path, "bid/")
		idm := idTailRe.FindStringSubmatch(slug)
		if idm == nil {
			continue
		}
		id := idm[1]
		full := "https://www.99freelas.com.br/project/" + path

		pl, ok := byID[id]
		if !ok {
			pl = &projectLink{id: id}
			byID[id] = pl
			order = append(order, id)
		}
		switch {
		case isBid && pl.bidURL == "":
			pl.bidURL = full
		case !isBid && pl.verURL == "":
			pl.verURL = full
		}
	}

	out := make([]projectLink, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out
}

// --- HTML → text ---

var (
	blockTagRe = regexp.MustCompile(`(?i)<\s*(br\s*/?|/p|/div|/li|/h[1-6]|/tr|/table)\s*>`)
	anyTagRe   = regexp.MustCompile(`(?s)<[^>]+>`)
	wsRe       = regexp.MustCompile(`[\s\x{00a0}]+`) // \x{00a0} = nbsp, which \s misses
)

// htmlToText renders e-mail HTML to text with ONE line per block element: block
// boundaries become newlines, while every other run of whitespace — including the
// source's own newlines inside a paragraph — collapses to a single space. That is
// what lets a pipe-separated metadata paragraph land on a single line, the shape
// ParseDigest expects.
func htmlToText(htmlBody string) string {
	const nl = "\x00" // sentinel for block boundaries, immune to the whitespace collapse
	s := blockTagRe.ReplaceAllString(htmlBody, nl)
	s = anyTagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = wsRe.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, nl, "\n")

	var lines []string
	for _, ln := range strings.Split(s, "\n") {
		if t := strings.TrimSpace(ln); t != "" {
			lines = append(lines, t)
		}
	}
	return strings.Join(lines, "\n")
}
