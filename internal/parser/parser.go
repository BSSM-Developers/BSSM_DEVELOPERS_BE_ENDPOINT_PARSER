package parser

import "endpoint-parser/internal/model"

// LanguageParser extracts HTTP endpoints from source code of a specific language.
type LanguageParser interface {
	Language() model.Language
	Parse(content []byte) []model.Endpoint
}

// registry maps each supported language to its parser.
// To add a new language: implement LanguageParser and add it here.
var registry = map[model.Language]LanguageParser{
	model.LangJava:       &javaParser{},
	model.LangPython:     &pythonParser{},
	model.LangJavaScript: &javaScriptParser{},
	model.LangTypeScript: &typeScriptParser{},
}

// Parse delegates to the appropriate LanguageParser for the file's language.
// Returns nil for unsupported languages or on parse panic.
func Parse(file model.FileContent) (endpoints []model.Endpoint) {
	defer func() {
		if r := recover(); r != nil {
			endpoints = nil
		}
	}()

	p, ok := registry[file.Lang]
	if !ok {
		return nil
	}
	return p.Parse(file.Content)
}
