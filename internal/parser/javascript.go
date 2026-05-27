package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"

	"endpoint-parser/internal/model"
)

var jsLang = javascript.GetLanguage()

var jsHTTPMethods = map[string]string{
	"get":    "GET",
	"post":   "POST",
	"put":    "PUT",
	"delete": "DELETE",
	"patch":  "PATCH",
}

func parseJavaScript(content []byte) []model.Endpoint {
	p := sitter.NewParser()
	p.SetLanguage(jsLang)
	tree := p.Parse(nil, content)
	defer tree.Close()

	return extractExpressEndpoints(tree.RootNode(), content)
}

func extractExpressEndpoints(root *sitter.Node, content []byte) []model.Endpoint {
	var endpoints []model.Endpoint
	walkTree(root, func(node *sitter.Node) {
		if node.Type() != "call_expression" {
			return
		}
		method, path := expressRouteFromCall(node, content)
		if method != "" && path != "" {
			endpoints = append(endpoints, model.Endpoint{Method: method, Path: path})
		}
	})
	return endpoints
}

func expressRouteFromCall(call *sitter.Node, content []byte) (string, string) {
	funcNode := call.ChildByFieldName("function")
	argsNode := call.ChildByFieldName("arguments")
	if funcNode == nil || argsNode == nil {
		return "", ""
	}
	if funcNode.Type() != "member_expression" {
		return "", ""
	}

	propNode := funcNode.ChildByFieldName("property")
	if propNode == nil {
		return "", ""
	}

	httpMethod, ok := jsHTTPMethods[strings.ToLower(nodeText(propNode, content))]
	if !ok {
		return "", ""
	}

	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		if child.Type() == "string" || child.Type() == "template_string" {
			return httpMethod, unquote(nodeText(child, content))
		}
	}
	return "", ""
}
