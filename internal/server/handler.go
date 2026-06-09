package server

import (
	"encoding/json"
	"log"
	"net/http"

	"endpoint-parser/internal/fetcher"
	"endpoint-parser/internal/model"
	"endpoint-parser/internal/parser"
	"golang.org/x/sync/singleflight"
)

type Handler struct {
	fetcher *fetcher.GitHubFetcher
	group   singleflight.Group
}

func NewHandler() *Handler {
	return &Handler{fetcher: fetcher.New()}
}

func (h *Handler) HandleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.RepoFullName == "" || req.Branch == "" || req.InstallationToken == "" {
		http.Error(w, "repoFullName, branch, installationToken are required", http.StatusBadRequest)
		return
	}

	key := req.RepoFullName + "@" + req.Branch
	v, err, _ := h.group.Do(key, func() (interface{}, error) {
		return h.fetcher.FetchRelevantFiles(req.RepoFullName, req.Branch, req.InstallationToken)
	})
	if err != nil {
		log.Printf("fetch error repo=%s: %v", req.RepoFullName, err)
		http.Error(w, "failed to fetch repository files", http.StatusInternalServerError)
		return
	}

	files := v.([]model.FileContent)
	var endpoints []model.Endpoint
	for _, file := range files {
		parsed := parser.Parse(file)
		log.Printf("parsed %s: %d endpoints", file.Path, len(parsed))
		endpoints = append(endpoints, parsed...)
	}

	if endpoints == nil {
		endpoints = []model.Endpoint{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.ParseResponse{Endpoints: endpoints})
}
