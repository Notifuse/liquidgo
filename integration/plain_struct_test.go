package integration

// Regression tests for what a template may read of Go values (audit of
// 2026-11-15, finding 2). As in Ruby, where only a Liquid::Drop has its
// methods invoked, a plain Go struct exposes nothing: no field, no method,
// and no text of its fields. A drop, a value with ToLiquid, a String or an
// Error method keep their behaviour.

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
	PasswordHash string
}

func (u liquidUser) ToLiquid() interface{} { return map[string]interface{}{"name": u.Name} }

// userDrop is a drop: its methods and fields are readable.
type userDrop struct {
	*liquid.Drop
	Name string
}

func (d *userDrop) Greeting() string { return "hi " + d.Name }

// valueDrop embeds Drop by value: only its pointer has the drop methods, and
// a typed slice holds values.
type valueDrop struct {
	liquid.Drop
	Name string
}

func renderPlain(t *testing.T, src string, data interface{}) string {
	t.Helper()
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	tpl, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	return tpl.Render(data, nil)
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
	}
	for _, src := range []string{
		`{{ u.password_hash }}{{ pu.password_hash }}{{ u.PasswordHash }}{{ u.name }}`,
		`{{ users[0].password_hash }}{{ users.first.password_hash }}{{ pus[0].password_hash }}`,
		`{{ users | map: "password_hash" | join }}{{ pus | map: "password_hash" | join }}`,
		`{{ users | where: "name", "bob" | size }}{{ users | sort: "password_hash" | map: "name" | join }}`,
		`{{ users | first | password_hash }}`,
		`{% for x in users %}{{ x.password_hash }}{% endfor %}`,
		`{{ u }}{{ pu }}{{ users }}{{ users | join: "," }}{{ u | append: "" }}`,
		`{{ m }}{{ m | inspect }}{{ tm }}{{ tm | inspect }}{{ mm }}`,
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
	// Data given as a plain struct instead of a map exposes nothing either
	u := &secretUser{Name: "bob", PasswordHash: "secret$hash"}
	if got := renderPlain(t, `{{ password_hash }}{{ name }}{{ delete }}`, u); got != "" || u.deleted {
		t.Errorf("a plain struct as data gives %q (deleted: %v)", got, u.deleted)
	}
}

func TestToLiquidAndDropsStayReadable(t *testing.T) {
	lu := liquidUser{Name: "ann", PasswordHash: "secret$hash"}
	data := map[string]interface{}{
		"lu":    lu,
		"lus":   []liquidUser{lu, {Name: "bea"}},
		"drop":  &userDrop{Drop: liquid.NewDrop(), Name: "dan"},
		"drops": []*userDrop{{Drop: liquid.NewDrop(), Name: "eve"}},
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

// A time.Time renders and takes the date filter; its methods are not read,
// as in Ruby, where Time is not a drop.
func TestATimeRendersButExposesNoMethod(t *testing.T) {
	now := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
	data := map[string]interface{}{"t": now, "tp": &now}
	for _, c := range []struct{ src, want string }{
		{`{{ t | date: "%Y-%m-%d" }}|{{ tp | date: "%Y-%m-%d" }}`, "2026-11-15|2026-11-15"},
		{`{{ t }}|{{ tp }}`, "2026-11-15 10:00:00 +0000 UTC|2026-11-15 10:00:00 +0000 UTC"},
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
