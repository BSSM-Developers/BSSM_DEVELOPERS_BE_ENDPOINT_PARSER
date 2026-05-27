package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// typed path params: {id:int}, {name:str}
var fastapiTypedPathParams = []byte(`
from fastapi import APIRouter

router = APIRouter()

@router.get("/users/{user_id:int}")
async def get_user(user_id: int): return {}

@router.get("/users/{user_id:int}/posts/{post_id:int}")
async def get_post(user_id: int, post_id: int): return {}

@router.delete("/files/{file_path:path}")
async def delete_file(file_path: str): pass

@router.get("/items/{name:str}/tags/{tag:str}")
async def get_item_tag(name: str, tag: str): return {}
`)

// variable or expression paths should NOT be captured
var fastapiVariableAndExprPaths = []byte(`
from fastapi import FastAPI, APIRouter

app = FastAPI()
router = APIRouter()

USERS_PATH = "/users"
BASE = "/api"

@app.get(USERS_PATH)
async def users_by_var(): return []

@router.get(BASE + "/items")
async def items_by_concat(): return []

@router.get(f"/dynamic/{some_var}")
async def dynamic(): return {}

@app.get("/valid")
async def valid(): return {}
`)

// non-HTTP decorators must not be captured
var fastapiNonHTTPDecoratorsExtended = []byte(`
from fastapi import FastAPI
from functools import lru_cache
import pytest

app = FastAPI()

@lru_cache(maxsize=128)
def get_config(): return {}

@pytest.fixture
def client(): return TestClient(app)

@staticmethod
def helper(): pass

@classmethod
def factory(cls): return cls()

@property
def value(self): return self._val

@app.get("/ok")
async def valid_route(): return {}
`)

// dependency injection and complex route kwargs
var fastapiComplexRouteKwargs = []byte(`
from fastapi import APIRouter, Depends, Security
from fastapi.security import HTTPBearer

router = APIRouter()

@router.get(
    "/users",
    response_model=list[UserResponse],
    summary="List users",
    description="Returns paginated list of users",
    tags=["users"],
    dependencies=[Depends(verify_token)],
    status_code=200,
)
async def list_users(): return []

@router.post(
    "/users",
    response_model=UserResponse,
    status_code=201,
    dependencies=[Security(HTTPBearer())],
)
async def create_user(): return {}
`)

// multiple decorators on same endpoint (stacked routes)
var fastapiStackedDecorators = []byte(`
from fastapi import FastAPI
app = FastAPI()

@app.get("/")
@app.get("/home")
@app.get("/index")
async def home(): return {"page": "home"}

@app.post("/submit")
@app.put("/submit")
async def submit(): return {}
`)

// generator/streaming endpoints
var fastapiStreamingEndpoints = []byte(`
from fastapi import FastAPI
from fastapi.responses import StreamingResponse

app = FastAPI()

@app.get("/stream")
async def stream_data():
    async def generator():
        yield b"data"
    return StreamingResponse(generator())

@app.get("/events")
async def sse_events():
    return EventSourceResponse(event_generator())
`)

// validator and model decorators should NOT be captured
var fastapiValidatorDecorators = []byte(`
from pydantic import BaseModel, validator, root_validator
from fastapi import FastAPI

app = FastAPI()

class User(BaseModel):
    name: str
    age: int

    @validator("name")
    def name_not_empty(cls, v): return v

    @root_validator
    def check_age(cls, values): return values

@app.get("/users")
async def list_users(): return []
`)

func TestFastAPITypedPathParams(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiTypedPathParams, Lang: model.LangPython})
	t.Logf("FastAPI typed path params - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "FastAPI/TypedPathParams", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users/{user_id:int}"},
		{Method: "GET", Path: "/users/{user_id:int}/posts/{post_id:int}"},
		{Method: "DELETE", Path: "/files/{file_path:path}"},
		{Method: "GET", Path: "/items/{name:str}/tags/{tag:str}"},
	})
}

func TestFastAPIVariablePathNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiVariableAndExprPaths, Lang: model.LangPython})
	t.Logf("FastAPI variable/expr paths - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// Only the literal string "/valid" should be captured
	assertEndpointsUnordered(t, "FastAPI/VariablePath", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/valid"},
	})
}

func TestFastAPINonHTTPDecoratorsExtended(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiNonHTTPDecoratorsExtended, Lang: model.LangPython})
	t.Logf("FastAPI non-HTTP decorators (extended) - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "FastAPI/NonHTTPExtended", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/ok"},
	})
}

func TestFastAPIComplexRouteKwargs(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiComplexRouteKwargs, Lang: model.LangPython})
	t.Logf("FastAPI complex kwargs - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "FastAPI/ComplexKwargs", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
	})
}

func TestFastAPIStackedDecorators(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiStackedDecorators, Lang: model.LangPython})
	t.Logf("FastAPI stacked decorators - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "FastAPI/StackedDecorators", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/"},
		{Method: "GET", Path: "/home"},
		{Method: "GET", Path: "/index"},
		{Method: "POST", Path: "/submit"},
		{Method: "PUT", Path: "/submit"},
	})
}

func TestFastAPIStreamingEndpoints(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiStreamingEndpoints, Lang: model.LangPython})
	t.Logf("FastAPI streaming - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "FastAPI/Streaming", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/stream"},
		{Method: "GET", Path: "/events"},
	})
}

func TestFastAPIValidatorDecoratorsNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: fastapiValidatorDecorators, Lang: model.LangPython})
	t.Logf("FastAPI validator decorators - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "FastAPI/ValidatorDecorators", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users"},
	})
}
