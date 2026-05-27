package parser

import "endpoint-parser/internal/model"

func Parse(file model.FileContent) (endpoints []model.Endpoint) {
	defer func() {
		if r := recover(); r != nil {
			endpoints = nil
		}
	}()

	switch file.Lang {
	case model.LangJava:
		return parseJava(file.Content)
	case model.LangPython:
		return parsePython(file.Content)
	case model.LangJavaScript:
		return parseJavaScript(file.Content)
	case model.LangTypeScript:
		return parseTypeScript(file.Content)
	}
	return nil
}
