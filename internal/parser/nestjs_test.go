package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// @Head(), @Options() decorators
var nestJSAdditionalDecorators = []byte(`
import { Controller, Get, Head, Options, Post, Delete } from '@nestjs/common';

@Controller('resources')
export class ResourceController {

    @Get()
    findAll() { return []; }

    @Get(':id')
    findOne() { return {}; }

    @Post()
    create() { return {}; }

    @Delete(':id')
    remove() { return {}; }

    @Head(':id')
    exists() { return null; }

    @Options()
    preflight() { return null; }
}
`)

// @Controller() with no path argument
var nestJSNoControllerPath = []byte(`
import { Controller, Get, Post } from '@nestjs/common';

@Controller()
export class AppController {

    @Get('/health')
    health() { return 'ok'; }

    @Post('/events')
    createEvent() { return {}; }

    @Get('/version')
    version() { return '1.0.0'; }
}
`)

// multiple NestJS controllers in one file
var nestJSMultiController = []byte(`
import { Controller, Get, Post, Put, Delete } from '@nestjs/common';

@Controller('users')
export class UserController {

    @Get()
    findAll() { return []; }

    @Get(':id')
    findOne() { return {}; }

    @Post()
    create() { return {}; }
}

@Controller('products')
export class ProductController {

    @Get()
    findAll() { return []; }

    @Get(':id')
    findOne() { return {}; }

    @Put(':id')
    update() { return {}; }

    @Delete(':id')
    remove() {}
}
`)

// NestJS with guards and interceptors (non-HTTP decorators should not be captured)
var nestJSWithGuards = []byte(`
import { Controller, Get, Post, UseGuards, UseInterceptors } from '@nestjs/common';

@Controller('admin')
@UseGuards(AdminGuard)
export class AdminController {

    @Get('dashboard')
    @UseInterceptors(LoggingInterceptor)
    dashboard() { return {}; }

    @Post('users')
    @UseGuards(PermissionGuard)
    createUser() { return {}; }
}
`)

// nested path segments in NestJS
var nestJSNestedPaths = []byte(`
import { Controller, Get, Post, Delete, Patch } from '@nestjs/common';

@Controller('orgs/:orgId/repos')
export class RepoController {

    @Get()
    listRepos() { return []; }

    @Get(':repoId')
    getRepo() { return {}; }

    @Get(':repoId/branches')
    listBranches() { return []; }

    @Delete(':repoId/branches/:branch')
    deleteBranch() {}
}
`)

func TestNestJSAdditionalDecorators(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSAdditionalDecorators, Lang: model.LangTypeScript})
	t.Logf("NestJS additional decorators - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/AdditionalDecorators", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/resources"},
		{Method: "GET", Path: "/resources/:id"},
		{Method: "POST", Path: "/resources"},
		{Method: "DELETE", Path: "/resources/:id"},
		{Method: "HEAD", Path: "/resources/:id"},
		{Method: "OPTIONS", Path: "/resources"},
	})
}

func TestNestJSNoControllerPath(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSNoControllerPath, Lang: model.LangTypeScript})
	t.Logf("NestJS no controller path - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/NoControllerPath", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/health"},
		{Method: "POST", Path: "/events"},
		{Method: "GET", Path: "/version"},
	})
}

func TestNestJSMultiController(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSMultiController, Lang: model.LangTypeScript})
	t.Logf("NestJS multi-controller - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/MultiController", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users"},
		{Method: "GET", Path: "/users/:id"},
		{Method: "POST", Path: "/users"},
		{Method: "GET", Path: "/products"},
		{Method: "GET", Path: "/products/:id"},
		{Method: "PUT", Path: "/products/:id"},
		{Method: "DELETE", Path: "/products/:id"},
	})
}

func TestNestJSWithGuardsShouldNotCaptureGuards(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSWithGuards, Lang: model.LangTypeScript})
	t.Logf("NestJS with guards - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/WithGuards", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/admin/dashboard"},
		{Method: "POST", Path: "/admin/users"},
	})
}

func TestNestJSNestedPaths(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSNestedPaths, Lang: model.LangTypeScript})
	t.Logf("NestJS nested paths - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/NestedPaths", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/orgs/:orgId/repos"},
		{Method: "GET", Path: "/orgs/:orgId/repos/:repoId"},
		{Method: "GET", Path: "/orgs/:orgId/repos/:repoId/branches"},
		{Method: "DELETE", Path: "/orgs/:orgId/repos/:repoId/branches/:branch"},
	})
}
