package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"

	"endpoint-parser/internal/model"
)

// springMappings maps Spring annotation names to HTTP methods.
var springMappings = map[string]string{
	"GetMapping":    "GET",
	"PostMapping":   "POST",
	"PutMapping":    "PUT",
	"DeleteMapping": "DELETE",
	"PatchMapping":  "PATCH",
}

type javaParser struct{}

func (javaParser) Language() model.Language { return model.LangJava }

func (javaParser) Parse(content []byte) []model.Endpoint {
	return parseWithTreeSitter(java.GetLanguage(), content, extractSpringEndpoints)
}

// extractSpringEndpoints collects @GetMapping / @PostMapping / @RequestMapping annotations.
func extractSpringEndpoints(root *sitter.Node, content []byte) []model.Endpoint {
	var endpoints []model.Endpoint
	walkTree(root, func(node *sitter.Node) {
		if endpoint, ok := extractSpringAnnotation(node, content); ok {
			endpoints = append(endpoints, endpoint)
		}
	})
	return endpoints
}

func extractSpringAnnotation(node *sitter.Node, content []byte) (model.Endpoint, bool) {
	if node.Type() != "annotation" {
		return model.Endpoint{}, false
	}
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return model.Endpoint{}, false
	}
	name := nodeText(nameNode, content)

	httpMethod, isShorthand := springMappings[name]
	if !isShorthand && name != "RequestMapping" {
		return model.Endpoint{}, false
	}

	path := extractAnnotationPath(node, content)
	if path == "" {
		return model.Endpoint{}, false
	}

	if name == "RequestMapping" {
		httpMethod = extractRequestMappingMethod(node, content)
		if httpMethod == "" {
			// no method attr = class-level prefix, not an endpoint
			return model.Endpoint{}, false
		}
	}

	return model.Endpoint{Method: httpMethod, Path: path}, true
}

// extractAnnotationPath reads the path string from annotation arguments.
// Handles both @GetMapping("/path") and @GetMapping(value="/path") / @GetMapping(path="/path").
func extractAnnotationPath(annotation *sitter.Node, content []byte) string {
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
			if (k == "value" || k == "path") && val.Type() == "string_literal" {
				return unquote(nodeText(val, content))
			}
		}
	}
	return ""
}

// extractRequestMappingMethod reads the method= attribute from @RequestMapping.
func extractRequestMappingMethod(annotation *sitter.Node, content []byte) string {
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
