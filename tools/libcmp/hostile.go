package main

import (
	gocontext "context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
	nchapman "github.com/nchapman/go-liquid"
)

// hostileTemplates are templates a hostile author writes to exhaust the
// process that renders them.
var hostileTemplates = []struct{ name, source string }{
	{"a billion iterations", `{% for i in (1..1000000000) %}{% endfor %}done`},
	{"a string doubled 40 times", `{% assign s = "xxxxxxxxxx" %}{% for i in (1..40) %}{% assign s = s | append: s %}{% endfor %}{{ s | size }}`},
	{"a billion bytes of output", `{% for a in (1..1000) %}{% for b in (1..1000) %}{% for c in (1..1000) %}x{% endfor %}{% endfor %}{% endfor %}`},
	{"100,000 nested blocks", strings.Repeat("{% if true %}", 100000) + "x" + strings.Repeat("{% endif %}", 100000)},
}

// limits are the bounds given to the libraries that take them, the same for
// each: 1 MiB of output, a render score of 100,000 and 1 MiB assigned.
const (
	limitLength = 1 << 20
	limitScore  = 100000
	limitAssign = 1 << 20
)

// limited returns a render of src with the resource limits of the library,
// for the libraries that have them.
func limited(name string) func(src string) (func(map[string]any) (string, error), error) {
	switch name {
	case "liquidgo":
		return func(src string) (func(map[string]any) (string, error), error) {
			env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard()})
			if err != nil {
				return nil, err
			}
			length, score, assign := limitLength, limitScore, limitAssign
			t, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
			if err != nil {
				return nil, err
			}
			return func(data map[string]any) (string, error) {
				res, err := t.Render(gocontext.Background(), data, liquid.WithLimits(liquid.ResourceLimitsConfig{
					RenderLengthLimit: &length, RenderScoreLimit: &score, AssignScoreLimit: &assign,
				}))
				return res.Output, err
			}, nil
		}
	case "nchapman":
		return func(src string) (func(map[string]any) (string, error), error) {
			t, err := nchapman.NewEnvironment().WithErrorMode(nchapman.ErrorModeLax).Parse(src)
			if err != nil {
				return nil, err
			}
			return func(data map[string]any) (string, error) {
				return t.Render(data, nchapman.WithLimits(&nchapman.ResourceLimits{
					RenderLengthLimit: limitLength, RenderScoreLimit: limitScore, AssignScoreLimit: limitAssign,
				}))
			}, nil
		}
	}
	return nil
}

// hostile renders hostile template i with lib, with its limits if it has
// any, and prints the outcome; the caller bounds the time and the memory.
func hostile(lib library, arg string) error {
	i, err := strconv.Atoi(arg)
	if err != nil || i < 0 || i >= len(hostileTemplates) {
		return fmt.Errorf("hostile: no template %q", arg)
	}
	parse := lib.parse
	if l := limited(lib.name); l != nil {
		parse = l
	}
	fn, err := parse(hostileTemplates[i].source)
	if err != nil {
		fmt.Printf("refused: %.80s\n", err)
		return nil
	}
	out, err := fn(map[string]any{})
	if err != nil {
		fmt.Printf("stopped: %.80s\n", err)
		return nil
	}
	fmt.Printf("rendered: %d bytes\n", len(out))
	return nil
}
