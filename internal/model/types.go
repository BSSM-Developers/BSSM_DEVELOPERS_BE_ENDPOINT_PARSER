package model

type Language string

const (
	LangJava       Language = "java"
	LangPython     Language = "python"
	LangJavaScript Language = "javascript"
	LangTypeScript Language = "typescript"
)

type ParseRequest struct {
	RepoFullName      string `json:"repoFullName"`
	Branch            string `json:"branch"`
	InstallationToken string `json:"installationToken"`
}

type Endpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type ParseResponse struct {
	Endpoints []Endpoint `json:"endpoints"`
}

type FileContent struct {
	Path    string
	Content []byte
	Lang    Language
}
