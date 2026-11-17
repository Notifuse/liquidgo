package integration

// Regression tests for what a template may read of Go values (audit of
// 2026-11-15, finding 2). As in Ruby, where only a Liquid::Drop has its
// methods invoked, a plain Go struct exposes nothing: no field, no method,
// and no text of its fields (Ruby raises an internal error for an object
// without to_liquid). A drop, a value with ToLiquid, a String or an Error
// method keep their behaviour.

import (
	"testing"
	"time"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// secretUser is a plain Go struct, not a drop.
type secretUser struct {
	Name         string
	PasswordHash string
	deleted      bool
}

func (u *secretUser) Delete() string { u.deleted = true; return "deleted" }

// liquidUser is the migration of secretUser: ToLiquid gives what a template reads.
type liquidUser struct {
	Name         string
	N            int
	PasswordHash string
}

func (u liquidUser) ToLiquid() interface{} { return map[string]interface{}{"name": u.Name, "n": u.N} }

// getterUser is a plain struct with a Get method, such as a configuration
// object: a template never calls it.
type getterUser struct {
	PasswordHash string
	calls        *int
}

func (g getterUser) Get(key string) interface{} { *g.calls++; return "get:" + key }

// userDrop is a drop: the keys it declares are readable.
type userDrop struct {
	Name string
}

func (d *userDrop) Greeting() string { return "hi " + d.Name }

// LiquidGet returns the key a template reads.
func (d *userDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "name":
		return d.Name, true, nil
	case "greeting":
		return d.Greeting(), true, nil
	}
	return nil, false, nil
}

// valueDrop is a drop by its pointer only, and a typed slice holds values.
type valueDrop struct {
	Name string
}

// LiquidGet returns the key a template reads.
func (d *valueDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	switch key {
	case "name":
		return d.Name, true, nil
	}
	return nil, false, nil
}

func renderPlain(t *testing.T, src string, data interface{}) string {
	t.Helper()
	env := mustEnv(liquid.EnvironmentConfig{Tags: tags.Standard(), FileSystem: &mapFileSystem{templates: map[string]string{
		"u": "{{ u.name }}{{ u.password_hash }};",
	}}})
	tpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return renderRuby(tpl, data, nil)
}

func TestAPlainStructExposesNothing(t *testing.T) {
	u := secretUser{Name: "bob", PasswordHash: "secret$hash"}
	data := map[string]interface{}{
		"u":     u,
		"pu":    &u,
		"users": []secretUser{u},
		"pus":   []*secretUser{&u},
		"m":     map[string]interface{}{"u": u},
		"tm":    map[string]secretUser{"u": u},
		"mm":    map[string]interface{}{"tm": map[string]secretUser{"u": u}},
		"list":  []interface{}{u, &u, map[string]interface{}{"u": u}},
	}
	for _, src := range []string{
		`{{ u.password_hash }}{{ pu.password_hash }}{{ u.PasswordHash }}{{ u.name }}`,
		`{{ users[0].password_hash }}{{ users.first.password_hash }}{{ pus[0].password_hash }}`,
		`{{ users | map: "password_hash" | join }}{{ pus | map: "password_hash" | join }}`,
		`{{ users | where: "name", "bob" | size }}{{ users | sort: "password_hash" | map: "name" | join }}`,
		`{{ users | first | password_hash }}{{ users | find: "name", "bob" | size }}`,
		`{% for x in users %}{{ x.password_hash }}{% cycle x.name, x.password_hash %}{% endfor %}`,
		`{% tablerow x in users %}{{ x.password_hash }}{% endtablerow %}`,
		`{% render "u" for users as u %}{% render "u", u: u %}{% include "u" with pu %}`,
		`{% case u.name %}{% when "bob" %}bob{% endcase %}{% if u.name == "bob" %}bob{% endif %}`,
		`{{ u }}{{ pu }}{{ users }}{{ users | join: "," }}{{ u | append: "" }}{{ users | append: "" }}`,
		`{{ m }}{{ tm }}{{ mm }}{{ list }}{{ list | append: "" }}{{ u | default: "x" }}`,
	} {
		if got := renderPlain(t, src, data); containsAny(got, "secret$hash", "bob") {
			t.Errorf("%s gives %q, which shows a field of a plain struct", src, got)
		}
	}
}

func TestATemplateCannotCallAMethodOfAPlainStruct(t *testing.T) {
	for _, src := range []string{`{{ u.delete }}`, `{{ u | delete }}`, `{{ u.Delete }}`} {
		u := &secretUser{Name: "bob"}
		renderPlain(t, src, map[string]interface{}{"u": u})
		if u.deleted {
			t.Errorf("%s called Delete()", src)
		}
	}
}

// listUser gives a list as its liquid form.
type listUser struct{ Names []string }

func (u listUser) ToLiquid() interface{} {
	out := make([]interface{}, len(u.Names))
	for i, n := range u.Names {
		out[i] = n
	}
	return out
}

// A value whose ToLiquid gives a list renders as that list.
func TestToLiquidGivingAList(t *testing.T) {
	data := map[string]interface{}{"l": listUser{[]string{"a", "b"}}}
	if got, want := renderPlain(t, `{{ l }}|{{ l | size }}|{{ l | join: "," }}`, data), "ab|2|a,b"; got != want {
		t.Errorf("gives %q, want %q", got, want)
	}
	// A list of them renders as the list of their lists
	ls := []listUser{{[]string{"a", "b"}}, {[]string{"c"}}}
	lists := []interface{}{[]interface{}{"a", "b"}, []interface{}{"c"}}
	src := `{{ X }}|{{ X | size }}|{{ X | first | join: "," }}`
	got := renderPlain(t, src, map[string]interface{}{"X": ls})
	if want := renderPlain(t, src, map[string]interface{}{"X": lists}); got != want {
		t.Errorf("%s gives %q, want %q as with the lists", src, got, want)
	}
}

// The array filters never call Get on a plain struct, whatever its key.
func TestTheArrayFiltersDoNotCallGetOnAPlainStruct(t *testing.T) {
	calls := 0
	gs := []getterUser{{PasswordHash: "PW", calls: &calls}}
	got := renderPlain(t, `{{ gs | map: "password" | join }}|{{ gs | where: "password", "get:password" | size }}`, map[string]interface{}{"gs": gs})
	if got != "|0" || calls != 0 {
		t.Errorf("gives %q with %d calls of Get, want %q and none", got, calls, "|0")
	}
}

// A pointer renders as what it points to, and a nil pointer as nothing.
func TestAPointerRendersAsItsTarget(t *testing.T) {
	u := &secretUser{Name: "bob", PasswordHash: "secret$hash"}
	n := 42
	pn := &n
	data := map[string]interface{}{
		"pps": &u,
		"np":  (*secretUser)(nil),
		"ntp": (*time.Time)(nil),
		"ppn": &pn,
		"nl":  []interface{}{(*secretUser)(nil), &u},
	}
	src := `[{{ pps }}][{{ pps.name }}][{{ np }}][{{ ntp }}][{{ ppn }}][{{ nl | append: "" }}]`
	if got, want := renderPlain(t, src, data), "[][][][][42][[nil, nil]]"; got != want {
		t.Errorf("gives %q, want %q", got, want)
	}
}

// A hash renders with its keys in order, so that a render does not change
// from one run to the next, for a map of any type.
func TestAHashRendersInKeyOrder(t *testing.T) {
	data := map[string]interface{}{
		"tm": map[string]string{"c": "3", "a": "1", "b": "2", "e": "5", "d": "4"},
		"m":  map[string]interface{}{"c": 3, "a": 1, "b": 2, "e": 5, "d": 4},
	}
	for i := 0; i < 20; i++ {
		got := renderPlain(t, `{{ tm }}|{{ m }}`, data)
		if want := `{"a"=>"1", "b"=>"2", "c"=>"3", "d"=>"4", "e"=>"5"}|{"a"=>1, "b"=>2, "c"=>3, "d"=>4, "e"=>5}`; got != want {
			t.Fatalf("gives %q, want %q", got, want)
		}
	}
}

// A list of values with ToLiquid renders as the list of the maps they give,
// for every filter and tag that reads list items.
func TestToLiquidBehavesAsItsMap(t *testing.T) {
	// The first and the last differ as structs but give the same map
	users := []liquidUser{{Name: "bea", N: 2, PasswordHash: "secret$hash"}, {Name: "ann", N: 1}, {Name: "bea", N: 2}}
	maps := make([]interface{}, len(users))
	for i, u := range users {
		maps[i] = u.ToLiquid()
	}
	for _, src := range []string{
		`{{ X | size }}|{{ X | uniq | size }}|{{ X | join: "," }}|{{ X | reverse | first | map: "name" }}`,
		`{{ X | first | map: "name" }}|{{ X | last | map: "name" }}|{{ X[1].name }}|{{ X.first.name }}`,
		`{{ X | map: "name" | join: "," }}|{{ X | where: "name", "bea" | size }}|{{ X | reject: "name", "bea" | size }}`,
		`{{ X | sort: "n" | map: "name" | join }}|{{ X | sort_natural: "name" | map: "n" | join }}`,
		`{{ X | uniq: "name" | size }}|{{ X | compact: "name" | size }}|{{ X | sum: "n" }}`,
		`{{ X | find: "name", "ann" | map: "n" }}|{{ X | find_index: "name", "ann" }}|{% if X | has: "name", "ann" %}has{% endif %}`,
		`{% for u in X %}{{ u.name }}{% cycle u.n, "-" %}{% endfor %}|{% tablerow u in X %}{{ u.name }}{% endtablerow %}`,
		`{% render "u" for X as u %}|{% for u in X %}{% include "u" %}{% endfor %}`,
		`{{ X }}|{{ X | append: "" }}|{{ X | first | append: "" }}`,
	} {
		got := renderPlain(t, src, map[string]interface{}{"X": users})
		want := renderPlain(t, src, map[string]interface{}{"X": maps})
		if got != want || containsAny(got, "secret$hash") {
			t.Errorf("%s gives %q, want %q as with the maps", src, got, want)
		}
	}
}

func TestToLiquidAndDropsStayReadable(t *testing.T) {
	lu := liquidUser{Name: "ann", PasswordHash: "secret$hash"}
	data := map[string]interface{}{
		"lu":    lu,
		"lus":   []liquidUser{lu, {Name: "bea"}},
		"drop":  &userDrop{Name: "dan"},
		"drops": []*userDrop{{Name: "eve"}},
		"vds":   []valueDrop{{Name: "val"}},
	}
	for _, c := range []struct{ src, want string }{
		{`{{ lu.name }}[{{ lu.password_hash }}]`, "ann[]"},
		{`{{ lus[1].name }}|{{ lus.first.name }}`, "bea|ann"},
		{`{{ lus | map: "name" | join: "," }}|{{ lus | where: "name", "bea" | size }}`, "ann,bea|1"},
		{`{{ lus | sort: "name" | map: "name" | last }}`, "bea"},
		{`{{ drop.name }}|{{ drop.greeting }}|{{ drops[0].greeting }}|{{ drops | map: "name" | join }}`, "dan|hi dan|hi eve|eve"},
		{`{{ vds[0].name }}|{{ vds | map: "name" | join }}`, "val|val"},
	} {
		if got := renderPlain(t, c.src, data); got != c.want {
			t.Errorf("%s gives %q, want %q", c.src, got, c.want)
		}
	}
}

// A time.Time renders as Ruby's Time#to_s and takes the date filter; its
// methods are not read, as in Ruby, where Time is not a drop.
func TestATimeRendersButExposesNoMethod(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	paris := time.FixedZone("CET", 3600)
	var unset *time.Time
	data := map[string]interface{}{
		"t":     now,
		"tp":    &now,
		"local": now.In(paris),
		"frac":  map[string]interface{}{"at": now.Add(500 * time.Millisecond)},
		"unset": unset,
	}
	for _, c := range []struct{ src, want string }{
		{`{{ t | date: "%Y-%m-%d" }}|{{ tp | date: "%Y-%m-%d" }}`, "2026-11-15|2026-11-15"},
		{`{{ t }}|{{ tp }}`, "2026-11-15 10:00:00 UTC|2026-11-15 10:00:00 UTC"},
		{`{{ local }}|{{ t | append: "" }}|[{{ unset }}]`, "2026-11-15 11:00:00 +0100|2026-11-15 10:00:00 UTC|[]"},
		// inside a hash, a time is written as Time#inspect, with its fraction of a second
		{`{{ frac }}`, `{"at"=>2026-11-15 10:00:00.5 UTC}`},
		{`[{{ t.year }}][{{ tp.year }}]`, "[][]"},
	} {
		if got := renderPlain(t, c.src, data); got != c.want {
			t.Errorf("%s gives %q, want %q", c.src, got, c.want)
		}
	}
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
	}
	return false
}

// secretDrop is a drop with no String method: its fields are not keys.
type secretDrop struct {
	Name         string
	PasswordHash string
}

func (d *secretDrop) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	if key == "name" {
		return d.Name, true, nil
	}
	return nil, false, nil
}

// A drop is written by its String method only: one without it never writes
// its fields, alone, in a list or in a hash, nor do the loop drops.
func TestADropNeverWritesItsFields(t *testing.T) {
	d := &secretDrop{Name: "bob", PasswordHash: "secret$hash"}
	data := map[string]interface{}{
		"d":  d,
		"ds": []*secretDrop{d},
		"vs": []secretDrop{*d},
		"h":  map[string]interface{}{"d": d},
	}
	src := `{{ d }}|{{ ds }}|{{ vs }}|{{ h }}|{{ d.name }}|{% for i in (1..1) %}{{ forloop }}{% endfor %}|{% tablerow i in (1..1) %}{{ tablerowloop }}{% endtablerow %}`
	got := renderPlain(t, src, data)
	if containsAny(got, "secret$hash", "i-(1..1)", "&{") {
		t.Errorf("%s gives %q: a drop wrote its fields", src, got)
	}
	if !containsAny(got, "|bob|") {
		t.Errorf("%s gives %q: the key of the drop is not readable", src, got)
	}
}

// secretOrders is a drop on a named slice of plain structs.
type secretOrder struct {
	ID         int
	CardNumber string
}

type secretOrders []secretOrder

func (o *secretOrders) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	if key == "count" {
		return len(*o), true, nil
	}
	return nil, false, nil
}

type secretOrderMap map[string]secretOrder

func (o *secretOrderMap) LiquidGet(_ *liquid.Context, key string) (interface{}, bool, error) {
	return nil, false, nil
}

// A drop on a slice or a map of plain structs, given by pointer, never writes
// the fields of what it holds, in the output, through a filter or in a hash.
func TestADropOnASliceNeverWritesTheFieldsItHolds(t *testing.T) {
	orders := secretOrders{{1, "4242-S3CR3T"}}
	byKey := secretOrderMap{"k": {2, "4343-S3CR3T"}}
	data := map[string]interface{}{"o": &orders, "m": &byKey, "h": map[string]interface{}{"o": &orders}, "list": []interface{}{&orders}}
	src := `{{ o }}|{{ o | append: "" }}|{{ o | upcase }}|{{ o | default: "" }}|{{ list | join: "," }}|{{ list | first }}|{{ h }}|{{ m }}|{{ o.count }}`
	got := renderPlain(t, src, data)
	if containsAny(got, "S3CR3T") {
		t.Errorf("%s gives %q: a drop wrote the fields it holds", src, got)
	}
	if !containsAny(got, "|1") {
		t.Errorf("%s gives %q: the key of the drop is not readable", src, got)
	}
}
