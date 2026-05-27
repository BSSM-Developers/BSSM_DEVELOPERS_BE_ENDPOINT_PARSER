package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// @All() catches all HTTP methods
var nestJSAllDecorator = []byte(`
import { Controller, Get, All } from '@nestjs/common';

@Controller('resources')
export class ResourceController {

    @Get()
    findAll() { return []; }

    @All('catch-all')
    catchAll() { return {}; }

    @All()
    fallback() { return null; }
}
`)

// non-HTTP decorators on class and method — must not be captured
var nestJSNonHTTPDecoratorsOnClass = []byte(`
import { Controller, Get, Injectable } from '@nestjs/common';

@Injectable()
export class UserService {

    findAll() { return []; }
}

@Controller('users')
export class UserController {

    @Get()
    findAll() { return []; }
}
`)

// abstract base controller — routes should still be captured
var nestJSAbstractController = []byte(`
import { Controller, Get, Post } from '@nestjs/common';

@Controller('base')
export abstract class BaseController {

    @Get('health')
    health() { return 'ok'; }

    @Get('version')
    abstract version(): string;
}

@Controller('users')
export class UserController {

    @Get()
    list() { return []; }

    @Post()
    create() { return {}; }
}
`)

// method-level non-HTTP decorators stacked before HTTP decorator
var nestJSStackedNonHTTPFirst = []byte(`
import { Controller, Get, Post, UseGuards, UseInterceptors, UsePipes } from '@nestjs/common';

@Controller('admin')
export class AdminController {

    @UseGuards(AuthGuard)
    @UseInterceptors(LoggingInterceptor)
    @Get('dashboard')
    getDashboard() { return {}; }

    @UsePipes(ValidationPipe)
    @UseGuards(AdminGuard)
    @Post('settings')
    updateSettings() { return {}; }

    @UseGuards(AuthGuard)
    @UseGuards(AdminGuard)
    @Get('reports')
    getReports() { return []; }
}
`)

// NestJS with versioning-style controller object arg (not yet supported, should not crash)
var nestJSComplexControllerArg = []byte(`
import { Controller, Get, Post } from '@nestjs/common';

@Controller({ path: 'tasks', version: '1' })
export class TaskV1Controller {

    @Get()
    listV1() { return []; }
}

@Controller('tasks')
export class TaskController {

    @Get()
    list() { return []; }

    @Post()
    create() { return {}; }
}
`)

// private / protected methods — decorators should still be captured
// (NestJS routing ignores TS access modifiers at runtime)
var nestJSAccessModifiers = []byte(`
import { Controller, Get, Post, Delete } from '@nestjs/common';

@Controller('items')
export class ItemController {

    @Get()
    public findAll() { return []; }

    @Get(':id')
    protected findOne() { return {}; }

    @Post()
    private create() { return {}; }

    @Delete(':id')
    async remove() {}
}
`)

// root path in controller
var nestJSRootPaths = []byte(`
import { Controller, Get, Post } from '@nestjs/common';

@Controller()
export class AppController {

    @Get('/')
    root() { return 'ok'; }

    @Post('/')
    submit() { return {}; }
}

@Controller('api')
export class ApiController {

    @Get('/')
    apiRoot() { return {}; }

    @Get('/health')
    health() { return 'ok'; }
}
`)

// NestJS with only non-HTTP decorator on a method (no route → not captured)
var nestJSMethodWithOnlyNonHTTP = []byte(`
import { Controller, Get, UseGuards } from '@nestjs/common';

@Controller('secure')
export class SecureController {

    @Get('data')
    getData() { return {}; }

    @UseGuards(AuthGuard)
    checkAuth() { return true; }
}
`)

func TestNestJSAllDecorator(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSAllDecorator, Lang: model.LangTypeScript})
	t.Logf("NestJS @All decorator - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/AllDecorator", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/resources"},
		{Method: "ALL", Path: "/resources/catch-all"},
		{Method: "ALL", Path: "/resources"},
	})
}

func TestNestJSNonHTTPClassNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSNonHTTPDecoratorsOnClass, Lang: model.LangTypeScript})
	t.Logf("NestJS non-HTTP class - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// @Injectable() class has no routes; @Controller() class has one
	assertEndpointsUnordered(t, "NestJS/NonHTTPClass", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users"},
	})
}

func TestNestJSAbstractController(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSAbstractController, Lang: model.LangTypeScript})
	t.Logf("NestJS abstract controller - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/AbstractController", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/base/health"},
		{Method: "GET", Path: "/base/version"},
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
	})
}

func TestNestJSStackedNonHTTPDecoratorsFirst(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSStackedNonHTTPFirst, Lang: model.LangTypeScript})
	t.Logf("NestJS stacked non-HTTP first - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/StackedNonHTTPFirst", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/admin/dashboard"},
		{Method: "POST", Path: "/admin/settings"},
		{Method: "GET", Path: "/admin/reports"},
	})
}

func TestNestJSComplexControllerArgDoesNotCrash(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSComplexControllerArg, Lang: model.LangTypeScript})
	t.Logf("NestJS complex controller arg - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// @Controller({path:'tasks'}) object arg: base path = "" (not a string arg → graceful skip)
	// @Get() with no prefix → "/"
	// @Controller('tasks'): base path = "tasks"
	assertEndpointsUnordered(t, "NestJS/ComplexControllerArg", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/"},
		{Method: "GET", Path: "/tasks"},
		{Method: "POST", Path: "/tasks"},
	})
}

func TestNestJSAccessModifiersAllCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSAccessModifiers, Lang: model.LangTypeScript})
	t.Logf("NestJS access modifiers - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// public/protected/private don't affect routing in NestJS
	assertEndpointsUnordered(t, "NestJS/AccessModifiers", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/items"},
		{Method: "GET", Path: "/items/:id"},
		{Method: "POST", Path: "/items"},
		{Method: "DELETE", Path: "/items/:id"},
	})
}

func TestNestJSRootPaths(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSRootPaths, Lang: model.LangTypeScript})
	t.Logf("NestJS root paths - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "NestJS/RootPaths", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/"},
		{Method: "POST", Path: "/"},
		{Method: "GET", Path: "/api"},
		{Method: "GET", Path: "/api/health"},
	})
}

func TestNestJSMethodWithOnlyNonHTTPNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: nestJSMethodWithOnlyNonHTTP, Lang: model.LangTypeScript})
	t.Logf("NestJS method with only non-HTTP decorator - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// checkAuth has only @UseGuards, no HTTP decorator → not a route
	assertEndpointsUnordered(t, "NestJS/MethodOnlyNonHTTP", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/secure/data"},
	})
}
