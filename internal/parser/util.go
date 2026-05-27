package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"

	"endpoint-parser/internal/model"
)

// parseWithTreeSitter handles the tree-sitter parser lifecycle and delegates extraction.
// Each language parser uses this to avoid repeating setup/teardown boilerplate.
func parseWithTreeSitter(
	lang *sitter.Language,
	content []byte,
	extract func(*sitter.Node, []byte) []model.Endpoint,
) []model.Endpoint {
	p := sitter.NewParser()
	p.SetLanguage(lang)
	tree := p.Parse(nil, content)
	defer tree.Close()
	return extract(tree.RootNode(), content)
}

// walkTree visits every node in the AST via pre-order traversal.
func walkTree(node *sitter.Node, visit func(*sitter.Node)) {
	visit(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		walkTree(node.Child(i), visit)
	}
}

// nodeText returns the source text for a given AST node.
func nodeText(node *sitter.Node, content []byte) string {
	return string(content[node.StartByte():node.EndByte()])
}

// unquote removes surrounding quotes from a string literal (single, double, or backtick).
func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	c := s[0]
	if (c == '"' || c == '\'' || c == '`') && s[len(s)-1] == c {
		return s[1 : len(s)-1]
	}
	return s
}

// deduplicateEndpoints removes duplicate (method, path) pairs while preserving order.
func deduplicateEndpoints(endpoints []model.Endpoint) []model.Endpoint {
	seen := make(map[string]bool, len(endpoints))
	result := make([]model.Endpoint, 0, len(endpoints))
	for _, e := range endpoints {
		key := e.Method + ":" + e.Path
		if !seen[key] {
			seen[key] = true
			result = append(result, e)
		}
	}
	return result
}

// joinPaths concatenates a prefix and a relative path, ensuring a single leading slash.
// joinPaths("/api/v1", "users") → "/api/v1/users"
// joinPaths("/api/v1", "")     → "/api/v1"
// joinPaths("", "/users")      → "/users"
func joinPaths(base, path string) string {
	if base != "" && !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	base = strings.TrimRight(base, "/")
	if path == "" {
		return base
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if base == "" {
		return path
	}
	return base + path
}
