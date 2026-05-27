package parser

import (
	"testing"

	"endpoint-parser/internal/model"
)

// @RequestMapping(method = {GET, POST}) → 2 endpoints
var springMultipleMethodsArray = []byte(`
@RestController
public class ItemController {

    @RequestMapping(value = "/items", method = {RequestMethod.GET, RequestMethod.POST})
    public Object listOrCreate() { return null; }

    @RequestMapping(value = "/items/{id}", method = {RequestMethod.PUT, RequestMethod.PATCH})
    public Object updateOrPatch() { return null; }
}
`)

// @RequestMapping(method={...}) + class prefix
var springMultipleMethodsWithPrefix = []byte(`
@RestController
@RequestMapping("/api/orders")
public class OrderController {

    @RequestMapping(method = {RequestMethod.GET, RequestMethod.POST})
    public Object listOrCreate() { return null; }

    @RequestMapping(value = "/{id}", method = {RequestMethod.PUT, RequestMethod.DELETE})
    public Object updateOrDelete() { return null; }
}
`)

// non-routing annotations must NOT be captured
var springNonRoutingAnnotations = []byte(`
@Service
@Transactional
public class UserService {

    @Autowired
    private UserRepository userRepository;

    @Bean
    public UserMapper userMapper() { return new UserMapperImpl(); }

    @Override
    public User findById(Long id) { return userRepository.findById(id).orElseThrow(); }
}

@Component
public class EventListener {

    @EventListener
    public void onStartup(ApplicationReadyEvent event) {}

    @Scheduled(cron = "0 * * * * *")
    public void cleanUp() {}
}
`)

// annotation with only consumes/produces and no path — maps to class base
var springProducesConsumesSample = []byte(`
@RestController
@RequestMapping("/api/documents")
public class DocumentController {

    @PostMapping(consumes = "application/json")
    public Document upload(@RequestBody DocumentRequest req) { return null; }

    @GetMapping(produces = "application/pdf")
    public byte[] download() { return null; }

    @PutMapping(value = "/{id}", consumes = "application/json", produces = "application/json")
    public Document update(@PathVariable Long id, @RequestBody DocumentRequest req) { return null; }
}
`)

// interface-based controller (Feign client + actual controller share interface)
var springInterfaceController = []byte(`
public interface UserApi {

    @GetMapping("/users")
    List<User> listUsers();

    @GetMapping("/users/{id}")
    User getUser(@PathVariable Long id);

    @PostMapping("/users")
    User createUser(@RequestBody UserRequest req);
}

@RestController
public class UserController implements UserApi {

    @Override
    public List<User> listUsers() { return null; }

    @Override
    public User getUser(Long id) { return null; }

    @Override
    public User createUser(UserRequest req) { return null; }
}
`)

// abstract base controller with shared prefix
var springAbstractController = []byte(`
@RequestMapping("/api/v1")
public abstract class BaseApiController {

    @GetMapping("/health")
    public String health() { return "ok"; }
}

@RestController
public class ConcreteController extends BaseApiController {

    @GetMapping("/users")
    public List<User> users() { return null; }

    @PostMapping("/users")
    public User createUser() { return null; }
}
`)

// nested class should NOT inherit outer class prefix
var springNestedClass = []byte(`
@RestController
@RequestMapping("/api/outer")
public class OuterController {

    @GetMapping("/route")
    public String outerRoute() { return "outer"; }

    public static class InnerHelper {

        @GetMapping("/inner-ignored")
        public String innerRoute() { return "inner"; }
    }
}
`)

// @PathVariable, @RequestParam etc. should NEVER be captured
var springMethodParamAnnotations = []byte(`
@RestController
@RequestMapping("/users")
public class UserController {

    @GetMapping("/{id}")
    public User getUser(
        @PathVariable("id") Long userId,
        @RequestParam(value = "fields", required = false) String fields,
        @RequestHeader("Authorization") String auth,
        @RequestBody(required = false) UserFilter filter
    ) {
        return null;
    }
}
`)

func TestSpringBootMultipleMethodsArray(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springMultipleMethodsArray, Lang: model.LangJava})
	t.Logf("Spring Boot method array - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/MultiMethodArray", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/items"},
		{Method: "POST", Path: "/items"},
		{Method: "PUT", Path: "/items/{id}"},
		{Method: "PATCH", Path: "/items/{id}"},
	})
}

func TestSpringBootMultipleMethodsWithPrefix(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springMultipleMethodsWithPrefix, Lang: model.LangJava})
	t.Logf("Spring Boot method array + prefix - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/MultiMethodWithPrefix", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/api/orders"},
		{Method: "POST", Path: "/api/orders"},
		{Method: "PUT", Path: "/api/orders/{id}"},
		{Method: "DELETE", Path: "/api/orders/{id}"},
	})
}

func TestSpringBootNonRoutingAnnotationsNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springNonRoutingAnnotations, Lang: model.LangJava})
	t.Logf("Spring Boot non-routing - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	if len(endpoints) != 0 {
		t.Errorf("[SpringBoot/NonRouting] expected 0 endpoints, got %d", len(endpoints))
	}
}

func TestSpringBootProducesConsumesOnly(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springProducesConsumesSample, Lang: model.LangJava})
	t.Logf("Spring Boot produces/consumes - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/ProducesConsumes", endpoints, []model.Endpoint{
		{Method: "POST", Path: "/api/documents"},
		{Method: "GET", Path: "/api/documents"},
		{Method: "PUT", Path: "/api/documents/{id}"},
	})
}

func TestSpringBootInterfaceController(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springInterfaceController, Lang: model.LangJava})
	t.Logf("Spring Boot interface controller - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// Interface annotations are captured; @Override methods have no annotations → only 3 from interface
	assertEndpointsUnordered(t, "SpringBoot/InterfaceController", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users"},
		{Method: "GET", Path: "/users/{id}"},
		{Method: "POST", Path: "/users"},
	})
}

func TestSpringBootAbstractController(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springAbstractController, Lang: model.LangJava})
	t.Logf("Spring Boot abstract controller - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/AbstractController", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/api/v1/health"},
		{Method: "GET", Path: "/users"},
		{Method: "POST", Path: "/users"},
	})
}

func TestSpringBootNestedClassDoesNotInheritPrefix(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springNestedClass, Lang: model.LangJava})
	t.Logf("Spring Boot nested class - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	// InnerHelper is a nested static class — its @GetMapping should use its OWN enclosing class
	// (InnerHelper has no @RequestMapping, so inner route has no prefix)
	assertEndpointsUnordered(t, "SpringBoot/NestedClass", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/api/outer/route"},
		{Method: "GET", Path: "/inner-ignored"}, // inner class has no prefix of its own
	})
}

func TestSpringBootMethodParamAnnotationsNotCaptured(t *testing.T) {
	endpoints := Parse(model.FileContent{Content: springMethodParamAnnotations, Lang: model.LangJava})
	t.Logf("Spring Boot method param annotations - parsed %d endpoints:", len(endpoints))
	for _, e := range endpoints {
		t.Logf("  %-7s %s", e.Method, e.Path)
	}
	assertEndpointsUnordered(t, "SpringBoot/MethodParamAnnotations", endpoints, []model.Endpoint{
		{Method: "GET", Path: "/users/{id}"},
	})
}
