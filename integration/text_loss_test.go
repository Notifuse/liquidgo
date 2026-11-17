package integration

import (
	gocontext "context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/Notifuse/liquidgo/v5/liquid"
)

// A template that parses and renders keeps the text around its tags: the text
// before its first tag or output starts the output, and the text after its
// last one ends it. A delimiter read wrong can make a render lose the rest of
// a template without an error (category 6 of the security check); this
// catches it. An end next to a tag that strips whitespace ({%- or -}}) is not
// checked, nor the tail of a template that may break or continue: a break or
// a continue outside a loop ends the render there, as in Ruby
// (differential cases tags-20 and tags-21).

// textEnds returns the text before the first tag or output of source and the
// text after the last one, and whether each is to be checked. A source with
// no tag is all text.
func textEnds(source string) (head, tail string, checkHead, checkTail bool) {
	first := firstIndex(source, "{{", "{%")
	if first < 0 {
		return source, "", true, false
	}
	head = source[:first]
	checkHead = !strings.HasPrefix(source[first+2:], "-")
	last := lastIndex(source, "}}", "%}")
	if last < first {
		return head, "", checkHead, false
	}
	tail = source[last+2:]
	checkTail = last == 0 || source[last-1] != '-'
	// a tail that opens a tag or an output is no text: the parse decides
	if strings.Contains(tail, "{{") || strings.Contains(tail, "{%") || mayInterrupt.MatchString(source) {
		checkTail = false
	}
	return head, tail, checkHead, checkTail
}

// mayInterrupt finds the words break and continue, which may be a tag that
// ends the render.
var mayInterrupt = regexp.MustCompile(`\b(?:break|continue)\b`)

func firstIndex(s string, subs ...string) int {
	best := -1
	for _, sub := range subs {
		if i := strings.Index(s, sub); i >= 0 && (best < 0 || i < best) {
			best = i
		}
	}
	return best
}

func lastIndex(s string, subs ...string) int {
	best := -1
	for _, sub := range subs {
		if i := strings.LastIndex(s, sub); i > best {
			best = i
		}
	}
	return best
}

// keptEnds returns an error when output does not keep the text ends of
// source.
func keptEnds(source, output string) error {
	head, tail, checkHead, checkTail := textEnds(source)
	if firstIndex(source, "{{", "{%") < 0 {
		if output != source {
			return fmt.Errorf("a template with no tag renders %q", output)
		}
		return nil
	}
	if checkHead && !strings.HasPrefix(output, head) {
		return fmt.Errorf("the output %q does not start with the text %q before the first tag", output, head)
	}
	if checkTail && !strings.HasSuffix(output, tail) {
		return fmt.Errorf("the output %q does not end with the text %q after the last tag", output, tail)
	}
	return nil
}

// endsRender renders source in mode with data, and returns its output when
// the parse and the render succeed.
func endsRender(mode liquid.ErrorMode, source string, data map[string]interface{}) (string, bool) {
	env := newEnv(mode, allLimits)
	tmpl, err := liquid.ParseTemplate(source, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		return "", false
	}
	res, err := tmpl.Render(gocontext.Background(), data)
	return res.Output, err == nil
}

// Every template of the differential corpus, written by hand or generated,
// keeps its text ends, in lax and strict mode.
func TestTheCorpusKeepsTheTextAroundItsTags(t *testing.T) {
	raw, err := os.ReadFile(differentialDir + "/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected struct {
		Results []differentialResult `json:"results"`
	}
	if err := json.Unmarshal(raw, &expected); err != nil {
		t.Fatal(err)
	}
	data := differentialData(t)
	checked := 0
	for _, r := range expected.Results {
		output, ok := endsRender(liquid.ErrorMode(r.Mode), r.Template, data)
		if !ok {
			continue
		}
		checked++
		if err := keptEnds(r.Template, output); err != nil {
			t.Errorf("%s: %v", r.key(), err)
		}
	}
	if checked < 500 {
		t.Errorf("checked %d results, want most of the corpus", checked)
	}
}

// The text ends of a few sources, as textEnds reads them.
func TestTextEnds(t *testing.T) {
	for _, c := range []struct {
		source, head, tail   string
		checkHead, checkTail bool
	}{
		{"plain", "plain", "", true, false},
		{"a{{ x }}b", "a", "b", true, true},
		{"a {%- if x %}y{% endif -%} b", "a ", " b", false, false},
		{"a{% raw %}{{ x }}{% endraw %}z", "a", "z", true, true},
		{"a{{ x }}b{{ y", "a", "b{{ y", true, false},
	} {
		head, tail, ch, ct := textEnds(c.source)
		if head != c.head || tail != c.tail || ch != c.checkHead || ct != c.checkTail {
			t.Errorf("%q: %q %q %v %v, want %q %q %v %v", c.source, head, tail, ch, ct, c.head, c.tail, c.checkHead, c.checkTail)
		}
	}
	if err := keptEnds("a{{ x }}b", "a1"); err == nil {
		t.Error("a lost tail is not reported")
	}
}

// The end tag of a comment decides whether the text after it is trimmed, as
// Ruby's Comment#parse_body: a {%- on the opening tag, which trims the
// inside of the comment, does not reach the text after endcomment.
func TestTheEndOfACommentDecidesItsTrim(t *testing.T) {
	for source, want := range map[string]string{
		"{%comment-%}{%endcomment%} ":        " ",
		"{% comment -%}x{% endcomment %} y":  " y",
		"{% comment -%}x{% endcomment -%} y": "y",
		"{% comment %}x{% endcomment -%} y":  "y",
	} {
		for _, mode := range []liquid.ErrorMode{"lax", "strict"} {
			if got, ok := endsRender(mode, source, nil); !ok || got != want {
				t.Errorf("%s: %q renders %q (%v), want %q", mode, source, got, ok, want)
			}
		}
	}
}
