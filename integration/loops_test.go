package integration

import (
	gocontext "context"
	"errors"
	"testing"
	"time"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// TestLoopsSliceAsRuby checks offset, limit and reversed on a range, which
// is iterated rather than built, and on an array: the results are Ruby's,
// and an offset plus a limit past the int range saturates.
func TestLoopsSliceAsRuby(t *testing.T) {
	data := map[string]interface{}{"arr": []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	for _, c := range []struct{ source, want string }{
		{"{% for i in (1..10) offset: 2 limit: 9223372036854775807 %}{{ i }}{% endfor %}", "345678910"},
		{"{% for i in arr offset: 2 limit: 9223372036854775807 %}{{ i }}{% endfor %}", "345678910"},
		{"{% for i in (1..10) reversed offset: 2 limit: 3 %}{{ i }}{% endfor %}", "543"},
		{"{% for i in arr reversed offset: 2 limit: 3 %}{{ i }}{% endfor %}", "543"},
		{"{% for i in (1..10) limit: 3 %}{{ i }}{% endfor %}|{% for i in (1..10) offset: continue limit: 3 %}{{ i }}{% endfor %}|{% for i in (1..10) offset: continue %}{{ i }}{% endfor %}", "123|456|78910"},
		{"{% for i in (-3..-1) offset: -5 %}{{ i }}{% endfor %}", "-3-2-1"},
		{"{% for i in (5..1) %}{{ i }}{% else %}none{% endfor %}", "none"},
		{"{% for i in (1..10) %}{{ forloop.length }}{% break %}{% endfor %}", "10"},
		{"{% tablerow i in (1..10) offset: 8 %}{{ i }}{% endtablerow %}", "<tr class=\"row1\">\n<td class=\"col1\">9</td><td class=\"col2\">10</td></tr>\n"},
		{"{% for i in (-9223372036854775807..9223372036854775807) limit: 2 %}{{ forloop.length }}{% endfor %}", "22"},
	} {
		if got := renderWith(newEnv("", limits{}), c.source, data); got.output != c.want || got.parseErr != nil {
			t.Errorf("%q renders %q (%v), want %q", c.source, got.output, got.parseErr, c.want)
		}
	}
}

// TestRenderStopsWhenItsContextIsDone checks that a render that would run
// for hours stops soon after the deadline of its context.Context, in a loop,
// a capture and a partial, and that the error wraps the context's.
func TestRenderStopsWhenItsContextIsDone(t *testing.T) {
	config := envConfig("", limits{})
	loops := "{% for i in (1..100000) %}{% for j in (1..100000) %}x{% endfor %}{% endfor %}"
	config.FileSystem = NewStubFileSystem(map[string]string{"p": loops})
	env := mustEnv(config)
	const deadline = 50 * time.Millisecond
	slack := 50 * time.Millisecond
	if raceEnabled {
		slack = 200 * time.Millisecond
	}
	for _, source := range []string{
		loops,
		"{% capture c %}" + loops + "{% endcapture %}",
		"{% render 'p' %}",
		"{% include 'p' %}",
		"{% tablerow i in (1..100000) %}{% tablerow j in (1..100000) %}{% endtablerow %}{% endtablerow %}",
	} {
		tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			t.Fatalf("%q: %v", source, err)
		}
		ctx, cancel := gocontext.WithTimeout(gocontext.Background(), deadline)
		start := time.Now()
		renderRuby(tmpl, nil, &rubyOptions{Context: ctx})
		took := time.Since(start)
		cancel()
		if took > deadline+slack {
			t.Errorf("%q stopped after %v, more than %v after the deadline", source, took, slack)
		}
		errs := rubyErrors(tmpl)
		if len(errs) == 0 || !errors.Is(errs[len(errs)-1], gocontext.DeadlineExceeded) {
			t.Errorf("%q: errors %v, want one that wraps context.DeadlineExceeded", source, errs)
		}
	}
}

// TestRenderWithACancelledContextOfItsOwn checks a Context that the caller
// builds: SetCancellation stops its render too.
func TestRenderWithACancelledContextOfItsOwn(t *testing.T) {
	tmpl, err := liquid.ParseTemplate("a{% for i in (1..3) %}{{ i }}{% endfor %}", &liquid.TemplateOptions{Environment: newEnv("", limits{})})
	if err != nil {
		t.Fatal(err)
	}
	ctx := newRubyContext(newEnv("", limits{}))
	cancelled, cancel := gocontext.WithCancel(gocontext.Background())
	cancel()
	ctx.SetCancellation(cancelled)
	out := renderRuby(tmpl, ctx, nil)
	var canceledErr *liquid.CanceledError
	if errs := ctx.Errors(); len(errs) == 0 || !errors.As(errs[0], &canceledErr) || !errors.Is(errs[0], gocontext.Canceled) {
		t.Errorf("errors %v, want a CanceledError that wraps context.Canceled (output %q)", ctx.Errors(), out)
	}
}

// TestACancellationEndsWithItsRender checks that the ctx of a render stops
// only the render it is given, on a Context of the caller too; that the
// render of a Context whose cancellation is set records the stop, as it
// records a MemoryError; and that the cause of the cancellation is wrapped.
func TestACancellationEndsWithItsRender(t *testing.T) {
	env := newEnv("", limits{})
	tmpl, err := liquid.ParseTemplate("a{{ 1 }}b", &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := gocontext.WithCancel(gocontext.Background())
	cancel()

	ctx := newRubyContext(env)
	renderRuby(tmpl, ctx, &rubyOptions{Context: cancelled})
	if out := renderRuby(tmpl, ctx, nil); out != "a1b" {
		t.Errorf("a second render of the same Context renders %q, want %q", out, "a1b")
	}

	ctx = newRubyContext(env)
	ctx.SetCancellation(cancelled)
	renderRuby(tmpl, ctx, nil)
	if errs := ctx.Errors(); len(errs) == 0 || !errors.Is(errs[len(errs)-1], gocontext.Canceled) {
		t.Errorf("a render of a cancelled Context: errors %v, want one that wraps context.Canceled", errs)
	}

	cause := errors.New("the request ended")
	withCause, cancelWithCause := gocontext.WithCancelCause(gocontext.Background())
	cancelWithCause(cause)
	renderRuby(tmpl, nil, &rubyOptions{Context: withCause})
	if errs := rubyErrors(tmpl); len(errs) == 0 || !errors.Is(errs[0], cause) || !errors.Is(errs[0], gocontext.Canceled) {
		t.Errorf("errors %v, want one that wraps the cause and context.Canceled", errs)
	}
}
