package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// multiple middleware handlers - path is still first arg
var expressMultiMiddleware = []byte(`
const express = require('express');
const router = express.Router();

router.get('/users', authMiddleware, rateLimiter, getAllUsers);
router.post('/users', authMiddleware, validateBody, createUser);
router.put('/users/:id', authMiddleware, validateBody, updateUser);
router.delete('/users/:id', authMiddleware, deleteUser);
`)

// chained router.route()
var expressChainedRoute = []byte(`
const express = require('express');
const router = express.Router();

router.route('/articles')
    .get(listArticles)
    .post(createArticle);

router.route('/articles/:id')
    .get(getArticle)
    .put(updateArticle)
    .delete(deleteArticle)
    .patch(patchArticle);
`)

// template literal paths
var expressTemplateLiteral = []byte(`
const express = require('express');
const router = express.Router();
const BASE = '/api/v1';

router.get(` + "`" + `/posts` + "`" + `, handler);
router.post(` + "`" + `/posts` + "`" + `, handler);
router.get(` + "`" + `/posts/:id` + "`" + `, handler);
`)

// use() should not be captured, all() not supported (intentional)
var expressNonRoute = []byte(`
const express = require('express');
const app = express();
const router = express.Router();

app.use('/api', router);
app.use(express.json());
router.use('/prefix', subRouter);

app.get('/health', (req, res) => res.json({ status: 'ok' }));
`)

// nested path params
var expressNestedParams = []byte(`
const express = require('express');
const router = express.Router();

router.get('/orgs/:orgId/repos/:repoId/issues/:issueId', getIssue);
router.put('/orgs/:orgId/repos/:repoId/issues/:issueId', updateIssue);
router.get('/users/:userId/posts/:postId/comments', listComments);
`)

// arrow function inline handler
var expressInlineHandler = []byte(`
const express = require('express');
const app = express();

app.get('/ping', (req, res) => {
    res.json({ pong: true });
});

app.post('/echo', async (req, res) => {
    res.json(req.body);
});

app.get('/redirect', (req, res) => res.redirect('/ping'));
`)

// mixed app and router
var expressMixedAppRouter = []byte(`
const express = require('express');
const app = express();
const v1 = express.Router();
const v2 = express.Router();

app.get('/health', healthHandler);

v1.get('/users', listUsersV1);
v1.post('/users', createUserV1);

v2.get('/users', listUsersV2);
v2.post('/users', createUserV2);
v2.delete('/users/:id', deleteUserV2);
`)

func TestExpressMultiMiddleware(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressMultiMiddleware, Lang: model.LangJavaScript})
	t.Logf("Express multi-middleware - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "Express/MultiMiddleware", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
		{Method: "PUT", Path: "/users/:id"},
		{Method: "DELETE", Path: "/users/:id"},
	})
}

func TestExpressChainedRoute(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressChainedRoute, Lang: model.LangJavaScript})
	t.Logf("Express chained route() - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// AST traversal visits outermost call first, so order depends on chain depth — use unordered comparison
	assertEndpointsUnordered(t, "Express/ChainedRoute", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/articles"},
		{Method: "POST", Path: "/articles"},
		{Method: "GET", Path: "/articles/:id"},
		{Method: "PUT", Path: "/articles/:id"},
		{Method: "DELETE", Path: "/articles/:id"},
		{Method: "PATCH", Path: "/articles/:id"},
	})
}

func TestExpressTemplateLiteral(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressTemplateLiteral, Lang: model.LangJavaScript})
	t.Logf("Express template literal - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "Express/TemplateLiteral", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/posts"},
		{Method: "POST", Path: "/posts"},
		{Method: "GET", Path: "/posts/:id"},
	})
}

func TestExpressNonRouteShouldNotCapture(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressNonRoute, Lang: model.LangJavaScript})
	t.Logf("Express non-route - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// app.use() should not be captured
	assertEndpoints(t, "Express/NonRoute", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/health"},
	})
}

func TestExpressNestedParams(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressNestedParams, Lang: model.LangJavaScript})
	t.Logf("Express nested params - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "Express/NestedParams", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/orgs/:orgId/repos/:repoId/issues/:issueId"},
		{Method: "PUT", Path: "/orgs/:orgId/repos/:repoId/issues/:issueId"},
		{Method: "GET", Path: "/users/:userId/posts/:postId/comments"},
	})
}

func TestExpressInlineHandler(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressInlineHandler, Lang: model.LangJavaScript})
	t.Logf("Express inline handler - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "Express/InlineHandler", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/ping"},
		{Method: "POST", Path: "/echo"},
		{Method: "GET", Path: "/redirect"},
	})
}

func TestExpressMixedAppRouter(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressMixedAppRouter, Lang: model.LangJavaScript})
	t.Logf("Express mixed app+router - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpoints(t, "Express/MixedAppRouter", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/health"},
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
		{Method: "DELETE", Path: "/users/:id"},
	})
}
