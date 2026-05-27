package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"

	"endpoint-parser/internal/model"
)

// jsHTTPMethods maps Express method names to HTTP verbs.
var jsHTTPMethods = map[string]string{
	"get":    "GET",
	"post":   "POST",
	"put":    "PUT",
	"delete": "DELETE",
	"patch":  "PATCH",
}

type javaScriptParser struct{}

func (javaScriptParser) Language() model.Language { return model.LangJavaScript }

func (javaScriptParser) Parse(content []byte) []model.Endpoint {
	return parseWithTreeSitter(javascript.GetLanguage(), content, extractExpressEndpoints)
}

// extractExpressEndpoints collects router.get('/path', ...) / app.post('/path', ...) calls.
// Also shared by the TypeScript parser for Express-style TS files.
func extractExpressEndpoints(root *sitter.Node, content []byte) []model.Endpoint {
	var endpoints []model.Endpoint
	walkTree(root, func(node *sitter.Node) {
		if node.Type() != "call_expression" {
			return
		}
		if method, path, ok := extractExpressRoute(node, content); ok {
			endpoints = append(endpoints, model.Endpoint{Method: method, Path: path})
		}
	})
	return endpoints
}

func extractExpressRoute(call *sitter.Node, content []byte) (method, path string, ok bool) {
	funcNode := call.ChildByFieldName("function")
	argsNode := call.ChildByFieldName("arguments")
	if funcNode == nil || argsNode == nil {
		return
	}
	if funcNode.Type() != "member_expression" {
		return
	}
	propNode := funcNode.ChildByFieldName("property")
	if propNode == nil {
		return
	}

	httpMethod, supported := jsHTTPMethods[strings.ToLower(nodeText(propNode, content))]
	if !supported {
		return
	}

	// Direct path argument: router.get('/path', handler)
	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		if child.Type() == "string" || child.Type() == "template_string" {
			return httpMethod, unquote(nodeText(child, content)), true
		}
	}

	// Chained call: router.route('/path').get(fn) — walk the object chain for route()
	objNode := funcNode.ChildByFieldName("object")
	if chainedPath := findChainedRoutePath(objNode, content); chainedPath != "" {
		return httpMethod, chainedPath, true
	}
	return
}

// findChainedRoutePath traverses a method-call chain to find the path from .route('/path').
// Handles: router.route('/path').get(fn).post(fn).delete(fn)
func findChainedRoutePath(node *sitter.Node, content []byte) string {
	if node == nil || node.Type() != "call_expression" {
		return ""
	}
	funcNode := node.ChildByFieldName("function")
	argsNode := node.ChildByFieldName("arguments")
	if funcNode == nil || funcNode.Type() != "member_expression" {
		return ""
	}
	propNode := funcNode.ChildByFieldName("property")
	if propNode == nil {
		return ""
	}

	if nodeText(propNode, content) == "route" {
		if argsNode != nil {
			for i := 0; i < int(argsNode.ChildCount()); i++ {
				child := argsNode.Child(i)
				if child.Type() == "string" || child.Type() == "template_string" {
					return unquote(nodeText(child, content))
				}
			}
		}
	}

	// Another HTTP method in the chain — recurse into its object
	return findChainedRoutePath(funcNode.ChildByFieldName("object"), content)
}
