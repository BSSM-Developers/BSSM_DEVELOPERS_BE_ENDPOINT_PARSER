package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// ─── Spring Boot ──────────────────────────────────────────────────────────────

var springBootSample = []byte(`
package com.example.demo.controller;

import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/v1")
public class UserController {

    @GetMapping("/users")
    public List<UserResponse> getUsers() {
        return userService.findAll();
    }

    @GetMapping("/users/{id}")
    public UserResponse getUser(@PathVariable Long id) {
        return userService.findById(id);
    }

    @PostMapping("/users")
    public UserResponse createUser(@RequestBody CreateUserRequest req) {
        return userService.create(req);
    }

    @PutMapping("/users/{id}")
    public UserResponse updateUser(@PathVariable Long id, @RequestBody UpdateUserRequest req) {
        return userService.update(id, req);
    }

    @DeleteMapping("/users/{id}")
    public void deleteUser(@PathVariable Long id) {
        userService.delete(id);
    }

    @PatchMapping("/users/{id}/role")
    public void updateRole(@PathVariable Long id, @RequestBody RoleRequest req) {
        userService.updateRole(id, req);
    }

    @RequestMapping(value = "/users/{id}/status", method = RequestMethod.POST)
    public void updateStatus(@PathVariable Long id) {
        userService.activate(id);
    }
}
`)

var springBootSample2 = []byte(`
package com.example.demo.controller;

import org.springframework.web.bind.annotation.*;

@RestController
public class ProductController {

    @GetMapping(value = "/products")
    public List<Product> list() { return productService.findAll(); }

    @PostMapping(path = "/products")
    public Product create(@RequestBody ProductRequest req) { return productService.create(req); }

    @GetMapping("/products/{id}/reviews")
    public List<Review> reviews(@PathVariable Long id) { return reviewService.findByProduct(id); }
}
`)

// ─── FastAPI ──────────────────────────────────────────────────────────────────

var fastAPISample = []byte(`
from fastapi import FastAPI, APIRouter

app = FastAPI()
router = APIRouter(prefix="/api/v1")

@app.get("/health")
async def health():
    return {"status": "ok"}

@router.get("/users")
async def get_users():
    return []

@router.post("/users")
async def create_user(req: UserCreate):
    return {}

@router.get("/users/{user_id}")
async def get_user(user_id: int):
    return {}

@router.put("/users/{user_id}")
async def update_user(user_id: int, req: UserUpdate):
    return {}

@router.delete("/users/{user_id}")
async def delete_user(user_id: int):
    return {}

@router.patch("/users/{user_id}/role")
async def update_role(user_id: int, req: RoleUpdate):
    return {}
`)

// ─── Express ─────────────────────────────────────────────────────────────────

var expressSample = []byte(`
const express = require('express');
const router = express.Router();
const app = express();

app.get('/health', (req, res) => res.json({ status: 'ok' }));

router.get('/posts', async (req, res) => {
    const posts = await Post.findAll();
    res.json(posts);
});

router.get('/posts/:id', async (req, res) => {
    const post = await Post.findByPk(req.params.id);
    res.json(post);
});

router.post('/posts', async (req, res) => {
    const post = await Post.create(req.body);
    res.status(201).json(post);
});

router.put('/posts/:id', async (req, res) => {
    await Post.update(req.body, { where: { id: req.params.id } });
    res.json({ ok: true });
});

router.delete('/posts/:id', async (req, res) => {
    await Post.destroy({ where: { id: req.params.id } });
    res.status(204).end();
});
`)

// ─── NestJS ───────────────────────────────────────────────────────────────────

var nestJSSample = []byte(`
import { Controller, Get, Post, Put, Delete, Patch, Param, Body } from '@nestjs/common';

@Controller('comments')
export class CommentController {
    constructor(private readonly commentService: CommentService) {}

    @Get()
    findAll() {
        return this.commentService.findAll();
    }

    @Get(':id')
    findOne(@Param('id') id: string) {
        return this.commentService.findOne(+id);
    }

    @Post()
    create(@Body() dto: CreateCommentDto) {
        return this.commentService.create(dto);
    }

    @Put(':id')
    update(@Param('id') id: string, @Body() dto: UpdateCommentDto) {
        return this.commentService.update(+id, dto);
    }

    @Delete(':id')
    remove(@Param('id') id: string) {
        return this.commentService.remove(+id);
    }

    @Patch(':id/approve')
    approve(@Param('id') id: string) {
        return this.commentService.approve(+id);
    }
}
`)

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestParseSpringBoot(t *testing.T) {
	endpoints := Parse(model.FileContent{
		Path:    "UserController.java",
		Content: springBootSample,
		Lang:    model.LangJava,
	})

	t.Logf("Spring Boot UserController - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}

	expected := []model.Endpoint{
		{Method: "GET", Path: "/users"},
		{Method: "GET", Path: "/users/{id}"},
		{Method: "POST", Path: "/users"},
		{Method: "PUT", Path: "/users/{id}"},
		{Method: "DELETE", Path: "/users/{id}"},
		{Method: "PATCH", Path: "/users/{id}/role"},
		{Method: "POST", Path: "/users/{id}/status"},
	}

	assertEndpoints(t, "SpringBoot UserController", endpoints, expected)
}

func TestParseSpringBootValuePath(t *testing.T) {
	endpoints := Parse(model.FileContent{
		Path:    "ProductController.java",
		Content: springBootSample2,
		Lang:    model.LangJava,
	})

	t.Logf("Spring Boot ProductController - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}

	expected := []model.Endpoint{
		{Method: "GET", Path: "/products"},
		{Method: "POST", Path: "/products"},
		{Method: "GET", Path: "/products/{id}/reviews"},
	}

	assertEndpoints(t, "SpringBoot ProductController", endpoints, expected)
}

func TestParseFastAPI(t *testing.T) {
	endpoints := Parse(model.FileContent{
		Path:    "users.py",
		Content: fastAPISample,
		Lang:    model.LangPython,
	})

	t.Logf("FastAPI - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}

	expected := []model.Endpoint{
		{Method: "GET", Path: "/health"},
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
		{Method: "GET", Path: "/users/{user_id}"},
		{Method: "PUT", Path: "/users/{user_id}"},
		{Method: "DELETE", Path: "/users/{user_id}"},
		{Method: "PATCH", Path: "/users/{user_id}/role"},
	}

	assertEndpoints(t, "FastAPI", endpoints, expected)
}

func TestParseExpress(t *testing.T) {
	endpoints := Parse(model.FileContent{
		Path:    "routes.js",
		Content: expressSample,
		Lang:    model.LangJavaScript,
	})

	t.Logf("Express - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}

	expected := []model.Endpoint{
		{Method: "GET", Path: "/health"},
		{Method: "GET", Path: "/posts"},
		{Method: "GET", Path: "/posts/:id"},
		{Method: "POST", Path: "/posts"},
		{Method: "PUT", Path: "/posts/:id"},
		{Method: "DELETE", Path: "/posts/:id"},
	}

	assertEndpoints(t, "Express", endpoints, expected)
}

func TestParseNestJS(t *testing.T) {
	endpoints := Parse(model.FileContent{
		Path:    "comment.controller.ts",
		Content: nestJSSample,
		Lang:    model.LangTypeScript,
	})

	t.Logf("NestJS - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}

	expected := []model.Endpoint{
		{Method: "GET", Path: "/comments"},
		{Method: "GET", Path: "/comments/:id"},
		{Method: "POST", Path: "/comments"},
		{Method: "PUT", Path: "/comments/:id"},
		{Method: "DELETE", Path: "/comments/:id"},
		{Method: "PATCH", Path: "/comments/:id/approve"},
	}

	assertEndpoints(t, "NestJS", endpoints, expected)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func assertEndpoints(t *testing.T, name string, got []model.Endpoint, want []model.Endpoint) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("[%s] endpoint count: got %d, want %d", name, len(got), len(want))
		return
	}
	for i, w := range want {
		g := got[i]
		if g.Method != w.Method || g.Path != w.Path {
			t.Errorf("[%s] endpoint[%d]: got %s %s, want %s %s", name, i, g.Method, g.Path, w.Method, w.Path)
		}
	}
}

// assertEndpointsUnordered checks same set of endpoints regardless of order.
func assertEndpointsUnordered(t *testing.T, name string, got []model.Endpoint, want []model.Endpoint) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("[%s] endpoint count: got %d, want %d", name, len(got), len(want))
		return
	}
	counts := make(map[string]int)
	for _, e := range want {
		counts[e.Method+":"+e.Path]++
	}
	for _, e := range got {
		key := e.Method + ":" + e.Path
		if counts[key] <= 0 {
			t.Errorf("[%s] unexpected endpoint: %s %s", name, e.Method, e.Path)
		}
		counts[key]--
	}
}
