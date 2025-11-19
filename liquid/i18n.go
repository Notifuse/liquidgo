package liquid

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	// DefaultLocalePath is the default path to the English locale file
	DefaultLocalePath = filepath.Join("liquid", "locales", "en.yml")
)

// I18n handles internationalization for Liquid templates.
type I18n struct {
	path   string
	locale map[string]interface{}
}

// NewI18n creates a new I18n instance with the given locale path.
func NewI18n(path string) *I18n {
	if path == "" {
		path = DefaultLocalePath
	}
	return &I18n{path: path}
}

// Translate translates a key using the locale, with optional variables for interpolation.
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

	parts := strings.Split(name, ".")
	current := locale

	for _, part := range parts {
		if val, ok := current[part]; ok {
			if str, ok := val.(string); ok {
				return str
			}
			if m, ok := val.(map[string]interface{}); ok {
				current = m
			} else {
				panic(fmt.Sprintf("Translation for %s does not exist in locale %s", name, i.path))
			}
		} else {
			panic(fmt.Sprintf("Translation for %s does not exist in locale %s", name, i.path))
		}
	}

	// If we get here, return the last value as string
	return fmt.Sprintf("%v", current)
}
