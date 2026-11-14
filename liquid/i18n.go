package liquid

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	// DefaultLocalePath is the path that names the English locale of
	// Shopify Liquid, which is embedded: NewI18n reads no file for it.
	DefaultLocalePath = filepath.Join("liquid", "locales", "en.yml")
)

// builtinLocalePath is the value DefaultLocalePath starts with.
var builtinLocalePath = DefaultLocalePath

//go:embed locales/en.yml
var builtinLocaleSource []byte

var (
	builtinLocaleOnce sync.Once
	builtinLocale     map[string]interface{}
)

// I18n handles internationalization for Liquid templates.
type I18n struct {
	locale map[string]interface{}
	path   string
}

// NewI18n creates a new I18n instance with the given locale path. An empty
// path, "en" or DefaultLocalePath gives the English locale of Shopify Liquid
// 5.11.0, embedded in the package; any other path is read from the file
// system when a message is first translated.
func NewI18n(path string) *I18n {
	if path == "" {
		path = DefaultLocalePath
	}
	return &I18n{path: path}
}

// Translate translates a key using the locale, with optional variables for
// interpolation. A key the locale does not have translates to itself.
func (i *I18n) Translate(name string, vars map[string]interface{}) string {
	if vars == nil {
		vars = make(map[string]interface{})
	}
	translation := i.deepFetchTranslation(name)
	return i.interpolate(translation, vars)
}

// T is an alias for Translate.
func (i *I18n) T(name string, vars map[string]interface{}) string {
	return i.Translate(name, vars)
}

// Locale returns the loaded locale data.
func (i *I18n) Locale() (map[string]interface{}, error) {
	if i.locale != nil {
		return i.locale, nil
	}

	if i.path == "en" || i.path == builtinLocalePath {
		builtinLocaleOnce.Do(func() {
			if err := yaml.Unmarshal(builtinLocaleSource, &builtinLocale); err != nil {
				panic(fmt.Sprintf("the embedded locale does not parse: %v", err))
			}
		})
		i.locale = builtinLocale
		return i.locale, nil
	}

	data, err := os.ReadFile(i.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read locale file %s: %w", i.path, err)
	}

	var locale map[string]interface{}
	if err := yaml.Unmarshal(data, &locale); err != nil {
		return nil, fmt.Errorf("failed to parse locale file %s: %w", i.path, err)
	}

	i.locale = locale
	return i.locale, nil
}

func (i *I18n) interpolate(name string, vars map[string]interface{}) string {
	re := regexp.MustCompile(`%\{(\w+)\}`)
	return re.ReplaceAllStringFunc(name, func(match string) string {
		key := re.FindStringSubmatch(match)[1]
		if val, ok := vars[key]; ok {
			return fmt.Sprintf("%v", val)
		}
		return match
	})
}

func (i *I18n) deepFetchTranslation(name string) string {
	locale, err := i.Locale()
	if err != nil {
		// If locale file doesn't exist, return the key itself
		return name
	}

	// Ruby's locale files nest the keys under the language; a key may be
	// given with or without it.
	if en, ok := locale["en"].(map[string]interface{}); ok && !strings.HasPrefix(name, "en.") {
		locale = en
	}

	current := locale
	parts := strings.Split(name, ".")
	for n, part := range parts {
		val, ok := current[part]
		if !ok {
			return name
		}
		if n == len(parts)-1 {
			if str, ok := val.(string); ok {
				return str
			}
			return name
		}
		m, ok := val.(map[string]interface{})
		if !ok {
			return name
		}
		current = m
	}
	return name
}

// LocaleOf returns the locale of parseContext, or the embedded English locale
// of Shopify Liquid.
func LocaleOf(parseContext ParseContextInterface) *I18n {
	if pc, ok := parseContext.(*ParseContext); ok && pc != nil && pc.Locale() != nil {
		return pc.Locale()
	}
	return NewI18n("")
}

// NewSyntaxErrorFor returns the syntax error of the locale key, as Ruby's
// SyntaxError, options[:locale].t(key, vars): errors.syntax.for gives
// "Syntax Error in 'for loop' - Valid syntax: for [item] in [collection]".
func NewSyntaxErrorFor(parseContext ParseContextInterface, key string, vars map[string]interface{}) *SyntaxError {
	return NewSyntaxError(LocaleOf(parseContext).Translate(key, vars))
}
