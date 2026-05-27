package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/python"

	"endpoint-parser/internal/model"
)

var pythonLang = python.GetLanguage()

func parsePython(content []byte) []model.Endpoint {
	p := sitter.NewParser()
	p.SetLanguage(pythonLang)
	tree := p.Parse(nil, content)
	defer tree.Close()

	var endpoints []model.Endpoint

	walkTree(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() != "decorator" {
			return
		}
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() != "call" {
				continue
			}
			method, path := pythonRouteFromCall(child, content)
			if method != "" && path != "" {
				endpoints = append(endpoints, model.Endpoint{Method: method, Path: path})
			}
		}
	})

	return endpoints
}

func pythonRouteFromCall(call *sitter.Node, content []byte) (string, string) {
	funcNode := call.ChildByFieldName("function")
	argsNode := call.ChildByFieldName("arguments")
	if funcNode == nil || argsNode == nil {
		return "", ""
	}

	var methodName string
	switch funcNode.Type() {
	case "attribute":
		attr := funcNode.ChildByFieldName("attribute")
		if attr != nil {
			methodName = nodeText(attr, content)
		}
	case "identifier":
		methodName = nodeText(funcNode, content)
	}

	httpMethod := pythonHTTPMethod(methodName)
	if httpMethod == "" {
		return "", ""
	}

	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		if child.Type() == "string" {
			return httpMethod, unquote(nodeText(child, content))
		}
	}
	return "", ""
}

func pythonHTTPMethod(method string) string {
	switch strings.ToLower(method) {
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
	case "route":
		return "GET"
	}
	return ""
}
