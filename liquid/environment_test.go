package liquid

import (
	"testing"
)

func TestNewEnvironment(t *testing.T) {
	env := mustEnv(EnvironmentConfig{})
	if env == nil {
		t.Fatal("Expected Environment, got nil")
	}
	if env.ErrorMode() != "lax" {
		t.Errorf("Expected error mode 'lax', got '%s'", env.ErrorMode())
	}
}

// namedTag returns a constructor of a silent tag named name, whatever name
// the template gives: the tests tell constructors apart by what they build.
func namedTag(name string) TagConstructor {
	return func(_, markup string, pc ParseContextInterface) (Node, error) {
		return &silentTag{NewTag(name, markup, pc)}, nil
	}
}

// builtName returns the name of the tag that constructor builds.
func builtName(t *testing.T, constructor TagConstructor) string {
	t.Helper()
	if constructor == nil {
		t.Fatal("Expected a constructor, got nil")
	}
	node, err := constructor("", "", NewParseContext(ParseContextOptions{}))
	if err != nil {
		t.Fatalf("constructor: %v", err)
	}
	return node.(*silentTag).TagName()
}

func TestEnvironmentTag(t *testing.T) {
	envTags := map[string]TagConstructor{}
	envTags["test"] = namedTag("TestTag")
	env := mustEnv(EnvironmentConfig{Tags: envTags})

	if got := builtName(t, env.TagForName("test")); got != "TestTag" {
		t.Errorf("Expected 'TestTag', got %v", got)
	}
}

func TestEnvironmentFilters(t *testing.T) {
	env := mustEnv(EnvironmentConfig{Filters: []FilterDef{Filter("shout", rawFilter), Filter("one", rawFilter)}})

	names := env.FilterNames()
	if !containsName(names, "shout") || !containsName(names, "one") || !containsName(names, "upcase") {
		t.Errorf("Expected the filters and the standard ones, got %v", names)
	}
}

func TestEnvironmentFileSystem(t *testing.T) {
	fs := &BlankFileSystem{}
	env := mustEnv(EnvironmentConfig{FileSystem: fs})

	if env.FileSystem() != fs {
		t.Error("FileSystem mismatch")
	}
}

func TestEnvironmentErrorMode(t *testing.T) {
	env := mustEnv(EnvironmentConfig{ErrorMode: "strict"})
	if env.ErrorMode() != "strict" {
		t.Errorf("Expected error mode 'strict', got '%s'", env.ErrorMode())
	}
}

func TestEnvironmentTags(t *testing.T) {
	env := mustEnv(EnvironmentConfig{Tags: map[string]TagConstructor{"test1": namedTag("Tag1"), "test2": namedTag("Tag2")}})

	if got := builtName(t, env.TagForName("test1")); got != "Tag1" {
		t.Errorf("Expected 'Tag1' for test1, got %v", got)
	}
	if got := builtName(t, env.TagForName("test2")); got != "Tag2" {
		t.Errorf("Expected 'Tag2' for test2, got %v", got)
	}
	if env.TagForName("test3") != nil {
		t.Error("Expected no constructor for an unknown tag")
	}
}

// The environment keeps a copy of its config: what the caller changes after,
// in the config, its tag map or its limits, changes nothing.
func TestTheEnvironmentKeepsACopyOfItsConfig(t *testing.T) {
	limit := 10
	config := EnvironmentConfig{
		Tags:   map[string]TagConstructor{"a": namedTag("A")},
		Limits: ResourceLimitsConfig{RenderScoreLimit: &limit},
	}
	env := mustEnv(config)
	config.Tags["a"] = namedTag("B")
	config.Tags["b"] = namedTag("B")
	limit = 99
	config.ErrorMode = "strict"

	if got := builtName(t, env.TagForName("a")); got != "A" {
		t.Errorf("tag a = %v, want A", got)
	}
	if env.TagForName("b") != nil {
		t.Error("a tag added to the map after is in the environment")
	}
	if got := NewTemplate(&TemplateOptions{Environment: env}).limits.RenderScoreLimit; got == nil || *got != 10 {
		t.Errorf("render score limit = %v, want 10", got)
	}
	if env.ErrorMode() != "lax" {
		t.Errorf("error mode = %q, want lax", env.ErrorMode())
	}
}

func TestNewEnvironmentRefusesAnInvalidConfig(t *testing.T) {
	for name, config := range map[string]EnvironmentConfig{
		"unknown error mode":        {ErrorMode: "loose"},
		"tag without a name":        {Tags: map[string]TagConstructor{"": namedTag("A")}},
		"tag without constructor":   {Tags: map[string]TagConstructor{"a": nil}},
		"filter without a name":     {Filters: []FilterDef{Filter("", rawFilter)}},
		"filter without function":   {Filters: []FilterDef{Filter("x", nil)}},
		"negative partial partials": {PartialCache: &PartialCacheLimits{Partials: -1}},
		"negative partial bytes":    {PartialCache: &PartialCacheLimits{Bytes: -1}},
	} {
		if env, err := NewEnvironment(config); err == nil || env != nil {
			t.Errorf("%s: NewEnvironment = %v, %v; want an error", name, env, err)
		}
	}
	for _, mode := range []ErrorMode{"", ErrorModeLax, ErrorModeStrict, ErrorModeStrict2, ErrorModeRigid, ErrorModeWarn} {
		if _, err := NewEnvironment(EnvironmentConfig{ErrorMode: mode}); err != nil {
			t.Errorf("error mode %q: %v", mode, err)
		}
	}
}

// OutputEscape adds the raw filter, which a filter of the config replaces.
func TestOutputEscapeAddsRaw(t *testing.T) {
	if containsName(mustEnv(EnvironmentConfig{}).FilterNames(), "raw") {
		t.Error("raw without the output escape")
	}
	env := mustEnv(EnvironmentConfig{OutputEscape: true})
	if !env.OutputEscape() || !containsName(env.FilterNames(), "raw") {
		t.Errorf("output escape %v, filters %v", env.OutputEscape(), env.FilterNames())
	}
	shout := func(_ *Context, input interface{}, _ ...interface{}) (interface{}, error) { return "RAW", nil }
	env = mustEnv(EnvironmentConfig{OutputEscape: true, Filters: []FilterDef{Filter("raw", shout)}})
	if f, _ := (&Context{environment: env}).resolveFilter("raw"); f == nil {
		t.Fatal("no raw filter")
	} else if got, _ := f(nil, "x"); got != "RAW" {
		t.Errorf("raw = %v, want the config's filter", got)
	}
}
