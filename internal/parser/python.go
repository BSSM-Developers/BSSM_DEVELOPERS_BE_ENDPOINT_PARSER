package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"

	"endpoint-parser/internal/model"
)

type pythonParser struct{}

func (pythonParser) Language() model.Language { return model.LangPython }

func (pythonParser) Parse(content []byte) []model.Endpoint {
	return parseWithTreeSitter(python.GetLanguage(), content, extractFastAPIEndpoints)
}

// extractFastAPIEndpoints collects @app.get / @router.post / @app.route decorators.
func extractFastAPIEndpoints(root *sitter.Node, content []byte) []model.Endpoint {
	var endpoints []model.Endpoint
	walkTree(root, func(node *sitter.Node) {
		if node.Type() != "decorator" {
			return
		}
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "call" {
				endpoints = append(endpoints, extractRouteFromCall(child, content)...)
			}
		}
	})
	return endpoints
}

func extractRouteFromCall(call *sitter.Node, content []byte) []model.Endpoint {
	funcNode := call.ChildByFieldName("function")
	argsNode := call.ChildByFieldName("arguments")
	if funcNode == nil || argsNode == nil {
		return nil
	}

	methodName := extractCallMethodName(funcNode, content)
	if methodName == "" {
		return nil
	}

	// Flask / Werkzeug: @app.route("/path", methods=["GET", "POST"])
	if strings.ToLower(methodName) == "route" {
		return extractFlaskRoute(argsNode, content)
	}

	httpMethod := toHTTPMethod(methodName)
	if httpMethod == "" {
		return nil
	}

	path := extractRoutePathFromArgs(argsNode, content)
	if path == "" {
		return nil
	}
	return []model.Endpoint{{Method: httpMethod, Path: path}}
}

// extractCallMethodName returns the method name from an attribute call (router.get → "get")
// or a bare identifier call (get → "get").
func extractCallMethodName(funcNode *sitter.Node, content []byte) string {
	switch funcNode.Type() {
	case "attribute":
		attr := funcNode.ChildByFieldName("attribute")
		if attr != nil {
			return nodeText(attr, content)
		}
	case "identifier":
		return nodeText(funcNode, content)
	}
	return ""
}

// extractRoutePathFromArgs finds the route path: positional string first, then path= kwarg.
func extractRoutePathFromArgs(argsNode *sitter.Node, content []byte) string {
	var kwargPath string
	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		switch child.Type() {
		case "string":
			text := nodeText(child, content)
			if !isPythonStringLiteral(text) {
				continue
			}
			return unquote(text)
		case "keyword_argument":
			if path := extractPathKwarg(child, content); path != "" {
				kwargPath = path
			}
		}
	}
	return kwargPath
}

// extractPathKwarg reads the value from a `path="..."` keyword argument node.
func extractPathKwarg(kwarg *sitter.Node, content []byte) string {
	nameNode := kwarg.ChildByFieldName("name")
	valNode := kwarg.ChildByFieldName("value")
	if nameNode == nil || valNode == nil {
		return ""
	}
	if nodeText(nameNode, content) == "path" && valNode.Type() == "string" {
		text := nodeText(valNode, content)
		if !isPythonStringLiteral(text) {
			return ""
		}
		return unquote(text)
	}
	return ""
}

// isPythonStringLiteral returns false for f-strings (f"..." or f'...').
func isPythonStringLiteral(text string) bool {
	if len(text) < 2 {
		return false
	}
	return !(text[0] == 'f' && (text[1] == '"' || text[1] == '\''))
}

// extractFlaskRoute handles @app.route("/path", methods=["GET", "POST"]).
// Returns one Endpoint per HTTP method in the methods list.
func extractFlaskRoute(argsNode *sitter.Node, content []byte) []model.Endpoint {
	path := ""
	var methods []string

	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		switch child.Type() {
		case "string":
			if path == "" {
				text := nodeText(child, content)
				if isPythonStringLiteral(text) {
					path = unquote(text)
				}
			}
		case "keyword_argument":
			nameNode := child.ChildByFieldName("name")
			valNode := child.ChildByFieldName("value")
			if nameNode == nil || valNode == nil {
				continue
			}
			switch nodeText(nameNode, content) {
			case "path":
				if path == "" && valNode.Type() == "string" {
					path = unquote(nodeText(valNode, content))
				}
			case "methods":
				methods = extractFlaskMethods(valNode, content)
			}
		}
	}

	if path == "" {
		return nil
	}
	if len(methods) == 0 {
		methods = []string{"GET"}
	}

	endpoints := make([]model.Endpoint, 0, len(methods))
	for _, m := range methods {
		endpoints = append(endpoints, model.Endpoint{Method: m, Path: path})
	}
	return endpoints
}

// extractFlaskMethods extracts HTTP verb strings from a list literal: ["GET", "POST"].
func extractFlaskMethods(node *sitter.Node, content []byte) []string {
	if node == nil || node.Type() != "list" {
		return nil
	}
	var methods []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "string" {
			m := strings.ToUpper(unquote(nodeText(child, content)))
			if isHTTPVerb(m) {
				methods = append(methods, m)
			}
		}
	}
	return methods
}

func toHTTPMethod(name string) string {
	switch strings.ToLower(name) {
	case "get":
		return "GET"
	case "post":
		return "POST"
	case "put":
		return "PUT"
	case "delete":
		return "DELETE"
	case "patch":
		return "PATCH"
	}
	return ""
}

func isHTTPVerb(s string) bool {
	switch s {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		return true
	}
	return false
}
