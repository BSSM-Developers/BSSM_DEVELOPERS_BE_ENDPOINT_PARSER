package fetcher

import "testing"

func TestIsRouteFile(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		// alwaysIncludeNames
		{"app/main.py", true},
		{"app.py", true},
		{"server.py", true},
		{"src/app.js", true},
		{"src/server.js", true},
		// routeNamePatterns in filename
		{"src/userController.java", true},
		{"routes/index.js", true},
		{"api/userHandler.go", true}, // dir match; lang filtering is caller's job (detectLang)
		{"src/router.ts", true},
		// routeDirPatterns — parent directory match
		{"app/api/book.py", true},
		{"app/api/paraphrase.py", true},
		{"app/apis/user.py", true},
		{"app/controllers/user.py", true},
		{"app/endpoints/health.py", true},
		{"app/views/index.py", true},
		{"app/handlers/auth.py", true},
		// no match
		{"app/services/chunking.py", false},
		{"app/db/chroma.py", false},
		{"app/config.py", false},
		{"app/models/user.py", false},
	}
	for _, c := range cases {
		got := isRouteFile(c.path)
		if got != c.want {
			t.Errorf("isRouteFile(%q) = %v, want %v", c.path, got, c.want)
		}
	}
}
