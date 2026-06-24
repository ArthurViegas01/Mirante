// Package skills is the shared kernel for the career-search domains (jobs, cv).
// It owns a canonical vocabulary of technical skills with synonyms/aliases and a
// light category ontology, all as in-code Go data — no database, no HTTP, no
// external dependencies. Per ADR-0001 it is a conscious shared kernel: domains
// may import it (they do not import each other). Its API is deterministic:
// Normalize(raw) maps a token to its canonical skill; Match(text) extracts the
// canonical skills mentioned in free text (e.g. a job posting). This determinism
// is the floor under any later LLM-based aderência scoring (ADR-0004).
package skills

import "sort"

// Category is the broad bucket a skill belongs to.
type Category string

const (
	CatLinguagem  Category = "linguagem"
	CatFramework  Category = "framework"
	CatBanco      Category = "banco"
	CatInfra      Category = "infra"
	CatCloud      Category = "cloud"
	CatPlataforma Category = "plataforma"
	CatFerramenta Category = "ferramenta"
	CatConceito   Category = "conceito"
)

// Skill is a canonical entry in the vocabulary. Canonical is the display form;
// Aliases are alternative spellings/synonyms (case-insensitive) that Normalize
// and Match resolve back to Canonical. Related lists adjacent canonical skills
// (the ontology edge), e.g. SvelteKit → Svelte.
type Skill struct {
	Canonical string
	Category  Category
	Aliases   []string
	Related   []string
}

// catalog is the source of truth. Order here is the order All() and Match()
// return, so keep related skills grouped for readable output.
var catalog = []Skill{
	// --- Linguagens ---
	{Canonical: "Go", Category: CatLinguagem, Aliases: []string{"golang"}},
	{Canonical: "JavaScript", Category: CatLinguagem, Aliases: []string{"js", "ecmascript"}},
	{Canonical: "TypeScript", Category: CatLinguagem, Aliases: []string{"ts"}, Related: []string{"JavaScript"}},
	{Canonical: "Python", Category: CatLinguagem, Aliases: []string{"py"}},
	{Canonical: "Java", Category: CatLinguagem},
	{Canonical: "C#", Category: CatLinguagem, Aliases: []string{"csharp", "c-sharp"}, Related: []string{".NET"}},
	{Canonical: "C++", Category: CatLinguagem, Aliases: []string{"cpp", "cplusplus"}},
	{Canonical: "Ruby", Category: CatLinguagem},
	{Canonical: "PHP", Category: CatLinguagem},
	{Canonical: "Rust", Category: CatLinguagem},
	{Canonical: "Kotlin", Category: CatLinguagem},
	{Canonical: "Swift", Category: CatLinguagem},
	{Canonical: "Dart", Category: CatLinguagem},
	{Canonical: "SQL", Category: CatLinguagem},
	{Canonical: "HTML", Category: CatLinguagem, Aliases: []string{"html5"}},
	{Canonical: "CSS", Category: CatLinguagem, Aliases: []string{"css3"}},
	{Canonical: "Shell", Category: CatLinguagem, Aliases: []string{"bash", "shell script", "shellscript"}},

	// --- Frameworks / bibliotecas ---
	{Canonical: "React", Category: CatFramework, Aliases: []string{"react.js", "reactjs"}, Related: []string{"JavaScript"}},
	{Canonical: "React Native", Category: CatFramework, Aliases: []string{"react-native"}, Related: []string{"React"}},
	{Canonical: "Flutter", Category: CatFramework, Related: []string{"Dart"}},
	{Canonical: "Vue", Category: CatFramework, Aliases: []string{"vue.js", "vuejs"}, Related: []string{"JavaScript"}},
	{Canonical: "Angular", Category: CatFramework, Aliases: []string{"angular.js", "angularjs"}, Related: []string{"TypeScript"}},
	{Canonical: "Svelte", Category: CatFramework, Related: []string{"JavaScript"}},
	{Canonical: "SvelteKit", Category: CatFramework, Aliases: []string{"svelte kit"}, Related: []string{"Svelte"}},
	{Canonical: "Next.js", Category: CatFramework, Aliases: []string{"nextjs", "next js"}, Related: []string{"React"}},
	{Canonical: "Node.js", Category: CatFramework, Aliases: []string{"nodejs", "node js", "node"}, Related: []string{"JavaScript"}},
	{Canonical: "Express", Category: CatFramework, Aliases: []string{"express.js", "expressjs"}, Related: []string{"Node.js"}},
	{Canonical: "Django", Category: CatFramework, Related: []string{"Python"}},
	{Canonical: "Flask", Category: CatFramework, Related: []string{"Python"}},
	{Canonical: "FastAPI", Category: CatFramework, Aliases: []string{"fast api"}, Related: []string{"Python"}},
	{Canonical: "Spring", Category: CatFramework, Aliases: []string{"spring boot", "springboot"}, Related: []string{"Java"}},
	{Canonical: "Rails", Category: CatFramework, Aliases: []string{"ruby on rails", "ror"}, Related: []string{"Ruby"}},
	{Canonical: "Laravel", Category: CatFramework, Related: []string{"PHP"}},
	{Canonical: ".NET", Category: CatFramework, Aliases: []string{"dotnet", "asp.net", ".net core"}, Related: []string{"C#"}},
	{Canonical: "Tailwind CSS", Category: CatFramework, Aliases: []string{"tailwind", "tailwindcss"}, Related: []string{"CSS"}},
	{Canonical: "GSAP", Category: CatFramework},

	// --- Plataformas / CMS / e-commerce / no-code ---
	{Canonical: "WordPress", Category: CatPlataforma, Aliases: []string{"wp"}, Related: []string{"PHP"}},
	{Canonical: "WooCommerce", Category: CatPlataforma, Aliases: []string{"woo commerce"}, Related: []string{"WordPress"}},
	{Canonical: "Elementor", Category: CatPlataforma, Related: []string{"WordPress"}},
	{Canonical: "Shopify", Category: CatPlataforma},
	{Canonical: "Wix", Category: CatPlataforma},
	{Canonical: "Webflow", Category: CatPlataforma},
	{Canonical: "Squarespace", Category: CatPlataforma},
	{Canonical: "Bubble", Category: CatPlataforma},

	// --- Bancos de dados ---
	{Canonical: "PostgreSQL", Category: CatBanco, Aliases: []string{"postgres", "psql"}, Related: []string{"SQL"}},
	{Canonical: "MySQL", Category: CatBanco, Related: []string{"SQL"}},
	{Canonical: "SQLite", Category: CatBanco, Related: []string{"SQL"}},
	{Canonical: "libSQL", Category: CatBanco, Aliases: []string{"turso"}, Related: []string{"SQLite"}},
	{Canonical: "MongoDB", Category: CatBanco, Aliases: []string{"mongo"}},
	{Canonical: "Redis", Category: CatBanco},
	{Canonical: "Elasticsearch", Category: CatBanco, Aliases: []string{"elastic search"}},

	// --- Infra / DevOps ---
	{Canonical: "Docker", Category: CatInfra},
	{Canonical: "Kubernetes", Category: CatInfra, Aliases: []string{"k8s"}, Related: []string{"Docker"}},
	{Canonical: "Terraform", Category: CatInfra},
	{Canonical: "CI/CD", Category: CatInfra, Aliases: []string{"cicd", "ci cd"}},
	{Canonical: "GitHub Actions", Category: CatInfra, Related: []string{"CI/CD"}},
	{Canonical: "Linux", Category: CatInfra},
	{Canonical: "Nginx", Category: CatInfra},

	// --- Cloud / hospedagem ---
	{Canonical: "AWS", Category: CatCloud, Aliases: []string{"amazon web services"}},
	{Canonical: "GCP", Category: CatCloud, Aliases: []string{"google cloud", "google cloud platform"}},
	{Canonical: "Azure", Category: CatCloud, Aliases: []string{"microsoft azure"}},
	{Canonical: "Fly.io", Category: CatCloud, Aliases: []string{"fly io"}},
	{Canonical: "Vercel", Category: CatCloud},
	{Canonical: "Netlify", Category: CatCloud},
	{Canonical: "Cloudflare", Category: CatCloud},
	{Canonical: "Supabase", Category: CatCloud, Related: []string{"PostgreSQL"}},
	{Canonical: "Railway", Category: CatCloud},

	// --- Ferramentas / protocolos ---
	{Canonical: "Git", Category: CatFerramenta},
	{Canonical: "REST", Category: CatFerramenta, Aliases: []string{"rest api", "restful"}},
	{Canonical: "GraphQL", Category: CatFerramenta},
	{Canonical: "gRPC", Category: CatFerramenta},
	{Canonical: "OpenTelemetry", Category: CatFerramenta, Aliases: []string{"otel"}},
	{Canonical: "Figma", Category: CatFerramenta},
	{Canonical: "Pentaho", Category: CatFerramenta, Aliases: []string{"pentaho pdi", "pdi/kettle", "kettle", "pdi"}},
	{Canonical: "Power BI", Category: CatFerramenta, Aliases: []string{"powerbi"}},
	{Canonical: "Tableau", Category: CatFerramenta},
	{Canonical: "n8n", Category: CatFerramenta},
	{Canonical: "Make.com", Category: CatFerramenta, Aliases: []string{"integromat"}},
	{Canonical: "Zapier", Category: CatFerramenta},
	{Canonical: "Twilio", Category: CatFerramenta},

	// --- Conceitos / práticas ---
	{Canonical: "Microserviços", Category: CatConceito, Aliases: []string{"microservices", "microsserviços"}},
	{Canonical: "TDD", Category: CatConceito, Aliases: []string{"test driven development", "test-driven development"}},
	{Canonical: "Scrum", Category: CatConceito, Aliases: []string{"agile", "ágil"}},
}

// All returns the full catalog in canonical order.
func All() []Skill {
	out := make([]Skill, len(catalog))
	copy(out, catalog)
	return out
}

// Get returns the Skill for a canonical name (exact, case-sensitive on canonical).
func Get(canonical string) (Skill, bool) {
	s, ok := byCanonical[canonical]
	return s, ok
}

// ByCategory groups the catalog by category, each slice in canonical order.
func ByCategory() map[Category][]Skill {
	out := map[Category][]Skill{}
	for _, s := range catalog {
		out[s.Category] = append(out[s.Category], s)
	}
	return out
}

// Categories returns the category keys present, in a stable order.
func Categories() []Category {
	seen := map[Category]bool{}
	var out []Category
	for _, s := range catalog {
		if !seen[s.Category] {
			seen[s.Category] = true
			out = append(out, s.Category)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
