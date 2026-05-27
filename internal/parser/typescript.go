package parser

import (
	sitter "github.com/smacker/go-tree-sitter"
	tsLang "github.com/smacker/go-tree-sitter/typescript/typescript"

	"endpoint-parser/internal/model"
)

// nestDecorators maps NestJS HTTP decorator names to HTTP verbs.
var nestDecorators = map[string]string{
	"Get":     "GET",
	"Post":    "POST",
	"Put":     "PUT",
	"Delete":  "DELETE",
	"Patch":   "PATCH",
	"Head":    "HEAD",
	"Options": "OPTIONS",
	"All":     "ALL",
}

type typeScriptParser struct{}

func (typeScriptParser) Language() model.Language { return model.LangTypeScript }

func (typeScriptParser) Parse(content []byte) []model.Endpoint {
	return parseWithTreeSitter(tsLang.GetLanguage(), content, extractTypeScriptEndpoints)
}

// extractTypeScriptEndpoints handles NestJS class-based controllers and Express-style routes.
// NestJS is checked first; Express-style routes are collected as a fallback (deduped).
func extractTypeScriptEndpoints(root *sitter.Node, content []byte) []model.Endpoint {
	var endpoints []model.Endpoint

	walkTree(root, func(node *sitter.Node) {
		if node.Type() == "class_declaration" || node.Type() == "abstract_class_declaration" {
			endpoints = append(endpoints, extractNestControllerEndpoints(node, content)...)
		}
	})

	endpoints = append(endpoints, extractExpressEndpoints(root, content)...)
	return deduplicateEndpoints(endpoints)
}

// extractNestControllerEndpoints extracts routes from a NestJS @Controller class.
func extractNestControllerEndpoints(classNode *sitter.Node, content []byte) []model.Endpoint {
	basePath := resolveControllerBasePath(classNode, content)
	classBody := findClassBody(classNode)
	if classBody == nil {
		return nil
	}
	return extractMethodEndpoints(classBody, basePath, content)
}

// resolveControllerBasePath finds the @Controller('prefix') path for the class.
// The decorator may appear on the class itself or on its parent export_statement.
func resolveControllerBasePath(classNode *sitter.Node, content []byte) string {
	if path, ok := findControllerDecorator(classNode, content); ok {
		return path
	}
	if path, ok := findControllerDecorator(classNode.Parent(), content); ok {
		return path
	}
	return ""
}

func findControllerDecorator(node *sitter.Node, content []byte) (string, bool) {
	if node == nil {
		return "", false
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "decorator" {
			if path, ok := decoratorArg(child, "Controller", content); ok {
				return path, true
			}
		}
	}
	return "", false
}

// extractMethodEndpoints iterates class_body children, associating decorators with the
// method_definition that immediately follows them.
func extractMethodEndpoints(classBody *sitter.Node, basePath string, content []byte) []model.Endpoint {
	var endpoints []model.Endpoint
	var pendingDecorators []*sitter.Node

	for i := 0; i < int(classBody.ChildCount()); i++ {
		member := classBody.Child(i)
		switch member.Type() {
		case "decorator":
			pendingDecorators = append(pendingDecorators, member)
		case "method_definition", "abstract_method_signature":
			for _, dec := range pendingDecorators {
				for decoratorName, httpMethod := range nestDecorators {
					if path, ok := decoratorArg(dec, decoratorName, content); ok {
						endpoints = append(endpoints, model.Endpoint{
							Method: httpMethod,
							Path:   joinPaths(basePath, path),
						})
					}
				}
			}
			pendingDecorators = nil
		case "public_field_definition":
			// abstract decorated members have their decorator as a child, not a sibling
			pendingDecorators = nil
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
		default:
			pendingDecorators = nil
		}
	}
	return endpoints
}

func findClassBody(classNode *sitter.Node) *sitter.Node {
	for i := 0; i < int(classNode.ChildCount()); i++ {
		if classNode.Child(i).Type() == "class_body" {
			return classNode.Child(i)
		}
	}
	return nil
}

// decoratorArg returns the first string argument of a decorator call with the given name.
// e.g. @Get(':id') → ":id", true   |   @Get() → "", true   |   @Body() → "", false
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

