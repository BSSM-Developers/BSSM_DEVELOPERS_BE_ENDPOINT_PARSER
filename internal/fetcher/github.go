package fetcher

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"endpoint-parser/internal/model"
)

const maxFileSizeBytes = 100 * 1024 // 100KB

var routeNamePatterns = []string{
	"controller", "router", "routes", "handler", "resource",
}

var routeDirPatterns = []string{
	"api", "apis", "routes", "controllers", "endpoints", "views", "handlers",
}

var alwaysIncludeNames = []string{
	"app.js", "app.ts", "index.js", "server.js", "main.py", "app.py", "server.py",
}

type GitHubFetcher struct {
	client *http.Client
}

func New() *GitHubFetcher {
	return &GitHubFetcher{client: &http.Client{}}
}

type treeResponse struct {
	Tree []treeEntry `json:"tree"`
}

type treeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Size int    `json:"size"`
}

type contentsResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func (f *GitHubFetcher) FetchRelevantFiles(repoFullName, branch, token string) ([]model.FileContent, error) {
	parts := strings.SplitN(repoFullName, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repoFullName: %s", repoFullName)
	}
	owner, repo := parts[0], parts[1]

	tree, err := f.fetchTree(owner, repo, branch, token)
	if err != nil {
		return nil, fmt.Errorf("fetch tree: %w", err)
	}

	var files []model.FileContent
	for _, entry := range tree.Tree {
		if entry.Type != "blob" || entry.Size > maxFileSizeBytes {
			continue
		}
		lang, ok := detectLang(entry.Path)
		if !ok || !isRouteFile(entry.Path) {
			continue
		}
		content, err := f.fetchContent(owner, repo, entry.Path, branch, token)
		if err != nil {
			continue
		}
		files = append(files, model.FileContent{Path: entry.Path, Content: content, Lang: lang})
	}
	return files, nil
}

func (f *GitHubFetcher) fetchTree(owner, repo, branch, token string) (*treeResponse, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, branch)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result treeResponse
	return &result, json.NewDecoder(resp.Body).Decode(&result)
}

func (f *GitHubFetcher) fetchContent(owner, repo, path, branch, token string) ([]byte, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cr contentsResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, err
	}
	if cr.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected encoding: %s", cr.Encoding)
	}
	return base64.StdEncoding.DecodeString(strings.ReplaceAll(cr.Content, "\n", ""))
}

func detectLang(path string) (model.Language, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".java":
		return model.LangJava, true
	case ".py":
		return model.LangPython, true
	case ".js":
		if strings.HasSuffix(path, ".min.js") || strings.HasSuffix(path, ".test.js") || strings.HasSuffix(path, ".spec.js") {
			return "", false
		}
		return model.LangJavaScript, true
	case ".ts":
		if strings.HasSuffix(path, ".d.ts") || strings.HasSuffix(path, ".spec.ts") || strings.HasSuffix(path, ".test.ts") {
			return "", false
		}
		return model.LangTypeScript, true
	}
	return "", false
}

func isRouteFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	for _, name := range alwaysIncludeNames {
		if base == name {
			return true
		}
	}
	for _, pattern := range routeNamePatterns {
		if strings.Contains(base, pattern) {
			return true
		}
	}
	dir := strings.ToLower(filepath.ToSlash(filepath.Dir(path)))
	for _, part := range strings.Split(dir, "/") {
		for _, pattern := range routeDirPatterns {
			if part == pattern {
				return true
			}
		}
	}
	return false
}
