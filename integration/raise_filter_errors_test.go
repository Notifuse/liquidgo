package integration

import (
	gocontext "context"
	"errors"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

func renderFilterError(t *testing.T, src string, opts ...liquid.RenderOption) (liquid.Result, error) {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard()})
	tmpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return tmpl.Render(gocontext.Background(), nil, opts...)
}

// With RaiseErrors, a filter error ends the render as Ruby's render! raises it:
// the render returns the error, recorded once, and the output stops before it.
func TestRaiseErrorsEndsTheRenderOnAFilterError(t *testing.T) {
	res, err := renderFilterError(t, "<p>a</p>{{ 10 | divided_by: 0 }}<p>b</p>", liquid.RaiseErrors())
	var liquidErr liquid.LiquidError
	if !errors.As(err, &liquidErr) || !strings.Contains(err.Error(), "divided by 0") {
		t.Fatalf("err = %v; want the Liquid error divided by 0", err)
	}
	if res.Output != "<p>a</p>" {
		t.Errorf("output = %q; want what was written before the error", res.Output)
	}
	if len(res.Errors) != 1 {
		t.Errorf("recorded %d errors; want 1", len(res.Errors))
	}

	t.Run("inside enclosing blocks, recorded once", func(t *testing.T) {
		res, err := renderFilterError(t, "{% if true %}{% for i in (1..3) %}{{ i | divided_by: 0 }}{% endfor %}{% endif %}", liquid.RaiseErrors())
		if err == nil || !strings.Contains(err.Error(), "divided by 0") {
			t.Fatalf("err = %v; want divided by 0", err)
		}
		if len(res.Errors) != 1 {
			t.Errorf("recorded %d errors; want 1", len(res.Errors))
		}
	})

	t.Run("too few arguments", func(t *testing.T) {
		_, err := renderFilterError(t, "{{ 'x' | replace }}", liquid.RaiseErrors())
		if err == nil || !strings.Contains(err.Error(), "wrong number of arguments") {
			t.Fatalf("err = %v; want wrong number of arguments", err)
		}
	})

	t.Run("a Go error of a custom filter", func(t *testing.T) {
		cause := errors.New("feed unavailable")
		fail := liquid.Filter("fail", func(_ *liquid.Context, _ interface{}, _ ...interface{}) (interface{}, error) {
			return nil, cause
		})
		_, err := renderFilterError(t, "{{ 'x' | fail }}", liquid.RaiseErrors(), liquid.WithFilters(fail))
		var internal *liquid.InternalError
		if !errors.As(err, &internal) || !errors.Is(err, cause) {
			t.Fatalf("err = %v; want an *InternalError that unwraps to the filter's error", err)
		}
	})
}

// Without RaiseErrors, a filter error is still recorded and writes nothing.
func TestAFilterErrorWithoutRaiseErrorsIsRecorded(t *testing.T) {
	res, err := renderFilterError(t, "<p>a</p>{{ 10 | divided_by: 0 }}<p>b</p>")
	if err != nil {
		t.Fatalf("err = %v; want nil", err)
	}
	if res.Output != "<p>a</p><p>b</p>" {
		t.Errorf("output = %q; want the filter's output left empty", res.Output)
	}
	if len(res.Errors) != 1 || !strings.Contains(res.Errors[0].Error(), "divided by 0") {
		t.Errorf("errors = %v; want divided by 0 recorded", res.Errors)
	}
}
