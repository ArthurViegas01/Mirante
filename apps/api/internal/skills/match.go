package skills

import "strings"

// Indices derived from the catalog at startup.
var (
	byCanonical = map[string]Skill{}  // canonical -> Skill
	aliasIndex  = map[string]string{} // lowercased canonical/alias -> canonical
)

func init() {
	for _, s := range catalog {
		byCanonical[s.Canonical] = s
		aliasIndex[strings.ToLower(s.Canonical)] = s.Canonical
		for _, a := range s.Aliases {
			aliasIndex[strings.ToLower(a)] = s.Canonical
		}
	}
}

// Normalize maps a raw token (a single skill mention) to its canonical form,
// returning ("", false) when it is not a known skill or alias.
func Normalize(raw string) (string, bool) {
	canon, ok := aliasIndex[strings.ToLower(strings.TrimSpace(raw))]
	return canon, ok
}

// Match extracts the canonical skills mentioned anywhere in free text (e.g. a job
// description). Matching is case-insensitive and respects token boundaries, so
// "go" matches the Go language but not the "go" inside "categoria". Each skill is
// returned at most once, in catalog order.
func Match(text string) []string {
	hay := strings.ToLower(text)
	out := []string{}
	for _, s := range catalog {
		if mentioned(hay, s) {
			out = append(out, s.Canonical)
		}
	}
	return out
}

func mentioned(hay string, s Skill) bool {
	if containsToken(hay, strings.ToLower(s.Canonical)) {
		return true
	}
	for _, a := range s.Aliases {
		if containsToken(hay, strings.ToLower(a)) {
			return true
		}
	}
	return false
}

// containsToken reports whether needle occurs in hay bounded by non-word
// characters on both sides (so it is not part of a larger word). Both arguments
// are assumed already lowercased.
func containsToken(hay, needle string) bool {
	if needle == "" {
		return false
	}
	for from := 0; from+len(needle) <= len(hay); {
		i := strings.Index(hay[from:], needle)
		if i < 0 {
			return false
		}
		start := from + i
		end := start + len(needle)
		beforeOK := start == 0 || !isWordByte(hay[start-1])
		afterOK := end == len(hay) || !isWordByte(hay[end])
		if beforeOK && afterOK {
			return true
		}
		from = start + 1
	}
	return false
}

func isWordByte(b byte) bool {
	return b == '_' ||
		(b >= '0' && b <= '9') ||
		(b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z')
}
