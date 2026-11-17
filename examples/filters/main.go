package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Notifuse/liquidgo/v5/liquid"
	"github.com/Notifuse/liquidgo/v5/liquid/tags"
)

// reverse reverses the input string.
func reverse(_ *liquid.Context, input interface{}, _ ...interface{}) (interface{}, error) {
	runes := []rune(fmt.Sprint(input))
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes), nil
}

// shout converts the input string to uppercase and appends exclamation marks.
func shout(_ *liquid.Context, input interface{}, _ ...interface{}) (interface{}, error) {
	return strings.ToUpper(fmt.Sprint(input)) + "!!!", nil
}

func main() {
	// Create environment with standard tags and custom filters: reverse
	// takes the place of the standard reverse
	env, err := liquid.NewEnvironment(liquid.EnvironmentConfig{
		Tags:    tags.Standard(),
		Filters: []liquid.FilterDef{liquid.Filter("reverse", reverse), liquid.Filter("shout", shout)},
	})
	if err != nil {
		panic(err)
	}

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

	res, err := tmpl.Render(context.Background(), map[string]interface{}{
		"text": "hello",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(res.Output)
}
