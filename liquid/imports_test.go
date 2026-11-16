package liquid

import (
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The packages of the library reach no network, no process and no plugin,
// so that a template cannot either; they read files only through the
// FileSystem a host gives (file_system.go) and a locale file a host names
// (i18n.go).
var (
	forbiddenImports = []string{"os/exec", "net", "net/http", "net/rpc", "net/smtp", "net/mail", "plugin", "os/signal", "os/user"}
	osAllowedIn      = map[string]bool{"liquid/file_system.go": true, "liquid/i18n.go": true}
)

func TestTheLibraryImportsNoSystemAccess(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"tags", "tag"} {
		more, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		files = append(files, more...)
	}
	fset := token.NewFileSet()
	checked := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		checked++
		name := filepath.ToSlash(filepath.Join("liquid", file))
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			for _, forbidden := range forbiddenImports {
				if path == forbidden {
					t.Errorf("%s imports %s", name, path)
				}
			}
			if path == "os" && !osAllowedIn[name] {
				t.Errorf("%s imports os: only %v may read files", name, osAllowedIn)
			}
		}
	}
	if checked < 50 {
		t.Fatalf("checked %d files, want the whole library", checked)
	}

	// and none of the packages they depend on
	goTool, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go tool to list the dependencies")
	}
	out, err := exec.Command(goTool, "list", "-deps", "./...").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	for _, dep := range strings.Fields(string(out)) {
		for _, forbidden := range forbiddenImports {
			if dep == forbidden {
				t.Errorf("the library depends on %s", dep)
			}
		}
	}
}
