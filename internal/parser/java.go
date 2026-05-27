package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"

	"endpoint-parser/internal/model"
)

var javaLang = java.GetLanguage()

var springMappings = map[string]string{
	"GetMapping":    "GET",
	"PostMapping":   "POST",
	"PutMapping":    "PUT",
	"DeleteMapping": "DELETE",
	"PatchMapping":  "PATCH",
}

func parseJava(content []byte) []model.Endpoint {
	p := sitter.NewParser()
	p.SetLanguage(javaLang)
	tree := p.Parse(nil, content)
	defer tree.Close()

	var endpoints []model.Endpoint

	walkTree(tree.RootNode(), func(node *sitter.Node) {
		if node.Type() != "annotation" {
			return
		}
		nameNode := node.ChildByFieldName("name")
		if nameNode == nil {
			return
		}
		name := nodeText(nameNode, content)

		httpMethod, ok := springMappings[name]
		if !ok {
			if name != "RequestMapping" {
				return
			}
		}

		path := javaAnnotationPath(node, content)
		if path == "" {
			return
		}

		if name == "RequestMapping" {
			httpMethod = javaRequestMappingMethod(node, content)
			if httpMethod == "" {
				return // no method attr = class-level prefix, skip
			}
		}

		endpoints = append(endpoints, model.Endpoint{Method: httpMethod, Path: path})
	})

	return endpoints
}

func javaAnnotationPath(annotation *sitter.Node, content []byte) string {
	argList := annotation.ChildByFieldName("arguments")
	if argList == nil {
		return ""
	}
	for i := 0; i < int(argList.ChildCount()); i++ {
		child := argList.Child(i)
		switch child.Type() {
		case "string_literal":
			return unquote(nodeText(child, content))
		case "element_value_pair":
			key := child.ChildByFieldName("key")
			val := child.ChildByFieldName("value")
			if key == nil || val == nil {
				continue
			}
			k := nodeText(key, content)
			if k != "value" && k != "path" {
				continue
			}
			if val.Type() == "string_literal" {
				return unquote(nodeText(val, content))
			}
		}
	}
	return ""
}

func javaRequestMappingMethod(annotation *sitter.Node, content []byte) string {
	argList := annotation.ChildByFieldName("arguments")
	if argList == nil {
		return ""
	}
	for i := 0; i < int(argList.ChildCount()); i++ {
		child := argList.Child(i)
		if child.Type() != "element_value_pair" {
			continue
		}
		key := child.ChildByFieldName("key")
		if key == nil || nodeText(key, content) != "method" {
			continue
		}
		val := child.ChildByFieldName("value")
		if val == nil {
			continue
		}
		text := nodeText(val, content)
		for _, m := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
			if strings.Contains(text, m) {
				return m
			}
		}
	}
	return ""
}
