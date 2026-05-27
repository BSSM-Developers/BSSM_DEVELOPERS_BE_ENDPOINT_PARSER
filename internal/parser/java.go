package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"

	"endpoint-parser/internal/model"
)

// springMappings maps Spring shorthand annotation names to HTTP methods.
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

// extractSpringEndpoints visits every annotation and marker_annotation node.
// "annotation" has arguments (e.g. @GetMapping("/path")); "marker_annotation" has none (e.g. @GetMapping).
func extractSpringEndpoints(root *sitter.Node, content []byte) []model.Endpoint {
	var endpoints []model.Endpoint
	walkTree(root, func(node *sitter.Node) {
		if node.Type() == "annotation" || node.Type() == "marker_annotation" {
			endpoints = append(endpoints, extractSpringAnnotationEndpoints(node, content)...)
		}
	})
	return endpoints
}

func extractSpringAnnotationEndpoints(node *sitter.Node, content []byte) []model.Endpoint {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	name := nodeText(nameNode, content)

	var httpMethods []string
	if httpMethod, isShorthand := springMappings[name]; isShorthand {
		httpMethods = []string{httpMethod}
	} else if name == "RequestMapping" {
		httpMethods = extractRequestMappingMethods(node, content)
		if len(httpMethods) == 0 {
			return nil // no method attr = class-level prefix annotation, not an endpoint
		}
	} else {
		return nil
	}

	paths := extractAnnotationPaths(node, content)
	if len(paths) == 0 {
		return nil
	}

	basePath := findEnclosingClassBasePath(node, content)

	endpoints := make([]model.Endpoint, 0, len(httpMethods)*len(paths))
	for _, httpMethod := range httpMethods {
		for _, path := range paths {
			endpoints = append(endpoints, model.Endpoint{
				Method: httpMethod,
				Path:   joinPaths(basePath, path),
			})
		}
	}
	return endpoints
}

// extractAnnotationPaths returns all route paths from a Spring annotation.
//
// Returns [""] (single empty string) when the annotation has no arguments,
// which means the route maps directly to the enclosing class's base path.
// Supports: @GetMapping, @GetMapping(), @GetMapping("/path"),
//           @GetMapping(value="/path"), @GetMapping({"/a", "/b"}).
func extractAnnotationPaths(annotation *sitter.Node, content []byte) []string {
	argList := annotation.ChildByFieldName("arguments")
	if argList == nil {
		// marker_annotation: @GetMapping — maps to class base path
		return []string{""}
	}
	for i := 0; i < int(argList.ChildCount()); i++ {
		child := argList.Child(i)
		switch child.Type() {
		case "string_literal":
			return []string{unquote(nodeText(child, content))}
		case "element_value_array_initializer":
			return extractArrayLiteralPaths(child, content)
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
			switch val.Type() {
			case "string_literal":
				return []string{unquote(nodeText(val, content))}
			case "element_value_array_initializer":
				return extractArrayLiteralPaths(val, content)
			}
		}
	}
	// @GetMapping() with empty parens, or annotation with only non-path attributes
	return []string{""}
}

// extractArrayLiteralPaths collects string literals from @GetMapping({"/path1", "/path2"}).
func extractArrayLiteralPaths(array *sitter.Node, content []byte) []string {
	var paths []string
	for i := 0; i < int(array.ChildCount()); i++ {
		child := array.Child(i)
		if child.Type() == "string_literal" {
			paths = append(paths, unquote(nodeText(child, content)))
		}
	}
	return paths
}

// findEnclosingClassBasePath walks up the AST to find the @RequestMapping prefix on the
// enclosing class declaration. Returns "" if the class has no prefix annotation.
func findEnclosingClassBasePath(node *sitter.Node, content []byte) string {
	for p := node.Parent(); p != nil; p = p.Parent() {
		if p.Type() == "class_declaration" {
			return extractClassBasePath(p, content)
		}
	}
	return ""
}

// extractClassBasePath finds the @RequestMapping path declared on a class (no method= attr).
// A class-level @RequestMapping without method= is treated as a URL prefix for all methods.
func extractClassBasePath(classNode *sitter.Node, content []byte) string {
	for i := 0; i < int(classNode.ChildCount()); i++ {
		child := classNode.Child(i)
		if child.Type() != "modifiers" {
			continue
		}
		for j := 0; j < int(child.ChildCount()); j++ {
			annot := child.Child(j)
			if annot.Type() != "annotation" {
				continue
			}
			nameNode := annot.ChildByFieldName("name")
			if nameNode == nil || nodeText(nameNode, content) != "RequestMapping" {
				continue
			}
			// Only use as prefix if no explicit method= attribute
			if len(extractRequestMappingMethods(annot, content)) > 0 {
				continue
			}
			paths := extractAnnotationPaths(annot, content)
			if len(paths) > 0 && paths[0] != "" {
				return paths[0]
			}
		}
	}
	return ""
}

// extractRequestMappingMethods reads the method= attribute from @RequestMapping.
// Supports single (method=RequestMethod.GET) and array (method={GET,POST}) values.
// Returns nil if no method attribute is present (class-level prefix annotation).
func extractRequestMappingMethods(annotation *sitter.Node, content []byte) []string {
	argList := annotation.ChildByFieldName("arguments")
	if argList == nil {
		return nil
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
		if val.Type() == "element_value_array_initializer" {
			return extractMethodsFromArray(val, content)
		}
		return extractMethodsFromText(nodeText(val, content))
	}
	return nil
}

// extractMethodsFromArray collects HTTP verbs from {RequestMethod.GET, RequestMethod.POST}.
func extractMethodsFromArray(array *sitter.Node, content []byte) []string {
	var methods []string
	for i := 0; i < int(array.ChildCount()); i++ {
		methods = append(methods, extractMethodsFromText(nodeText(array.Child(i), content))...)
	}
	return methods
}

// extractMethodsFromText scans a text fragment for known HTTP verb names.
func extractMethodsFromText(text string) []string {
	var methods []string
	for _, m := range []string{"GET", "POST", "PUT", "DELETE", "PATCH"} {
		if strings.Contains(text, m) {
			methods = append(methods, m)
		}
	}
	return methods
}
