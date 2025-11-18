package main

import (
	"fmt"
	"strings"

	"github.com/Notifuse/liquidgo/liquid"
	"github.com/Notifuse/liquidgo/liquid/tags"
)

// Custom filter
type MyFilters struct{}

func (f *MyFilters) Reverse(input string) string {
	runes := []rune(input)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func (f *MyFilters) Shout(input string) string {
	return strings.ToUpper(input) + "!!!"
}

func main() {
	// Create environment with standard tags and custom filters
	env := liquid.NewEnvironment()
	tags.RegisterStandardTags(env)
	env.RegisterFilter(&MyFilters{})

	template := `
Original: {{ text }}
Uppercase: {{ text | upcase }}
Reversed: {{ text | reverse }}
Shouted: {{ text | shout }}
Combined: {{ text | reverse | shout }}
`

	tmpl, err := liquid.ParseTemplate(template, &liquid.TemplateOptions{
		Environment: env,
	})
	if err != nil {
		panic(err)
	}

	output := tmpl.Render(map[string]interface{}{
		"text": "hello",
	}, nil)

	fmt.Println(output)
}

