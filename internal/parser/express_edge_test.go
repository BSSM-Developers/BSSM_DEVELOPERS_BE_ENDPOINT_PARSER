package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// regex paths should NOT be captured (not a string literal)
var expressRegexPaths = []byte(`
const express = require('express');
const router = express.Router();

router.get(/^\/api\/v\d+\/health$/, (req, res) => res.json({ ok: true }));
router.post(/users/, handler);

router.get('/valid', handler);
router.post('/valid', handler);
`)

// app.use() and app.listen() and app.set() should NOT be captured
var expressNonHttpCalls = []byte(`
const express = require('express');
const app = express();

app.set('view engine', 'ejs');
app.set('trust proxy', 1);
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use('/static', express.static('public'));
app.listen(3000, () => console.log('running'));

app.get('/health', handler);
`)

// error-handler (4-param function) and app.use middleware
var expressErrorHandlers = []byte(`
const express = require('express');
const app = express();

app.use((req, res, next) => {
    next();
});

app.use((err, req, res, next) => {
    res.status(500).json({ error: err.message });
});

app.get('/users', listUsers);
app.post('/users', createUser);
`)

// various router/app variable naming
var expressVariousRouterNames = []byte(`
const express = require('express');
const api = express.Router();
const v1 = express.Router();
const authRouter = express.Router();

api.get('/products', listProducts);
api.post('/products', createProduct);

v1.get('/orders', listOrders);
v1.delete('/orders/:id', deleteOrder);

authRouter.post('/login', login);
authRouter.post('/logout', logout);
authRouter.get('/me', getMe);
`)

// chaining off app.route() (not router.route())
var expressAppRouteSyntax = []byte(`
const express = require('express');
const app = express();

app.route('/books')
    .get(getBooks)
    .post(createBook);

app.route('/books/:id')
    .get(getBook)
    .put(updateBook)
    .delete(deleteBook);
`)

// array-of-middleware as second argument
var expressArrayMiddlewareArg = []byte(`
const express = require('express');
const router = express.Router();

const auth = (req, res, next) => next();
const validate = (req, res, next) => next();
const rateLimit = (req, res, next) => next();

router.get('/admin', [auth, validate, rateLimit], getAdmin);
router.post('/admin/users', [auth, validate], createAdminUser);
router.delete('/admin/users/:id', [auth], deleteAdminUser);
`)

// async/await handlers
var expressAsyncHandlers = []byte(`
const express = require('express');
const router = express.Router();

router.get('/async-users', async (req, res) => {
    const users = await User.findAll();
    res.json(users);
});

router.post('/async-users', async (req, res, next) => {
    try {
        const user = await User.create(req.body);
        res.status(201).json(user);
    } catch (err) {
        next(err);
    }
});

router.delete('/async-users/:id', async (req, res) => {
    await User.destroy({ where: { id: req.params.id } });
    res.status(204).end();
});
`)

// deeply nested path params
var expressDeepNestedParams = []byte(`
const express = require('express');
const router = express.Router();

router.get('/orgs/:org/teams/:team/members/:member', getMember);
router.put('/orgs/:org/teams/:team/members/:member', updateMember);
router.delete('/orgs/:org/teams/:team/members/:member', removeMember);

router.get('/v1/namespaces/:ns/pods/:pod/exec', execInPod);
`)

// root path "/"
var expressRootPaths = []byte(`
const express = require('express');
const app = express();
const router = express.Router();

app.get('/', homeHandler);
app.post('/', submitHandler);
router.get('/', listHandler);
`)

func TestExpressRegexPathsNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressRegexPaths, Lang: model.LangJavaScript})
	t.Logf("Express regex paths - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// Only literal string paths captured
	assertEndpointsUnordered(t, "Express/RegexPaths", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/valid"},
		{Method: "POST", Path: "/valid"},
	})
}

func TestExpressNonHttpCallsNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressNonHttpCalls, Lang: model.LangJavaScript})
	t.Logf("Express non-HTTP calls - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "Express/NonHttpCalls", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/health"},
	})
}

func TestExpressErrorHandlersNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressErrorHandlers, Lang: model.LangJavaScript})
	t.Logf("Express error handlers - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "Express/ErrorHandlers", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
	})
}

func TestExpressVariousRouterNames(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressVariousRouterNames, Lang: model.LangJavaScript})
	t.Logf("Express various router names - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "Express/VariousRouterNames", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/products"},
		{Method: "POST", Path: "/products"},
		{Method: "GET", Path: "/orders"},
		{Method: "DELETE", Path: "/orders/:id"},
		{Method: "POST", Path: "/login"},
		{Method: "POST", Path: "/logout"},
		{Method: "GET", Path: "/me"},
	})
}

func TestExpressAppRouteSyntax(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressAppRouteSyntax, Lang: model.LangJavaScript})
	t.Logf("Express app.route() syntax - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "Express/AppRoute", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/books"},
		{Method: "POST", Path: "/books"},
		{Method: "GET", Path: "/books/:id"},
		{Method: "PUT", Path: "/books/:id"},
		{Method: "DELETE", Path: "/books/:id"},
	})
}

func TestExpressArrayMiddlewareArg(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressArrayMiddlewareArg, Lang: model.LangJavaScript})
	t.Logf("Express array middleware arg - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "Express/ArrayMiddleware", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/admin"},
		{Method: "POST", Path: "/admin/users"},
		{Method: "DELETE", Path: "/admin/users/:id"},
	})
}

func TestExpressAsyncHandlers(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressAsyncHandlers, Lang: model.LangJavaScript})
	t.Logf("Express async handlers - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "Express/AsyncHandlers", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/async-users"},
		{Method: "POST", Path: "/async-users"},
		{Method: "DELETE", Path: "/async-users/:id"},
	})
}

func TestExpressDeepNestedParams(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressDeepNestedParams, Lang: model.LangJavaScript})
	t.Logf("Express deep nested params - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "Express/DeepNestedParams", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/orgs/:org/teams/:team/members/:member"},
		{Method: "PUT", Path: "/orgs/:org/teams/:team/members/:member"},
		{Method: "DELETE", Path: "/orgs/:org/teams/:team/members/:member"},
		{Method: "GET", Path: "/v1/namespaces/:ns/pods/:pod/exec"},
	})
}

func TestExpressRootPaths(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: expressRootPaths, Lang: model.LangJavaScript})
	t.Logf("Express root paths - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "Express/RootPaths", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/"},
		{Method: "POST", Path: "/"},
		{Method: "GET", Path: "/"},
	})
}
