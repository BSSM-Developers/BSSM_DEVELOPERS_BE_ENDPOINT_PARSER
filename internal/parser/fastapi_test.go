package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// extra kwargs in decorator
var fastapiExtraKwargs = []byte(`
from fastapi import FastAPI
app = FastAPI()

@app.get("/users", response_model=list[UserResponse], tags=["users"], deprecated=False)
async def get_users():
    return []

@app.post("/users", status_code=201, response_model=UserResponse)
async def create_user(req: UserCreate):
    return {}

@app.delete("/users/{user_id}", status_code=204)
async def delete_user(user_id: int):
    pass
`)

// path= keyword argument
var fastapiPathKwarg = []byte(`
from fastapi import APIRouter
router = APIRouter()

@router.get(path="/users/{user_id}", response_model=UserResponse)
async def get_user(user_id: int):
    return {}

@router.post(path="/users", status_code=201)
async def create_user(req: UserCreate):
    return {}
`)

// Flask-style @app.route with methods list
var flaskStyleRoutes = []byte(`
from flask import Flask
app = Flask(__name__)

@app.route("/items", methods=["GET"])
def list_items():
    return []

@app.route("/items", methods=["POST"])
def create_item():
    return {}, 201

@app.route("/items/<int:item_id>", methods=["PUT"])
def update_item(item_id):
    return {}

@app.route("/items/<int:item_id>", methods=["DELETE"])
def delete_item(item_id):
    pass
`)

// nested path params
var fastapiNestedParams = []byte(`
from fastapi import APIRouter
router = APIRouter()

@router.get("/users/{user_id}/posts/{post_id}/comments/{comment_id}")
async def get_comment(user_id: int, post_id: int, comment_id: int):
    return {}

@router.delete("/orgs/{org}/repos/{repo}/issues/{number}")
async def delete_issue(org: str, repo: str, number: int):
    pass
`)

// root path and multiple methods on same path
var fastapiRootAndMulti = []byte(`
from fastapi import FastAPI
app = FastAPI()

@app.get("/")
async def root():
    return {"message": "ok"}

@app.get("/health")
@app.head("/health")
async def health():
    return {}
`)

// non-HTTP decorators should not be captured
var fastapiNonHTTP = []byte(`
from fastapi import FastAPI, Depends
app = FastAPI()

def require_auth():
    pass

@app.middleware("http")
async def add_cors(request, call_next):
    return await call_next(request)

@require_auth
@app.get("/protected")
async def protected():
    return {}

@app.get("/public")
async def public():
    return {}
`)

// multiple routers in same file
var fastapiMultiRouter = []byte(`
from fastapi import APIRouter

users_router = APIRouter(prefix="/users")
posts_router = APIRouter(prefix="/posts")

@users_router.get("/")
async def list_users():
    return []

@users_router.post("/")
async def create_user():
    return {}

@posts_router.get("/")
async def list_posts():
    return []

@posts_router.get("/{post_id}")
async def get_post(post_id: int):
    return {}

@posts_router.delete("/{post_id}")
async def delete_post(post_id: int):
    pass
`)

func TestFastAPIExtraKwargs(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiExtraKwargs, Lang: model.LangPython})
	t.Logf("FastAPI extra kwargs - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "FastAPI/ExtraKwargs", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
		{Method: "DELETE", Path: "/users/{user_id}"},
	})
}

func TestFastAPIPathKwarg(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiPathKwarg, Lang: model.LangPython})
	t.Logf("FastAPI path= kwarg - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "FastAPI/PathKwarg", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users/{user_id}"},
		{Method: "POST", Path: "/users"},
	})
}

func TestFlaskStyleRoutes(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: flaskStyleRoutes, Lang: model.LangPython})
	t.Logf("Flask-style routes - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "Flask/MethodsKwarg", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/items"},
		{Method: "POST", Path: "/items"},
		{Method: "PUT", Path: "/items/<int:item_id>"},
		{Method: "DELETE", Path: "/items/<int:item_id>"},
	})
}

func TestFastAPINestedParams(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiNestedParams, Lang: model.LangPython})
	t.Logf("FastAPI nested params - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "FastAPI/NestedParams", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users/{user_id}/posts/{post_id}/comments/{comment_id}"},
		{Method: "DELETE", Path: "/orgs/{org}/repos/{repo}/issues/{number}"},
	})
}

func TestFastAPIRootAndMultiDecorator(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiRootAndMulti, Lang: model.LangPython})
	t.Logf("FastAPI root & multi-decorator - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// @app.head is not in our HTTP methods, only GET /health captured alongside GET /
	assertEndpoints(t, "FastAPI/RootAndMulti", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/"},
		{Method: "GET", Path: "/health"},
	})
}

func TestFastAPINonHTTPDecorators(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiNonHTTP, Lang: model.LangPython})
	t.Logf("FastAPI non-HTTP decorators - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// middleware and require_auth should not be captured
	assertEndpoints(t, "FastAPI/NonHTTP", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/protected"},
		{Method: "GET", Path: "/public"},
	})
}

func TestFastAPIMultiRouter(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiMultiRouter, Lang: model.LangPython})
	t.Logf("FastAPI multiple routers - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "FastAPI/MultiRouter", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/"},
		{Method: "POST", Path: "/"},
		{Method: "GET", Path: "/"},
		{Method: "GET", Path: "/{post_id}"},
		{Method: "DELETE", Path: "/{post_id}"},
	})
}
