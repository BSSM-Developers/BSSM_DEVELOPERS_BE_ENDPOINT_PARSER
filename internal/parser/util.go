package parser

import sitter "github.com/smacker/go-tree-sitter"

func walkTree(node *sitter.Node, visit func(*sitter.Node)) {
	visit(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		walkTree(node.Child(i), visit)
	}
}

func nodeText(node *sitter.Node, content []byte) string {
	return string(content[node.StartByte():node.EndByte()])
}

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
