package main

import (
	"bytes"
	gocontext "context"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
	acstech "github.com/acstech/liquid"
	goliquid "github.com/go-liquid/liquid"
	nchapman "github.com/nchapman/go-liquid"
	osteele "github.com/osteele/liquid"
	tobi "github.com/tobi/liquid-go"
)

// A library parses a template once and renders it, in lax mode where it has
// one, as Shopify's outputs are.
type library struct {
	name  string
	parse func(src string) (func(data map[string]any) (string, error), error)
}

var libraries = []library{
	{"liquidgo", func(src string) (func(map[string]any) (string, error), error) {
		env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{Tags: tags.Standard()})
		if err != nil {
			return nil, err
		}
		t, err := liquid.ParseTemplate(src, &liquid.TemplateOptions{Environment: env})
		if err != nil {
			return nil, err
		}
		return func(data map[string]any) (string, error) {
			res, err := t.Render(gocontext.Background(), data)
			return res.Output, err
		}, nil
	}},
	{"osteele", func(src string) (func(map[string]any) (string, error), error) {
		t, err := osteele.NewEngine().ParseString(src)
		if err != nil {
			return nil, err
		}
		return func(data map[string]any) (string, error) {
			out, err := t.RenderString(data)
			if err != nil {
				return out, err
			}
			return out, nil
		}, nil
	}},
	{"acstech", func(src string) (func(map[string]any) (string, error), error) {
		t, err := acstech.ParseString(src, acstech.NoCache)
		if err != nil {
			return nil, err
		}
		return func(data map[string]any) (string, error) {
			var w bytes.Buffer
			t.Render(&w, data)
			return w.String(), nil
		}, nil
	}},
	{"tobi", func(src string) (func(map[string]any) (string, error), error) {
		t, err := tobi.Parse(src, tobi.ParseOptions{ErrorMode: "lax"})
		if err != nil {
			return nil, err
		}
		return func(data map[string]any) (string, error) {
			res, err := t.Render(tobi.RenderOptions{Environment: data})
			if res == nil {
				return "", err
			}
			return res.Output, err
		}, nil
	}},
	{"goliquid", func(src string) (func(map[string]any) (string, error), error) {
		t, err := goliquid.Parse(src, goliquid.WithErrorMode(goliquid.Lax))
		if err != nil {
			return nil, err
		}
		return t.Render, nil
	}},
	{"nchapman", func(src string) (func(map[string]any) (string, error), error) {
		t, err := nchapman.NewEnvironment().WithErrorMode(nchapman.ErrorModeLax).Parse(src)
		if err != nil {
			return nil, err
		}
		return func(data map[string]any) (string, error) { return t.Render(data) }, nil
	}},
}
