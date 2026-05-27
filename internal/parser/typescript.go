package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	tsLang "github.com/smacker/go-tree-sitter/typescript/typescript"

	"endpoint-parser/internal/model"
)

var typescriptLang = tsLang.GetLanguage()

var nestDecorators = map[string]string{
	"Get":    "GET",
	"Post":   "POST",
	"Put":    "PUT",
	"Delete": "DELETE",
	"Patch":  "PATCH",
}

func parseTypeScript(content []byte) []model.Endpoint {
	p := sitter.NewParser()
	p.SetLanguage(typescriptLang)
	tree := p.Parse(nil, content)
	defer tree.Close()

	root := tree.RootNode()
	var endpoints []model.Endpoint

	// NestJS: walk class declarations
	walkTree(root, func(node *sitter.Node) {
		if node.Type() == "class_declaration" {
			endpoints = append(endpoints, nestClassEndpoints(node, content)...)
		}
	})

	// Express-style TS routes (fallback)
	endpoints = append(endpoints, extractExpressEndpoints(root, content)...)

	return deduplicateEndpoints(endpoints)
}

func nestClassEndpoints(classNode *sitter.Node, content []byte) []model.Endpoint {
	basePath := ""

	// Find @Controller on the class
	for i := 0; i < int(classNode.ChildCount()); i++ {
		child := classNode.Child(i)
		if child.Type() == "decorator" {
			if path, ok := decoratorArg(child, "Controller", content); ok {
				basePath = path
			}
		}
	}

	// Find class body
	var classBody *sitter.Node
	for i := 0; i < int(classNode.ChildCount()); i++ {
		if classNode.Child(i).Type() == "class_body" {
			classBody = classNode.Child(i)
			break
		}
	}
	if classBody == nil {
		return nil
	}

	var endpoints []model.Endpoint

	for i := 0; i < int(classBody.ChildCount()); i++ {
		member := classBody.Child(i)
		if member.Type() != "method_definition" {
			continue
		}
		for j := 0; j < int(member.ChildCount()); j++ {
			child := member.Child(j)
			if child.Type() != "decorator" {
				continue
			}
			for decoratorName, httpMethod := range nestDecorators {
				if path, ok := decoratorArg(child, decoratorName, content); ok {
					endpoints = append(endpoints, model.Endpoint{
						Method: httpMethod,
						Path:   joinPaths(basePath, path),
					})
				}
			}
		}
	}
	return endpoints
}

func decoratorArg(decoratorNode *sitter.Node, name string, content []byte) (string, bool) {
	for i := 0; i < int(decoratorNode.ChildCount()); i++ {
		child := decoratorNode.Child(i)
		if child.Type() != "call_expression" {
			continue
		}
		funcNode := child.ChildByFieldName("function")
		if funcNode == nil || nodeText(funcNode, content) != name {
			continue
		}
		argsNode := child.ChildByFieldName("arguments")
		if argsNode == nil {
			return "", true
		}
		for j := 0; j < int(argsNode.ChildCount()); j++ {
			arg := argsNode.Child(j)
			if arg.Type() == "string" {
				return unquote(nodeText(arg, content)), true
			}
		}
		return "", true
	}
	return "", false
}

func joinPaths(base, path string) string {
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

func deduplicateEndpoints(endpoints []model.Endpoint) []model.Endpoint {
	seen := make(map[string]bool)
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
