package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// class-level @RequestMapping prefix + no-arg method annotations
var springClassPrefixSample = []byte(`
package com.example.demo;

import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1/users")
public class UserController {

    @GetMapping
    public List<User> list() { return service.list(); }

    @PostMapping
    public User create(@RequestBody UserRequest req) { return service.create(req); }

    @GetMapping("/{id}")
    public User get(@PathVariable Long id) { return service.get(id); }

    @PutMapping("/{id}")
    public User update(@PathVariable Long id, @RequestBody UserRequest req) { return service.update(id, req); }

    @DeleteMapping("/{id}")
    public void delete(@PathVariable Long id) { service.delete(id); }

    @PatchMapping("/{id}/status")
    public void updateStatus(@PathVariable Long id) { service.updateStatus(id); }
}
`)

// array paths: @GetMapping({"/path1", "/path2"})
var springArrayPathsSample = []byte(`
@RestController
public class HealthController {

    @GetMapping({"/health", "/healthz", "/ping"})
    public String health() { return "ok"; }

    @RequestMapping(value = {"/status", "/alive"}, method = RequestMethod.GET)
    public String status() { return "alive"; }
}
`)

// multiple controllers in one file, each with different prefixes
var springMultiControllerSample = []byte(`
@RestController
@RequestMapping("/api/users")
public class UserController {

    @GetMapping
    public List<User> list() { return null; }

    @PostMapping
    public User create() { return null; }

    @GetMapping("/{id}")
    public User get() { return null; }
}

@RestController
@RequestMapping("/api/posts")
public class PostController {

    @GetMapping
    public List<Post> list() { return null; }

    @GetMapping("/{id}")
    public Post get() { return null; }

    @DeleteMapping("/{id}")
    public void delete() {}
}
`)

// explicit empty parens @GetMapping() — same as no args
var springEmptyArgsSample = []byte(`
@RestController
@RequestMapping("/api/items")
public class ItemController {

    @GetMapping()
    public List<Item> list() { return null; }

    @PostMapping()
    public Item create() { return null; }
}
`)

// no class prefix — paths should stay as-is
var springNoPrefixSample = []byte(`
@RestController
public class VersionController {

    @GetMapping("/version")
    public String version() { return "1.0"; }

    @GetMapping(value = "/info")
    public String info() { return "{}"; }

    @PostMapping(path = "/reset")
    public void reset() {}
}
`)

// @RequestMapping with explicit method= on method level + class prefix
var springRequestMappingMethodSample = []byte(`
@RestController
@RequestMapping("/api/legacy")
public class LegacyController {

    @RequestMapping(value = "/items", method = RequestMethod.GET)
    public List<Item> list() { return null; }

    @RequestMapping(path = "/items", method = RequestMethod.POST)
    public Item create() { return null; }
}
`)

// deeply nested path segments
var springDeepPathSample = []byte(`
@RestController
@RequestMapping("/api/v2/orgs/{orgId}")
public class OrgRepoController {

    @GetMapping("/repos")
    public List<Repo> listRepos() { return null; }

    @GetMapping("/repos/{repoId}/branches")
    public List<Branch> listBranches() { return null; }

    @DeleteMapping("/repos/{repoId}/branches/{branch}")
    public void deleteBranch() {}
}
`)

func TestSpringBootClassPrefixCombination(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springClassPrefixSample, Lang: model.LangJava})
	t.Logf("Spring Boot class prefix - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/ClassPrefix", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/api/v1/users"},
		{Method: "POST", Path: "/api/v1/users"},
		{Method: "GET", Path: "/api/v1/users/{id}"},
		{Method: "PUT", Path: "/api/v1/users/{id}"},
		{Method: "DELETE", Path: "/api/v1/users/{id}"},
		{Method: "PATCH", Path: "/api/v1/users/{id}/status"},
	})
}

func TestSpringBootArrayPaths(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springArrayPathsSample, Lang: model.LangJava})
	t.Logf("Spring Boot array paths - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/ArrayPaths", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/health"},
		{Method: "GET", Path: "/healthz"},
		{Method: "GET", Path: "/ping"},
		{Method: "GET", Path: "/status"},
		{Method: "GET", Path: "/alive"},
	})
}

func TestSpringBootMultiController(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springMultiControllerSample, Lang: model.LangJava})
	t.Logf("Spring Boot multi-controller - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/MultiController", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/api/users"},
		{Method: "POST", Path: "/api/users"},
		{Method: "GET", Path: "/api/users/{id}"},
		{Method: "GET", Path: "/api/posts"},
		{Method: "GET", Path: "/api/posts/{id}"},
		{Method: "DELETE", Path: "/api/posts/{id}"},
	})
}

func TestSpringBootEmptyArgAnnotation(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springEmptyArgsSample, Lang: model.LangJava})
	t.Logf("Spring Boot empty args - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/EmptyArgs", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/api/items"},
		{Method: "POST", Path: "/api/items"},
	})
}

func TestSpringBootNoPrefix(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springNoPrefixSample, Lang: model.LangJava})
	t.Logf("Spring Boot no prefix - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "SpringBoot/NoPrefix", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/version"},
		{Method: "GET", Path: "/info"},
		{Method: "POST", Path: "/reset"},
	})
}

func TestSpringBootRequestMappingWithMethod(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springRequestMappingMethodSample, Lang: model.LangJava})
	t.Logf("Spring Boot @RequestMapping + method= - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/RequestMappingMethod", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/api/legacy/items"},
		{Method: "POST", Path: "/api/legacy/items"},
	})
}

func TestSpringBootDeepPaths(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springDeepPathSample, Lang: model.LangJava})
	t.Logf("Spring Boot deep paths - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/DeepPaths", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/api/v2/orgs/{orgId}/repos"},
		{Method: "GET", Path: "/api/v2/orgs/{orgId}/repos/{repoId}/branches"},
		{Method: "DELETE", Path: "/api/v2/orgs/{orgId}/repos/{repoId}/branches/{branch}"},
	})
}
