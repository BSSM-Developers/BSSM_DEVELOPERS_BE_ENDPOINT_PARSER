package parser

import (
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
